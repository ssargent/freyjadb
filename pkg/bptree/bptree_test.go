package bptree

import (
	"bytes"
	"fmt"
	"os"
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

	// Insert enough keys to force root split and height=3
	keys := make([][]byte, 0)
	values := make([]ksuid.KSUID, 0)

	// Insert 8 keys to ensure we get height=2
	for i := 0; i < 8; i++ {
		key := []byte(fmt.Sprintf("%02d", i))
		val := ksuid.New()
		keys = append(keys, key)
		values = append(values, val)
		tree.Insert(key, val)
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
