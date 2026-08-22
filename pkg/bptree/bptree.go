// File: bptree.go
// Package bptree provides a thread-safe B+Tree implementation.
// It supports concurrent reads and writes using per-node RWMutex and latch coupling.
// All operations (Insert, Search, Delete) are safe for concurrent use.
package bptree

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/segmentio/ksuid"
)

// DefaultOrder is the fallback branching factor if a user-supplied order is too small.
const DefaultOrder = 4

// safeUint32 converts a non-negative int to uint32, returning an error instead
// of silently truncating/overflowing when the value is out of range. Used when
// serializing sizes and counts that are assumed small but come from int-typed
// fields or len() results.
func safeUint32(n int) (uint32, error) {
	if n < 0 {
		return 0, fmt.Errorf("value %d is negative, cannot convert to uint32", n)
	}
	if n > math.MaxUint32 {
		return 0, fmt.Errorf("value %d overflows uint32", n)
	}
	return uint32(n), nil
}

// findChildIndex determines which child pointer to follow for a given search key in an internal node.
// This implements the B+Tree navigation logic where:
// - For internal node with keys [k1, k2, ..., kn] and children [c0, c1, ..., cn]
// - If searchKey < k1, return 0 (follow c0)
// - If k1 <= searchKey < k2, return 1 (follow c1)
// - ...
// - If searchKey >= kn, return n (follow cn)
//
// Uses linear search for simplicity; could be optimized with binary search for large orders.
// Time complexity: O(order)
func findChildIndex(keys [][]byte, searchKey []byte) int {
	for i, k := range keys {
		if bytes.Compare(searchKey, k) < 0 {
			return i
		}
	}
	return len(keys)
}

// BPlusTree represents a thread-safe B+Tree data structure optimized for concurrent access.
// Key features:
// - Thread-safe for concurrent reads and writes
// - Uses per-node RWMutex for fine-grained concurrency control
// - Implements latch coupling for efficient tree traversal
// - All keys are stored in leaf nodes, internal nodes contain separator keys
// - Supports range scans via leaf node linking
//
// NOTE: Current implementation uses tree-level exclusive locking during node splits,
// which can limit concurrency under high write loads. For applications requiring
// higher write throughput, consider implementing a lock-free B+Tree or using
// finer-grained synchronization techniques.
type BPlusTree struct {
	root             *node        // Root node of the tree
	order            int          // Maximum number of keys per node
	height           int          // Height of the tree (1 for single leaf)
	m                sync.RWMutex // Protects root and height modifications
	checkpointTicker *time.Ticker // Ticker for periodic checkpoints
	checkpointDone   chan bool    // Channel to stop checkpointing

}

// Height returns the current height of the B+Tree.
// Height is defined as the number of levels in the tree (1 for a single leaf node).
// This method is thread-safe and can be called concurrently with other operations.
func (tree *BPlusTree) Height() int {
	tree.m.RLock()
	h := tree.height
	tree.m.RUnlock()
	return h
}

