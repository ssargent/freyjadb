/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

# FreyjaDB REST API

This is the REST API for FreyjaDB, an embeddable key-value store.

Version: 1.0.0
Host: localhost:8080
BasePath: /api/v1

SecurityDefinitions:
  - ApiKeyAuth:
    type: apiKey
    in: header
    name: X-API-Key

swagger:meta
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/spf13/cobra"
	"github.com/ssargent/freyjadb/pkg/store"
	httpSwagger "github.com/swaggo/http-swagger"
)

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// RelationshipRequest represents a relationship creation/deletion request
type RelationshipRequest struct {
	FromKey  string `json:"from_key"`
	ToKey    string `json:"to_key"`
	Relation string `json:"relation"`
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the REST API server",
	Long: `Start the FreyjaDB REST API server with authentication.

Example:
  freyja serve --api-key=mysecretkey --port=8080`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")
		apiKey, _ := cmd.Flags().GetString("api-key")

		if apiKey == "" {
			fmt.Println("Error: --api-key is required")
			return
		}

		// Get store from context
		kv, ok := cmd.Context().Value("store").(*store.KVStore)
		if !ok {
			fmt.Printf("Error: store not found in context\n")
			return
		}

		// Start HTTP server
		startServer(port, apiKey, kv)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntP("port", "p", 8080, "Port to listen on")
	serveCmd.Flags().String("api-key", "", "API key for authentication (required)")
	serveCmd.MarkFlagRequired("api-key")
}

// startServer starts the HTTP server with all routes
func startServer(port int, apiKey string, kv *store.KVStore) {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// API key authentication middleware
	r.Use(apiKeyMiddleware(apiKey))

	// Routes
	r.Route("/api/v1", func(r chi.Router) {
		// Health check
		r.Get("/health", handleHealth)

		// KV operations
		r.Put("/kv/{key}", handlePut(kv))
		r.Get("/kv/{key}", handleGet(kv))
		r.Delete("/kv/{key}", handleDelete(kv))
		r.Get("/kv", handleListKeys(kv))

		// Relationships
		r.Post("/relationships", handleCreateRelationship(kv))
		r.Delete("/relationships", handleDeleteRelationship(kv))
		r.Get("/relationships", handleGetRelationships(kv))

		// Diagnostics
		r.Get("/explain", handleExplain(kv))
		r.Get("/stats", handleStats(kv))
	})

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Starting FreyjaDB REST API server on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

// apiKeyMiddleware validates the X-API-Key header
func apiKeyMiddleware(expectedKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				sendError(w, "Missing X-API-Key header", http.StatusUnauthorized)
				return
			}
			if apiKey != expectedKey {
				sendError(w, "Invalid API key", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// handleHealth godoc
// @Summary Health check
// @Description Get the health status of the FreyjaDB server
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=map[string]string}
// @Security ApiKeyAuth
// @Router /health [get]
func handleHealth(w http.ResponseWriter, r *http.Request) {
	sendSuccess(w, map[string]string{"status": "healthy"})
}

// handlePut godoc
// @Summary Store a key-value pair
// @Description Store a key-value pair in the database
// @Tags kv
// @Accept plain
// @Produce json
// @Param key path string true "Key"
// @Param value body string true "Value"
// @Success 200 {object} APIResponse{data=map[string]string}
// @Failure 400 {object} APIResponse
// @Failure 401 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security ApiKeyAuth
// @Router /kv/{key} [put]
func handlePut(kv *store.KVStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "key")
		if key == "" {
			sendError(w, "Key is required", http.StatusBadRequest)
			return
		}

		body := make([]byte, r.ContentLength)
		_, err := r.Body.Read(body)
		if err != nil && err.Error() != "EOF" {
			sendError(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		if err := kv.Put([]byte(key), body); err != nil {
			sendError(w, fmt.Sprintf("Failed to put key-value: %v", err), http.StatusInternalServerError)
			return
		}

		sendSuccess(w, map[string]string{"message": "Key-value pair stored successfully"})
	}
}

// handleGet godoc
// @Summary Get a value by key
// @Description Retrieve a value for a given key
// @Tags kv
// @Accept json
// @Produce octet-stream
// @Param key path string true "Key"
// @Success 200 {string} string "Value"
// @Failure 400 {object} APIResponse
// @Failure 401 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security ApiKeyAuth
// @Router /kv/{key} [get]
func handleGet(kv *store.KVStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "key")
		if key == "" {
			sendError(w, "Key is required", http.StatusBadRequest)
			return
		}

		value, err := kv.Get([]byte(key))
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				sendError(w, "Key not found", http.StatusNotFound)
			} else {
				sendError(w, fmt.Sprintf("Failed to get value: %v", err), http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(value)
	}
}

func handleDelete(kv *store.KVStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "key")
		if key == "" {
			sendError(w, "Key is required", http.StatusBadRequest)
			return
		}

		if err := kv.Delete([]byte(key)); err != nil {
			sendError(w, fmt.Sprintf("Failed to delete key: %v", err), http.StatusInternalServerError)
			return
		}

		sendSuccess(w, map[string]string{"message": "Key deleted successfully"})
	}
}

