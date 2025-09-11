package api

import (
	"bytes"
	"encoding/json"
	"testing"
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
