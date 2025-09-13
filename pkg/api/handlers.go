package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ssargent/freyjadb/pkg/store"
)

// KeyValueResponse represents the response when including relationships
type KeyValueResponse struct {
	Value         interface{}                `json:"value"`
	ContentType   string                     `json:"content_type,omitempty"`
	Relationships []store.RelationshipResult `json:"relationships,omitempty"`
}

// Server holds the API server state
type Server struct {
	store         IKVStore
	systemService *SystemService
	config        ServerConfig
	metrics       *Metrics
}

// NewServer creates a new API server
func NewServer(store IKVStore, systemService *SystemService, config ServerConfig, metrics *Metrics) *Server {
	return &Server{
		store:         store,
		systemService: systemService,
		config:        config,
		metrics:       metrics,
	}
}

// handleHealth godoc
//
//	@Summary		Health check
//	@Description	Get the health status of the API
//	@Tags			health
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/health [get]
//	@Security		ApiKeyAuth
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.metrics.RecordHealthCheck(true)
	sendSuccess(w, map[string]string{"status": "healthy"})
}

// handlePut godoc
//
//	@Summary		Put a key-value pair
//	@Description	Store a key-value pair in the database
//	@Tags			kv
//	@Accept			octet-stream,json
//	@Produce		json
//	@Param			key		path		string				true	"Key"
//	@Param			body	body		[]byte				true	"Value"
//	@Param			Content-Type	header		string				false	"Content type (application/json or application/octet-stream)"
//	@Success		200		{object}	map[string]string
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Security		ApiKeyAuth
//	@Router			/kv/{key} [put]
func (s *Server) handlePut(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	key := chi.URLParam(r, "key")
	if key == "" {
		if s.metrics != nil {
			s.metrics.RecordDBOperation("put", false, time.Since(start))
		}
		sendError(w, "Key is required", http.StatusBadRequest)
		return
	}

	// Read the request body
	body := make([]byte, r.ContentLength)
	_, err := r.Body.Read(body)
	if err != nil && err.Error() != "EOF" {
		if s.metrics != nil {
			s.metrics.RecordDBOperation("put", false, time.Since(start))
		}
		sendError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Determine content type from header
	contentTypeHeader := r.Header.Get("Content-Type")
	contentType := getContentTypeFromHeader(contentTypeHeader)

	var dataToStore []byte

	// Handle JSON marshaling if content type is JSON
	if contentType == ContentTypeJSON {
		// Validate that the body is valid JSON
		var jsonData interface{}
		if err := json.Unmarshal(body, &jsonData); err != nil {
			if s.metrics != nil {
				s.metrics.RecordDBOperation("put", false, time.Since(start))
			}
			sendError(w, "Invalid JSON in request body", http.StatusBadRequest)
			return
		}
		// Re-marshal to ensure consistent formatting
		formattedJSON, err := json.Marshal(jsonData)
		if err != nil {
			if s.metrics != nil {
				s.metrics.RecordDBOperation("put", false, time.Since(start))
			}
			sendError(w, "Failed to format JSON", http.StatusInternalServerError)
			return
		}
		dataToStore = formattedJSON
	} else {
		dataToStore = body
	}

	// Encode data with content type metadata
	encodedData := encodeDataWithContentType(dataToStore, contentType)

	unescapedKey, err := url.QueryUnescape(chi.URLParam(r, "key"))
	if err != nil {
		if s.metrics != nil {
			s.metrics.RecordDBOperation("put", false, time.Since(start))
		}
		sendError(w, "Invalid key encoding", http.StatusBadRequest)
		return
	}
	if err := s.store.Put([]byte(unescapedKey), encodedData); err != nil {
		if s.metrics != nil {
			s.metrics.RecordDBOperation("put", false, time.Since(start))
		}
		sendError(w, fmt.Sprintf("Failed to put key-value: %v", err), http.StatusInternalServerError)
		return
	}

	if s.metrics != nil {
		s.metrics.RecordDBOperation("put", true, time.Since(start))
	}
	sendSuccess(w, map[string]string{"message": "Key-value pair stored successfully"})
}

// handleGet godoc
//
//	@Summary		Get a value by key
//	@Description	Retrieve the value for a given key. Use ?include=relationships to include relationship data.
//	@Tags			kv
//	@Accept			json
//	@Produce		octet-stream,json
//	@Param			key		path		string	true	"Key"
//	@Param			include	query		string	false	"Include additional data (relationships)"
//	@Success		200		{string}	byte
//	@Success		200		{object}	KeyValueResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/kv/{key} [get]
//	@Security		ApiKeyAuth
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	key := chi.URLParam(r, "key")
	if key == "" {
		s.metrics.RecordDBOperation("get", false, time.Since(start))
		sendError(w, "Key is required", http.StatusBadRequest)
		return
	}

	includeRelationships := r.URL.Query().Get("include") == "relationships"

	encodedValue, err := s.store.Get([]byte(key))
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

	// Decode the data and extract content type
	data, contentType := decodeDataWithContentType(encodedValue)

	s.metrics.RecordDBOperation("get", true, time.Since(start))

	if includeRelationships {
		// Fetch relationships
		query := store.RelationshipQuery{
			Key:       key,
			Direction: "both",
			Limit:     100, // Default limit
		}
		relationships, err := s.store.GetRelationships(query)
		if err != nil {
			sendError(w, fmt.Sprintf("Failed to get relationships: %v", err), http.StatusInternalServerError)
			return
		}

		// Prepare response
		response := KeyValueResponse{
			Relationships: relationships,
			ContentType:   getContentTypeHeader(contentType),
		}

		// Handle value based on content type
		if contentType == ContentTypeJSON {
			var jsonValue interface{}
			if err := json.Unmarshal(data, &jsonValue); err != nil {
				sendError(w, "Failed to parse JSON value", http.StatusInternalServerError)
				return
			}
			response.Value = jsonValue
		} else {
			response.Value = string(data)
		}

		w.Header().Set("Content-Type", "application/json")
		sendSuccess(w, response)
	} else {
		// Original behavior: return raw data
		contentTypeHeader := getContentTypeHeader(contentType)
		w.Header().Set("Content-Type", contentTypeHeader)
		if _, err := w.Write(data); err != nil {
			sendError(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	}
}

// handleDelete godoc
//
//	@Summary		Delete a key-value pair
//	@Description	Delete the key-value pair for a given key
//	@Tags			kv
//	@Accept			json
//	@Produce		json
//	@Param			key	path		string	true	"Key"
//	@Success		200	{object}	map[string]string
//	@Failure		400	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/kv/{key} [delete]
//	@Security		ApiKeyAuth
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

// handleListKeys godoc
//
//	@Summary		List keys
//	@Description	List all keys with optional prefix
//	@Tags			kv
//	@Accept			json
//	@Produce		json
//	@Param			prefix	query		string	false	"Key prefix"
//	@Success		200	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]string
//	@Router			/kv [get]
//	@Security		ApiKeyAuth
func (s *Server) handleListKeys(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")

	keys, err := s.store.ListKeys([]byte(prefix))
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to list keys: %v", err), http.StatusInternalServerError)
		return
	}

	sendSuccess(w, map[string]interface{}{"keys": keys})
}

// handleCreateRelationship godoc
//
//	@Summary		Create a relationship
//	@Description	Create a relationship between two keys
//	@Tags			relationships
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RelationshipRequest	true	"Relationship request"
//	@Success		200		{object}	map[string]string
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/relationships [post]
//	@Security		ApiKeyAuth
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

// handleDeleteRelationship godoc
//
//	@Summary		Delete a relationship
//	@Description	Delete a relationship between two keys
//	@Tags			relationships
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RelationshipRequest	true	"Relationship request"
//	@Success		200		{object}	map[string]string
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/relationships [delete]
//	@Security		ApiKeyAuth
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

// handleGetRelationships godoc
//
//	@Summary		Get relationships
//	@Description	Get relationships for a key with optional filters
//	@Tags			relationships
//	@Accept			json
//	@Produce		json
//	@Param			key			query		string	false	"Key to get relationships for"
//	@Param			direction	query		string	false	"Direction (both, incoming, outgoing)"
//	@Param			relation	query		string	false	"Relationship type filter"
//	@Param			limit		query		int		false	"Maximum number of results"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	map[string]string
//	@Failure		500			{object}	map[string]string
//	@Router			/relationships [get]
//	@Security		ApiKeyAuth
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

// handleExplain godoc
//
//	@Summary		Get database explain information
//	@Description	Get detailed information about database structure and performance
//	@Tags			diagnostics
//	@Accept			json
//	@Produce		json
//	@Param			pk	query		string	false	"Primary key to explain"
//	@Success		200	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]string
//	@Router			/explain [get]
//	@Security		ApiKeyAuth
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

// handleStats godoc
//
//	@Summary		Get database statistics
//	@Description	Get statistics about the database including key count and data size
//	@Tags			diagnostics
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]string
//	@Router			/stats [get]
//	@Security		ApiKeyAuth
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := s.store.Stats()
	// Update metrics with current stats
	s.metrics.UpdateDBStats(stats.Keys, stats.DataSize)
	sendSuccess(w, stats)
}

