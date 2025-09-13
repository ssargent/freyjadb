package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	statusSuccess = "success"
	statusError   = "error"
)

// Metrics holds all Prometheus metrics for the API
type Metrics struct {
	// HTTP request metrics
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight *prometheus.GaugeVec

	// Database operation metrics
	dbOperationsTotal   *prometheus.CounterVec
	dbOperationDuration *prometheus.HistogramVec
	dbKeysTotal         prometheus.Gauge
	dbDataSizeBytes     prometheus.Gauge

	// API key authentication metrics
	authRequestsTotal *prometheus.CounterVec

	// Relationship metrics
	relationshipOperationsTotal *prometheus.CounterVec

	// Health check metrics
	healthChecksTotal *prometheus.CounterVec
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics() *Metrics {
	m := &Metrics{
		// HTTP request metrics
		httpRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "freyja_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),

		httpRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "freyja_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),

		httpRequestsInFlight: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "freyja_http_requests_in_flight",
				Help: "Number of HTTP requests currently being processed",
			},
			[]string{"method", "endpoint"},
		),

		// Database operation metrics
		dbOperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "freyja_db_operations_total",
				Help: "Total number of database operations",
			},
			[]string{"operation", "status"},
		),

		dbOperationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "freyja_db_operation_duration_seconds",
				Help:    "Database operation duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation"},
		),

		dbKeysTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "freyja_db_keys_total",
				Help: "Total number of keys in the database",
			},
		),

		dbDataSizeBytes: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "freyja_db_data_size_bytes",
				Help: "Total size of data in the database in bytes",
			},
		),

		// Authentication metrics
		authRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "freyja_auth_requests_total",
				Help: "Total number of authentication requests",
			},
			[]string{"status"},
		),

		// Relationship metrics
		relationshipOperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "freyja_relationship_operations_total",
				Help: "Total number of relationship operations",
			},
			[]string{"operation", "status"},
		),

		// Health check metrics
		healthChecksTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "freyja_health_checks_total",
				Help: "Total number of health checks",
			},
			[]string{"status"},
		),
	}

	return m
}

// RecordHTTPRequest records an HTTP request
func (m *Metrics) RecordHTTPRequest(method, endpoint string, statusCode int, duration time.Duration) {
	statusCodeStr := strconv.Itoa(statusCode)

	m.httpRequestsTotal.WithLabelValues(method, endpoint, statusCodeStr).Inc()
	m.httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordDBOperation records a database operation
func (m *Metrics) RecordDBOperation(operation string, success bool, duration time.Duration) {
	status := statusSuccess
	if !success {
		status = statusError
	}

	m.dbOperationsTotal.WithLabelValues(operation, status).Inc()
	m.dbOperationDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// UpdateDBStats updates database statistics
func (m *Metrics) UpdateDBStats(keys int, dataSize int64) {
	m.dbKeysTotal.Set(float64(keys))
	m.dbDataSizeBytes.Set(float64(dataSize))
}

// RecordAuthRequest records an authentication request
func (m *Metrics) RecordAuthRequest(success bool) {
	status := statusSuccess
	if !success {
		status = statusError
	}
	m.authRequestsTotal.WithLabelValues(status).Inc()
}

// RecordRelationshipOperation records a relationship operation
func (m *Metrics) RecordRelationshipOperation(operation string, success bool) {
	status := statusSuccess
	if !success {
		status = statusError
	}
	m.relationshipOperationsTotal.WithLabelValues(operation, status).Inc()
}

// RecordHealthCheck records a health check
func (m *Metrics) RecordHealthCheck(success bool) {
	status := statusSuccess
	if !success {
		status = statusError
	}
	m.healthChecksTotal.WithLabelValues(status).Inc()
}

// InstrumentHandler instruments an HTTP handler with metrics
func (m *Metrics) InstrumentHandler(method, endpoint string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Record request in flight
		gauge := m.httpRequestsInFlight.WithLabelValues(method, endpoint)
		gauge.Inc()
		defer gauge.Dec()

		// Create response writer wrapper to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the original handler
		handler(rw, r)

		// Record metrics
		duration := time.Since(start)
		m.RecordHTTPRequest(method, endpoint, rw.statusCode, duration)
	}
}

// InstrumentAuthMiddleware instruments the authentication middleware
func (m *Metrics) InstrumentAuthMiddleware(next func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if API key is present
			apiKey := r.Header.Get("X-API-Key")
			hasAPIKey := apiKey != ""

			// Call the auth middleware
			next(h).ServeHTTP(w, r)

			// Record auth metrics based on response status
			if rw, ok := w.(*responseWriter); ok {
				success := rw.statusCode != http.StatusUnauthorized
				if hasAPIKey {
					m.RecordAuthRequest(success)
				}
			}
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
