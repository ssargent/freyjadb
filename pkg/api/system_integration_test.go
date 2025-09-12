package api

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSystemServiceIntegration(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "freyja_system_integration")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Run("SystemService lifecycle", func(t *testing.T) {
		// Create system service
		config := SystemConfig{
			DataDir:          tmpDir,
			EncryptionKey:    "12345678901234567890123456789012",
			EnableEncryption: true,
		}

		service, err := NewSystemService(config)
		assert.NoError(t, err)
		assert.NotNil(t, service)
		assert.False(t, service.IsOpen())

		// Open service
		err = service.Open()
		assert.NoError(t, err)
		assert.True(t, service.IsOpen())

		// Test basic operations
		keys, err := service.ListAPIKeys()
		assert.NoError(t, err)
		assert.Len(t, keys, 0) // Should be empty initially

		// Store a simple config value
		err = service.StoreSystemConfig("test", "integration test")
		assert.NoError(t, err)

		// Retrieve config value
		var value string
		err = service.GetSystemConfig("test", &value)
		assert.NoError(t, err)
		assert.Equal(t, "integration test", value)

		// Close service
		err = service.Close()
		assert.NoError(t, err)
		assert.False(t, service.IsOpen())
	})
}
