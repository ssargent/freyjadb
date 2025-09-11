package codec

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"
)

func TestRecordCodec_EncodeDecodeRoundTrip(t *testing.T) {
	codec := NewRecordCodec()

	testCases := []struct {
		name  string
		key   []byte
		value []byte
	}{
		{
			name:  "simple string key-value",
			key:   []byte("user:123"),
			value: []byte("john@example.com"),
		},
		{
			name:  "empty key",
			key:   []byte(""),
			value: []byte("some value"),
		},
		{
			name:  "empty value",
			key:   []byte("some key"),
			value: []byte(""),
		},
		{
			name:  "both empty",
			key:   []byte(""),
			value: []byte(""),
		},
		{
			name:  "binary data",
			key:   []byte{0x00, 0x01, 0x02, 0x03},
			value: []byte{0xFF, 0xFE, 0xFD, 0xFC},
		},
		{
			name:  "large key",
			key:   bytes.Repeat([]byte("k"), 1024),
			value: []byte("small value"),
		},
		{
			name:  "large value",
			key:   []byte("small key"),
			value: bytes.Repeat([]byte("v"), 10240),
		},
		{
			name:  "unicode data",
			key:   []byte("🔑 unicode key"),
			value: []byte("🎯 unicode value with émojis"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Encode the key-value pair
			encoded, err := codec.Encode(tc.key, tc.value)
			if err != nil {
				t.Fatalf("Encode failed: %v", err)
			}

			// Decode the binary data
			record, err := codec.Decode(encoded)
			if err != nil {
				t.Fatalf("Decode failed: %v", err)
			}

			// Validate the record
			if err := record.Validate(); err != nil {
				t.Fatalf("Record validation failed: %v", err)
			}

			// Check that decoded data matches original
			if !bytes.Equal(record.Key, tc.key) {
				t.Errorf("Key mismatch: got %v, want %v", record.Key, tc.key)
			}

			if !bytes.Equal(record.Value, tc.value) {
				t.Errorf("Value mismatch: got %v, want %v", record.Value, tc.value)
			}

			// Check sizes
			if record.KeySize != uint32(len(tc.key)) {
				t.Errorf("KeySize mismatch: got %d, want %d", record.KeySize, len(tc.key))
			}

			if record.ValueSize != uint32(len(tc.value)) {
				t.Errorf("ValueSize mismatch: got %d, want %d", record.ValueSize, len(tc.value))
			}

			// Check timestamp is reasonable (within last minute)
			now := time.Now().UnixNano()
			if record.Timestamp > uint64(now) || record.Timestamp < uint64(now-int64(time.Minute)) {
				t.Errorf("Timestamp seems unreasonable: %d", record.Timestamp)
			}
		})
	}
}

func TestRecordCodec_CRCValidation(t *testing.T) {
	codec := NewRecordCodec()

	t.Run("valid CRC passes validation", func(t *testing.T) {
		key := []byte("test key")
		value := []byte("test value")

		encoded, err := codec.Encode(key, value)
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}

		record, err := codec.Decode(encoded)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		if err := record.Validate(); err != nil {
			t.Errorf("Valid record failed validation: %v", err)
		}
	})

	t.Run("corrupted CRC fails validation", func(t *testing.T) {
		key := []byte("test key")
		value := []byte("test value")

		encoded, err := codec.Encode(key, value)
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}

		// Corrupt the CRC32 field (first 4 bytes)
		encoded[0] ^= 0xFF

		record, err := codec.Decode(encoded)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		// Validation should fail due to CRC mismatch
		if err := record.Validate(); err == nil {
			t.Error("Expected validation to fail for corrupted CRC, but it passed")
		}
	})

	t.Run("corrupted key data fails validation", func(t *testing.T) {
		key := []byte("test key")
		value := []byte("test value")

		encoded, err := codec.Encode(key, value)
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}

		// Corrupt key data (after header, first byte of key)
		if len(encoded) > 20 {
			encoded[20] ^= 0xFF
		}

		record, err := codec.Decode(encoded)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		// Validation should fail due to CRC mismatch
		if err := record.Validate(); err == nil {
			t.Error("Expected validation to fail for corrupted key data, but it passed")
		}
	})

	t.Run("corrupted value data fails validation", func(t *testing.T) {
		key := []byte("test key")
		value := []byte("test value")

		encoded, err := codec.Encode(key, value)
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}

		// Corrupt value data (after header + key)
		valueOffset := 20 + len(key)
		if len(encoded) > valueOffset {
			encoded[valueOffset] ^= 0xFF
		}

		record, err := codec.Decode(encoded)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		// Validation should fail due to CRC mismatch
		if err := record.Validate(); err == nil {
			t.Error("Expected validation to fail for corrupted value data, but it passed")
		}
	})
}

