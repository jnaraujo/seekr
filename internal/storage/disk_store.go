package storage

import (
	"bufio"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/jnaraujo/seekr/internal/document"
	"github.com/jnaraujo/seekr/internal/vector"
)

type DiskStore struct {
	filePath  string
	mu        sync.RWMutex
	file      *os.File
	documents []document.Document
}

// checks if DiskStore implements the Store interface
var _ Store = &DiskStore{}

func NewDiskStore(path string) (*DiskStore, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}

	ds := &DiskStore{filePath: path, file: f, documents: make([]document.Document, 0)}

	err = ds.load()
	if err != nil {
		f.Close()
		return nil, err
	}

	_, err = f.Seek(0, io.SeekEnd)
	return ds, err
}

func (s *DiskStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.documents = s.documents[:0]
	scanner := bufio.NewScanner(s.file)
	buf := make([]byte, 0, 64*1024) // 64KB buffer
	scanner.Buffer(buf, 1024*1024)  // 1MB max line size
	for scanner.Scan() {
		var doc document.Document
		if err := json.Unmarshal(scanner.Bytes(), &doc); err != nil {
			return err
		}
		s.documents = append(s.documents, doc)
	}
	return scanner.Err()
}

func (ds *DiskStore) Index(ctx context.Context, document document.Document) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	_, err := ds.getInternal(ctx, document.ID)
	if !errors.Is(err, ErrNotFound) {
		return err
	}

	ds.documents = append(ds.documents, document)

	line, err := json.Marshal(document)
	if err != nil {
		return err
	}
	if _, err := ds.file.Write(append(line, '\n')); err != nil {
		return err
	}
	// Ensure write is flushed to disk
	return ds.file.Sync()
}

func (ds *DiskStore) Remove(ctx context.Context, id string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	foundIndex := -1
	for i, e := range ds.documents {
		if e.ID == id {
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		return ErrNotFound
	}

	tempFileDir := filepath.Dir(ds.filePath)
	tempFile, err := os.CreateTemp(tempFileDir, filepath.Base(ds.filePath)+".*.tmp")
	if err != nil {
		return fmt.Errorf("Remove: failed to create temporary file: %w", err)
	}

	tempFilePath := tempFile.Name()
	operationSuccessful := false
	defer func() {
		tempFile.Close()
		if !operationSuccessful {
			os.Remove(tempFilePath)
		}
	}()

	writer := bufio.NewWriter(tempFile)
	for i, doc := range ds.documents {
		if i == foundIndex {
			continue
		}
		line, err := json.Marshal(doc)
		if err != nil {
			return fmt.Errorf("remove: failed to marshal document ID %s: %w", doc.ID, err)
		}
		if _, writeErr := writer.Write(append(line, '\n')); writeErr != nil {
			return fmt.Errorf("remove: failed to write document ID %s to temporary file: %w", doc.ID, writeErr)
		}
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	if err = tempFile.Sync(); err != nil {
		return fmt.Errorf("remove: failed to sync temporary file: %w", err)
	}

	if err = tempFile.Close(); err != nil {
		return fmt.Errorf("remove: failed to close temporary file: %w", err)
	}

	if err = ds.file.Close(); err != nil {
		return fmt.Errorf("remove: failed to close main data file (%s): %w "+
			"Temporary data not applied", ds.filePath, err)
	}

	if err = os.Rename(tempFilePath, ds.filePath); err != nil {
		currentFile, err := os.OpenFile(ds.filePath, os.O_CREATE|os.O_RDWR, 0o644)
		if err == nil {
			ds.file = currentFile
		}
		return fmt.Errorf("remove: CRITICAL - failed to rename temporary file from %s to %s: %w. "+
			"The new data *might* still be in %s (check cleanup logic). Original file state: %s",
			tempFilePath, ds.filePath, err, tempFilePath, ds.filePath)
	}

	newFile, err := os.OpenFile(ds.filePath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return fmt.Errorf("remove: CRITICAL - data file %s updated successfully on disk, but failed to reopen: %w. "+
			"DiskStore's file handle is invalid. In-memory document list is OUT OF SYNC with disk", ds.filePath, err)
	}
	ds.file = newFile
	ds.documents = slices.Delete(ds.documents, foundIndex, foundIndex+1)
	operationSuccessful = true
	return nil
}

func (ds *DiskStore) Search(ctx context.Context, query []float32, topK int) ([]SearchResult, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	query = vector.Normalize(query)

	if len(ds.documents) == 0 {
		return nil, ErrNotFound
	}

	results := make([]SearchResult, 0, len(ds.documents))
	for _, doc := range ds.documents {
		var bestScore float32
		var bestChunkIndex = 0
		for i, chunk := range doc.Chunks {
			score := vector.CosineSimilarity(query, chunk.Embedding)
			if score > bestScore {
				bestScore = score
				bestChunkIndex = i
			}
		}
		if bestScore <= 0 {
			continue
		}
		results = append(results, SearchResult{
			Document:          doc,
			Score:             bestScore,
			BestMatchingChunk: bestChunkIndex,
		})
	}

	slices.SortFunc(results, func(a, b SearchResult) int {
		return cmp.Compare(b.Score, a.Score)
	})

	return results, nil
}

func (ds *DiskStore) Get(ctx context.Context, id string) (document.Document, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	return ds.getInternal(ctx, id)
}

func (ds *DiskStore) getInternal(_ context.Context, id string) (document.Document, error) {
	for _, doc := range ds.documents {
		if doc.ID == id {
			return doc, nil
		}
	}

	return document.Document{}, ErrNotFound
}

func (ds *DiskStore) List(ctx context.Context) ([]document.Document, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.documents, nil
}

func (ds *DiskStore) Close() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	return ds.file.Close()
}
