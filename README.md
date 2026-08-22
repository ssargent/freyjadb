# FreyjaDB

A **small, embeddable key-value database** built with ❤️ in Go. FreyjaDB provides a simple, fast, and reliable way to store and retrieve data with an HTTP REST API.

## ✨ What FreyjaDB Offers

- **🔑 Key-Value API**: Simple HTTP REST API for storing and retrieving data
- **🔐 System Store**: Secure, encrypted storage for API keys and system configuration
- **🌳 B+ Tree Package**: High-performance, thread-safe B+ tree implementation for advanced use cases
- **🛡️ Crash Safety**: Automatic recovery with data integrity checks
- **⚡ High Performance**: Optimized for SSDs with fast lookups and writes

## 🛠️ Development

### Using Just (Recommended)

This project now includes a `justfile` for streamlined development workflows. Install [Just](https://github.com/casey/just) and use:

```bash
# Show all available commands
just help

# Build the project
just build

# Run tests
just test

# Full development cycle
just dev
```

All the traditional Makefile commands are available through Just with improved reliability and cross-platform support.

### Traditional Make (Legacy)

The project also maintains a Makefile for compatibility:

### Option 1: One-Command Setup (Recommended)

```bash
# Bootstrap and start FreyjaDB in one command
freyja up

# Or with custom settings
freyja up --data-dir="./mydata" --port=9000
```

**What's included:**
- ✅ Automatic configuration generation
- ✅ Secure key generation and storage
- ✅ System store initialization
- ✅ Server startup on http://localhost:8080

**Get your API key:**
```bash
# The generated client API key is displayed on first run
freyja up --print-keys
```

### Option 2: Manual Setup (Advanced)

#### 1. Initialize the System Store

```bash
# Initialize with encryption and auto-generated API key
freyja init --system-key="your-secure-system-key-here" --data-dir="./data"

# Or specify a custom system API key
freyja init --system-key="your-secure-system-key-here" --system-api-key="your-custom-api-key" --data-dir="./data"
```

#### 2. Start the Server

```bash
# Start with system store support
freyja serve \
  --api-key="your-user-api-key" \
  --system-key="your-secure-system-key-here" \
  --data-dir="./data" \
  --enable-encryption \
  --port=8080
```

### 3. Use the API

```bash
# Store data
curl -X PUT http://localhost:8080/api/v1/kv/mykey \
  -H "X-API-Key: your-generated-api-key" \
  -d "Hello, FreyjaDB!"

# Retrieve data
curl -X GET http://localhost:8080/api/v1/kv/mykey \
  -H "X-API-Key: your-generated-api-key"
```

### Option 3: Systemd Service (Production)

```bash
# Install as systemd service
sudo freyja service install

# Service will start automatically on boot
# Check status
sudo systemctl status freyja.service

# View logs
sudo journalctl -u freyja.service -f
```

## 🖥️ CLI Commands

FreyjaDB provides a modern, ergonomic CLI with both simple and advanced usage patterns:

### Core Commands

- **`freyja up`** - One-command bootstrap and server start (recommended for most users)
- **`freyja service`** - Systemd service management for production deployments
- **`freyja init`** - Manual system store initialization (advanced users)
- **`freyja serve`** - Manual server start (advanced users)

### Command Reference

#### freyja up
```bash
freyja up [options]

Options:
  --data-dir string     Data directory (default "./data")
  --port int           Port to listen on (default 8080)
  --bind string        Address to bind server to (default "127.0.0.1")
  --config string      Path to config file (default OS-specific location)
  --print-keys         Print generated API keys to console
```

#### freyja service
```bash
freyja service <command> [options]

Commands:
  install    Install FreyjaDB as a systemd service
  start      Start the FreyjaDB service
  stop       Stop the FreyjaDB service
  restart    Restart the FreyjaDB service
  status     Show service status
  logs       Show service logs
  uninstall  Remove the systemd service

Install options:
  --data-dir string    Data directory (default "/var/lib/freyjadb")
  --config string      Path to config file
  --user string        User to run service as (default "freyja")
  --port int          Port for service (default 8080)
  --start             Start service after installation (default true)
```

### Migration Guide

**From old workflow:**
```bash
freyja init --system-key=... --data-dir=./data
freyja serve --api-key=... --system-key=... --data-dir=./data
```

**To new workflow:**
```bash
freyja up  # That's it!
```

The old commands (`init` and `serve`) are still available for advanced users and backward compatibility.

## 🔌 Usage Options

FreyjaDB can be used in three ways:

### Option 1: CLI with HTTP REST API (Recommended for most use cases)

Use the modern CLI commands to start a server and interact via HTTP endpoints:

```bash
freyja up  # Start server
curl -X PUT http://localhost:8080/api/v1/kv/mykey -H "X-API-Key: $(freyja up --print-keys | grep 'Client API Key')" -d "Hello!"
```

### Option 2: HTTP REST API Only (Advanced)

Start the server manually and use HTTP endpoints for data operations.

### Option 3: Embedded Database (Direct Integration into Go Applications)

For scenarios where you want to embed the key-value store directly into your Go application—bypassing the HTTP server and API key authentication entirely—FreyjaDB provides a lightweight, high-performance embedded mode. This approach is ideal for:

- **Single-process applications**: No network overhead or serialization costs.
- **High-throughput needs**: Direct in-memory and disk access for faster operations.
- **Simplified security**: No API keys required; access is controlled at the application level (e.g., via your app's authentication).
- **Offline or edge computing**: Works without external dependencies on a server process.
- **Custom integrations**: Full control over the store's lifecycle within your application's context.

In embedded mode, you import the `pkg/store` package and manage the KVStore instance directly in your code. The store handles its own persistence to a data directory, with automatic crash recovery and indexing.

#### Basic Embedded Usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/ssargent/freyjadb/pkg/store"
)

func main() {
    // Configure the KV store (optional: add encryption via SystemKey if needed)
    config := store.KVStoreConfig{
        DataDir:    "./my_data",  // Directory for log files and indexes
        // SystemKey: "your-secure-key",  // Optional: Enable encryption
    }

    // Create and open the store
    kvStore, err := store.NewKVStore(config)
    if err != nil {
        log.Fatal("Failed to create store:", err)
    }
    defer kvStore.Close()

    _, err = kvStore.Open()  // Loads existing data or initializes new
    if err != nil {
        log.Fatal("Failed to open store:", err)
    }

    // Basic operations: Put, Get, Delete
    key := []byte("user:123")
    value := []byte(`{"name": "John", "email": "john@example.com"}`)

    if err := kvStore.Put(key, value); err != nil {
        log.Fatal("Failed to store data:", err)
    }

    retrieved, err := kvStore.Get(key)
    if err != nil {
        log.Fatal("Failed to retrieve data:", err)
    }
    fmt.Printf("Retrieved: %s\n", string(retrieved))

    if err := kvStore.Delete(key); err != nil {
        log.Fatal("Failed to delete:", err)
    }

    // Prefix-based key listing
    keys, err := kvStore.ListKeys([]byte("user:"))
    if err != nil {
        log.Fatal("Failed to list keys:", err)
    }
    fmt.Printf("Found %d user keys\n", len(keys))
}
```

#### Advanced Embedded Configuration

- **Encryption**: Set `SystemKey` in `KVStoreConfig` to enable transparent encryption for data at rest. This uses the same secure key management as the server mode but without API keys.
  
  ```go
  config := store.KVStoreConfig{
      DataDir:   "./encrypted_data",
      SystemKey: "your-secure-system-key-here",  // Enables encryption
  }
  ```

- **Concurrency**: The store supports multiple readers and a single writer out-of-the-box. For write-heavy workloads, consider serializing writes via a mutex in your application.

- **Lifecycle Management**: Always call `Open()` after creation to load data, and `Close()` to flush changes and release resources. The store is not automatically persisted on every operation—changes are logged and flushed periodically.

- **Error Handling**: Embedded mode provides direct error returns (e.g., `store.ErrKeyNotFound`). Wrap operations in your app's error handling as needed.

#### When to Use Embedded vs. API Mode

- **Embedded**: Best for internal app storage, microservices with direct integration, or when minimizing latency is critical. No setup for servers or API keys.
- **API Mode**: Suited for multi-client access, microservices communication, or when you need HTTP-based interfaces with authentication via API keys.

For more examples, see the [examples/](examples/) directory or the [System Store User Guide](docs/system-store-user-guide.md) for integration tips.
## � Documentation

- **[System Store User Guide](docs/system-store-user-guide.md)**: Complete guide for setup, API key management, and authentication
- **[System Store Architecture](docs/system_kv_store_architecture.md)**: Technical details about the system store implementation
- **[Development Guide](docs/DEVELOPMENT.md)**: Contributing guidelines and development practices

## 🏗️ Architecture

FreyjaDB consists of two main components:

### Key-Value Store
- **Log-structured storage** with in-memory hash index
- **HTTP REST API** with JSON responses
- **Crash recovery** with automatic data validation
- **Concurrent access** (multiple readers, single writer)

### System Store
- **Encrypted storage** for sensitive system data
- **API key management** with secure authentication
- **System configuration** storage
- **Complete separation** from user data for security

## 🌳 B+ Tree Package

FreyjaDB includes a standalone B+ tree implementation that can be used independently:

```go
package main

import (
    "fmt"
    "github.com/ssargent/freyjadb/pkg/bptree"
)

func main() {
    // Create a new B+ tree
    tree := bptree.NewBPlusTree(4)

    // Insert key-value pairs
    tree.Insert([]byte("key1"), []byte("value1"))
    tree.Insert([]byte("key2"), []byte("value2"))

    // Search for values
    if value, found := tree.Search([]byte("key1")); found {
        fmt.Printf("Found: %s\n", string(value))
    }

    // Persistence support
    tree.Save("./my_tree.dat")
    loadedTree, _ := bptree.LoadBPlusTree("./my_tree.dat")
}
```

## ⚠️ Important Notes

**FreyjaDB is a passion project** and is not currently designed or optimized for production workloads. It serves as:

- A learning platform for database internals
- A foundation for understanding distributed systems concepts
- A codebase for experimenting with storage engine designs

For production use cases, consider established databases like Redis, PostgreSQL, or DynamoDB.

## 🛠️ Development

### Prerequisites
- Go 1.21+
- Make or Just (Just recommended)

### Building and Testing

With Just (recommended):
```bash
# Build the project
just build

# Run tests
just test

# Run linter
just lint

# Run all checks
just all
```

With Make (legacy):
```bash
# Build the project
make build

# Run tests
make test

# Run linter
make lint

# Run all checks
make all
```

### Project Structure

```
freyjadb/
├── cmd/freyja/          # CLI commands
├── pkg/
│   ├── api/            # HTTP API and system store
│   ├── bptree/         # B+ tree implementation
│   ├── codec/          # Record encoding/decoding
│   ├── index/          # Indexing components
│   ├── query/          # Query engine
│   └── store/          # Core storage engine
├── docs/               # Documentation
└── examples/           # Usage examples
```

## 🤝 Contributing

FreyjaDB welcomes contributions! Please see our [Development Guide](docs/DEVELOPMENT.md) for:

- Code style guidelines
- Testing practices
- How to add new features
- Pull request process

## 📄 License

FreyjaDB is released under the BSD 3-Clause License. See [LICENSE](LICENSE) for details.

---

> *FreyjaDB is named after Freyja, the Norse goddess associated with wisdom and foresight—qualities essential for a database that must reliably store and retrieve your data.*

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