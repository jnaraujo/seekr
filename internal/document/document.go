package document

import (
	"errors"
	"time"

	"github.com/jnaraujo/seekr/internal/embeddings"
	"github.com/jnaraujo/seekr/internal/vector"
)

type Metadata = map[string]string

type Document struct {
	ID        string
	Chunks    []embeddings.Chunk
	CreatedAt time.Time
	Path      string
}

func NewDocument(id string, chunks []embeddings.Chunk, createdAt time.Time, path string) (Document, error) {
	if id == "" {
		return Document{}, errors.New("id is empty")
	}

	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	normalizedChunks := make([]embeddings.Chunk, len(chunks))
	for i, chunk := range chunks {
		chunk.Embedding = vector.Normalize(chunk.Embedding)
		normalizedChunks[i] = chunk
	}

	return Document{
		ID:        id,
		Chunks:    normalizedChunks,
		CreatedAt: createdAt,
		Path:      path,
	}, nil
}
