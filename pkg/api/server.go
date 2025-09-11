// Package api FreyjaDB REST API
//
// @title           FreyjaDB REST API
// @version         1.0.0
// @description     This is the REST API for FreyjaDB, an embeddable key-value store.
// @host            localhost:9200
// @BasePath        /api/v1
//
// @securityDefinitions.apikey ApiKeyAuth
// @in              header
// @name            X-API-Key
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
	"github.com/swaggo/swag"
)

// StartServer starts the HTTP server with all routes configured
func StartServer(store *store.KVStore, config ServerConfig) error {
	// Set Swagger host with port
	if SwaggerInfo != nil {
		SwaggerInfo.Host = fmt.Sprintf("localhost:%d", config.Port)
	}

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
	r.Get("/swagger/*", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/swagger/" || path == "/swagger/index.html" {
			// Serve the Swagger UI HTML
			w.Header().Set("Content-Type", "text/html")
			html := `<!DOCTYPE html>
<html>
<head>
	 <title>FreyjaDB API Documentation</title>
	 <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@3.25.0/swagger-ui.css" />
</head>
<body>
	 <div id="swagger-ui"></div>
	 <script src="https://unpkg.com/swagger-ui-dist@3.25.0/swagger-ui-bundle.js"></script>
	 <script>
	   window.onload = function() {
	     SwaggerUIBundle({
	       url: '/swagger/swagger.json',
	       dom_id: '#swagger-ui',
	       presets: [
	         SwaggerUIBundle.presets.apis,
	         SwaggerUIBundle.presets.standalone
	       ]
	     });
	   };
	 </script>
</body>
</html>`
			w.Write([]byte(html))
			return
		}

		if path == "/swagger/swagger.json" {
			// Serve the dynamically generated Swagger JSON
			doc, err := swag.ReadDoc("swagger")
			if err != nil {
				fmt.Printf("Error generating swagger doc: %v\n", err)
				http.Error(w, "Failed to generate Swagger documentation", 500)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(doc))
			return
		}

		if path == "/swagger/swagger.yaml" {
			// Serve the dynamically generated Swagger YAML
			doc, err := swag.ReadDoc("swagger")
			if err != nil {
				fmt.Printf("Error generating swagger doc: %v\n", err)
				http.Error(w, "Failed to generate Swagger documentation", 500)
				return
			}
			w.Header().Set("Content-Type", "application/yaml")
			w.Write([]byte(doc)) // Note: This serves JSON as YAML, for true YAML conversion you'd need a JSON to YAML converter
			return
		}

		// For any other paths, return 404
		http.NotFound(w, r)
	})

	// Start background metrics updater
	go server.startMetricsUpdater()

	addr := fmt.Sprintf(":%d", config.Port)
	fmt.Printf("Starting FreyjaDB REST API server on %s\n", addr)
	fmt.Printf("Metrics available at: http://localhost:%d/metrics\n", config.Port)
	log.Fatal(http.ListenAndServe(addr, r))

	return nil
}
