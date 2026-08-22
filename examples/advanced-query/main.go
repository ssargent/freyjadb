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

// Field names and fixture values reused across the sample data and queries below.
const (
	fieldAge    = "age"
	fieldCity   = "city"
	cityNewYork = "New York"

	demoTempDir = "/tmp/freyjadb_demo"
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
	ageIndex := indexManager.GetOrCreateIndex(fieldAge)
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

// sampleUsers returns the demo dataset used to populate the KV store and indexes.
func sampleUsers() []User {
	return []User{
		{
			ID:        "user:1",
			Name:      "Alice Johnson",
			Age:       25,
			City:      cityNewYork,
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
			City:      cityNewYork,
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
			City:      cityNewYork,
			Email:     "eve@example.com",
			CreatedAt: time.Now().Add(-120 * time.Hour),
		},
	}
}

// setupKVStore creates the temp data directory and opens a KV store rooted there.
func setupKVStore(tempDir string) (*store.KVStore, error) {
	kvConfig := store.KVStoreConfig{
		DataDir:       tempDir,
		FsyncInterval: 100 * time.Millisecond,
		MaxRecordSize: 4096, // 4KB default
	}

	kvStore, err := store.NewKVStore(kvConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create KV store: %w", err)
	}

	if _, err := kvStore.Open(); err != nil {
		return nil, fmt.Errorf("failed to open KV store: %w", err)
	}

	return kvStore, nil
}

// insertUsers stores each user in the KV store and indexes its age and city fields.
func insertUsers(kvStore *store.KVStore, indexManager *index.IndexManager, users []User) error {
	fmt.Println("\n📝 Inserting sample data...")

	for _, user := range users {
		userJSON, err := json.Marshal(user)
		if err != nil {
			return fmt.Errorf("failed to marshal user: %w", err)
		}

		key := []byte(user.ID)
		if err := kvStore.Put(key, userJSON); err != nil {
			return fmt.Errorf("failed to store user %s: %w", user.ID, err)
		}

		// Automatically index fields for querying
		ageIndex := indexManager.GetOrCreateIndex(fieldAge)
		if err := ageIndex.Insert(float64(user.Age), key); err != nil {
			return fmt.Errorf("failed to index age for %s: %w", user.ID, err)
		}

		cityIndex := indexManager.GetOrCreateIndex(fieldCity)
		if err := cityIndex.Insert(user.City, key); err != nil {
			return fmt.Errorf("failed to index city for %s: %w", user.ID, err)
		}

		fmt.Printf("✅ Stored and indexed user: %s (%s, age %d)\n", user.Name, user.City, user.Age)
	}

	fmt.Printf("\n📊 Successfully stored %d users with automatic indexing\n", len(users))
	return nil
}

// closeIterator closes a query iterator, logging (rather than failing) on error
// since it runs during best-effort demo cleanup.
func closeIterator(it query.QueryIterator, label string) {
	if err := it.Close(); err != nil {
		log.Printf("Warning: failed to close %s iterator: %v", label, err)
	}
}

// queryExactAge runs the "age == 25" demo query. Unlike the other demo queries,
// a record that fails to unmarshal is treated as fatal here to mirror the
// original example's behavior for this specific query.
func queryExactAge(ctx context.Context, engine query.QueryEngine, extractor query.FieldExtractor) error {
	fmt.Println("\n1️⃣ Finding users aged exactly 25:")
	ageQuery := query.FieldQuery{
		Field:    fieldAge,
		Operator: "=",
		Value:    25.0, // JSON numbers are float64
	}

	iterator, err := engine.ExecuteQuery(ctx, "users", ageQuery, extractor)
	if err != nil {
		return fmt.Errorf("age query failed: %w", err)
	}
	defer closeIterator(iterator, "age")

	count := 0
	for iterator.Next() {
		result := iterator.Result()
		count++

		var user User
		if unmarshalErr := json.Unmarshal(result.Value, &user); unmarshalErr != nil {
			return fmt.Errorf("failed to unmarshal user: %w", unmarshalErr)
		}

		fmt.Printf("   👤 %s (%s) - Age: %d\n", user.Name, user.Email, user.Age)
	}
	fmt.Printf("   📊 Found %d users aged 25\n", count)
	return nil
}

// queryAgeRange runs the "age >= 25" demo query.
func queryAgeRange(ctx context.Context, engine query.QueryEngine, extractor query.FieldExtractor) error {
	fmt.Println("\n2️⃣ Finding users aged 25 or older:")
	rangeQuery := query.FieldQuery{
		Field:    fieldAge,
		Operator: ">=",
		Value:    25.0,
	}

	rangeIterator, err := engine.ExecuteQuery(ctx, "users", rangeQuery, extractor)
	if err != nil {
		return fmt.Errorf("range query failed: %w", err)
	}
	defer closeIterator(rangeIterator, "range")

	rangeCount := 0
	for rangeIterator.Next() {
		result := rangeIterator.Result()
		rangeCount++

		var user User
		if unmarshalErr := json.Unmarshal(result.Value, &user); unmarshalErr != nil {
			log.Printf("Warning: failed to unmarshal user: %v", unmarshalErr)
			continue
		}
		fmt.Printf("   👤 %s - Age: %d\n", user.Name, user.Age)
	}
	fmt.Printf("   📊 Found %d users aged 25+\n", rangeCount)
	return nil
}

// queryCity runs the "city == New York" demo query.
func queryCity(ctx context.Context, engine query.QueryEngine, extractor query.FieldExtractor) error {
	fmt.Println("\n3️⃣ Finding users in New York:")
	cityQuery := query.FieldQuery{
		Field:    fieldCity,
		Operator: "=",
		Value:    cityNewYork,
	}

	cityIterator, err := engine.ExecuteQuery(ctx, "users", cityQuery, extractor)
	if err != nil {
		return fmt.Errorf("city query failed: %w", err)
	}
	defer closeIterator(cityIterator, "city")

	cityCount := 0
	for cityIterator.Next() {
		result := cityIterator.Result()
		cityCount++

		var user User
		if unmarshalErr := json.Unmarshal(result.Value, &user); unmarshalErr != nil {
			log.Printf("Warning: failed to unmarshal user: %v", unmarshalErr)
			continue
		}
		fmt.Printf("   🏙️  %s (%s) - %s\n", user.Name, user.Email, user.City)
	}
	fmt.Printf("   📊 Found %d users in New York\n", cityCount)
	return nil
}

// queryAgeBetween runs the "25 <= age <= 35" demo range query.
func queryAgeBetween(ctx context.Context, engine query.QueryEngine, extractor query.FieldExtractor) error {
	fmt.Println("\n4️⃣ Finding users aged between 25 and 35:")
	startQuery := query.FieldQuery{
		Field:    fieldAge,
		Operator: ">=",
		Value:    25.0,
	}
	endQuery := query.FieldQuery{
		Field:    fieldAge,
		Operator: "<=",
		Value:    35.0,
	}

	betweenIterator, err := engine.ExecuteRangeQuery(ctx, "users", startQuery, endQuery, extractor)
	if err != nil {
		return fmt.Errorf("between query failed: %w", err)
	}
	defer closeIterator(betweenIterator, "between")

	betweenCount := 0
	for betweenIterator.Next() {
		result := betweenIterator.Result()
		betweenCount++

		var user User
		if unmarshalErr := json.Unmarshal(result.Value, &user); unmarshalErr != nil {
			log.Printf("Warning: failed to unmarshal user: %v", unmarshalErr)
			continue
		}
		fmt.Printf("   📅 %s - Age: %d\n", user.Name, user.Age)
	}
	fmt.Printf("   📊 Found %d users aged 25-35\n", betweenCount)
	return nil
}

// runQueries executes the full set of demo queries in sequence.
func runQueries(ctx context.Context, engine query.QueryEngine, extractor query.FieldExtractor) error {
	fmt.Println("\n🔍 Executing queries...")

	if err := queryExactAge(ctx, engine, extractor); err != nil {
		return err
	}
	if err := queryAgeRange(ctx, engine, extractor); err != nil {
		return err
	}
	if err := queryCity(ctx, engine, extractor); err != nil {
		return err
	}
	if err := queryAgeBetween(ctx, engine, extractor); err != nil {
		return err
	}

	return nil
}

// printStats prints summary statistics about the KV store's contents.
func printStats(kvStore *store.KVStore) {
	fmt.Println("\n📈 System Statistics:")
	stats := kvStore.Stats()
	fmt.Printf("   📊 Total keys: %d\n", stats.Keys)
	fmt.Printf("   💾 Data size: %d bytes\n", stats.DataSize)
}

// run wires up the full demo: it sets up the KV store and indexes, inserts
// sample data, executes the demo queries, and prints statistics. All cleanup
// is handled via defer so it runs regardless of where an error occurs; the
// caller (main) is responsible for deciding whether/how to exit on error.
func run() error {
	fmt.Println("🚀 Setting up FreyjaDB with Advanced Query Support")

	// Create temporary directory for demo
	_ = os.RemoveAll(demoTempDir) // Clean up previous runs, ignore error
	if err := os.MkdirAll(demoTempDir, 0750); err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(demoTempDir); err != nil {
			log.Printf("Warning: failed to clean up temp dir: %v", err)
		} else {
			fmt.Println("\n🧹 Demo completed and cleaned up!")
		}
	}()

	kvStore, err := setupKVStore(demoTempDir)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := kvStore.Close(); closeErr != nil {
			log.Printf("Warning: failed to close KV store: %v", closeErr)
		}
	}()

	// Initialize Index Manager
	indexManager := index.NewIndexManager(4)

	// Create Query Engine with both KV store and index manager
	engine := query.NewSimpleQueryEngine(indexManager, kvStore)
	extractor := &query.JSONFieldExtractor{}

	fmt.Println("✅ System initialized successfully")

	users := sampleUsers()
	if err := insertUsers(kvStore, indexManager, users); err != nil {
		return err
	}

	ctx := context.Background()
	if err := runQueries(ctx, engine, extractor); err != nil {
		return err
	}

	printStats(kvStore)

	fmt.Println("\n🎉 FreyjaDB Advanced Query Demo Complete!")
	fmt.Println("   ✅ B+tree indexes working")
	fmt.Println("   ✅ KV store integration working")
	fmt.Println("   ✅ Automatic indexing working")
	fmt.Println("   ✅ Complex queries working")
	fmt.Println("   ✅ Full record retrieval working")

	return nil
}

func main() {
	// Debug the indexing first
	debugIndexing()

	if err := run(); err != nil {
		log.Fatal(err)
	}
}
