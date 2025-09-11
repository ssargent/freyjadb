package store

import (
	"fmt"
	"strings"
	"time"
)

// Relationship represents a relationship between two entities
type Relationship struct {
	FromKey   string                 `json:"from_key"`
	ToKey     string                 `json:"to_key"`
	Relation  string                 `json:"relation"`
	CreatedAt time.Time              `json:"created_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// RelationshipQuery represents a query for relationships
type RelationshipQuery struct {
	Key       string // Entity key to find relationships for
	Relation  string // Optional: filter by relationship type
	Direction string // "outgoing", "incoming", or "both"
	Limit     int    // Maximum number of results
}

// RelationshipResult represents the result of a relationship query
type RelationshipResult struct {
	Relationship *Relationship `json:"relationship"`
	OtherKey     string        `json:"other_key"`
	Direction    string        `json:"direction"` // "outgoing" or "incoming"
}

// makeRelationshipKey generates a relationship key
// Format: relationship:<direction>:<from_key>:<relation>:<to_key>
// Note: We replace colons in keys with a safe separator to avoid parsing issues
func makeRelationshipKey(direction, fromKey, relation, toKey string) string {
	// Replace colons in keys with a safe separator
	safeFromKey := strings.ReplaceAll(fromKey, ":", "|")
	safeToKey := strings.ReplaceAll(toKey, ":", "|")
	return fmt.Sprintf("relationship:%s:%s:%s:%s", direction, safeFromKey, relation, safeToKey)
}

// parseRelationshipKey extracts components from a relationship key
func parseRelationshipKey(key string) (direction, fromKey, relation, toKey string, err error) {
	parts := strings.Split(key, ":")
	if len(parts) != 5 || parts[0] != "relationship" {
		return "", "", "", "", fmt.Errorf("invalid relationship key format: %s", key)
	}

	direction = parts[1]
	fromKey = strings.ReplaceAll(parts[2], "|", ":") // Restore colons
	relation = parts[3]
	toKey = strings.ReplaceAll(parts[4], "|", ":") // Restore colons
	return
}

// validateRelationshipKeys checks if both keys exist
// Note: This function assumes the caller already holds the mutex
func (kv *KVStore) validateRelationshipKeys(fromKey, toKey string) error {
	if !kv.isOpen {
		return &KVError{"store is not open"}
	}

	// Check if fromKey exists
	_, err := kv.getInternal([]byte(fromKey))
	if err != nil {
		if err == ErrKeyNotFound {
			return fmt.Errorf("source entity does not exist: %s", fromKey)
		}
		return fmt.Errorf("failed to validate source entity: %w", err)
	}

	// Check if toKey exists
	_, err = kv.getInternal([]byte(toKey))
	if err != nil {
		if err == ErrKeyNotFound {
			return fmt.Errorf("target entity does not exist: %s", toKey)
		}
		return fmt.Errorf("failed to validate target entity: %w", err)
	}

	return nil
}
