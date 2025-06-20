# Record Codec Implementation Guide

This directory contains the complete test suite and structure for implementing FreyjaDB's record codec (Issue #1).

## 📁 Structure Created

```text
pkg/codec/
├── doc.go                    # Package documentation
├── record.go                 # Main implementation (TODO: implement methods)
├── record_test.go           # Comprehensive unit tests
├── record_bench_test.go     # Performance benchmarks  
├── record_fuzz_test.go      # Property-based/fuzz tests
├── example_test.go          # Usage examples
└── structure_test.go        # Basic structure validation
```

## 🎯 Implementation Tasks

The following methods in `record.go` need to be implemented:

### 1. `RecordCodec.Encode(key, value []byte) ([]byte, error)`

**Goal**: Serialize a key-value pair into binary format.

**Format**: `[CRC32(4)][KeySize(4)][ValueSize(4)][Timestamp(8)][Key][Value]`

**Steps**:

1. Create a Record with current timestamp
2. Allocate byte slice for total size (20 + len(key) + len(value))
3. Write header fields using `binary.LittleEndian`
4. Append key and value data
5. Calculate CRC32 over everything except CRC field
6. Write CRC32 to the beginning
7. Return complete binary record

### 2. `RecordCodec.Decode(data []byte) (*Record, error)`

**Goal**: Deserialize binary data into a Record struct.

**Steps**:

1. Validate minimum size (at least 20 bytes for header)
2. Parse header fields using `binary.LittleEndian`
3. Validate declared sizes match available data
4. Extract key and value slices
5. Create Record struct with parsed data
6. Return Record (CRC validation happens in Validate())

### 3. `Record.Validate() error`

**Goal**: Verify record integrity using CRC32 checksum.

**Steps**:

1. Call `r.calculateCRC32()` to compute expected CRC
2. Compare with `r.CRC32` field
3. Return error if mismatch, nil if valid

**Note**: The `calculateCRC32()` method is already implemented as a reference.

## 📋 Test Coverage

The test suite covers:

- ✅ **Round-trip encoding/decoding** with various data sizes
- ✅ **CRC32 validation** and corruption detection  
- ✅ **Error handling** for malformed data
- ✅ **Edge cases** (empty keys/values, binary data, large data)
- ✅ **Performance benchmarks** for encode/decode operations
- ✅ **Fuzz testing** for property verification
- ✅ **Usage examples** for documentation

## 🧪 Running Tests

```bash
# Run all tests
go test ./pkg/codec -v

# Run specific test
go test ./pkg/codec -v -run TestStructureSetup

# Run benchmarks
go test ./pkg/codec -bench=.

# Run fuzz tests (Go 1.18+)
go test ./pkg/codec -fuzz=FuzzRecordCodec_RoundTrip

# Check test coverage
go test ./pkg/codec -cover
```

## 📊 Expected Performance

Target benchmarks (will vary by hardware):

- **Encode**: < 100ns for small records, O(n) for data size
- **Decode**: < 50ns for small records, O(n) for data size  
- **Validate**: < 30ns for small records, O(n) for data size
- **Memory**: Minimal allocations, primarily for output buffer

## 🔍 Implementation Hints

1. **Use binary.LittleEndian** for all multi-byte fields
2. **Pre-allocate buffers** to minimize allocations
3. **Validate inputs** before processing to provide clear errors
4. **Consider using unsafe** for performance-critical paths (optional)
5. **Test with the provided test suite** after each method implementation

## ✅ Acceptance Criteria

When implementation is complete, all tests should pass:

- [ ] `TestStructureSetup` - basic structure validation
- [ ] `TestRecordCodec_EncodeDecodeRoundTrip` - round-trip integrity
- [ ] `TestRecordCodec_CRCValidation` - corruption detection
- [ ] `TestRecordCodec_MalformedData` - error handling
- [ ] All benchmark tests run without errors
- [ ] All fuzz tests pass
- [ ] Examples in `example_test.go` work correctly

## 🎉 Next Steps

After implementing the record codec:

1. **Run full test suite**: `make test`
2. **Check performance**: `go test ./pkg/codec -bench=.`
3. **Update project status**: Mark Issue #1 as complete
4. **Move to Issue #2**: Implement append-only log writer

Good luck with the implementation! The test suite will guide you and verify correctness at each step.