func TestRecordCodec_MalformedData(t *testing.T) {
	codec := NewRecordCodec()

	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "too short for header",
			data: []byte{0x01, 0x02, 0x03},
		},
		{
			name: "insufficient data for declared key size",
			data: func() []byte {
				buf := make([]byte, 20)
				binary.LittleEndian.PutUint32(buf[4:8], 100) // KeySize = 100
				binary.LittleEndian.PutUint32(buf[8:12], 0)  // ValueSize = 0
				// But only 20 bytes total, can't fit 100-byte key
				return buf
			}(),
		},
		{
			name: "insufficient data for declared value size",
			data: func() []byte {
				buf := make([]byte, 25)                       // 20 header + 5 key bytes
				binary.LittleEndian.PutUint32(buf[4:8], 5)    // KeySize = 5
				binary.LittleEndian.PutUint32(buf[8:12], 100) // ValueSize = 100
				// But only 25 bytes total, can't fit 100-byte value
				return buf
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := codec.Decode(tc.data)
			if err == nil {
				t.Errorf("Expected decode to fail for malformed data, but it succeeded (%s)", tc.name)
			}
		})
	}
}

func TestRecord_Size(t *testing.T) {
	testCases := []struct {
		name         string
		key          []byte
		value        []byte
		expectedSize int
	}{
		{
			name:         "empty key and value",
			key:          []byte(""),
			value:        []byte(""),
			expectedSize: 20, // Header only
		},
		{
			name:         "small key and value",
			key:          []byte("key"),
			value:        []byte("value"),
			expectedSize: 20 + 3 + 5, // Header + key + value
		},
		{
			name:         "large data",
			key:          bytes.Repeat([]byte("k"), 1000),
			value:        bytes.Repeat([]byte("v"), 2000),
			expectedSize: 20 + 1000 + 2000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			record := NewRecord(tc.key, tc.value)
			if record.Size() != tc.expectedSize {
				t.Errorf("Size mismatch: got %d, want %d", record.Size(), tc.expectedSize)
			}
		})
	}
}

func TestNewRecord(t *testing.T) {
	key := []byte("test key")
	value := []byte("test value")

	record := NewRecord(key, value)

	// Check fields are set correctly
	if record.KeySize != uint32(len(key)) {
		t.Errorf("KeySize mismatch: got %d, want %d", record.KeySize, len(key))
	}

	if record.ValueSize != uint32(len(value)) {
		t.Errorf("ValueSize mismatch: got %d, want %d", record.ValueSize, len(value))
	}

	if !bytes.Equal(record.Key, key) {
		t.Errorf("Key mismatch: got %v, want %v", record.Key, key)
	}

	if !bytes.Equal(record.Value, value) {
		t.Errorf("Value mismatch: got %v, want %v", record.Value, value)
	}

	// Check timestamp is reasonable
	now := time.Now().UnixNano()
	if record.Timestamp > uint64(now) || record.Timestamp < uint64(now-int64(time.Second)) {
		t.Errorf("Timestamp seems unreasonable: %d", record.Timestamp)
	}

	// CRC32 should be zero initially (set during encoding)
	if record.CRC32 != 0 {
		t.Errorf("Expected CRC32 to be zero initially, got %d", record.CRC32)
	}
}

func TestRecord_CalculateCRC32(t *testing.T) {
	key := []byte("test key")
	value := []byte("test value")
	record := NewRecord(key, value)

	// Calculate CRC32
	crc := record.calculateCRC32()

	// Should be non-zero for non-empty data
	if crc == 0 {
		t.Error("Expected non-zero CRC32 for non-empty record")
	}

	// Should be deterministic
	crc2 := record.calculateCRC32()
	if crc != crc2 {
		t.Errorf("CRC32 calculation is not deterministic: %d vs %d", crc, crc2)
	}

	// Different data should produce different CRC
	record2 := NewRecord([]byte("different key"), value)
	crc3 := record2.calculateCRC32()
	if crc == crc3 {
		t.Error("Different records produced same CRC32 (highly unlikely)")
	}
}
