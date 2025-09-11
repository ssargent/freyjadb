package store

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ssargent/freyjadb/pkg/codec"
)

// KVStore provides the main key-value store interface
type KVStore struct {
	config   KVStoreConfig
	writer   *LogWriter
	reader   *LogReader
	index    *HashIndex
	dataFile string
	mutex    sync.Mutex
	isOpen   bool
}

// NewKVStore creates a new key-value store instance
func NewKVStore(config KVStoreConfig) (*KVStore, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		return nil, err
	}

	dataFile := filepath.Join(config.DataDir, "active.data")

	store := &KVStore{
		config:   config,
		dataFile: dataFile,
		index:    NewHashIndex(HashIndexConfig{}),
		isOpen:   false,
	}

	return store, nil
}

// Open initializes the store and loads existing data with crash recovery
func (kv *KVStore) Open() (*RecoveryResult, error) {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	if kv.isOpen {
		return &RecoveryResult{
			RecordsValidated: 0,
			RecordsTruncated: 0,
			FileSizeBefore:   0,
			FileSizeAfter:    0,
			IndexRebuilt:     false,
			RecoveryTime:     0,
		}, nil
	}

	// Validate log file and recover from corruption
	recoveryResult, err := kv.validateLogFile(kv.dataFile)
	if err != nil {
		return nil, err
	}

	// Create log writer
	writerConfig := LogWriterConfig{
		FilePath:      kv.dataFile,
		FsyncInterval: kv.config.FsyncInterval,
		BufferSize:    64 * 1024, // 64KB buffer
	}
	writer, err := NewLogWriter(writerConfig)
	if err != nil {
		return nil, err
	}
	kv.writer = writer

	// Create log reader
	readerConfig := LogReaderConfig{
		FilePath:    kv.dataFile,
		StartOffset: 0,
	}
	reader, err := NewLogReader(readerConfig)
	if err != nil {
		kv.writer.Close()
		return nil, err
	}
	kv.reader = reader

	// Build index from validated data
	if err := kv.index.BuildFromLog(kv.reader); err != nil {
		kv.reader.Close()
		kv.writer.Close()
		return nil, err
	}

	kv.isOpen = true
	return recoveryResult, nil
}

// Get retrieves a value for a key
func (kv *KVStore) Get(key []byte) ([]byte, error) {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	if !kv.isOpen {
		return nil, &KVError{"store is not open"}
	}

	// Use index for O(1) lookup
	entry, exists := kv.index.Get(key)
	if !exists {
		return nil, ErrKeyNotFound
	}

	// Read record directly from the stored offset
	record, err := kv.reader.ReadAt(entry.Offset)
	if err != nil {
		return nil, err
	}

	// Check if it's a tombstone (empty value indicates deletion)
	if len(record.Value) == 0 {
		return nil, ErrKeyNotFound
	}

	return record.Value, nil
}

// putInternal stores a key-value pair without acquiring the mutex
// This is for internal use when the mutex is already held
func (kv *KVStore) putInternal(key, value []byte) error {
	if !kv.isOpen {
		return &KVError{"store is not open"}
	}

	if len(key) == 0 {
		return ErrInvalidKey
	}

	// Write record to log
	offset, err := kv.writer.Put(key, value)
	if err != nil {
		return err
	}

	// Update index
	record := codec.NewRecord(key, value)
	entry := &IndexEntry{
		FileID:    0,      // Single file for now
		Offset:    offset, // LogWriter.Put() returns the starting offset
		Size:      uint32(record.Size()),
		Timestamp: record.Timestamp,
	}
	kv.index.Put(key, entry)

	return nil
}

// deleteInternal removes a key-value pair without acquiring the mutex
// This is for internal use when the mutex is already held
func (kv *KVStore) deleteInternal(key []byte) error {
	if !kv.isOpen {
		return &KVError{"store is not open"}
	}

	if len(key) == 0 {
		return ErrInvalidKey
	}

	// Write tombstone record (empty value)
	_, err := kv.writer.Put(key, []byte{})
	if err != nil {
		return err
	}

	// Remove from index
	kv.index.Delete(key)

	return nil
}

