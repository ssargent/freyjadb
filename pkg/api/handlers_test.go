package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ssargent/freyjadb/pkg/store"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestContentTypeHandling(t *testing.T) {
	// Create a mock store (you would need to implement this)
	// For now, we'll test the helper functions

	t.Run("encode/decode with content type", func(t *testing.T) {
		originalData := []byte(`{"name": "test", "value": 123}`)
		contentType := ContentTypeJSON

		encoded := encodeDataWithContentType(originalData, contentType)
		decoded, decodedType, err := decodeDataWithContentType(encoded)

		if err != nil {
			t.Fatalf("Failed to decode: %v", err)
		}

		if decodedType != contentType {
			t.Errorf("Expected content type %d, got %d", contentType, decodedType)
		}

		if !bytes.Equal(decoded, originalData) {
			t.Errorf("Decoded data doesn't match original")
		}
	})

	t.Run("backward compatibility - no header", func(t *testing.T) {
		originalData := []byte("raw data without header")

		// Data without header should be treated as raw bytes
		decoded, decodedType, err := decodeDataWithContentType(originalData)

		if err != nil {
			t.Fatalf("Failed to decode: %v", err)
		}

		if decodedType != ContentTypeRaw {
			t.Errorf("Expected content type %d for raw data, got %d", ContentTypeRaw, decodedType)
		}

		if !bytes.Equal(decoded, originalData) {
			t.Errorf("Decoded data doesn't match original")
		}
	})

	t.Run("content type header parsing", func(t *testing.T) {
		tests := []struct {
			header   string
			expected int
		}{
			{"application/json", ContentTypeJSON},
			{"application/json; charset=utf-8", ContentTypeJSON},
			{"text/plain", ContentTypeRaw},
			{"", ContentTypeRaw},
			{"application/octet-stream", ContentTypeRaw},
		}

		for _, test := range tests {
			result := getContentTypeFromHeader(test.header)
			if result != test.expected {
				t.Errorf("Header '%s': expected %d, got %d", test.header, test.expected, result)
			}
		}
	})

	t.Run("content type header generation", func(t *testing.T) {
		tests := []struct {
			contentType int
			expected    string
		}{
			{ContentTypeJSON, "application/json"},
			{ContentTypeRaw, "application/octet-stream"},
		}

		for _, test := range tests {
			result := getContentTypeHeader(test.contentType)
			if result != test.expected {
				t.Errorf("Content type %d: expected '%s', got '%s'", test.contentType, test.expected, result)
			}
		}
	})
}

