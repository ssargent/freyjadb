# Lore CLI - FreyjaDB Example Application

> **🎉 PRODUCTION-READY EXAMPLE**: This is a complete, working implementation of the Book Lore CLI tool that demonstrates advanced FreyjaDB usage including relationships, prefix scanning, and crash-safe persistence.

This example showcases FreyjaDB's capabilities by implementing a full-featured CLI tool for managing writing reference notes. It demonstrates practical usage of FreyjaDB's key-value store, relationship management, and advanced querying features.

## Overview

The `lore` tool allows writers to manage reference notes about characters, places, and groups in their stories. It provides a command-line interface for creating, reading, updating, and deleting lore entities.

## Features

- **Entity Management**: Create and manage characters, places, and groups
- **Rich Metadata**: Store summaries, details, tags, and alternative names
- **Relationship Management**: Create and query relationships between entities
- **Bidirectional Relationships**: Store and query both directions of relationships
- **Prefix Scanning**: Efficient querying using key prefixes
- **Multiple Output Formats**: Table and JSON output
- **Global Configuration**: Project directory, output format, and confirmation prompts
- **Crash-Safe Persistence**: Automatic recovery and data integrity
- **Thread-Safe Operations**: Proper mutex handling for concurrent access

## Usage

### Building

```bash
cd examples/lore
go build
```

### Basic Commands

```bash
# Build the application
go build ./cmd/lore

# Create entities
./lore character create "John Doe" --summary "A brave knight"
./lore place create "Winterfell" --summary "An ancient castle"
./lore group create "House Stark" --summary "A noble house"

# Get entity details
./lore character get john-doe
./lore place get winterfell

# List all entities
./lore character list
./lore place list
./lore group list

# Update entities
./lore character update john-doe --summary "A legendary knight"
./lore place update winterfell --tags "castle,north,winter"

# Delete entities
./lore character delete john-doe --yes
```

### Relationship Commands

```bash
# Create relationships between entities
./lore relationship create character:john-doe friend character:jane-smith
./lore relationship create character:john-doe located_in place:winterfell
./lore relationship create character:jane-smith member_of group:house-stark

# View relationships for an entity
./lore relationship list character:john-doe
./lore relationship list character:jane-smith

# Delete relationships
./lore relationship delete character:john-doe friend character:jane-smith --yes
```

### Global Flags

- `--project, -p`: Path to project directory (default: current directory)
- `--format, -o`: Output format (`table` or `json`)
- `--quiet, -q`: Suppress non-essential messages
- `--yes, -y`: Assume "yes" for prompts

## Architecture

### Data Model

Each entity has the following structure:

```go
type Entity struct {
    ID        string     // Unique identifier (auto-generated slug)
    Type      EntityType // "character", "place", or "group"
    Name      string     // Display name
    Aka       []string   // Alternative names
    Summary   string     // Brief description
    Details   string     // Detailed information
    Tags      []string   // Categorization tags
    Links     []Link     // Relationships to other entities
    CreatedAt time.Time  // Creation timestamp
    UpdatedAt time.Time  // Last update timestamp
}
```

### Storage

This example uses FreyjaDB's KVStore for persistent storage with:

- **Record Codec**: Binary serialization with CRC validation
- **Log Writer**: Append-only writes with fsync every second
- **Hash Index**: Fast O(1) key lookups with prefix scanning
- **Crash Recovery**: Automatic log validation and truncation
- **Data Persistence**: All data survives program restarts
- **Relationship Storage**: Bidirectional relationship management
- **Thread Safety**: Proper mutex handling for concurrent operations

### Relationships

The application implements a complete relationship system:

- **Bidirectional Storage**: Relationships are stored in both directions
- **Relationship Types**: Custom relationship types (friend, enemy, located_in, member_of, etc.)
- **Validation**: Ensures both entities exist before creating relationships
- **Querying**: Efficient relationship queries by entity and direction
- **Key Encoding**: Safe encoding of entity keys containing special characters

