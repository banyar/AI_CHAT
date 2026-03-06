package config

import "os"

// Config holds all service configuration
type Config struct {
	// LLM provider: "ollama" (default) or "gemini"
	LLMProvider string

	// Ollama settings
	OllamaURL   string
	OllamaModel string

	// Gemini settings
	GeminiAPIKey string
	GeminiModel  string

	// Embedding + vector store
	EmbedModel     string
	QdrantURL      string
	CollectionName string
	EmbedDim       int
	TopK           int
}

// Default returns a Config with sensible defaults.
// Values can be overridden via environment variables:
//
//	LLM_PROVIDER  – "ollama" (default) or "gemini"
//	OLLAMA_URL    – e.g. https://abc123.ngrok-free.app  (Google Colab)
//	OLLAMA_MODEL  – override the chat model
//	GEMINI_API_KEY – Google AI Studio API key
//	GEMINI_MODEL  – e.g. gemini-2.0-flash-lite
//	QDRANT_URL    – e.g. http://localhost:6333
func Default() *Config {
	return &Config{
		LLMProvider: getEnv("LLM_PROVIDER", "ollama"),

		OllamaURL: getEnv("OLLAMA_URL", "http://localhost:11434"),
		// Burmese-capable models (choose based on your RAM):
		// "qwen2.5:3b"                   → 2.0 GB  (8GB RAM PC)
		// "qwen2.5:7b"                   → 4.7 GB  (16GB RAM PC)
		// "qwen2.5:14b"                  → 9.0 GB  (32GB RAM PC)
		// "yxchia/seallms-v3-7b:Q2_K"   → 3.1 GB  (SEA specialized, CPU ok)
		// "yxchia/seallms-v3-7b:Q4_K_M" → 5.4 GB  (SEA specialized, GPU required)
		OllamaModel: getEnv("OLLAMA_MODEL", "frontiir-ai:latest"),

		GeminiAPIKey: getEnv("GEMINI_API_KEY", ""),
		GeminiModel:  getEnv("GEMINI_MODEL", "gemini-2.0-flash-lite"),

		EmbedModel:     "nomic-embed-text",
		QdrantURL:      getEnv("QDRANT_URL", "http://localhost:6333"),
		CollectionName: "documents",
		EmbedDim:       768,
		TopK:           3,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
