---
name: Feature request
about: Suggest an idea for FreyjaDB
title: '[FEATURE] Implement Compaction/Merge Utility'
labels: enhancement, priority-high, complexity-high
assignees: ''
---

## Is your feature request related to a problem? Please describe

FreyjaDB needs compaction to reclaim disk space by removing obsolete records (old versions, tombstones) and consolidating data files. Without compaction, disk usage grows indefinitely even when data is deleted.

## Describe the solution you'd like

Implement an offline compaction utility that:

- Reads all data files and identifies the latest version of each key
- Removes tombstone records and obsolete key versions
- Writes compacted data to fresh files with corresponding hint files
- Provides significant disk space reclamation

## Describe alternatives you've considered

- Online compaction (more complex, addressed in later issues)
- Partial compaction (less efficient space reclamation)
- External compaction tools (would require format knowledge)

## Use Case

This feature addresses the need for:

- Disk space reclamation in production environments
- Removing accumulated tombstones and old versions
- Improved read performance through data consolidation
- Foundation for online compaction mechanisms

## API Design (if applicable)

```go
type CompactionConfig struct {
    InputDir    string
    OutputDir   string
    MaxFileSize int64
    Concurrency int
}

type CompactionResult struct {
    InputFiles       []string
    OutputFiles      []string
    SpaceSaved       int64
    RecordsProcessed int64
    RecordsKept      int64
    CompactionTime   time.Duration
}

type Compactor struct {
    config CompactionConfig
}

func NewCompactor(config CompactionConfig) *Compactor {
    // Implementation details
}

func (c *Compactor) Compact() (*CompactionResult, error) {
    // Perform offline compaction
}

func (c *Compactor) scanAllFiles() (map[string]*LatestRecord, error) {
    // Build map of latest record for each key
}

func (c *Compactor) writeCompactedFiles(records map[string]*LatestRecord) error {
    // Write compacted data and hint files
}

type LatestRecord struct {
    Key       []byte
    Value     []byte
    Timestamp uint64
    IsTombstone bool
}
```

## Implementation Details

- [x] Multi-file scanning to find latest record versions
- [x] Tombstone filtering and obsolete record removal
- [x] Output file generation with size limits
- [x] Hint file generation for fast bootstrap
- [x] Progress reporting and statistics collection

## Performance Considerations

- Expected performance impact: High I/O during compaction, space savings afterward
- Memory usage changes: Significant increase during compaction for record tracking
- Disk space impact: Major reduction in space usage

## Testing Strategy

- [x] Unit tests with known data patterns
- [x] Space reclamation verification
- [x] Data integrity validation post-compaction
- [x] Large dataset compaction performance
- [x] Tombstone and version handling correctness

## Documentation Requirements

- [x] Compaction process documentation
- [x] Configuration options and tuning
- [x] Performance characteristics and resource usage
- [x] Best practices for compaction scheduling

## Additional Context

This is item #10 from the project roadmap with complexity rating 4/5. Compaction is crucial for production databases but requires careful implementation to ensure data integrity.

## Acceptance Criteria

- [ ] Reads all data files and processes every record
- [ ] Keeps only the latest version of each key
- [ ] Removes tombstone records completely
- [ ] Generates fresh data files with compacted content
- [ ] Creates corresponding hint files for fast bootstrap
- [ ] Achieves significant space reduction
- [ ] Maintains data integrity throughout process

## Priority

- [x] High - Essential for production disk space management

## Related Issues

- Depends on: #001 (Record Codec)
- Depends on: #003 (Sequential Log Reader)
- Depends on: #007 (Hint File Format)
- Depends on: #009 (Log Rotation)
- Enables: #011 (Hot-Swap Compaction)
