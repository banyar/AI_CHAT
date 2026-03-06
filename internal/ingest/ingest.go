package ingest

import (
	"fmt"
	"time"

	"go-ai-rag/internal/config"
	"go-ai-rag/internal/embedder"
	"go-ai-rag/internal/vectorstore"
)

// Ingestor handles adding text documents into the vector store
type Ingestor struct {
	embedder    *embedder.Ollama
	vectorstore *vectorstore.Qdrant
	cfg         *config.Config
}

// NewIngestor creates a new Ingestor
func NewIngestor(e *embedder.Ollama, vs *vectorstore.Qdrant, cfg *config.Config) *Ingestor {
	return &Ingestor{embedder: e, vectorstore: vs, cfg: cfg}
}

// IngestText embeds and stores a text document in Qdrant
func (i *Ingestor) IngestText(text string) error {
	vec, err := i.embedder.Embed(text)
	if err != nil {
		return fmt.Errorf("embed text: %w", err)
	}

	doc := vectorstore.Document{
		ID:     fmt.Sprintf("%d", time.Now().UnixNano()),
		Text:   text,
		Vector: vec,
	}

	if err := i.vectorstore.Upsert(i.cfg.CollectionName, doc); err != nil {
		return fmt.Errorf("store document: %w", err)
	}

	return nil
}
