---
name: Feature request
about: Suggest an idea for FreyjaDB
title: '[FEATURE] Implement Memory Packing with Slab Allocator'
labels: enhancement, priority-medium
assignees: ''
---

## Is your feature request related to a problem? Please describe

FreyjaDB's current hash index stores keys as individual byte slices, leading to memory fragmentation and excessive per-key overhead. A slab allocator approach can significantly reduce memory usage and improve cache locality.

## Describe the solution you'd like

Implement memory packing that:

- Stores key bytes in contiguous "key log" allocations
- Uses u64 pointers + lengths instead of byte slices in index
- Implements slab allocator for efficient memory management
- Achieves ≥30% reduction in RAM per entry

## Describe alternatives you've considered

- Individual allocations (current approach, high overhead)
- Memory pools (less efficient packing)
- External memory management (adds complexity)

## Use Case

This feature addresses the need for:

- Reduced memory footprint for large key sets
- Better cache locality for index operations
- Lower garbage collection pressure in Go
- More predictable memory usage patterns

## API Design (if applicable)

```go
type KeySlab struct {
    data   []byte
    offset int64
    size   int64
}

type SlabAllocator struct {
    slabs     []*KeySlab
    slabSize  int64
    current   *KeySlab
    mutex     sync.Mutex
}

type PackedKeyRef struct {
    SlabID uint32
    Offset uint32
    Length uint32
}

type PackedHashIndex struct {
    entries   map[uint64]*PackedIndexEntry  // key_hash -> entry
    allocator *SlabAllocator
    mutex     sync.RWMutex
}

type PackedIndexEntry struct {
    KeyRef    PackedKeyRef
    FileID    uint32
    Offset    int64
    Size      uint32
    Timestamp uint64
}

func NewSlabAllocator(slabSize int64) *SlabAllocator {
    // Implementation details
}

func (sa *SlabAllocator) Allocate(keyData []byte) (PackedKeyRef, error) {
    // Allocate space in slab and return reference
}

func (sa *SlabAllocator) Resolve(ref PackedKeyRef) ([]byte, error) {
    // Resolve reference to actual key bytes
}

func (sa *SlabAllocator) Stats() *SlabStats {
    // Return memory usage statistics
}

type SlabStats struct {
    TotalSlabs    int
    AllocatedBytes int64
    WastedBytes   int64
    FragmentationRatio float64
}
```

## Implementation Details

- [x] Contiguous memory allocation for key storage
- [x] Pointer-based index entries with packed references
- [x] Slab management with configurable slab sizes
- [x] Memory usage tracking and statistics
- [x] Garbage collection for unused slabs

## Performance Considerations

- Expected performance impact: Positive (better cache locality, less GC pressure)
- Memory usage changes: Target ≥30% reduction in RAM per entry
- Disk space impact: No change

## Testing Strategy

- [x] Memory usage benchmarks vs current implementation
- [x] Cache locality performance testing
- [x] Slab allocation and deallocation correctness
- [x] Large dataset memory profiling
- [x] Fragmentation analysis under various workloads

## Documentation Requirements

- [x] Memory layout documentation
- [x] Performance characteristics and benchmarks
- [x] Configuration tuning guidelines
- [x] Memory usage optimization tips

## Additional Context

This is item #14 from the project roadmap with complexity rating 3/5. Memory packing is important for large-scale deployments where memory efficiency is critical.

## Acceptance Criteria

- [ ] Key bytes stored in contiguous slab allocations
- [ ] Index uses u64 pointers + lengths instead of byte slices
- [ ] Memory usage per entry reduced by ≥30%
- [ ] Performance maintained or improved vs current approach
- [ ] Proper memory management with slab cleanup
- [ ] Comprehensive memory usage statistics

## Priority

- [x] Medium - Important for memory efficiency at scale

## Related Issues

- Depends on: #004 (In-Memory Hash Index)
- Enhances: #013 (Read/Write Concurrency)
- Performance optimization for large datasets
