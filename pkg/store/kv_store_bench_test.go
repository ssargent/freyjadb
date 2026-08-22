//go:build bench
// +build bench

package store

import (
	"fmt"
	"os"
	"testing"
)

// BenchmarkKVStore_Put benchmarks the Put operation with various key/value sizes
func BenchmarkKVStore_Put(b *testing.B) {
	tests := []struct {
		name      string
		keySize   int
		valueSize int
	}{
		{"Small_Key_Small_Value", 10, 100},
		{"Small_Key_Large_Value", 10, 10000},
		{"Large_Key_Small_Value", 1000, 100},
		{"Large_Key_Large_Value", 1000, 10000},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			// Create temporary directory for test
			tmpDir, err := os.MkdirTemp("", "freyja_bench")
			if err != nil {
				b.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Create KV store
			config := KVStoreConfig{
				DataDir:       tmpDir,
				FsyncInterval: 0,           // Disable fsync for benchmarking
				MaxRecordSize: 1024 * 1024, // 1MB max record size
			}

			store, err := NewKVStore(config)
			if err != nil {
				b.Fatalf("Failed to create KV store: %v", err)
			}

			// Open the store
			_, err = store.Open()
			if err != nil {
				b.Fatalf("Failed to open KV store: %v", err)
			}
			defer store.Close()

			// Generate test data
			key := make([]byte, tt.keySize)
			value := make([]byte, tt.valueSize)
			for i := 0; i < tt.keySize; i++ {
				key[i] = byte(i % 256)
			}
			for i := 0; i < tt.valueSize; i++ {
				value[i] = byte(i % 256)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				testKey := fmt.Sprintf("key_%d", i)
				if err := store.Put([]byte(testKey), value); err != nil {
					b.Fatalf("Failed to put key-value: %v", err)
				}
			}
		})
	}
}

// BenchmarkKVStore_Get benchmarks the Get operation with various key/value sizes
func BenchmarkKVStore_Get(b *testing.B) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "freyja_bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create KV store
	config := KVStoreConfig{
		DataDir:       tmpDir,
		FsyncInterval: 0,           // Disable fsync for benchmarking
		MaxRecordSize: 1024 * 1024, // 1MB max record size
	}

	store, err := NewKVStore(config)
	if err != nil {
		b.Fatalf("Failed to create KV store: %v", err)
	}

	// Open the store
	_, err = store.Open()
	if err != nil {
		b.Fatalf("Failed to open KV store: %v", err)
	}
	defer store.Close()

	// Pre-populate store with test data
	value := make([]byte, 1000)
	for i := 0; i < 1000; i++ {
		value[i] = byte(i % 256)
	}

	for i := 0; i < 1000; i++ {
		testKey := fmt.Sprintf("key_%d", i)
		if err := store.Put([]byte(testKey), value); err != nil {
			b.Fatalf("Failed to put key-value: %v", err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.Run("Existing_Key", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			testKey := fmt.Sprintf("key_%d", i%1000)
			_, err := store.Get([]byte(testKey))
			if err != nil {
				b.Fatalf("Failed to get value: %v", err)
			}
		}
	})

	b.Run("Non_Existent_Key", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			testKey := fmt.Sprintf("nonexistent_key_%d", i)
			_, err := store.Get([]byte(testKey))
			if err != ErrKeyNotFound {
				b.Fatalf("Expected ErrKeyNotFound, got %v", err)
			}
		}
	})
}

