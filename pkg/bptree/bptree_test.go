//go:build bench
// +build bench

package bptree

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/segmentio/ksuid"
)

func TestNewBPlusTree(t *testing.T) {
	tree := NewBPlusTree(3)
	if tree == nil {
		t.Fatal("Expected non-nil tree")
	}
	if tree.order != 3 {
		t.Fatalf("Expected order 3, got %d", tree.order)
	}
	if tree.height != 1 {
		t.Fatalf("Expected height 1, got %d", tree.height)
	}
}

func TestBPlusTree_InsertAndSearch(t *testing.T) {
	tree := NewBPlusTree(3)

	key1 := []byte("key1")
	val1 := ksuid.New()
	tree.Insert(key1, val1)

	key2 := []byte("key2")
	val2 := ksuid.New()
	tree.Insert(key2, val2)

	// Test search for existing keys
	if v, found := tree.Search(key1); !found || !bytes.Equal(v.Bytes(), val1.Bytes()) {
		t.Fatalf("Expected to find key1 with value %v, got %v", val1, v)
	}

	if v, found := tree.Search(key2); !found || !bytes.Equal(v.Bytes(), val2.Bytes()) {
		t.Fatalf("Expected to find key2 with value %v, got %v", val2, v)
	}

	// Test search for non-existing key
	if _, found := tree.Search([]byte("key3")); found {
		t.Fatal("Expected not to find key3")
	}
}

func TestBPlusTree_SplitLeaf(t *testing.T) {
	tree := NewBPlusTree(3)

	keys := [][]byte{[]byte("key1"), []byte("key2"), []byte("key3"), []byte("key4")}
	values := []ksuid.KSUID{ksuid.New(), ksuid.New(), ksuid.New(), ksuid.New()}

	for i := range keys {
		tree.Insert(keys[i], values[i])
	}

	// Check if the tree height increased due to splits
	if tree.height != 2 {
		t.Fatalf("Expected tree height 2, got %d", tree.height)
	}

	// Check if all keys are present
	for i, key := range keys {
		if v, found := tree.Search(key); !found || !bytes.Equal(v.Bytes(), values[i].Bytes()) {
			t.Fatalf("Expected to find %s with value %v, got %v", key, values[i], v)
		}
	}
}

func TestBPlusTree_SaveLoad(t *testing.T) {
	tree := NewBPlusTree(3)

	// Insert some data
	keys := [][]byte{[]byte("key1"), []byte("key2"), []byte("key3"), []byte("key4")}
	values := []ksuid.KSUID{ksuid.New(), ksuid.New(), ksuid.New(), ksuid.New()}

	for i := range keys {
		tree.Insert(keys[i], values[i])
	}

	// Save to file
	filename := "/tmp/bptree_test.dat"
	defer os.Remove(filename)

	if err := tree.Save(filename); err != nil {
		t.Fatalf("Failed to save tree: %v", err)
	}

	// Load from file
	loadedTree, err := LoadBPlusTree(filename)
	if err != nil {
		t.Fatalf("Failed to load tree: %v", err)
	}

	// Verify loaded tree has same properties
	if loadedTree.order != tree.order {
		t.Fatalf("Expected order %d, got %d", tree.order, loadedTree.order)
	}
	if loadedTree.height != tree.height {
		t.Fatalf("Expected height %d, got %d", tree.height, loadedTree.height)
	}

	// Verify all keys are present with correct values
	for i, key := range keys {
		if v, found := loadedTree.Search(key); !found || !bytes.Equal(v.Bytes(), values[i].Bytes()) {
			t.Fatalf("Expected to find %s with value %v, got %v", key, values[i], v)
		}
	}

	// Verify non-existing key is not found
	if _, found := loadedTree.Search([]byte("nonexistent")); found {
		t.Fatal("Expected not to find nonexistent key")
	}
}

