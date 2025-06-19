---
name: Concurrency/Race Condition
about: Report data races, deadlocks, or other concurrency issues
title: '[CONCURRENCY] '
labels: concurrency, bug
assignees: ''
---

## Concurrency Issue Type

- [ ] Data race
- [ ] Deadlock
- [ ] Livelock
- [ ] Performance degradation under concurrency
- [ ] Inconsistent behavior under load
- [ ] Other concurrency issue

## Race Detector Output

Run your test with the race detector and paste the output:

```bash
go test -race ./...
```

```text
<!-- Paste race detector output here -->
```

## Reproduction Steps

1. Set up concurrent environment with...
2. Execute operations...
3. Observe issue...

## Minimal Reproduction Case

```go
package main

import (
    "sync"
    "testing"
    "github.com/ssargent/freyjadb/pkg/bptree"
    "github.com/segmentio/ksuid"
)

func TestConcurrencyIssue(t *testing.T) {
    tree := bptree.NewBPlusTree(4)
    var wg sync.WaitGroup
    
    // Your reproduction code here
}
```

## Expected Behavior

What should happen when running concurrent operations?

## Actual Behavior

What actually happens? Include any:

- Panics
- Incorrect results
- Performance degradation
- Hangs/timeouts

## Concurrency Details

- **Number of goroutines**: 
- **Types of operations**: [Insert/Search/Delete/Mixed]
- **Data size**: 
- **Duration of test**: 
- **Frequency of issue**: [Always/Often/Sometimes/Rarely]

## Environment

- **OS**: [e.g. macOS 14.0, Ubuntu 22.04]
- **Go Version**: [e.g. 1.21.0]
- **FreyjaDB Version/Commit**: [e.g. v0.1.0 or commit hash]
- **Architecture**: [e.g. amd64, arm64]
- **CPU cores**: [e.g. 8 cores]
- **GOMAXPROCS**: [if set explicitly]

## Additional Tools Output

If you've used other debugging tools, include their output:

### Go Trace
```bash
go test -trace=trace.out
go tool trace trace.out
```

### Deadlock Detection
```bash
# If using github.com/sasha-s/go-deadlock
```

### CPU Profile
```bash
go test -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

## Impact

- [ ] Development only - doesn't affect production
- [ ] Potential production issue
- [ ] Active production issue
- [ ] Critical - causes data corruption

## Potential Root Cause

If you have ideas about what might be causing this:

- [ ] Missing mutex protection
- [ ] Lock ordering issue
- [ ] Reader-writer lock misuse
- [ ] Atomic operation issue
- [ ] Channel operation issue
- [ ] Other: ___________

## Additional Context

Any other information that might help diagnose the issue.