// Put stores a key-value pair
func (kv *KVStore) Put(key, value []byte) error {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	if !kv.isOpen {
		return &KVError{"store is not open"}
	}

	if len(key) == 0 {
		return ErrInvalidKey
	}

	// Write record to log
	offset, err := kv.writer.Put(key, value)
	if err != nil {
		return err
	}

	// Update index
	record := codec.NewRecord(key, value)
	entry := &IndexEntry{
		FileID:    0,      // Single file for now
		Offset:    offset, // LogWriter.Put() returns the starting offset
		Size:      uint32(record.Size()),
		Timestamp: record.Timestamp,
	}
	kv.index.Put(key, entry)

	return nil
}

// Delete removes a key-value pair (tombstone)
func (kv *KVStore) Delete(key []byte) error {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	if !kv.isOpen {
		return &KVError{"store is not open"}
	}

	if len(key) == 0 {
		return ErrInvalidKey
	}

	// Write tombstone record (empty value)
	_, err := kv.writer.Put(key, []byte{})
	if err != nil {
		return err
	}

	// Remove from index
	kv.index.Delete(key)

	return nil
}

// Close shuts down the store
func (kv *KVStore) Close() error {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	if !kv.isOpen {
		return nil
	}

	kv.isOpen = false

	// Close writer first (ensures all data is flushed)
	if kv.writer != nil {
		if err := kv.writer.Close(); err != nil {
			kv.reader.Close()
			return err
		}
	}

	// Close reader
	if kv.reader != nil {
		if err := kv.reader.Close(); err != nil {
			return err
		}
	}

	return nil
}

// validateLogFile validates the log file integrity and truncates corrupted records
func (kv *KVStore) validateLogFile(filePath string) (*RecoveryResult, error) {
	startTime := time.Now()

	// Get file size before validation
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, nothing to validate
			return &RecoveryResult{
				RecordsValidated: 0,
				RecordsTruncated: 0,
				FileSizeBefore:   0,
				FileSizeAfter:    0,
				IndexRebuilt:     true,
				RecoveryTime:     time.Since(startTime).Nanoseconds(),
			}, nil
		}
		return nil, err
	}

	fileSizeBefore := fileInfo.Size()

	// Create a temporary reader for validation
	reader, err := NewLogReader(LogReaderConfig{
		FilePath:    filePath,
		StartOffset: 0,
	})
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var recordsValidated int64
	var lastValidOffset int64 = -1
	var corruptionFound bool

	// Read through the file until we find corruption
	for {
		record, err := reader.ReadNext()
		if err != nil {
			if err == io.EOF {
				break // End of file reached
			}
			// Corruption detected
			corruptionFound = true
			break
		}

		// Validate CRC
		if err := record.Validate(); err != nil {
			corruptionFound = true
			break
		}

		recordsValidated++
		lastValidOffset = reader.Offset()
	}

	// If corruption was found, truncate the file
	var fileSizeAfter int64 = fileSizeBefore
	var recordsTruncated int64

	if corruptionFound && lastValidOffset >= 0 {
		// Truncate the file to the last valid record
		file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
		if err != nil {
			return nil, err
		}

		if err := file.Truncate(lastValidOffset); err != nil {
			file.Close()
			return nil, err
		}

		file.Close()
		fileSizeAfter = lastValidOffset
		recordsTruncated = 1 // We assume one corrupted record at the end
	}

	return &RecoveryResult{
		RecordsValidated: recordsValidated,
		RecordsTruncated: recordsTruncated,
		FileSizeBefore:   fileSizeBefore,
		FileSizeAfter:    fileSizeAfter,
		IndexRebuilt:     true,
		RecoveryTime:     time.Since(startTime).Nanoseconds(),
	}, nil
}

// Stats returns store statistics
func (kv *KVStore) Stats() *StoreStats {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	if !kv.isOpen {
		return &StoreStats{}
	}

	return &StoreStats{
		Keys:     kv.index.Size(),
		DataSize: kv.writer.Size(),
	}
}

// StoreStats holds statistics about the store
type StoreStats struct {
	Keys     int
	DataSize int64
}