// node represents a single node in the B+Tree, which can be either an internal node or a leaf node.
// Each node maintains its own RWMutex for fine-grained concurrency control.
//
// For internal nodes (!isLeaf):
// - keys: separator keys that guide navigation to child nodes
// - children: pointers to child nodes (len(children) = len(keys) + 1)
// - values: nil (not used)
//
// For leaf nodes (isLeaf):
// - keys: the actual data keys
// - children: nil (not used)
// - values: the corresponding values for each key
// - next: pointer to the next leaf node for range scan support
//
// Thread safety: Each node has its own RWMutex that protects all its fields.
// Multiple readers can access a node simultaneously, but writers get exclusive access.
type node struct {
	mutex    sync.RWMutex   // Per-node latch for concurrency control
	isLeaf   bool           // True if this is a leaf node, false for internal node
	keys     [][]byte       // Keys stored in this node
	children []*node        // Child nodes (internal nodes only)
	values   []*ksuid.KSUID // Values corresponding to keys (leaf nodes only)
	parent   *node          // Parent node (nil for root)
	next     *node          // Next leaf node for range scans (leaf nodes only)
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

// Search performs a point lookup for the given key in the B+Tree.
// Returns the associated value and true if the key exists, or nil and false if not found.
//
// This method is thread-safe and can be called concurrently with other operations.
// It uses latch coupling for efficient traversal:
// 1. Acquires tree-level read lock to safely access the root
// 2. Traverses down the tree using read locks on each node
// 3. Uses latch coupling: acquires child lock before releasing parent lock
// 4. Performs linear search on the leaf node keys
//
// Time complexity: O(log n) for tree traversal + O(order) for leaf search
// Space complexity: O(1) additional space
func (tree *BPlusTree) Search(key []byte) (*ksuid.KSUID, bool) {
	tree.m.RLock()
	current := tree.root
	if current == nil {
		tree.m.RUnlock()
		return nil, false
	}

	// Lock the current node in shared mode
	current.mutex.RLock()
	tree.m.RUnlock()

	// Traverse from root to leaf using latch coupling
	// Latch coupling ensures we always hold at least one lock during traversal
	for !current.isLeaf {
		idx := findChildIndex(current.keys, key)
		child := current.children[idx]

		// Latch coupling: acquire child's read lock BEFORE releasing parent's lock
		// This prevents race conditions during tree modifications
		child.mutex.RLock()
		current.mutex.RUnlock()

		// Now safely move to child (we still hold child's lock)
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
	}

	// Key not found
	current.mutex.RUnlock()
	return nil, false
}

// Insert adds or updates a key-value pair in the B+Tree.
// If the key already exists, its value is updated. If the key is new, it's inserted.
//
// This method is thread-safe and can be called concurrently with other operations.
// It uses a hybrid locking strategy for optimal concurrency:
// 1. Acquires tree-level read lock to safely access the root
// 2. Traverses down the tree using read locks (latch coupling)
// 3. Switches to exclusive lock on the target leaf node for modification
// 4. If the leaf overflows, acquires tree-level exclusive lock for splitting
//
// NOTE: Tree-level exclusive lock during splitting can limit concurrency under
// high write loads. This is a trade-off between correctness and performance.
// The insertion process:
// - Finds the correct leaf node for the key
// - Inserts the key-value pair in sorted order
// - Handles node splitting if capacity is exceeded
// - Maintains B+Tree balance and invariants
//
// Time complexity: O(log n) for traversal + O(order) for insertion/splitting
// Space complexity: O(order) for temporary operations during splitting
func (tree *BPlusTree) Insert(key []byte, value ksuid.KSUID) {
	tree.m.RLock()
	// If there's no root, create one (edge case)
	if tree.root == nil {
		tree.m.RUnlock()
		tree.m.Lock()
		if tree.root == nil {
			tree.root = &node{
				isLeaf: true,
				keys:   [][]byte{key},
				values: []*ksuid.KSUID{&value},
			}
			tree.height = 1
		}
		tree.m.Unlock()
		return
	}

	current := tree.root
	current.mutex.RLock()
	tree.m.RUnlock()
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

// Delete removes a key-value pair from the B+Tree if the key exists.
// Returns true if the key was found and removed, false if the key was not found.
//
// This method is thread-safe and can be called concurrently with other operations.
// It uses the same locking strategy as Insert:
// 1. Acquires tree-level read lock to safely access the root
// 2. Traverses down the tree using read locks (latch coupling)
// 3. Switches to exclusive lock on the target leaf node for modification
//
// NOTE: Like Insert, this method may acquire tree-level exclusive lock during
// structural modifications, which can limit concurrency under high loads.
//
// Note: This implementation provides basic deletion without rebalancing.
// In a production B+Tree, you would typically implement redistribution and merging
// to maintain optimal space utilization when nodes become underfull.
//
// Time complexity: O(log n) for traversal + O(order) for key removal
// Space complexity: O(1) additional space
func (tree *BPlusTree) Delete(key []byte) bool {
	tree.m.RLock()
	current := tree.root
	if current == nil {
		tree.m.RUnlock()
		return false
	}
	current.mutex.RLock()
	tree.m.RUnlock()

	// Traverse down with read locks until we reach a leaf
	for !current.isLeaf {
		idx := findChildIndex(current.keys, key)
		child := current.children[idx]

		// Latch coupling: acquire child's read lock before releasing the current node
		child.mutex.RLock()
		current.mutex.RUnlock()

		current = child
	}

	// We have a read lock on a leaf node, but need to modify => exclusive lock
	current.mutex.RUnlock()
	current.mutex.Lock()
	defer current.mutex.Unlock()

	// Find and remove the key
	for i, k := range current.keys {
		if bytes.Equal(key, k) {
			// Remove the key and value
			current.keys = append(current.keys[:i], current.keys[i+1:]...)
			current.values = append(current.values[:i], current.values[i+1:]...)
			return true
		}
	}

	// Key not found
	return false
}

// insertKeyValueInLeaf inserts a key-value pair into a leaf node at the correct sorted position.
// If the key already exists, it updates the value. The leaf node must be locked exclusively.
//
// Algorithm:
// 1. Find the insertion point using binary search (linear scan here for simplicity)
// 2. If key exists, update the value in place
// 3. If key is new, make room by shifting elements and insert at the correct position
//
// This maintains the sorted order invariant of B+Tree leaf nodes.
func insertKeyValueInLeaf(leaf *node, key []byte, value *ksuid.KSUID) {
	// Find insertion point (could be optimized with binary search)
	idx := 0
	for idx < len(leaf.keys) && bytes.Compare(leaf.keys[idx], key) < 0 {
		idx++
	}

	// Check if the key already exists at this position
	if idx < len(leaf.keys) && bytes.Equal(leaf.keys[idx], key) {
		leaf.values[idx] = value // Update existing value
		return
	}

	// Insert new key-value pair
	// First, append placeholders to extend slices
	leaf.keys = append(leaf.keys, key)
	leaf.values = append(leaf.values, value)

	// Shift elements to the right to make room at idx
	copy(leaf.keys[idx+1:], leaf.keys[idx:])
	copy(leaf.values[idx+1:], leaf.values[idx:])

	// Insert the new key-value pair at the correct position
	leaf.keys[idx] = key
	leaf.values[idx] = value
}

// splitLeaf handles splitting a leaf node that has overflowed after insertion.
// This maintains the B+Tree invariant that no node exceeds the maximum order.
//
// Algorithm:
// 1. Split the leaf node into two halves at the midpoint
// 2. Create a new leaf node with the right half of keys/values
// 3. Update the linked list pointers for range scan support
// 4. If this is the root, create a new root with the split key
// 5. Otherwise, insert the split key into the parent node
//
// The split key (first key of the new leaf) is promoted to the parent level.
// This ensures balanced tree growth and maintains search properties.
//
// Must be called with leaf node locked in exclusive mode and tree.m held.
func (tree *BPlusTree) splitLeaf(leaf *node) {
	// Calculate split point (middle of the node)
	mid := len(leaf.keys) / 2

	// Create new leaf node with right half of keys and values
	newLeaf := &node{
		isLeaf: true,
		keys:   append(make([][]byte, 0), leaf.keys[mid:]...),         // Copy right half of keys
		values: append(make([]*ksuid.KSUID, 0), leaf.values[mid:]...), // Copy right half of values
		next:   leaf.next,                                             // Link to the original next leaf
		parent: leaf.parent,
	}

	// Adjust the original leaf to contain only left half
	leaf.keys = leaf.keys[:mid]
	leaf.values = leaf.values[:mid]
	leaf.next = newLeaf // Update linked list pointer

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

	insertKeyInParent(tree, parent, newLeaf.keys[0], newLeaf)
}

// insertKeyInParent inserts `key` and links `rightChild` in the parent.
// The left sibling is already linked in the parent's children slice, so only
// the new right child needs to be inserted.
// Must be called with the parent locked in exclusive mode.
func insertKeyInParent(tree *BPlusTree,
	parent *node, key []byte,
	rightChild *node) {
	idx := 0
	for idx < len(parent.keys) && bytes.Compare(parent.keys[idx], key) < 0 {
		idx++
	}

	parent.keys = append(parent.keys, key)
	copy(parent.keys[idx+1:], parent.keys[idx:])
	parent.keys[idx] = key

	parent.children = append(parent.children, nil)
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

	insertKeyInParent(tree, parent, splitKey, newInternal)
}

// collectNodesBFS assigns a stable ID to every node reachable from root
// (breadth-first), for use as node references in the serialized file.
func collectNodesBFS(root *node) (map[*node]uint32, []*node) {
	nodeMap := make(map[*node]uint32)
	var nodes []*node
	id := uint32(0)

	queue := []*node{root}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if _, exists := nodeMap[current]; !exists {
			nodeMap[current] = id
			nodes = append(nodes, current)
			id++

			// Add children to queue
			for _, child := range current.children {
				if child != nil {
					queue = append(queue, child)
				}
			}
		}
	}

	return nodeMap, nodes
}

// writeTreeMetadata writes the order/height/rootID/nodeCount header block
// shared by Save and writeEmptyTree.
func writeTreeMetadata(file *os.File, order, height int, rootID uint32, nodeCount int) error {
	orderU32, err := safeUint32(order)
	if err != nil {
		return fmt.Errorf("invalid tree order: %w", err)
	}
	if err = binary.Write(file, binary.LittleEndian, orderU32); err != nil {
		return fmt.Errorf("failed to write order: %w", err)
	}

	heightU32, err := safeUint32(height)
	if err != nil {
		return fmt.Errorf("invalid tree height: %w", err)
	}
	if err = binary.Write(file, binary.LittleEndian, heightU32); err != nil {
		return fmt.Errorf("failed to write height: %w", err)
	}

	if err = binary.Write(file, binary.LittleEndian, rootID); err != nil {
		return fmt.Errorf("failed to write root ID: %w", err)
	}

	nodeCountU32, err := safeUint32(nodeCount)
	if err != nil {
		return fmt.Errorf("invalid node count: %w", err)
	}
	if err = binary.Write(file, binary.LittleEndian, nodeCountU32); err != nil {
		return fmt.Errorf("failed to write node count: %w", err)
	}

	return nil
}

// Save serializes the B+Tree to a binary file.
// This method is thread-safe and can be called concurrently with other operations.
// It acquires an exclusive lock on the tree to ensure consistency during serialization.
func (tree *BPlusTree) Save(filename string) error {
	tree.m.Lock()
	defer tree.m.Unlock()

	// Clean the filename to prevent path traversal
	filename = filepath.Clean(filename)
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() { _ = file.Close() }() // Ignore error in cleanup

	// If tree is empty, just write empty metadata
	if tree.root == nil {
		return tree.writeEmptyTree(file)
	}

	// Collect all nodes with IDs using breadth-first traversal
	nodeMap, nodes := collectNodesBFS(tree.root)

	if err = writeTreeMetadata(file, tree.order, tree.height, nodeMap[tree.root], len(nodes)); err != nil {
		return err
	}

	// Write each node
	for _, n := range nodes {
		if err = tree.writeNode(file, n, nodeMap); err != nil {
			return fmt.Errorf("failed to write node: %w", err)
		}
	}

	return nil
}

// writeEmptyTree writes metadata for an empty tree
func (tree *BPlusTree) writeEmptyTree(file *os.File) error {
	return writeTreeMetadata(file, tree.order, tree.height, 0, 0)
}

// writeNode serializes a single node to the file
func (tree *BPlusTree) writeNode(file *os.File, n *node, nodeMap map[*node]uint32) error {
	if err := writeNodeKeys(file, n); err != nil {
		return err
	}

	if n.isLeaf {
		if err := writeLeafValues(file, n, nodeMap); err != nil {
			return err
		}
	} else {
		if err := writeInternalChildren(file, n, nodeMap); err != nil {
			return err
		}
	}

	return writeNodeParent(file, n, nodeMap)
}

// writeNodeKeys writes the isLeaf flag followed by the node's keys.
func writeNodeKeys(file *os.File, n *node) error {
	isLeaf := uint8(0)
	if n.isLeaf {
		isLeaf = 1
	}
	if err := binary.Write(file, binary.LittleEndian, isLeaf); err != nil {
		return err
	}

	keyCount, err := safeUint32(len(n.keys))
	if err != nil {
		return fmt.Errorf("invalid key count: %w", err)
	}
	if err := binary.Write(file, binary.LittleEndian, keyCount); err != nil {
		return err
	}

	for _, key := range n.keys {
		keyLen, err := safeUint32(len(key))
		if err != nil {
			return fmt.Errorf("invalid key length: %w", err)
		}
		if err := binary.Write(file, binary.LittleEndian, keyLen); err != nil {
			return err
		}
		if _, err := file.Write(key); err != nil {
			return err
		}
	}

	return nil
}

// writeLeafValues writes a leaf node's values and its next-leaf pointer.
func writeLeafValues(file *os.File, n *node, nodeMap map[*node]uint32) error {
	for _, value := range n.values {
		if value == nil {
			// Write zero length for nil
			if err := binary.Write(file, binary.LittleEndian, uint32(0)); err != nil {
				return err
			}
			continue
		}

		ksuidBytes := value.Bytes()
		valueLen, err := safeUint32(len(ksuidBytes))
		if err != nil {
			return fmt.Errorf("invalid value length: %w", err)
		}
		if err := binary.Write(file, binary.LittleEndian, valueLen); err != nil {
			return err
		}
		if _, err := file.Write(ksuidBytes); err != nil {
			return err
		}
	}

	nextID := uint32(0)
	if n.next != nil {
		if id, exists := nodeMap[n.next]; exists {
			nextID = id
		}
	}
	return binary.Write(file, binary.LittleEndian, nextID)
}

// writeInternalChildren writes an internal node's child pointers as node IDs.
func writeInternalChildren(file *os.File, n *node, nodeMap map[*node]uint32) error {
	for _, child := range n.children {
		childID := uint32(0)
		if child != nil {
			if id, exists := nodeMap[child]; exists {
				childID = id
			}
		}
		if err := binary.Write(file, binary.LittleEndian, childID); err != nil {
			return err
		}
	}
	return nil
}

// writeNodeParent writes the node's parent pointer as a node ID.
func writeNodeParent(file *os.File, n *node, nodeMap map[*node]uint32) error {
	parentID := uint32(0)
	if n.parent != nil {
		if id, exists := nodeMap[n.parent]; exists {
			parentID = id
		}
	}
	return binary.Write(file, binary.LittleEndian, parentID)
}

// treeHeader holds the metadata block at the start of a serialized B+Tree file.
type treeHeader struct {
	order     uint32
	height    uint32
	rootID    uint32
	nodeCount uint32
}

// readTreeHeader reads the fixed metadata block written by Save/writeEmptyTree.
func readTreeHeader(file *os.File) (treeHeader, error) {
	var h treeHeader
	if err := binary.Read(file, binary.LittleEndian, &h.order); err != nil {
		return treeHeader{}, fmt.Errorf("failed to read order: %w", err)
	}
	if err := binary.Read(file, binary.LittleEndian, &h.height); err != nil {
		return treeHeader{}, fmt.Errorf("failed to read height: %w", err)
	}
	if err := binary.Read(file, binary.LittleEndian, &h.rootID); err != nil {
		return treeHeader{}, fmt.Errorf("failed to read root ID: %w", err)
	}
	if err := binary.Read(file, binary.LittleEndian, &h.nodeCount); err != nil {
		return treeHeader{}, fmt.Errorf("failed to read node count: %w", err)
	}
	return h, nil
}

// readAllTempNodes reads nodeCount serialized nodes from the file into their
// intermediate (pointer-less) representation.
func readAllTempNodes(file *os.File, nodeCount uint32) ([]*tempNode, error) {
	tempNodes := make([]*tempNode, nodeCount)
	for i := uint32(0); i < nodeCount; i++ {
		temp, err := readTempNode(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read node %d: %w", i, err)
		}
		tempNodes[i] = temp
	}
	return tempNodes, nil
}

// buildNodesFromTemp allocates the real *node objects (without cross-node
// pointers wired up yet) from their temp representations.
func buildNodesFromTemp(tempNodes []*tempNode) map[uint32]*node {
	idToNode := make(map[uint32]*node, len(tempNodes))
	for i, temp := range tempNodes {
		idToNode[uint32(i)] = &node{
			isLeaf:   temp.isLeaf,
			keys:     temp.keys,
			children: make([]*node, len(temp.childrenIDs)),
			values:   temp.values,
		}
	}
	return idToNode
}

// linkNodePointers wires up next/children/parent pointers between the real
// nodes based on the IDs recorded on each temp node.
func linkNodePointers(tempNodes []*tempNode, idToNode map[uint32]*node) {
	for _, temp := range tempNodes {
		n := idToNode[temp.id]
		if temp.isLeaf {
			// Reconstruct next pointer
			if temp.nextID != 0 {
				if nextNode, exists := idToNode[temp.nextID]; exists {
					n.next = nextNode
				}
			}
		} else {
			// Reconstruct children pointers
			for j, childID := range temp.childrenIDs {
				if childID != 0 {
					if childNode, exists := idToNode[childID]; exists {
						n.children[j] = childNode
					}
				}
			}
		}

		// Reconstruct parent pointer
		if temp.parentID != 0 {
			if parentNode, exists := idToNode[temp.parentID]; exists {
				n.parent = parentNode
			}
		}
	}
}

// Load deserializes a B+Tree from a binary file.
// Returns a new BPlusTree instance loaded from the file.
func LoadBPlusTree(filename string) (*BPlusTree, error) {
	// Clean the filename to prevent path traversal
	filename = filepath.Clean(filename)
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }() // Ignore error in cleanup

	header, err := readTreeHeader(file)
	if err != nil {
		return nil, err
	}

	// If no nodes, return empty tree
	if header.nodeCount == 0 {
		return NewBPlusTree(int(header.order)), nil
	}

	tempNodes, err := readAllTempNodes(file, header.nodeCount)
	if err != nil {
		return nil, err
	}

	// Convert temp nodes to real nodes and reconstruct pointers
	idToNode := buildNodesFromTemp(tempNodes)
	linkNodePointers(tempNodes, idToNode)

	tree := &BPlusTree{
		root:   idToNode[header.rootID],
		order:  int(header.order),
		height: int(header.height),
	}

	return tree, nil
}

