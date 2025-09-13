package store

import (
	"time"

	"github.com/ssargent/freyjadb/pkg/codec"
)

// IndexEntry represents the location of a key-value pair in the log
type IndexEntry struct {
	FileID    uint32 // ID of the data file
	Offset    int64  // Byte offset within the file
	Size      uint32 // Size of the record in bytes
	Timestamp uint64 // Record timestamp
}

// LogWriterConfig holds configuration for the log writer
type LogWriterConfig struct {
	FilePath      string        // Path to the active data file
	FsyncInterval time.Duration // How often to fsync (0 = every write)
	BufferSize    int           // Write buffer size
}

// LogReaderConfig holds configuration for the log reader
type LogReaderConfig struct {
	FilePath    string // Path to the data file
	StartOffset int64  // Offset to start reading from
}

// HashIndexConfig holds configuration for the hash index
type HashIndexConfig struct {
	// Future: max memory, persistence options, etc.
}

// KVStoreConfig holds configuration for the key-value store
type KVStoreConfig struct {
	DataDir       string        // Directory for data files
	FsyncInterval time.Duration // Fsync interval for durability
	MaxRecordSize int           // Maximum size of a single record in bytes
}

// RecoveryResult holds statistics about crash recovery operations
type RecoveryResult struct {
	RecordsValidated int64 // Number of records successfully validated
	RecordsTruncated int64 // Number of corrupted records truncated
	FileSizeBefore   int64 // File size before recovery
	FileSizeAfter    int64 // File size after recovery
	IndexRebuilt     bool  // Whether index was rebuilt
	RecoveryTime     int64 // Time taken for recovery in nanoseconds
}

// RecordIterator provides streaming access to records
type RecordIterator interface {
	Next() bool
	Record() *codec.Record
	Close() error
}

// Errors
var (
	ErrKeyNotFound        = &KVError{"key not found"}
	ErrInvalidKey         = &KVError{"invalid key"}
	ErrCorruption         = &KVError{"data corruption detected"}
	ErrRecordSizeExceeded = &KVError{"record size exceeds maximum allowed size"}
)

// KVError represents a key-value store error
type KVError struct {
	Message string
}

func (e *KVError) Error() string {
	return e.Message
}
