# FreyjaDB System Store User Guide

This guide explains how to set up and use FreyjaDB's system key-value store for managing API keys and system configuration. The system store provides secure, encrypted storage for sensitive system data separate from user data.

## Table of Contents

- [Overview](#overview)
- [Fresh Installation Setup](#fresh-installation-setup)
- [Creating API Keys](#creating-api-keys)
- [Using API Keys for Authentication](#using-api-keys-for-authentication)
- [Managing System Configuration](#managing-system-configuration)
- [API Reference](#api-reference)
- [Troubleshooting](#troubleshooting)

## Overview

FreyjaDB now supports two separate data stores:

1. **User Store**: Regular key-value data accessible via standard APIs
2. **System Store**: Encrypted system data (API keys, configuration) accessible only via special system APIs

The system store provides:
- AES-256-GCM encryption for sensitive data
- Secure API key management
- System configuration storage
- Complete separation from user data

## Fresh Installation Setup

### Step 1: Initialize the System Store

First, initialize FreyjaDB's system store with encryption:

```bash
# Initialize with encryption enabled
freyja init --system-key="your-secure-system-key-here" --data-dir="./data"

# Or initialize without encryption (not recommended for production)
freyja init --system-key="your-system-key" --data-dir="./data"
```

**Parameters:**
- `--system-key`: A secure key for encrypting system data (minimum 32 characters recommended)
- `--data-dir`: Directory where FreyjaDB will store data (default: "./data")

**What happens during initialization:**
1. Creates the system data directory (`./data/system/`)
2. Initializes the encrypted system key-value store
3. Creates a root system API key for administrative operations
4. Stores default system configuration

**Example output:**
```
Initializing FreyjaDB system...
Data directory: ./data
System key: your-secure-syst...
✅ FreyjaDB system initialization completed successfully!
System API key: your-secure-system-key-here
Data directory: ./data

You can now start the server with:
  freyja serve --api-key=your-user-key --system-key=your-secure-system-key-here --data-dir=./data
```

### Step 2: Start the Server

Start FreyjaDB with system store support:

```bash
# Start with system store encryption
freyja serve \
  --api-key="your-user-api-key" \
  --system-key="your-secure-system-key-here" \
  --data-dir="./data" \
  --enable-encryption \
  --system-encryption-key="your-secure-system-key-here" \
  --port=8080
```

**Parameters:**
- `--api-key`: API key for user data access (can be any string initially)
- `--system-key`: The system key from initialization
- `--data-dir`: Data directory (must match initialization)
- `--enable-encryption`: Enable encryption for system data
- `--system-encryption-key`: Encryption key (same as system-key)
- `--port`: Port to listen on (default: 8080)

## Creating API Keys

### Method 1: Using System API (Recommended)

Once the server is running, create API keys through the system API:

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
```

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "API key created successfully",
    "id": "user-api-key-1"
  }
}
```

### Method 2: Using CLI Tools

You can also create API keys programmatically in your application:

```go
// Example Go code to create an API key
systemService := // Get system service instance
apiKey := api.APIKey{
    ID:          "app-key-001",
    Key:         "generated-secure-key",
    Description: "Application API key",
    IsActive:    true,
}

err := systemService.StoreAPIKey(apiKey)
if err != nil {
    log.Fatal("Failed to store API key:", err)
}
```

### API Key Structure

```json
{
  "id": "unique-key-identifier",
  "key": "the-actual-api-key-value",
  "description": "optional description",
  "created_at": "2025-09-12T12:00:00Z",
  "expires_at": "2025-12-31T23:59:59Z",  // optional
  "is_active": true
}
```

## Using API Keys for Authentication

### Accessing User Data APIs

Once you have created API keys, use them to authenticate with the standard FreyjaDB APIs:

```bash
# Store data using your API key
curl -X PUT http://localhost:8080/api/v1/kv/mykey \
  -H "Content-Type: application/json" \
  -H "X-API-Key: secure-random-api-key-12345" \
  -d '{"name": "example", "value": 42}'

# Retrieve data
curl -X GET http://localhost:8080/api/v1/kv/mykey \
  -H "X-API-Key: secure-random-api-key-12345"

# List keys
curl -X GET http://localhost:8080/api/v1/kv \
  -H "X-API-Key: secure-random-api-key-12345"
```

### Managing API Keys

```bash
# List all API keys
curl -X GET http://localhost:8080/system/api-keys \
  -H "X-API-Key: your-secure-system-key-here"

# Get specific API key details
curl -X GET http://localhost:8080/system/api-keys/user-api-key-1 \
  -H "X-API-Key: your-secure-system-key-here"

# Delete an API key
curl -X DELETE http://localhost:8080/system/api-keys/user-api-key-1 \
  -H "X-API-Key: your-secure-system-key-here"
```

## Managing System Configuration

### Setting System Configuration

```bash
# Set system configuration
curl -X PUT http://localhost:8080/system/config/max_connections \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secure-system-key-here" \
  -d '100'

# Set complex configuration
curl -X PUT http://localhost:8080/system/config/database \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secure-system-key-here" \
  -d '{
    "host": "localhost",
    "port": 5432,
    "max_connections": 100,
    "timeout": "30s"
  }'
```

### Retrieving System Configuration

```bash
# Get system configuration
curl -X GET http://localhost:8080/system/config/max_connections \
  -H "X-API-Key: your-secure-system-key-here"
```

## Complete Setup Example

Here's a complete example from fresh installation to using APIs:

### 1. Initialize System

```bash
freyja init --system-key="MySuperSecureSystemKey12345!" --data-dir="./freyja-data"
```

### 2. Start Server

```bash
freyja serve \
  --api-key="initial-user-key" \
  --system-key="MySuperSecureSystemKey12345!" \
  --data-dir="./freyja-data" \
  --enable-encryption \
  --system-encryption-key="MySuperSecureSystemKey12345!" \
  --port=8080
```

### 3. Create API Key

```bash
curl -X POST http://localhost:8080/system/api-keys \
  -H "Content-Type: application/json" \
  -H "X-API-Key: MySuperSecureSystemKey12345!" \
  -d '{
    "id": "production-api-key",
    "key": "prod-api-key-secure-random-123",
    "description": "Production API key for user data",
    "is_active": true
  }'
```

### 4. Use API Key for Data Operations

```bash
# Store user data
curl -X PUT http://localhost:8080/api/v1/kv/user:123 \
  -H "Content-Type: application/json" \
  -H "X-API-Key: prod-api-key-secure-random-123" \
  -d '{"name": "John Doe", "email": "john@example.com"}'

# Retrieve user data
curl -X GET http://localhost:8080/api/v1/kv/user:123 \
  -H "X-API-Key: prod-api-key-secure-random-123"
```

## API Reference

### System API Endpoints

All system endpoints require the system API key in the `X-API-Key` header.

#### API Key Management

- `POST /system/api-keys` - Create new API key
- `GET /system/api-keys` - List all API keys
- `GET /system/api-keys/{id}` - Get specific API key
- `DELETE /system/api-keys/{id}` - Delete API key

#### System Configuration

- `PUT /system/config/{key}` - Set configuration value
- `GET /system/config/{key}` - Get configuration value

### User Data Endpoints

These endpoints require a user API key created through the system APIs.

- `PUT /api/v1/kv/{key}` - Store key-value pair
- `GET /api/v1/kv/{key}` - Retrieve value by key
- `DELETE /api/v1/kv/{key}` - Delete key-value pair
- `GET /api/v1/kv` - List keys (with optional prefix)
- `POST /api/v1/relationships` - Create relationship
- `GET /api/v1/relationships` - Query relationships
- `DELETE /api/v1/relationships` - Delete relationship

## Security Best Practices

### API Key Management

1. **Use strong, random API keys** - Generate cryptographically secure random keys
2. **Rotate keys regularly** - Implement key rotation policies
3. **Set expiration dates** - Use the `expires_at` field for temporary keys
4. **Monitor key usage** - Log API key usage for security auditing
5. **Use different keys for different applications** - Separate keys by environment/service

### System Configuration

1. **Encrypt sensitive configuration** - Always enable encryption for production
2. **Use strong encryption keys** - Minimum 32 characters, randomly generated
3. **Secure key storage** - Store encryption keys securely (environment variables, key vaults)
4. **Regular backups** - Backup system store data regularly

### Network Security

1. **Use HTTPS** - Always use TLS in production
2. **Firewall configuration** - Restrict access to necessary ports only
3. **API key transmission** - Never transmit API keys in URLs or logs

## Troubleshooting

### Common Issues

#### "System not initialized" Error

**Problem:** Trying to use system APIs before initialization.

**Solution:**
```bash
freyja init --system-key="your-system-key" --data-dir="./data"
```

#### "Invalid API key" Error

**Problem:** API key authentication failing.

**Solutions:**
1. Verify the API key exists:
   ```bash
   curl -X GET http://localhost:8080/system/api-keys \
     -H "X-API-Key: your-system-key"
   ```

2. Check if the key is active and not expired
3. Ensure correct `X-API-Key` header format

#### "Failed to decrypt" Error

**Problem:** System data cannot be decrypted.

**Solutions:**
1. Verify encryption key matches between init and serve commands
2. Check that `--enable-encryption` and `--system-encryption-key` are set correctly
3. Ensure the system store wasn't corrupted

#### Server Won't Start

**Problem:** Server fails to start with system store errors.

**Solutions:**
1. Check data directory permissions
2. Verify system key is correct
3. Check for file system issues
4. Review server logs for detailed error messages

### Getting Help

For additional support:
1. Check the [System KV Store Architecture](system_kv_store_architecture.md) document
2. Review server logs for detailed error messages
3. Test with the provided examples in this guide

## Migration from Legacy Setup

If you have an existing FreyjaDB installation without the system store:

1. **Backup your data** - Always backup before migration
2. **Initialize system store** - Run `freyja init` with your chosen system key
3. **Create API keys** - Use system APIs to create user API keys
4. **Update applications** - Update your applications to use the new API keys
5. **Test thoroughly** - Verify all functionality works with new authentication

The system store is designed to be backward compatible, so existing user data remains accessible.