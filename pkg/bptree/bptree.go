// File: bptree.go
package bptree

import (
	"bytes"
	"sync"

	"github.com/segmentio/ksuid"
)

// DefaultOrder is the fallback branching factor if a user-supplied order is too small.
const DefaultOrder = 4

// compare is a helper function for ordering keys.
// Replace this with something more flexible (e.g., a comparator interface or Go 1.18+ generics constraints).
/* func compare[K comparable](a, b K) int {
	switch a := any(a).(type) {
	case int:
		b := any(b).(int)
		return a - b
	case string:
		b := any(b).(string)
		if a < b {
			return -1
		} else if a > b {
			return 1
		}
		return 0
	}
	panic("compare: type not supported or not implemented yet")
} */

// findChildIndex determines which child pointer to follow
// (or where to insert a new key) in an internal node.
func findChildIndex(keys [][]byte, searchKey []byte) int {
	// Linear scan for simplicity; you might use binary search for better performance.
	for i, k := range keys {
		if bytes.Equal(searchKey, k) {
			return i // Exact match found
		}
	}
	return len(keys)
}

// BPlusTree is our main structure.
type BPlusTree struct {
	root   *node
	order  int
	height int
	m      sync.RWMutex
}

func (tree *BPlusTree) Height() int {
	return tree.height
}

// node represents both internal and leaf nodes in the B+Tree.
// Each node has its own RWMutex for concurrency control.
type node struct {
	mutex    sync.RWMutex // per-node latch
	isLeaf   bool
	keys     [][]byte
	children []*node        // used if !isLeaf
	values   []*ksuid.KSUID // used if isLeaf
	parent   *node
	next     *node // leaf-link pointer, for range scans
}

// NewBPlusTree creates and returns a B+Tree with the given order.
// If the specified order < 3, we fall back to DefaultOrder.
func NewBPlusTree(order int) *BPlusTree {
	if order < 3 {
		order = DefaultOrder
	}
	rootNode := &node{
		isLeaf:   true,
		keys:     make([][]byte, 0, order),
		values:   make([]*ksuid.KSUID, 0, order),
		children: make([]*node, 0),
	}
	return &BPlusTree{
		root:   rootNode,
		order:  order,
		height: 1,
	}
}

// Search locates the value associated with `key` (if it exists).
// Demonstrates "latch coupling" to move down the tree with shared (read) locks.
func (tree *BPlusTree) Search(key []byte) (*ksuid.KSUID, bool) {
	current := tree.root
	if current == nil {
		return nil, false
	}

	// Lock the current node in shared mode
	current.mutex.RLock()

	for !current.isLeaf {
		idx := findChildIndex(current.keys, key)
		child := current.children[idx]

		// Latch coupling: acquire child's read lock before releasing the current node
		child.mutex.RLock()
		current.mutex.RUnlock()

		// Move to child
		current = child
	}

	// We're now at a leaf, holding its read lock
	// Search its keys
	for i, k := range current.keys {
		if bytes.Equal(key, k) {
			val := current.values[i]
			// Release leaf lock and return
			current.mutex.RUnlock()
			return val, true
		}

		// we might not have an exact match, we need to support starts with
		if bytes.HasPrefix(key, k) {
			val := current.values[i]
			current.mutex.RUnlock()
			return val, true
		}
	}

	// Key not found
	current.mutex.RUnlock()
	return nil, false
}

// Insert adds a (key, value) pair to the B+Tree. Demonstrates a
// simplified concurrency approach (read-latching during descent,
// then switching to an exclusive lock at the leaf).
func (tree *BPlusTree) Insert(key []byte, value ksuid.KSUID) {
	// If there's no root, create one (edge case)
	if tree.root == nil {
		tree.root = &node{
			isLeaf: true,
			keys:   [][]byte{key},
			values: []*ksuid.KSUID{&value},
		}
		return
	}

	current := tree.root
	current.mutex.RLock()
	// Traverse down with read locks until we reach a leaf
	for !current.isLeaf {
		idx := findChildIndex(current.keys, key)
		child := current.children[idx]

		// Acquire child's read lock before releasing parent's read lock
		child.mutex.RLock()
		current.mutex.RUnlock()

		current = child
	}

	// We have a read lock on a leaf node, but need to modify => exclusive lock
	current.mutex.RUnlock()
	// Lock the current node in exclusive mode
	current.mutex.Lock()
	defer current.mutex.Unlock()

	// Insert the key/value in sorted order
	insertKeyValueInLeaf(current, key, &value)

	// Check overflow
	if len(current.keys) > tree.order {
		tree.m.Lock()
		defer tree.m.Unlock()
		tree.splitLeaf(current)
	}
}