// tempNode holds node data during deserialization
type tempNode struct {
	id          uint32
	isLeaf      bool
	keys        [][]byte
	values      []*ksuid.KSUID
	childrenIDs []uint32
	parentID    uint32
	nextID      uint32
}

// readTempNodeKeys reads the leaf/internal flag, key count, and keys for a
// serialized node.
func readTempNodeKeys(file *os.File) (isLeaf bool, keys [][]byte, keyCount uint32, err error) {
	var isLeafByte uint8
	if err := binary.Read(file, binary.LittleEndian, &isLeafByte); err != nil {
		return false, nil, 0, err
	}

	if err := binary.Read(file, binary.LittleEndian, &keyCount); err != nil {
		return false, nil, 0, err
	}

	keys = make([][]byte, keyCount)
	for i := uint32(0); i < keyCount; i++ {
		var keyLen uint32
		if err := binary.Read(file, binary.LittleEndian, &keyLen); err != nil {
			return false, nil, 0, err
		}
		key := make([]byte, keyLen)
		if _, err := io.ReadFull(file, key); err != nil {
			return false, nil, 0, err
		}
		keys[i] = key
	}

	return isLeafByte == 1, keys, keyCount, nil
}

// readTempLeafData reads a leaf node's values and next-leaf ID.
func readTempLeafData(file *os.File, keyCount uint32) (values []*ksuid.KSUID, nextID uint32, err error) {
	values = make([]*ksuid.KSUID, keyCount)
	for i := uint32(0); i < keyCount; i++ {
		var valueLen uint32
		if err := binary.Read(file, binary.LittleEndian, &valueLen); err != nil {
			return nil, 0, err
		}
		if valueLen == 0 {
			values[i] = nil
			continue
		}
		valueBytes := make([]byte, valueLen)
		if _, err := io.ReadFull(file, valueBytes); err != nil {
			return nil, 0, err
		}
		id, err := ksuid.FromBytes(valueBytes)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid KSUID bytes: %w", err)
		}
		values[i] = &id
	}

	if err := binary.Read(file, binary.LittleEndian, &nextID); err != nil {
		return nil, 0, err
	}
	return values, nextID, nil
}

