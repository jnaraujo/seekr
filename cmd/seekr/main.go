package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jnaraujo/seekr/internal/embeddings"
	"github.com/jnaraujo/seekr/internal/env"
)

func main() {
	var embeddingProvider embeddings.Provider = embeddings.NewOllamaProvider("nomic-embed-text", env.Env.OllamaAPIUrl)

	embVector, err := embeddingProvider.Embed(context.Background(), "ola como vai")
	if err != nil {
		log.Fatalf("Failed to get embedding: %v", err)
	}

	fmt.Println("Embedding vector:")
	for i, v := range embVector {
		fmt.Printf("[%d]: %.6f\n", i, v)
	}
}
