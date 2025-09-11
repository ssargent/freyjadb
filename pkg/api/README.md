# FreyjaDB REST API Package

This package provides a complete REST API for FreyjaDB, an embeddable key-value store. The API is designed to be reusable across different parts of the FreyjaDB ecosystem.

## Features

- **Complete REST API**: Full CRUD operations for key-value pairs
- **Relationship Management**: Create and query relationships between entities
- **Authentication**: API key-based authentication
- **Swagger Documentation**: Auto-generated OpenAPI documentation
- **Health Monitoring**: Health check and diagnostic endpoints
- **CORS Support**: Cross-origin resource sharing enabled

## Quick Start

```go
package main

import (
    "github.com/ssargent/freyjadb/pkg/api"
    "github.com/ssargent/freyjadb/pkg/store"
)

func main() {
    // Initialize your store
    config := store.KVStoreConfig{DataDir: "./data"}
    kv, err := store.NewKVStore(config)
    if err != nil {
        panic(err)
    }
    _, err = kv.Open()
    if err != nil {
        panic(err)
    }

    // Start the API server
    serverConfig := api.ServerConfig{
        Port:   8080,
        APIKey: "your-secret-api-key",
    }

    if err := api.StartServer(kv, serverConfig); err != nil {
        panic(err)
    }
}
```

## API Endpoints

### Key-Value Operations
- `PUT /api/v1/kv/{key}` - Store a key-value pair
- `GET /api/v1/kv/{key}` - Retrieve a value
- `DELETE /api/v1/kv/{key}` - Delete a key
- `GET /api/v1/kv?prefix={prefix}` - List keys with prefix

### Relationships
- `POST /api/v1/relationships` - Create relationship
- `DELETE /api/v1/relationships` - Delete relationship
- `GET /api/v1/relationships?key={key}&direction={dir}&relation={rel}&limit={limit}` - Query relationships

### Diagnostics
- `GET /api/v1/health` - Health check
- `GET /api/v1/stats` - Store statistics
- `GET /api/v1/explain` - Detailed diagnostics

### Documentation
- `GET /swagger/` - Interactive API documentation

## Authentication

All endpoints require the `X-API-Key` header:

```bash
curl -H "X-API-Key: your-secret-key" http://localhost:8080/api/v1/health
```

## Usage in Other Packages

The API package is designed to be reusable. For example, the `lore` package could use it like this:

```go
package lore

import "github.com/ssargent/freyjadb/pkg/api"

func startLoreAPI(store *store.KVStore) error {
    config := api.ServerConfig{
        Port:   8080,
        APIKey: "lore-api-key",
    }
    return api.StartServer(store, config)
}
```

## Architecture

The package is organized into several files:

- `types.go` - Shared types and structures
- `middleware.go` - Authentication and utility middleware
- `handlers.go` - HTTP request handlers
- `server.go` - Server setup and routing

This separation allows for easy testing and extension of individual components.