---
name: Feature request
about: Suggest an idea for FreyjaDB
title: '[FEATURE] Implement Crash-Safe Reopen and Recovery'
labels: enhancement, priority-high
assignees: ''
---

## Status

**Completed** - Crash recovery implementation verified in `pkg/store/kv_store.go` with Open() method and validateLogFile functionality.

## Is your feature request related to a problem? Please describe

FreyjaDB needs robust crash recovery to handle partial writes, corrupted records, and ensure data consistency after unexpected shutdowns. This is critical for production database reliability.

## Describe the solution you'd like

Implement crash-safe reopen functionality that:

- Validates log file integrity on startup using CRC checks
- Truncates corrupted tail records that may result from partial writes
- Rebuilds the in-memory index from validated log data
- Ensures no data loss for successfully written records

## Describe alternatives you've considered

- WAL (Write-Ahead Logging) - more complex than needed for Bitcask
- Shadow paging - doesn't fit append-only model
- Ignoring corruption - unacceptable for data integrity

## Use Case

This feature addresses the need for:

- Reliable database operation after crashes or power failures
- Data integrity guarantees for committed writes
- Automatic recovery without manual intervention
- Foundation for production deployment confidence

## API Design (if applicable)

```go
type RecoveryResult struct {
    RecordsValidated int64
    RecordsTruncated int64
    FileSizeBefore   int64
    FileSizeAfter    int64
    IndexRebuilt     bool
}

func (kv *KVStore) Open() (*RecoveryResult, error) {
    // Validate and recover from log files
}

func (kv *KVStore) validateLogFile(filePath string) (*RecoveryResult, error) {
    // Read until CRC failure, truncate if needed
}

func (kv *KVStore) rebuildIndex() error {
    // Scan validated log and rebuild index
}

// Recovery configuration
type RecoveryConfig struct {
    MaxCorruptionBytes int64 // Max bytes to truncate
    ValidateAllRecords bool  // Full validation vs fast recovery
}
```

## Implementation Details

- [x] Sequential log validation using existing LogReader
- [x] CRC failure detection and truncation logic
- [x] Index rebuilding from validated records
- [x] Atomic file operations for truncation
- [x] Recovery statistics and logging

## Performance Considerations

- Expected performance impact: Startup cost, runtime neutral
- Memory usage changes: Temporary increase during index rebuild
- Disk space impact: May decrease due to truncation

## Testing Strategy

- [x] Unit tests for clean restart (no corruption)
- [x] Corruption simulation: partial writes, bad CRC
- [x] Recovery validation: data before corruption point preserved
- [x] Large file recovery performance
- [x] Multiple corruption scenarios

## Documentation Requirements

- [x] Recovery process documentation
- [x] Data consistency guarantees
- [x] Performance impact of recovery
- [x] Configuration options

## Additional Context

This is item #6 from the project roadmap. Crash safety is essential for any production database system and builds confidence in the reliability of FreyjaDB.

## Acceptance Criteria

- [x] Clean restarts rebuild index correctly
- [x] Corrupted tail records are detected and truncated
- [x] Data before corruption point is preserved
- [x] Index accurately reflects recovered data
- [x] Recovery process completes automatically
- [x] Recovery statistics are available for monitoring

## Priority

- [x] High - Critical for production reliability

## Related Issues

- Depends on: #001 (Record Codec)
- Depends on: #003 (Sequential Log Reader)
- Depends on: #004 (In-Memory Hash Index)
- Depends on: #005 (Basic KV API)
- Enables: #009 (Log Rotation)
