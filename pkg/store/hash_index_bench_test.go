//go:build bench
// +build bench

package store

import (
	"fmt"
	"testing"
)

// BenchmarkHashIndex_Put benchmarks the Put operation with various key sizes
func BenchmarkHashIndex_Put(b *testing.B) {
	keySizes := []int{10, 100, 1000}

	for _, keySize := range keySizes {
		b.Run(fmt.Sprintf("Key_Size_%d", keySize), func(b *testing.B) {
			idx := NewHashIndex(HashIndexConfig{})

			// Generate test key
			key := make([]byte, keySize)
			for i := range keySize {
				key[i] = byte(i % 256)
			}

			entry := &IndexEntry{
				FileID:    1,
				Offset:    100,
				Size:      200,
				Timestamp: 1234567890,
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := range b.N {
				testKey := fmt.Sprintf("%s_%d", string(key[:keySize/2]), i)
				idx.Put([]byte(testKey), entry)
			}
		})
	}
}

// BenchmarkHashIndex_Get benchmarks the Get operation
func BenchmarkHashIndex_Get(b *testing.B) {
	// Pre-populate index with test data
	idx := NewHashIndex(HashIndexConfig{})

	for i := range 10000 {
		key := []byte(fmt.Sprintf("test_key_%d", i))
		entry := &IndexEntry{
			FileID:    1,
			Offset:    int64(i * 100),
			Size:      uint32(i % 1000),
			Timestamp: 1234567890,
		}
		idx.Put(key, entry)
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.Run("Existing_Key", func(b *testing.B) {
		for i := range b.N {
			key := []byte(fmt.Sprintf("test_key_%d", i%10000))
			idx.Get(key)
		}
	})

	b.Run("Non_Existent_Key", func(b *testing.B) {
		for i := range b.N {
			key := []byte(fmt.Sprintf("nonexistent_key_%d", i))
			idx.Get(key)
		}
	})
}

// BenchmarkHashIndex_Delete benchmarks the Delete operation
func BenchmarkHashIndex_Delete(b *testing.B) {
	b.Run("Delete_Existing_Key", func(b *testing.B) {
		for range b.N {
			// Create a new index for each iteration
			idx := NewHashIndex(HashIndexConfig{})

			// Pre-populate with a single key
			key := []byte("test_key")
			entry := &IndexEntry{
				FileID:    1,
				Offset:    100,
				Size:      200,
				Timestamp: 1234567890,
			}
			idx.Put(key, entry)

			b.ResetTimer()
			b.ReportAllocs()

			// Delete the key
			idx.Delete(key)

			b.StopTimer()
		}
	})
}

// BenchmarkHashIndex_BuildFromLog benchmarks building index from log
func BenchmarkHashIndex_BuildFromLog(b *testing.B) {
	// Note: This would require mocking a LogReader
	// For now, we'll benchmark the core operations that would be used in BuildFromLog
	b.Run("Simulated_BuildFromLog", func(b *testing.B) {
		for range b.N {
			idx := NewHashIndex(HashIndexConfig{})

			// Simulate inserting many entries like BuildFromLog would do
			b.StartTimer()

			for i := range 10000 {
				key := []byte(fmt.Sprintf("log_key_%d", i))
				entry := &IndexEntry{
					FileID:    1,
					Offset:    int64(i * 100),
					Size:      uint32(i % 1000),
					Timestamp: 1234567890,
				}
				idx.Put(key, entry)
			}

			b.StopTimer()
		}
	})
}

// BenchmarkHashIndex_Concurrent benchmarks concurrent operations
func BenchmarkHashIndex_Concurrent(b *testing.B) {
	// Pre-populate index with test data
	idx := NewHashIndex(HashIndexConfig{})

	for i := range 10000 {
		key := []byte(fmt.Sprintf("test_key_%d", i))
		entry := &IndexEntry{
			FileID:    1,
			Offset:    int64(i * 100),
			Size:      uint32(i % 1000),
			Timestamp: 1234567890,
		}
		idx.Put(key, entry)
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.Run("Concurrent_Reads", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := []byte(fmt.Sprintf("test_key_%d", i%10000))
				idx.Get(key)
				i++
			}
		})
	})
}
