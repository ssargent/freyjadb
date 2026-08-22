//go:build bench
// +build bench

package store

import (
	"fmt"
	"os"
	"testing"
)

// BenchmarkLogReader_ReadNext benchmarks the ReadNext operation
func BenchmarkLogReader_ReadNext(b *testing.B) {
	// Create temporary file with test data
	tmpFile, err := os.CreateTemp("", "freyja_log_reader_bench")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write test data using log writer
	writerConfig := LogWriterConfig{
		FilePath:      tmpFile.Name(),
		FsyncInterval: 0,         // Disable fsync for benchmarking
		BufferSize:    64 * 1024, // 64KB buffer
	}

	writer, err := NewLogWriter(writerConfig)
	if err != nil {
		b.Fatalf("Failed to create log writer: %v", err)
	}

	// Write multiple records
	recordCount := 10000
	for i := range recordCount {
		key := []byte(fmt.Sprintf("key_%d", i))
		value := []byte(fmt.Sprintf("value_%d_with_some_additional_content", i))
		_, err := writer.Put(key, value)
		if err != nil {
			b.Fatalf("Failed to put record: %v", err)
		}
	}

	// Flush and close writer
	err = writer.Flush()
	if err != nil {
		b.Fatalf("Failed to flush writer: %v", err)
	}
	err = writer.Close()
	if err != nil {
		b.Fatalf("Failed to close writer: %v", err)
	}

	// Create log reader
	readerConfig := LogReaderConfig{
		FilePath:    tmpFile.Name(),
		StartOffset: 0,
	}

	reader, err := NewLogReader(readerConfig)
	if err != nil {
		b.Fatalf("Failed to create log reader: %v", err)
	}
	defer reader.Close()

	b.ResetTimer()
	b.ReportAllocs()

	b.Run("Read_Next_Sequential", func(b *testing.B) {
		// Reset reader to start
		err = reader.SeekTo(0)
		if err != nil {
			b.Fatalf("Failed to seek: %v", err)
		}

		for range b.N {
			_, err := reader.ReadNext()
			if err != nil {
				break // EOF reached
			}
		}
	})
}

// BenchmarkLogReader_ReadAt benchmarks the ReadAt operation
func BenchmarkLogReader_ReadAt(b *testing.B) {
	// Create temporary file with test data
	tmpFile, err := os.CreateTemp("", "freyja_log_reader_bench")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write test data using log writer
	writerConfig := LogWriterConfig{
		FilePath:      tmpFile.Name(),
		FsyncInterval: 0,         // Disable fsync for benchmarking
		BufferSize:    64 * 1024, // 64KB buffer
	}

	writer, err := NewLogWriter(writerConfig)
	if err != nil {
		b.Fatalf("Failed to create log writer: %v", err)
	}

	// Write multiple records and collect offsets
	offsets := make([]int64, 0, 1000)
	for i := range 1000 {
		// Get current offset before writing
		// Note: This is approximate since we don't have direct access to the file position
		key := []byte(fmt.Sprintf("key_%d", i))
		value := []byte(fmt.Sprintf("value_%d_with_some_additional_content_to_make_it_larger", i))
		_, err := writer.Put(key, value)
		if err != nil {
			b.Fatalf("Failed to put record: %v", err)
		}

		// Store offset for later reading
		if i%100 == 0 {
			// Approximate offset - in a real implementation we'd track this properly
			offsets = append(offsets, int64(i*100)) // Rough approximation
		}
	}

	// Flush and close writer
	err = writer.Flush()
	if err != nil {
		b.Fatalf("Failed to flush writer: %v", err)
	}
	err = writer.Close()
	if err != nil {
		b.Fatalf("Failed to close writer: %v", err)
	}

	// Create log reader
	readerConfig := LogReaderConfig{
		FilePath:    tmpFile.Name(),
		StartOffset: 0,
	}

	reader, err := NewLogReader(readerConfig)
	if err != nil {
		b.Fatalf("Failed to create log reader: %v", err)
	}
	defer reader.Close()

	b.ResetTimer()
	b.ReportAllocs()

	b.Run("Read_At_Specific_Offset", func(b *testing.B) {
		for i := range b.N {
			offsetIdx := i % len(offsets)
			offset := offsets[offsetIdx]

			_, err := reader.ReadAt(offset)
			if err != nil {
				// Reset to beginning if we hit EOF
				err = reader.SeekTo(0)
				if err != nil {
					b.Fatalf("Failed to seek: %v", err)
				}
				continue
			}
		}
	})
}

