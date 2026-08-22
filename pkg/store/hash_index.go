package store

import (
	"fmt"
	"math"
	"strings"
	"sync"
)

// HashIndex provides O(1) average-case lookups for key locations
type HashIndex struct {
	entries map[string]*IndexEntry
	mutex   sync.RWMutex
}

// NewHashIndex creates a new hash index
func NewHashIndex(config HashIndexConfig) *HashIndex {
	return &HashIndex{
		entries: make(map[string]*IndexEntry),
	}
}

// Put adds or updates an index entry for a key
func (idx *HashIndex) Put(key []byte, entry *IndexEntry) {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	keyStr := string(key)
	idx.entries[keyStr] = entry
}

// Get retrieves the index entry for a key
func (idx *HashIndex) Get(key []byte) (*IndexEntry, bool) {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	entry, exists := idx.entries[string(key)]
	return entry, exists
}

// Delete removes a key from the index
func (idx *HashIndex) Delete(key []byte) {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	keyStr := string(key)
	delete(idx.entries, keyStr)
}

// Size returns the number of keys in the index
func (idx *HashIndex) Size() int {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	return len(idx.entries)
}

// Clear removes all entries from the index
func (idx *HashIndex) Clear() {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	idx.entries = make(map[string]*IndexEntry)
}

// Keys returns all keys in the index (for debugging/testing)
func (idx *HashIndex) Keys() []string {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	keys := make([]string, 0, len(idx.entries))
	for key := range idx.entries {
		keys = append(keys, key)
	}
	return keys
}

// KeysWithPrefix returns all keys that start with the given prefix
func (idx *HashIndex) KeysWithPrefix(prefix string) []string {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	var keys []string
	for key := range idx.entries {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	return keys
}

// ScanPrefix returns a channel of keys that match the prefix
func (idx *HashIndex) ScanPrefix(prefix string) <-chan string {
	idx.mutex.RLock()
	keys := make([]string, 0, len(idx.entries))

	// Collect matching keys
	for key := range idx.entries {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	idx.mutex.RUnlock()

	// Buffer sized to the full result set: every send below is
	// non-blocking, so no goroutine or cancellation channel is needed.
	ch := make(chan string, len(keys))
	for _, key := range keys {
		ch <- key
	}
	close(ch)

	return ch
}

// BuildFromLog scans a log file and populates the index
func (idx *HashIndex) BuildFromLog(reader *LogReader) error {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	// Clear existing entries
	idx.entries = make(map[string]*IndexEntry)

	// Reset reader to beginning
	if err := reader.SeekTo(0); err != nil {
		return err
	}

	iterator := reader.Iterator()
	defer func() { _ = iterator.Close() }() // Ignore error in cleanup

	for iterator.Next() {
		record := iterator.Record()
		if record == nil {
			continue
		}

		size := record.Size()
		if size < 0 || size > math.MaxUint32 {
			return fmt.Errorf("record size %d exceeds uint32 range", size)
		}

		keyStr := string(record.Key)
		entry := &IndexEntry{
			FileID:    0, // Single file for now
			Offset:    reader.Offset() - int64(size),
			Size:      uint32(size),
			Timestamp: record.Timestamp,
		}

		// Handle tombstones (empty value indicates deletion)
		if len(record.Value) == 0 {
			delete(idx.entries, keyStr)
		} else {
			idx.entries[keyStr] = entry
		}
	}

	return nil
}

// Stats returns index statistics
func (idx *HashIndex) Stats() *IndexStats {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	return &IndexStats{
		TotalKeys:  len(idx.entries),
		Tombstones: 0, // TODO: Implement tombstone counting
	}
}

// IndexStats holds statistics about the index
type IndexStats struct {
	TotalKeys  int
	Tombstones int
}
