# FreyjaDB REST API

## Overview

FreyjaDB provides a REST API for storing and retrieving key-value data with automatic content-type handling.

## Content-Type Aware Storage

FreyjaDB now supports automatic content-type detection and handling, making it easier to work with JSON data without manual marshaling/unmarshaling.

### Supported Content Types

- `application/json` - JSON data (automatically validated and formatted)
- `application/octet-stream` - Raw binary data (default)

### API Endpoints

#### PUT /api/v1/kv/{key}

Store a key-value pair with automatic content-type handling.

**Headers:**
- `Content-Type`: `application/json` or `application/octet-stream` (optional, defaults to octet-stream)
- `X-API-Key`: Your API key (required)

**Example - Store JSON:**
```bash
curl -X PUT http://localhost:9200/api/v1/kv/user/123 \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{"name": "John Doe", "age": 30, "email": "john@example.com"}'
```

**Example - Store Raw Bytes:**
```bash
curl -X PUT http://localhost:9200/api/v1/kv/binary/data \
  -H "X-API-Key: your-api-key" \
  --data-binary @file.bin
```

#### GET /api/v1/kv/{key}

Retrieve a value with automatic content-type detection.

**Headers:**
- `X-API-Key`: Your API key (required)

**Response Headers:**
- `Content-Type`: The original content type of the stored data

**Example:**
```bash
curl http://localhost:9200/api/v1/kv/user/123 \
  -H "X-API-Key: your-api-key"
# Returns: {"name": "John Doe", "age": 30, "email": "john@example.com"}
# Content-Type: application/json
```

### Backward Compatibility

Existing data stored without content-type headers continues to work exactly as before. Such data is treated as raw bytes and returned with `Content-Type: application/octet-stream`.

### Implementation Details

- **Metadata Storage**: Content-type information is stored as a 2-byte header prefix
- **JSON Validation**: JSON data is validated on storage and reformatted for consistency
- **Automatic Detection**: Content-type is automatically detected from HTTP headers
- **Minimal Overhead**: Only 2 bytes of metadata per entry

### Error Handling

- **400 Bad Request**: Invalid JSON in request body
- **404 Not Found**: Key does not exist
- **500 Internal Server Error**: Storage or retrieval errors

### Examples

#### JavaScript/Node.js
```javascript
// Store JSON
const response = await fetch('http://localhost:9200/api/v1/kv/user/123', {
  method: 'PUT',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'your-api-key'
  },
  body: JSON.stringify({ name: 'John', age: 30 })
});

// Retrieve JSON
const data = await fetch('http://localhost:9200/api/v1/kv/user/123', {
  headers: { 'X-API-Key': 'your-api-key' }
});
const user = await data.json(); // Automatically parsed
```

#### Python
```python
import requests
import json

# Store JSON
user_data = {'name': 'John', 'age': 30}
response = requests.put(
    'http://localhost:9200/api/v1/kv/user/123',
    json=user_data,
    headers={'X-API-Key': 'your-api-key'}
)

# Retrieve JSON
response = requests.get(
    'http://localhost:9200/api/v1/kv/user/123',
    headers={'X-API-Key': 'your-api-key'}
)
user = response.json()  # Automatically parsed
```

### Migration

No migration is required. Existing applications continue to work unchanged. New applications can opt into JSON handling by setting the `Content-Type: application/json` header.