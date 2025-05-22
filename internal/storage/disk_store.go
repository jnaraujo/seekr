package storage

import (
	"bufio"
	"cmp"
	"context"
	"encoding/json"
	"io"
	"os"
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

	for i, e := range ds.documents {
		if e.ID == document.ID {
			ds.documents = slices.Delete(ds.documents, i, i+1)
			break
		}
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
		for _, emb := range doc.Embeddings {
			score := vector.CosineSimilarity(query, emb)
			if score > bestScore {
				bestScore = score
			}
		}
		if bestScore <= 0 {
			continue
		}
		results = append(results, SearchResult{
			Document: doc,
			Score:    bestScore,
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

	for _, doc := range ds.documents {
		if doc.ID == id {
			return doc, nil
		}
	}

	return document.Document{}, ErrNotFound
}
