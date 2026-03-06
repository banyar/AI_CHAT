package vectorstore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go-ai-rag/internal/config"
)

// Document represents a text document with its embedding
type Document struct {
	ID     string
	Text   string
	Vector []float32
}

// Qdrant is the vector store client for Qdrant REST API
type Qdrant struct {
	cfg    *config.Config
	client *http.Client
}

// NewQdrant creates a new Qdrant client
func NewQdrant(cfg *config.Config) *Qdrant {
	return &Qdrant{cfg: cfg, client: &http.Client{}}
}

// EnsureCollection creates the collection if it does not already exist
func (q *Qdrant) EnsureCollection(name string, dim int) error {
	body, _ := json.Marshal(map[string]interface{}{
		"vectors": map[string]interface{}{
			"size":     dim,
			"distance": "Cosine",
		},
	})

	req, _ := http.NewRequest(http.MethodPut,
		q.cfg.QdrantURL+"/collections/"+name,
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := q.client.Do(req)
	if err != nil {
		return fmt.Errorf("qdrant create collection: %w", err)
	}
	defer resp.Body.Close()
	// 200 = created, 400 with "already exists" is also acceptable
	return nil
}

type point struct {
	ID      uint64                 `json:"id"`
	Vector  []float32              `json:"vector"`
	Payload map[string]interface{} `json:"payload"`
}

// Upsert inserts or updates a document in the collection
func (q *Qdrant) Upsert(collection string, doc Document) error {
	body, err := json.Marshal(map[string]interface{}{
		"points": []point{
			{
				ID:     fnv64(doc.ID),
				Vector: doc.Vector,
				Payload: map[string]interface{}{
					"text": doc.Text,
					"id":   doc.ID,
				},
			},
		},
	})
	if err != nil {
		return err
	}

	req, _ := http.NewRequest(http.MethodPut,
		q.cfg.QdrantURL+"/collections/"+collection+"/points",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := q.client.Do(req)
	if err != nil {
		return fmt.Errorf("qdrant upsert: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("qdrant upsert status %d: %s", resp.StatusCode, string(data))
	}
	return nil
}

type searchResponse struct {
	Result []struct {
		Score   float32                `json:"score"`
		Payload map[string]interface{} `json:"payload"`
	} `json:"result"`
}

// Search finds the topK most similar documents to the query vector
func (q *Qdrant) Search(collection string, vector []float32, topK int) ([]string, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"vector":       vector,
		"limit":        topK,
		"with_payload": true,
	})

	resp, err := q.client.Post(
		q.cfg.QdrantURL+"/collections/"+collection+"/points/search",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("qdrant search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("qdrant search status %d: %s", resp.StatusCode, string(data))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result searchResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("qdrant search parse: %w", err)
	}

	texts := make([]string, 0, len(result.Result))
	for _, r := range result.Result {
		if text, ok := r.Payload["text"].(string); ok {
			texts = append(texts, text)
		}
	}
	return texts, nil
}

// fnv64 is a simple FNV-1a hash to convert string ID → uint64 point ID
func fnv64(s string) uint64 {
	var h uint64 = 14695981039346656037
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 1099511628211
	}
	return h
}
