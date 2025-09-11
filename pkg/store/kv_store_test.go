package store

import (
	"os"
	"testing"
)

func TestKVStore_BasicOperations(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "freyja_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create KV store
	config := KVStoreConfig{
		DataDir:       tmpDir,
		FsyncInterval: 0, // Immediate sync for testing
	}

	store, err := NewKVStore(config)
	if err != nil {
		t.Fatalf("Failed to create KV store: %v", err)
	}

	// Open the store
	if err := store.Open(); err != nil {
		t.Fatalf("Failed to open KV store: %v", err)
	}
	defer store.Close()

	// Test Put operation
	key := []byte("test_key")
	value := []byte("test_value")

	if err := store.Put(key, value); err != nil {
		t.Fatalf("Failed to put key-value: %v", err)
	}

	// Test Get operation
	retrievedValue, err := store.Get(key)
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}

	if string(retrievedValue) != string(value) {
		t.Errorf("Retrieved value mismatch: got %s, want %s", string(retrievedValue), string(value))
	}

	// Test Get non-existent key
	_, err = store.Get([]byte("non_existent"))
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}

	// Test Delete operation
	if err := store.Delete(key); err != nil {
		t.Fatalf("Failed to delete key: %v", err)
	}

	// Verify key is deleted
	_, err = store.Get(key)
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound after delete, got %v", err)
	}
}

func TestKVStore_UpdateValue(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "freyja_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create KV store
	config := KVStoreConfig{
		DataDir:       tmpDir,
		FsyncInterval: 0,
	}

	store, err := NewKVStore(config)
	if err != nil {
		t.Fatalf("Failed to create KV store: %v", err)
	}

	if err := store.Open(); err != nil {
		t.Fatalf("Failed to open KV store: %v", err)
	}
	defer store.Close()

	key := []byte("update_key")

	// Put initial value
	initialValue := []byte("initial")
	if err := store.Put(key, initialValue); err != nil {
		t.Fatalf("Failed to put initial value: %v", err)
	}

	// Verify initial value
	retrieved, err := store.Get(key)
	if err != nil {
		t.Fatalf("Failed to get initial value: %v", err)
	}
	if string(retrieved) != string(initialValue) {
		t.Errorf("Initial value mismatch: got %s, want %s", string(retrieved), string(initialValue))
	}

	// Update value
	updatedValue := []byte("updated")
	if err := store.Put(key, updatedValue); err != nil {
		t.Fatalf("Failed to put updated value: %v", err)
	}

	// Verify updated value
	retrieved, err = store.Get(key)
	if err != nil {
		t.Fatalf("Failed to get updated value: %v", err)
	}
	if string(retrieved) != string(updatedValue) {
		t.Errorf("Updated value mismatch: got %s, want %s", string(retrieved), string(updatedValue))
	}
}

func TestKVStore_Reopen(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "freyja_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := KVStoreConfig{
		DataDir:       tmpDir,
		FsyncInterval: 0,
	}

	// First instance
	store1, err := NewKVStore(config)
	if err != nil {
		t.Fatalf("Failed to create first KV store: %v", err)
	}

	if err := store1.Open(); err != nil {
		t.Fatalf("Failed to open first KV store: %v", err)
	}

	// Put some data
	key := []byte("persistent_key")
	value := []byte("persistent_value")
	if err := store1.Put(key, value); err != nil {
		t.Fatalf("Failed to put data: %v", err)
	}

	// Close first instance
	if err := store1.Close(); err != nil {
		t.Fatalf("Failed to close first KV store: %v", err)
	}

	// Create second instance (reopen)
	store2, err := NewKVStore(config)
	if err != nil {
		t.Fatalf("Failed to create second KV store: %v", err)
	}

	if err := store2.Open(); err != nil {
		t.Fatalf("Failed to open second KV store: %v", err)
	}
	defer store2.Close()

	// Verify data persists
	retrieved, err := store2.Get(key)
	if err != nil {
		t.Fatalf("Failed to get persisted data: %v", err)
	}
	if string(retrieved) != string(value) {
		t.Errorf("Persisted value mismatch: got %s, want %s", string(retrieved), string(value))
	}
}

func TestKVStore_Stats(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "freyja_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := KVStoreConfig{
		DataDir:       tmpDir,
		FsyncInterval: 0,
	}

	store, err := NewKVStore(config)
	if err != nil {
		t.Fatalf("Failed to create KV store: %v", err)
	}

	if err := store.Open(); err != nil {
		t.Fatalf("Failed to open KV store: %v", err)
	}
	defer store.Close()

	// Initially empty
	stats := store.Stats()
	if stats.Keys != 0 {
		t.Errorf("Expected 0 keys initially, got %d", stats.Keys)
	}

	// Add some keys
	keys := [][]byte{
		[]byte("key1"),
		[]byte("key2"),
		[]byte("key3"),
	}
	values := [][]byte{
		[]byte("value1"),
		[]byte("value2"),
		[]byte("value3"),
	}

	for i, key := range keys {
		if err := store.Put(key, values[i]); err != nil {
			t.Fatalf("Failed to put key %d: %v", i, err)
		}
	}

	// Check stats
	stats = store.Stats()
	if stats.Keys != len(keys) {
		t.Errorf("Expected %d keys, got %d", len(keys), stats.Keys)
	}

	if stats.DataSize <= 0 {
		t.Errorf("Expected positive data size, got %d", stats.DataSize)
	}
}
