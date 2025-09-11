package main

import (
	"encoding/json"
	"time"
)

// EntityType represents the type of lore entity
type EntityType string

const (
	EntityTypeCharacter EntityType = "character"
	EntityTypePlace     EntityType = "place"
	EntityTypeGroup     EntityType = "group"
)

// Link represents a relationship between entities
type Link struct {
	Type     EntityType `json:"type"`
	ID       string     `json:"id"`
	Relation string     `json:"relation"`
}

// Entity represents a lore entity with common fields
type Entity struct {
	ID        string     `json:"id"`
	Type      EntityType `json:"type"`
	Name      string     `json:"name"`
	Aka       []string   `json:"aka,omitempty"`
	Summary   string     `json:"summary,omitempty"`
	Details   string     `json:"details,omitempty"`
	Tags      []string   `json:"tags,omitempty"`
	Links     []Link     `json:"links,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// Character represents a character entity
type Character struct {
	Entity
}

// Place represents a place/location entity
type Place struct {
	Entity
}

// Group represents a group/faction/organization entity
type Group struct {
	Entity
}

// NewEntity creates a new entity with the given type and name
func NewEntity(entityType EntityType, name string) *Entity {
	now := time.Now()
	return &Entity{
		Type:      entityType,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Validate checks if the entity is valid
func (e *Entity) Validate() error {
	if e.ID == "" {
		return &LoreError{"entity ID is required"}
	}
	if e.Type == "" {
		return &LoreError{"entity type is required"}
	}
	if e.Name == "" {
		return &LoreError{"entity name is required"}
	}
	return nil
}

// ToJSON converts the entity to JSON bytes
func (e *Entity) ToJSON() ([]byte, error) {
	return json.MarshalIndent(e, "", "  ")
}

// EntityFromJSON creates an entity from JSON bytes
func EntityFromJSON(data []byte) (*Entity, error) {
	var entity Entity
	if err := json.Unmarshal(data, &entity); err != nil {
		return nil, err
	}
	return &entity, nil
}

// LoreError represents a lore-specific error
type LoreError struct {
	Message string
}

func (e *LoreError) Error() string {
	return e.Message
}