// BenchmarkKVStore_Delete benchmarks the Delete operation
func BenchmarkKVStore_Delete(b *testing.B) {
	b.Run("Delete_Existing_Key", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Create temporary directory for test
			tmpDir, err := os.MkdirTemp("", "freyja_bench")
			if err != nil {
				b.Fatalf("Failed to create temp dir: %v", err)
			}

			// Clean up after each iteration
			defer os.RemoveAll(tmpDir)

			// Create KV store
			config := KVStoreConfig{
				DataDir:       tmpDir,
				FsyncInterval: 0,           // Disable fsync for benchmarking
				MaxRecordSize: 1024 * 1024, // 1MB max record size
			}

			store, err := NewKVStore(config)
			if err != nil {
				b.Fatalf("Failed to create KV store: %v", err)
			}

			// Open the store
			_, err = store.Open()
			if err != nil {
				b.Fatalf("Failed to open KV store: %v", err)
			}

			// Pre-populate with a single key
			key := []byte("test_key")
			value := []byte("test_value")
			if err := store.Put(key, value); err != nil {
				b.Fatalf("Failed to put key-value: %v", err)
			}

			b.ResetTimer()
			b.ReportAllocs()

			// Delete the key
			if err := store.Delete(key); err != nil {
				b.Fatalf("Failed to delete key: %v", err)
			}

			b.StopTimer()
			store.Close()
		}
	})
}

// BenchmarkKVStore_Concurrent benchmarks concurrent operations
func BenchmarkKVStore_Concurrent(b *testing.B) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "freyja_bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create KV store
	config := KVStoreConfig{
		DataDir:       tmpDir,
		FsyncInterval: 0,           // Disable fsync for benchmarking
		MaxRecordSize: 1024 * 1024, // 1MB max record size
	}

	store, err := NewKVStore(config)
	if err != nil {
		b.Fatalf("Failed to create KV store: %v", err)
	}

	// Open the store
	_, err = store.Open()
	if err != nil {
		b.Fatalf("Failed to open KV store: %v", err)
	}
	defer store.Close()

	// Pre-populate store with test data
	value := make([]byte, 100)
	for i := 0; i < 100; i++ {
		value[i] = byte(i % 256)
	}

	for i := 0; i < 1000; i++ {
		testKey := fmt.Sprintf("key_%d", i)
		if err := store.Put([]byte(testKey), value); err != nil {
			b.Fatalf("Failed to put key-value: %v", err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.Run("Concurrent_Reads", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				testKey := fmt.Sprintf("key_%d", i%1000)
				_, err := store.Get([]byte(testKey))
				if err != nil && err != ErrKeyNotFound {
					b.Fatalf("Failed to get value: %v", err)
				}
				i++
			}
		})
	})
}

// BenchmarkKVStore_BatchOperations benchmarks batch put operations
func BenchmarkKVStore_BatchOperations(b *testing.B) {
	b.Run("Batch_Put_Small", func(b *testing.B) {
		batchSizes := []int{10, 100, 1000}

		for _, batchSize := range batchSizes {
			b.Run(fmt.Sprintf("%d_operations", batchSize), func(b *testing.B) {
				// Create temporary directory for test
				tmpDir, err := os.MkdirTemp("", "freyja_bench")
				if err != nil {
					b.Fatalf("Failed to create temp dir: %v", err)
				}
				defer os.RemoveAll(tmpDir)

				// Create KV store
				config := KVStoreConfig{
					DataDir:       tmpDir,
					FsyncInterval: 0,           // Disable fsync for benchmarking
					MaxRecordSize: 1024 * 1024, // 1MB max record size
				}

				store, err := NewKVStore(config)
				if err != nil {
					b.Fatalf("Failed to create KV store: %v", err)
				}

				// Open the store
				_, err = store.Open()
				if err != nil {
					b.Fatalf("Failed to open KV store: %v", err)
				}
				defer store.Close()

				// Prepare batch data
				keys := make([][]byte, batchSize)
				values := make([][]byte, batchSize)

				for i := 0; i < batchSize; i++ {
					keys[i] = []byte(fmt.Sprintf("batch_key_%d", i))
					values[i] = []byte(fmt.Sprintf("batch_value_%d_with_some_content", i))
				}

				b.ResetTimer()
				b.ReportAllocs()

				for i := 0; i < b.N; i++ {
					// Execute batch operations
					for j := 0; j < batchSize; j++ {
						if err := store.Put(keys[j], values[j]); err != nil {
							b.Fatalf("Failed to put key-value: %v", err)
						}
					}
				}
			})
		}
	})
}
