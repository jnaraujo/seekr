package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jnaraujo/seekr/internal/config"
	"github.com/jnaraujo/seekr/internal/textsplitter"
	"github.com/jnaraujo/seekr/internal/vector"
)

type OllamaProvider struct {
	baseURL string
	model   string
	client  *http.Client
}

var _ Provider = &OllamaProvider{}

const defaultBaseURLOllama = "http://localhost:11434/api"

func NewOllamaProvider(model, baseURL string) *OllamaProvider {
	if baseURL == "" {
		baseURL = defaultBaseURLOllama
	}
	if model == "" {
		model = config.DefaultEmbeddingModel
	}

	return &OllamaProvider{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

// embedRequest matches the JSON structure sent to the Ollama API.
type embedRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embedResponse struct {
	Model           string      `json:"model"`
	Embedding       [][]float32 `json:"embeddings"`
	TotalDuration   int         `json:"total_duration"`
	LoadDuration    int         `json:"load_duration"`
	PromptEvalCount int         `json:"prompt_eval_count"`
}

var splitter = textsplitter.NewRecursiveCharacterTextSplitter(config.MaxChunkChars, config.ChunkOverlapping)

func (p *OllamaProvider) Embed(ctx context.Context, text string) ([]Chunk, error) {
	blocks := splitter.SplitText(text)
	chunks := make([]Chunk, 0, len(blocks))
	for _, block := range blocks {
		emb, err := p.embedBlock(ctx, block)
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, Chunk{
			Block:     block,
			Embedding: emb,
		})
	}

	return chunks, nil
}

func (p *OllamaProvider) embedBlock(ctx context.Context, text string) ([]float32, error) {
	reqBody, err := json.Marshal(embedRequest{
		Model: p.model,
		Input: text,
	})
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/embed", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader([]byte(reqBody)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama API returned status %d", resp.StatusCode)
	}

	var er embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	embedding := er.Embedding[0]

	if len(embedding) == 0 {
		return nil, errors.New("no embeddings returned")
	}

	if len(embedding) != config.EmbeddingDimension {
		return nil, fmt.Errorf("expected %d dimensions, got %d", config.EmbeddingDimension, len(er.Embedding))
	}

	if !vector.IsNormalized(embedding) {
		slog.Info("embedding not normalized, normalizing")
		embedding = vector.Normalize(embedding)
	}

	return embedding, nil
}
