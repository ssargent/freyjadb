//go:build bench
// +build bench

package codec

import (
	"fmt"
	"testing"
)

// BenchmarkRecordCodec benchmarks the record codec with various data sizes
func BenchmarkRecordCodec(b *testing.B) {
	dataSizes := []int{100, 1000, 10000, 100000}

	for _, dataSize := range dataSizes {
		b.Run(fmt.Sprintf("Data_Size_%d", dataSize), func(b *testing.B) {
			codec := NewRecordCodec()

			// Generate test data
			key := make([]byte, 50)
			value := make([]byte, dataSize)
			for i := range 50 {
				key[i] = byte(i % 256)
			}
			for i := range dataSize {
				value[i] = byte(i % 256)
			}

			b.ResetTimer()
			b.ReportAllocs()

			b.Run("Encode", func(b *testing.B) {
				for range b.N {
					_, err := codec.Encode(key, value)
					if err != nil {
						b.Fatalf("Failed to encode: %v", err)
					}
				}
			})

			// For decode benchmark, we need encoded data
			encodedData, err := codec.Encode(key, value)
			if err != nil {
				b.Fatalf("Failed to encode for decode benchmark: %v", err)
			}

			b.Run("Decode", func(b *testing.B) {
				for range b.N {
					_, err := codec.Decode(encodedData)
					if err != nil {
						b.Fatalf("Failed to decode: %v", err)
					}
				}
			})

			b.Run("Round_Trip", func(b *testing.B) {
				for range b.N {
					record, err := codec.Encode(key, value)
					if err != nil {
						b.Fatalf("Failed to encode: %v", err)
					}

					_, err = codec.Decode(record)
					if err != nil {
						b.Fatalf("Failed to decode: %v", err)
					}
				}
			})
		})
	}
}

// BenchmarkRecord_Validate benchmarks the record validation
func BenchmarkRecord_Validate(b *testing.B) {
	codec := NewRecordCodec()

	// Create a valid record
	key := []byte("test_key")
	value := []byte("test_value")
	recordData, err := codec.Encode(key, value)
	if err != nil {
		b.Fatalf("Failed to encode: %v", err)
	}

	record, err := codec.Decode(recordData)
	if err != nil {
		b.Fatalf("Failed to decode: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		err := record.Validate()
		if err != nil {
			b.Fatalf("Validation failed: %v", err)
		}
	}
}

// BenchmarkRecord_CRC32 benchmarks CRC32 calculation
func BenchmarkRecord_CRC32(b *testing.B) {
	// Create test data of various sizes
	dataSizes := []int{100, 1000, 10000, 100000}

	for _, dataSize := range dataSizes {
		b.Run(fmt.Sprintf("Data_Size_%d", dataSize), func(b *testing.B) {
			data := make([]byte, dataSize)
			for i := range dataSize {
				data[i] = byte(i % 256)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				_ = calculateCRC32(data)
			}
		})
	}
}

// BenchmarkRecord_Marshal benchmarks JSON marshaling operations
func BenchmarkRecord_Marshal(b *testing.B) {
	// Test with different numbers of records
	recordCounts := []int{1, 10, 100, 1000}

	for _, recordCount := range recordCounts {
		b.Run(fmt.Sprintf("Record_Count_%d", recordCount), func(b *testing.B) {
			records := make([]*Record, recordCount)
			codec := NewRecordCodec()

			// Create test records
			for i := range recordCount {
				key := []byte(fmt.Sprintf("key_%d", i))
				value := []byte(fmt.Sprintf("value_%d_with_some_additional_content", i))
				recordData, err := codec.Encode(key, value)
				if err != nil {
					b.Fatalf("Failed to encode: %v", err)
				}

				record, err := codec.Decode(recordData)
				if err != nil {
					b.Fatalf("Failed to decode: %v", err)
				}
				records[i] = record
			}

			b.ResetTimer()
			b.ReportAllocs()

			b.Run("Marshal_Single", func(b *testing.B) {
				for range b.N {
					for _, record := range records {
						_, err := record.MarshalJSON()
						if err != nil {
							b.Fatalf("Failed to marshal: %v", err)
						}
					}
				}
			})

			b.Run("Unmarshal_Single", func(b *testing.B) {
				// Create JSON data for unmarshaling
				jsonData := make([][]byte, recordCount)
				for i, record := range records {
					data, err := record.MarshalJSON()
					if err != nil {
						b.Fatalf("Failed to marshal: %v", err)
					}
					jsonData[i] = data
				}

				for range b.N {
					for _, data := range jsonData {
						record := &Record{}
						err := record.UnmarshalJSON(data)
						if err != nil {
							b.Fatalf("Failed to unmarshal: %v", err)
						}
					}
				}
			})
		})
	}
}

// BenchmarkRecord_Allocations benchmarks memory allocations
func BenchmarkRecord_Allocations(b *testing.B) {
	codec := NewRecordCodec()

	// Test different allocation patterns
	b.Run("Encode_No_Alloc", func(b *testing.B) {
		key := []byte("test_key")
		value := []byte("test_value")

		b.ReportAllocs()
		b.ResetTimer()

		for range b.N {
			_, err := codec.Encode(key, value)
			if err != nil {
				b.Fatalf("Failed to encode: %v", err)
			}
		}
	})

	b.Run("Decode_No_Alloc", func(b *testing.B) {
		// Create encoded data once
		key := []byte("test_key")
		value := []byte("test_value")
		encodedData, err := codec.Encode(key, value)
		if err != nil {
			b.Fatalf("Failed to encode: %v", err)
		}

		b.ReportAllocs()
		b.ResetTimer()

		for range b.N {
			_, err := codec.Decode(encodedData)
			if err != nil {
				b.Fatalf("Failed to decode: %v", err)
			}
		}
	})
}
