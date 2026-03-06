// internal/llm/gemini.go
// Gemini API chat provider — drop-in replacement for Ollama.
// Uses Google Generative Language REST API v1beta.
package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go-ai-rag/internal/config"
)

const geminiBaseURL = "https://generativelanguage.googleapis.com/v1beta/models"

// Gemini is the LLM client for Google Gemini API
type Gemini struct {
	cfg    *config.Config
	client *http.Client
}

// NewGemini creates a new Gemini LLM client
func NewGemini(cfg *config.Config) *Gemini {
	return &Gemini{cfg: cfg, client: &http.Client{}}
}

// ── Gemini API request/response types ────────────────────────────────────────

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiSystemInstruction struct {
	Parts []geminiPart `json:"parts"`
}

type geminiRequest struct {
	SystemInstruction *geminiSystemInstruction `json:"system_instruction,omitempty"`
	Contents          []geminiContent          `json:"contents"`
}

type geminiCandidate struct {
	Content geminiContent `json:"content"`
}

type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
}

// ── helpers ───────────────────────────────────────────────────────────────────

// buildGeminiRequest converts []Message (Ollama format) to Gemini API format.
// Gemini rules:
//   - "system" role → goes into SystemInstruction (not contents)
//   - "assistant" role → becomes "model" in Gemini
//   - "user" role → stays "user"
func buildGeminiRequest(messages []Message) geminiRequest {
	req := geminiRequest{}

	for _, m := range messages {
		switch m.Role {
		case "system":
			req.SystemInstruction = &geminiSystemInstruction{
				Parts: []geminiPart{{Text: m.Content}},
			}
		case "assistant":
			req.Contents = append(req.Contents, geminiContent{
				Role:  "model",
				Parts: []geminiPart{{Text: m.Content}},
			})
		default: // "user"
			req.Contents = append(req.Contents, geminiContent{
				Role:  "user",
				Parts: []geminiPart{{Text: m.Content}},
			})
		}
	}
	return req
}

// ── Chat (non-streaming) ──────────────────────────────────────────────────────

// Chat sends messages to Gemini and returns the full response text.
func (g *Gemini) Chat(messages []Message) (string, error) {
	url := fmt.Sprintf("%s/%s:generateContent?key=%s",
		geminiBaseURL, g.cfg.GeminiModel, g.cfg.GeminiAPIKey)

	body, err := json.Marshal(buildGeminiRequest(messages))
	if err != nil {
		return "", err
	}

	resp, err := g.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("gemini chat request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gemini chat status %d: %s", resp.StatusCode, string(data))
	}

	var result geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("gemini chat parse: %w", err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini returned empty response")
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}

// ── ChatStream (SSE streaming) ────────────────────────────────────────────────

// ChatStream calls Gemini with streamGenerateContent and calls onToken for each token.
func (g *Gemini) ChatStream(messages []Message, onToken func(token string) error) error {
	url := fmt.Sprintf("%s/%s:streamGenerateContent?key=%s&alt=sse",
		geminiBaseURL, g.cfg.GeminiModel, g.cfg.GeminiAPIKey)

	body, err := json.Marshal(buildGeminiRequest(messages))
	if err != nil {
		return err
	}

	resp, err := g.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("gemini stream request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gemini stream status %d: %s", resp.StatusCode, string(data))
	}

	// Gemini SSE format: each line is "data: {json}"
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		jsonData := strings.TrimPrefix(line, "data: ")
		if jsonData == "" {
			continue
		}

		var chunk geminiResponse
		if err := json.Unmarshal([]byte(jsonData), &chunk); err != nil {
			continue
		}

		if len(chunk.Candidates) > 0 && len(chunk.Candidates[0].Content.Parts) > 0 {
			token := chunk.Candidates[0].Content.Parts[0].Text
			if token != "" {
				if err := onToken(token); err != nil {
					return err
				}
			}
		}
	}
	return scanner.Err()
}
