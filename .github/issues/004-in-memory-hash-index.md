---
name: Feature request
about: Suggest an idea for FreyjaDB
title: '[FEATURE] Implement In-Memory Hash Index Builder'
labels: enhancement, priority-high
assignees: ''
---

## Is your feature request related to a problem? Please describe

FreyjaDB needs an in-memory hash index to provide fast key lookups. The index maps keys to their latest file positions, enabling O(1) average-case lookups for the Bitcask architecture.

## Describe the solution you'd like

Implement an in-memory hash index that:

- Scans log files to build a key → (file_id, offset, size) mapping
- Handles duplicate keys by keeping only the latest occurrence
- Provides fast lookup operations for key existence and location
- Supports incremental updates as new records are written

## Describe alternatives you've considered

- B-tree index (more complex, not needed for point lookups)
- External sorting (too slow for real-time operations)
- Bloom filters only (can't provide exact locations)

## Use Case

This feature addresses the need for:

- Fast key existence checking and value retrieval
- Mapping keys to their physical storage locations
- Handling key updates and deletions (tombstones)
- Foundation for the core KV API operations

## API Design (if applicable)

```go
type IndexEntry struct {
    FileID    uint32
    Offset    int64
    Size      uint32
    Timestamp uint64
}

type HashIndex struct {
    entries map[string]*IndexEntry
    mutex   sync.RWMutex
}

func NewHashIndex() *HashIndex {
    // Implementation details
}

func (idx *HashIndex) Put(key []byte, entry *IndexEntry) {
    // Add or update index entry
}

func (idx *HashIndex) Get(key []byte) (*IndexEntry, bool) {
    // Lookup index entry for key
}

func (idx *HashIndex) Delete(key []byte) {
    // Remove key from index
}

func (idx *HashIndex) BuildFromLog(reader *LogReader) error {
    // Scan log file and populate index
}

func (idx *HashIndex) Size() int {
    // Return number of keys in index
}
```

## Implementation Details

- [x] HashMap-based storage for O(1) average lookups
- [x] Integration with LogReader for index building
- [x] Duplicate key handling (latest wins semantics)
- [x] Thread-safe operations with RWMutex
- [x] Memory-efficient key storage

## Performance Considerations

- Expected performance impact: Positive (enables fast lookups)
- Memory usage changes: Linear with number of unique keys
- Disk space impact: No change (in-memory only)

## Testing Strategy

- [x] Unit tests for basic put/get/delete operations
- [x] Duplicate key handling verification
- [x] Index building from log files with known data
- [x] Concurrent access testing
- [x] Memory usage benchmarks

## Documentation Requirements

- [x] API documentation for HashIndex interface
- [x] Usage examples for index operations
- [x] Memory usage characteristics
- [x] Concurrency safety documentation

## Additional Context

This is item #4 from the project roadmap. The hash index is critical for achieving the performance characteristics expected from a Bitcask-style database.

## Acceptance Criteria

- [ ] Can build index by scanning log files
- [ ] Duplicate keys result in index pointing to newest record
- [ ] O(1) average-case lookup performance
- [ ] Thread-safe concurrent operations
- [ ] Memory usage scales linearly with unique key count
- [ ] Handles empty keys and large keys correctly

## Priority

- [x] High - Critical for lookup performance

## Related Issues

- Depends on: #001 (Record Codec)
- Depends on: #003 (Sequential Log Reader)
- Enables: #005 (Basic KV API)
