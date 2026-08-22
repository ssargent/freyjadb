package store

import (
	"os"
	"testing"
)

const relationLocatedIn = "located_in"

// relationshipTestFixture holds the store and entity keys shared by TestRelationships' subtests.
type relationshipTestFixture struct {
	kv           *KVStore
	characterKey string
	placeKey     string
	placeKey2    string
}

func newRelationshipTestFixture(t *testing.T) *relationshipTestFixture {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "freyja_relationships_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	config := KVStoreConfig{
		DataDir:       tmpDir,
		FsyncInterval: 0, // Immediate sync for testing
	}

	kv, err := NewKVStore(config)
	if err != nil {
		t.Fatalf("Failed to create KVStore: %v", err)
	}

	if _, err := kv.Open(); err != nil {
		t.Fatalf("Failed to open KVStore: %v", err)
	}
	t.Cleanup(func() { _ = kv.Close() })

	fixture := &relationshipTestFixture{
		kv:           kv,
		characterKey: "character:john-doe",
		placeKey:     "character:jane-smith",
		placeKey2:    "place:winterfell",
	}

	if err := kv.Put([]byte(fixture.characterKey), []byte(`{"name": "John Doe", "type": "character"}`)); err != nil {
		t.Fatalf("Failed to store character: %v", err)
	}
	if err := kv.Put([]byte(fixture.placeKey), []byte(`{"name": "Jane Smith", "type": "character"}`)); err != nil {
		t.Fatalf("Failed to store character 2: %v", err)
	}
	if err := kv.Put([]byte(fixture.placeKey2), []byte(`{"name": "Winterfell", "type": "place"}`)); err != nil {
		t.Fatalf("Failed to store place: %v", err)
	}

	return fixture
}

func TestRelationships(t *testing.T) {
	f := newRelationshipTestFixture(t)

	t.Run("PutRelationship", f.testPutRelationship)
	t.Run("GetRelationships", f.testGetRelationships)
	t.Run("GetIncomingRelationships", f.testGetIncomingRelationships)
	t.Run("DeleteRelationship", f.testDeleteRelationship)
	t.Run("RelationshipValidation", f.testRelationshipValidation)
}

func (f *relationshipTestFixture) testPutRelationship(t *testing.T) {
	if err := f.kv.PutRelationship(f.characterKey, f.placeKey, "friend"); err != nil {
		t.Fatalf("Failed to create relationship: %v", err)
	}

	if err := f.kv.PutRelationship(f.characterKey, f.placeKey2, relationLocatedIn); err != nil {
		t.Fatalf("Failed to create location relationship: %v", err)
	}
}

func (f *relationshipTestFixture) testGetRelationships(t *testing.T) {
	query := RelationshipQuery{
		Key:       f.characterKey,
		Direction: relationDirectionOutgoing,
		Limit:     10,
	}

	results, err := f.kv.GetRelationships(query)
	if err != nil {
		t.Fatalf("Failed to get relationships: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 relationships, got %d", len(results))
	}

	foundFriend := false
	foundLocation := false
	for _, result := range results {
		if result.Relationship.Relation == "friend" && result.OtherKey == f.placeKey {
			foundFriend = true
		}
		if result.Relationship.Relation == relationLocatedIn && result.OtherKey == f.placeKey2 {
			foundLocation = true
		}
	}

	if !foundFriend {
		t.Error("Friend relationship not found")
	}
	if !foundLocation {
		t.Error("Location relationship not found")
	}
}

func (f *relationshipTestFixture) testGetIncomingRelationships(t *testing.T) {
	query := RelationshipQuery{
		Key:       f.placeKey2,
		Direction: relationDirectionIncoming,
		Limit:     10,
	}

	results, err := f.kv.GetRelationships(query)
	if err != nil {
		t.Fatalf("Failed to get incoming relationships: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 incoming relationship, got %d", len(results))
	}

	if results[0].Relationship.Relation != relationLocatedIn {
		t.Errorf("Expected '%s' relationship, got '%s'", relationLocatedIn, results[0].Relationship.Relation)
	}
}

func (f *relationshipTestFixture) testDeleteRelationship(t *testing.T) {
	if err := f.kv.DeleteRelationship(f.characterKey, f.placeKey, "friend"); err != nil {
		t.Fatalf("Failed to delete relationship: %v", err)
	}

	query := RelationshipQuery{
		Key:       f.characterKey,
		Direction: relationDirectionOutgoing,
		Relation:  "friend",
		Limit:     10,
	}

	results, err := f.kv.GetRelationships(query)
	if err != nil {
		t.Fatalf("Failed to query after deletion: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 friend relationships after deletion, got %d", len(results))
	}
}

func (f *relationshipTestFixture) testRelationshipValidation(t *testing.T) {
	if err := f.kv.PutRelationship("nonexistent:key", f.characterKey, "test"); err == nil {
		t.Error("Expected error when creating relationship with non-existent entity")
	}
}

func TestRelationshipKeyGeneration(t *testing.T) {
	fromKey := "character:john"
	toKey := "place:winterfell"
	relation := relationLocatedIn

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
