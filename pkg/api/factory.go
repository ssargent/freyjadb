// Package api provides factory implementations for dependency injection
package api

import (
	"github.com/ssargent/freyjadb/pkg/store"
)

// DefaultSystemServiceFactory is the default implementation of SystemServiceFactory
type DefaultSystemServiceFactory struct{}

// NewSystemServiceFactory creates a new system service factory
func NewSystemServiceFactory() SystemServiceFactory {
	return &DefaultSystemServiceFactory{}
}

// DefaultServerFactory is the default implementation of ServerFactory
type DefaultServerFactory struct{}

// NewServerFactory creates a new server factory
func NewServerFactory() ServerFactory {
	return &DefaultServerFactory{}
}

// CreateServerStarter creates a server starter
func (f *DefaultServerFactory) CreateServerStarter() ServerStarter {
	return &DefaultServerStarter{}
}

// DefaultServerStarter is the default implementation of ServerStarter
type DefaultServerStarter struct{}

// StartServer starts the API server with the given configuration
func (s *DefaultServerStarter) StartServer(
	kvStore *store.KVStore,
	port int,
	apiKey, systemKey, dataDir, systemEncryptionKey string,
	enableEncryption bool,
) error {
	config := ServerConfig{
		Port:                port,
		APIKey:              apiKey,
		SystemKey:           systemKey,
		DataDir:             dataDir,
		SystemDataDir:       dataDir,
		SystemEncryptionKey: systemEncryptionKey,
		EnableEncryption:    enableEncryption,
	}
	return StartServer(kvStore, config)
}

// CreateSystemService creates a new system service with the given config
func (f *DefaultSystemServiceFactory) CreateSystemService(
	dataDir, encryptionKey string,
	enableEncryption bool,
	maxRecordSize int,
) (SystemInitializer, error) {
	config := SystemConfig{
		DataDir:          dataDir,
		EncryptionKey:    encryptionKey,
		EnableEncryption: enableEncryption,
		MaxRecordSize:    maxRecordSize,
	}
	return NewSystemService(config)
}
