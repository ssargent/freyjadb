//go:build bench
// +build bench

package codec

import (
	"bytes"
	"testing"
)

func BenchmarkRecordCodec_Encode(b *testing.B) {
	codec := NewRecordCodec()

	benchmarks := []struct {
		name  string
		key   []byte
		value []byte
	}{
		{
			name:  "small",
			key:   []byte("user:123"),
			value: []byte("john@example.com"),
		},
		{
			name:  "medium",
			key:   bytes.Repeat([]byte("k"), 100),
			value: bytes.Repeat([]byte("v"), 1000),
		},
		{
			name:  "large",
			key:   bytes.Repeat([]byte("k"), 1000),
			value: bytes.Repeat([]byte("v"), 10000),
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := codec.Encode(bm.key, bm.value)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkRecordCodec_Decode(b *testing.B) {
	codec := NewRecordCodec()

	benchmarks := []struct {
		name  string
		key   []byte
		value []byte
	}{
		{
			name:  "small",
			key:   []byte("user:123"),
			value: []byte("john@example.com"),
		},
		{
			name:  "medium",
			key:   bytes.Repeat([]byte("k"), 100),
			value: bytes.Repeat([]byte("v"), 1000),
		},
		{
			name:  "large",
			key:   bytes.Repeat([]byte("k"), 1000),
			value: bytes.Repeat([]byte("v"), 10000),
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Pre-encode the data
			encoded, err := codec.Encode(bm.key, bm.value)
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := codec.Decode(encoded)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkRecordCodec_RoundTrip(b *testing.B) {
	codec := NewRecordCodec()

	benchmarks := []struct {
		name  string
		key   []byte
		value []byte
	}{
		{
			name:  "small",
			key:   []byte("user:123"),
			value: []byte("john@example.com"),
		},
		{
			name:  "medium",
			key:   bytes.Repeat([]byte("k"), 100),
			value: bytes.Repeat([]byte("v"), 1000),
		},
		{
			name:  "large",
			key:   bytes.Repeat([]byte("k"), 1000),
			value: bytes.Repeat([]byte("v"), 10000),
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				encoded, err := codec.Encode(bm.key, bm.value)
				if err != nil {
					b.Fatal(err)
				}

				_, err = codec.Decode(encoded)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkRecord_Validate(b *testing.B) {
	codec := NewRecordCodec()
	key := []byte("benchmark key")
	value := bytes.Repeat([]byte("v"), 1000)

	encoded, err := codec.Encode(key, value)
	if err != nil {
		b.Fatal(err)
	}

	record, err := codec.Decode(encoded)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := record.Validate()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRecord_CalculateCRC32(b *testing.B) {
	key := []byte("benchmark key")
	value := bytes.Repeat([]byte("v"), 1000)
	record := NewRecord(key, value)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = record.calculateCRC32()
	}
}

// Benchmark memory allocations
func BenchmarkRecordCodec_EncodeAllocs(b *testing.B) {
	codec := NewRecordCodec()
	key := []byte("user:123")
	value := []byte("john@example.com")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := codec.Encode(key, value)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRecordCodec_DecodeAllocs(b *testing.B) {
	codec := NewRecordCodec()
	key := []byte("user:123")
	value := []byte("john@example.com")

	encoded, err := codec.Encode(key, value)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := codec.Decode(encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}
