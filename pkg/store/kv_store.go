package store

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/ssargent/freyjadb/pkg/codec"
)

const errStoreNotOpen = "store is not open"

const (
	relationDirectionOutgoing = "outgoing"
	relationDirectionIncoming = "incoming"
	relationDirectionBoth     = "both"
)

// KVStore provides the main key-value store interface
type KVStore struct {
	config     KVStoreConfig
	writer     *LogWriter
	reader     *LogReader
	index      *HashIndex
	dataFile   string
	mutex      sync.Mutex
	isOpen     bool
	tombstones int64     // Track number of tombstones
	startTime  time.Time // Track when the store was opened
}

// NewKVStore creates a new key-value store instance
func NewKVStore(config KVStoreConfig) (*KVStore, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(config.DataDir, 0750); err != nil {
		return nil, err
	}

	dataFile := filepath.Join(config.DataDir, "active.data")

	store := &KVStore{
		config:     config,
		dataFile:   dataFile,
		index:      NewHashIndex(HashIndexConfig{}),
		isOpen:     false,
		tombstones: 0,
		startTime:  time.Time{}, // Will be set when opened
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
		if closeErr := kv.writer.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Error closing writer: %v\n", closeErr)
		}
		return nil, err
	}
	kv.reader = reader

	// Build index from validated data
	if err := kv.index.BuildFromLog(kv.reader); err != nil {
		if closeErr := kv.reader.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Error closing reader: %v\n", closeErr)
		}
		if closeErr := kv.writer.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Error closing writer: %v\n", closeErr)
		}
		return nil, err
	}

	kv.isOpen = true
	kv.startTime = time.Now()
	return recoveryResult, nil
}

// Get retrieves a value for a key
func (kv *KVStore) Get(key []byte) ([]byte, error) {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	if !kv.isOpen {
		return nil, &KVError{errStoreNotOpen}
	}

	// Use index for O(1) lookup
	entry, exists := kv.index.Get(key)
	if !exists {
		return nil, ErrKeyNotFound
	}

	// Force sync to ensure all buffered writes are on disk
	if err := kv.writer.Sync(); err != nil {
		return nil, err
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
		return &KVError{errStoreNotOpen}
	}

	if len(key) == 0 {
		return ErrInvalidKey
	}

	// Validate record size
	recordSize := len(key) + len(value)
	if kv.config.MaxRecordSize > 0 && recordSize > kv.config.MaxRecordSize {
		return ErrRecordSizeExceeded
	}

	// Write record to log
	offset, err := kv.writer.Put(key, value)
	if err != nil {
		return err
	}

	// Update index
	record := codec.NewRecord(key, value)
	entry := &IndexEntry{
		FileID:    0,                     // Single file for now
		Offset:    offset,                // LogWriter.Put() returns the starting offset
		Size:      uint32(record.Size()), //nolint: gosec // Size is uint32
		Timestamp: record.Timestamp,
	}
	kv.index.Put(key, entry)

	return nil
}

// deleteInternal removes a key-value pair without acquiring the mutex
// This is for internal use when the mutex is already held
func (kv *KVStore) deleteInternal(key []byte) error {
	if !kv.isOpen {
		return &KVError{errStoreNotOpen}
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
		return &KVError{errStoreNotOpen}
	}

	if len(key) == 0 {
		return ErrInvalidKey
	}

	// Validate record size
	recordSize := len(key) + len(value)
	if kv.config.MaxRecordSize > 0 && recordSize > kv.config.MaxRecordSize {
		return ErrRecordSizeExceeded
	}

	// Write record to log
	offset, err := kv.writer.Put(key, value)
	if err != nil {
		return err
	}

	// Update index
	record := codec.NewRecord(key, value)
	entry := &IndexEntry{
		FileID:    0,                     // Single file for now
		Offset:    offset,                // LogWriter.Put() returns the starting offset
		Size:      uint32(record.Size()), //nolint: gosec // Size is uint32
		Timestamp: record.Timestamp,
	}

	// Update index
	kv.index.Put(key, entry)

	return nil
}

// Delete removes a key-value pair (tombstone)

