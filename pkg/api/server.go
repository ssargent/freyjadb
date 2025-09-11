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
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ssargent/freyjadb/pkg/store"
	httpSwagger "github.com/swaggo/http-swagger"
)

// StartServer starts the HTTP server with all routes configured
func StartServer(store *store.KVStore, config ServerConfig) error {
	// Initialize metrics
	metrics := NewMetrics()

	server := NewServer(store, config, metrics)

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

	// Prometheus metrics endpoint (unprotected for scraping)
	r.Handle("/metrics", promhttp.Handler())

	// API key authentication middleware for protected routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(metrics.InstrumentAuthMiddleware(apiKeyMiddleware(config.APIKey)))

		// Health check
		r.Get("/health", metrics.InstrumentHandler("GET", "/api/v1/health", server.handleHealth))

		// KV operations
		r.Put("/kv/{key}", metrics.InstrumentHandler("PUT", "/api/v1/kv/{key}", server.handlePut))
		r.Get("/kv/{key}", metrics.InstrumentHandler("GET", "/api/v1/kv/{key}", server.handleGet))
		r.Delete("/kv/{key}", metrics.InstrumentHandler("DELETE", "/api/v1/kv/{key}", server.handleDelete))
		r.Get("/kv", metrics.InstrumentHandler("GET", "/api/v1/kv", server.handleListKeys))

		// Relationships
		r.Post("/relationships", metrics.InstrumentHandler("POST", "/api/v1/relationships", server.handleCreateRelationship))
		r.Delete("/relationships", metrics.InstrumentHandler("DELETE", "/api/v1/relationships", server.handleDeleteRelationship))
		r.Get("/relationships", metrics.InstrumentHandler("GET", "/api/v1/relationships", server.handleGetRelationships))

		// Diagnostics
		r.Get("/explain", metrics.InstrumentHandler("GET", "/api/v1/explain", server.handleExplain))
		r.Get("/stats", metrics.InstrumentHandler("GET", "/api/v1/stats", server.handleStats))
	})

	// Swagger documentation (unprotected)
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://localhost:%d/swagger/doc.json", config.Port)),
	))

	// Start background metrics updater
	go server.startMetricsUpdater()

	addr := fmt.Sprintf(":%d", config.Port)
	fmt.Printf("Starting FreyjaDB REST API server on %s\n", addr)
	fmt.Printf("Metrics available at: http://localhost:%d/metrics\n", config.Port)
	log.Fatal(http.ListenAndServe(addr, r))

	return nil
}
