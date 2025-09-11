package index

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/segmentio/ksuid"
	"github.com/ssargent/freyjadb/pkg/bptree"
)

// SecondaryIndex manages a B+Tree-based index for a specific field
type SecondaryIndex struct {
	fieldName string
	tree      *bptree.BPlusTree
	mutex     sync.RWMutex
}

// NewSecondaryIndex creates a new secondary index for a field
func NewSecondaryIndex(fieldName string, order int) *SecondaryIndex {
	return &SecondaryIndex{
		fieldName: fieldName,
		tree:      bptree.NewBPlusTree(order),
	}
}

// Insert adds a record to the secondary index
// The index key is: field_value + primary_key (to ensure uniqueness)
func (idx *SecondaryIndex) Insert(fieldValue interface{}, primaryKey []byte) error {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	indexKey := idx.createIndexKey(fieldValue, primaryKey)
	// Create a KSUID from the primary key bytes for the index value
	ksuidValue, err := ksuid.FromBytes(primaryKey)
	if err != nil {
		return fmt.Errorf("failed to create KSUID from primary key: %w", err)
	}
	idx.tree.Insert(indexKey, ksuidValue)
	return nil
}

// Delete removes a record from the secondary index
func (idx *SecondaryIndex) Delete(fieldValue interface{}, primaryKey []byte) bool {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	indexKey := idx.createIndexKey(fieldValue, primaryKey)
	return idx.tree.Delete(indexKey)
}

// Search finds records with exact field value match
func (idx *SecondaryIndex) Search(fieldValue interface{}) ([][]byte, error) {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	// For exact match, we need to find all keys that start with the field value
	_ = idx.createFieldPrefix(fieldValue) // TODO: Use for prefix search

	// TODO: Implement prefix search in B+Tree
	// For now, return empty (this needs B+Tree range query support)
	return [][]byte{}, nil
}

// SearchRange finds records within a field value range
func (idx *SecondaryIndex) SearchRange(startValue, endValue interface{}) ([][]byte, error) {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	// TODO: Implement range search in B+Tree
	// This requires extending the B+Tree with range query capabilities
	return [][]byte{}, nil
}

// Save persists the index to disk
func (idx *SecondaryIndex) Save(dir string) error {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	filename := filepath.Join(dir, fmt.Sprintf("index_%s.dat", idx.fieldName))
	return idx.tree.Save(filename)
}

// Load restores the index from disk
func (idx *SecondaryIndex) Load(dir string) error {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	filename := filepath.Join(dir, fmt.Sprintf("index_%s.dat", idx.fieldName))
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// Index doesn't exist yet, keep empty tree
		return nil
	}

	tree, err := bptree.LoadBPlusTree(filename)
	if err != nil {
		return fmt.Errorf("failed to load index for field %s: %w", idx.fieldName, err)
	}

	idx.tree = tree
	return nil
}

// createIndexKey creates a composite key: field_value + primary_key
func (idx *SecondaryIndex) createIndexKey(fieldValue interface{}, primaryKey []byte) []byte {
	var buf bytes.Buffer

	// Serialize field value
	idx.serializeValue(&buf, fieldValue)

	// Append primary key
	buf.Write(primaryKey)

	return buf.Bytes()
}

// createFieldPrefix creates a key prefix for field value matching
func (idx *SecondaryIndex) createFieldPrefix(fieldValue interface{}) []byte {
	var buf bytes.Buffer
	idx.serializeValue(&buf, fieldValue)
	return buf.Bytes()
}

// serializeValue serializes different value types for indexing
func (idx *SecondaryIndex) serializeValue(buf *bytes.Buffer, value interface{}) {
	switch v := value.(type) {
	case int:
		buf.WriteByte(0) // Type marker for int
		binary.Write(buf, binary.BigEndian, int64(v))
	case int64:
		buf.WriteByte(0)
		binary.Write(buf, binary.BigEndian, v)
	case float64:
		buf.WriteByte(1) // Type marker for float64
		binary.Write(buf, binary.BigEndian, v)
	case string:
		buf.WriteByte(2) // Type marker for string
		buf.WriteString(v)
		buf.WriteByte(0) // Null terminator
	default:
		// For unknown types, convert to string
		buf.WriteByte(2)
		buf.WriteString(fmt.Sprintf("%v", v))
		buf.WriteByte(0)
	}
}

// IndexManager manages multiple secondary indexes for a partition
type IndexManager struct {
	indexes map[string]*SecondaryIndex
	mutex   sync.RWMutex
	order   int
}

// NewIndexManager creates a new index manager
func NewIndexManager(order int) *IndexManager {
	return &IndexManager{
		indexes: make(map[string]*SecondaryIndex),
		order:   order,
	}
}

// GetOrCreateIndex gets an existing index or creates a new one for a field
func (im *IndexManager) GetOrCreateIndex(fieldName string) *SecondaryIndex {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	if idx, exists := im.indexes[fieldName]; exists {
		return idx
	}

	idx := NewSecondaryIndex(fieldName, im.order)
	im.indexes[fieldName] = idx
	return idx
}

// SaveAll saves all indexes to disk
func (im *IndexManager) SaveAll(dir string) error {
	im.mutex.RLock()
	defer im.mutex.RUnlock()

	for _, idx := range im.indexes {
		if err := idx.Save(dir); err != nil {
			return err
		}
	}
	return nil
}

// LoadAll loads all indexes from disk
func (im *IndexManager) LoadAll(dir string) error {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	// Find all index files
	pattern := filepath.Join(dir, "index_*.dat")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	for _, file := range files {
		filename := filepath.Base(file)
		if len(filename) < 10 { // "index_.dat" is 10 chars minimum
			continue
		}

		// Extract field name from filename
		fieldName := filename[6 : len(filename)-4] // Remove "index_" prefix and ".dat" suffix

		idx := NewSecondaryIndex(fieldName, im.order)
		if err := idx.Load(dir); err != nil {
			return err
		}

		im.indexes[fieldName] = idx
	}

	return nil
}
