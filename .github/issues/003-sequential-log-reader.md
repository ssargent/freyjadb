---
name: Feature request
about: Suggest an idea for FreyjaDB
title: '[FEATURE] Implement Sequential Log Reader'
labels: enhancement, priority-high
assignees: ''
---

## Status

**Completed** - LogReader implementation verified in `pkg/store/log_reader.go` with CRC validation and sequential access.

## Is your feature request related to a problem? Please describe

FreyjaDB needs a sequential log reader to iterate through records in log files for index building, compaction, and recovery operations. This complements the append-only log writer.

## Describe the solution you'd like

Implement a sequential log reader that:

- Iterates through records from a specified offset
- Validates CRC checksums for each record
- Handles end-of-file and malformed record detection
- Provides efficient sequential access patterns

## Describe alternatives you've considered

- Random access reads (inefficient for sequential scanning)
- Memory mapping entire files (memory usage concerns)
- Stream-based parsing (more complex error handling)

## Use Case

This feature addresses the need for:

- Building in-memory indexes from log files
- Compaction operations that scan all records
- Crash recovery by validating file integrity
- Data export and analysis tools

## API Design (if applicable)

```go
type LogReader struct {
    file   *os.File
    codec  *RecordCodec
    offset int64
}

type LogReaderConfig struct {
    FilePath    string
    StartOffset int64
}

func NewLogReader(config LogReaderConfig) (*LogReader, error) {
    // Implementation details
}

func (r *LogReader) Next() (*Record, error) {
    // Read next record from current offset
}

func (r *LogReader) Seek(offset int64) error {
    // Move to specific file offset
}

func (r *LogReader) Close() error {
    // Cleanup resources
}

// Iterator interface for range-based loops
func (r *LogReader) Iterator() <-chan RecordResult {
    // Channel-based iteration
}

type RecordResult struct {
    Record *Record
    Error  error
}
```

## Implementation Details

- [x] Sequential file reading with buffering
- [x] CRC validation for each record
- [x] Graceful handling of EOF and truncated records
- [x] Offset tracking for resumable reads
- [x] Integration with RecordCodec for decoding

## Performance Considerations

- Expected performance impact: Positive (optimized for sequential reads)
- Memory usage changes: Small increase for read buffers
- Disk space impact: No change (read-only operation)

## Testing Strategy

- [x] Unit tests for sequential record iteration
- [x] CRC validation and corruption detection
- [x] EOF handling and boundary conditions
- [x] Multiple files with known record counts
- [x] Performance benchmarks for read throughput

## Documentation Requirements

- [x] API documentation for LogReader interface
- [x] Usage examples for iteration patterns
- [x] Error handling documentation
- [x] Performance characteristics

## Additional Context

This is item #3 from the project roadmap. The sequential reader is essential for all operations that need to scan log files, including index building and compaction.

## Acceptance Criteria

- [x] Can iterate through records from offset 0
- [x] Validates CRC for each record read
- [x] Stops gracefully at end of file
- [x] Handles corrupted/truncated records appropriately
- [x] File with N records yields exactly N record objects
- [x] Maintains accurate offset tracking

## Priority

- [x] High - Critical for index building and recovery

## Related Issues

- Depends on: #001 (Record Codec)
- Depends on: #002 (Append-Only Log Writer)
- Enables: #004 (In-Memory Hash Index Builder)