// readTempInternalChildren reads an internal node's child IDs.
func readTempInternalChildren(file *os.File, keyCount uint32) ([]uint32, error) {
	childrenCount := keyCount + 1
	childrenIDs := make([]uint32, childrenCount)
	for i := uint32(0); i < childrenCount; i++ {
		if err := binary.Read(file, binary.LittleEndian, &childrenIDs[i]); err != nil {
			return nil, err
		}
	}
	return childrenIDs, nil
}

// readTempNode deserializes a single temp node from the file
func readTempNode(file *os.File) (*tempNode, error) {
	isLeaf, keys, keyCount, err := readTempNodeKeys(file)
	if err != nil {
		return nil, err
	}

	temp := &tempNode{
		isLeaf: isLeaf,
		keys:   keys,
	}

	if temp.isLeaf {
		values, nextID, err := readTempLeafData(file, keyCount)
		if err != nil {
			return nil, err
		}
		temp.values = values
		temp.nextID = nextID
	} else {
		childrenIDs, err := readTempInternalChildren(file, keyCount)
		if err != nil {
			return nil, err
		}
		temp.childrenIDs = childrenIDs
	}

	// Read parent ID
	var parentID uint32
	if err := binary.Read(file, binary.LittleEndian, &parentID); err != nil {
		return nil, err
	}
	temp.parentID = parentID

	return temp, nil
}

