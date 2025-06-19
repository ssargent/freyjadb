---
name: Feature request
about: Suggest an idea for FreyjaDB
title: '[FEATURE] Implement Log Rotation'
labels: enhancement, priority-high
assignees: ''
---

## Is your feature request related to a problem? Please describe

FreyjaDB needs log rotation to manage disk space and enable efficient compaction. Without rotation, the active log file would grow indefinitely, making operations slower and compaction impossible.

## Describe the solution you'd like

Implement log rotation that:

- Closes the current active.data file when it reaches a size limit
- Creates a new numbered data file (e.g., 001.data, 002.data)
- Seamlessly continues write operations across file boundaries
- Maintains proper file numbering and sequence

## Describe alternatives you've considered

- Time-based rotation (less predictable file sizes)
- Fixed number of records (variable file sizes)
- Manual rotation only (not suitable for production)

## Use Case

This feature addresses the need for:

- Manageable file sizes for backup and compaction
- Bounded memory usage for individual file operations
- Preparation for multi-file compaction
- Controlled disk space growth

## API Design (if applicable)

```go
type LogRotationConfig struct {
    MaxFileSize   int64         // Bytes before rotation
    DataDir       string        // Directory for data files
    FilePrefix    string        // Prefix for data files
}

type RotatedLogWriter struct {
    config      LogRotationConfig
    currentFile *LogWriter
    fileID      uint32
    currentSize int64
    mutex       sync.Mutex
}

func NewRotatedLogWriter(config LogRotationConfig) (*RotatedLogWriter, error) {
    // Implementation details
}

func (rlw *RotatedLogWriter) Put(key, value []byte) error {
    // Check size limit and rotate if needed
}

func (rlw *RotatedLogWriter) rotate() error {
    // Close current file and open next numbered file
}

func (rlw *RotatedLogWriter) getCurrentFileName() string {
    // Generate filename with current file ID
}
```

## Implementation Details

- [x] Size-based rotation with configurable thresholds
- [x] Atomic file switching to prevent data loss
- [x] Sequential file numbering (001.data, 002.data, etc.)
- [x] Thread-safe rotation operations
- [x] Proper cleanup and resource management

## Performance Considerations

- Expected performance impact: Neutral (brief pause during rotation)
- Memory usage changes: No significant change
- Disk space impact: Better control over individual file sizes

## Testing Strategy

- [x] Unit tests for rotation threshold detection
- [x] Cross-boundary write testing (writes spanning rotation)
- [x] File naming and sequencing verification
- [x] Concurrent write safety during rotation
- [x] Performance impact measurement

## Documentation Requirements

- [x] Configuration options for rotation settings
- [x] File naming conventions
- [x] Performance impact during rotation
- [x] Best practices for file size limits

## Additional Context

This is item #9 from the project roadmap. Log rotation is essential for enabling compaction and managing disk usage in production environments.

## Acceptance Criteria

- [ ] Files rotate when reaching configured size limit
- [ ] New files follow sequential naming convention
- [ ] Writes continue seamlessly across rotation boundaries
- [ ] No data loss during rotation operations
- [ ] File IDs increment correctly
- [ ] Configurable rotation thresholds work as expected

## Priority

- [x] High - Essential for production deployment

## Related Issues

- Depends on: #002 (Append-Only Log Writer)
- Depends on: #006 (Crash-Safe Reopen)
- Enables: #010 (Compaction/Merge Utility)
- Enables: #012 (Directory Bootstrap)
