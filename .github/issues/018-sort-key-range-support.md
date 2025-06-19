---
name: Feature request
about: Suggest an idea for FreyjaDB
title: '[FEATURE] Implement Sort Key Range Support with B-Tree'
labels: enhancement, priority-medium, complexity-high
assignees: ''
---

## Is your feature request related to a problem? Please describe

FreyjaDB needs support for sort keys within partitions to enable range queries similar to DynamoDB's sort key functionality. This requires maintaining sorted order within each partition for efficient range operations.

## Describe the solution you'd like

Implement sort key support that:

- Maintains an in-memory B-tree sorted by sort key within each partition
- Supports range queries: query(pk, start_sk, end_sk)
- Streams results in sorted order
- Integrates with existing partition layer

## Describe alternatives you've considered

- External sorting for queries (too slow)
- Skip lists (more complex implementation)
- Simple sorted arrays (poor insert performance)

## Use Case

This feature addresses the need for:

- Time-series data queries (timestamps as sort keys)
- Leaderboards and rankings
- Event logs with chronological ordering
- User activity streams and timelines

## API Design (if applicable)

```go
type SortKeyIndex struct {
    btree    *BPlusTree  // Your existing B+ tree implementation
    sortKeys map[string][]byte  // key -> sort_key mapping
    mutex    sync.RWMutex
}

type PartitionWithSortKeys struct {
    store     *KVStore
    sortIndex *SortKeyIndex
    mutex     sync.RWMutex
}

type RangeQuery struct {
    PartitionKey string
    StartSortKey []byte
    EndSortKey   []byte
    Limit        int
    Reverse      bool
}

type RangeResult struct {
    Key      []byte
    SortKey  []byte
    Value    []byte
}

func (psk *PartitionWithSortKeys) PutWithSortKey(key, sortKey, value []byte) error {
    // Store in both hash index and B-tree
}

func (psk *PartitionWithSortKeys) QueryRange(query RangeQuery) (<-chan RangeResult, error) {
    // Stream sorted results in range
}

func (psk *PartitionWithSortKeys) GetWithSortKey(key, sortKey []byte) ([]byte, error) {
    // Point lookup with sort key
}
```

## Implementation Details

- [x] Integration with existing B+ tree implementation
- [x] Dual indexing: hash for point lookups, B-tree for ranges
- [x] Sort key extraction and storage
- [x] Range iteration with proper ordering
- [x] Memory management for sort indexes

## Performance Considerations

- Expected performance impact: Positive for range queries, slight overhead for point operations
- Memory usage changes: Additional B-tree memory per partition
- Disk space impact: Minimal (sort keys stored in memory)

## Testing Strategy

- [x] Range query correctness with various ranges
- [x] Sort order validation
- [x] Integration with partition layer
- [x] Performance benchmarks for range operations
- [x] Memory usage profiling

## Documentation Requirements

- [x] Sort key design patterns
- [x] Range query API documentation
- [x] Performance characteristics
- [x] Integration with partition keys

## Additional Context

This is item #18 from the project roadmap with complexity rating 4/5. This feature leverages your existing B+ tree implementation and completes the DynamoDB-like API layer.

## Acceptance Criteria

- [ ] Range queries return results in sort key order
- [ ] Point lookups work with sort key specification
- [ ] Integration with partition layer is seamless
- [ ] B-tree maintains proper sort order
- [ ] Memory usage scales reasonably with data size
- [ ] Range boundaries are handled correctly

## Priority

- [x] Medium - Important for advanced query patterns

## Related Issues

- Depends on: Current B+ tree implementation
- Depends on: #017 (Partition Layer)
- Completes DynamoDB-like functionality
- May relate to: Your current concurrency work
