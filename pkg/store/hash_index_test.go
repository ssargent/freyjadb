package store

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHashIndex(t *testing.T) {
	config := HashIndexConfig{}
	idx := NewHashIndex(config)

	assert.NotNil(t, idx)
	assert.NotNil(t, idx.entries)
	assert.Equal(t, 0, idx.Size())
}

func TestHashIndex_PutAndGet(t *testing.T) {
	idx := NewHashIndex(HashIndexConfig{})

	key := []byte("test_key")
	entry := &IndexEntry{
		FileID:    1,
		Offset:    100,
		Size:      50,
		Timestamp: 1234567890,
	}

	// Put entry
	idx.Put(key, entry)

	// Get entry
	retrieved, exists := idx.Get(key)
	assert.True(t, exists)
	assert.NotNil(t, retrieved)
	assert.Equal(t, entry.FileID, retrieved.FileID)
	assert.Equal(t, entry.Offset, retrieved.Offset)
	assert.Equal(t, entry.Size, retrieved.Size)
	assert.Equal(t, entry.Timestamp, retrieved.Timestamp)
}

func TestHashIndex_Get_NonExistent(t *testing.T) {
	idx := NewHashIndex(HashIndexConfig{})

	key := []byte("non_existent_key")
	entry, exists := idx.Get(key)

	assert.False(t, exists)
	assert.Nil(t, entry)
}

func TestHashIndex_Put_Overwrite(t *testing.T) {
	idx := NewHashIndex(HashIndexConfig{})

	key := []byte("test_key")

	// Put first entry
	entry1 := &IndexEntry{
		FileID:    1,
		Offset:    100,
		Size:      50,
		Timestamp: 1234567890,
	}
	idx.Put(key, entry1)

	// Put second entry with same key (should overwrite)
	entry2 := &IndexEntry{
		FileID:    2,
		Offset:    200,
		Size:      75,
		Timestamp: 1234567891,
	}
	idx.Put(key, entry2)

	// Get should return the second entry
	retrieved, exists := idx.Get(key)
	assert.True(t, exists)
	assert.Equal(t, entry2.FileID, retrieved.FileID)
	assert.Equal(t, entry2.Offset, retrieved.Offset)
	assert.Equal(t, entry2.Size, retrieved.Size)
	assert.Equal(t, entry2.Timestamp, retrieved.Timestamp)
}

func TestHashIndex_Delete(t *testing.T) {
	idx := NewHashIndex(HashIndexConfig{})

	key := []byte("test_key")
	entry := &IndexEntry{
		FileID:    1,
		Offset:    100,
		Size:      50,
		Timestamp: 1234567890,
	}

	// Put entry
	idx.Put(key, entry)

	// Verify it exists
	_, exists := idx.Get(key)
	assert.True(t, exists)

	// Delete entry
	idx.Delete(key)

	// Verify it's gone
	_, exists = idx.Get(key)
	assert.False(t, exists)
}

func TestHashIndex_Size(t *testing.T) {
	idx := NewHashIndex(HashIndexConfig{})

	// Initially empty
	assert.Equal(t, 0, idx.Size())

	// Add entries
	idx.Put([]byte("key1"), &IndexEntry{})
	idx.Put([]byte("key2"), &IndexEntry{})
	idx.Put([]byte("key3"), &IndexEntry{})

	assert.Equal(t, 3, idx.Size())

	// Delete one
	idx.Delete([]byte("key2"))
	assert.Equal(t, 2, idx.Size())

	// Clear all
	idx.Clear()
	assert.Equal(t, 0, idx.Size())
}

func TestHashIndex_Keys(t *testing.T) {
	idx := NewHashIndex(HashIndexConfig{})

	// Add some keys
	keys := [][]byte{
		[]byte("key1"),
		[]byte("key2"),
		[]byte("key3"),
	}

	for _, key := range keys {
		idx.Put(key, &IndexEntry{})
	}

	retrievedKeys := idx.Keys()
	assert.Len(t, retrievedKeys, 3)

	// Convert to map for easier checking
	keyMap := make(map[string]bool)
	for _, key := range retrievedKeys {
		keyMap[key] = true
	}

	for _, expectedKey := range keys {
		assert.True(t, keyMap[string(expectedKey)])
	}
}

func TestHashIndex_KeysWithPrefix(t *testing.T) {
	idx := NewHashIndex(HashIndexConfig{})

	// Add keys with different prefixes
	keys := []string{
		"user:1",
		"user:2",
		"item:1",
		"item:2",
		"order:1",
	}

	for _, key := range keys {
		idx.Put([]byte(key), &IndexEntry{})
	}

	// Test prefix matching
	userKeys := idx.KeysWithPrefix("user:")
	assert.Len(t, userKeys, 2)
	assert.Contains(t, userKeys, "user:1")
	assert.Contains(t, userKeys, "user:2")

	itemKeys := idx.KeysWithPrefix("item:")
	assert.Len(t, itemKeys, 2)
	assert.Contains(t, itemKeys, "item:1")
	assert.Contains(t, itemKeys, "item:2")

	orderKeys := idx.KeysWithPrefix("order:")
	assert.Len(t, orderKeys, 1)
	assert.Contains(t, orderKeys, "order:1")

	// Test non-matching prefix
	nonExistentKeys := idx.KeysWithPrefix("nonexistent:")
	assert.Len(t, nonExistentKeys, 0)
}

