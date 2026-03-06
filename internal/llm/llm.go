// internal/llm/llm.go
// LLM is the common interface for all chat providers (Ollama, Gemini, etc.)
package llm

// LLM defines the contract every chat provider must satisfy.
type LLM interface {
	Chat(messages []Message) (string, error)
	ChatStream(messages []Message, onToken func(token string) error) error
}