// Explain gathers diagnostic information about the store
func (kv *KVStore) Explain(ctx context.Context, opts ExplainOptions) (*ExplainResult, error) {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	if !kv.isOpen {
		return nil, &KVError{"store is not open"}
	}

	res := &ExplainResult{}
	res.Global.TotalKeys = kv.index.Size()
	res.Global.ActiveKeys = kv.index.Size() // TODO: Subtract tombstones
	res.Global.Tombstones = 0               // TODO: Count tombstones
	res.Global.TotalSizeMB = float64(kv.writer.Size()) / (1024 * 1024)
	res.Global.LiveSizeMB = res.Global.TotalSizeMB // TODO: Calculate live size
	res.Global.Uptime = time.Since(time.Now())     // TODO: Track start time
	res.Global.IndexMemoryMB = 0                   // TODO: Estimate index memory

	// Segments (stub for now)
	res.Segments = []Segment{
		{ID: "active", Keys: kv.index.Size(), DeadPct: 0.0, SizeMB: res.Global.TotalSizeMB},
	}

	// Partitions (stub)
	res.Partitions = map[string]PKStats{}

	// Samples
	if opts.WithSamples > 0 {
		// TODO: Sample actual records
		res.Diagnostics.Samples = []Sample{}
	}

	// Warnings
	if opts.PK != "" {
		res.Warnings = append(res.Warnings, fmt.Sprintf("Partition filtering not implemented for PK: %s", opts.PK))
	}

	res.Diagnostics.CRCErrors = 0

	if opts.WithMetrics {
		res.Diagnostics.Metrics.AvgGetLatencyMs = 0 // TODO: Track metrics
		res.Diagnostics.Metrics.IORateMBs = 0
	}

	return res, nil
}

// KeyValuePair represents a key-value pair for scanning operations
type KeyValuePair struct {
	Key   []byte
	Value []byte
}

// ListKeys returns all keys that match the given prefix
func (kv *KVStore) ListKeys(prefix []byte) ([]string, error) {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	if !kv.isOpen {
		return nil, &KVError{"store is not open"}
	}

	prefixStr := string(prefix)
	return kv.index.KeysWithPrefix(prefixStr), nil
}

// ScanPrefix returns a channel of key-value pairs that match the prefix
func (kv *KVStore) ScanPrefix(prefix []byte) (<-chan KeyValuePair, error) {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	if !kv.isOpen {
		return nil, &KVError{"store is not open"}
	}

	ch := make(chan KeyValuePair, 100)

	go func() {
		defer close(ch)

		prefixStr := string(prefix)
		keyChan := kv.index.ScanPrefix(prefixStr)

		for keyStr := range keyChan {
			// Get the value for this key
			key := []byte(keyStr)
			entry, exists := kv.index.Get(key)
			if !exists {
				continue // Key was deleted while scanning
			}

			// Read the record from disk
			record, err := kv.reader.ReadAt(entry.Offset)
			if err != nil {
				continue // Skip corrupted records
			}

			// Skip tombstones
			if len(record.Value) == 0 {
				continue
			}

			select {
			case ch <- KeyValuePair{Key: key, Value: record.Value}:
			case <-ch: // Channel closed by receiver
				return
			}
		}
	}()

	return ch, nil
}

// listKeysInternal returns all keys that match the given prefix without acquiring the mutex
// This is for internal use when the mutex is already held
func (kv *KVStore) listKeysInternal(prefix []byte) ([]string, error) {
	if !kv.isOpen {
		return nil, &KVError{"store is not open"}
	}

	prefixStr := string(prefix)
	return kv.index.KeysWithPrefix(prefixStr), nil
}

// PutRelationship creates a relationship between two entities
func (kv *KVStore) PutRelationship(fromKey, toKey, relation string) error {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	if !kv.isOpen {
		return &KVError{"store is not open"}
	}

	// Validate that both entities exist
	if err := kv.validateRelationshipKeys(fromKey, toKey); err != nil {
		return err
	}

	// Create relationship object
	relationship := &Relationship{
		FromKey:   fromKey,
		ToKey:     toKey,
		Relation:  relation,
		CreatedAt: time.Now(),
	}

	// Store forward relationship
	forwardKey := makeRelationshipKey("forward", fromKey, relation, toKey)
	forwardData, err := json.Marshal(relationship)
	if err != nil {
		return fmt.Errorf("failed to marshal relationship: %w", err)
	}
	if err := kv.putInternal([]byte(forwardKey), forwardData); err != nil {
		return fmt.Errorf("failed to store forward relationship: %w", err)
	}

	// Store reverse relationship
	reverseKey := makeRelationshipKey("reverse", toKey, relation, fromKey)
	reverseData, err := json.Marshal(relationship)
	if err != nil {
		return fmt.Errorf("failed to marshal reverse relationship: %w", err)
	}
	if err := kv.putInternal([]byte(reverseKey), reverseData); err != nil {
		return fmt.Errorf("failed to store reverse relationship: %w", err)
	}

	return nil
}

