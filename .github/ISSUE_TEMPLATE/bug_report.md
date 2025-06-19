---
name: Bug report
about: Create a report to help us improve FreyjaDB
title: '[BUG] '
labels: bug
assignees: ''
---

## Bug Description
A clear and concise description of what the bug is.

## To Reproduce
Steps to reproduce the behavior:
1. Initialize FreyjaDB with '...'
2. Execute operation '....'
3. Observe behavior '....'
4. See error

## Expected Behavior
A clear and concise description of what you expected to happen.

## Actual Behavior
A clear and concise description of what actually happened.

## Error Messages/Stack Traces
```
Paste any error messages, stack traces, or log output here
```

## Code Sample
```go
// Minimal code sample that reproduces the issue
package main

import "github.com/ssargent/freyjadb"

func main() {
    // Your code here
}
```

## Environment
- **OS**: [e.g. macOS 14.0, Ubuntu 22.04, Windows 11]
- **Go Version**: [e.g. 1.21.0]
- **FreyjaDB Version/Commit**: [e.g. v0.1.0 or commit hash]
- **Architecture**: [e.g. amd64, arm64]
- **Database Size**: [e.g. number of records, approximate data size]

## Race Detector Output (if applicable)
If you're experiencing concurrency issues, please run with `-race` flag and include output:
```
go test -race ./...
```

## Performance Impact
- [ ] No performance impact
- [ ] Minor performance degradation
- [ ] Major performance degradation
- [ ] System becomes unusable

## Additional Context
Add any other context about the problem here, such as:
- When did this start happening?
- Does it happen consistently or intermittently?
- Any recent changes that might be related?

## Possible Solution
If you have ideas on how to fix this, please describe them here.

## Related Issues
Link to any related issues or PRs using #issue_number