// Content type constants
const (
	ContentTypeRaw    = 0
	ContentTypeJSON   = 1
	ContentTypeHeader = 2 // Size of the header (type byte + null terminator)
)

// encodeDataWithContentType encodes data with content-type metadata
func encodeDataWithContentType(data []byte, contentType int) []byte {
	header := make([]byte, ContentTypeHeader)
	header[0] = byte(contentType)
	header[1] = 0 // null terminator

	return append(header, data...)
}

// decodeDataWithContentType decodes data and extracts content-type metadata
func decodeDataWithContentType(encodedData []byte) ([]byte, int) {
	if len(encodedData) < ContentTypeHeader {
		// No header present, treat as raw bytes (backward compatibility)
		return encodedData, ContentTypeRaw
	}

	contentType := int(encodedData[0])
	if encodedData[1] != 0 {
		// Invalid header format, treat as raw bytes
		return encodedData, ContentTypeRaw
	}

	data := encodedData[ContentTypeHeader:]
	return data, contentType
}

// getContentTypeFromHeader extracts content type from HTTP Content-Type header
func getContentTypeFromHeader(contentTypeHeader string) int {
	if strings.Contains(contentTypeHeader, "application/json") {
		return ContentTypeJSON
	}
	return ContentTypeRaw
}

// getContentTypeHeader returns the appropriate HTTP Content-Type header for a content type
func getContentTypeHeader(contentType int) string {
	switch contentType {
	case ContentTypeJSON:
		return "application/json"
	default:
		return "application/octet-stream"
	}
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

// System API handlers

// handleCreateAPIKey godoc
//
//	@Summary		Create a new API key
//	@Description	Create a new API key for user authentication
//	@Tags			system
//	@Accept			json
//	@Produce		json
//	@Param			request	body		APIKey					true	"API key details"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/system/api-keys [post]
//	@Security		ApiKeyAuth
func (s *Server) handleCreateAPIKey(w http.ResponseWriter, r *http.Request) {
	var apiKey APIKey
	if err := json.NewDecoder(r.Body).Decode(&apiKey); err != nil {
		sendError(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	if apiKey.ID == "" || apiKey.Key == "" {
		sendError(w, "id and key are required", http.StatusBadRequest)
		return
	}

	// Set creation time if not provided
	if apiKey.CreatedAt.IsZero() {
		apiKey.CreatedAt = time.Now()
	}

	// Set active if not specified
	if !apiKey.IsActive {
		apiKey.IsActive = true
	}

	if err := s.systemService.StoreAPIKey(apiKey); err != nil {
		sendError(w, fmt.Sprintf("Failed to create API key: %v", err), http.StatusInternalServerError)
		return
	}

	sendSuccess(w, map[string]interface{}{
		"message": "API key created successfully",
		"id":      apiKey.ID,
	})
}

// handleListAPIKeys godoc
//
//	@Summary		List all API keys
//	@Description	Get a list of all API key IDs
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]string
//	@Router			/system/api-keys [get]
//	@Security		ApiKeyAuth
func (s *Server) handleListAPIKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := s.systemService.ListAPIKeys()
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to list API keys: %v", err), http.StatusInternalServerError)
		return
	}

	sendSuccess(w, map[string]interface{}{"api_keys": keys})
}

