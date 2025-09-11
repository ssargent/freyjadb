package api

import (
	"encoding/json"
	"net/http"
)

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

// sendSuccess sends a successful JSON response
func sendSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	response := APIResponse{
		Success: true,
		Data:    data,
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// sendError sends an error JSON response
func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := APIResponse{
		Success: false,
		Error:   message,
	}
	_ = json.NewEncoder(w).Encode(response)
}
