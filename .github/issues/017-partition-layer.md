---
name: Feature request
about: Suggest an idea for FreyjaDB
title: '[FEATURE] Implement Partition Layer for Multi-Tenant Support'
labels: enhancement, priority-medium
assignees: ''
---

## Is your feature request related to a problem? Please describe

FreyjaDB needs support for logical data partitioning similar to DynamoDB's partition key concept. This enables multi-tenant applications and provides logical data separation within a single database instance.

## Describe the solution you'd like

Implement a partition layer that:

- Creates sub-directories for each partition key (PK)
- Opens/closes Bitcask instances on demand per partition
- Ensures keys in different partitions never collide
- Provides efficient partition management and cleanup

## Describe alternatives you've considered

- Single global namespace (no tenant isolation)
- External partitioning (requires application-level management)
- Database-per-tenant (too much overhead)

## Use Case

This feature addresses the need for:

- Multi-tenant SaaS applications with data isolation
- Logical data organization by customer/region/service
- Efficient resource usage with on-demand partition loading
- Foundation for partition-level operations and analytics

## API Design (if applicable)

```go
type PartitionedKVStore struct {
    baseDir    string
    partitions map[string]*KVStore
    mutex      sync.RWMutex
    config     PartitionConfig
}

type PartitionConfig struct {
    BaseDir           string
    MaxOpenPartitions int
    PartitionTTL      time.Duration
    AutoCreateDirs    bool
}

func NewPartitionedKVStore(config PartitionConfig) (*PartitionedKVStore, error) {
    // Implementation details
}

func (pks *PartitionedKVStore) Get(partitionKey string, key []byte) ([]byte, error) {
    // Get from specific partition
}

func (pks *PartitionedKVStore) Put(partitionKey string, key, value []byte) error {
    // Put to specific partition
}

func (pks *PartitionedKVStore) Delete(partitionKey string, key []byte) error {
    // Delete from specific partition
}

func (pks *PartitionedKVStore) getOrCreatePartition(partitionKey string) (*KVStore, error) {
    // Lazy partition loading
}

func (pks *PartitionedKVStore) ListPartitions() ([]string, error) {
    // List available partitions
}
```

## Implementation Details

- [x] Directory-based partition organization
- [x] Lazy loading of partition instances
- [x] LRU eviction for partition cache management
- [x] Thread-safe partition access and creation
- [x] Partition lifecycle management

## Performance Considerations

- Expected performance impact: Neutral (adds partitioning overhead)
- Memory usage changes: Scales with number of active partitions
- Disk space impact: Organized by partition directories

## Testing Strategy

- [x] Multi-partition isolation testing
- [x] Concurrent partition access
- [x] Partition creation and cleanup
- [x] Key collision prevention across partitions
- [x] Resource usage with many partitions

## Documentation Requirements

- [x] Partition key design guidelines
- [x] Directory structure documentation
- [x] Performance implications
- [x] Multi-tenant usage patterns

## Additional Context

This is item #17 from the project roadmap with complexity rating 3/5. Partitioning is important for multi-tenant applications and provides the foundation for sort key support.

## Acceptance Criteria

- [ ] Each partition key maps to separate sub-directory
- [ ] Keys in different partitions never collide
- [ ] Partitions are created on-demand
- [ ] Resource usage scales reasonably with partition count
- [ ] Partition instances can be safely closed/reopened
- [ ] Directory structure is clean and organized

## Priority

- [x] Medium - Important for multi-tenant applications

## Related Issues

- Depends on: #005 (Basic KV API)
- Depends on: #013 (Read/Write Concurrency)
- Enables: #018 (Sort Key Range Support)
- Part of DynamoDB-like API layer
