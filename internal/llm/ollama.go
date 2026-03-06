package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go-ai-rag/internal/config"
)

// Message represents a single chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Ollama is the LLM client for Ollama chat API
type Ollama struct {
	cfg    *config.Config
	client *http.Client
}

// NewOllama creates a new Ollama LLM client
func NewOllama(cfg *config.Config) *Ollama {
	return &Ollama{cfg: cfg, client: &http.Client{}}
}

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type chatResponse struct {
	Message Message `json:"message"`
}

// Chat sends messages to Ollama and returns the assistant response
func (o *Ollama) Chat(messages []Message) (string, error) {
	body, err := json.Marshal(chatRequest{
		Model:    o.cfg.OllamaModel,
		Messages: messages,
		Stream:   false,
	})
	if err != nil {
		return "", err
	}

	resp, err := o.client.Post(o.cfg.OllamaURL+"/api/chat", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ollama chat request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama chat status %d: %s", resp.StatusCode, string(data))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result chatResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("ollama chat parse: %w", err)
	}

	return result.Message.Content, nil
}

type streamChunk struct {
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

// ChatStream calls Ollama with stream:true and calls onToken for each token.
// This avoids ngrok/proxy timeouts for slow models.
func (o *Ollama) ChatStream(messages []Message, onToken func(token string) error) error {
	body, err := json.Marshal(chatRequest{
		Model:    o.cfg.OllamaModel,
		Messages: messages,
		Stream:   true,
	})
	if err != nil {
		return err
	}

	resp, err := o.client.Post(o.cfg.OllamaURL+"/api/chat", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("ollama stream request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ollama stream status %d: %s", resp.StatusCode, string(data))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var chunk streamChunk
		if err := json.Unmarshal(line, &chunk); err != nil {
			continue
		}
		if chunk.Message.Content != "" {
			if err := onToken(chunk.Message.Content); err != nil {
				return err
			}
		}
		if chunk.Done {
			break
		}
	}
	return scanner.Err()
}