func handleListKeys(kv *store.KVStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		prefix := r.URL.Query().Get("prefix")

		keys, err := kv.ListKeys([]byte(prefix))
		if err != nil {
			sendError(w, fmt.Sprintf("Failed to list keys: %v", err), http.StatusInternalServerError)
			return
		}

		sendSuccess(w, map[string]interface{}{"keys": keys})
	}
}

// Relationship handlers
func handleCreateRelationship(kv *store.KVStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RelationshipRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, "Invalid JSON request", http.StatusBadRequest)
			return
		}

		if req.FromKey == "" || req.ToKey == "" || req.Relation == "" {
			sendError(w, "from_key, to_key, and relation are required", http.StatusBadRequest)
			return
		}

		if err := kv.PutRelationship(req.FromKey, req.ToKey, req.Relation); err != nil {
			sendError(w, fmt.Sprintf("Failed to create relationship: %v", err), http.StatusInternalServerError)
			return
		}

		sendSuccess(w, map[string]string{"message": "Relationship created successfully"})
	}
}

func handleDeleteRelationship(kv *store.KVStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RelationshipRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, "Invalid JSON request", http.StatusBadRequest)
			return
		}

		if req.FromKey == "" || req.ToKey == "" || req.Relation == "" {
			sendError(w, "from_key, to_key, and relation are required", http.StatusBadRequest)
			return
		}

		if err := kv.DeleteRelationship(req.FromKey, req.ToKey, req.Relation); err != nil {
			sendError(w, fmt.Sprintf("Failed to delete relationship: %v", err), http.StatusInternalServerError)
			return
		}

		sendSuccess(w, map[string]string{"message": "Relationship deleted successfully"})
	}
}

func handleGetRelationships(kv *store.KVStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		results, err := kv.GetRelationships(query)
		if err != nil {
			sendError(w, fmt.Sprintf("Failed to get relationships: %v", err), http.StatusInternalServerError)
			return
		}

		sendSuccess(w, map[string]interface{}{"relationships": results})
	}
}

// Diagnostic handlers
func handleExplain(kv *store.KVStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		opts := store.ExplainOptions{
			WithSamples: 10,
			WithMetrics: true,
		}

		if pk := r.URL.Query().Get("pk"); pk != "" {
			opts.PK = pk
		}

		result, err := kv.Explain(r.Context(), opts)
		if err != nil {
			sendError(w, fmt.Sprintf("Failed to get explain data: %v", err), http.StatusInternalServerError)
			return
		}

		sendSuccess(w, result)
	}
}

func handleStats(kv *store.KVStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := kv.Stats()
		sendSuccess(w, stats)
	}
}

// Helper functions
func sendSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	response := APIResponse{
		Success: true,
		Data:    data,
	}
	json.NewEncoder(w).Encode(response)
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := APIResponse{
		Success: false,
		Error:   message,
	}
	json.NewEncoder(w).Encode(response)
}