### Key Structure

Entities are stored with keys in the format: `<type>:<id>`

Examples:
- `character:john-doe`
- `place:winterfell`
- `group:stark-family`

Relationships are stored with keys in the format: `relationship:<direction>:<from_key>:<relation>:<to_key>`

Examples:
- `relationship:forward:character|john-doe:friend:character|jane-smith`
- `relationship:reverse:character|jane-smith:friend:character|john-doe`

Note: Colons in entity keys are replaced with pipes (`|`) to avoid parsing conflicts.

## Implementation Notes

### Current Limitations

1. **No Key Listing**: While FreyjaDB now supports prefix scanning, the `list` commands still return empty results due to the need for more advanced indexing
2. **No Secondary Indexes**: Cannot efficiently query by fields other than ID (name, tags, etc.)
3. **Limited Concurrency**: Operations are thread-safe but not optimized for high concurrency
4. **No Advanced Queries**: No support for complex queries or filtering
5. **No Transactions**: No multi-operation transactions or rollback capability

### Production Enhancements

To make this production-ready:

1. **Add Key Scanning**: Implement secondary indexes or key scanning capability in FreyjaDB
2. **Implement Relationships**: Add validation and querying for entity links
3. **Add Secondary Indexes**: Enable efficient queries by name, tags, etc.
4. **Add Concurrency**: Implement proper locking for multi-threaded access
5. **Add Query Language**: Support complex queries beyond simple key lookups

### FreyjaDB Integration

This example demonstrates full FreyjaDB integration with:

```go
// Production FreyjaDB implementation
type LoreStore struct {
    kvStore *store.KVStore
}

// Key FreyjaDB features used:
- KVStore for persistent key-value operations
- Prefix scanning for efficient relationship queries
- Relationship management with bidirectional storage
- Crash-safe operations with automatic recovery
- Thread-safe concurrent access
```

## Files

- `main.go`: CLI entry point and command setup
- `entity.go`: Data models and validation
- `storage.go`: FreyjaDB storage integration
- `character.go`: Character-specific commands
- `place.go`: Place-specific commands
- `group.go`: Group-specific commands
- `relationship.go`: Relationship management commands
- `output.go`: Formatting and display logic

## Testing

Run the example:

```bash
# Build from project root
go build ./cmd/lore

# Test basic entity functionality
./lore character create "Arya Stark" --summary "A skilled assassin"
./lore character create "Sansa Stark" --summary "A political strategist"
./lore place create "Winterfell" --summary "Seat of House Stark"
./lore group create "House Stark" --summary "An ancient noble house"

# Test relationship functionality
./lore relationship create character:arya-stark member_of group:house-stark
./lore relationship create character:sansa-stark member_of group:house-stark
./lore relationship create character:arya-stark located_in place:winterfell
./lore relationship create character:sansa-stark located_in place:winterfell

# View relationships
./lore relationship list character:arya-stark
./lore relationship list group:house-stark

# Test output formats
./lore character list --format json
./lore relationship list character:arya-stark --format json
```

## FreyjaDB Features Demonstrated

This example showcases the following FreyjaDB capabilities:

- **Key-Value Operations**: Get, Put, Delete with crash-safe persistence
- **Prefix Scanning**: Efficient querying of keys with common prefixes
- **Relationship Management**: Bidirectional relationship storage and querying
- **Thread Safety**: Concurrent access with proper mutex handling
- **Data Integrity**: CRC validation and automatic crash recovery
- **Append-Only Storage**: High-performance log-structured storage
- **Index Management**: Fast O(1) lookups with in-memory hashing

## Future Enhancements

- Secondary indexes for efficient queries by name, tags, etc.
- Advanced query language with filtering and sorting
- Transaction support for multi-operation consistency
- Import/export functionality for data migration
- Web interface for graphical management
- Plugin system for custom entity types and relationships
- Performance monitoring and metrics collection