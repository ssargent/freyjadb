//go:build bench
// +build bench

package bptree

import (
	"fmt"
	"github.com/segmentio/ksuid"
	"testing"
)

// BenchmarkBPTree_Insert benchmarks the Insert operation with various tree orders
func BenchmarkBPTree_Insert(b *testing.B) {
	orders := []int{4, 16, 64, 256}

	for _, order := range orders {
		b.Run(fmt.Sprintf("Order_%d", order), func(b *testing.B) {
			tree := NewBPlusTree(order)

			b.ResetTimer()
			b.ReportAllocs()

			for i := range b.N {
				key := []byte(fmt.Sprintf("key_%010d", i))
				value := ksuid.New()
				tree.Insert(key, value)
			}
		})
	}
}

// BenchmarkBPTree_Search benchmarks the Search operation
func BenchmarkBPTree_Search(b *testing.B) {
	// Create a tree and populate it with data
	tree := NewBPlusTree(16)

	// Pre-populate with test data
	for i := range 10000 {
		key := []byte(fmt.Sprintf("key_%010d", i))
		value := ksuid.New()
		tree.Insert(key, value)
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.Run("Existing_Key", func(b *testing.B) {
		for i := range b.N {
			key := []byte(fmt.Sprintf("key_%010d", i%10000))
			tree.Search(key)
		}
	})

	b.Run("Non_Existent_Key", func(b *testing.B) {
		for i := range b.N {
			key := []byte(fmt.Sprintf("nonexistent_key_%010d", i))
			tree.Search(key)
		}
	})
}

// BenchmarkBPTree_Delete benchmarks the Delete operation
func BenchmarkBPTree_Delete(b *testing.B) {
	b.Run("Delete_From_Full_Tree", func(b *testing.B) {
		for range b.N {
			// Create a new tree for each iteration
			tree := NewBPlusTree(16)

			// Pre-populate with test data
			for i := range 1000 {
				key := []byte(fmt.Sprintf("key_%010d", i))
				value := ksuid.New()
				tree.Insert(key, value)
			}

			b.ResetTimer()

			// Delete keys
			for i := range 1000 {
				key := []byte(fmt.Sprintf("key_%010d", i))
				tree.Delete(key)
			}

			b.StopTimer()
		}
	})
}

// BenchmarkBPTree_Concurrent benchmarks concurrent operations
func BenchmarkBPTree_Concurrent(b *testing.B) {
	// Create a tree and populate it with data
	tree := NewBPlusTree(16)

	// Pre-populate with test data
	for i := range 10000 {
		key := []byte(fmt.Sprintf("key_%010d", i))
		value := ksuid.New()
		tree.Insert(key, value)
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.Run("Concurrent_Reads", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := []byte(fmt.Sprintf("key_%010d", i%10000))
				tree.Search(key)
				i++
			}
		})
	})

	b.Run("Concurrent_Writes", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := []byte(fmt.Sprintf("parallel_key_%010d_%d", i, b.N))
				value := ksuid.New()
				tree.Insert(key, value)
				i++
			}
		})
	})
}

// BenchmarkBPTree_Height benchmarks tree height with increasing data sizes
func BenchmarkBPTree_Height(b *testing.B) {
	sizes := []int{1000, 10000, 100000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%d_Items", size), func(b *testing.B) {
			for range b.N {
				tree := NewBPlusTree(16)

				for i := range size {
					key := []byte(fmt.Sprintf("key_%010d", i))
					value := ksuid.New()
					tree.Insert(key, value)
				}

				// Measure height
				_ = tree.Height()
			}
		})
	}
}
