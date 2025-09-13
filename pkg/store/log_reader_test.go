package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogReader(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "log_reader_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.log")

	// Create a test file
	err = os.WriteFile(filePath, []byte("test data"), 0600)
	require.NoError(t, err)

	config := LogReaderConfig{
		FilePath: filePath,
	}

	reader, err := NewLogReader(config)
	require.NoError(t, err)
	assert.NotNil(t, reader)

	err = reader.Close()
	assert.NoError(t, err)
}

func TestNewLogReader_NonExistentFile(t *testing.T) {
	config := LogReaderConfig{
		FilePath: "/non/existent/file.log",
	}

	reader, err := NewLogReader(config)
	assert.Error(t, err)
	assert.Nil(t, reader)
}

func TestNewLogReader_WithStartOffset(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "log_reader_offset_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.log")

	// Create a test file with some data
	testData := []byte("0123456789abcdef")
	err = os.WriteFile(filePath, testData, 0600)
	require.NoError(t, err)

	config := LogReaderConfig{
		FilePath:    filePath,
		StartOffset: 5,
	}

	reader, err := NewLogReader(config)
	require.NoError(t, err)
	assert.NotNil(t, reader)
	assert.Equal(t, int64(5), reader.Offset())

	err = reader.Close()
	assert.NoError(t, err)
}

func TestLogReader_ReadNext_EOF(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "log_reader_eof_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "empty.log")

	// Create an empty file
	err = os.WriteFile(filePath, []byte{}, 0600)
	require.NoError(t, err)

	config := LogReaderConfig{
		FilePath: filePath,
	}

	reader, err := NewLogReader(config)
	require.NoError(t, err)
	defer reader.Close()

	// Reading from empty file should return EOF or corruption error
	record, err := reader.ReadNext()
	assert.Nil(t, record)
	assert.Error(t, err) // Should be some kind of error (EOF or corruption)
}

func TestLogReader_Seek(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "log_reader_seek_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.log")

	// Create a test file
	testData := []byte("0123456789abcdef")
	err = os.WriteFile(filePath, testData, 0600)
	require.NoError(t, err)

	config := LogReaderConfig{
		FilePath: filePath,
	}

	reader, err := NewLogReader(config)
	require.NoError(t, err)
	defer reader.Close()

	// Initial offset should be 0
	assert.Equal(t, int64(0), reader.Offset())

	// Seek to position 5
	err = reader.Seek(5)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), reader.Offset())
}

func TestLogReader_Iterator(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "log_reader_iterator_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.log")

	// Create an empty file for this test
	err = os.WriteFile(filePath, []byte{}, 0600)
	require.NoError(t, err)

	config := LogReaderConfig{
		FilePath: filePath,
	}

	reader, err := NewLogReader(config)
	require.NoError(t, err)
	defer reader.Close()

	iterator := reader.Iterator()
	assert.NotNil(t, iterator)

	// Empty file should not have records
	assert.False(t, iterator.Next())

	err = iterator.Close()
	assert.NoError(t, err)
}

func TestLogReader_ReadAt(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "log_reader_readat_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.log")

	// Create a test file
	testData := []byte("0123456789abcdef")
	err = os.WriteFile(filePath, testData, 0600)
	require.NoError(t, err)

	config := LogReaderConfig{
		FilePath: filePath,
	}

	reader, err := NewLogReader(config)
	require.NoError(t, err)
	defer reader.Close()

	// Read at offset 5
	record, err := reader.ReadAt(5)
	// This will likely fail because the data doesn't represent a valid record format
	// But we can test that the function doesn't panic and handles the error
	if err != nil {
		assert.Contains(t, err.Error(), "corruption")
	}
	assert.Nil(t, record)
}

func TestLogReader_MultipleOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "log_reader_multi_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.log")

	// Create a test file
	testData := []byte("0123456789abcdef")
	err = os.WriteFile(filePath, testData, 0600)
	require.NoError(t, err)

	config := LogReaderConfig{
		FilePath: filePath,
	}

	reader, err := NewLogReader(config)
	require.NoError(t, err)
	defer reader.Close()

	// Test multiple seeks
	err = reader.Seek(0)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), reader.Offset())

	err = reader.Seek(10)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), reader.Offset())

	err = reader.Seek(5)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), reader.Offset())
}

func TestLogReaderConfig_Validation(t *testing.T) {
	// Test with empty file path
	config := LogReaderConfig{
		FilePath: "",
	}

	reader, err := NewLogReader(config)
	assert.Error(t, err)
	assert.Nil(t, reader)
}

func BenchmarkLogReader_Seek(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "log_reader_bench_seek")
	require.NoError(b, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.log")

	// Create a test file
	testData := make([]byte, 1024*1024) // 1MB file
	for i := range testData {
		testData[i] = byte(i % 256)
	}
	err = os.WriteFile(filePath, testData, 0600)
	require.NoError(b, err)

	config := LogReaderConfig{
		FilePath: filePath,
	}

	reader, err := NewLogReader(config)
	require.NoError(b, err)
	defer reader.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		offset := int64(i % 1000000)
		if err := reader.Seek(offset); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLogReader_ReadAt(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "log_reader_bench_readat")
	require.NoError(b, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.log")

	// Create a test file
	testData := make([]byte, 1024*1024) // 1MB file
	for i := range testData {
		testData[i] = byte(i % 256)
	}
	err = os.WriteFile(filePath, testData, 0600)
	require.NoError(b, err)

	config := LogReaderConfig{
		FilePath: filePath,
	}

	reader, err := NewLogReader(config)
	require.NoError(b, err)
	defer reader.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		offset := int64(i % 1000000)
		if _, err := reader.ReadAt(offset); err != nil {
			// Ignore error in benchmark
		}
	}
}
