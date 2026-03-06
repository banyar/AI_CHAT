package memory

import "go-ai-rag/internal/llm"

// Simple keeps the last N chat messages in memory
type Simple struct {
	messages []llm.Message
	maxLen   int
}

// NewSimple creates a new Simple memory with a maximum message count
func NewSimple(maxLen int) *Simple {
	return &Simple{maxLen: maxLen}
}

// Add appends a message and trims to maxLen
func (m *Simple) Add(role, content string) {
	m.messages = append(m.messages, llm.Message{Role: role, Content: content})
	if len(m.messages) > m.maxLen {
		m.messages = m.messages[len(m.messages)-m.maxLen:]
	}
}

// Get returns all stored messages
func (m *Simple) Get() []llm.Message {
	return m.messages
}

// Clear resets the conversation history
func (m *Simple) Clear() {
	m.messages = nil
}
