// File: bptree.go
package bptree

import (
	"sync"
)

// DefaultOrder is the fallback branching factor if a user-supplied order is too small.
const DefaultOrder = 4

// compare is a helper function for ordering keys.
// Replace this with something more flexible (e.g., a comparator interface or Go 1.18+ generics constraints).
func compare[K comparable](a, b K) int {
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
}

// findChildIndex determines which child pointer to follow
// (or where to insert a new key) in an internal node.
func findChildIndex[K comparable](keys []K, searchKey K) int {
	// Linear scan for simplicity; you might use binary search for better performance.
	for i, k := range keys {
		if compare(searchKey, k) < 0 {
			return i
		}
	}
	return len(keys)
}

// BPlusTree is our main structure.
type BPlusTree[K comparable, V any] struct {
	root   *node[K, V]
	order  int
	height int
	m      sync.RWMutex
}

func (tree *BPlusTree[K, V]) Height() int {
	return tree.height
}

// node represents both internal and leaf nodes in the B+Tree.
// Each node has its own RWMutex for concurrency control.
type node[K comparable, V any] struct {
	mutex    sync.RWMutex // per-node latch
	isLeaf   bool
	keys     []K
	children []*node[K, V] // used if !isLeaf
	values   []V           // used if isLeaf
	parent   *node[K, V]
	next     *node[K, V] // leaf-link pointer, for range scans
}

// NewBPlusTree creates and returns a B+Tree with the given order.
// If the specified order < 3, we fall back to DefaultOrder.
func NewBPlusTree[K comparable, V any](order int) *BPlusTree[K, V] {
	if order < 3 {
		order = DefaultOrder
	}
	rootNode := &node[K, V]{
		isLeaf:   true,
		keys:     make([]K, 0, order),
		values:   make([]V, 0, order),
		children: make([]*node[K, V], 0),
	}
	return &BPlusTree[K, V]{
		root:   rootNode,
		order:  order,
		height: 1,
	}
}

// Search locates the value associated with `key` (if it exists).
// Demonstrates "latch coupling" to move down the tree with shared (read) locks.
func (tree *BPlusTree[K, V]) Search(key K) (V, bool) {
	current := tree.root
	if current == nil {
		var zero V
		return zero, false
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
		if k == key {
			val := current.values[i]
			// Release leaf lock and return
			current.mutex.RUnlock()
			return val, true
		}
	}

	// Key not found
	current.mutex.RUnlock()
	var zero V
	return zero, false
}

// Insert adds a (key, value) pair to the B+Tree. Demonstrates a
// simplified concurrency approach (read-latching during descent,
// then switching to an exclusive lock at the leaf).
func (tree *BPlusTree[K, V]) Insert(key K, value V) {
	// If there's no root, create one (edge case)
	if tree.root == nil {
		tree.root = &node[K, V]{
			isLeaf: true,
			keys:   []K{key},
			values: []V{value},
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
	insertKeyValueInLeaf(current, key, value)

	// Check overflow
	if len(current.keys) > tree.order {
		tree.m.Lock()
		defer tree.m.Unlock()
		tree.splitLeaf(current)
	}
}

// tree.m.Unlock() // Removed redundant unlock
func insertKeyValueInLeaf[K comparable, V any](leaf *node[K, V], key K, value V) {
	idx := 0
	for idx < len(leaf.keys) && compare(leaf.keys[idx], key) < 0 {
		idx++
	}
	// Check if the key already exists
	if idx < len(leaf.keys) && compare(leaf.keys[idx], key) == 0 {
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
func (tree *BPlusTree[K, V]) splitLeaf(leaf *node[K, V]) {
	mid := len(leaf.keys) / 2

	newLeaf := &node[K, V]{
		isLeaf: true,
		keys:   append([]K{}, leaf.keys[mid:]...),
		values: append([]V{}, leaf.values[mid:]...),
		next:   leaf.next,
		parent: leaf.parent,
	}

	// Adjust the original leaf
	leaf.keys = leaf.keys[:mid]
	leaf.values = leaf.values[:mid]
	leaf.next = newLeaf

	// If the leaf is the root (no parent), create a new root
	if leaf.parent == nil {
		newRoot := &node[K, V]{
			isLeaf:   false,
			keys:     []K{newLeaf.keys[0]},
			children: []*node[K, V]{leaf, newLeaf},
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
func insertKeyInParent[K comparable, V any](tree *BPlusTree[K, V],
	parent *node[K, V], key K,
	leftChild, rightChild *node[K, V]) {

	idx := 0
	for idx < len(parent.keys) && compare(parent.keys[idx], key) < 0 {
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
func splitInternalNode[K comparable, V any](tree *BPlusTree[K, V], internal *node[K, V]) {
	mid := len(internal.keys) / 2
	splitKey := internal.keys[mid]

	newInternal := &node[K, V]{
		isLeaf:   false,
		keys:     append([]K{}, internal.keys[mid+1:]...),
		children: append([]*node[K, V]{}, internal.children[mid+1:]...),
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
		newRoot := &node[K, V]{
			isLeaf:   false,
			keys:     []K{splitKey},
			children: []*node[K, V]{internal, newInternal},
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
