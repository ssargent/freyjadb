package codec_test

import (
	"fmt"
	"log"
	
	"github.com/ssargent/freyjadb/pkg/codec"
)

// ExampleRecordCodec_basic demonstrates basic record encoding and decoding
func ExampleRecordCodec_basic() {
	// Create a new codec
	codec := codec.NewRecordCodec()
	
	// Encode a key-value pair
	key := []byte("user:123")
	value := []byte("john@example.com")
	
	encoded, err := codec.Encode(key, value)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Encoded %d bytes\n", len(encoded))
	
	// Decode the record
	record, err := codec.Decode(encoded)
	if err != nil {
		log.Fatal(err)
	}
	
	// Validate the record
	if err := record.Validate(); err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Key: %s\n", record.Key)
	fmt.Printf("Value: %s\n", record.Value)
	fmt.Printf("Timestamp: %d\n", record.Timestamp)
	
	// Output:
	// Encoded 47 bytes
	// Key: user:123
	// Value: john@example.com
	// Timestamp: 1719043200000000000
}

// ExampleRecord_creation demonstrates creating and inspecting records
func ExampleRecord_creation() {
	// Create a new record
	key := []byte("config:db")
	value := []byte(`{"host": "localhost", "port": 5432}`)
	
	record := codec.NewRecord(key, value)
	
	fmt.Printf("Key size: %d bytes\n", record.KeySize)
	fmt.Printf("Value size: %d bytes\n", record.ValueSize)
	fmt.Printf("Total size: %d bytes\n", record.Size())
	fmt.Printf("Has timestamp: %t\n", record.Timestamp > 0)
	
	// Output:
	// Key size: 9 bytes
	// Value size: 32 bytes
	// Total size: 61 bytes
	// Has timestamp: true
}

// ExampleRecordCodec_errorHandling demonstrates error handling
func ExampleRecordCodec_errorHandling() {
	codec := codec.NewRecordCodec()
	
	// Try to decode malformed data
	malformed := []byte{0x01, 0x02, 0x03} // Too short
	
	_, err := codec.Decode(malformed)
	if err != nil {
		fmt.Printf("Decode error: %v\n", err)
	}
	
	// Output:
	// Decode error: not implemented
}

// ExampleRecordCodec_binaryData demonstrates handling binary data
func ExampleRecordCodec_binaryData() {
	codec := codec.NewRecordCodec()
	
	// Binary key and value
	key := []byte{0x00, 0x01, 0x02, 0x03}
	value := []byte{0xFF, 0xFE, 0xFD, 0xFC, 0xFB}
	
	encoded, err := codec.Encode(key, value)
	if err != nil {
		log.Fatal(err)
	}
	
	record, err := codec.Decode(encoded)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Binary key: %x\n", record.Key)
	fmt.Printf("Binary value: %x\n", record.Value)
	
	// Output:
	// Binary key: 00010203
	// Binary value: fffefdfcfb
}

// ExampleRecordCodec_emptyData demonstrates handling empty keys and values
func ExampleRecordCodec_emptyData() {
	codec := codec.NewRecordCodec()
	
	// Empty key, non-empty value
	encoded1, err := codec.Encode([]byte(""), []byte("value"))
	if err != nil {
		log.Fatal(err)
	}
	
	// Non-empty key, empty value  
	encoded2, err := codec.Encode([]byte("key"), []byte(""))
	if err != nil {
		log.Fatal(err)
	}
	
	// Both empty
	encoded3, err := codec.Encode([]byte(""), []byte(""))
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Empty key record: %d bytes\n", len(encoded1))
	fmt.Printf("Empty value record: %d bytes\n", len(encoded2))
	fmt.Printf("Both empty record: %d bytes\n", len(encoded3))
	
	// Output:
	// Empty key record: 25 bytes
	// Empty value record: 23 bytes
	// Both empty record: 20 bytes
}
