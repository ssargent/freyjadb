# FreyjaDB Benchmark Suite

This document describes the comprehensive benchmark suite for FreyjaDB components.

## Running Benchmarks

To run all benchmarks, use the following command:

```bash
go test -bench=. -benchmem ./...
```

To run only specific benchmark suites, you can filter by package:

```bash
# Run KV store benchmarks
go test -bench=. -benchmem ./pkg/store -run=^$ -bench=BenchmarkKVStore

# Run B+Tree benchmarks
go test -bench=. -benchmem ./pkg/bptree -run=^$ -bench=BenchmarkBPTree

# Run codec benchmarks
go test -bench=. -benchmem ./pkg/codec -run=^$ -bench=BenchmarkRecord
```

To run benchmarks with specific build tags:

```bash
go test -tags=bench -bench=. -benchmem ./...
```

## Benchmark Suites

### 1. KV Store Benchmarks (`pkg/store/kv_store_bench_test.go`)

Tests core KV store operations:
- Put operations with various key/value sizes
- Get operations for existing and non-existent keys
- Delete operations
- Concurrent read/write operations
- Batch operations

### 2. B+Tree Benchmarks (`pkg/bptree/bptree_bench_test.go`)

Tests B+Tree performance:
- Insert operations with different tree orders
- Search operations
- Delete operations
- Concurrent operations
- Tree height scaling

### 3. Hash Index Benchmarks (`pkg/store/hash_index_bench_test.go`)

Tests hash index performance:
- Put operations with various key sizes
- Get operations
- Delete operations
- Concurrent operations

### 4. Log Writer Benchmarks (`pkg/store/log_writer_bench_test.go`)

Tests log writer performance:
- Put operations with various record sizes
- Flush operations
- Different buffer sizes
- Concurrent writes

### 5. Log Reader Benchmarks (`pkg/store/log_reader_bench_test.go`)

Tests log reader performance:
- ReadNext operations
- ReadAt operations
- Iterator performance
- Concurrent reads

### 6. Codec Benchmarks (`pkg/codec/codec_bench_test.go`)

Tests record codec performance:
- Encode/decode operations with various data sizes
- Record validation
- CRC32 calculation
- JSON marshaling/unmarshaling
- Memory allocation patterns

### 7. Relationships Benchmarks (`pkg/store/relationships_bench_test.go`)

Tests relationship operations:
- PutRelationship operations
- GetRelationships operations
- DeleteRelationship operations
- Query performance
- Concurrent operations

## Interpreting Results

Benchmark results include:
- `ns/op`: Nanoseconds per operation
- `B/op`: Bytes allocated per operation
- `allocs/op`: Allocations per operation

Lower numbers are better for all metrics.

## Profiling

To profile benchmarks and identify bottlenecks:

```bash
# CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./pkg/store
go tool pprof cpu.prof

# Memory profiling
go test -bench=. -memprofile=mem.prof ./pkg/store
go tool pprof mem.prof

# Allocation profiling
go test -bench=. -allocsprofile=allocs.prof ./pkg/store
go tool pprof allocs.prof
```

## Continuous Performance Monitoring

The benchmark suite should be run regularly to detect performance regressions. Consider integrating with CI/CD pipelines to automatically run benchmarks and compare results against baseline measurements.

## Contributing New Benchmarks

When adding new benchmarks:
1. Place them in the appropriate `_bench_test.go` file
2. Use descriptive benchmark names
3. Include various workload sizes when applicable
4. Use `b.ResetTimer()` and `b.ReportAllocs()` appropriately
5. Document any special setup or assumptions