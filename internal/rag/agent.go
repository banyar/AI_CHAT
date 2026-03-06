package rag

import (
	"fmt"
	"strings"

	"go-ai-rag/internal/config"
	"go-ai-rag/internal/cpe"
	"go-ai-rag/internal/customer"
	"go-ai-rag/internal/embedder"
	"go-ai-rag/internal/llm"
	"go-ai-rag/internal/memory"
	"go-ai-rag/internal/postprocess"
	"go-ai-rag/internal/vectorstore"
)

// Agent is the RAG chatbot agent that combines LLM + embeddings + vector store + memory
type Agent struct {
	llm         llm.LLM
	embedder    *embedder.Ollama
	vectorstore *vectorstore.Qdrant
	memory      *memory.Simple
	cfg         *config.Config
}

// NewAgent creates a new RAG Agent
func NewAgent(
	l llm.LLM,
	e *embedder.Ollama,
	vs *vectorstore.Qdrant,
	m *memory.Simple,
	cfg *config.Config,
) *Agent {
	return &Agent{llm: l, embedder: e, vectorstore: vs, memory: m, cfg: cfg}
}

// buildContext combines Qdrant RAG docs + CPE API data into one context string.
func buildContext(docs []string, userMsg string) string {
	var parts []string

	// CPE: if message contains a CPE ID, fetch live data from API
	if cpeID := cpe.ExtractID(userMsg); cpeID != "" {
		parts = append(parts, cpe.FetchInfo(cpeID))
	}

	// Customer: if message contains a Myanmar phone number, fetch customer data from API
	if phone := customer.ExtractPhone(userMsg); phone != "" {
		parts = append(parts, customer.FetchInfo(phone))
	}

	parts = append(parts, docs...)
	return strings.Join(parts, "\n\n---\n\n")
}

const basePrompt = `You are Frontiir AI Assistant — a helpful customer support assistant for Frontiir, Myanmar's leading fiber internet provider.

LANGUAGE RULES (strictly follow):
- If the user writes in Burmese (မြန်မာ), reply ONLY in pure Burmese. Do NOT mix English words mid-sentence.
- If the user writes in English, reply ONLY in English.
- Never combine two languages in one word. Write "ဥပမာ" OR "example" — never "ဥပမable".
- Technical terms (CPE, ID, router, ONT, signal, status) may stay in English inside a Burmese reply.
- Use "ဥပမာ -" (not "ဥပမာ:" or "example:") when giving examples in Burmese.

Be helpful, clear, and concise.`

// Chat processes a user message through the full RAG pipeline and returns the response
func (a *Agent) Chat(userMsg string) (string, error) {
	queryVec, err := a.embedder.Embed(userMsg)
	if err != nil {
		return "", fmt.Errorf("embed query: %w", err)
	}

	docs, _ := a.vectorstore.Search(a.cfg.CollectionName, queryVec, a.cfg.TopK)

	systemPrompt := basePrompt
	if ctx := buildContext(docs, userMsg); ctx != "" {
		systemPrompt = fmt.Sprintf(
			"%s\n\nUse the following context to answer. If not relevant, use your own knowledge.\n\nContext:\n%s",
			basePrompt, ctx,
		)
	}

	messages := []llm.Message{{Role: "system", Content: systemPrompt}}
	messages = append(messages, a.memory.Get()...)
	messages = append(messages, llm.Message{Role: "user", Content: userMsg})

	response, err := a.llm.Chat(messages)
	if err != nil {
		return "", fmt.Errorf("llm chat: %w", err)
	}

	// Clean mixed-language artifacts from model output
	response = postprocess.Clean(response)

	a.memory.Add("user", userMsg)
	a.memory.Add("assistant", response)

	return response, nil
}

// ChatStream runs the full RAG pipeline and streams tokens via onToken callback.
func (a *Agent) ChatStream(userMsg string, onToken func(token string) error) error {
	queryVec, err := a.embedder.Embed(userMsg)
	if err != nil {
		return fmt.Errorf("embed query: %w", err)
	}

	docs, _ := a.vectorstore.Search(a.cfg.CollectionName, queryVec, a.cfg.TopK)

	systemPrompt := basePrompt
	if ctx := buildContext(docs, userMsg); ctx != "" {
		systemPrompt = fmt.Sprintf(
			"%s\n\nUse the following context to answer. If not relevant, use your own knowledge.\n\nContext:\n%s",
			basePrompt, ctx,
		)
	}

	messages := []llm.Message{{Role: "system", Content: systemPrompt}}
	messages = append(messages, a.memory.Get()...)
	messages = append(messages, llm.Message{Role: "user", Content: userMsg})

	// Collect full response then clean before streaming to memory
	var fullResponse strings.Builder
	err = a.llm.ChatStream(messages, func(token string) error {
		fullResponse.WriteString(token)
		// Clean each token before sending to browser
		cleaned := postprocess.Clean(token)
		return onToken(cleaned)
	})
	if err != nil {
		return fmt.Errorf("llm stream: %w", err)
	}

	cleaned := postprocess.Clean(fullResponse.String())
	a.memory.Add("user", userMsg)
	a.memory.Add("assistant", cleaned)
	return nil
}
