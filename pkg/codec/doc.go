// Package codec provides record serialization and deserialization for FreyjaDB.
//
// The codec package implements a binary record format for storing key-value pairs
// with integrity checking and metadata. This is the foundation for FreyjaDB's
// log-structured storage engine.
//
// # Record Format
//
// Records are serialized in a binary format with the following structure:
//
//	[CRC32(4)][KeySize(4)][ValueSize(4)][Timestamp(8)][Key][Value]
//
// Fields:
//   - CRC32: 32-bit CRC checksum for integrity validation (little-endian)
//   - KeySize: 32-bit unsigned integer indicating key length in bytes (little-endian)
//   - ValueSize: 32-bit unsigned integer indicating value length in bytes (little-endian)
//   - Timestamp: 64-bit Unix timestamp in nanoseconds (little-endian)
//   - Key: Variable-length key data
//   - Value: Variable-length value data
//
// The total record size is: 20 bytes (header) + len(key) + len(value)
//
// # CRC32 Calculation
//
// The CRC32 checksum is calculated over all fields except the CRC32 field itself:
//   - KeySize (4 bytes)
//   - ValueSize (4 bytes)
//   - Timestamp (8 bytes)
//   - Key data (KeySize bytes)
//   - Value data (ValueSize bytes)
//
// This ensures that any corruption in the record header or data will be detected
// during validation.
//
// # Usage
//
// Basic encoding and decoding:
//
//	codec := codec.NewRecordCodec()
//
//	// Encode a record
//	encoded, err := codec.Encode([]byte("key"), []byte("value"))
//	if err != nil {
//	    return err
//	}
//
//	// Decode a record
//	record, err := codec.Decode(encoded)
//	if err != nil {
//	    return err
//	}
//
//	// Validate integrity
//	if err := record.Validate(); err != nil {
//	    return err // Record is corrupted
//	}
//
// # Error Handling
//
// The codec provides comprehensive error handling for:
//   - Malformed binary data (insufficient length, invalid headers)
//   - CRC32 validation failures (data corruption)
//   - Size mismatches between declared and actual data lengths
//
// All methods return descriptive errors that can be used for debugging
// and recovery scenarios.
//
// # Performance Considerations
//
// The codec is designed for high performance:
//   - Zero-copy decoding where possible
//   - Efficient CRC32 calculation using hardware acceleration when available
//   - Minimal memory allocations during encoding/decoding
//   - Streaming-friendly format for large values
//
// Benchmark your specific use case to understand performance characteristics.
// See the benchmark tests for performance examples with different data sizes.
//
// # Thread Safety
//
// RecordCodec instances are safe for concurrent use. Record structs are
// immutable after creation and safe to share between goroutines.
//
// # Compatibility
//
// The record format is designed to be stable and backwards-compatible.
// Future versions may add optional fields but will maintain compatibility
// with the current format for existing records.
package codec
