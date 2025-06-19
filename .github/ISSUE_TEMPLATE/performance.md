---
name: Performance issue
about: Report performance problems or bottlenecks in FreyjaDB
title: '[PERFORMANCE] '
labels: performance
assignees: ''
---

## Performance Issue Description

A clear and concise description of the performance problem.

## Current Performance

Describe the current performance metrics:

- **Operations per second**: 
- **Memory usage**: 
- **CPU usage**: 
- **Response time**: 
- **Database size**: 
- **Concurrent connections**: 

## Expected Performance

Describe what performance you expected:

- **Expected operations per second**: 
- **Expected memory usage**: 
- **Expected CPU usage**: 
- **Expected response time**: 

## Benchmark/Test Case

```go
// Include the benchmark or test case that demonstrates the issue
func BenchmarkExample(b *testing.B) {
    tree := bptree.NewBPlusTree(4)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        key := []byte(fmt.Sprintf("key%d", i))
        value := ksuid.New()
        tree.Insert(key, value)
    }
}
```

## Profiling Data

If you have profiling data, please attach it here:

```bash
# CPU Profile
go test -cpuprofile=cpu.prof -bench=.

# Memory Profile
go test -memprofile=mem.prof -bench=.

# Trace
go test -trace=trace.out -bench=.
```

## Environment

- **OS**: [e.g. macOS 14.0, Ubuntu 22.04]
- **Go Version**: [e.g. 1.21.0]
- **FreyjaDB Version/Commit**: [e.g. v0.1.0 or commit hash]
- **Hardware**: 
  - CPU: [e.g. Intel i7, Apple M2, AMD Ryzen]
  - RAM: [e.g. 16GB, 32GB]
  - Storage: [e.g. SSD, NVMe, HDD]
- **Dataset size**: [number of records, data size]
- **Concurrency level**: [number of goroutines/threads]

## Reproduction Steps

1. Set up environment with...
2. Load data using...
3. Execute operation...
4. Measure performance with...

## Additional Context

Add any other context about the performance issue here:

- Does this happen with all data sizes?
- Is it specific to certain operations?
- Any patterns you've noticed?

## Potential Solutions

If you have ideas for optimizations, please describe them here:

- [ ] Algorithm improvements
- [ ] Data structure changes
- [ ] Caching strategies
- [ ] Concurrency optimizations
- [ ] Memory management improvements

## Impact

- [ ] Minor - Slightly slower than expected
- [ ] Moderate - Noticeable performance degradation
- [ ] Major - Significantly impacts usability
- [ ] Critical - Makes system unusable
