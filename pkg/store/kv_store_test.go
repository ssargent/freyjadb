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
	_, err = store.Open()
	if err != nil {
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

	_, err = store.Open()
	if err != nil {
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

	_, err = store1.Open()
	if err != nil {
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

	_, err = store2.Open()
	if err != nil {
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

	_, err = store.Open()
	if err != nil {
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

func TestKVStore_CrashSafeReopen_CleanFile(t *testing.T) {
	// Test clean restart with no corruption
	tmpDir, err := os.MkdirTemp("", "freyja_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := KVStoreConfig{
		DataDir:       tmpDir,
		FsyncInterval: 0,
	}

	// First instance - create and populate data
	store1, err := NewKVStore(config)
	if err != nil {
		t.Fatalf("Failed to create first KV store: %v", err)
	}

	recoveryResult, err := store1.Open()
	if err != nil {
		t.Fatalf("Failed to open first KV store: %v", err)
	}

	// Verify clean startup (no corruption)
	if recoveryResult.RecordsTruncated != 0 {
		t.Errorf("Expected no records truncated on clean startup, got %d", recoveryResult.RecordsTruncated)
	}

	if recoveryResult.FileSizeBefore != 0 {
		t.Errorf("Expected file size before to be 0 on clean startup, got %d", recoveryResult.FileSizeBefore)
	}

	// Add some data
	keys := [][]byte{[]byte("key1"), []byte("key2"), []byte("key3")}
	values := [][]byte{[]byte("value1"), []byte("value2"), []byte("value3")}

	for i, key := range keys {
		if err := store1.Put(key, values[i]); err != nil {
			t.Fatalf("Failed to put key %d: %v", i, err)
		}
	}

	if err := store1.Close(); err != nil {
		t.Fatalf("Failed to close first KV store: %v", err)
	}

	// Second instance - reopen and verify recovery
	store2, err := NewKVStore(config)
	if err != nil {
		t.Fatalf("Failed to create second KV store: %v", err)
	}

	recoveryResult2, err := store2.Open()
	if err != nil {
		t.Fatalf("Failed to open second KV store: %v", err)
	}
	defer store2.Close()

	// Verify recovery statistics
	if recoveryResult2.RecordsValidated != 3 {
		t.Errorf("Expected 3 records validated, got %d", recoveryResult2.RecordsValidated)
	}

	if recoveryResult2.RecordsTruncated != 0 {
		t.Errorf("Expected no records truncated, got %d", recoveryResult2.RecordsTruncated)
	}

	if !recoveryResult2.IndexRebuilt {
		t.Error("Expected index to be rebuilt")
	}

	// Verify data integrity
	for i, key := range keys {
		retrieved, err := store2.Get(key)
		if err != nil {
			t.Fatalf("Failed to get key %d: %v", i, err)
		}
		if string(retrieved) != string(values[i]) {
			t.Errorf("Data mismatch for key %d: got %s, want %s", i, string(retrieved), string(values[i]))
		}
	}
}

// TODO: Add corruption test once file format is better understood
// The current implementation provides the framework for corruption detection
// but requires deeper understanding of the exact record format for reliable testing

func TestKVStore_CrashSafeReopen_EmptyFile(t *testing.T) {
	// Test recovery from empty/non-existent file
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

	recoveryResult, err := store.Open()
	if err != nil {
		t.Fatalf("Failed to open KV store: %v", err)
	}
	defer store.Close()

	// Verify empty file recovery
	if recoveryResult.RecordsValidated != 0 {
		t.Errorf("Expected 0 records validated for empty file, got %d", recoveryResult.RecordsValidated)
	}

	if recoveryResult.RecordsTruncated != 0 {
		t.Errorf("Expected 0 records truncated for empty file, got %d", recoveryResult.RecordsTruncated)
	}

	if recoveryResult.FileSizeBefore != 0 {
		t.Errorf("Expected file size before to be 0 for empty file, got %d", recoveryResult.FileSizeBefore)
	}

	if !recoveryResult.IndexRebuilt {
		t.Error("Expected index to be marked as rebuilt even for empty file")
	}
}
