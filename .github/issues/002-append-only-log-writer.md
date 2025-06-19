---
name: Feature request
about: Suggest an idea for FreyjaDB
title: '[FEATURE] Implement Append-Only Log Writer'
labels: enhancement, priority-high
assignees: ''
---

## Is your feature request related to a problem? Please describe

FreyjaDB needs an append-only log writer that provides crash-safe, sequential writes to disk. This is essential for the Bitcask-style architecture where all writes are appended to an active log file.

## Describe the solution you'd like

Implement an append-only log writer that:

- Accepts key-value pairs and appends encoded records to active.data file
- Provides configurable fsync intervals for durability vs performance tuning
- Handles file growth and maintains write ordering
- Ensures atomic record writes (no partial records)

## Describe alternatives you've considered

- Direct file writes without buffering (poor performance)
- In-memory buffering without fsync (data loss risk)
- SQLite WAL mode (too complex for our use case)

## Use Case

This feature addresses the need for:

- High-throughput sequential writes optimized for SSDs
- Configurable durability guarantees
- Foundation for crash recovery mechanisms
- Simple append-only storage semantics

## API Design (if applicable)

```go
type LogWriter struct {
    file       *os.File
    codec      *RecordCodec
    fsyncTimer *time.Timer
    mutex      sync.Mutex
}

type LogWriterConfig struct {
    FilePath     string
    FsyncInterval time.Duration
    BufferSize   int
}

func NewLogWriter(config LogWriterConfig) (*LogWriter, error) {
    // Implementation details
}

func (w *LogWriter) Put(key, value []byte) error {
    // Encode record and append to file
}

func (w *LogWriter) Sync() error {
    // Force fsync
}

func (w *LogWriter) Close() error {
    // Cleanup resources
}
```

## Implementation Details

- [x] File management with O_APPEND mode
- [x] Record encoding integration with RecordCodec
- [x] Configurable fsync intervals with timer-based flushing
- [x] Thread-safe writes with appropriate locking
- [x] Error handling for disk full, permission errors

## Performance Considerations

- Expected performance impact: Positive (optimized for sequential writes)
- Memory usage changes: Small increase for write buffers
- Disk space impact: Linear growth with data volume

## Testing Strategy

- [x] Unit tests for put operations and file growth
- [x] Fsync interval testing and configuration
- [x] Concurrent write testing
- [x] Error condition testing (disk full, permissions)
- [x] Performance benchmarks for write throughput

## Documentation Requirements

- [x] API documentation for LogWriter interface
- [x] Configuration options documentation
- [x] Usage examples for basic operations
- [x] Performance tuning guidelines

## Additional Context

This is item #2 from the project roadmap. The log writer must be robust and performant as it's the foundation for all write operations in the database.

## Acceptance Criteria

- [ ] Records are appended to active.data file
- [ ] Configurable fsync intervals work correctly
- [ ] File grows as expected with write operations
- [ ] Tail bytes match encoded record format
- [ ] Thread-safe concurrent writes
- [ ] Proper error handling and recovery

## Priority

- [x] High - Critical for basic write functionality

## Related Issues

- Depends on: #001 (Record Codec)
- Enables: #003 (Sequential Log Reader)
