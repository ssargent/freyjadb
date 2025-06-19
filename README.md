# FreyjaDB

A **small, embeddable, log-structured key-value database** inspired by DynamoDB and built with ❤️ in Go.

FreyjaDB is designed to be a high-performance, crash-safe storage engine that supports:

- **SSD-friendly speed** with append-only writes and single random reads per GET
- **Crash safety** with CRC-guarded records and automatic recovery
- **High concurrency** with many readers and single writer architecture
- **Optional DynamoDB-like semantics** with Partition Key / Sort Key support
- **Educational codebase** with comprehensive tests and incremental development

## 🎯 Project Goals

FreyjaDB implements a **Bitcask-style storage engine** with the following characteristics:

- **Point lookups** with excellent performance on SSDs
- **Append-only log structure** for write optimization
- **Crash recovery** with automatic tail truncation and index rebuilding
- **Concurrent access** using lock-light reader/writer patterns
- **Space efficiency** through background compaction and merge operations
- **DynamoDB-inspired API** with partition and sort key semantics

## 🚀 Current Status

This project is under active development following a structured roadmap of 21 incremental milestones. Each component is built with comprehensive unit tests and integration tests.

### Development Workflow

```bash
# Build the project
make build

# Run tests
make test

# Run linter
make lint

# Run all checks (format, lint, test, build)
make all

# See all available targets
make help
```

## 🏗️ Architecture Overview

FreyjaDB uses a **log-structured merge (LSM) tree** approach with:

1. **Append-only log files** for all writes
2. **In-memory hash index** for fast key lookups  
3. **Background compaction** to reclaim space from deleted/updated records
4. **Hint files** for fast startup and index rebuilding
5. **Partition layer** for DynamoDB-style data organization

### Core Components

- **Record Codec**: Serialization format with CRC32 validation
- **Log Writer**: Append-only writes with fsync batching
- **Log Reader**: Sequential scanning with corruption detection
- **Hash Index**: In-memory key → file offset mapping
- **Compaction Engine**: Background merge and space reclamation
- **Partition Manager**: Multi-keyspace support with sort key ranges

## 📖 Usage

> **Note**: FreyjaDB is currently in early development. The API is subject to change.

```go
// Basic usage example (planned API)
db, err := freyjadb.Open("./data")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Simple key-value operations
err = db.Put("user:123", []byte("john@example.com"))
value, err := db.Get("user:123")
err = db.Delete("user:123")

// Partition key operations (DynamoDB-style)
err = db.PutItem("users", "john@example.com", map[string]interface{}{
    "email": "john@example.com",
    "name":  "John Doe",
    "age":   30,
})

item, err := db.GetItem("users", "john@example.com")
```

## 🛠️ Development

FreyjaDB follows a test-driven development approach with comprehensive automation:

- **Linting**: golangci-lint with strict rules
- **Testing**: Unit tests, integration tests, and property-based testing  
- **Benchmarking**: Performance testing for all critical paths
- **CI/CD**: Automated builds and testing
- **Documentation**: Inline docs and architecture guides

### Contributing

See our [Project Plan](docs/project_plan.md) for the complete development roadmap and current progress.

## 📄 License

FreyjaDB is released under the MIT License. See [LICENSE](LICENSE) for details.

---

> *FreyjaDB is named after Freyja, the Norse goddess associated with wisdom and foresight—qualities essential for a database that must reliably store and retrieve your data.*