func TestBPlusTree_SaveLoadEmpty(t *testing.T) {
	tree := NewBPlusTree(5)

	filename := "/tmp/bptree_empty_test.dat"
	defer os.Remove(filename)

	if err := tree.Save(filename); err != nil {
		t.Fatalf("Failed to save empty tree: %v", err)
	}

	loadedTree, err := LoadBPlusTree(filename)
	if err != nil {
		t.Fatalf("Failed to load empty tree: %v", err)
	}

	if loadedTree.order != tree.order {
		t.Fatalf("Expected order %d, got %d", tree.order, loadedTree.order)
	}
	if loadedTree.height != tree.height {
		t.Fatalf("Expected height %d, got %d", tree.height, loadedTree.height)
	}
}

func TestBPlusTree_SaveLoadConcurrent(t *testing.T) {
	tree := NewBPlusTree(3)

	// Insert some data
	for i := 0; i < 10; i++ {
		key := []byte(fmt.Sprintf("key%d", i))
		val := ksuid.New()
		tree.Insert(key, val)
	}

	filename := "/tmp/bptree_concurrent_test.dat"
	defer os.Remove(filename)

	var wg sync.WaitGroup
	wg.Add(2)

	// Save in one goroutine
	go func() {
		defer wg.Done()
		if err := tree.Save(filename); err != nil {
			t.Errorf("Failed to save tree: %v", err)
		}
	}()

	// Insert more data in another goroutine
	go func() {
		defer wg.Done()
		for i := 10; i < 20; i++ {
			key := []byte(fmt.Sprintf("key%d", i))
			val := ksuid.New()
			tree.Insert(key, val)
		}
	}()

	wg.Wait()

	// Load and verify
	loadedTree, err := LoadBPlusTree(filename)
	if err != nil {
		t.Fatalf("Failed to load tree: %v", err)
	}

	// Should have the first 10 keys
	for i := 0; i < 10; i++ {
		key := []byte(fmt.Sprintf("key%d", i))
		if _, found := loadedTree.Search(key); !found {
			t.Errorf("Key %s not found in loaded tree", key)
		}
	}
}

func TestBPlusTree_Delete(t *testing.T) {
	tree := NewBPlusTree(3)

	key1 := []byte("key1")
	val1 := ksuid.New()
	tree.Insert(key1, val1)

	if _, found := tree.Search(key1); !found {
		t.Fatal("Key should be found after insert")
	}

	if !tree.Delete(key1) {
		t.Fatal("Delete should return true for existing key")
	}

	if _, found := tree.Search(key1); found {
		t.Fatal("Key should not be found after delete")
	}

	if tree.Delete(key1) {
		t.Fatal("Delete should return false for non-existing key")
	}
}

func TestBPlusTree_ConcurrentInsertSearch(t *testing.T) {
	tree := NewBPlusTree(3)
	var wg sync.WaitGroup
	numGoroutines := 10
	keysPerGoroutine := 100

	// Insert concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < keysPerGoroutine; j++ {
				key := []byte(fmt.Sprintf("key%d_%d", id, j))
				val := ksuid.New()
				tree.Insert(key, val)
			}
		}(i)
	}
	wg.Wait()

	// Search concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < keysPerGoroutine; j++ {
				key := []byte(fmt.Sprintf("key%d_%d", id, j))
				if _, found := tree.Search(key); !found {
					t.Errorf("Key %s not found", key)
				}
			}
		}(i)
	}
	wg.Wait()
}

func TestBPlusTree_ConcurrentInsertDelete(t *testing.T) {
	tree := NewBPlusTree(3)
	var wg sync.WaitGroup
	numGoroutines := 10
	keysPerGoroutine := 50

	// Insert concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < keysPerGoroutine; j++ {
				key := []byte(fmt.Sprintf("key%d_%d", id, j))
				val := ksuid.New()
				tree.Insert(key, val)
			}
		}(i)
	}
	wg.Wait()

	// Delete concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < keysPerGoroutine; j++ {
				key := []byte(fmt.Sprintf("key%d_%d", id, j))
				if !tree.Delete(key) {
					t.Errorf("Failed to delete key %s", key)
				}
			}
		}(i)
	}
	wg.Wait()

	// Verify all deleted
	for i := 0; i < numGoroutines; i++ {
		for j := 0; j < keysPerGoroutine; j++ {
			key := []byte(fmt.Sprintf("key%d_%d", i, j))
			if _, found := tree.Search(key); found {
				t.Errorf("Key %s should be deleted", key)
			}
		}
	}
}

