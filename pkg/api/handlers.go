package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ssargent/freyjadb/pkg/store"
)

// Server holds the API server state
type Server struct {
	store   *store.KVStore
	config  ServerConfig
	metrics *Metrics
}

// NewServer creates a new API server
func NewServer(store *store.KVStore, config ServerConfig, metrics *Metrics) *Server {
	return &Server{
		store:   store,
		config:  config,
		metrics: metrics,
	}
}

// Health check handler
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.metrics.RecordHealthCheck(true)
	sendSuccess(w, map[string]string{"status": "healthy"})
}

// KV handlers
func (s *Server) handlePut(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	key := chi.URLParam(r, "key")
	if key == "" {
		s.metrics.RecordDBOperation("put", false, time.Since(start))
		sendError(w, "Key is required", http.StatusBadRequest)
		return
	}

	body := make([]byte, r.ContentLength)
	_, err := r.Body.Read(body)
	if err != nil && err.Error() != "EOF" {
		s.metrics.RecordDBOperation("put", false, time.Since(start))
		sendError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	if err := s.store.Put([]byte(key), body); err != nil {
		s.metrics.RecordDBOperation("put", false, time.Since(start))
		sendError(w, fmt.Sprintf("Failed to put key-value: %v", err), http.StatusInternalServerError)
		return
	}

	s.metrics.RecordDBOperation("put", true, time.Since(start))
	sendSuccess(w, map[string]string{"message": "Key-value pair stored successfully"})
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	key := chi.URLParam(r, "key")
	if key == "" {
		s.metrics.RecordDBOperation("get", false, time.Since(start))
		sendError(w, "Key is required", http.StatusBadRequest)
		return
	}

	value, err := s.store.Get([]byte(key))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.metrics.RecordDBOperation("get", false, time.Since(start))
			sendError(w, "Key not found", http.StatusNotFound)
		} else {
			s.metrics.RecordDBOperation("get", false, time.Since(start))
			sendError(w, fmt.Sprintf("Failed to get value: %v", err), http.StatusInternalServerError)
		}
		return
	}

	s.metrics.RecordDBOperation("get", true, time.Since(start))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(value)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	key := chi.URLParam(r, "key")
	if key == "" {
		s.metrics.RecordDBOperation("delete", false, time.Since(start))
		sendError(w, "Key is required", http.StatusBadRequest)
		return
	}

	if err := s.store.Delete([]byte(key)); err != nil {
		s.metrics.RecordDBOperation("delete", false, time.Since(start))
		sendError(w, fmt.Sprintf("Failed to delete key: %v", err), http.StatusInternalServerError)
		return
	}

	s.metrics.RecordDBOperation("delete", true, time.Since(start))
	sendSuccess(w, map[string]string{"message": "Key deleted successfully"})
}

func (s *Server) handleListKeys(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")

	keys, err := s.store.ListKeys([]byte(prefix))
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to list keys: %v", err), http.StatusInternalServerError)
		return
	}

	sendSuccess(w, map[string]interface{}{"keys": keys})
}

// Relationship handlers
func (s *Server) handleCreateRelationship(w http.ResponseWriter, r *http.Request) {
	var req RelationshipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.metrics.RecordRelationshipOperation("create", false)
		sendError(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	if req.FromKey == "" || req.ToKey == "" || req.Relation == "" {
		s.metrics.RecordRelationshipOperation("create", false)
		sendError(w, "from_key, to_key, and relation are required", http.StatusBadRequest)
		return
	}

	if err := s.store.PutRelationship(req.FromKey, req.ToKey, req.Relation); err != nil {
		s.metrics.RecordRelationshipOperation("create", false)
		sendError(w, fmt.Sprintf("Failed to create relationship: %v", err), http.StatusInternalServerError)
		return
	}

	s.metrics.RecordRelationshipOperation("create", true)
	sendSuccess(w, map[string]string{"message": "Relationship created successfully"})
}

func (s *Server) handleDeleteRelationship(w http.ResponseWriter, r *http.Request) {
	var req RelationshipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	if req.FromKey == "" || req.ToKey == "" || req.Relation == "" {
		sendError(w, "from_key, to_key, and relation are required", http.StatusBadRequest)
		return
	}

	if err := s.store.DeleteRelationship(req.FromKey, req.ToKey, req.Relation); err != nil {
		sendError(w, fmt.Sprintf("Failed to delete relationship: %v", err), http.StatusInternalServerError)
		return
	}

	sendSuccess(w, map[string]string{"message": "Relationship deleted successfully"})
}

func (s *Server) handleGetRelationships(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	direction := r.URL.Query().Get("direction")
	relation := r.URL.Query().Get("relation")
	limitStr := r.URL.Query().Get("limit")

	if key == "" {
		sendError(w, "key parameter is required", http.StatusBadRequest)
		return
	}

	if direction == "" {
		direction = "both"
	}

	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	query := store.RelationshipQuery{
		Key:       key,
		Direction: direction,
		Relation:  relation,
		Limit:     limit,
	}

	results, err := s.store.GetRelationships(query)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to get relationships: %v", err), http.StatusInternalServerError)
		return
	}

	sendSuccess(w, map[string]interface{}{"relationships": results})
}

// Diagnostic handlers
func (s *Server) handleExplain(w http.ResponseWriter, r *http.Request) {
	opts := store.ExplainOptions{
		WithSamples: 10,
		WithMetrics: true,
	}

	if pk := r.URL.Query().Get("pk"); pk != "" {
		opts.PK = pk
	}

	result, err := s.store.Explain(r.Context(), opts)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to get explain data: %v", err), http.StatusInternalServerError)
		return
	}

	sendSuccess(w, result)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := s.store.Stats()
	// Update metrics with current stats
	s.metrics.UpdateDBStats(stats.Keys, stats.DataSize)
	sendSuccess(w, stats)
}

// startMetricsUpdater periodically updates database metrics
func (s *Server) startMetricsUpdater() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := s.store.Stats()
		s.metrics.UpdateDBStats(stats.Keys, stats.DataSize)
	}
}
