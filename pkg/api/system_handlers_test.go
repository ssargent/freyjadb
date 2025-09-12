package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/ssargent/freyjadb/pkg/store"
	"github.com/stretchr/testify/assert"
)

func setupSystemTestServer(t *testing.T) (*Server, func()) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "freyja_system_handlers_test")
	assert.NoError(t, err)

	// Create KV store
	config := store.KVStoreConfig{
		DataDir:       tmpDir,
		FsyncInterval: 0,
	}

	kvStore, err := store.NewKVStore(config)
	assert.NoError(t, err)

	_, err = kvStore.Open()
	assert.NoError(t, err)

	// Create system service
	systemConfig := SystemConfig{
		DataDir:          tmpDir,
		EncryptionKey:    "12345678901234567890123456789012",
		EnableEncryption: true,
	}

	systemService, err := NewSystemService(systemConfig)
	assert.NoError(t, err)

	err = systemService.Open()
	assert.NoError(t, err)

	// Create server
	serverConfig := ServerConfig{
		Port:                8080,
		APIKey:              "test-user-key",
		SystemKey:           "test-system-key",
		DataDir:             tmpDir,
		SystemDataDir:       tmpDir,
		SystemEncryptionKey: "12345678901234567890123456789012",
		EnableEncryption:    true,
	}

	server := NewServer(kvStore, systemService, serverConfig, &Metrics{})

	// Return cleanup function
	cleanup := func() {
		// Close in reverse order to avoid file handle issues
		kvStore.Close()
		systemService.Close()
		os.RemoveAll(tmpDir)
	}

	return server, cleanup
}

func TestSystemAPIKeyHandlers(t *testing.T) {
	server, cleanup := setupSystemTestServer(t)
	defer cleanup()

	t.Run("Create API key", func(t *testing.T) {
		apiKeyData := APIKey{
			ID:          "test-api-key",
			Key:         "test-key-value",
			Description: "Test API key",
			IsActive:    true,
		}

		body, _ := json.Marshal(apiKeyData)
		req := httptest.NewRequest("POST", "/system/api-keys", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-system-key")

		w := httptest.NewRecorder()
		server.handleCreateAPIKey(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "API key created successfully", data["message"])
		assert.Equal(t, "test-api-key", data["id"])
	})

	t.Run("List API keys", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/system/api-keys", nil)
		req.Header.Set("X-API-Key", "test-system-key")

		w := httptest.NewRecorder()
		server.handleListAPIKeys(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		data := response["data"].(map[string]interface{})
		assert.Contains(t, data, "api_keys")
	})

	t.Run("Get specific API key", func(t *testing.T) {
		server, cleanup := setupSystemTestServer(t)
		defer cleanup()

		// First create the API key
		apiKeyData := APIKey{
			ID:          "test-api-key",
			Key:         "test-key-value",
			Description: "Test API key",
			IsActive:    true,
		}

		body, _ := json.Marshal(apiKeyData)
		req := httptest.NewRequest("POST", "/system/api-keys", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-system-key")

		w := httptest.NewRecorder()
		server.handleCreateAPIKey(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Now get the API key
		req2 := httptest.NewRequest("GET", "/system/api-keys/test-api-key", nil)
		req2.Header.Set("X-API-Key", "test-system-key")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "test-api-key")
		req2 = req2.WithContext(context.WithValue(req2.Context(), chi.RouteCtxKey, rctx))

		w2 := httptest.NewRecorder()
		server.handleGetAPIKey(w2, req2)

		assert.Equal(t, http.StatusOK, w2.Code)

		var apiResponse APIResponse
		err := json.Unmarshal(w2.Body.Bytes(), &apiResponse)
		assert.NoError(t, err)
		assert.True(t, apiResponse.Success)

		// Extract the API key from the response data
		apiKeyResponse, ok := apiResponse.Data.(map[string]interface{})
		assert.True(t, ok)

		assert.Equal(t, "test-api-key", apiKeyResponse["id"])
		assert.Equal(t, "test-key-value", apiKeyResponse["key"])
	})

	t.Run("Delete API key", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/system/api-keys/test-api-key", nil)
		req.Header.Set("X-API-Key", "test-system-key")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "test-api-key")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		server.handleDeleteAPIKey(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "API key deleted successfully", data["message"])
	})

	t.Run("Get non-existent API key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/system/api-keys/non-existent", nil)
		req.Header.Set("X-API-Key", "test-system-key")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "non-existent")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		server.handleGetAPIKey(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestSystemConfigHandlers(t *testing.T) {
	server, cleanup := setupSystemTestServer(t)
	defer cleanup()

	t.Run("Set system config", func(t *testing.T) {
		configData := map[string]interface{}{
			"max_connections": 100,
			"timeout":         "30s",
		}

		body, _ := json.Marshal(configData)
		req := httptest.NewRequest("PUT", "/system/config/database", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-system-key")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("key", "database")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		server.handleSetSystemConfig(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "Configuration updated successfully", data["message"])
	})

	t.Run("Get system config", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/system/config/database", nil)
		req.Header.Set("X-API-Key", "test-system-key")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("key", "database")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		server.handleGetSystemConfig(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		data := response["data"].(map[string]interface{})
		assert.Contains(t, data, "key")
		assert.Contains(t, data, "value")
	})

	t.Run("Get non-existent config", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/system/config/non-existent", nil)
		req.Header.Set("X-API-Key", "test-system-key")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("key", "non-existent")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		server.handleGetSystemConfig(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestSystemAPIKeyValidation(t *testing.T) {
	server, cleanup := setupSystemTestServer(t)
	defer cleanup()

	t.Run("Invalid JSON in create request", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/system/api-keys", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-system-key")

		w := httptest.NewRecorder()
		server.handleCreateAPIKey(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Missing required fields", func(t *testing.T) {
		apiKeyData := APIKey{
			Description: "Missing ID and Key",
		}

		body, _ := json.Marshal(apiKeyData)
		req := httptest.NewRequest("POST", "/system/api-keys", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-system-key")

		w := httptest.NewRecorder()
		server.handleCreateAPIKey(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid JSON in set config", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/system/config/test", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-system-key")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("key", "test")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		server.handleSetSystemConfig(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
