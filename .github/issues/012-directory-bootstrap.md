---
name: Feature request
about: Suggest an idea for FreyjaDB
title: '[FEATURE] Implement Directory Bootstrap for Multi-File Management'
labels: enhancement, priority-medium
assignees: ''
---

## Is your feature request related to a problem? Please describe

FreyjaDB needs to handle directories containing multiple data files from log rotation and compaction. The system must automatically discover, order, and bootstrap from all available data files in the correct sequence.

## Describe the solution you'd like

Implement directory bootstrap that:

- Scans directory for all *.data files using glob patterns
- Orders files by numeric ID (001.data, 002.data, etc.)
- Loads corresponding hint files when available
- Rebuilds complete index from all files in sequence
- Survives restarts and merges gracefully

## Describe alternatives you've considered

- Manual file specification (not suitable for production)
- Single file only (limits scalability)
- External file management (adds operational complexity)

## Use Case

This feature addresses the need for:

- Automatic discovery of rotated log files
- Consistent ordering of data files for index building
- Integration testing across restarts and compaction cycles
- Production deployment with multiple data segments

## API Design (if applicable)

```go
type DirectoryBootstrap struct {
    dataDir   string
    fileGlob  string
    index     *HashIndex
}

type BootstrapSummary struct {
    FilesFound    []string
    FilesLoaded   []string
    HintsUsed     []string
    KeysLoaded    int64
    TotalSize     int64
    BootstrapTime time.Duration
    Errors        []error
}

func NewDirectoryBootstrap(dataDir string) *DirectoryBootstrap {
    // Implementation details
}

func (db *DirectoryBootstrap) ScanDataFiles() ([]string, error) {
    // Glob for *.data files and sort by ID
}

func (db *DirectoryBootstrap) LoadAllFiles() (*BootstrapSummary, error) {
    // Load all data files in correct order
}

func (db *DirectoryBootstrap) extractFileID(filename string) (uint32, error) {
    // Parse numeric ID from filename
}

func (db *DirectoryBootstrap) validateFileSequence(files []string) error {
    // Ensure no gaps in file sequence
}
```

## Implementation Details

- [x] Glob-based file discovery with pattern matching
- [x] Numeric sorting of file IDs for correct ordering
- [x] Integration with existing hint-driven bootstrap
- [x] Gap detection in file sequences
- [x] Error handling for corrupted or missing files

## Performance Considerations

- Expected performance impact: Startup cost, runtime neutral
- Memory usage changes: Temporary increase during bootstrap
- Disk space impact: No change (read-only operation)

## Testing Strategy

- [x] Multi-file bootstrap with known sequences
- [x] Missing file and gap handling
- [x] Integration test: restart after rotation and compaction
- [x] Performance testing with many files
- [x] Hint file integration verification

## Documentation Requirements

- [x] File naming convention requirements
- [x] Directory structure documentation
- [x] Bootstrap process flow
- [x] Error recovery procedures

## Additional Context

This is item #12 from the project roadmap with complexity rating 2/5. Directory bootstrap is essential for production deployment and enables automated recovery from complex file layouts.

## Acceptance Criteria

- [ ] Automatically discovers all *.data files in directory
- [ ] Orders files correctly by numeric ID
- [ ] Loads files in sequence to rebuild index
- [ ] Handles missing files gracefully
- [ ] Integration tests pass for restart scenarios
- [ ] Works correctly after compaction operations

## Priority

- [x] Medium - Important for production file management

## Related Issues

- Depends on: #003 (Sequential Log Reader)
- Depends on: #008 (Hint-Driven Bootstrap)
- Depends on: #009 (Log Rotation)
- Enhances: #010 (Compaction/Merge Utility)
