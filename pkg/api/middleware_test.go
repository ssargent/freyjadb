package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIKeyMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		apiKey         string
		requestHeader  string
		expectedStatus int
	}{
		{
			name:           "valid API key",
			apiKey:         "test-key",
			requestHeader:  "test-key",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing API key header",
			apiKey:         "test-key",
			requestHeader:  "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid API key",
			apiKey:         "test-key",
			requestHeader:  "wrong-key",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "empty API key in request",
			apiKey:         "test-key",
			requestHeader:  "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler that just returns 200
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Apply the middleware
			middleware := apiKeyMiddleware(tt.apiKey)
			handler := middleware(testHandler)

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.requestHeader != "" {
				req.Header.Set("X-API-Key", tt.requestHeader)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request
			handler.ServeHTTP(w, req)

			// Check status
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestSendSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"message": "test"}

	sendSuccess(w, data)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	// Check that response contains expected data
	body := w.Body.String()
	if len(body) == 0 {
		t.Error("Expected non-empty response body")
	}
}

func TestSendError(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		statusCode     int
		expectedStatus int
	}{
		{
			name:           "bad request error",
			message:        "Invalid request",
			statusCode:     http.StatusBadRequest,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized error",
			message:        "Not authorized",
			statusCode:     http.StatusUnauthorized,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "internal server error",
			message:        "Server error",
			statusCode:     http.StatusInternalServerError,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			sendError(w, tt.message, tt.statusCode)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", contentType)
			}

			// Check that response contains error message
			body := w.Body.String()
			if len(body) == 0 {
				t.Error("Expected non-empty response body")
			}
		})
	}
}
