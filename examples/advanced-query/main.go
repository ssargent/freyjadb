// Example usage of the integrated B+tree/KV query system
// This file demonstrates how to use FreyjaDB's advanced query capabilities
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ssargent/freyjadb/pkg/index"
	"github.com/ssargent/freyjadb/pkg/query"
	"github.com/ssargent/freyjadb/pkg/store"
)

// User represents a user record
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Age       int       `json:"age"`
	City      string    `json:"city"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// Debug function to understand B+tree indexing
func debugIndexing() {
	fmt.Println("🔍 Debugging B+tree indexing...")

	// Create index manager
	indexManager := index.NewIndexManager(4)

	// Test data
	testKey := []byte("user:1")
	testValue := 25.0

	// Create and populate index
	ageIndex := indexManager.GetOrCreateIndex("age")
	err := ageIndex.Insert(testValue, testKey)
	if err != nil {
		log.Fatalf("Failed to insert: %v", err)
	}

	fmt.Printf("✅ Inserted key=%s, value=%v\n", string(testKey), testValue)

	// Try to search
	results, err := ageIndex.Search(testValue)
	if err != nil {
		log.Fatalf("Search failed: %v", err)
	}

	fmt.Printf("🔍 Search results: %d keys found\n", len(results))
	for i, key := range results {
		fmt.Printf("   Result %d: %s\n", i, string(key))
	}

	// Try range search
	rangeResults, err := ageIndex.SearchRange(testValue, nil)
	if err != nil {
		log.Fatalf("Range search failed: %v", err)
	}

	fmt.Printf("🔍 Range search results: %d keys found\n", len(rangeResults))
	for i, key := range rangeResults {
		fmt.Printf("   Range result %d: %s\n", i, string(key))
	}
}

func main() {
	// Debug the indexing first
	debugIndexing()

	// 1. Set up the integrated system
	fmt.Println("🚀 Setting up FreyjaDB with Advanced Query Support")

	// Create temporary directory for demo
	tempDir := "/tmp/freyjadb_demo"
	os.RemoveAll(tempDir) // Clean up previous runs
	if err := os.MkdirAll(tempDir, 0750); err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}

	// Initialize KV Store
	kvConfig := store.KVStoreConfig{
		DataDir:       tempDir,
		FsyncInterval: 100 * time.Millisecond,
		MaxRecordSize: 4096, // 4KB default
	}
	kvStore, err := store.NewKVStore(kvConfig)
	if err != nil {
		log.Fatalf("Failed to create KV store: %v", err)
	}
	defer func() {
		if err := kvStore.Close(); err != nil {
			log.Printf("Warning: failed to close KV store: %v", err)
		}
	}()

	// Open the store
	_, err = kvStore.Open()
	if err != nil {
		log.Fatalf("Failed to open KV store: %v", err)
	}

	// Initialize Index Manager
	indexManager := index.NewIndexManager(4)

	// Create Query Engine with both KV store and index manager
	engine := query.NewSimpleQueryEngine(indexManager, kvStore)
	extractor := &query.JSONFieldExtractor{}

	fmt.Println("✅ System initialized successfully")

	// 2. Insert sample data with automatic indexing
	fmt.Println("\n📝 Inserting sample data...")

	users := []User{
		{
			ID:        "user:1",
			Name:      "Alice Johnson",
			Age:       25,
			City:      "New York",
			Email:     "alice@example.com",
			CreatedAt: time.Now().Add(-24 * time.Hour),
		},
		{
			ID:        "user:2",
			Name:      "Bob Smith",
			Age:       30,
			City:      "San Francisco",
			Email:     "bob@example.com",
			CreatedAt: time.Now().Add(-48 * time.Hour),
		},
		{
			ID:        "user:3",
			Name:      "Charlie Brown",
			Age:       25,
			City:      "New York",
			Email:     "charlie@example.com",
			CreatedAt: time.Now().Add(-72 * time.Hour),
		},
		{
			ID:        "user:4",
			Name:      "Diana Prince",
			Age:       35,
			City:      "Chicago",
			Email:     "diana@example.com",
			CreatedAt: time.Now().Add(-96 * time.Hour),
		},
		{
			ID:        "user:5",
			Name:      "Eve Wilson",
			Age:       28,
			City:      "New York",
			Email:     "eve@example.com",
			CreatedAt: time.Now().Add(-120 * time.Hour),
		},
	}

	// Insert each user and automatically create indexes
	for _, user := range users {
		// Convert to JSON
		userJSON, err := json.Marshal(user)
		if err != nil {
			log.Fatalf("Failed to marshal user: %v", err)
		}

		// Store in KV store
		key := []byte(user.ID)
		err = kvStore.Put(key, userJSON)
		if err != nil {
			log.Fatalf("Failed to store user %s: %v", user.ID, err)
		}

		// Automatically index fields for querying
		// Age index
		ageIndex := indexManager.GetOrCreateIndex("age")
		err = ageIndex.Insert(float64(user.Age), key)
		if err != nil {
			log.Fatalf("Failed to index age for %s: %v", user.ID, err)
		}

		// City index
		cityIndex := indexManager.GetOrCreateIndex("city")
		err = cityIndex.Insert(user.City, key)
		if err != nil {
			log.Fatalf("Failed to index city for %s: %v", user.ID, err)
		}

		fmt.Printf("✅ Stored and indexed user: %s (%s, age %d)\n", user.Name, user.City, user.Age)
	}

	fmt.Printf("\n📊 Successfully stored %d users with automatic indexing\n", len(users))

	// 3. Execute various queries
	fmt.Println("\n🔍 Executing queries...")

	ctx := context.Background()

	// Query 1: Find users aged exactly 25
	fmt.Println("\n1️⃣ Finding users aged exactly 25:")
	ageQuery := query.FieldQuery{
		Field:    "age",
		Operator: "=",
		Value:    25.0, // JSON numbers are float64
	}

	iterator, err := engine.ExecuteQuery(ctx, "users", ageQuery, extractor)
	if err != nil {
		log.Fatalf("Age query failed: %v", err)
	}

	count := 0
	for iterator.Next() {
		result := iterator.Result()
		count++

		// Parse the full user record
		var user User
		err := json.Unmarshal(result.Value, &user)
		if err != nil {
			log.Fatalf("Failed to unmarshal user: %v", err)
		}

		fmt.Printf("   👤 %s (%s) - Age: %d\n", user.Name, user.Email, user.Age)
	}
	if err := iterator.Close(); err != nil {
		log.Printf("Warning: failed to close iterator: %v", err)
	}
	fmt.Printf("   📊 Found %d users aged 25\n", count)

	// Query 2: Find users aged 25 or older
	fmt.Println("\n2️⃣ Finding users aged 25 or older:")
	rangeQuery := query.FieldQuery{
		Field:    "age",
		Operator: ">=",
		Value:    25.0,
	}

	rangeIterator, err := engine.ExecuteQuery(ctx, "users", rangeQuery, extractor)
	if err != nil {
		log.Fatalf("Range query failed: %v", err)
	}

	rangeCount := 0
	for rangeIterator.Next() {
		result := rangeIterator.Result()
		rangeCount++

		var user User
		if err := json.Unmarshal(result.Value, &user); err != nil {
			log.Printf("Warning: failed to unmarshal user: %v", err)
			continue
		}
		fmt.Printf("   👤 %s - Age: %d\n", user.Name, user.Age)
	}
	if err := rangeIterator.Close(); err != nil {
		log.Printf("Warning: failed to close range iterator: %v", err)
	}
	fmt.Printf("   📊 Found %d users aged 25+\n", rangeCount)

	// Query 3: Find users in New York
	fmt.Println("\n3️⃣ Finding users in New York:")
	cityQuery := query.FieldQuery{
		Field:    "city",
		Operator: "=",
		Value:    "New York",
	}

	cityIterator, err := engine.ExecuteQuery(ctx, "users", cityQuery, extractor)
	if err != nil {
		log.Fatalf("City query failed: %v", err)
	}

	cityCount := 0
	for cityIterator.Next() {
		result := cityIterator.Result()
		cityCount++

		var user User
		if err := json.Unmarshal(result.Value, &user); err != nil {
			log.Printf("Warning: failed to unmarshal user: %v", err)
			continue
		}
		fmt.Printf("   🏙️  %s (%s) - %s\n", user.Name, user.Email, user.City)
	}
	if err := cityIterator.Close(); err != nil {
		log.Printf("Warning: failed to close city iterator: %v", err)
	}
	fmt.Printf("   📊 Found %d users in New York\n", cityCount)

	// Query 4: Range query between ages
	fmt.Println("\n4️⃣ Finding users aged between 25 and 35:")
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
		log.Fatalf("Between query failed: %v", err)
	}

	betweenCount := 0
	for betweenIterator.Next() {
		result := betweenIterator.Result()
		betweenCount++

		var user User
		if err := json.Unmarshal(result.Value, &user); err != nil {
			log.Printf("Warning: failed to unmarshal user: %v", err)
			continue
		}
		fmt.Printf("   📅 %s - Age: %d\n", user.Name, user.Age)
	}
	if err := betweenIterator.Close(); err != nil {
		log.Printf("Warning: failed to close between iterator: %v", err)
	}
	fmt.Printf("   📊 Found %d users aged 25-35\n", betweenCount)

	// 4. Demonstrate statistics
	fmt.Println("\n📈 System Statistics:")
	stats := kvStore.Stats()
	fmt.Printf("   📊 Total keys: %d\n", stats.Keys)
	fmt.Printf("   💾 Data size: %d bytes\n", stats.DataSize)

	// Clean up
	if err := os.RemoveAll(tempDir); err != nil {
		log.Printf("Warning: failed to clean up temp dir: %v", err)
	}
	fmt.Println("\n🧹 Demo completed and cleaned up!")

	fmt.Println("\n🎉 FreyjaDB Advanced Query Demo Complete!")
	fmt.Println("   ✅ B+tree indexes working")
	fmt.Println("   ✅ KV store integration working")
	fmt.Println("   ✅ Automatic indexing working")
	fmt.Println("   ✅ Complex queries working")
	fmt.Println("   ✅ Full record retrieval working")
}
