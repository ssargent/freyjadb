package api

//go:generate mockgen -destination=./mock_store.go -package=api . IKVStore

import (
	"context"

	"github.com/ssargent/freyjadb/pkg/store"
)

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// RelationshipRequest represents a relationship creation/deletion request
type RelationshipRequest struct {
	FromKey  string `json:"from_key"`
	ToKey    string `json:"to_key"`
	Relation string `json:"relation"`
}

// ServerConfig holds configuration for the API server
type ServerConfig struct {
	Port                int
	APIKey              string
	SystemKey           string // System API key for administrative operations
	DataDir             string
	SystemDataDir       string // Directory for system KV store
	SystemEncryptionKey string // Encryption key for system data
	EnableEncryption    bool   // Whether to encrypt system data
}

// IKVStore defines the interface for the key-value store operations
type IKVStore interface {
	Put(key, value []byte) error
	Get(key []byte) ([]byte, error)
	Delete(key []byte) error
	ListKeys(prefix []byte) ([]string, error)

	// Relationship methods
	PutRelationship(fromKey, toKey, relation string) error
	DeleteRelationship(fromKey, toKey, relation string) error
	GetRelationships(store.RelationshipQuery) ([]store.RelationshipResult, error)

	// Diagnostics
	Explain(context.Context, store.ExplainOptions) (*store.ExplainResult, error)
	Stats() *store.StoreStats
}
