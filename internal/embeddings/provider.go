package embeddings

import "context"

type Chunk struct {
	Embedding []float32
}

type Provider interface {
	Embed(ctx context.Context, text string) ([]Chunk, error)
}
