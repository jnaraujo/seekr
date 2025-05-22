package document

import (
	"errors"

	"github.com/jnaraujo/seekr/internal/vector"
)

type Metadata = map[string]string

type Document struct {
	ID         string
	Metadata   Metadata
	Embeddings [][]float32
	Content    string
}

func NewDocument(id string, metadata Metadata, embeddings [][]float32, content string) (Document, error) {
	if id == "" {
		return Document{}, errors.New("id is empty")
	}

	normalizedEmbeddings := make([][]float32, len(embeddings))
	for i, emb := range embeddings {
		normalizedEmbeddings[i] = vector.Normalize(emb)
	}

	return Document{
		ID:         id,
		Metadata:   metadata,
		Embeddings: normalizedEmbeddings,
		Content:    content,
	}, nil
}