// handleGetAPIKey godoc
//
//	@Summary		Get API key details
//	@Description	Get details of a specific API key
//	@Tags			system
//	@Produce		json
//	@Param			id	path		string	true	"API key ID"
//	@Success		200	{object}	APIKey
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/system/api-keys/{id} [get]
//	@Security		ApiKeyAuth
func (s *Server) handleGetAPIKey(w http.ResponseWriter, r *http.Request) {
	keyID := chi.URLParam(r, "id")
	if keyID == "" {
		sendError(w, "API key ID is required", http.StatusBadRequest)
		return
	}

	apiKey, err := s.systemService.GetAPIKey(keyID)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to get API key: %v", err), http.StatusInternalServerError)
		return
	}

	sendSuccess(w, apiKey)
}

// handleDeleteAPIKey godoc
//
//	@Summary		Delete an API key
//	@Description	Delete a specific API key
//	@Tags			system
//	@Produce		json
//	@Param			id	path		string	true	"API key ID"
//	@Success		200	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/system/api-keys/{id} [delete]
//	@Security		ApiKeyAuth
func (s *Server) handleDeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	keyID := chi.URLParam(r, "id")
	if keyID == "" {
		sendError(w, "API key ID is required", http.StatusBadRequest)
		return
	}

	if err := s.systemService.DeleteAPIKey(keyID); err != nil {
		sendError(w, fmt.Sprintf("Failed to delete API key: %v", err), http.StatusInternalServerError)
		return
	}

	sendSuccess(w, map[string]string{"message": "API key deleted successfully"})
}

