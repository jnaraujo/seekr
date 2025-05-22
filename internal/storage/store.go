package storage

import (
	"context"
	"errors"

	"github.com/jnaraujo/seekr/internal/document"
)

var ErrNotFound = errors.New("document not found")

type SearchResult struct {
	Document          document.Document
	Score             float32
	BestMatchingChunk int
}

type Store interface {
	// Index stores an embedding vector under a document ID with optional metadata.
	Index(ctx context.Context, document document.Document) error
	// Search finds the top K closest embeddings to the given query vector.
	Search(ctx context.Context, query []float32, topK int) ([]SearchResult, error)
	// Get retrieves a stored embedding and metadata by document ID.
	Get(ctx context.Context, id string) (document.Document, error)
	// Returns a list of all stored documents.
	List(ctx context.Context) ([]document.Document, error)
	Remove(ctx context.Context, id string) error

	// Closes the store
	Close() error
}
