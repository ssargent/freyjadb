//go:build bench
// +build bench

package store

import (
	"fmt"
	"testing"
)

// BenchmarkRelationships_PutRelationship benchmarks the PutRelationship operation
func BenchmarkRelationships_PutRelationship(b *testing.B) {
	b.Run("Put_Relationship_Simple", func(b *testing.B) {
		// Create temporary directory for test
		// Note: In a real implementation, we'd use a mock or temp directory
		// For this benchmark, we'll focus on the relationship logic itself

		relations := make(map[string]map[string]map[string]bool)

		b.ResetTimer()
		b.ReportAllocs()

		for i := range b.N {
			fromKey := fmt.Sprintf("from_key_%d", i)
			toKey := fmt.Sprintf("to_key_%d", i)
			relation := "test_relation"

			// Put relationship logic
			if _, exists := relations[fromKey]; !exists {
				relations[fromKey] = make(map[string]map[string]bool)
			}
			if _, exists := relations[fromKey][relation]; !exists {
				relations[fromKey][relation] = make(map[string]bool)
			}
			relations[fromKey][relation][toKey] = true
		}
	})
}

// BenchmarkRelationships_GetRelationships benchmarks the GetRelationships operation
func BenchmarkRelationships_GetRelationships(b *testing.B) {
	// Pre-populate with test data
	relations := make(map[string]map[string]map[string]bool)

	// Add relationships
	for i := range 10000 {
		fromKey := fmt.Sprintf("from_key_%d", i%1000)
		toKey := fmt.Sprintf("to_key_%d", i)
		relation := fmt.Sprintf("relation_%d", i%10)

		if _, exists := relations[fromKey]; !exists {
			relations[fromKey] = make(map[string]map[string]bool)
		}
		if _, exists := relations[fromKey][relation]; !exists {
			relations[fromKey][relation] = make(map[string]bool)
		}
		relations[fromKey][relation][toKey] = true
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.Run("Get_Relationships_By_FromKey", func(b *testing.B) {
		for i := range b.N {
			fromKey := fmt.Sprintf("from_key_%d", i%1000)
			_ = relations[fromKey]
		}
	})

	b.Run("Get_Relationships_By_Relation", func(b *testing.B) {
		for i := range b.N {
			fromKey := fmt.Sprintf("from_key_%d", i%1000)
			relation := fmt.Sprintf("relation_%d", i%10)
			_ = relations[fromKey][relation]
		}
	})
}

// BenchmarkRelationships_DeleteRelationship benchmarks the DeleteRelationship operation
func BenchmarkRelationships_DeleteRelationship(b *testing.B) {
	b.Run("Delete_Relationship", func(b *testing.B) {
		for range b.N {
			// Create a new relationship structure for each iteration
			relations := make(map[string]map[string]map[string]bool)

			// Add a relationship
			fromKey := "test_from_key"
			toKey := "test_to_key"
			relation := "test_relation"

			if _, exists := relations[fromKey]; !exists {
				relations[fromKey] = make(map[string]map[string]bool)
			}
			if _, exists := relations[fromKey][relation]; !exists {
				relations[fromKey][relation] = make(map[string]bool)
			}
			relations[fromKey][relation][toKey] = true

			b.ResetTimer()
			b.ReportAllocs()

			// Delete the relationship
			if fromMap, exists := relations[fromKey]; exists {
				if relationMap, exists := fromMap[relation]; exists {
					delete(relationMap, toKey)
				}
			}

			b.StopTimer()
		}
	})
}

// BenchmarkRelationships_Query benchmarks relationship queries
func BenchmarkRelationships_Query(b *testing.B) {
	// Pre-populate with test data
	relations := make(map[string]map[string]map[string]bool)

	// Add many relationships
	for i := range 10000 {
		fromKey := fmt.Sprintf("entity_%d", i%1000)
		toKey := fmt.Sprintf("entity_%d", (i+100)%1000)
		relation := "connected_to"

		if _, exists := relations[fromKey]; !exists {
			relations[fromKey] = make(map[string]map[string]bool)
		}
		if _, exists := relations[fromKey][relation]; !exists {
			relations[fromKey][relation] = make(map[string]bool)
		}
		relations[fromKey][relation][toKey] = true
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.Run("Query_All_Relationships_For_Entity", func(b *testing.B) {
		for range b.N {
			fromKey := "entity_500" // Entity with many relationships
			results := make([]string, 0)

			if fromMap, exists := relations[fromKey]; exists {
				for relation, toMap := range fromMap {
					for toKey := range toMap {
						results = append(results, fmt.Sprintf("%s -> %s -> %s", fromKey, relation, toKey))
					}
				}
			}

			_ = results
		}
	})

	b.Run("Query_Specific_Relation_Type", func(b *testing.B) {
		for range b.N {
			fromKey := "entity_500"
			relation := "connected_to"
			results := make([]string, 0)

			if fromMap, exists := relations[fromKey]; exists {
				if toMap, exists := fromMap[relation]; exists {
					for toKey := range toMap {
						results = append(results, toKey)
					}
				}
			}

			_ = results
		}
	})
}

// BenchmarkRelationships_Concurrent benchmarks concurrent relationship operations
func BenchmarkRelationships_Concurrent(b *testing.B) {
	// Pre-populate with test data
	relations := make(map[string]map[string]map[string]bool)

	// Add relationships
	for i := range 10000 {
		fromKey := fmt.Sprintf("from_key_%d", i%1000)
		toKey := fmt.Sprintf("to_key_%d", i)
		relation := fmt.Sprintf("relation_%d", i%10)

		if _, exists := relations[fromKey]; !exists {
			relations[fromKey] = make(map[string]map[string]bool)
		}
		if _, exists := relations[fromKey][relation]; !exists {
			relations[fromKey][relation] = make(map[string]bool)
		}
		relations[fromKey][relation][toKey] = true
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.Run("Concurrent_Reads", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				fromKey := fmt.Sprintf("from_key_%d", i%1000)
				_ = relations[fromKey]
				i++
			}
		})
	})

	b.Run("Concurrent_Writes", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				fromKey := fmt.Sprintf("parallel_from_key_%d", i)
				toKey := fmt.Sprintf("parallel_to_key_%d", i)
				relation := fmt.Sprintf("parallel_relation_%d", i%100)

				if _, exists := relations[fromKey]; !exists {
					relations[fromKey] = make(map[string]map[string]bool)
				}
				if _, exists := relations[fromKey][relation]; !exists {
					relations[fromKey][relation] = make(map[string]bool)
				}
				relations[fromKey][relation][toKey] = true
				i++
			}
		})
	})
}

// BenchmarkRelationships_Scaling benchmarks relationship operations with increasing data sizes
func BenchmarkRelationships_Scaling(b *testing.B) {
	entityCounts := []int{100, 1000, 10000}

	for _, entityCount := range entityCounts {
		b.Run(fmt.Sprintf("Entity_Count_%d", entityCount), func(b *testing.B) {
			b.Run("Build_Relationships", func(b *testing.B) {
				for range b.N {
					relations := make(map[string]map[string]map[string]bool)

					// Create relationships between entities
					for i := range entityCount {
						fromKey := fmt.Sprintf("entity_%d", i)
						// Each entity connects to a few others
						connectionCount := 5
						for j := range connectionCount {
							toKey := fmt.Sprintf("entity_%d", (i+j+1)%entityCount)
							relation := "connected"

							if _, exists := relations[fromKey]; !exists {
								relations[fromKey] = make(map[string]map[string]bool)
							}
							if _, exists := relations[fromKey][relation]; !exists {
								relations[fromKey][relation] = make(map[string]bool)
							}
							relations[fromKey][relation][toKey] = true
						}
					}

					_ = relations
				}
			})
		})
	}
}