func TestJSONValidation(t *testing.T) {
	t.Run("valid JSON", func(t *testing.T) {
		validJSON := []byte(`{"key": "value", "number": 42}`)
		var data interface{}
		err := json.Unmarshal(validJSON, &data)
		if err != nil {
			t.Errorf("Valid JSON should not error: %v", err)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		invalidJSON := []byte(`{"key": "value", invalid}`)
		var data interface{}
		err := json.Unmarshal(invalidJSON, &data)
		if err == nil {
			t.Errorf("Invalid JSON should error")
		}
	})
}

// mockKVStore wraps KVStore to allow mocking Put method
type mockKVStore struct {
	*store.KVStore
	putFunc func(key, value []byte) error
}

func (m *mockKVStore) Put(key, value []byte) error {
	if m.putFunc != nil {
		return m.putFunc(key, value)
	}
	return m.KVStore.Put(key, value)
}

// mockMetrics embeds Metrics but overrides methods to be no-ops for testing
type mockMetrics struct {
	*Metrics
}

func (m *mockMetrics) RecordHTTPRequest(method, endpoint string, statusCode int, duration time.Duration) {
}
func (m *mockMetrics) RecordDBOperation(operation string, success bool, duration time.Duration) {}
func (m *mockMetrics) UpdateDBStats(keys int, dataSize int64)                                   {}
func (m *mockMetrics) RecordAuthRequest(success bool)                                           {}
func (m *mockMetrics) RecordRelationshipOperation(operation string, success bool)               {}
func (m *mockMetrics) RecordHealthCheck(success bool)                                           {}
func (m *mockMetrics) InstrumentHandler(method, endpoint string, handler http.HandlerFunc) http.HandlerFunc {
	return handler
}
func (m *mockMetrics) InstrumentAuthMiddleware(next func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return next
}

func helperEncodeJsonWithContentType(t *testing.T, data string) []byte {
	var mything interface{}
	err := json.Unmarshal([]byte(data), &mything)
	assert.NoError(t, err)

	encodedData, err := json.Marshal(mything)
	assert.NoError(t, err)

	return encodeDataWithContentType(encodedData, ContentTypeJSON)
}

// TestHandlePut tests the handlePut function with various scenarios
// Note on macos this test may warn about a very odd linker error:
// ld: warning: '/private/var/folders/mn/5l10pmk93_l4j6hv6k4cgd280000gn/T/go-link-2562642374/000013.o' has malformed LC_DYSYMTAB, expected 98 undefined symbols to start at index 1626, found 95 undefined symbols starting at index 1626

func TestHandlePut(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		body           string
		contentType    string
		mockPutError   error
		expectedStatus int
		expectedBody   string
		mocks          func(store *MockIKVStore)
	}{
		{
			name:           "valid JSON put",
			key:            "testkey",
			body:           `{"name": "test", "value": 12345}`,
			contentType:    "application/json",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"success":true,"data":{"message":"Key-value pair stored successfully"}}`,
			mocks: func(store *MockIKVStore) {
				store.
					EXPECT().
					Put(
						[]byte("testkey"),
						helperEncodeJsonWithContentType(t, `{"name": "test", "value": 12345}`),
					).
					Return(nil)
			},
		},
		{
			name:           "valid raw put",
			key:            "testkey",
			body:           "raw data content",
			contentType:    "application/octet-stream",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"success":true,"data":{"message":"Key-value pair stored successfully"}}`,
			mocks: func(store *MockIKVStore) {
				store.
					EXPECT().
					Put(
						[]byte("testkey"),
						encodeDataWithContentType([]byte("raw data content"), ContentTypeRaw),
					).
					Return(nil)
			},
		},
		{
			name:           "missing key",
			key:            "",
			body:           "some data",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":"Key is required"}`,
			mocks:          func(store *MockIKVStore) {},
		},
		{
			name:           "invalid JSON",
			key:            "testkey",
			body:           `{"invalid": json}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":"Invalid JSON in request body"}`,
			mocks:          func(store *MockIKVStore) {},
		},
		{
			name:           "store put error",
			key:            "testkey",
			body:           "data",
			mockPutError:   errors.New("mock put error"), // This will cause the store to not be opened
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"success":false,"error":"Failed to put key-value: store is not open"}`,
			mocks: func(store *MockIKVStore) {
				store.
					EXPECT().
					Put([]byte("testkey"), encodeDataWithContentType([]byte("data"), ContentTypeRaw)).
					Return(errors.New("store is not open"))
			},
		},
		{
			name:           "empty body raw",
			key:            "testkey",
			body:           "",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"success":true,"data":{"message":"Key-value pair stored successfully"}}`,
			mocks: func(store *MockIKVStore) {
				store.EXPECT().Put([]byte("testkey"), encodeDataWithContentType([]byte(""), ContentTypeRaw)).Return(nil)
			},
		},
		{
			name:           "url encoded key",
			key:            "user%2F123", // URL encoded "user/123"
			body:           `{"info": "some user data"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"success":true,"data":{"message":"Key-value pair stored successfully"}}`,
			mocks: func(store *MockIKVStore) {
				store.
					EXPECT().
					Put(
						[]byte("user/123"),
						helperEncodeJsonWithContentType(t, `{"info": "some user data"}`),
					).
					Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a real KVStore for testing
			kvStore, err := store.NewKVStore(store.KVStoreConfig{DataDir: "/tmp/test_" + tt.name})
			if err != nil {
				t.Fatalf("Failed to create KV store: %v", err)
			}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := NewMockIKVStore(ctrl)
			tt.mocks(mockStore)

			if tt.mockPutError == nil {
				// For normal cases, open the store
				_, err = kvStore.Open()
				if err != nil {
					t.Fatalf("Failed to open KV store: %v", err)
				}
				defer kvStore.Close()
			}
			// For error case, don't open the store, so Put will fail

			// Create mock system service
			mockSystemService := &SystemService{} // Will be closed, so no-op is fine

			// Create server with store
			server := NewServer(mockStore, mockSystemService, ServerConfig{}, nil)

			// Create request
			req := httptest.NewRequest(http.MethodPut, "/kv/"+tt.key, strings.NewReader(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			req.Header.Set("Content-Length", string(rune(len(tt.body))))

			// Set up chi router context for URL param
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("key", tt.key)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			server.handlePut(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check response body
			if strings.TrimSpace(w.Body.String()) != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}
