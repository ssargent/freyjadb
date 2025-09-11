# FreyjaDB Development Guide

This guide provides comprehensive instructions for developers working on FreyjaDB. Whether you're adding new features, fixing bugs, or optimizing performance, this document will help you maintain code quality and follow best practices.

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [Development Workflow](#development-workflow)
- [Adding New Features](#adding-new-features)
- [Testing Guidelines](#testing-guidelines)
- [Code Quality Standards](#code-quality-standards)
- [Performance Considerations](#performance-considerations)
- [Documentation Requirements](#documentation-requirements)
- [Release Process](#release-process)

## 🚀 Quick Start

### Prerequisites

```bash
# Install Go 1.21+
go version

# Install development tools
make tools

# Clone and setup
git clone https://github.com/ssargent/freyjadb.git
cd freyjadb
go mod download
```

### Development Commands

```bash
# Fast development cycle
make test          # Run fast unit tests (~0.5s)
make build         # Build the binary
make dev           # Full development build

# Comprehensive testing
make test-all      # All tests including slow ones
make bench         # Performance benchmarks
make fuzz          # Fuzz tests
```

## 🔄 Development Workflow

### 1. Choose Your Development Path

**For New Features:**
```bash
# Start with fast feedback
make test
# Implement feature incrementally
# Run tests frequently
make test-parallel
```

**For Bug Fixes:**
```bash
# Reproduce the issue
# Write failing test first
make test
# Fix the bug
# Verify fix
make test-race
```

**For Performance Optimization:**
```bash
# Establish baseline
make bench
# Implement optimization
# Measure improvement
make bench-cpu
```

### 2. Branch Strategy

```bash
# Create feature branch
git checkout -b feature/your-feature-name

# Create bug fix branch
git checkout -b fix/issue-description

# Create performance branch
git checkout -b perf/optimization-name
```

### 3. Commit Guidelines

```bash
# Write clear, descriptive commits
git commit -m "feat: add secondary index support for field queries

- Implement B+Tree-based secondary indexes
- Add JSON field extraction
- Support equality and range queries
- Include comprehensive tests

Closes #123"

# Use conventional commit format
# Types: feat, fix, docs, style, refactor, perf, test, chore
```

## ✨ Adding New Features

### Step 1: Plan Your Feature

1. **Check existing issues** in `.github/issues/` for similar features
2. **Review project roadmap** in `docs/project_plan.md`
3. **Consider impact** on existing architecture
4. **Plan testing strategy** upfront

### Step 2: Implement Incrementally

```go
// Example: Adding a new query operator

// 1. Define the interface
type QueryOperator interface {
    Evaluate(value interface{}) bool
    String() string
}

// 2. Implement the operator
type LikeOperator struct {
    pattern string
}

func (op *LikeOperator) Evaluate(value interface{}) bool {
    // Implementation
    return false // TODO
}

// 3. Add to parser
func parseOperator(op string) (QueryOperator, error) {
    switch op {
    case "LIKE":
        return &LikeOperator{}, nil
    // ... existing operators
    }
    return nil, fmt.Errorf("unknown operator: %s", op)
}
```

### Step 3: Test Early and Often

```go
// Write tests first (TDD approach)
func TestLikeOperator_Evaluate(t *testing.T) {
    tests := []struct {
        name     string
        pattern  string
        value    string
        expected bool
    }{
        {"exact match", "john", "john", true},
        {"wildcard", "j%n", "john", true},
        {"no match", "jane", "john", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            op := &LikeOperator{pattern: tt.pattern}
            result := op.Evaluate(tt.value)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Step 4: Performance Considerations

```go
// Add benchmarks for new features
func BenchmarkLikeOperator_Evaluate(b *testing.B) {
    op := &LikeOperator{pattern: "test%pattern"}
    value := "test string to match"

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        op.Evaluate(value)
    }
}

// Profile memory usage
func BenchmarkLikeOperator_Memory(b *testing.B) {
    b.ReportAllocs()
    // ... benchmark code
}
```

### Step 5: Documentation

```go
// Document your feature in code
type LikeOperator struct {
    // Pattern supports SQL LIKE syntax with % wildcards
    // Examples: "john%", "%smith", "j%n"
    pattern string
}

// Add usage examples to README
// ## LIKE Operator
//
// Query records with pattern matching:
//
// ```go
// query := FieldQuery{
//     Field:    "name",
//     Operator: "LIKE",
//     Value:    "John%",
// }
// ```
```

## 🧪 Testing Guidelines

### Test Categories

| Test Type | Location | Run Command | Purpose |
|-----------|----------|-------------|---------|
| Unit Tests | `*_test.go` | `make test` | Isolated functionality |
| Integration | `tests/integration/` | `make test-integration` | Cross-component |
| Benchmarks | `*_bench_test.go` | `make bench` | Performance |
| Fuzz Tests | `*_fuzz_test.go` | `make fuzz` | Property-based |
| Race Tests | Any test | `make test-race` | Concurrency |

### Writing Effective Tests

#### 1. Unit Tests

```go
func TestFeatureName_Scenario(t *testing.T) {
    // Given: Setup test data
    setup := createTestSetup()

    // When: Execute the feature
    result, err := setup.feature.Execute(input)

    // Then: Verify expectations
    require.NoError(t, err)
    assert.Equal(t, expectedResult, result)
}
```

#### 2. Table-Driven Tests

```go
func TestFieldExtractor_Extract(t *testing.T) {
    extractor := &JSONFieldExtractor{}

    tests := []struct {
        name     string
        json     string
        field    string
        expected interface{}
        wantErr  bool
    }{
        {"string field", `{"name":"John"}`, "name", "John", false},
        {"number field", `{"age":25}`, "age", float64(25), false},
        {"missing field", `{"name":"John"}`, "email", nil, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := extractor.Extract([]byte(tt.json), tt.field)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected, result)
            }
        })
    }
}
```

#### 3. Benchmark Tests

```go
func BenchmarkFeature_Operation(b *testing.B) {
    // Setup
    setup := createBenchmarkSetup()

    // Reset timer before actual benchmark
    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        setup.feature.Operation()
    }
}
```

#### 4. Fuzz Tests

```go
func FuzzFeature_ParseInput(f *testing.F) {
    // Add seed corpus
    f.Add([]byte("valid input"))
    f.Add([]byte("another valid case"))

    f.Fuzz(func(t *testing.T, input []byte) {
        // Skip invalid inputs that would cause panics
        if len(input) > 10000 {
            t.Skip("Input too large")
        }

        // Test that parsing doesn't crash
        result, err := feature.ParseInput(input)
        if err != nil {
            // Errors are OK, panics are not
            return
        }

        // If parsing succeeded, validate result
        assert.NotNil(t, result)
    })
}
```

### Test Coverage Goals

- **Unit Tests**: 80%+ coverage
- **Integration Tests**: Key user journeys
- **Performance Tests**: Critical paths
- **Edge Cases**: Error conditions, boundary values

## 📏 Code Quality Standards

### Go Best Practices

```go
// ✅ Good: Clear naming and structure
type UserService struct {
    repo UserRepository
}

func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    if id == "" {
        return nil, errors.New("user ID cannot be empty")
    }

    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user %s: %w", id, err)
    }

    return user, nil
}

