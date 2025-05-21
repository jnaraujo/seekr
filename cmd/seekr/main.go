package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jnaraujo/seekr/internal/embeddings"
	"github.com/jnaraujo/seekr/internal/env"
	"github.com/jnaraujo/seekr/internal/vector"
)

func main() {
	var embeddingProvider embeddings.Provider = embeddings.NewOllamaProvider("nomic-embed-text", env.Env.OllamaAPIUrl)

	vec1, err := embeddingProvider.Embed(context.Background(), "the car is red")
	if err != nil {
		log.Fatalf("Failed to get embedding: %v", err)
	}

	vec2, err := embeddingProvider.Embed(context.Background(), "the sky is blue")
	if err != nil {
		log.Fatalf("Failed to get embedding: %v", err)
	}

	fmt.Println(vector.CosineSimilarity(vec1, vec2))
}
