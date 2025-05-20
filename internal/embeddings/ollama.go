package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
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
		model = "nomic-embed-text"
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

// embedResponse matches the JSON structure returned by the Ollama API.
type embedResponse struct {
	Model           string      `json:"model"`
	Embedding       [][]float32 `json:"embeddings"`
	TotalDuration   int         `json:"total_duration"`
	LoadDuration    int         `json:"load_duration"`
	PromptEvalCount int         `json:"prompt_eval_count"`
}

func (p *OllamaProvider) Embed(ctx context.Context, text string) ([]float32, error) {
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

	if len(er.Embedding) == 0 {
		return nil, errors.New("no embeddings returned")
	}

	return er.Embedding[0], nil
}
