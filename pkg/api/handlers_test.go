package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/ssargent/freyjadb/pkg/store"
)

func setupTestServer(t *testing.T) (*Server, func()) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "freyja_api_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create KV store
	config := store.KVStoreConfig{
		DataDir:       tmpDir,
		FsyncInterval: 0,
	}

	kvStore, err := store.NewKVStore(config)
	if err != nil {
		t.Fatalf("Failed to create KV store: %v", err)
	}

	_, err = kvStore.Open()
	if err != nil {
		t.Fatalf("Failed to open KV store: %v", err)
	}

	// Create metrics
	metrics := NewMetrics()

	// Create API server
	server := NewServer(kvStore, ServerConfig{}, metrics)

	// Cleanup function
	cleanup := func() {
		kvStore.Close()
		os.RemoveAll(tmpDir)
	}

	return server, cleanup
}

func TestServer_handleHealth(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response APIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}

	if response.Data == nil {
		t.Error("Expected data to be present")
	}
}

func TestServer_handlePut(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	tests := []struct {
		name           string
		key            string
		value          string
		expectedStatus int
	}{
		{
			name:           "valid put",
			key:            "testkey",
			value:          "testvalue",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "empty key",
			key:            "",
			value:          "testvalue",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty value",
			key:            "testkey",
			value:          "",
			expectedStatus: http.StatusOK, // Empty values are allowed (tombstones)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with URL param
			req := httptest.NewRequest("PUT", "/kv/"+tt.key, strings.NewReader(tt.value))
			req.Header.Set("Content-Type", "text/plain")

			// Set up chi context for URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("key", tt.key)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()

			handler := server.handlePut
			handler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response APIResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if !response.Success {
					t.Error("Expected success to be true")
				}
			}
		})
	}
}

func TestServer_handleGet(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// First put a value
	key := "testkey"
	value := "testvalue"
	if err := server.store.Put([]byte(key), []byte(value)); err != nil {
		t.Fatalf("Failed to put test data: %v", err)
	}

	tests := []struct {
		name           string
		key            string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "existing key",
			key:            "testkey",
			expectedStatus: http.StatusOK,
			expectedBody:   "testvalue",
		},
		{
			name:           "non-existing key",
			key:            "nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
		},
		{
			name:           "empty key",
			key:            "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/kv/"+tt.key, nil)

			// Set up chi context for URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("key", tt.key)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()

			handler := server.handleGet
			handler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				body := w.Body.String()
				if body != tt.expectedBody {
					t.Errorf("Expected body %q, got %q", tt.expectedBody, body)
				}

				contentType := w.Header().Get("Content-Type")
				if contentType != "application/octet-stream" {
					t.Errorf("Expected Content-Type application/octet-stream, got %s", contentType)
				}
			}
		})
	}
}

func TestServer_handleDelete(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// First put a value
	key := "testkey"
	value := "testvalue"
	if err := server.store.Put([]byte(key), []byte(value)); err != nil {
		t.Fatalf("Failed to put test data: %v", err)
	}

	tests := []struct {
		name           string
		key            string
		expectedStatus int
	}{
		{
			name:           "existing key",
			key:            "testkey",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existing key",
			key:            "nonexistent",
			expectedStatus: http.StatusOK, // Delete is idempotent
		},
		{
			name:           "empty key",
			key:            "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/kv/"+tt.key, nil)

			// Set up chi context for URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("key", tt.key)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()

			handler := server.handleDelete
			handler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response APIResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if !response.Success {
					t.Error("Expected success to be true")
				}
			}
		})
	}
}

func TestServer_handleListKeys(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Put some test data
	testData := map[string]string{
		"user:1": "John",
		"user:2": "Jane",
		"item:1": "Laptop",
		"item:2": "Phone",
	}

	for key, value := range testData {
		if err := server.store.Put([]byte(key), []byte(value)); err != nil {
			t.Fatalf("Failed to put test data: %v", err)
		}
	}

	tests := []struct {
		name           string
		prefix         string
		expectedCount  int
		expectedStatus int
	}{
		{
			name:           "all keys",
			prefix:         "",
			expectedCount:  4,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "user prefix",
			prefix:         "user",
			expectedCount:  2,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "item prefix",
			prefix:         "item",
			expectedCount:  2,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existing prefix",
			prefix:         "nonexistent",
			expectedCount:  0,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/kv?prefix="+tt.prefix, nil)
			w := httptest.NewRecorder()

			handler := server.handleListKeys
			handler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response APIResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if !response.Success {
					t.Error("Expected success to be true")
				}

				data, ok := response.Data.(map[string]interface{})
				if !ok {
					t.Fatal("Expected data to be a map")
				}

				// Handle the case where keys might be nil or empty
				if keysData, exists := data["keys"]; exists {
					if keys, ok := keysData.([]interface{}); ok {
						if len(keys) != tt.expectedCount {
							t.Errorf("Expected %d keys, got %d", tt.expectedCount, len(keys))
						}
					} else {
						// If it's not an array, it might be nil or another type
						if tt.expectedCount != 0 {
							t.Errorf("Expected %d keys, but keys field is not an array", tt.expectedCount)
						}
					}
				} else if tt.expectedCount != 0 {
					t.Errorf("Expected %d keys, but keys field is missing", tt.expectedCount)
				}
			}
		})
	}
}

func TestServer_handleCreateRelationship(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Create test entities first
	if err := server.store.Put([]byte("user:1"), []byte("John")); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	if err := server.store.Put([]byte("item:1"), []byte("Laptop")); err != nil {
		t.Fatalf("Failed to create test item: %v", err)
	}

	tests := []struct {
		name           string
		request        RelationshipRequest
		expectedStatus int
	}{
		{
			name: "valid relationship",
			request: RelationshipRequest{
				FromKey:  "user:1",
				ToKey:    "item:1",
				Relation: "owns",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing from_key",
			request: RelationshipRequest{
				ToKey:    "item:1",
				Relation: "owns",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing to_key",
			request: RelationshipRequest{
				FromKey:  "user:1",
				Relation: "owns",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing relation",
			request: RelationshipRequest{
				FromKey: "user:1",
				ToKey:   "item:1",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "non-existent from_key",
			request: RelationshipRequest{
				FromKey:  "user:999",
				ToKey:    "item:1",
				Relation: "owns",
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/relationships", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			handler := server.handleCreateRelationship
			handler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestServer_handleStats(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()

	server.handleStats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response APIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}

	if response.Data == nil {
		t.Error("Expected data to be present")
	}
}