// ❌ Bad: Unclear naming, no error handling
func getUser(id string) *User {
    return repo.find(id) // No context, no error handling
}
```

### Code Organization

```
pkg/
├── feature/           # Feature package
│   ├── feature.go     # Main implementation
│   ├── feature_test.go # Unit tests
│   └── feature_bench_test.go # Benchmarks
├── internal/          # Internal packages
└── tests/             # Integration tests
    └── integration/
```

### Error Handling

```go
// ✅ Good: Structured errors
var (
    ErrUserNotFound    = errors.New("user not found")
    ErrInvalidUserID   = errors.New("invalid user ID")
)

func (s *UserService) GetUser(id string) (*User, error) {
    if !isValidID(id) {
        return nil, fmt.Errorf("%w: %s", ErrInvalidUserID, id)
    }

    user := s.findUser(id)
    if user == nil {
        return nil, fmt.Errorf("%w: %s", ErrUserNotFound, id)
    }

    return user, nil
}
```

## ⚡ Performance Considerations

### Profiling Workflow

```bash
# CPU profiling
make bench-cpu
go tool pprof cpu.prof

# Memory profiling
make bench-mem
go tool pprof mem.prof

# Trace analysis
go test -trace=trace.out ./...
go tool trace trace.out
```

### Performance Best Practices

1. **Minimize allocations** in hot paths
2. **Use object pooling** for frequently created objects
3. **Profile before optimizing**
4. **Consider cache effects** in data structures
5. **Use appropriate data types** for your use case

### Memory Management

```go
// ✅ Good: Reuse buffers
type BufferPool struct {
    pool sync.Pool
}

