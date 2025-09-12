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
	r := NewRecord(key, value)
	r.CRC32 = r.calculateCRC32()

	buf := make([]byte, r.Size())

	binary.LittleEndian.PutUint32(buf[0:], r.CRC32)
	binary.LittleEndian.PutUint32(buf[4:], r.KeySize)
	binary.LittleEndian.PutUint32(buf[8:], r.ValueSize)
	binary.LittleEndian.PutUint64(buf[12:], r.Timestamp)
	copy(buf[20:], r.Key)
	copy(buf[20+r.KeySize:], r.Value)

	return buf, nil
}

// Decode deserializes a binary record into a Record struct
func (c *RecordCodec) Decode(data []byte) (*Record, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("data too short for record header")
	}

	r := &Record{}
	r.CRC32 = binary.LittleEndian.Uint32(data[0:4])
	r.KeySize = binary.LittleEndian.Uint32(data[4:8])
	r.ValueSize = binary.LittleEndian.Uint32(data[8:12])
	r.Timestamp = binary.LittleEndian.Uint64(data[12:20])
	// Validate sizes
	if len(data) < int(20+r.KeySize+r.ValueSize) {
		return nil, fmt.Errorf("data too short for key/value sizes: %d < %d", len(data), 20+r.KeySize+r.ValueSize)
	}

	r.Key = data[20 : 20+r.KeySize]
	r.Value = data[20+r.KeySize : 20+r.KeySize+r.ValueSize]

	return r, nil
}

// Validate checks the integrity of a record using CRC32
func (r *Record) Validate() error {
	if r.CRC32 != r.calculateCRC32() {
		return fmt.Errorf("CRC32 mismatch: %d != %d", r.CRC32, r.calculateCRC32())
	}

	return nil
}

// Size returns the total size of the record when encoded
func (r *Record) Size() int {
	// Header: CRC32(4) + KeySize(4) + ValueSize(4) + Timestamp(8) = 20 bytes
	// Data: len(Key) + len(Value)
	return 20 + len(r.Key) + len(r.Value)
}

// NewRecord creates a new record with current timestamp
func NewRecord(key, value []byte) *Record {
	keyLen := len(key)
	valLen := len(value)
	if keyLen > int(^uint32(0)) {
		panic("key too large")
	}
	if valLen > int(^uint32(0)) {
		panic("value too large")
	}
	return &Record{
		KeySize:   uint32(keyLen),
		ValueSize: uint32(valLen),
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
	if err := binary.Write(crc, binary.LittleEndian, r.KeySize); err != nil {
		return 0
	}
	if err := binary.Write(crc, binary.LittleEndian, r.ValueSize); err != nil {
		return 0
	}
	if err := binary.Write(crc, binary.LittleEndian, r.Timestamp); err != nil {
		return 0
	}

	// Write data
	if _, err := crc.Write(r.Key); err != nil {
		return 0
	}
	if _, err := crc.Write(r.Value); err != nil {
		return 0
	}

	return crc.Sum32()
}
