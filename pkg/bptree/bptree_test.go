package bptree

import (
	"bytes"
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
