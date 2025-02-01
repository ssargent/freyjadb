package bptree_test

import (
	"sync"
	"testing"

	"github.com/ssargent/freyjadb/pkg/bptree"
)

func TestBPlusTree_InsertAndSearch(t *testing.T) {
	tests := map[string]struct {
		tree     *bptree.BPlusTree[int, string]
		actions  []func(tree *bptree.BPlusTree[int, string])
		searches []struct {
			key      int
			expected string
			found    bool
		}
	}{
		"Insert and search integers": {
			tree: bptree.NewBPlusTree[int, string](4),
			actions: []func(tree *bptree.BPlusTree[int, string]){
				func(tree *bptree.BPlusTree[int, string]) { tree.Insert(1, "one") },
				func(tree *bptree.BPlusTree[int, string]) { tree.Insert(2, "two") },
				func(tree *bptree.BPlusTree[int, string]) { tree.Insert(3, "three") },
				func(tree *bptree.BPlusTree[int, string]) { tree.Insert(4, "four") },
				func(tree *bptree.BPlusTree[int, string]) { tree.Insert(5, "five") },
			},
			searches: []struct {
				key      int
				expected string
				found    bool
			}{
				{1, "one", true},
				{2, "two", true},
				{3, "three", true},
				{4, "four", true},
				{5, "five", true},
				{6, "", false},
			},
		},
		"Insert duplicate keys": {
			tree: bptree.NewBPlusTree[int, string](4),
			actions: []func(tree *bptree.BPlusTree[int, string]){
				func(tree *bptree.BPlusTree[int, string]) { tree.Insert(1, "one") },
				func(tree *bptree.BPlusTree[int, string]) { tree.Insert(1, "uno") },
			},
			searches: []struct {
				key      int
				expected string
				found    bool
			}{
				{1, "uno", true},
			},
		},
		"Search empty tree": {
			tree:    bptree.NewBPlusTree[int, string](4),
			actions: []func(tree *bptree.BPlusTree[int, string]){},
			searches: []struct {
				key      int
				expected string
				found    bool
			}{
				{1, "", false},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			for _, action := range tt.actions {
				action(tt.tree)
			}
			for _, search := range tt.searches {
				value, found := tt.tree.Search(search.key)
				if found != search.found || value != search.expected {
					t.Errorf("Search(%d) = %v, %v; want %v, %v", search.key, value, found, search.expected, search.found)
				}
			}
		})
	}
}

func TestBPlusTree_Concurrency(t *testing.T) {
	tree := bptree.NewBPlusTree[int, string](4)

	// Insert keys concurrently
	var wg sync.WaitGroup
	for i := 1; i <= 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			tree.Insert(i, string(rune('a'+i-1)))
		}(i)
	}
	wg.Wait()

	// Search for keys concurrently
	for i := 1; i <= 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if _, found := tree.Search(i); !found {
				t.Errorf("Expected to find key %d", i)
			}
		}(i)
	}
	wg.Wait()
}