func TestHashIndex_ScanPrefix(t *testing.T) {
	idx := NewHashIndex(HashIndexConfig{})

	// Add keys with prefixes
	keys := []string{
		"user:1",
		"user:2",
		"user:3",
		"item:1",
	}

	for _, key := range keys {
		idx.Put([]byte(key), &IndexEntry{})
	}

	// Scan for user keys
	ch := idx.ScanPrefix("user:")
	var userKeys []string
	for key := range ch {
		userKeys = append(userKeys, key)
	}

	assert.Len(t, userKeys, 3)
	for _, key := range userKeys {
		assert.Contains(t, key, "user:")
	}
}

func TestHashIndex_ScanPrefix_EmptyResult(t *testing.T) {
	idx := NewHashIndex(HashIndexConfig{})

	// Add some keys
	idx.Put([]byte("user:1"), &IndexEntry{})
	idx.Put([]byte("item:1"), &IndexEntry{})

	// Scan for non-existent prefix
	ch := idx.ScanPrefix("nonexistent:")
	var keys []string
	for key := range ch {
		keys = append(keys, key)
	}

	assert.Len(t, keys, 0)
}

func TestHashIndex_Clear(t *testing.T) {
	idx := NewHashIndex(HashIndexConfig{})

	// Add some entries
	idx.Put([]byte("key1"), &IndexEntry{})
	idx.Put([]byte("key2"), &IndexEntry{})
	assert.Equal(t, 2, idx.Size())

	// Clear
	idx.Clear()
	assert.Equal(t, 0, idx.Size())

	// Verify entries are gone
	_, exists := idx.Get([]byte("key1"))
	assert.False(t, exists)
	_, exists = idx.Get([]byte("key2"))
	assert.False(t, exists)
}

func TestHashIndex_Stats(t *testing.T) {
	idx := NewHashIndex(HashIndexConfig{})

	// Initially empty
	stats := idx.Stats()
	assert.Equal(t, 0, stats.TotalKeys)

	// Add entries
	idx.Put([]byte("key1"), &IndexEntry{})
	idx.Put([]byte("key2"), &IndexEntry{})
	idx.Put([]byte("key3"), &IndexEntry{})

	stats = idx.Stats()
	assert.Equal(t, 3, stats.TotalKeys)
}

func TestHashIndex_ConcurrentAccess(t *testing.T) {
	idx := NewHashIndex(HashIndexConfig{})

	done := make(chan bool, 3)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			key := []byte(fmt.Sprintf("key_%d", i))
			entry := &IndexEntry{
				FileID:    uint32(i % 10),
				Offset:    int64(i * 100),
				Size:      50,
				Timestamp: uint64(i),
			}
			idx.Put(key, entry)
		}
		done <- true
	}()

	// Reader goroutine 1
	go func() {
		for i := 0; i < 50; i++ {
			key := []byte(fmt.Sprintf("key_%d", i%100))
			idx.Get(key)
		}
		done <- true
	}()

	// Reader goroutine 2
	go func() {
		for i := 0; i < 50; i++ {
			idx.Size()
			idx.Keys()
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done
}

func BenchmarkHashIndex_Put(b *testing.B) {
	idx := NewHashIndex(HashIndexConfig{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("bench_key_%d", i))
		entry := &IndexEntry{
			FileID:    uint32(i % 10),
			Offset:    int64(i * 100),
			Size:      50,
			Timestamp: uint64(i),
		}
		idx.Put(key, entry)
	}
}

func BenchmarkHashIndex_Get(b *testing.B) {
	idx := NewHashIndex(HashIndexConfig{})

	// Pre-populate
	for i := 0; i < 10000; i++ {
		key := []byte(fmt.Sprintf("bench_key_%d", i))
		entry := &IndexEntry{
			FileID:    uint32(i % 10),
			Offset:    int64(i * 100),
			Size:      50,
			Timestamp: uint64(i),
		}
		idx.Put(key, entry)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("bench_key_%d", i%10000))
		idx.Get(key)
	}
}

func BenchmarkHashIndex_KeysWithPrefix(b *testing.B) {
	idx := NewHashIndex(HashIndexConfig{})

	// Pre-populate with prefixed keys
	for i := 0; i < 10000; i++ {
		key := []byte(fmt.Sprintf("user:%d", i))
		entry := &IndexEntry{
			FileID:    uint32(i % 10),
			Offset:    int64(i * 100),
			Size:      50,
			Timestamp: uint64(i),
		}
		idx.Put(key, entry)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx.KeysWithPrefix("user:")
	}
}