// Checkpoint starts a background goroutine that periodically saves the B+Tree to the specified file.
// The interval is specified in seconds. Call StopCheckpoint to stop the checkpointing.
func (tree *BPlusTree) StartCheckpoint(filename string, intervalSeconds int) {
	tree.stopCheckpoint() // Stop any existing checkpointing
	tree.checkpointTicker = time.NewTicker(time.Duration(intervalSeconds) * time.Second)
	tree.checkpointDone = make(chan bool)

	go func() {
		for {
			select {
			case <-tree.checkpointTicker.C:
				if err := tree.Save(filename); err != nil {
					// Log error but continue checkpointing
					fmt.Printf("Warning: checkpoint save failed: %v\n", err)
				}
			case <-tree.checkpointDone:
				return
			}
		}
	}()
}

// StopCheckpoint stops the background checkpointing goroutine.
func (tree *BPlusTree) StopCheckpoint() {
	tree.stopCheckpoint()
}

// stopCheckpoint is a helper to stop checkpointing
func (tree *BPlusTree) stopCheckpoint() {
	if tree.checkpointTicker != nil {
		tree.checkpointTicker.Stop()
		tree.checkpointTicker = nil
	}
	if tree.checkpointDone != nil {
		tree.checkpointDone <- true
		tree.checkpointDone = nil
	}
}
