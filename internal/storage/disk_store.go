package storage

import (
	"bytes"
	"cmp"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"math"
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

func (ds *DiskStore) load() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	data, err := os.ReadFile(ds.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	if len(data) == 0 {
		return nil
	}

	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&ds.documents); err != nil {
		if err == io.EOF {
			return nil
		}
		return fmt.Errorf("failed to decode gob data: %w", err)
	}

	return nil
}

func (ds *DiskStore) persist() error {
	tempFile, err := os.CreateTemp(filepath.Dir(ds.filePath), filepath.Base(ds.filePath)+".*.tmp")
	if err != nil {
		return fmt.Errorf("persist: failed to create temporary file: %w", err)
	}
	tempFilePath := tempFile.Name()

	encoder := gob.NewEncoder(tempFile)
	if err := encoder.Encode(ds.documents); err != nil {
		tempFile.Close()
		os.Remove(tempFilePath)
		return fmt.Errorf("persist: failed to encode documents to gob: %w", err)
	}

	if err := tempFile.Sync(); err != nil {
		tempFile.Close()
		os.Remove(tempFilePath)
		return fmt.Errorf("persist: failed to sync temporary file: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		os.Remove(tempFilePath)
		return fmt.Errorf("persist: failed to close temporary file: %w", err)
	}

	if err := os.Rename(tempFilePath, ds.filePath); err != nil {
		return fmt.Errorf("persist: failed to rename temporary file: %w", err)
	}

	return nil
}

func (ds *DiskStore) Index(ctx context.Context, document document.Document) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	_, err := ds.getInternal(ctx, document.ID)
	if !errors.Is(err, ErrNotFound) {
		return err
	}

	ds.documents = append(ds.documents, document)

	return ds.persist()
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

	ds.documents = slices.Delete(ds.documents, foundIndex, foundIndex+1)
	return ds.persist()
}

func (ds *DiskStore) Search(ctx context.Context, query []float32, topK int) ([]SearchResult, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	query = vector.Normalize(query)

	if len(ds.documents) == 0 {
		return []SearchResult{}, nil
	}

	results := make([]SearchResult, 0, len(ds.documents))

	for _, doc := range ds.documents {
		var bestScore float32
		var bestChunkIndex = 0
		for i, chunk := range doc.Chunks {
			score := vector.FastCosineSimilarity(query, chunk.Embedding)
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

	max := int(math.Min(float64(topK), float64(len(ds.documents))))
	return results[:max], nil
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
	docsCopy := make([]document.Document, len(ds.documents))
	copy(docsCopy, ds.documents)
	return docsCopy, nil
}

func (ds *DiskStore) Close() error {
	return nil
}
