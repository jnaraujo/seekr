package document

import (
	"errors"
	"time"

	"github.com/jnaraujo/seekr/internal/vector"
)

type Metadata = map[string]string

type Chunk struct {
	Embedding []float32
	Block     string
}

type Document struct {
	ID        string
	Chunks    []Chunk
	Content   string
	CreatedAt time.Time
	Path      string
}

func NewDocument(id string, chunks []Chunk, content string, createdAt time.Time, path string) (Document, error) {
	if id == "" {
		return Document{}, errors.New("id is empty")
	}

	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	normalizedChunks := make([]Chunk, len(chunks))
	for i, chunk := range chunks {
		chunk.Embedding = vector.Normalize(chunk.Embedding)
		normalizedChunks[i] = chunk
	}

	return Document{
		ID:        id,
		Chunks:    normalizedChunks,
		Content:   content,
		CreatedAt: createdAt,
	}, nil
}
