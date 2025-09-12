//go:build ignore

package bptree

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/segmentio/ksuid"
)

func TestBPlusTree_SaveLoadConcurrent(t *testing.T) {
	tree := NewBPlusTree(3)

	// Insert some data
	for i := 0; i < 5; i++ {
		key := []byte(fmt.Sprintf("key%d", i))
		val := ksuid.New()
		tree.Insert(key, val)
	}

	filename := "/tmp/bptree_concurrent_test.dat"
	defer os.Remove(filename)

	// Save first
	if err := tree.Save(filename); err != nil {
		t.Fatalf("Failed to save tree: %v", err)
	}

	// Insert more data
	for i := 5; i < 10; i++ {
		key := []byte(fmt.Sprintf("key%d", i))
		val := ksuid.New()
		tree.Insert(key, val)
	}

	// Load and verify
	loadedTree, err := LoadBPlusTree(filename)
	if err != nil {
		t.Fatalf("Failed to load tree: %v", err)
	}

	// Should have the first 5 keys
	for i := 0; i < 5; i++ {
		key := []byte(fmt.Sprintf("key%d", i))
		if _, found := loadedTree.Search(key); !found {
			t.Errorf("Key %s not found in loaded tree", key)
		}
	}
}

func TestBPlusTree_ConcurrentInsertSearch(t *testing.T) {
	tree := NewBPlusTree(3)
	var wg sync.WaitGroup
	numGoroutines := 10
	keysPerGoroutine := 10

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
	keysPerGoroutine := 5

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
	operations := 5

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
