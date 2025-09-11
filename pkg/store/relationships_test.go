package store

import (
	"os"
	"testing"
	"time"
)

func TestRelationships(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "freyja_relationships_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create KVStore
	config := KVStoreConfig{
		DataDir:       tmpDir,
		FsyncInterval: time.Second,
	}

	kv, err := NewKVStore(config)
	if err != nil {
		t.Fatalf("Failed to create KVStore: %v", err)
	}

	// Open the store
	_, err = kv.Open()
	if err != nil {
		t.Fatalf("Failed to open KVStore: %v", err)
	}
	defer kv.Close()

	// Create test entities
	characterKey := "character:john-doe"
	placeKey := "character:jane-smith"
	placeKey2 := "place:winterfell"

	// Store test entities
	err = kv.Put([]byte(characterKey), []byte(`{"name": "John Doe", "type": "character"}`))
	if err != nil {
		t.Fatalf("Failed to store character: %v", err)
	}

	err = kv.Put([]byte(placeKey), []byte(`{"name": "Jane Smith", "type": "character"}`))
	if err != nil {
		t.Fatalf("Failed to store character 2: %v", err)
	}

	err = kv.Put([]byte(placeKey2), []byte(`{"name": "Winterfell", "type": "place"}`))
	if err != nil {
		t.Fatalf("Failed to store place: %v", err)
	}

	// Test creating relationships
	t.Run("PutRelationship", func(t *testing.T) {
		// Create friendship relationship
		err := kv.PutRelationship(characterKey, placeKey, "friend")
		if err != nil {
			t.Fatalf("Failed to create relationship: %v", err)
		}

		// Create location relationship
		err = kv.PutRelationship(characterKey, placeKey2, "located_in")
		if err != nil {
			t.Fatalf("Failed to create location relationship: %v", err)
		}
	})

	// Test querying relationships
	t.Run("GetRelationships", func(t *testing.T) {
		// Query outgoing relationships for John Doe
		query := RelationshipQuery{
			Key:       characterKey,
			Direction: "outgoing",
			Limit:     10,
		}

		results, err := kv.GetRelationships(query)
		if err != nil {
			t.Fatalf("Failed to get relationships: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 relationships, got %d", len(results))
		}

		// Check that we have both relationships
		foundFriend := false
		foundLocation := false
		for _, result := range results {
			if result.Relationship.Relation == "friend" && result.OtherKey == placeKey {
				foundFriend = true
			}
			if result.Relationship.Relation == "located_in" && result.OtherKey == placeKey2 {
				foundLocation = true
			}
		}

		if !foundFriend {
			t.Error("Friend relationship not found")
		}
		if !foundLocation {
			t.Error("Location relationship not found")
		}
	})

	// Test querying incoming relationships
	t.Run("GetIncomingRelationships", func(t *testing.T) {
		// Query incoming relationships for Winterfell
		query := RelationshipQuery{
			Key:       placeKey2,
			Direction: "incoming",
			Limit:     10,
		}

		results, err := kv.GetRelationships(query)
		if err != nil {
			t.Fatalf("Failed to get incoming relationships: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 incoming relationship, got %d", len(results))
		}

		if results[0].Relationship.Relation != "located_in" {
			t.Errorf("Expected 'located_in' relationship, got '%s'", results[0].Relationship.Relation)
		}
	})

	// Test deleting relationships
	t.Run("DeleteRelationship", func(t *testing.T) {
		// Delete the friendship relationship
		err := kv.DeleteRelationship(characterKey, placeKey, "friend")
		if err != nil {
			t.Fatalf("Failed to delete relationship: %v", err)
		}

		// Verify it's gone
		query := RelationshipQuery{
			Key:       characterKey,
			Direction: "outgoing",
			Relation:  "friend",
			Limit:     10,
		}

		results, err := kv.GetRelationships(query)
		if err != nil {
			t.Fatalf("Failed to query after deletion: %v", err)
		}

		if len(results) != 0 {
			t.Errorf("Expected 0 friend relationships after deletion, got %d", len(results))
		}
	})

	// Test relationship validation
	t.Run("RelationshipValidation", func(t *testing.T) {
		// Try to create relationship with non-existent entity
		err := kv.PutRelationship("nonexistent:key", characterKey, "test")
		if err == nil {
			t.Error("Expected error when creating relationship with non-existent entity")
		}
	})
}

func TestRelationshipKeyGeneration(t *testing.T) {
	fromKey := "character:john"
	toKey := "place:winterfell"
	relation := "located_in"

	forwardKey := makeRelationshipKey("forward", fromKey, relation, toKey)
	expectedForward := "relationship:forward:character|john:located_in:place|winterfell"

	if forwardKey != expectedForward {
		t.Errorf("Expected forward key '%s', got '%s'", expectedForward, forwardKey)
	}

	reverseKey := makeRelationshipKey("reverse", toKey, relation, fromKey)
	expectedReverse := "relationship:reverse:place|winterfell:located_in:character|john"

	if reverseKey != expectedReverse {
		t.Errorf("Expected reverse key '%s', got '%s'", expectedReverse, reverseKey)
	}

	// Test parsing
	direction, parsedFrom, parsedRelation, parsedTo, err := parseRelationshipKey(forwardKey)
	if err != nil {
		t.Fatalf("Failed to parse relationship key: %v", err)
	}

	if direction != "forward" || parsedFrom != fromKey || parsedRelation != relation || parsedTo != toKey {
		t.Errorf("Parsed values don't match: direction=%s, from=%s, relation=%s, to=%s",
			direction, parsedFrom, parsedRelation, parsedTo)
	}
}
