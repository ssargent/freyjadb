package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestKVStore_BasicOperations(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "freyja_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }() // Ignore error in cleanup

	// Create KV store
	config := KVStoreConfig{
		DataDir:       tmpDir,
		FsyncInterval: 0, // Immediate sync for testing
		MaxRecordSize: 4096,
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
	defer func() { _ = store.Close() }() // Ignore error in cleanup

	// Test Put operation
	key := []byte("test_key")
	value := []byte("test_value")

	if putErr := store.Put(key, value); putErr != nil {
		t.Fatalf("Failed to put key-value: %v", putErr)
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
	if delErr := store.Delete(key); delErr != nil {
		t.Fatalf("Failed to delete key: %v", delErr)
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
	defer func() { _ = os.RemoveAll(tmpDir) }() // Ignore error in cleanup

	// Create KV store
	config := KVStoreConfig{
		DataDir:       tmpDir,
		FsyncInterval: 0,
		MaxRecordSize: 4096,
	}

	store, err := NewKVStore(config)
	if err != nil {
		t.Fatalf("Failed to create KV store: %v", err)
	}

	_, err = store.Open()
	if err != nil {
		t.Fatalf("Failed to open KV store: %v", err)
	}
	defer func() { _ = store.Close() }() // Ignore error in cleanup

	key := []byte("update_key")

	// Put initial value
	initialValue := []byte("initial")
	if putErr := store.Put(key, initialValue); putErr != nil {
		t.Fatalf("Failed to put initial value: %v", putErr)
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
	if putErr := store.Put(key, updatedValue); putErr != nil {
		t.Fatalf("Failed to put updated value: %v", putErr)
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
	defer func() { _ = os.RemoveAll(tmpDir) }() // Ignore error in cleanup

	config := KVStoreConfig{
		DataDir:       tmpDir,
		FsyncInterval: 0,
		MaxRecordSize: 4096,
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
	if putErr := store1.Put(key, value); putErr != nil {
		t.Fatalf("Failed to put data: %v", putErr)
	}

	// Close first instance
	if closeErr := store1.Close(); closeErr != nil {
		t.Fatalf("Failed to close first KV store: %v", closeErr)
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
	defer func() { _ = store2.Close() }() // Ignore error in cleanup

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
	defer func() { _ = os.RemoveAll(tmpDir) }() // Ignore error in cleanup

	config := KVStoreConfig{
		DataDir:       tmpDir,
		FsyncInterval: 0,
		MaxRecordSize: 4096,
	}

	store, err := NewKVStore(config)
	if err != nil {
		t.Fatalf("Failed to create KV store: %v", err)
	}

	_, err = store.Open()
	if err != nil {
		t.Fatalf("Failed to open KV store: %v", err)
	}
	defer func() { _ = store.Close() }() // Ignore error in cleanup

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
	defer func() { _ = os.RemoveAll(tmpDir) }() // Ignore error in cleanup

	config := KVStoreConfig{
		DataDir:       tmpDir,
		FsyncInterval: 0,
		MaxRecordSize: 4096,
	}

	keys := [][]byte{[]byte("key1"), []byte("key2"), []byte("key3")}
	values := [][]byte{[]byte("value1"), []byte("value2"), []byte("value3")}

	// First instance - create and populate data
	store1, err := NewKVStore(config)
	if err != nil {
		t.Fatalf("Failed to create first KV store: %v", err)
	}

	recoveryResult, err := store1.Open()
	if err != nil {
		t.Fatalf("Failed to open first KV store: %v", err)
	}
	assertCleanStartupRecovery(t, recoveryResult)

	putAllOrFail(t, store1, keys, values)

	if closeErr := store1.Close(); closeErr != nil {
		t.Fatalf("Failed to close first KV store: %v", closeErr)
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
	defer func() { _ = store2.Close() }() // Ignore error in cleanup

	assertReopenRecoveryStats(t, recoveryResult2, int64(len(keys)))
	assertAllRetrievable(t, store2, keys, values)
}

func assertCleanStartupRecovery(t *testing.T, r *RecoveryResult) {
	t.Helper()
	if r.RecordsTruncated != 0 {
		t.Errorf("Expected no records truncated on clean startup, got %d", r.RecordsTruncated)
	}
	if r.FileSizeBefore != 0 {
		t.Errorf("Expected file size before to be 0 on clean startup, got %d", r.FileSizeBefore)
	}
}

func assertReopenRecoveryStats(t *testing.T, r *RecoveryResult, expectedRecords int64) {
	t.Helper()
	if r.RecordsValidated != expectedRecords {
		t.Errorf("Expected %d records validated, got %d", expectedRecords, r.RecordsValidated)
	}
	if r.RecordsTruncated != 0 {
		t.Errorf("Expected no records truncated, got %d", r.RecordsTruncated)
	}
	if !r.IndexRebuilt {
		t.Error("Expected index to be rebuilt")
	}
}

func putAllOrFail(t *testing.T, store *KVStore, keys, values [][]byte) {
	t.Helper()
	for i, key := range keys {
		if err := store.Put(key, values[i]); err != nil {
			t.Fatalf("Failed to put key %d: %v", i, err)
		}
	}
}

func assertAllRetrievable(t *testing.T, store *KVStore, keys, values [][]byte) {
	t.Helper()
	for i, key := range keys {
		retrieved, err := store.Get(key)
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
	defer func() { _ = os.RemoveAll(tmpDir) }() // Ignore error in cleanup

	config := KVStoreConfig{
		DataDir:       tmpDir,
		FsyncInterval: 0,
		MaxRecordSize: 4096,
	}

	store, err := NewKVStore(config)
	if err != nil {
		t.Fatalf("Failed to create KV store: %v", err)
	}

	recoveryResult, err := store.Open()
	if err != nil {
		t.Fatalf("Failed to open KV store: %v", err)
	}
	defer func() { _ = store.Close() }() // Ignore error in cleanup

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

func TestKVStore_ValidateLogFile_DecomposedFunctions(t *testing.T) {
	// Test the decomposed functions individually
	tmpDir, err := os.MkdirTemp("", "freyja_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }() // Ignore error in cleanup

	store, err := NewKVStore(KVStoreConfig{
		DataDir:       tmpDir,
		FsyncInterval: 0,
		MaxRecordSize: 4096,
	})
	if err != nil {
		t.Fatalf("Failed to create KV store: %v", err)
	}

	// Test createEmptyRecoveryResult
	startTime := time.Now()
	result := store.createEmptyRecoveryResult(startTime)
	if result.RecordsValidated != 0 {
		t.Errorf("Expected 0 records validated, got %d", result.RecordsValidated)
	}
	if result.IndexRebuilt != true {
		t.Error("Expected IndexRebuilt to be true")
	}
	if result.RecoveryTime < 0 {
		t.Error("Expected non-negative recovery time")
	}

	// Test with non-existent file
	nonExistentPath := filepath.Join(tmpDir, "nonexistent.data")
	result, err = store.validateLogFile(nonExistentPath)
	if err != nil {
		t.Fatalf("Expected no error for non-existent file, got %v", err)
	}
	if result.RecordsValidated != 0 {
		t.Errorf("Expected 0 records validated for non-existent file, got %d", result.RecordsValidated)
	}
}

func TestKVStore_RecordSizeValidation(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "freyja_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }() // Ignore error in cleanup

	// Create KV store with small max record size for testing
	config := KVStoreConfig{
		DataDir:       tmpDir,
		FsyncInterval: 0,
		MaxRecordSize: 100, // Small size for testing
	}

	store, err := NewKVStore(config)
	if err != nil {
		t.Fatalf("Failed to create KV store: %v", err)
	}

	_, err = store.Open()
	if err != nil {
		t.Fatalf("Failed to open KV store: %v", err)
	}
	defer func() { _ = store.Close() }() // Ignore error in cleanup

	// Test with record size within limit
	smallKey := []byte("small_key")
	smallValue := make([]byte, 50) // 50 bytes
	for i := range smallValue {
		smallValue[i] = byte(i % 256)
	}

	if err := store.Put(smallKey, smallValue); err != nil {
		t.Fatalf("Failed to put small record: %v", err)
	}

	// Test with record size exceeding limit
	largeKey := []byte("large_key")
	largeValue := make([]byte, 200) // 200 bytes, exceeds 100 byte limit
	for i := range largeValue {
		largeValue[i] = byte(i % 256)
	}

	if err := store.Put(largeKey, largeValue); err != ErrRecordSizeExceeded {
		t.Errorf("Expected ErrRecordSizeExceeded, got %v", err)
	}

	// Test with record size exactly at limit
	exactKey := []byte("exact_key")
	exactValue := make([]byte, 100-len(exactKey)) // Exactly at limit

	if err := store.Put(exactKey, exactValue); err != nil {
		t.Fatalf("Failed to put record at size limit: %v", err)
	}
}
