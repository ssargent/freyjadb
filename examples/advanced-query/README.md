# FreyjaDB Advanced Query Example

This example demonstrates how to use FreyjaDB's integrated B+tree/KV query system for high-performance data retrieval.

## What This Example Shows

- ✅ **Setting up the integrated system** (B+tree + KV store)
- ✅ **Inserting data with automatic indexing**
- ✅ **Executing various types of queries** (equality, range, between)
- ✅ **Retrieving full records** from the KV store
- ✅ **Real-world usage patterns**

## Running the Example

```bash
# From the FreyjaDB root directory
cd examples/advanced-query
go run main.go
```

## Expected Output

```
🚀 Setting up FreyjaDB with Advanced Query Support
✅ System initialized successfully

📝 Inserting sample data...
✅ Stored and indexed user: Alice Johnson (New York, age 25)
✅ Stored and indexed user: Bob Smith (San Francisco, age 30)
✅ Stored and indexed user: Charlie Brown (New York, age 25)
✅ Stored and indexed user: Diana Prince (Chicago, age 35)
✅ Stored and indexed user: Eve Wilson (New York, age 28)

📊 Successfully stored 5 users with automatic indexing

🔍 Executing queries...

1️⃣ Finding users aged exactly 25:
   👤 Alice Johnson (alice@example.com) - Age: 25
   👤 Charlie Brown (charlie@example.com) - Age: 25
   📊 Found 2 users aged 25

2️⃣ Finding users aged 25 or older:
   👤 Alice Johnson - Age: 25
   👤 Bob Smith - Age: 25
   👤 Charlie Brown - Age: 25
   👤 Diana Prince - Age: 35
   👤 Eve Wilson - Age: 28
   📊 Found 5 users aged 25+

3️⃣ Finding users in New York:
   🏙️  Alice Johnson (alice@example.com) - New York
   🏙️  Charlie Brown (charlie@example.com) - New York
   🏙️  Eve Wilson (eve@example.com) - New York
   📊 Found 3 users in New York

4️⃣ Finding users aged between 25 and 35:
   📅 Alice Johnson - Age: 25
   📅 Bob Smith - Age: 30
   📅 Charlie Brown - Age: 25
   📅 Diana Prince - Age: 35
   📅 Eve Wilson - Age: 28
   📊 Found 5 users aged 25-35

📈 System Statistics:
   📊 Total keys: 5
   💾 Data size: 1024 bytes

🧹 Demo completed and cleaned up!

🎉 FreyjaDB Advanced Query Demo Complete!
   ✅ B+tree indexes working
   ✅ KV store integration working
   ✅ Automatic indexing working
   ✅ Complex queries working
   ✅ Full record retrieval working
```

## How It Works

### 1. System Setup

```go
// Initialize components
kvStore, _ := store.NewKVStore(config)
indexManager := index.NewIndexManager(4)
engine := query.NewSimpleQueryEngine(indexManager, kvStore)
```

### 2. Data Insertion with Indexing

```go
// Store record in KV store
kvStore.Put([]byte(user.ID), userJSON)

// Automatically index fields
ageIndex := indexManager.GetOrCreateIndex("age")
ageIndex.Insert(float64(user.Age), []byte(user.ID))

cityIndex := indexManager.GetOrCreateIndex("city")
cityIndex.Insert(user.City, []byte(user.ID))
```

### 3. Query Execution

```go
// Equality query
query := query.FieldQuery{
    Field:    "age",
    Operator: "=",
    Value:    25.0,
}
iterator, _ := engine.ExecuteQuery(ctx, "users", query, extractor)

// Range query
rangeQuery := query.FieldQuery{
    Field:    "age",
    Operator: ">=",
    Value:    25.0,
}

// Between query
startQuery := query.FieldQuery{Field: "age", Operator: ">=", Value: 25.0}
endQuery := query.FieldQuery{Field: "age", Operator: "<=", Value: 35.0}
iterator, _ := engine.ExecuteRangeQuery(ctx, "users", startQuery, endQuery, extractor)
```

### 4. Result Processing

```go
for iterator.Next() {
    result := iterator.Result()
    // result.Key contains the primary key
    // result.Value contains the full JSON record
    var user User
    json.Unmarshal(result.Value, &user)
}
```

## Key Features Demonstrated

### 🔍 **Query Types**
- **Equality**: `age = 25`
- **Range**: `age >= 25`
- **Between**: `age BETWEEN 25 AND 35`
- **String matching**: `city = "New York"`

### 📊 **Automatic Indexing**
- Indexes are created on-demand
- Composite keys ensure uniqueness
- B+tree provides O(log n) lookup performance

### 💾 **Full Record Retrieval**
- B+tree stores primary keys
- KV store provides O(1) record access
- Complete JSON records returned

### 🚀 **Performance Characteristics**
- **Index Lookup**: O(log n) via B+tree
- **Record Fetch**: O(1) via hash index
- **Memory Efficient**: Streaming iterators
- **Thread Safe**: Concurrent operations supported

## Integration Pattern

The example shows the complete integration pattern:

```
User Data (JSON) → KV Store (primary storage)
                    ↓
              Index Manager (B+tree indexes)
                    ↓
              Query Engine (executes queries)
                    ↓
              Results (full records from KV store)
```

This pattern gives you:
- **Fast queries** via B+tree indexes
- **Complete data** via KV store retrieval
- **Automatic indexing** for any field
- **Scalable architecture** for large datasets