func (bp *BufferPool) Get() *bytes.Buffer {
    if buf := bp.pool.Get(); buf != nil {
        return buf.(*bytes.Buffer)
    }
    return &bytes.Buffer{}
}

func (bp *BufferPool) Put(buf *bytes.Buffer) {
    buf.Reset()
    bp.pool.Put(buf)
}

// ❌ Bad: Creating buffers repeatedly
func processData(data []byte) []byte {
    buf := &bytes.Buffer{} // Allocates every time
    // ... process data
    return buf.Bytes()
}
```

## 📚 Documentation Requirements

### Code Documentation

```go
// Package feature provides functionality for [describe purpose].
//
// It supports [key features] and is designed for [use cases].
//
// Example usage:
//
//	feature := NewFeature()
//	result := feature.Process(input)
package feature

// Feature represents a [describe what it does].
//
// It maintains [important state] and provides [key methods].
type Feature struct {
    // Config contains feature configuration
    config Config
}

// NewFeature creates a new Feature instance.
//
// It initializes with default settings and validates the configuration.
func NewFeature(config Config) (*Feature, error) {
    // Implementation
}

// Process performs the main feature operation.
//
// It takes input data and returns processed results.
// The operation is thread-safe and can be called concurrently.
//
// Parameters:
//   - input: The data to process
//
// Returns:
//   - result: The processed data
//   - error: Any error that occurred during processing
func (f *Feature) Process(input []byte) ([]byte, error) {
    // Implementation
}
```

### README Updates

When adding features, update the main README.md:

```markdown
## New Feature: Field Queries

FreyjaDB now supports field-based queries on JSON values:

```go
// Query users by age
query := FieldQuery{
    Field:    "age",
    Operator: ">=",
    Value:    18,
}
```

### Key Features

- Support for equality and range queries
- Automatic secondary index creation
- JSON field extraction
- Streaming results
```

### API Documentation

```go
// API documentation should include:
// - Method signatures with parameter descriptions
// - Return value descriptions
// - Error conditions
// - Thread safety guarantees
// - Performance characteristics
// - Usage examples
```

## 🚀 Release Process

### Pre-Release Checklist

- [ ] All tests pass: `make test-all`
- [ ] Benchmarks show no regressions: `make bench`
- [ ] Code coverage meets requirements: `make test-cover`
- [ ] Linting passes: `make lint`
- [ ] Documentation updated
- [ ] Changelog updated
- [ ] Breaking changes documented

### Release Commands

```bash
# Final verification
make ci

# Create release
git tag v1.2.3
git push origin v1.2.3

# Build release binaries
make build-linux
```

### Version Numbering

Follow [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

## 🎯 Development Tips

### 1. Start Small
- Break features into small, testable increments
- Get feedback early and often
- Use feature flags for experimental features

### 2. Test-Driven Development
- Write tests before implementation
- Use table-driven tests for comprehensive coverage
- Test error conditions thoroughly

### 3. Performance-First Mindset
- Profile early, profile often
- Consider memory usage in design decisions
- Use appropriate data structures

### 4. Documentation Culture
- Document as you code
- Keep READMEs up to date
- Write clear commit messages

### 5. Code Review Checklist
- [ ] Tests included and passing
- [ ] Code follows Go conventions
- [ ] Documentation updated
- [ ] Performance impact considered
- [ ] Error handling appropriate
- [ ] Thread safety verified

## 📞 Getting Help

- **Issues**: Use GitHub issues for bugs and feature requests
- **Discussions**: Use GitHub discussions for questions
- **Code Reviews**: All changes require review before merge
- **Architecture**: Review `docs/project_plan.md` for design decisions

---

Happy coding! Remember: quality over speed, but fast feedback loops are essential. 🏗️✨