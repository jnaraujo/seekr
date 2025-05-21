package document

import (
	"errors"

	"github.com/jnaraujo/seekr/internal/vector"
)

type Metadata = map[string]string

type Document struct {
	ID        string
	Metadata  Metadata
	Embedding []float32
	Content   string
}

func NewDocument(id string, metadata Metadata, embedding []float32, content string) (Document, error) {
	if id == "" {
		return Document{}, errors.New("id is empty")
	}

	embedding = vector.Normalize(embedding)

	return Document{
		ID:        id,
		Metadata:  metadata,
		Embedding: embedding,
		Content:   content,
	}, nil
}