// DeleteRelationship removes a relationship between two entities
func (kv *KVStore) DeleteRelationship(fromKey, toKey, relation string) error {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	if !kv.isOpen {
		return &KVError{"store is not open"}
	}

	// Delete forward relationship
	forwardKey := makeRelationshipKey("forward", fromKey, relation, toKey)
	if err := kv.deleteInternal([]byte(forwardKey)); err != nil && err != ErrKeyNotFound {
		return fmt.Errorf("failed to delete forward relationship: %w", err)
	}

	// Delete reverse relationship
	reverseKey := makeRelationshipKey("reverse", toKey, relation, fromKey)
	if err := kv.deleteInternal([]byte(reverseKey)); err != nil && err != ErrKeyNotFound {
		return fmt.Errorf("failed to delete reverse relationship: %w", err)
	}

	return nil
}

// GetRelationships returns all relationships for a given key
func (kv *KVStore) GetRelationships(query RelationshipQuery) ([]RelationshipResult, error) {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	if !kv.isOpen {
		return nil, &KVError{"store is not open"}
	}

	var results []RelationshipResult
	limit := query.Limit
	if limit == 0 {
		limit = 100 // Default limit
	}

	// Query outgoing relationships
	if query.Direction == "outgoing" || query.Direction == "both" {
		safeKey := strings.ReplaceAll(query.Key, ":", "|")
		prefix := fmt.Sprintf("relationship:forward:%s", safeKey)
		if query.Relation != "" {
			prefix += fmt.Sprintf(":%s", query.Relation)
		}

		keys, err := kv.listKeysInternal([]byte(prefix))
		if err != nil {
			return nil, fmt.Errorf("failed to list outgoing relationships: %w", err)
		}

		for _, key := range keys {
			if len(results) >= limit {
				break
			}

			data, err := kv.getInternal([]byte(key))
			if err != nil {
				continue // Skip if can't read
			}

			var rel Relationship
			if err := json.Unmarshal(data, &rel); err != nil {
				continue // Skip if can't parse
			}

			results = append(results, RelationshipResult{
				Relationship: &rel,
				OtherKey:     rel.ToKey,
				Direction:    "outgoing",
			})
		}
	}

	// Query incoming relationships
	if query.Direction == "incoming" || query.Direction == "both" {
		safeKey := strings.ReplaceAll(query.Key, ":", "|")
		prefix := fmt.Sprintf("relationship:reverse:%s", safeKey)
		if query.Relation != "" {
			prefix += fmt.Sprintf(":%s", query.Relation)
		}

		keys, err := kv.listKeysInternal([]byte(prefix))
		if err != nil {
			return nil, fmt.Errorf("failed to list incoming relationships: %w", err)
		}

		for _, key := range keys {
			if len(results) >= limit {
				break
			}

			data, err := kv.getInternal([]byte(key))
			if err != nil {
				continue // Skip if can't read
			}

			var rel Relationship
			if err := json.Unmarshal(data, &rel); err != nil {
				continue // Skip if can't parse
			}

			results = append(results, RelationshipResult{
				Relationship: &rel,
				OtherKey:     rel.FromKey,
				Direction:    "incoming",
			})
		}
	}

	return results, nil
}

// getInternal retrieves a value for a key without acquiring the mutex
// This is for internal use when the mutex is already held
func (kv *KVStore) getInternal(key []byte) ([]byte, error) {
	if !kv.isOpen {
		return nil, &KVError{"store is not open"}
	}

	// Use index for O(1) lookup
	entry, exists := kv.index.Get(key)
	if !exists {
		return nil, ErrKeyNotFound
	}

	// Read record directly from the stored offset
	record, err := kv.reader.ReadAt(entry.Offset)
	if err != nil {
		return nil, err
	}

	// Check if it's a tombstone (empty value indicates deletion)
	if len(record.Value) == 0 {
		return nil, ErrKeyNotFound
	}

	return record.Value, nil
}
