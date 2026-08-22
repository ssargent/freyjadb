# FreyjaDB Comprehensive Developer Guide

This guide provides a complete overview of FreyjaDB for developers, covering everything from initial setup to advanced features. FreyjaDB is a small, embeddable key-value database built with Go that offers high performance, crash safety, and advanced querying capabilities.

## Table of Contents

1. [Introduction](#introduction)
2. [Installation and Setup](#installation-and-setup)
3. [Basic Operations](#basic-operations)
4. [Querying Data](#querying-data)
5. [Indexing](#indexing)
6. [Advanced Features](#advanced-features)
7. [Embedding in Applications](#embedding-in-applications)
8. [Security and Authentication](#security-and-authentication)
9. [Performance Tuning](#performance-tuning)
10. [Troubleshooting](#troubleshooting)

## Introduction

FreyjaDB is a log-structured key-value database that provides:

- **High Performance**: Optimized for SSDs with fast lookups and writes
- **Crash Safety**: Automatic recovery with data integrity checks
- **Concurrency**: Multiple readers, single writer architecture
- **Advanced Querying**: Field-based queries with secondary indexes
- **Security**: Encrypted system store for API keys and configuration
- **Flexibility**: Can be used via HTTP REST API or embedded directly

The database is composed of several key components:
- **Log-Structured Storage**: Append-only writes for performance
- **Hash Index**: In-memory index for O(1) key lookups
- **B+Tree Indexes**: Secondary indexes for field-based queries
- **System Store**: Encrypted storage for API keys and configuration

## Installation and Setup

### Prerequisites

- Go 1.21 or higher
- Git

### Building from Source

```bash
# Clone the repository
git clone https://github.com/ssargent/freyjadb.git
cd freyjadb

# Build the project
make build
# or
just build
```

### Quick Start with CLI

The easiest way to get started is using the `up` command:

```bash
# One-command setup and server start
freyja up

# With custom settings
freyja up --data-dir="./mydata" --port=9000

# Print generated API keys
freyja up --print-keys
```

### Manual Setup

For more control, you can initialize and start the server manually:

```bash
# Initialize the system store
freyja init --system-key="your-secure-system-key-here" --data-dir="./data"

# Start the server
freyja serve \
  --api-key="your-user-api-key" \
  --system-key="your-secure-system-key-here" \
  --data-dir="./data" \
  --enable-encryption \
  --port=8080
```

### Systemd Service (Production)

For production deployments on Linux systems:

```bash
# Install as systemd service
sudo freyja service install

# Service will start automatically on boot
sudo systemctl status freyja.service

# View logs
sudo journalctl -u freyja.service -f
```

## Basic Operations

### Using the HTTP REST API

Once the server is running, you can interact with it using HTTP requests:

```bash
# Store data
curl -X PUT http://localhost:8080/api/v1/kv/mykey \
  -H "X-API-Key: your-generated-api-key" \
  -H "Content-Type: application/json" \
  -d '{"name": "John", "age": 30, "city": "New York"}'

# Retrieve data
curl -X GET http://localhost:8080/api/v1/kv/mykey \
  -H "X-API-Key: your-generated-api-key"

# Delete data
curl -X DELETE http://localhost:8080/api/v1/kv/mykey \
  -H "X-API-Key: your-generated-api-key"

# List keys (with optional prefix)
curl -X GET http://localhost:8080/api/v1/kv?prefix=user: \
  -H "X-API-Key: your-generated-api-key"
```

### Working with Relationships

FreyjaDB supports relationships between entities:

```bash
# Create a relationship
curl -X POST http://localhost:8080/api/v1/relationships \
  -H "X-API-Key: your-generated-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "from_key": "user:123",
    "to_key": "order:456",
    "relation": "placed"
  }'

# Query relationships
curl -X GET "http://localhost:8080/api/v1/relationships?from_key=user:123" \
  -H "X-API-Key: your-generated-api-key"
```

## Querying Data

FreyjaDB supports field-based queries on JSON-encoded values through its integrated query engine.

### Setting Up Query Capabilities

To use advanced queries, you need to set up the query engine with indexes:

```go
// Initialize components
kvStore, _ := store.NewKVStore(config)
indexManager := index.NewIndexManager(4)
engine := query.NewSimpleQueryEngine(indexManager, kvStore)
extractor := &query.JSONFieldExtractor{}
```

### Equality Queries

Find records with exact field matches:

```go
// Query for users aged exactly 25
ageQuery := query.FieldQuery{
    Field:    "age",
    Operator: "=",
    Value:    25.0,
}

iterator, err := engine.ExecuteQuery(ctx, "users", ageQuery, extractor)
if err != nil {
    log.Fatal(err)
}
defer iterator.Close()

// Process results
for iterator.Next() {
    result := iterator.Result()
    fmt.Printf("Found user: %s\n", string(result.Value))
}
```

### Range Queries

Find records within a range:

```go
// Query for users aged 25 or older
rangeQuery := query.FieldQuery{
    Field:    "age",
    Operator: ">=",
    Value:    25.0,
}

rangeIterator, err := engine.ExecuteQuery(ctx, "users", rangeQuery, extractor)
if err != nil {
    log.Fatal(err)
}
defer rangeIterator.Close()

for rangeIterator.Next() {
    result := rangeIterator.Result()
    fmt.Printf("Found user: %s\n", string(result.Value))
}
```

### Between Queries

Find records between two values:

```go
// Query for users aged between 25 and 35
startQuery := query.FieldQuery{
    Field:    "age",
    Operator: ">=",
    Value:    25.0,
}
endQuery := query.FieldQuery{
    Field:    "age",
    Operator: "<=",
    Value:    35.0,
}

betweenIterator, err := engine.ExecuteRangeQuery(ctx, "users", startQuery, endQuery, extractor)
if err != nil {
    log.Fatal(err)
}
defer betweenIterator.Close()

for betweenIterator.Next() {
    result := betweenIterator.Result()
    fmt.Printf("Found user: %s\n", string(result.Value))
}
```

### Supported Query Operators

- `=` : Equality
- `>` : Greater than
- `<` : Less than
- `>=` : Greater than or equal
- `<=` : Less than or equal

## Indexing

FreyjaDB uses B+Tree-based secondary indexes to enable fast field-based queries.

### Automatic Index Creation

Indexes are created automatically when you query a field:

```go
// First query on "age" field creates the index
index := indexManager.GetOrCreateIndex("age")

// Subsequent queries on "age" reuse the same index
```

### Manual Index Management

You can also manage indexes manually:

```go
// Create a new index
ageIndex := index.NewSecondaryIndex("age", 4)

// Insert a record into the index
err := ageIndex.Insert(25.0, []byte("user:123"))

// Search the index
keys, err := ageIndex.Search(25.0)

// Range search
keys, err := ageIndex.SearchRange(20.0, 30.0)
```

### Index Persistence

Indexes can be saved to and loaded from disk:

```go
// Save index to disk
err := index.Save("./indexes")

// Load index from disk
err := index.Load("./indexes")
```

## Advanced Features

### B+Tree Package

FreyjaDB includes a standalone, thread-safe B+Tree implementation that can be used independently:

```go
// Create a new B+ tree
tree := bptree.NewBPlusTree(4)

// Insert key-value pairs
tree.Insert([]byte("key1"), []byte("value1"))
tree.Insert([]byte("key2"), []byte("value2"))

// Search for values
if value, found := tree.Search([]byte("key1")); found {
    fmt.Printf("Found: %s\n", string(value))
}

// Delete a key
if tree.Delete([]byte("key2")) {
    fmt.Println("Successfully deleted key2")
}

// Persistence support
tree.Save("./my_tree.dat")
loadedTree, _ := bptree.LoadBPlusTree("./my_tree.dat")
```

### Checkpointing for Long-Running Applications

```go
// Start automatic checkpointing every 30 seconds
checkpointFile := "./tree_checkpoint.dat"
tree.StartCheckpoint(checkpointFile, 30)
fmt.Println("Checkpointing started")

// Stop checkpointing when done
tree.StopCheckpoint()
```

### Thread-Safe Concurrent Operations

The B+Tree implementation supports concurrent operations:

```go
var wg sync.WaitGroup

// Start multiple goroutines performing operations
for i := 0; i < 5; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        
        // Each goroutine inserts its own set of keys
        for j := 0; j < 20; j++ {
            key := []byte(fmt.Sprintf("goroutine:%d:key:%d", id, j))
            val := []byte(fmt.Sprintf("value:%d", j))
            tree.Insert(key, val)
        }
    }(i)
}

wg.Wait()
```

## Embedding in Applications

FreyjaDB can be embedded directly into Go applications for maximum performance and minimal latency.

### Basic Embedded Usage

```go
package main

import (
    "fmt"
    "log"
    "github.com/ssargent/freyjadb/pkg/store"
)

func main() {
    // Configure the KV store
    config := store.KVStoreConfig{
        DataDir: "./my_data",
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

### Advanced Embedded Configuration

```go
// With encryption
config := store.KVStoreConfig{
    DataDir: "./encrypted_data",
    SystemKey: "your-secure-system-key-here",  // Enables encryption
}

// With custom settings
config := store.KVStoreConfig{
    DataDir:       "./my_data",
    FsyncInterval: 100 * time.Millisecond,  // How often to fsync
    MaxRecordSize: 4096,                    // Maximum record size in bytes
}
```

### Embedded Query Engine

You can also use the query engine in embedded mode:

```go
// Initialize components
kvStore, _ := store.NewKVStore(config)
indexManager := index.NewIndexManager(4)
engine := query.NewSimpleQueryEngine(indexManager, kvStore)
extractor := &query.JSONFieldExtractor{}

// Store and index data
userJSON := []byte(`{"id": "user:123", "name": "John", "age": 25}`)
kvStore.Put([]byte("user:123"), userJSON)

// Index the age field
ageIndex := indexManager.GetOrCreateIndex("age")
ageIndex.Insert(25.0, []byte("user:123"))

// Query the data
query := query.FieldQuery{
    Field:    "age",
    Operator: "=",
    Value:    25.0,
}
iterator, _ := engine.ExecuteQuery(context.Background(), "users", query, extractor)
```

## Security and Authentication

### System Store Architecture

FreyjaDB separates system data from user data using a dual-store architecture:

1. **User Store**: Regular key-value data accessible via standard APIs
2. **System Store**: Encrypted system data (API keys, configuration) accessible only via special system APIs

### API Key Management

Create and manage API keys through the system API:

```bash
# Create a new API key
curl -X POST http://localhost:8080/system/api-keys \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secure-system-key-here" \
  -d '{
    "id": "user-api-key-1",
    "key": "secure-random-api-key-12345",
    "description": "API key for user application",
    "is_active": true
  }'

# List all API keys
curl -X GET http://localhost:8080/system/api-keys \
  -H "X-API-Key: your-secure-system-key-here"

# Delete an API key
curl -X DELETE http://localhost:8080/system/api-keys/user-api-key-1 \
  -H "X-API-Key: your-secure-system-key-here"
```

### Encryption

Enable encryption for sensitive data:

```bash
# Start server with encryption enabled
freyja serve \
  --api-key="your-user-key" \
  --system-key="your-secure-system-key-here" \
  --data-dir="./data" \
  --enable-encryption \
  --system-encryption-key="your-secure-system-key-here" \
  --port=8080
```

### Security Best Practices

1. **Use strong, random API keys** - Generate cryptographically secure random keys
2. **Rotate keys regularly** - Implement key rotation policies
3. **Set expiration dates** - Use the `expires_at` field for temporary keys
4. **Use different keys for different applications** - Separate keys by environment/service
5. **Encrypt sensitive configuration** - Always enable encryption for production
6. **Use strong encryption keys** - Minimum 32 characters, randomly generated
7. **Secure key storage** - Store encryption keys securely (environment variables, key vaults)
8. **Use HTTPS** - Always use TLS in production
9. **Firewall configuration** - Restrict access to necessary ports only

## Performance Tuning

### Configuration Options

Adjust these settings based on your workload:

```go
config := store.KVStoreConfig{
    DataDir:       "./data",
    FsyncInterval: 100 * time.Millisecond,  // Balance between performance and durability
    MaxRecordSize: 4096,                    // Adjust based on your typical record sizes
}
```

### Benchmarking

Run benchmarks to measure performance:

```bash
# Run benchmarks
make bench
# or
just bench
```

### Profiling

Profile your application to identify bottlenecks:

```bash
# CPU profiling
make bench-cpu
go tool pprof cpu.prof

# Memory profiling
make bench-mem
go tool pprof mem.prof
```

### Memory Management

Minimize allocations in hot paths:

```go
// Reuse buffers
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
```

## Troubleshooting

### Common Issues

#### "System not initialized" Error

Make sure you've initialized the system store:

```bash
# Initialize with auto-generated API key
freyja init --system-key="your-system-key" --data-dir="./data"
```

#### "Invalid API key" Error

Verify the API key exists and is active:

```bash
curl -X GET http://localhost:8080/system/api-keys \
  -H "X-API-Key: your-system-key"
```

#### "Failed to decrypt" Error

Ensure encryption keys match between init and serve commands:

```bash
# Make sure these match
freyja init --system-key="MySecureKey" --data-dir="./data"
freyja serve --system-key="MySecureKey" --data-dir="./data"
```

### Getting Help

1. **Check server logs** for detailed error messages
2. **Review documentation** in the `docs/` directory
3. **Test with provided examples** in the `examples/` directory
4. **Use GitHub issues** for bugs and feature requests

---

This guide covers the core functionality of FreyjaDB. For more detailed information about specific components, refer to the specialized documentation:

- [System Store User Guide](system-store-user-guide.md)
- [System Store Architecture](system_kv_store_architecture.md)
- [Development Guide](DEVELOPMENT.md)
- [Project Plan](project_plan.md)