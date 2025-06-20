package codec

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"time"
)

// Record represents a key-value record with metadata for storage
type Record struct {
	CRC32     uint32 // CRC32 checksum for integrity
	KeySize   uint32 // Size of the key in bytes
	ValueSize uint32 // Size of the value in bytes
	Timestamp uint64 // Unix timestamp in nanoseconds
	Key       []byte // Key data
	Value     []byte // Value data
}

// RecordCodec handles serialization and deserialization of records
type RecordCodec struct{}

// NewRecordCodec creates a new record codec instance
func NewRecordCodec() *RecordCodec {
	return &RecordCodec{}
}

// Encode serializes a key-value pair into a binary record format
// Format: [CRC32(4)][KeySize(4)][ValueSize(4)][Timestamp(8)][Key][Value]
func (c *RecordCodec) Encode(key, value []byte) ([]byte, error) {
	// TODO: Implement record encoding
	// 1. Create record with current timestamp
	// 2. Calculate sizes
	// 3. Serialize header + data
	// 4. Calculate and set CRC32
	// 5. Return complete binary record
	return nil, fmt.Errorf("not implemented")
}

// Decode deserializes a binary record into a Record struct
func (c *RecordCodec) Decode(data []byte) (*Record, error) {
	// TODO: Implement record decoding
	// 1. Validate minimum size
	// 2. Parse header fields
	// 3. Extract key and value data
	// 4. Validate CRC32
	// 5. Return Record struct
	return nil, fmt.Errorf("not implemented")
}

// Validate checks the integrity of a record using CRC32
func (r *Record) Validate() error {
	// TODO: Implement CRC32 validation
	// 1. Recalculate CRC32 for record data
	// 2. Compare with stored CRC32
	// 3. Return error if mismatch
	return fmt.Errorf("not implemented")
}

// Size returns the total size of the record when encoded
func (r *Record) Size() int {
	// Header: CRC32(4) + KeySize(4) + ValueSize(4) + Timestamp(8) = 20 bytes
	// Data: len(Key) + len(Value)
	return 20 + len(r.Key) + len(r.Value)
}

// NewRecord creates a new record with current timestamp
func NewRecord(key, value []byte) *Record {
	return &Record{
		KeySize:   uint32(len(key)),
		ValueSize: uint32(len(value)),
		Timestamp: uint64(time.Now().UnixNano()),
		Key:       key,
		Value:     value,
	}
}

// calculateCRC32 computes CRC32 checksum for record data (excluding the CRC field itself)
func (r *Record) calculateCRC32() uint32 {
	// TODO: Implement CRC32 calculation
	// Calculate checksum over: KeySize + ValueSize + Timestamp + Key + Value
	crc := crc32.NewIEEE()

	// Write header fields (excluding CRC32)
	binary.Write(crc, binary.LittleEndian, r.KeySize)
	binary.Write(crc, binary.LittleEndian, r.ValueSize)
	binary.Write(crc, binary.LittleEndian, r.Timestamp)

	// Write data
	crc.Write(r.Key)
	crc.Write(r.Value)

	return crc.Sum32()
}
