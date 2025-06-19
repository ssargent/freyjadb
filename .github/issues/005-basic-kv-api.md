---
name: Feature request
about: Suggest an idea for FreyjaDB
title: '[FEATURE] Implement Basic KV API (Single-threaded)'
labels: enhancement, priority-high
assignees: ''
---

## Is your feature request related to a problem? Please describe

FreyjaDB needs a complete single-threaded key-value API that ties together the log writer, reader, and hash index to provide the core database operations: Get, Put, and Delete.

## Describe the solution you'd like

Implement a basic KV API that:

- Provides Get, Put, and Delete operations
- Integrates log writer for persistence
- Uses hash index for fast lookups
- Handles tombstone records for deletions
- Maintains consistency between index and log files

## Describe alternatives you've considered

- Async API with futures (more complex for initial implementation)
- Batch operations (can be added later)
- Transactional operations (out of scope for Bitcask model)

## Use Case

This feature addresses the need for:

- Complete key-value database functionality
- Simple, synchronous API for applications
- Foundation for more advanced features
- Testing and validation of core components

## API Design (if applicable)

```go
type KVStore struct {
    writer *LogWriter
    index  *HashIndex
    mutex  sync.Mutex // Single-threaded for now
}

type KVStoreConfig struct {
    DataDir       string
    FsyncInterval time.Duration
}

func NewKVStore(config KVStoreConfig) (*KVStore, error) {
    // Implementation details
}

func (kv *KVStore) Get(key []byte) ([]byte, error) {
    // Lookup in index, read from log file
}

func (kv *KVStore) Put(key, value []byte) error {
    // Write to log, update index
}

func (kv *KVStore) Delete(key []byte) error {
    // Write tombstone, update index
}

func (kv *KVStore) Close() error {
    // Cleanup resources
}

// Error types
var (
    ErrKeyNotFound = errors.New("key not found")
    ErrInvalidKey  = errors.New("invalid key")
)
```

## Implementation Details

- [x] Integration of LogWriter, LogReader, and HashIndex
- [x] Tombstone record handling for deletions
- [x] Error handling for missing keys and I/O errors
- [x] Single mutex for thread safety (upgraded later)
- [x] Proper resource cleanup and file management

## Performance Considerations

- Expected performance impact: Positive (complete functionality)
- Memory usage changes: Combination of existing components
- Disk space impact: Linear growth with data and tombstones

## Testing Strategy

- [x] Unit tests for all CRUD operations
- [x] Property-based tests: get(after put) == value
- [x] Property-based tests: get(after delete) == KeyNotFound
- [x] Round-trip testing: put → get → delete → get
- [x] Error condition testing
- [x] Basic performance benchmarks

## Documentation Requirements

- [x] API documentation for all public methods
- [x] Usage examples and getting started guide
- [x] Error handling documentation
- [x] Configuration options

## Additional Context

This is item #5 from the project roadmap. This completes the basic single-threaded database functionality and provides a foundation for all subsequent enhancements.

## Acceptance Criteria

- [ ] Get operation retrieves values for existing keys
- [ ] Put operation stores key-value pairs persistently
- [ ] Delete operation removes keys (returns KeyNotFound on subsequent Get)
- [ ] Get on non-existent key returns ErrKeyNotFound
- [ ] All operations maintain consistency between index and log
- [ ] Proper error handling for I/O failures
- [ ] Performance meets basic expectations for small datasets

## Priority

- [x] High - Core database functionality

## Related Issues

- Depends on: #001 (Record Codec)
- Depends on: #002 (Append-Only Log Writer)
- Depends on: #003 (Sequential Log Reader)
- Depends on: #004 (In-Memory Hash Index)
- Enables: #006 (Crash-Safe Reopen)
