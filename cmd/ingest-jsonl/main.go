// cmd/ingest-jsonl/main.go
// Reads train.jsonl and stores all Q&A pairs into Qdrant vector database.
// Run this ONCE after starting Qdrant to populate the knowledge base.
//
// Usage:
//
//	go run ./cmd/ingest-jsonl
//	go run ./cmd/ingest-jsonl --file=train.jsonl
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"go-ai-rag/internal/config"
	"go-ai-rag/internal/embedder"
	"go-ai-rag/internal/ingest"
	"go-ai-rag/internal/vectorstore"
)

type trainEntry struct {
	Instruction string `json:"instruction"`
	Input       string `json:"input"`
	Output      string `json:"output"`
}

func main() {
	filePath := flag.String("file", "train.jsonl", "Path to train.jsonl file")
	flag.Parse()

	cfg := config.Default()

	fmt.Println("======================================")
	fmt.Println("  Frontiir AI — Ingest train.jsonl")
	fmt.Println("======================================")
	fmt.Printf("  File    : %s\n", *filePath)
	fmt.Printf("  Qdrant  : %s\n", cfg.QdrantURL)
	fmt.Printf("  Embed   : %s\n", cfg.EmbedModel)
	fmt.Printf("  Collection: %s\n", cfg.CollectionName)
	fmt.Println("======================================")
	fmt.Println()

	// Init components
	ollamaEmbed := embedder.NewOllama(cfg)
	qdrant := vectorstore.NewQdrant(cfg)
	ingestor := ingest.NewIngestor(ollamaEmbed, qdrant, cfg)

	// Ensure Qdrant collection exists
	if err := qdrant.EnsureCollection(cfg.CollectionName, cfg.EmbedDim); err != nil {
		log.Fatalf("❌ Cannot connect to Qdrant at %s\n   Make sure Qdrant is running: docker run -p 6333:6333 qdrant/qdrant\n   Error: %v", cfg.QdrantURL, err)
	}
	fmt.Println("✅ Qdrant collection ready")
	fmt.Println()

	// Open train.jsonl
	f, err := os.Open(*filePath)
	if err != nil {
		log.Fatalf("❌ Cannot open %s: %v", *filePath, err)
	}
	defer f.Close()

	var total, ingested, skipped int
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		total++

		var entry trainEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			log.Printf("  ⚠️  Skip line %d (invalid JSON): %v", total, err)
			skipped++
			continue
		}

		if entry.Instruction == "" || entry.Output == "" {
			skipped++
			continue
		}

		// Format as searchable Q&A text
		text := fmt.Sprintf("Q: %s\nA: %s", entry.Instruction, entry.Output)
		if entry.Input != "" {
			text = fmt.Sprintf("Q: %s %s\nA: %s", entry.Instruction, entry.Input, entry.Output)
		}

		if err := ingestor.IngestText(text); err != nil {
			log.Printf("  ❌ Ingest error (line %d): %v", total, err)
			skipped++
			continue
		}

		ingested++
		fmt.Printf("\r  Ingesting... %d/%d  ✅", ingested, total)

		// Small delay so Ollama embed model is not overwhelmed
		time.Sleep(50 * time.Millisecond)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("⚠️  Scanner error: %v", err)
	}

	fmt.Println()
	fmt.Println()
	fmt.Println("======================================")
	fmt.Println("  Done!")
	fmt.Printf("  Total entries : %d\n", total)
	fmt.Printf("  Ingested      : %d ✅\n", ingested)
	if skipped > 0 {
		fmt.Printf("  Skipped       : %d ⚠️\n", skipped)
	}
	fmt.Println("======================================")
	fmt.Println()
	fmt.Println("Now run the app:")
	fmt.Println("  go run .")
}
