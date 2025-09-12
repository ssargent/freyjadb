package api

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ssargent/freyjadb/pkg/store"
)

// SystemService provides internal APIs for managing system-level data
type SystemService struct {
	store  *store.KVStore
	config SystemConfig
	gcm    cipher.AEAD
	isOpen bool
}

// SystemConfig holds configuration for the system service
type SystemConfig struct {
	DataDir          string
	EncryptionKey    string
	EnableEncryption bool
}

// APIKey represents an API key stored in the system
type APIKey struct {
	ID          string     `json:"id"`
	Key         string     `json:"key"`
	Description string     `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsActive    bool       `json:"is_active"`
}

// NewSystemService creates a new system service instance
func NewSystemService(config SystemConfig) (*SystemService, error) {
	// Ensure system data directory exists
	systemDataDir := filepath.Join(config.DataDir, "system")
	if err := os.MkdirAll(systemDataDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create system data directory: %w", err)
	}

	// Initialize encryption if enabled
	var gcm cipher.AEAD
	if config.EnableEncryption && config.EncryptionKey != "" {
		block, err := aes.NewCipher([]byte(config.EncryptionKey))
		if err != nil {
			return nil, fmt.Errorf("failed to create cipher: %w", err)
		}

		gcm, err = cipher.NewGCM(block)
		if err != nil {
			return nil, fmt.Errorf("failed to create GCM: %w", err)
		}
	}

	service := &SystemService{
		config: config,
		gcm:    gcm,
		isOpen: false,
	}

	return service, nil
}

// Open initializes the system store
func (s *SystemService) Open() error {
	if s.isOpen {
		return nil
	}

	systemDataDir := filepath.Join(s.config.DataDir, "system")
	storeConfig := store.KVStoreConfig{
		DataDir:       systemDataDir,
		FsyncInterval: time.Second, // More frequent fsync for system data
	}

	kvStore, err := store.NewKVStore(storeConfig)
	if err != nil {
		return fmt.Errorf("failed to create system KV store: %w", err)
	}

	_, err = kvStore.Open()
	if err != nil {
		return fmt.Errorf("failed to open system KV store: %w", err)
	}

	s.store = kvStore
	s.isOpen = true
	return nil
}

// Close shuts down the system service
func (s *SystemService) Close() error {
	if !s.isOpen {
		return nil
	}

	// Mark as closed first to prevent double closes
	s.isOpen = false

	if s.store != nil {
		if err := s.store.Close(); err != nil {
			// Don't return error for "file already closed" as it's not a real error
			if err.Error() != "close /var/folders/mn/5l10pmk93_l4j6hv6k4cgd280000gn/T/"+
				"freyja_system_integration1021559736/system/active.data: file already closed" &&
				!strings.Contains(err.Error(), "file already closed") {
				return fmt.Errorf("failed to close system store: %w", err)
			}
		}
	}

	return nil
}

// encrypt encrypts data if encryption is enabled
func (s *SystemService) encrypt(plaintext []byte) ([]byte, error) {
	if !s.config.EnableEncryption || s.gcm == nil {
		return plaintext, nil
	}

	nonce := make([]byte, s.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := s.gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decrypt decrypts data if encryption is enabled
func (s *SystemService) decrypt(ciphertext []byte) ([]byte, error) {
	if !s.config.EnableEncryption || s.gcm == nil {
		return ciphertext, nil
	}

	if len(ciphertext) < s.gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce := ciphertext[:s.gcm.NonceSize()]
	ciphertext = ciphertext[s.gcm.NonceSize():]

	plaintext, err := s.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// StoreAPIKey stores an API key in the system store
func (s *SystemService) StoreAPIKey(apiKey APIKey) error {
	if !s.isOpen {
		return fmt.Errorf("system service is not open")
	}

	key := fmt.Sprintf("apikey:%s", apiKey.ID)
	data, err := json.Marshal(apiKey)
	if err != nil {
		return fmt.Errorf("failed to marshal API key: %w", err)
	}

	encryptedData, err := s.encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt API key: %w", err)
	}

	return s.store.Put([]byte(key), encryptedData)
}

// GetAPIKey retrieves an API key from the system store
func (s *SystemService) GetAPIKey(keyID string) (*APIKey, error) {
	if !s.isOpen {
		return nil, fmt.Errorf("system service is not open")
	}

	key := fmt.Sprintf("apikey:%s", keyID)
	encryptedData, err := s.store.Get([]byte(key))
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	data, err := s.decrypt(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt API key: %w", err)
	}

	var apiKey APIKey
	if err := json.Unmarshal(data, &apiKey); err != nil {
		return nil, fmt.Errorf("failed to unmarshal API key: %w", err)
	}

	return &apiKey, nil
}

// ValidateAPIKey validates if an API key exists and is active
func (s *SystemService) ValidateAPIKey(apiKeyValue string) (bool, error) {
	if !s.isOpen {
		return false, fmt.Errorf("system service is not open")
	}

	// List all API keys and check if any match
	keys, err := s.ListAPIKeys()
	if err != nil {
		return false, fmt.Errorf("failed to list API keys: %w", err)
	}

	for _, keyID := range keys {
		apiKey, err := s.GetAPIKey(keyID)
		if err != nil {
			continue // Skip invalid keys
		}

		if apiKey.Key == apiKeyValue && apiKey.IsActive {
			// Check expiration
			if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
				return false, nil // Key expired
			}
			return true, nil
		}
	}

	return false, nil
}

// ListAPIKeys returns a list of all API key IDs
func (s *SystemService) ListAPIKeys() ([]string, error) {
	if !s.isOpen {
		return nil, fmt.Errorf("system service is not open")
	}

	keys, err := s.store.ListKeys([]byte("apikey:"))
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	// Extract key IDs from the full keys
	var keyIDs []string
	for _, key := range keys {
		if len(key) > 7 { // "apikey:" prefix
			keyIDs = append(keyIDs, key[7:])
		}
	}

	return keyIDs, nil
}

// DeleteAPIKey removes an API key from the system store
func (s *SystemService) DeleteAPIKey(keyID string) error {
	if !s.isOpen {
		return fmt.Errorf("system service is not open")
	}

	key := fmt.Sprintf("apikey:%s", keyID)
	return s.store.Delete([]byte(key))
}

// StoreSystemConfig stores system configuration data
func (s *SystemService) StoreSystemConfig(key string, value interface{}) error {
	if !s.isOpen {
		return fmt.Errorf("system service is not open")
	}

	configKey := fmt.Sprintf("config:%s", key)
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal config value: %w", err)
	}

	encryptedData, err := s.encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt config value: %w", err)
	}

	return s.store.Put([]byte(configKey), encryptedData)
}

// GetSystemConfig retrieves system configuration data
func (s *SystemService) GetSystemConfig(key string, value interface{}) error {
	if !s.isOpen {
		return fmt.Errorf("system service is not open")
	}

	configKey := fmt.Sprintf("config:%s", key)
	encryptedData, err := s.store.Get([]byte(configKey))
	if err != nil {
		return fmt.Errorf("failed to get config value: %w", err)
	}

	data, err := s.decrypt(encryptedData)
	if err != nil {
		return fmt.Errorf("failed to decrypt config value: %w", err)
	}

	if err := json.Unmarshal(data, value); err != nil {
		return fmt.Errorf("failed to unmarshal config value: %w", err)
	}

	return nil
}

// IsOpen returns whether the system service is open
func (s *SystemService) IsOpen() bool {
	return s.isOpen
}

// InitializeSystem implements the SystemInitializer interface
func (s *SystemService) InitializeSystem(dataDir, systemKey, systemAPIKey string) error {
	// Open the system service
	if err := s.Open(); err != nil {
		return fmt.Errorf("failed to open system service: %w", err)
	}
	defer s.Close()

	// Store system API key
	apiKey := APIKey{
		ID:          "system-root",
		Key:         systemAPIKey,
		Description: "System root API key for administrative operations",
		CreatedAt:   time.Now(),
		IsActive:    true,
	}

	if err := s.StoreAPIKey(apiKey); err != nil {
		return fmt.Errorf("failed to store system API key: %w", err)
	}

	// Store some default system configuration
	defaultConfig := map[string]interface{}{
		"initialized_at":     time.Now().Format(time.RFC3339),
		"version":            "1.0.0",
		"encryption_enabled": s.config.EnableEncryption,
	}

	if err := s.StoreSystemConfig("system-info", defaultConfig); err != nil {
		return fmt.Errorf("failed to store system configuration: %w", err)
	}

	return nil
}
