package document

import (
	"errors"

	"github.com/jnaraujo/seekr/internal/vector"
)

type Metadata = map[string]string

type Chunk struct {
	Embedding []float32
	Block     string
}

type Document struct {
	ID       string
	Metadata Metadata
	Chunks   []Chunk
	Content  string
}

func NewDocument(id string, metadata Metadata, chunks []Chunk, content string) (Document, error) {
	if id == "" {
		return Document{}, errors.New("id is empty")
	}

	normalizedChunks := make([]Chunk, len(chunks))
	for i, chunk := range chunks {
		chunk.Embedding = vector.Normalize(chunk.Embedding)
		normalizedChunks[i] = chunk
	}

	return Document{
		ID:       id,
		Metadata: metadata,
		Chunks:   normalizedChunks,
		Content:  content,
	}, nil
}
