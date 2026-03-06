package embedder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go-ai-rag/internal/config"
)

// Ollama is the embeddings client using Ollama /api/embed endpoint
type Ollama struct {
	cfg    *config.Config
	client *http.Client
}

// NewOllama creates a new Ollama embedder
func NewOllama(cfg *config.Config) *Ollama {
	return &Ollama{cfg: cfg, client: &http.Client{}}
}

type embedRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

// Embed returns the embedding vector for a given text
func (o *Ollama) Embed(text string) ([]float32, error) {
	body, err := json.Marshal(embedRequest{
		Model: o.cfg.EmbedModel,
		Input: text,
	})
	if err != nil {
		return nil, err
	}

	resp, err := o.client.Post(o.cfg.OllamaURL+"/api/embed", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama embed request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama embed status %d: %s", resp.StatusCode, string(data))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result embedResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("ollama embed parse: %w", err)
	}

	if len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("ollama embed: empty response")
	}

	return result.Embeddings[0], nil
}
