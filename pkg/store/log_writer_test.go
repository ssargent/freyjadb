package store

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogWriter(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "log_writer_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.log")

	config := LogWriterConfig{
		FilePath:      filePath,
		FsyncInterval: 0, // Immediate fsync
		BufferSize:    4096,
	}

	writer, err := NewLogWriter(config)
	require.NoError(t, err)
	assert.NotNil(t, writer)

	// Verify file was created
	assert.FileExists(t, filePath)

	// Verify initial size is 0
	assert.Equal(t, int64(0), writer.Size())

	err = writer.Close()
	assert.NoError(t, err)
}

func TestNewLogWriter_DirectoryCreation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "log_writer_dir_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	nestedDir := filepath.Join(tmpDir, "nested", "deep", "path")
	filePath := filepath.Join(nestedDir, "test.log")

	config := LogWriterConfig{
		FilePath:      filePath,
		FsyncInterval: 0,
		BufferSize:    4096,
	}

	writer, err := NewLogWriter(config)
	require.NoError(t, err)
	assert.NotNil(t, writer)

	// Verify directory was created
	assert.DirExists(t, nestedDir)

	err = writer.Close()
	assert.NoError(t, err)
}

func TestNewLogWriter_InvalidPath(t *testing.T) {
	config := LogWriterConfig{
		FilePath:      "/invalid/path/that/cannot/be/created/test.log",
		FsyncInterval: 0,
		BufferSize:    4096,
	}

	writer, err := NewLogWriter(config)
	assert.Error(t, err)
	assert.Nil(t, writer)
}

func TestLogWriter_Put(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "log_writer_put_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.log")

	config := LogWriterConfig{
		FilePath:      filePath,
		FsyncInterval: 0, // Immediate fsync
		BufferSize:    4096,
	}

	writer, err := NewLogWriter(config)
	require.NoError(t, err)
	defer writer.Close()

	key := []byte("test_key")
	value := []byte("test_value")

	offset, err := writer.Put(key, value)
	require.NoError(t, err)

	// Offset should be 0 for first record
	assert.Equal(t, int64(0), offset)

	// Size should be greater than 0
	assert.Greater(t, writer.Size(), int64(0))
}

func TestLogWriter_MultiplePuts(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "log_writer_multi_put_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.log")

	config := LogWriterConfig{
		FilePath:      filePath,
		FsyncInterval: 0,
		BufferSize:    4096,
	}

	writer, err := NewLogWriter(config)
	require.NoError(t, err)
	defer writer.Close()

	// Put multiple records
	records := []struct {
		key   []byte
		value []byte
	}{
		{[]byte("key1"), []byte("value1")},
		{[]byte("key2"), []byte("value2")},
		{[]byte("key3"), []byte("value3")},
	}

	var offsets []int64
	for _, record := range records {
		offset, err := writer.Put(record.key, record.value)
		require.NoError(t, err)
		offsets = append(offsets, offset)
	}

	// Offsets should be increasing
	assert.Equal(t, int64(0), offsets[0])
	assert.Greater(t, offsets[1], offsets[0])
	assert.Greater(t, offsets[2], offsets[1])

	// File size should be reasonable
	assert.Greater(t, writer.Size(), int64(50))
}

func TestLogWriter_Sync(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "log_writer_sync_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.log")

	config := LogWriterConfig{
		FilePath:      filePath,
		FsyncInterval: time.Hour, // Long interval to prevent auto-sync
		BufferSize:    4096,
	}

	writer, err := NewLogWriter(config)
	require.NoError(t, err)
	defer writer.Close()

	// Put a record
	_, err = writer.Put([]byte("key"), []byte("value"))
	require.NoError(t, err)

	// Manually sync
	err = writer.Sync()
	assert.NoError(t, err)
}

