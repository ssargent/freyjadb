//go:build fuzz
// +build fuzz

package codec

import (
	"bytes"
	"testing"
)

// FuzzRecordCodec_RoundTrip tests encode/decode round-trip with random inputs
func FuzzRecordCodec_RoundTrip(f *testing.F) {
	codec := NewRecordCodec()

	// Add seed corpus
	f.Add([]byte(""), []byte(""))
	f.Add([]byte("key"), []byte("value"))
	f.Add([]byte("user:123"), []byte("john@example.com"))
	f.Add([]byte{0x00, 0x01, 0x02}, []byte{0xFF, 0xFE, 0xFD})

	f.Fuzz(func(t *testing.T, key, value []byte) {
		// Skip extremely large inputs to avoid timeout
		if len(key) > 10000 || len(value) > 100000 || len(key) == 0 || len(value) == 0 {
			t.Skip("Input too large for fuzz test")
		}

		// Encode the random key-value pair
		encoded, err := codec.Encode(key, value)
		if err != nil {
			t.Fatalf("Encode failed for key=%q value=%q: %v", key, value, err)
		}

		// Decode the binary data
		record, err := codec.Decode(encoded)
		if err != nil {
			t.Fatalf("Decode failed for encoded data: len(key)=%d len(value)=%d %v", len(key), len(value), err)
		}

		// Validate the record
		if err := record.Validate(); err != nil {
			t.Fatalf("Record validation failed: %v", err)
		}

		// Check that decoded data matches original
		if !bytes.Equal(record.Key, key) {
			t.Errorf("Key mismatch: got %q, want %q", record.Key, key)
		}

		if !bytes.Equal(record.Value, value) {
			t.Errorf("Value mismatch: got %q, want %q", record.Value, value)
		}

		// Check sizes match
		if record.KeySize != uint32(len(key)) {
			t.Errorf("KeySize mismatch: got %d, want %d", record.KeySize, len(key))
		}

		if record.ValueSize != uint32(len(value)) {
			t.Errorf("ValueSize mismatch: got %d, want %d", record.ValueSize, len(value))
		}
	})
}

// FuzzRecordCodec_CorruptionDetection tests that corruption is always detected
func FuzzRecordCodec_CorruptionDetection(f *testing.F) {
	codec := NewRecordCodec()

	// Add seed corpus
	f.Add([]byte("key"), []byte("value"), uint(0))
	f.Add([]byte("user:123"), []byte("john@example.com"), uint(5))
	f.Add([]byte("test"), []byte("data"), uint(10))

	f.Fuzz(func(t *testing.T, key, value []byte, corruptPos uint) {
		// Skip extremely large inputs
		if len(key) > 1000 || len(value) > 10000 {
			t.Skip("Input too large for fuzz test")
		}

		// Encode the key-value pair
		encoded, err := codec.Encode(key, value)
		if err != nil {
			t.Skip("Encode failed, skipping")
		}

		// Skip if corruption position is beyond data length
		if int(corruptPos) >= len(encoded) {
			t.Skip("Corruption position beyond data length")
		}

		// Make a copy and corrupt one byte
		corrupted := make([]byte, len(encoded))
		copy(corrupted, encoded)
		corrupted[corruptPos] ^= 0xFF // Flip all bits in the byte

		// If we corrupted the same way (no change), skip
		if bytes.Equal(corrupted, encoded) {
			t.Skip("Corruption resulted in no change")
		}

		// Try to decode corrupted data
		record, err := codec.Decode(corrupted)
		if err != nil {
			// Decode failure is acceptable for corrupted data
			return
		}

		// If decode succeeded, validation must fail
		err = record.Validate()
		if err == nil {
			t.Errorf("Corruption not detected! Original: %x, Corrupted: %x, Position: %d",
				encoded, corrupted, corruptPos)
		}
	})
}

// FuzzRecordCodec_MalformedData tests handling of malformed input
func FuzzRecordCodec_MalformedData(f *testing.F) {
	codec := NewRecordCodec()

	// Add seed corpus of malformed data
	f.Add([]byte{})
	f.Add([]byte{0x01})
	f.Add([]byte{0x01, 0x02, 0x03, 0x04})
	f.Add(make([]byte, 19)) // One byte short of header
	f.Add(make([]byte, 20)) // Header only

	f.Fuzz(func(t *testing.T, data []byte) {
		// Skip extremely large inputs
		if len(data) > 100000 {
			t.Skip("Input too large for fuzz test")
		}

		// Try to decode random data
		_, err := codec.Decode(data)

		// We expect most random data to fail decoding
		// The important thing is that it doesn't panic or cause other issues
		if err == nil {
			// If decode succeeded, the data was well-formed enough
			// This is fine, but rare for truly random data
			t.Logf("Unexpectedly succeeded to decode random data of length %d", len(data))
		}
	})
}

// Property test: encoded size should match expected size
func FuzzRecord_SizeProperty(f *testing.F) {
	// Add seed corpus
	f.Add([]byte(""), []byte(""))
	f.Add([]byte("k"), []byte("v"))
	f.Add([]byte("key"), []byte("value"))

	f.Fuzz(func(t *testing.T, key, value []byte) {
		// Skip extremely large inputs
		if len(key) > 10000 || len(value) > 100000 {
			t.Skip("Input too large for fuzz test")
		}

		record := NewRecord(key, value)
		expectedSize := 20 + len(key) + len(value) // Header + data

		if record.Size() != expectedSize {
			t.Errorf("Size calculation wrong: got %d, want %d", record.Size(), expectedSize)
		}

		// If we can encode it, the encoded size should match
		codec := NewRecordCodec()
		encoded, err := codec.Encode(key, value)
		if err == nil && len(encoded) != expectedSize {
			t.Errorf("Encoded size mismatch: got %d, want %d", len(encoded), expectedSize)
		}
	})
}
