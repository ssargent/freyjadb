//go:build bench
// +build bench

package store

import (
	"fmt"
	"os"
	"testing"
)

// BenchmarkLogWriter_Put benchmarks the Put operation with various record sizes
func BenchmarkLogWriter_Put(b *testing.B) {
	recordSizes := []int{100, 1000, 10000, 100000}

	for _, recordSize := range recordSizes {
		b.Run(fmt.Sprintf("Record_Size_%d", recordSize), func(b *testing.B) {
			// Create temporary file for test
			tmpFile, err := os.CreateTemp("", "freyja_log_writer_bench")
			if err != nil {
				b.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())
			defer tmpFile.Close()

			// Create log writer
			config := LogWriterConfig{
				FilePath:      tmpFile.Name(),
				FsyncInterval: 0,         // Disable fsync for benchmarking
				BufferSize:    64 * 1024, // 64KB buffer
			}

			writer, err := NewLogWriter(config)
			if err != nil {
				b.Fatalf("Failed to create log writer: %v", err)
			}
			defer writer.Close()

			// Generate test data
			key := []byte("test_key")
			value := make([]byte, recordSize)
			for i := range recordSize {
				value[i] = byte(i % 256)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := range b.N {
				testKey := fmt.Sprintf("%s_%d", string(key), i)
				_, err := writer.Put([]byte(testKey), value)
				if err != nil {
					b.Fatalf("Failed to put record: %v", err)
				}
			}
		})
	}
}

// BenchmarkLogWriter_Flush benchmarks the Flush operation
func BenchmarkLogWriter_Flush(b *testing.B) {
	b.Run("Flush_After_Writes", func(b *testing.B) {
		for range b.N {
			// Create temporary file for test
			tmpFile, err := os.CreateTemp("", "freyja_log_writer_bench")
			if err != nil {
				b.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())
			defer tmpFile.Close()

			// Create log writer
			config := LogWriterConfig{
				FilePath:      tmpFile.Name(),
				FsyncInterval: 0,         // Disable fsync for benchmarking
				BufferSize:    64 * 1024, // 64KB buffer
			}

			writer, err := NewLogWriter(config)
			if err != nil {
				b.Fatalf("Failed to create log writer: %v", err)
			}
			defer writer.Close()

			// Write some data to buffer
			key := []byte("test_key")
			value := []byte("test_value")
			for i := range 1000 {
				testKey := fmt.Sprintf("%s_%d", string(key), i)
				_, err := writer.Put([]byte(testKey), value)
				if err != nil {
					b.Fatalf("Failed to put record: %v", err)
				}
			}

			b.ResetTimer()
			b.ReportAllocs()

			// Flush the buffer
			err = writer.Flush()
			if err != nil {
				b.Fatalf("Failed to flush: %v", err)
			}

			b.StopTimer()
		}
	})
}

// BenchmarkLogWriter_BufferSizes benchmarks different buffer sizes
func BenchmarkLogWriter_BufferSizes(b *testing.B) {
	bufferSizes := []int{4 * 1024, 16 * 1024, 64 * 1024, 256 * 1024} // 4KB, 16KB, 64KB, 256KB

	for _, bufferSize := range bufferSizes {
		b.Run(fmt.Sprintf("Buffer_Size_%d", bufferSize), func(b *testing.B) {
			// Create temporary file for test
			tmpFile, err := os.CreateTemp("", "freyja_log_writer_bench")
			if err != nil {
				b.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())
			defer tmpFile.Close()

			// Create log writer
			config := LogWriterConfig{
				FilePath:      tmpFile.Name(),
				FsyncInterval: 0, // Disable fsync for benchmarking
				BufferSize:    bufferSize,
			}

			writer, err := NewLogWriter(config)
			if err != nil {
				b.Fatalf("Failed to create log writer: %v", err)
			}
			defer writer.Close()

			// Generate test data
			key := []byte("test_key")
			value := []byte("test_value_with_some_additional_content_to_make_it_larger")

			b.ResetTimer()
			b.ReportAllocs()

			for i := range b.N {
				testKey := fmt.Sprintf("%s_%d", string(key), i)
				_, err := writer.Put([]byte(testKey), value)
				if err != nil {
					b.Fatalf("Failed to put record: %v", err)
				}
			}
		})
	}
}

// BenchmarkLogWriter_Concurrent benchmarks concurrent writes
func BenchmarkLogWriter_Concurrent(b *testing.B) {
	// Create temporary file for test
	tmpFile, err := os.CreateTemp("", "freyja_log_writer_bench")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create log writer
	config := LogWriterConfig{
		FilePath:      tmpFile.Name(),
		FsyncInterval: 0,         // Disable fsync for benchmarking
		BufferSize:    64 * 1024, // 64KB buffer
	}

	writer, err := NewLogWriter(config)
	if err != nil {
		b.Fatalf("Failed to create log writer: %v", err)
	}
	defer writer.Close()

	b.ResetTimer()
	b.ReportAllocs()

	b.Run("Concurrent_Writes", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := []byte(fmt.Sprintf("parallel_key_%d", i))
				value := []byte(fmt.Sprintf("parallel_value_%d_with_some_content", i))
				_, err := writer.Put(key, value)
				if err != nil {
					b.Fatalf("Failed to put record: %v", err)
				}
				i++
			}
		})
	})
}