// BenchmarkLogReader_Iterator benchmarks the iterator operations
func BenchmarkLogReader_Iterator(b *testing.B) {
	// Create temporary file with test data
	tmpFile, err := os.CreateTemp("", "freyja_log_reader_bench")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write test data using log writer
	writerConfig := LogWriterConfig{
		FilePath:      tmpFile.Name(),
		FsyncInterval: 0,         // Disable fsync for benchmarking
		BufferSize:    64 * 1024, // 64KB buffer
	}

	writer, err := NewLogWriter(writerConfig)
	if err != nil {
		b.Fatalf("Failed to create log writer: %v", err)
	}

	// Write multiple records
	for i := range 10000 {
		key := []byte(fmt.Sprintf("key_%d", i))
		value := []byte(fmt.Sprintf("value_%d_with_some_additional_content", i))
		_, err := writer.Put(key, value)
		if err != nil {
			b.Fatalf("Failed to put record: %v", err)
		}
	}

	// Flush and close writer
	err = writer.Flush()
	if err != nil {
		b.Fatalf("Failed to flush writer: %v", err)
	}
	err = writer.Close()
	if err != nil {
		b.Fatalf("Failed to close writer: %v", err)
	}

	// Create log reader
	readerConfig := LogReaderConfig{
		FilePath:    tmpFile.Name(),
		StartOffset: 0,
	}

	reader, err := NewLogReader(readerConfig)
	if err != nil {
		b.Fatalf("Failed to create log reader: %v", err)
	}
	defer reader.Close()

	b.ResetTimer()
	b.ReportAllocs()

	b.Run("Iterator_Full_Scan", func(b *testing.B) {
		for range b.N {
			// Reset reader to start
			err = reader.SeekTo(0)
			if err != nil {
				b.Fatalf("Failed to seek: %v", err)
			}

			iterator := reader.Iterator()
			defer iterator.Close()

			count := 0
			for iterator.Next() {
				_ = iterator.Record()
				count++
				if count >= 10000 {
					break
				}
			}
		}
	})
}

// BenchmarkLogReader_Concurrent benchmarks concurrent read operations
func BenchmarkLogReader_Concurrent(b *testing.B) {
	// Create temporary file with test data
	tmpFile, err := os.CreateTemp("", "freyja_log_reader_bench")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write test data using log writer
	writerConfig := LogWriterConfig{
		FilePath:      tmpFile.Name(),
		FsyncInterval: 0,         // Disable fsync for benchmarking
		BufferSize:    64 * 1024, // 64KB buffer
	}

	writer, err := NewLogWriter(writerConfig)
	if err != nil {
		b.Fatalf("Failed to create log writer: %v", err)
	}

	// Write multiple records
	for i := range 10000 {
		key := []byte(fmt.Sprintf("key_%d", i))
		value := []byte(fmt.Sprintf("value_%d_with_some_additional_content", i))
		_, err := writer.Put(key, value)
		if err != nil {
			b.Fatalf("Failed to put record: %v", err)
		}
	}

	// Flush and close writer
	err = writer.Flush()
	if err != nil {
		b.Fatalf("Failed to flush writer: %v", err)
	}
	err = writer.Close()
	if err != nil {
		b.Fatalf("Failed to close writer: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.Run("Concurrent_Reads", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			// Each goroutine gets its own reader
			readerConfig := LogReaderConfig{
				FilePath:    tmpFile.Name(),
				StartOffset: 0,
			}

			reader, err := NewLogReader(readerConfig)
			if err != nil {
				b.Fatalf("Failed to create log reader: %v", err)
			}
			defer reader.Close()

			i := 0
			for pb.Next() {
				// Read at specific offset
				_, err := reader.ReadAt(int64((i % 1000) * 100))
				if err != nil {
					// Reset if we hit EOF
					err = reader.SeekTo(0)
					if err != nil {
						b.Fatalf("Failed to seek: %v", err)
					}
				}
				i++
			}
		})
	})
}
