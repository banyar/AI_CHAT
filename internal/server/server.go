package server

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"go-ai-rag/internal/config"
	"go-ai-rag/internal/guard"
	"go-ai-rag/internal/ingest"
	"go-ai-rag/internal/memory"
	"go-ai-rag/internal/rag"
)

// Server holds the HTTP server dependencies
type Server struct {
	agent    *rag.Agent
	ingestor *ingest.Ingestor
	memory   *memory.Simple
	cfg      *config.Config
	mux      *http.ServeMux
}

// New creates and configures a new Server
func New(agent *rag.Agent, ingestor *ingest.Ingestor, mem *memory.Simple, cfg *config.Config, staticFiles fs.FS) *Server {
	s := &Server{
		agent:    agent,
		ingestor: ingestor,
		memory:   mem,
		cfg:      cfg,
		mux:      http.NewServeMux(),
	}
	s.registerRoutes(staticFiles)
	return s
}

func (s *Server) registerRoutes(staticFiles fs.FS) {
	// Serve static files with no-cache headers so browser always gets latest HTML/JS
	s.mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		http.FileServer(http.FS(staticFiles)).ServeHTTP(w, r)
	}))

	// JSON API
	s.mux.HandleFunc("/api/chat", s.handleChat)
	s.mux.HandleFunc("/api/ingest", s.handleIngest)
	s.mux.HandleFunc("/api/clear", s.handleClear)
	s.mux.HandleFunc("/api/status", s.handleStatus)
}

// ServeHTTP implements http.Handler so Server can be passed to http.ListenAndServe
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// Start begins listening on addr (e.g. ":8080")
func (s *Server) Start(addr string) error {
	log.Printf("Web UI running at http://localhost%s", addr)
	return http.ListenAndServe(addr, s)
}

// ── handlers ─────────────────────────────────────────────────────────────────

type chatRequest struct {
	Message string `json:"message"`
}

type chatResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Message == "" {
		writeJSON(w, http.StatusBadRequest, chatResponse{Error: "invalid request body"})
		return
	}

	// Hard block — very rude messages never reach the LLM.
	if guard.IsDenied(req.Message) {
		writeJSON(w, http.StatusOK, chatResponse{Response: guard.DeniedResponse(req.Message)})
		return
	}

	// Soft warn — prepend a polite warning, then still answer.
	warnPrefix := ""
	if guard.IsWarned(req.Message) {
		warnPrefix = guard.WarnPrefix(req.Message)
	}

	// Use SSE (Server-Sent Events) to stream tokens to the browser.
	// This prevents ngrok/proxy timeouts for slow models.
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, chatResponse{Error: "streaming not supported"})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	enc := json.NewEncoder(w)

	sendToken := func(token string) error {
		fmt.Fprint(w, "data: ")
		if err := enc.Encode(map[string]string{"token": token}); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}

	// Send warning prefix before LLM response if needed.
	if warnPrefix != "" {
		sendToken(warnPrefix)
	}

	err := s.agent.ChatStream(req.Message, sendToken)

	if err != nil {
		fmt.Fprint(w, "data: ")
		enc.Encode(map[string]string{"error": err.Error()})
		flusher.Flush()
		return
	}

	// Signal end of stream
	fmt.Fprint(w, "data: ")
	enc.Encode(map[string]string{"done": "true"})
	flusher.Flush()
}

type ingestRequest struct {
	Text string `json:"text"`
}

type ingestResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

func (s *Server) handleIngest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ingestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Text == "" {
		writeJSON(w, http.StatusBadRequest, ingestResponse{Error: "invalid request body"})
		return
	}

	if err := s.ingestor.IngestText(req.Text); err != nil {
		writeJSON(w, http.StatusInternalServerError, ingestResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, ingestResponse{Message: "Document added to knowledge base."})
}

func (s *Server) handleClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.memory.Clear()
	writeJSON(w, http.StatusOK, map[string]string{"message": "Conversation history cleared."})
}

type statusResponse struct {
	OllamaURL   string `json:"ollama_url"`
	OllamaModel string `json:"ollama_model"`
	EmbedModel  string `json:"embed_model"`
	QdrantURL   string `json:"qdrant_url"`
	IsColab     bool   `json:"is_colab"`
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	isColab := s.cfg.OllamaURL != "http://localhost:11434"
	writeJSON(w, http.StatusOK, statusResponse{
		OllamaURL:   s.cfg.OllamaURL,
		OllamaModel: s.cfg.OllamaModel,
		EmbedModel:  s.cfg.EmbedModel,
		QdrantURL:   s.cfg.QdrantURL,
		IsColab:     isColab,
	})
}

// ── helpers ───────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