func TestBPlusTree_ConcurrentReadWrite(t *testing.T) {
	tree := NewBPlusTree(3)
	var wg sync.WaitGroup

	// Pre-insert some keys
	for i := 0; i < 10; i++ {
		key := []byte(fmt.Sprintf("pre%d", i))
		val := ksuid.New()
		tree.Insert(key, val)
	}

	numWriters := 2
	numReaders := 2
	operations := 10

	// Writers: insert new keys
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				key := []byte(fmt.Sprintf("write%d_%d", id, j))
				val := ksuid.New()
				tree.Insert(key, val)
			}
		}(i)
	}

	// Readers: search for existing keys (pre-inserted and being inserted)
	foundCount := int64(0)
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			localFound := 0
			for j := 0; j < operations*2; j++ { // More searches than inserts
				// Search for pre-inserted keys
				key := []byte(fmt.Sprintf("pre%d", j%50))
				if _, found := tree.Search(key); found {
					localFound++
				}
				// Search for keys being inserted
				key2 := []byte(fmt.Sprintf("write%d_%d", id, j%operations))
				if _, found := tree.Search(key2); found {
					localFound++
				}
			}
			atomic.AddInt64(&foundCount, int64(localFound))
		}(i)
	}

	wg.Wait()

	// Verify some searches found keys
	if foundCount == 0 {
		t.Error("No keys were found during concurrent read/write operations")
	}

	// Verify all inserted keys can be found after operations
	for i := 0; i < numWriters; i++ {
		for j := 0; j < operations; j++ {
			key := []byte(fmt.Sprintf("write%d_%d", i, j))
			if _, found := tree.Search(key); !found {
				t.Errorf("Key %s not found after concurrent operations", key)
			}
		}
	}
}

func BenchmarkBPlusTree_Insert(b *testing.B) {
	tree := NewBPlusTree(3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("key%d", i))
		val := ksuid.New()
		tree.Insert(key, val)
	}
}

func BenchmarkBPlusTree_Search(b *testing.B) {
	tree := NewBPlusTree(3)
	// Pre-insert
	for i := 0; i < 1000; i++ {
		key := []byte(fmt.Sprintf("key%d", i))
		val := ksuid.New()
		tree.Insert(key, val)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("key%d", i%1000))
		tree.Search(key)
	}
}

func BenchmarkBPlusTree_ConcurrentInsert(b *testing.B) {
	tree := NewBPlusTree(3)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := []byte(fmt.Sprintf("key%d", i))
			val := ksuid.New()
			tree.Insert(key, val)
			i++
		}
	})
}

func TestBPlusTree_SplitInternalNode(t *testing.T) {
	tree := NewBPlusTree(3)

	keys := [][]byte{[]byte("key1"), []byte("key2"), []byte("key3"), []byte("key4"), []byte("key5")}
	values := []ksuid.KSUID{ksuid.New(), ksuid.New(), ksuid.New(), ksuid.New(), ksuid.New()}

	for i := range keys {
		tree.Insert(keys[i], values[i])
	}

	// Check if the tree height increased due to splits
	if tree.height != 3 {
		t.Fatalf("Expected tree height 3, got %d", tree.height)
	}

	// Check if all keys are present
	for i, key := range keys {
		if v, found := tree.Search(key); !found || !bytes.Equal(v.Bytes(), values[i].Bytes()) {
			t.Fatalf("Expected to find %s with value %v, got %v", key, values[i], v)
		}
	}
}