// tree.m.Unlock() // Removed redundant unlock
func insertKeyValueInLeaf(leaf *node, key []byte, value *ksuid.KSUID) {
	idx := 0
	for idx < len(leaf.keys) && bytes.Compare(leaf.keys[idx], key) < 0 {
		idx++
	}
	// Check if the key already exists
	if idx < len(leaf.keys) && bytes.Equal(leaf.keys[idx], key) {
		leaf.values[idx] = value // Update the existing value
		return
	}
	leaf.keys = append(leaf.keys, key)
	leaf.values = append(leaf.values, value)

	// Shift elements to make room at idx
	copy(leaf.keys[idx+1:], leaf.keys[idx:])
	leaf.keys[idx] = key

	copy(leaf.values[idx+1:], leaf.values[idx:])
	leaf.values[idx] = value
}

// splitLeaf handles splitting a leaf node that has overflowed.
// Must be called with leaf node locked in exclusive mode.
func (tree *BPlusTree) splitLeaf(leaf *node) {
	mid := len(leaf.keys) / 2

	newLeaf := &node{
		isLeaf: true,
		keys:   append(make([][]byte, 0), leaf.keys[mid:]...),
		values: append(make([]*ksuid.KSUID, 0), leaf.values[mid:]...),
		next:   leaf.next,
		parent: leaf.parent,
	}

	// Adjust the original leaf
	leaf.keys = leaf.keys[:mid]
	leaf.values = leaf.values[:mid]
	leaf.next = newLeaf

	// If the leaf is the root (no parent), create a new root
	if leaf.parent == nil {
		newRoot := &node{
			isLeaf:   false,
			keys:     [][]byte{newLeaf.keys[0]},
			children: []*node{leaf, newLeaf},
		}

		leaf.parent = newRoot
		newLeaf.parent = newRoot

		tree.root = newRoot
		tree.height++

		return
	}

	// Otherwise, insert the new leaf's first key into the parent
	parent := leaf.parent

	// We must lock the parent in exclusive mode before modifying it
	parent.mutex.Lock()
	defer parent.mutex.Unlock()

	insertKeyInParent(tree, parent, newLeaf.keys[0], leaf, newLeaf)
}

// insertKeyInParent inserts `key` and links `leftChild` & `rightChild` in the parent.
// Must be called with the parent locked in exclusive mode.
func insertKeyInParent(tree *BPlusTree,
	parent *node, key []byte,
	leftChild, rightChild *node) {

	idx := 0
	for idx < len(parent.keys) && bytes.Compare(parent.keys[idx], key) < 0 {
		idx++
	}

	parent.keys = append(parent.keys, key)
	copy(parent.keys[idx+1:], parent.keys[idx:])
	parent.keys[idx] = key

	parent.children = append(parent.children, rightChild)
	copy(parent.children[idx+2:], parent.children[idx+1:])
	parent.children[idx+1] = rightChild

	rightChild.parent = parent

	// Check for overflow
	if len(parent.keys) > tree.order {
		splitInternalNode(tree, parent)
	}
}

// splitInternalNode handles splitting an internal node that has overflowed.
// Must be called with 'internal' locked in exclusive mode.
func splitInternalNode(tree *BPlusTree, internal *node) {
	mid := len(internal.keys) / 2
	splitKey := internal.keys[mid]

	newInternal := &node{
		isLeaf:   false,
		keys:     append(make([][]byte, 0), internal.keys[mid:]...),
		children: append([]*node{}, internal.children[mid+1:]...),
		parent:   internal.parent,
	}

	// Update children's parent pointers
	for _, child := range newInternal.children {
		child.parent = newInternal
	}

	// Adjust the original internal node
	internal.keys = internal.keys[:mid]
	internal.children = internal.children[:mid+1]

	if internal.parent == nil {
		// Create a new root
		newRoot := &node{
			isLeaf:   false,
			keys:     [][]byte{splitKey},
			children: []*node{internal, newInternal},
		}
		internal.parent = newRoot
		newInternal.parent = newRoot
		tree.root = newRoot
		tree.height++
		return
	}

	// Insert splitKey into parent
	parent := internal.parent
	parent.mutex.Lock()
	defer parent.mutex.Unlock()

	insertKeyInParent(tree, parent, splitKey, internal, newInternal)
}
