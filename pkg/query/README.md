# FreyjaDB Query Package

This package provides field-based query capabilities for FreyjaDB, enabling efficient queries on JSON-encoded record values.

## Features

- **Field-based queries**: Query records by field values (equality and range queries)
- **Secondary indexes**: Automatic index creation and management for queried fields
- **JSON support**: Built-in JSON field extraction
- **Streaming results**: Iterator-based result streaming for memory efficiency
- **Thread-safe**: Concurrent query execution

## Basic Usage

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
    // Create index manager
    indexManager := index.NewIndexManager(4)

    // Create query engine
    engine := query.NewSimpleQueryEngine(indexManager)

    // Create JSON field extractor
    extractor := &query.JSONFieldExtractor{}

    // Execute a field equality query
    q := query.FieldQuery{
        Field:    "age",
        Operator: "=",
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
        fmt.Printf("Found record: key=%s, value=%s\n",
            string(result.Key), string(result.Value))
    }
}
```

## Query Types

### Field Equality Query

```go
query := query.FieldQuery{
    Field:    "status",
    Operator: "=",
    Value:    "active",
}
```

### Field Range Query

```go
query := query.FieldQuery{
    Field:    "age",
    Operator: ">=",
    Value:    18,
}
```

### Range Between Query

```go
startQuery := query.FieldQuery{
    Field:    "age",
    Operator: ">=",
    Value:    18,
}

endQuery := query.FieldQuery{
    Field:    "age",
    Operator: "<=",
    Value:    65,
}

iterator, err := engine.ExecuteRangeQuery(ctx, "users", startQuery, endQuery, extractor)
```

## Supported Operators

- `=` : Equality
- `>` : Greater than
- `<` : Less than
- `>=` : Greater than or equal
- `<=` : Less than or equal

## Field Extractors

### JSON Field Extractor

The `JSONFieldExtractor` extracts fields from JSON-encoded values:

```go
extractor := &query.JSONFieldExtractor{}

// For JSON: {"name":"John","age":25}
value, err := extractor.Extract([]byte(`{"name":"John","age":25}`), "age")
// value = 25.0 (float64)
```

## Architecture

The query system consists of:

1. **SecondaryIndex**: B+Tree-based indexes for field values
2. **IndexManager**: Manages multiple secondary indexes
3. **QueryEngine**: Executes queries using indexes
4. **FieldExtractor**: Extracts field values from record data
5. **QueryIterator**: Streams query results

## Index Management

Indexes are automatically created when fields are queried:

```go
// First query on "age" field creates the index
index := indexManager.GetOrCreateIndex("age")

// Subsequent queries on "age" reuse the same index
```

## Performance Considerations

- Indexes are created on-demand for queried fields
- Results are streamed to minimize memory usage
- B+Tree provides O(log n) query performance
- Concurrent queries are supported through thread-safe indexes

## Future Enhancements

- Complex query expressions (AND/OR)
- Full-text search capabilities
- Custom field extractors
- Query optimization and planning
- Index statistics and monitoring