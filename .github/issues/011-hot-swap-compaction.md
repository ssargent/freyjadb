---
name: Feature request
about: Suggest an idea for FreyjaDB
title: '[FEATURE] Implement Hot-Swap Compaction'
labels: enhancement, priority-high, complexity-high
assignees: ''
---

## Is your feature request related to a problem? Please describe

FreyjaDB needs online compaction that doesn't require database downtime. The offline compaction utility requires stopping the database, which is unacceptable for production systems requiring high availability.

## Describe the solution you'd like

Implement hot-swap compaction that:

- Runs compaction concurrently with normal database operations
- Atomically swaps old files with compacted files
- Reloads the index with new file locations
- Ensures readers never miss keys during the swap

## Describe alternatives you've considered

- Stop-the-world compaction (unacceptable downtime)
- Incremental compaction (more complex coordination)
- Copy-on-write compaction (higher memory usage)

## Use Case

This feature addresses the need for:

- Zero-downtime disk space reclamation
- Production database maintenance without service interruption
- Automatic background compaction scheduling
- Maintaining read availability during compaction

## API Design (if applicable)

```go
type HotSwapCompactor struct {
    store     *KVStore
    compactor *Compactor
    swapMutex sync.Mutex
}

type HotSwapConfig struct {
    TempDir           string
    MinSpaceSaving    float64  // Only compact if saving > X%
    CompactionTimeout time.Duration
}

func NewHotSwapCompactor(store *KVStore, config HotSwapConfig) *HotSwapCompactor {
    // Implementation details
}

func (hsc *HotSwapCompactor) CompactInBackground() error {
    // Run compaction without blocking operations
}

func (hsc *HotSwapCompactor) atomicSwap(compactedFiles []string) error {
    // Atomically replace old files with compacted ones
}

func (hsc *HotSwapCompactor) reloadIndex() error {
    // Rebuild index from new files
}

// Events for monitoring
type CompactionEvent struct {
    Type      string // "started", "completed", "failed"
    Files     []string
    SpaceSaved int64
    Duration  time.Duration
}
```

## Implementation Details

- [x] Background compaction execution
- [x] Atomic file replacement using rename operations
- [x] Index hot-swapping with minimal lock time
- [x] Reader safety during file transitions
- [x] Rollback mechanisms for failed compactions

## Performance Considerations

- Expected performance impact: Brief pause during swap, long-term performance improvement
- Memory usage changes: Temporary increase during compaction
- Disk space impact: Major reduction after successful compaction

## Testing Strategy

- [x] Concurrent read/write operations during compaction
- [x] Atomic swap correctness under load
- [x] Reader consistency throughout compaction
- [x] Failure recovery and rollback scenarios
- [x] Performance impact measurement

## Documentation Requirements

- [x] Hot-swap process documentation
- [x] Safety guarantees and limitations
- [x] Performance impact during compaction
- [x] Monitoring and alerting recommendations

## Additional Context

This is item #11 from the project roadmap with complexity rating 4/5. Hot-swap compaction is critical for production systems but requires extremely careful implementation to maintain data consistency.

## Acceptance Criteria

- [ ] Compaction runs concurrently with database operations
- [ ] File replacement is atomic and crash-safe
- [ ] Readers never experience key lookup failures during swap
- [ ] Index is updated consistently with new file locations
- [ ] Failed compactions can be safely rolled back
- [ ] Monitoring events are emitted for operational visibility

## Priority

- [x] High - Essential for production zero-downtime operation

## Related Issues

- Depends on: #010 (Compaction/Merge Utility)
- Depends on: #004 (In-Memory Hash Index)
- Depends on: #013 (Read/Write Concurrency)
- Enables: #019 (Background Compaction Scheduler)