func (kv *KVStore) Delete(key []byte) error {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	if !kv.isOpen {
		return &KVError{errStoreNotOpen}
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

	// Increment tombstone counter
	kv.tombstones++

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
			_ = kv.reader.Close() // Ignore error in cleanup
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

	// Check if file exists and get initial stats
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, nothing to validate
			return kv.createEmptyRecoveryResult(startTime), nil
		}
		return nil, err
	}

	fileSizeBefore := fileInfo.Size()

	// Scan for corruption
	recordsValidated, lastValidOffset, corruptionFound, err := kv.scanForCorruption(filePath)
	if err != nil {
		return nil, err
	}

	// Handle corruption recovery if needed
	fileSizeAfter, recordsTruncated, err := kv.handleCorruptionRecovery(
		filePath, corruptionFound, lastValidOffset, fileSizeBefore)
	if err != nil {
		return nil, err
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

// createEmptyRecoveryResult creates a recovery result for non-existent files
func (kv *KVStore) createEmptyRecoveryResult(startTime time.Time) *RecoveryResult {
	return &RecoveryResult{
		RecordsValidated: 0,
		RecordsTruncated: 0,
		FileSizeBefore:   0,
		FileSizeAfter:    0,
		IndexRebuilt:     true,
		RecoveryTime:     time.Since(startTime).Nanoseconds(),
	}
}

// scanForCorruption scans the log file for corruption and returns validation results
func (kv *KVStore) scanForCorruption(filePath string) (int64, int64, bool, error) {
	reader, err := NewLogReader(LogReaderConfig{
		FilePath:    filePath,
		StartOffset: 0,
	})
	if err != nil {
		return 0, -1, false, err
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			// Log error but don't fail the operation
			fmt.Fprintf(os.Stderr, "Error closing reader: %v\n", closeErr)
		}
	}()

	var recordsValidated int64
	var lastValidOffset int64 = -1
	var corruptionFound bool
	var tombstones int64 // Track tombstones during recovery

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

		// Check if it's a tombstone (empty value indicates deletion)
		if len(record.Value) == 0 {
			tombstones++
		}

		recordsValidated++
		lastValidOffset = reader.Offset()
	}

	// Store tombstone count in the KVStore (this will be used during recovery)
	// Note: This is a bit of a hack - in a real implementation we'd want a cleaner way to pass this back
	kv.tombstones = tombstones

	return recordsValidated, lastValidOffset, corruptionFound, nil
}

// handleCorruptionRecovery handles file truncation when corruption is detected
func (kv *KVStore) handleCorruptionRecovery(
	filePath string,
	corruptionFound bool,
	lastValidOffset int64,
	fileSizeBefore int64,
) (int64, int64, error) {
	fileSizeAfter := fileSizeBefore
	var recordsTruncated int64

	if corruptionFound && lastValidOffset >= 0 {
		err := kv.truncateCorruptedFile(filePath, lastValidOffset)
		if err != nil {
			return 0, 0, err
		}
		fileSizeAfter = lastValidOffset
		recordsTruncated = 1 // We assume one corrupted record at the end
	}

	return fileSizeAfter, recordsTruncated, nil
}

// truncateCorruptedFile truncates the file to remove corrupted records
func (kv *KVStore) truncateCorruptedFile(filePath string, offset int64) error {
	cleanPath := filepath.Clean(filePath)
	file, err := os.OpenFile(cleanPath, os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log error but don't fail the operation
			fmt.Fprintf(os.Stderr, "Error closing file: %v\n", closeErr)
		}
	}()

	if err := file.Truncate(offset); err != nil {
		return err
	}

	return nil
}

// Stats returns store statistics
func (kv *KVStore) Stats() *StoreStats {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	if !kv.isOpen {
		return &StoreStats{}
	}

	return &StoreStats{
		Keys:       kv.index.Size(),
		DataSize:   kv.writer.Size(),
		Tombstones: int(kv.tombstones),
	}
}

// StoreStats holds statistics about the store
type StoreStats struct {
	Keys       int
	DataSize   int64
	Tombstones int
}

/*
	res.Global.TotalKeys = kv.index.Size() + int(kv.tombstones)
	res.Global.ActiveKeys = kv.index.Size()
	res.Global.Tombstones = int(kv.tombstones)
	res.Global.TotalSizeMB = float64(kv.writer.Size()) / (1024 * 1024)

	// Calculate live size by scanning through records and summing non-tombstone values
	// This is a simplified approximation - in a production system you'd want a more efficient approach
	res.Global.LiveSizeMB = res.Global.TotalSizeMB * 0.9 // Placeholder - would need actual calculation

	res.Global.Uptime = time.Since(kv.startTime)
	res.Global.IndexMemoryMB = 0 // TODO: Estimate index memory

	// Segments (stub for now)
	res.Segments = []Segment{
		{ID: "active", Keys: kv.index.Size(), DeadPct: 0.0, SizeMB: res.Global.TotalSizeMB},
	}

	// Partitions (stub)
	res.Partitions = map[string]PKStats{}
*/
/*
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
		// TODO: Collect actual metrics
		res.Metrics.QueriesPerSecond = 0
		res.Metrics.CacheHitRate = 0
		res.Metrics.AvgQueryLatency = 0
	}
*/

