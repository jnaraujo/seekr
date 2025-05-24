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
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 100*1024*1024)
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

type docProcessingResult struct {
	doc            *document.Document
	bestScore      float32
	bestChunkIndex int
}

func (ds *DiskStore) Search(ctx context.Context, query []float32, topK int) ([]SearchResult, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	query = vector.Normalize(query)

	if len(ds.documents) == 0 {
		return nil, ErrNotFound
	}

	numWorkers := 4

	docChan := make(chan *document.Document, numWorkers)
	resultsChan := make(chan docProcessingResult, len(ds.documents))
	var wg sync.WaitGroup

	wg.Add(numWorkers)
	for range numWorkers {
		go func() {
			defer wg.Done()

			for doc := range docChan {
				var bestScore float32 = -2.0
				var bestChunkIndex = -1

				for chunkIdx, chunk := range doc.Chunks {
					score := vector.FastCosineSimilarity(query, chunk.Embedding)
					if score > bestScore {
						bestScore = score
						bestChunkIndex = chunkIdx
					}
				}

				if bestChunkIndex != -1 && bestScore > 0 {
					resultsChan <- docProcessingResult{
						doc:            doc,
						bestScore:      bestScore,
						bestChunkIndex: bestChunkIndex,
					}
				}
			}

		}()
	}

	for _, doc := range ds.documents {
		docChan <- &doc
	}
	close(docChan)

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	results := make([]SearchResult, 0, len(ds.documents))
	for res := range resultsChan {
		results = append(results, SearchResult{
			Document:          *res.doc,
			Score:             res.bestScore,
			BestMatchingChunk: res.bestChunkIndex,
		})
	}

	slices.SortFunc(results, func(a, b SearchResult) int {
		return cmp.Compare(b.Score, a.Score)
	})

	return results[:topK], nil
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

func (ds *DiskStore) calculateDistribution(totalItems, numWorkers int) []int {
	if numWorkers <= 0 {
		// Handle invalid case, perhaps return nil or an error
		// For simplicity here, we might return nil or an empty slice depending on desired behavior.
		// Let's return nil for this example to indicate an issue.
		return nil
	}
	if totalItems == 0 {
		workloads := make([]int, numWorkers)
		// All workers receive 0 items
		return workloads
	}

	baseItemsPerWorker := totalItems / numWorkers
	remainingItems := totalItems % numWorkers

	workloads := make([]int, numWorkers)
	for i := 0; i < numWorkers; i++ {
		workloads[i] = baseItemsPerWorker
		if remainingItems > 0 {
			workloads[i]++
			remainingItems--
		}
	}
	return workloads
}