// handleGetSystemConfig godoc
//
//	@Summary		Get system configuration
//	@Description	Get a system configuration value
//	@Tags			system
//	@Produce		json
//	@Param			key	path		string	true	"Configuration key"
//	@Success		200	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]string
//	@Router			/system/config/{key} [get]
//	@Security		ApiKeyAuth
func (s *Server) handleGetSystemConfig(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		sendError(w, "Configuration key is required", http.StatusBadRequest)
		return
	}

	var value interface{}
	if err := s.systemService.GetSystemConfig(key, &value); err != nil {
		sendError(w, fmt.Sprintf("Failed to get config: %v", err), http.StatusInternalServerError)
		return
	}

	sendSuccess(w, map[string]interface{}{"key": key, "value": value})
}

// handleSetSystemConfig godoc
//
//	@Summary		Set system configuration
//	@Description	Set a system configuration value
//	@Tags			system
//	@Accept			json
//	@Produce		json
//	@Param			key		path		string					true	"Configuration key"
//	@Param			value	body		interface{}			true	"Configuration value"
//	@Success		200		{object}	map[string]string
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/system/config/{key} [put]
//	@Security		ApiKeyAuth
func (s *Server) handleSetSystemConfig(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		sendError(w, "Configuration key is required", http.StatusBadRequest)
		return
	}

	var value interface{}
	if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
		sendError(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	if err := s.systemService.StoreSystemConfig(key, value); err != nil {
		sendError(w, fmt.Sprintf("Failed to set config: %v", err), http.StatusInternalServerError)
		return
	}

	sendSuccess(w, map[string]string{"message": "Configuration updated successfully"})
}
