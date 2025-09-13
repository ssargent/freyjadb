// Package api provides interfaces for dependency injection
package api

import "github.com/ssargent/freyjadb/pkg/store"

// SystemInitializer defines the interface for system initialization operations
type SystemInitializer interface {
	// InitializeSystem sets up the system with the given configuration
	InitializeSystem(dataDir, systemKey, systemAPIKey string) error

	// Open initializes the system service
	Open() error

	// Close cleans up system resources
	Close() error

	// GetAPIKey retrieves an API key
	GetAPIKey(keyID string) (*APIKey, error)
}

// SystemServiceFactory creates system services
type SystemServiceFactory interface {
	// CreateSystemService creates a new system service with the given config
	CreateSystemService(dataDir, encryptionKey string, enableEncryption bool, maxRecordSize int) (SystemInitializer, error)
}

// ServerStarter defines the interface for starting the API server
type ServerStarter interface {
	// StartServer starts the API server with the given configuration
	StartServer(kvStore *store.KVStore,
		port int,
		apiKey, systemKey, dataDir, systemEncryptionKey string,
		enableEncryption bool,
	) error
}

// ServerFactory creates server instances
type ServerFactory interface {
	// CreateServerStarter creates a server starter
	CreateServerStarter() ServerStarter
}
