package store

import (
	"os"
	"path/filepath"
	"sync"

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

// Open initializes the store and loads existing data
func (kv *KVStore) Open() error {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	if kv.isOpen {
		return nil
	}

	// Create log writer
	writerConfig := LogWriterConfig{
		FilePath:      kv.dataFile,
		FsyncInterval: kv.config.FsyncInterval,
		BufferSize:    64 * 1024, // 64KB buffer
	}
	writer, err := NewLogWriter(writerConfig)
	if err != nil {
		return err
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
		return err
	}
	kv.reader = reader

	// Build index from existing data
	if err := kv.index.BuildFromLog(kv.reader); err != nil {
		kv.reader.Close()
		kv.writer.Close()
		return err
	}

	kv.isOpen = true
	return nil
}

// Get retrieves a value for a key
func (kv *KVStore) Get(key []byte) ([]byte, error) {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	if !kv.isOpen {
		return nil, &KVError{"store is not open"}
	}

	// Scan the log file sequentially to find the latest record for this key
	if err := kv.reader.Seek(0); err != nil {
		return nil, err
	}

	iterator := kv.reader.Iterator()
	defer iterator.Close()

	var latestValue []byte
	found := false

	for iterator.Next() {
		record := iterator.Record()
		if record == nil {
			continue
		}

		// Check if this is the key we're looking for
		if string(record.Key) == string(key) {
			found = true
			// Check if it's a tombstone (empty value indicates deletion)
			if len(record.Value) == 0 {
				latestValue = nil // Mark as deleted
			} else {
				latestValue = record.Value
			}
		}
	}

	if !found {
		return nil, ErrKeyNotFound
	}

	if latestValue == nil {
		return nil, ErrKeyNotFound
	}

	return latestValue, nil
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
	entry := &IndexEntry{
		FileID:    0, // Single file for now
		Offset:    offset - int64(codec.NewRecord(key, value).Size()),
		Size:      uint32(codec.NewRecord(key, value).Size()),
		Timestamp: codec.NewRecord(key, value).Timestamp,
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
