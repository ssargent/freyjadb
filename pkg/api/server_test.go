package api

import (
	"os"
	"testing"

	"github.com/ssargent/freyjadb/pkg/store"
)

// setupTestServer creates a test server with a temporary KV store
func setupTestServer(t *testing.T) (*Server, func()) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "freyja_server_test")
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

	// Create server
	serverConfig := ServerConfig{
		Port:   0, // Use random available port
		APIKey: "test-key",
	}

	// For tests, create a minimal metrics instance to avoid Prometheus registration conflicts
	metrics := &Metrics{} // Use empty metrics for tests
	server := NewServer(kvStore, serverConfig, metrics)

	// Return cleanup function
	cleanup := func() {
		kvStore.Close()
		os.RemoveAll(tmpDir)
	}

	return server, cleanup
}

func TestStartServer(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "freyja_server_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

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
	defer kvStore.Close()

	// Test server configuration
	serverConfig := ServerConfig{
		Port:   0, // Use random available port
		APIKey: "test-key",
	}

	// Note: We can't easily test the full server startup in unit tests
	// because it blocks on http.ListenAndServe. In integration tests,
	// we would start it in a goroutine and test the endpoints.

	// Create metrics
	metrics := NewMetrics()

	// For now, just test that the server can be created
	server := NewServer(kvStore, serverConfig, metrics)
	if server == nil {
		t.Error("Expected server to be created")
	}

	if server.store != kvStore {
		t.Error("Expected server to have the correct store")
	}

	if server.config.APIKey != "test-key" {
		t.Errorf("Expected API key to be 'test-key', got '%s'", server.config.APIKey)
	}
}

func TestServerConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   ServerConfig
		expected ServerConfig
	}{
		{
			name: "valid config",
			config: ServerConfig{
				Port:   8080,
				APIKey: "secret-key",
			},
			expected: ServerConfig{
				Port:   8080,
				APIKey: "secret-key",
			},
		},
		{
			name:   "empty config",
			config: ServerConfig{},
			expected: ServerConfig{
				Port:   0,
				APIKey: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.Port != tt.expected.Port {
				t.Errorf("Expected port %d, got %d", tt.expected.Port, tt.config.Port)
			}
			if tt.config.APIKey != tt.expected.APIKey {
				t.Errorf("Expected API key '%s', got '%s'", tt.expected.APIKey, tt.config.APIKey)
			}
		})
	}
}

func TestServer_Stats(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Put some test data
	if err := server.store.Put([]byte("test1"), []byte("value1")); err != nil {
		t.Fatalf("Failed to put test data: %v", err)
	}
	if err := server.store.Put([]byte("test2"), []byte("value2")); err != nil {
		t.Fatalf("Failed to put test data: %v", err)
	}

	// Get stats
	stats := server.store.Stats()

	if stats.Keys != 2 {
		t.Errorf("Expected 2 keys, got %d", stats.Keys)
	}

	if stats.DataSize <= 0 {
		t.Errorf("Expected positive data size, got %d", stats.DataSize)
	}
}

func TestServer_RelationshipOperations(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Create test entities
	if err := server.store.Put([]byte("user:1"), []byte("John")); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	if err := server.store.Put([]byte("item:1"), []byte("Laptop")); err != nil {
		t.Fatalf("Failed to create test item: %v", err)
	}

	// Test creating relationship
	if err := server.store.PutRelationship("user:1", "item:1", "owns"); err != nil {
		t.Fatalf("Failed to create relationship: %v", err)
	}

	// Test querying relationships
	query := store.RelationshipQuery{
		Key:       "user:1",
		Direction: "outgoing",
		Relation:  "owns",
		Limit:     10,
	}

	results, err := server.store.GetRelationships(query)
	if err != nil {
		t.Fatalf("Failed to query relationships: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 relationship, got %d", len(results))
	}

	if results[0].OtherKey != "item:1" {
		t.Errorf("Expected other key to be 'item:1', got '%s'", results[0].OtherKey)
	}

	if results[0].Relationship.Relation != "owns" {
		t.Errorf("Expected relation to be 'owns', got '%s'", results[0].Relationship.Relation)
	}

	// Test deleting relationship
	if err := server.store.DeleteRelationship("user:1", "item:1", "owns"); err != nil {
		t.Fatalf("Failed to delete relationship: %v", err)
	}

	// Verify relationship is deleted
	results, err = server.store.GetRelationships(query)
	if err != nil {
		t.Fatalf("Failed to query relationships after delete: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 relationships after delete, got %d", len(results))
	}
}
