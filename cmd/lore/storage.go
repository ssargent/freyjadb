package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ssargent/freyjadb/pkg/store"
)

// LoreStore manages persistence of lore entities using FreyjaDB
type LoreStore struct {
	kvStore *store.KVStore
	isOpen  bool
}

// NewLoreStore creates a new lore store
func NewLoreStore(projectDir string) (*LoreStore, error) {
	dataDir := filepath.Join(projectDir, ".lore")

	config := store.KVStoreConfig{
		DataDir:       dataDir,
		FsyncInterval: time.Second, // fsync every second for durability
	}

	kvStore, err := store.NewKVStore(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create KV store: %w", err)
	}

	return &LoreStore{
		kvStore: kvStore,
		isOpen:  false,
	}, nil
}

// Open initializes the store
func (ls *LoreStore) Open() error {
	if ls.isOpen {
		return nil
	}

	_, err := ls.kvStore.Open()
	if err != nil {
		return fmt.Errorf("failed to open KV store: %w", err)
	}

	ls.isOpen = true
	return nil
}

// Close closes the store
func (ls *LoreStore) Close() error {
	if !ls.isOpen {
		return nil
	}

	ls.isOpen = false
	return ls.kvStore.Close()
}

// makeKey creates a storage key for an entity
func makeKey(entityType EntityType, id string) []byte {
	return []byte(fmt.Sprintf("%s:%s", entityType, id))
}

// generateID creates a slug-like ID from a name
func generateID(name string) string {
	// Convert to lowercase
	id := strings.ToLower(name)

	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	id = reg.ReplaceAllString(id, "-")

	// Remove leading/trailing hyphens
	id = strings.Trim(id, "-")

	// Ensure it's not empty
	if id == "" {
		id = "unnamed"
	}

	return id
}

// PutEntity stores an entity
func (ls *LoreStore) PutEntity(entity *Entity) error {
	if !ls.isOpen {
		return fmt.Errorf("store is not open")
	}

	// Generate ID if not provided
	if entity.ID == "" {
		entity.ID = generateID(entity.Name)
	}

	if err := entity.Validate(); err != nil {
		return err
	}

	// Update timestamp
	entity.UpdatedAt = time.Now()

	key := makeKey(entity.Type, entity.ID)
	data, err := entity.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize entity: %w", err)
	}

	return ls.kvStore.Put(key, data)
}

// GetEntity retrieves an entity by type and ID
func (ls *LoreStore) GetEntity(entityType EntityType, id string) (*Entity, error) {
	if !ls.isOpen {
		return nil, fmt.Errorf("store is not open")
	}

	key := makeKey(entityType, id)
	data, err := ls.kvStore.Get(key)
	if err != nil {
		if err == store.ErrKeyNotFound {
			return nil, &LoreError{fmt.Sprintf("%s '%s' not found", entityType, id)}
		}
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	entity, err := EntityFromJSON(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize entity: %w", err)
	}

	return entity, nil
}

// DeleteEntity removes an entity
func (ls *LoreStore) DeleteEntity(entityType EntityType, id string) error {
	if !ls.isOpen {
		return fmt.Errorf("store is not open")
	}

	key := makeKey(entityType, id)

	// Check if entity exists first
	_, err := ls.kvStore.Get(key)
	if err != nil {
		if err == store.ErrKeyNotFound {
			return &LoreError{fmt.Sprintf("%s '%s' not found", entityType, id)}
		}
		return fmt.Errorf("failed to check entity existence: %w", err)
	}

	return ls.kvStore.Delete(key)
}

// ListEntities returns all entities of a given type
func (ls *LoreStore) ListEntities(entityType EntityType) ([]*Entity, error) {
	if !ls.isOpen {
		return nil, fmt.Errorf("store is not open")
	}

	prefix := fmt.Sprintf("%s:", entityType)
	keys, err := ls.kvStore.ListKeys([]byte(prefix))
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}

	var entities []*Entity
	for _, key := range keys {
		value, err := ls.kvStore.Get([]byte(key))
		if err != nil {
			continue // Skip keys that can't be read
		}

		entity, err := EntityFromJSON(value)
		if err != nil {
			continue // Skip corrupted entities
		}

		entities = append(entities, entity)
	}

	return entities, nil
}

// EntityExists checks if an entity exists
func (ls *LoreStore) EntityExists(entityType EntityType, id string) bool {
	if !ls.isOpen {
		return false
	}

	key := makeKey(entityType, id)
	_, err := ls.kvStore.Get(key)
	return err == nil
}

// PutRelationship creates a relationship between two Lore entities
func (ls *LoreStore) PutRelationship(fromType EntityType, fromID string,
	toType EntityType, toID string, relation string) error {
	if !ls.isOpen {
		return fmt.Errorf("store is not open")
	}

	fromKey := string(makeKey(fromType, fromID))
	toKey := string(makeKey(toType, toID))

	return ls.kvStore.PutRelationship(fromKey, toKey, relation)
}

// DeleteRelationship removes a relationship between two Lore entities
func (ls *LoreStore) DeleteRelationship(fromType EntityType, fromID string,
	toType EntityType, toID string, relation string) error {
	if !ls.isOpen {
		return fmt.Errorf("store is not open")
	}

	fromKey := string(makeKey(fromType, fromID))
	toKey := string(makeKey(toType, toID))

	return ls.kvStore.DeleteRelationship(fromKey, toKey, relation)
}

// GetEntityRelationships returns all relationships for a given entity
func (ls *LoreStore) GetEntityRelationships(entityType EntityType, id string,
	direction string, relation string) ([]store.RelationshipResult, error) {
	if !ls.isOpen {
		return nil, fmt.Errorf("store is not open")
	}

	key := string(makeKey(entityType, id))

	query := store.RelationshipQuery{
		Key:       key,
		Relation:  relation,
		Direction: direction,
		Limit:     100, // Reasonable default
	}

	return ls.kvStore.GetRelationships(query)
}

// GetEntityWithRelationships returns an entity along with its relationships
func (ls *LoreStore) GetEntityWithRelationships(entityType EntityType, id string) (*EntityWithRelationships, error) {
	if !ls.isOpen {
		return nil, fmt.Errorf("store is not open")
	}

	// Get the entity
	entity, err := ls.GetEntity(entityType, id)
	if err != nil {
		return nil, err
	}

	// Get all relationships
	outgoing, err := ls.GetEntityRelationships(entityType, id, "outgoing", "")
	if err != nil {
		return nil, err
	}

	incoming, err := ls.GetEntityRelationships(entityType, id, "incoming", "")
	if err != nil {
		return nil, err
	}

	return &EntityWithRelationships{
		Entity:   entity,
		Outgoing: outgoing,
		Incoming: incoming,
	}, nil
}

// EntityWithRelationships represents an entity with its relationship data
type EntityWithRelationships struct {
	Entity   *Entity                    `json:"entity"`
	Outgoing []store.RelationshipResult `json:"outgoing_relationships"`
	Incoming []store.RelationshipResult `json:"incoming_relationships"`
}
