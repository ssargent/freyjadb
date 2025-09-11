/*
FreyjaDB REST API

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
package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/ssargent/freyjadb/pkg/store"
	httpSwagger "github.com/swaggo/http-swagger"
)

// StartServer starts the HTTP server with all routes configured
func StartServer(store *store.KVStore, config ServerConfig) error {
	server := NewServer(store, config)

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
	r.Use(apiKeyMiddleware(config.APIKey))

	// Routes
	r.Route("/api/v1", func(r chi.Router) {
		// Health check
		r.Get("/health", server.handleHealth)

		// KV operations
		r.Put("/kv/{key}", server.handlePut)
		r.Get("/kv/{key}", server.handleGet)
		r.Delete("/kv/{key}", server.handleDelete)
		r.Get("/kv", server.handleListKeys)

		// Relationships
		r.Post("/relationships", server.handleCreateRelationship)
		r.Delete("/relationships", server.handleDeleteRelationship)
		r.Get("/relationships", server.handleGetRelationships)

		// Diagnostics
		r.Get("/explain", server.handleExplain)
		r.Get("/stats", server.handleStats)
	})

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://localhost:%d/swagger/doc.json", config.Port)),
	))

	addr := fmt.Sprintf(":%d", config.Port)
	fmt.Printf("Starting FreyjaDB REST API server on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, r))

	return nil
}
