package api

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSystemService(t *testing.T) {
	t.Run("NewSystemService", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "freyja_system_test_new")
		assert.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		config := SystemConfig{
			DataDir:          tmpDir,
			EncryptionKey:    "",
			EnableEncryption: false,
		}

		service, err := NewSystemService(config)
		assert.NoError(t, err)
		assert.NotNil(t, service)
		assert.False(t, service.IsOpen())
	})

	t.Run("Open and Close", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "freyja_system_test_open")
		assert.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		config := SystemConfig{
			DataDir:          tmpDir,
			EncryptionKey:    "",
			EnableEncryption: false,
		}

		service, err := NewSystemService(config)
		assert.NoError(t, err)

		// Open service
		err = service.Open()
		assert.NoError(t, err)
		assert.True(t, service.IsOpen())

		// Close service
		err = service.Close()
		assert.NoError(t, err)
		assert.False(t, service.IsOpen())
	})

	t.Run("API Key Management", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "freyja_system_test_api")
		assert.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		config := SystemConfig{
			DataDir:          tmpDir,
			EncryptionKey:    "12345678901234567890123456789012", // 32 bytes for AES-256
			EnableEncryption: true,
		}

		service, err := NewSystemService(config)
		assert.NoError(t, err)
		defer service.Close()

		err = service.Open()
		assert.NoError(t, err)

		// Create API key
		apiKey := APIKey{
			ID:          "test-key-1",
			Key:         "secret123",
			Description: "Test API key",
			CreatedAt:   time.Now(),
			IsActive:    true,
		}

		// Store API key
		err = service.StoreAPIKey(apiKey)
		assert.NoError(t, err)

		// Retrieve API key
		retrieved, err := service.GetAPIKey("test-key-1")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, "test-key-1", retrieved.ID)
		assert.Equal(t, "secret123", retrieved.Key)
		assert.Equal(t, "Test API key", retrieved.Description)
		assert.True(t, retrieved.IsActive)

		// Validate API key
		valid, err := service.ValidateAPIKey("secret123")
		assert.NoError(t, err)
		assert.True(t, valid)

		// Validate invalid API key
		valid, err = service.ValidateAPIKey("wrong-key")
		assert.NoError(t, err)
		assert.False(t, valid)

		// List API keys
		keys, err := service.ListAPIKeys()
		assert.NoError(t, err)
		assert.Len(t, keys, 1)
		assert.Equal(t, "test-key-1", keys[0])

		// Delete API key
		err = service.DeleteAPIKey("test-key-1")
		assert.NoError(t, err)

		// Verify deletion
		_, err = service.GetAPIKey("test-key-1")
		assert.Error(t, err)

		// Verify validation fails after deletion
		valid, err = service.ValidateAPIKey("secret123")
		assert.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("System Config Management", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "freyja_system_test_config")
		assert.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		config := SystemConfig{
			DataDir:          tmpDir,
			EncryptionKey:    "",
			EnableEncryption: false,
		}

		service, err := NewSystemService(config)
		assert.NoError(t, err)
		defer service.Close()

		err = service.Open()
		assert.NoError(t, err)

		// Store config
		testConfig := map[string]string{
			"server": "localhost",
			"port":   "8080",
		}

		err = service.StoreSystemConfig("test-config", testConfig)
		assert.NoError(t, err)

		// Retrieve config
		var retrieved map[string]string
		err = service.GetSystemConfig("test-config", &retrieved)
		assert.NoError(t, err)
		assert.Equal(t, testConfig, retrieved)
	})

	t.Run("Encryption", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "freyja_system_test_encrypt")
		assert.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		config := SystemConfig{
			DataDir:          tmpDir,
			EncryptionKey:    "12345678901234567890123456789012", // 32 bytes for AES-256
			EnableEncryption: true,
		}

		service, err := NewSystemService(config)
		assert.NoError(t, err)
		defer service.Close()

		err = service.Open()
		assert.NoError(t, err)

		// Create API key
		apiKey := APIKey{
			ID:          "encrypted-key",
			Key:         "super-secret-key",
			Description: "Encrypted API key",
			CreatedAt:   time.Now(),
			IsActive:    true,
		}

		// Store API key (should be encrypted)
		err = service.StoreAPIKey(apiKey)
		assert.NoError(t, err)

		// Retrieve and validate API key (should be decrypted)
		retrieved, err := service.GetAPIKey("encrypted-key")
		assert.NoError(t, err)
		assert.Equal(t, "super-secret-key", retrieved.Key)
	})

	t.Run("Key Derivation", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "freyja_system_test_keyderiv")
		assert.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Test with various key lengths - all should work due to SHA-256 derivation
		testKeys := []string{
			"short",                            // 5 bytes
			"cuddly-kitten",                    // 13 bytes (original failing case)
			"medium-length-key-for-testing",    // 28 bytes
			"12345678901234567890123456789012", // 32 bytes (exact)
		}

		for _, testKey := range testKeys {
			t.Run(fmt.Sprintf("key_%d_bytes", len(testKey)), func(t *testing.T) {
				config := SystemConfig{
					DataDir:          tmpDir,
					EncryptionKey:    testKey,
					EnableEncryption: true,
				}

				service, err := NewSystemService(config)
				assert.NoError(t, err)
				defer service.Close()

				err = service.Open()
				assert.NoError(t, err)

				// Create API key
				apiKey := APIKey{
					ID:          "test-key-" + testKey,
					Key:         "test-value",
					Description: "Test API key with derived key",
					CreatedAt:   time.Now(),
					IsActive:    true,
				}

				// Store API key (should be encrypted with derived key)
				err = service.StoreAPIKey(apiKey)
				assert.NoError(t, err)

				// Retrieve and validate API key (should be decrypted with same derived key)
				retrieved, err := service.GetAPIKey("test-key-" + testKey)
				assert.NoError(t, err)
				assert.Equal(t, "test-value", retrieved.Key)
			})
		}
	})
}
