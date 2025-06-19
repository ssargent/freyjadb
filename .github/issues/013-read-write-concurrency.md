---
name: Feature request
about: Suggest an idea for FreyjaDB
title: '[FEATURE] Implement Read/Write Concurrency with Lock-Light Design'
labels: enhancement, priority-high, complexity-high
assignees: ''
---

## Is your feature request related to a problem? Please describe

FreyjaDB currently uses a single mutex for all operations, which severely limits throughput under concurrent load. A lock-light design is needed to enable high-performance concurrent reads while maintaining single-writer semantics.

## Describe the solution you'd like

Implement lock-light concurrency that:

- Uses a single writer mutex for write operations
- Allows concurrent readers on immutable index snapshots
- Implements atomic index swapping using ArcSwap or RCU techniques
- Maximizes read throughput while maintaining consistency

## Describe alternatives you've considered

- Reader-writer locks (still too much contention)
- Lock-free data structures (too complex for initial implementation)
- Multiple writer queues (breaks Bitcask semantics)

## Use Case

This feature addresses the need for:

- High-throughput concurrent read operations
- Scalable performance with multiple CPU cores
- Low-latency reads without writer blocking
- Foundation for production database workloads

## API Design (if applicable)

```go
type ConcurrentKVStore struct {
    writer      *LogWriter
    indexSwap   *atomic.Value  // Holds *HashIndex
    writerMutex sync.Mutex     // Single writer
    stats       *ConcurrencyStats
}

type ConcurrencyStats struct {
    ConcurrentReads  int64
    WriterBlocks     int64
    IndexSwaps       int64
    ReadLatencyP99   time.Duration
}

func NewConcurrentKVStore(config KVStoreConfig) (*ConcurrentKVStore, error) {
    // Implementation details
}

func (ckv *ConcurrentKVStore) Get(key []byte) ([]byte, error) {
    // Lock-free read using atomic index snapshot
}

func (ckv *ConcurrentKVStore) Put(key, value []byte) error {
    // Single writer with atomic index update
}

func (ckv *ConcurrentKVStore) swapIndex(newIndex *HashIndex) {
    // Atomic index replacement
}

func (ckv *ConcurrentKVStore) Stats() *ConcurrencyStats {
    // Return concurrency performance metrics
}
```

## Implementation Details

- [x] Atomic pointer operations for index swapping
- [x] Single writer mutex with minimal critical sections
- [x] Immutable index snapshots for readers
- [x] Memory management for old index cleanup
- [x] Performance monitoring and statistics

## Performance Considerations

- Expected performance impact: Highly positive (orders of magnitude improvement for reads)
- Memory usage changes: Temporary increase during index swaps
- Disk space impact: No change

## Testing Strategy

- [x] Stress testing with multiple concurrent readers
- [x] Single writer correctness under load
- [x] Index consistency during swaps
- [x] Performance benchmarks vs single-threaded version
- [x] Memory leak detection for index cleanup

## Documentation Requirements

- [x] Concurrency model explanation
- [x] Performance characteristics and benchmarks
- [x] Thread safety guarantees
- [x] Best practices for concurrent usage

## Additional Context

This is item #13 from the project roadmap with complexity rating 4/5. This feature is crucial for achieving production-grade performance and enables the database to scale with multiple CPU cores.

## Acceptance Criteria

- [ ] Multiple readers can operate concurrently without blocking
- [ ] Single writer semantics are maintained
- [ ] Index swaps are atomic and consistent
- [ ] No data races under concurrent load
- [ ] Read performance scales with CPU cores
- [ ] Memory management prevents leaks during index swaps

## Priority

- [x] High - Critical for production performance

## Related Issues

- Depends on: #004 (In-Memory Hash Index)
- Depends on: #005 (Basic KV API)
- Enhances: #011 (Hot-Swap Compaction)
- Enables production-scale deployment
