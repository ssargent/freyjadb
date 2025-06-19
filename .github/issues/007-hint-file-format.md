---
name: Feature request
about: Suggest an idea for FreyjaDB
title: '[FEATURE] Implement Hint File Format for Fast Bootstrap'
labels: enhancement, priority-medium
assignees: ''
---

## Is your feature request related to a problem? Please describe

FreyjaDB needs a hint file format to enable fast database startup by avoiding full log scans. Hint files contain compact index information that can be quickly loaded into memory.

## Describe the solution you'd like

Implement a hint file format that:

- Stores compact index information: {key_hash, file_id, offset, size}
- Is generated during compaction/merge operations
- Enables O(keys) startup time instead of O(data_size)
- Maintains consistency with corresponding data files

## Describe alternatives you've considered

- Full log scanning on startup (too slow for large databases)
- Persistent index files (more complex, less crash-safe)
- No hint files (acceptable for small databases only)

## Use Case

This feature addresses the need for:

- Fast database startup times regardless of data size
- Efficient bootstrap for large databases
- Foundation for hint-driven index loading
- Preparation for compaction operations

## API Design (if applicable)

```go
type HintRecord struct {
    KeyHash  uint64  // Hash of the key for verification
    FileID   uint32  // Which data file contains this key
    Offset   int64   // Byte offset within the file
    Size     uint32  // Size of the record
}

type HintWriter struct {
    file  *os.File
    codec *HintCodec
}

type HintReader struct {
    file  *os.File
    codec *HintCodec
}

func NewHintWriter(filePath string) (*HintWriter, error) {
    // Implementation details
}

func (hw *HintWriter) WriteHint(key []byte, fileID uint32, offset int64, size uint32) error {
    // Write hint record
}

func NewHintReader(filePath string) (*HintReader, error) {
    // Implementation details
}

func (hr *HintReader) ReadAllHints() ([]HintRecord, error) {
    // Read all hint records
}
```

## Implementation Details

- [x] Compact binary format for hint records
- [x] Key hashing for verification without storing full keys
- [x] File ID and offset tracking
- [x] Integration with compaction process
- [x] Checksums for hint file integrity

## Performance Considerations

- Expected performance impact: Positive (faster startup)
- Memory usage changes: Small increase for hint processing
- Disk space impact: Small increase for hint files

## Testing Strategy

- [x] Unit tests for hint record encoding/decoding
- [x] Hint file generation during compaction
- [x] Verification of hint record accuracy
- [x] Performance comparison: hint vs full scan
- [x] Corruption handling for hint files

## Documentation Requirements

- [x] Hint file format specification
- [x] Performance characteristics
- [x] Integration with compaction process
- [x] Troubleshooting guide

## Additional Context

This is item #7 from the project roadmap. Hint files are preparation for fast bootstrap (#8) and are generated during merge operations (#10).

## Acceptance Criteria

- [ ] Hint records accurately represent key locations
- [ ] Hint files are generated during merge operations
- [ ] Binary format is compact and efficient
- [ ] Key hashing provides reasonable collision resistance
- [ ] Hint files maintain consistency with data files
- [ ] File format is documented and stable

## Priority

- [x] Medium - Important for performance, not critical for functionality

## Related Issues

- Enables: #008 (Hint-Driven Fast Bootstrap)
- Part of: #010 (Compaction/Merge Utility)
- Depends on: #001 (Record Codec)
