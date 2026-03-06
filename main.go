package main

import (
	"bufio"
	"embed"
	"io/fs"
	"log"
	"os"
	"strings"

	"go-ai-rag/internal/config"
	"go-ai-rag/internal/embedder"
	"go-ai-rag/internal/ingest"
	"go-ai-rag/internal/llm"
	"go-ai-rag/internal/memory"
	"go-ai-rag/internal/rag"
	"go-ai-rag/internal/server"
	"go-ai-rag/internal/vectorstore"
)

//go:embed static
var staticFiles embed.FS

// loadDotEnv reads .env file and sets environment variables (skips already-set ones).
func loadDotEnv() {
	f, err := os.Open(".env")
	if err != nil {
		return // .env not found — use system env vars
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}

func main() {
	loadDotEnv()

	cfg := config.Default()

	// Select LLM provider based on LLM_PROVIDER env var
	var chatLLM llm.LLM
	switch cfg.LLMProvider {
	case "gemini":
		if cfg.GeminiAPIKey == "" {
			log.Fatal("LLM_PROVIDER=gemini but GEMINI_API_KEY is not set in .env")
		}
		chatLLM = llm.NewGemini(cfg)
		log.Printf("LLM provider: Gemini (%s)", cfg.GeminiModel)
	default:
		chatLLM = llm.NewOllama(cfg)
		log.Printf("LLM provider: Ollama (%s @ %s)", cfg.OllamaModel, cfg.OllamaURL)
	}

	ollamaEmbed := embedder.NewOllama(cfg)
	qdrant := vectorstore.NewQdrant(cfg)
	mem := memory.NewSimple(10)
	ingestor := ingest.NewIngestor(ollamaEmbed, qdrant, cfg)
	agent := rag.NewAgent(chatLLM, ollamaEmbed, qdrant, mem, cfg)

	// Ensure Qdrant collection exists
	if err := qdrant.EnsureCollection(cfg.CollectionName, cfg.EmbedDim); err != nil {
		log.Printf("Warning: could not ensure Qdrant collection: %v", err)
	}

	// Serve static files from the embedded "static" subdirectory
	staticRoot, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("failed to get static sub-filesystem: %v", err)
	}

	// Start the web server
	srv := server.New(agent, ingestor, mem, cfg, staticRoot)
	if err := srv.Start(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
