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

## 🌳 B+ Tree Package

FreyjaDB includes a thread-safe B+ tree implementation that supports persistence and concurrent operations. The B+ tree is used internally for sort key range queries and can also be used as a standalone data structure.

### Basic B+ Tree Usage

```go
package main

import (
   "fmt"
   "log"

   "github.com/ssargent/freyjadb/pkg/bptree"
   "github.com/segmentio/ksuid"
)

func main() {
   // Create a new B+ tree with order 4
   tree := bptree.NewBPlusTree(4)

   // Insert key-value pairs
   key1 := []byte("user:alice")
   val1 := ksuid.New()
   tree.Insert(key1, val1)

   key2 := []byte("user:bob")
   val2 := ksuid.New()
   tree.Insert(key2, val2)

   // Search for values
   if value, found := tree.Search(key1); found {
       fmt.Printf("Found user:alice with ID %s\n", value.String())
   }

   // Delete a key
   if tree.Delete(key2) {
       fmt.Println("Successfully deleted user:bob")
   }
}
```

### Persistence: Saving and Loading B+ Trees

```go
package main

import (
   "fmt"
   "log"

   "github.com/ssargent/freyjadb/pkg/bptree"
   "github.com/segmentio/ksuid"
)

func main() {
   // Create and populate a B+ tree
   tree := bptree.NewBPlusTree(4)

   // Insert some data
   users := []string{"alice", "bob", "charlie", "diana"}
   for _, user := range users {
       key := []byte("user:" + user)
       val := ksuid.New()
       tree.Insert(key, val)
   }

   // Save the tree to disk
   filename := "./my_tree.dat"
   if err := tree.Save(filename); err != nil {
       log.Fatalf("Failed to save tree: %v", err)
   }
   fmt.Println("Tree saved to", filename)

   // Later, load the tree from disk
   loadedTree, err := bptree.LoadBPlusTree(filename)
   if err != nil {
       log.Fatalf("Failed to load tree: %v", err)
   }
   fmt.Printf("Loaded tree with height %d\n", loadedTree.Height())

   // Verify data integrity
   for _, user := range users {
       key := []byte("user:" + user)
       if _, found := loadedTree.Search(key); !found {
           log.Fatalf("Data integrity check failed for %s", user)
       }
   }
   fmt.Println("All data verified successfully")
}
```

### Checkpointing for Long-Running Applications

```go
package main

import (
   "fmt"
   "time"

   "github.com/ssargent/freyjadb/pkg/bptree"
   "github.com/segmentio/ksuid"
)

func main() {
   tree := bptree.NewBPlusTree(4)

   // Start automatic checkpointing every 30 seconds
   checkpointFile := "./tree_checkpoint.dat"
   tree.StartCheckpoint(checkpointFile, 30)
   fmt.Println("Checkpointing started - tree will be saved every 30 seconds")

   // Simulate ongoing operations
   for i := 0; i < 100; i++ {
       key := []byte(fmt.Sprintf("key:%d", i))
       val := ksuid.New()
       tree.Insert(key, val)

       // Simulate some work
       time.Sleep(100 * time.Millisecond)
   }

   // Stop checkpointing when done
   tree.StopCheckpoint()
   fmt.Println("Checkpointing stopped")

   // Final save
   if err := tree.Save("./final_tree.dat"); err != nil {
       fmt.Printf("Final save failed: %v\n", err)
   }
}
```

### Thread-Safe Concurrent Operations

```go
package main

import (
   "fmt"
   "sync"

   "github.com/ssargent/freyjadb/pkg/bptree"
   "github.com/segmentio/ksuid"
)

func main() {
   tree := bptree.NewBPlusTree(4)

   var wg sync.WaitGroup

   // Start multiple goroutines performing operations
   for i := 0; i < 5; i++ {
       wg.Add(1)
       go func(id int) {
           defer wg.Done()

           // Each goroutine inserts its own set of keys
           for j := 0; j < 20; j++ {
               key := []byte(fmt.Sprintf("goroutine:%d:key:%d", id, j))
               val := ksuid.New()
               tree.Insert(key, val)
           }

           // Each goroutine searches for some keys
           for j := 0; j < 10; j++ {
               key := []byte(fmt.Sprintf("goroutine:%d:key:%d", id, j))
               if _, found := tree.Search(key); !found {
                   fmt.Printf("Goroutine %d: key not found\n", id)
               }
           }
       }(i)
   }

   wg.Wait()
   fmt.Printf("All operations completed. Tree height: %d\n", tree.Height())
}
```

## 🔍 Query System

FreyjaDB now supports field-based queries on JSON-encoded record values, enabling efficient retrieval of records matching specific criteria.

### Field-Based Queries

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/ssargent/freyjadb/pkg/index"
    "github.com/ssargent/freyjadb/pkg/query"
)

func main() {
    // Create index manager and query engine
    indexManager := index.NewIndexManager(4)
    engine := query.NewSimpleQueryEngine(indexManager)
    extractor := &query.JSONFieldExtractor{}

    // Query for users aged 25 or older
    q := query.FieldQuery{
        Field:    "age",
        Operator: ">=",
        Value:    25,
    }

    iterator, err := engine.ExecuteQuery(context.Background(), "users", q, extractor)
    if err != nil {
        log.Fatal(err)
    }
    defer iterator.Close()

    // Stream results
    for iterator.Next() {
        result := iterator.Result()
        fmt.Printf("Found user: %s\n", string(result.Value))
    }
}
```

### Supported Query Types

- **Equality queries**: `field = value`
- **Range queries**: `field > value`, `field < value`, `field >= value`, `field <= value`
- **Range between**: Find records where field value is between two bounds

### Key Features

- **Automatic indexing**: Secondary indexes are created on-demand for queried fields
- **JSON support**: Built-in field extraction from JSON-encoded values
- **Streaming results**: Memory-efficient result iteration
- **Thread-safe**: Concurrent query execution
- **B+Tree powered**: Leverages existing B+Tree implementation for optimal performance

## 🛠️ Development

FreyjaDB follows a test-driven development approach with comprehensive automation:

- **Linting**: golangci-lint with strict rules
- **Testing**: Unit tests, integration tests, and property-based testing
- **Benchmarking**: Performance testing for all critical paths
- **CI/CD**: Automated builds and testing
- **Documentation**: Inline docs and architecture guides

### Contributing

- **[Development Guide](docs/DEVELOPMENT.md)**: Complete guide for adding features and following best practices
- **[Project Plan](docs/project_plan.md)**: Complete development roadmap and current progress
- **[Testing Strategy](docs/DEVELOPMENT.md#testing-guidelines)**: How to write and run tests effectively

## 📄 License

FreyjaDB is released under the BSD 3-Clause License. See [LICENSE](LICENSE) for details.

---

> *FreyjaDB is named after Freyja, the Norse goddess associated with wisdom and foresight—qualities essential for a database that must reliably store and retrieve your data.*