/*
return res, nil
}
*/

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
		return nil, &KVError{errStoreNotOpen}
	}

	prefixStr := string(prefix)
	return kv.index.KeysWithPrefix(prefixStr), nil
}

// ScanPrefix returns a channel of key-value pairs that match the prefix.
// Callers must range over the channel to completion; abandoning it partway
// through will leak the background goroutine.
func (kv *KVStore) ScanPrefix(prefix []byte) (<-chan KeyValuePair, error) {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	if !kv.isOpen {
		return nil, &KVError{errStoreNotOpen}
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

			ch <- KeyValuePair{Key: key, Value: record.Value}
		}
	}()

	return ch, nil
}

// listKeysInternal returns all keys that match the given prefix without acquiring the mutex
// This is for internal use when the mutex is already held
func (kv *KVStore) listKeysInternal(prefix []byte) ([]string, error) {
	if !kv.isOpen {
		return nil, &KVError{errStoreNotOpen}
	}

	prefixStr := string(prefix)
	return kv.index.KeysWithPrefix(prefixStr), nil
}

// PutRelationship creates a relationship between two entities
func (kv *KVStore) PutRelationship(fromKey, toKey, relation string) error {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	if !kv.isOpen {
		return &KVError{errStoreNotOpen}
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
	if putErr := kv.putInternal([]byte(forwardKey), forwardData); putErr != nil {
		return fmt.Errorf("failed to store forward relationship: %w", putErr)
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
		return &KVError{errStoreNotOpen}
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
		return nil, &KVError{errStoreNotOpen}
	}

	limit := query.Limit
	if limit == 0 {
		limit = 100 // Default limit
	}

	var results []RelationshipResult
	var err error

	if query.Direction == relationDirectionOutgoing || query.Direction == relationDirectionBoth {
		results, err = kv.queryRelationshipsByDirection(results, "forward", relationDirectionOutgoing, query, limit)
		if err != nil {
			return nil, err
		}
	}

	if query.Direction == relationDirectionIncoming || query.Direction == relationDirectionBoth {
		results, err = kv.queryRelationshipsByDirection(results, "reverse", relationDirectionIncoming, query, limit)
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

// queryRelationshipsByDirection lists and decodes relationships for a single direction
// (forward/outgoing or reverse/incoming). GetRelationships previously duplicated this
// batch-read-with-fallback logic once per direction; both directions share it here.
func (kv *KVStore) queryRelationshipsByDirection(
	results []RelationshipResult, prefixKind, direction string, query RelationshipQuery, limit int,
) ([]RelationshipResult, error) {
	safeKey := strings.ReplaceAll(query.Key, ":", "|")
	prefix := fmt.Sprintf("relationship:%s:%s", prefixKind, safeKey)
	if query.Relation != "" {
		prefix += fmt.Sprintf(":%s", query.Relation)
	}

	keys, err := kv.listKeysInternal([]byte(prefix))
	if err != nil {
		return nil, fmt.Errorf("failed to list %s relationships: %w", direction, err)
	}

	// Process relationships in batches to reduce I/O overhead
	const batchSize = 50
	for i := 0; i < len(keys) && len(results) < limit; i += batchSize {
		end := min(i+batchSize, len(keys))

		// Convert string keys to byte keys for batch operation
		byteKeys := make([][]byte, end-i)
		for j, key := range keys[i:end] {
			byteKeys[j] = []byte(key)
		}

		batchData, batchErr := kv.getBatchInternal(byteKeys)
		if batchErr != nil {
			results = kv.appendRelationshipsIndividually(results, keys[i:end], direction, limit)
		} else {
			results = appendRelationshipsFromBatch(results, batchData, direction, limit)
		}
	}

	return results, nil
}

// appendRelationshipsIndividually is the fallback path used when a batch read fails.
func (kv *KVStore) appendRelationshipsIndividually(
	results []RelationshipResult, keys []string, direction string, limit int,
) []RelationshipResult {
	for _, key := range keys {
		if len(results) >= limit {
			break
		}

		data, err := kv.getInternal([]byte(key))
		if err != nil {
			continue // Skip if can't read
		}

		if rel, ok := decodeRelationship(data); ok {
			results = append(results, newRelationshipResult(rel, direction))
		}
	}
	return results
}

func appendRelationshipsFromBatch(
	results []RelationshipResult, batchData [][]byte, direction string, limit int,
) []RelationshipResult {
	for _, data := range batchData {
		if len(results) >= limit {
			break
		}

		if data == nil {
			continue // Skip if no data
		}

		if rel, ok := decodeRelationship(data); ok {
			results = append(results, newRelationshipResult(rel, direction))
		}
	}
	return results
}

func decodeRelationship(data []byte) (*Relationship, bool) {
	var rel Relationship
	if err := json.Unmarshal(data, &rel); err != nil {
		return nil, false // Skip if can't parse
	}
	return &rel, true
}

func newRelationshipResult(rel *Relationship, direction string) RelationshipResult {
	otherKey := rel.ToKey
	if direction == relationDirectionIncoming {
		otherKey = rel.FromKey
	}
	return RelationshipResult{
		Relationship: rel,
		OtherKey:     otherKey,
		Direction:    direction,
	}
}

// getBatchInternal retrieves multiple values for keys in a single operation
// This is for internal use when the mutex is already held
func (kv *KVStore) getBatchInternal(keys [][]byte) ([][]byte, error) {
	if !kv.isOpen {
		return nil, &KVError{errStoreNotOpen}
	}

	results := make([][]byte, len(keys))

	// Process keys in batches to reduce I/O overhead
	for i, key := range keys {
		// Use index for O(1) lookup
		entry, exists := kv.index.Get(key)
		if !exists {
			results[i] = nil
			continue
		}

		// Read record directly from the stored offset
		record, err := kv.reader.ReadAt(entry.Offset)
		if err != nil {
			results[i] = nil
			continue
		}

		// Check if it's a tombstone (empty value indicates deletion)
		if len(record.Value) == 0 {
			results[i] = nil
			continue
		}

		results[i] = record.Value
	}

	return results, nil
}

// This is for internal use when the mutex is already held
func (kv *KVStore) getInternal(key []byte) ([]byte, error) {
	if !kv.isOpen {
		return nil, &KVError{errStoreNotOpen}
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

// Explain gathers stats about the key-value store
func (kv *KVStore) Explain(ctx context.Context, opts ExplainOptions) (*ExplainResult, error) {
	res := &ExplainResult{}

	// Lock the mutex to safely access store data
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	// Global stats
	totalKeys := kv.index.Size()
	res.Global.TotalKeys = totalKeys
	res.Global.ActiveKeys = totalKeys // For simplicity, assume all keys are active
	res.Global.Tombstones = int(kv.tombstones)
	res.Global.Uptime = time.Since(kv.startTime)

	// Calculate file sizes
	var totalSize int64
	if info, err := os.Stat(kv.dataFile); err == nil {
		totalSize = info.Size()
	}
	res.Global.TotalSizeMB = float64(totalSize) / (1024 * 1024)
	res.Global.LiveSizeMB = res.Global.TotalSizeMB // For simplicity

	// Memory usage
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	res.Global.IndexMemoryMB = float64(m.Alloc) / (1024 * 1024)

	// Segments (simplified representation)
	res.Segments = []Segment{
		{ID: "active", Keys: totalKeys, DeadPct: 0.0, SizeMB: res.Global.TotalSizeMB},
	}

	// Partitions (empty for now)
	res.Partitions = make(map[string]PKStats)

	// Diagnostics
	if opts.WithSamples > 0 {
		// Simplified sampling - just indicate we support it
		res.Diagnostics.Samples = []Sample{}
	}

	// Metrics
	if opts.WithMetrics {
		// Placeholder metrics
		res.Diagnostics.Metrics.AvgGetLatencyMs = 0.5
		res.Diagnostics.Metrics.IORateMBs = 10.0
	}

	// PK filtering
	if opts.PK != "" {
		// Filter partitions by PK
		// This is a placeholder implementation
		stats := PKStats{
			Keys:        0,
			SKRanges:    []SKRange{},
			Cardinality: "unknown",
		}
		res.Partitions[opts.PK] = stats
	}

	return res, nil
}
