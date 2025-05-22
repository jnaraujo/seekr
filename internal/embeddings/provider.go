package embeddings

import "context"

type Chunk struct {
	Embedding []float32
	Block     string
}

type Provider interface {
	Embed(ctx context.Context, text string) ([]Chunk, error)
}