func TestLogWriter_FsyncInterval(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "log_writer_fsync_interval_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.log")

	config := LogWriterConfig{
		FilePath:      filePath,
		FsyncInterval: 10 * time.Millisecond, // Short interval
		BufferSize:    4096,
	}

	writer, err := NewLogWriter(config)
	require.NoError(t, err)
	defer writer.Close()

	// Put a record - should trigger fsync after interval
	_, err = writer.Put([]byte("key"), []byte("value"))
	require.NoError(t, err)

	// Wait for fsync timer
	time.Sleep(50 * time.Millisecond)
}

func TestLogWriter_Path(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "log_writer_path_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.log")

	config := LogWriterConfig{
		FilePath:      filePath,
		FsyncInterval: 0,
		BufferSize:    4096,
	}

	writer, err := NewLogWriter(config)
	require.NoError(t, err)
	defer writer.Close()

	assert.Equal(t, filePath, writer.Path())
}

func TestLogWriter_Size(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "log_writer_size_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.log")

	config := LogWriterConfig{
		FilePath:      filePath,
		FsyncInterval: 0,
		BufferSize:    4096,
	}

	writer, err := NewLogWriter(config)
	require.NoError(t, err)
	defer writer.Close()

	// Initial size should be 0
	assert.Equal(t, int64(0), writer.Size())

	// Put a record
	_, err = writer.Put([]byte("key"), []byte("value"))
	require.NoError(t, err)

	// Size should increase
	initialSize := writer.Size()
	assert.Greater(t, initialSize, int64(0))

	// Put another record
	_, err = writer.Put([]byte("key2"), []byte("value2"))
	require.NoError(t, err)

	// Size should increase further
	finalSize := writer.Size()
	assert.Greater(t, finalSize, initialSize)
}

func TestLogWriter_BufferSize(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "log_writer_buffer_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.log")

	config := LogWriterConfig{
		FilePath:      filePath,
		FsyncInterval: 0,
		BufferSize:    1024, // Small buffer
	}

	writer, err := NewLogWriter(config)
	require.NoError(t, err)
	defer writer.Close()

	// Put a large record that exceeds buffer size
	largeValue := make([]byte, 2000)
	for i := range largeValue {
		largeValue[i] = byte(i % 256)
	}

	_, err = writer.Put([]byte("large_key"), largeValue)
	assert.NoError(t, err)
}

func TestLogWriter_ConcurrentAccess(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "log_writer_concurrent_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.log")

	config := LogWriterConfig{
		FilePath:      filePath,
		FsyncInterval: time.Hour, // Disable auto-fsync
		BufferSize:    4096,
	}

	writer, err := NewLogWriter(config)
	require.NoError(t, err)
	defer writer.Close()

	done := make(chan bool, 2)

	// Goroutine 1: Write operations
	go func() {
		for i := 0; i < 100; i++ {
			key := []byte(fmt.Sprintf("key_%d", i))
			value := []byte(fmt.Sprintf("value_%d", i))
			writer.Put(key, value)
		}
		done <- true
	}()

	// Goroutine 2: Sync operations
	go func() {
		for i := 0; i < 10; i++ {
			writer.Sync()
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done
}

func BenchmarkLogWriter_Put(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "log_writer_bench_put")
	require.NoError(b, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.log")

	config := LogWriterConfig{
		FilePath:      filePath,
		FsyncInterval: time.Hour, // Disable auto-fsync for benchmark
		BufferSize:    4096,
	}

	writer, err := NewLogWriter(config)
	require.NoError(b, err)
	defer writer.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("bench_key_%d", i))
		value := []byte(fmt.Sprintf("bench_value_%d", i))
		writer.Put(key, value)
	}
}

func BenchmarkLogWriter_PutWithFsync(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "log_writer_bench_fsync")
	require.NoError(b, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.log")

	config := LogWriterConfig{
		FilePath:      filePath,
		FsyncInterval: 0, // Immediate fsync
		BufferSize:    4096,
	}

	writer, err := NewLogWriter(config)
	require.NoError(b, err)
	defer writer.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("bench_key_%d", i))
		value := []byte(fmt.Sprintf("bench_value_%d", i))
		writer.Put(key, value)
	}
}
