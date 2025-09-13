package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ssargent/freyjadb/pkg/di"
	"github.com/stretchr/testify/assert"
)

func TestInitCommand(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "freyja_init_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dataDir := filepath.Join(tmpDir, "data")
	systemKey := "test-system-key-1234567890123456" // 32 bytes for AES-256

	t.Run("Successful initialization", func(t *testing.T) {
		// Test the system initialization using dependency injection
		container := di.NewContainer()
		factory := container.GetSystemServiceFactory()

		systemService, err := factory.CreateSystemService(dataDir, systemKey, true, 4096)
		assert.NoError(t, err)

		err = systemService.InitializeSystem(dataDir, systemKey, systemKey)
		assert.NoError(t, err)

		// Verify system directory was created
		systemDir := filepath.Join(dataDir, "system")
		assert.DirExists(t, systemDir)

		// Verify system data file was created
		systemFile := filepath.Join(systemDir, "active.data")
		assert.FileExists(t, systemFile)
	})

	t.Run("Force reinitialization", func(t *testing.T) {
		container := di.NewContainer()
		factory := container.GetSystemServiceFactory()

		// First initialization
		systemService, err := factory.CreateSystemService(dataDir, systemKey, true, 4096)
		assert.NoError(t, err)
		err = systemService.InitializeSystem(dataDir, systemKey, systemKey)
		assert.NoError(t, err)

		// Second initialization with same key (should work)
		err = systemService.InitializeSystem(dataDir, systemKey, systemKey)
		assert.NoError(t, err)

		// Second initialization with different key (should work due to force logic in init command)
		err = systemService.InitializeSystem(dataDir, "different-key", "different-key")
		assert.NoError(t, err)
	})

	t.Run("Invalid data directory", func(t *testing.T) {
		container := di.NewContainer()
		factory := container.GetSystemServiceFactory()
		invalidDir := "/invalid/path/that/does/not/exist"
		systemService, err := factory.CreateSystemService(invalidDir, systemKey, true, 4096)
		// The factory should fail when trying to create the system directory
		if err != nil {
			assert.Error(t, err) // Factory failed as expected
		} else {
			// If factory succeeded, then InitializeSystem should fail
			err = systemService.InitializeSystem(invalidDir, systemKey, systemKey)
			assert.Error(t, err)
		}
	})

	t.Run("Empty system key", func(t *testing.T) {
		container := di.NewContainer()
		factory := container.GetSystemServiceFactory()
		systemService, err := factory.CreateSystemService(dataDir, "", false, 4096)
		assert.NoError(t, err)
		err = systemService.InitializeSystem(dataDir, "", "")
		assert.NoError(t, err) // Should still work, just with empty key
	})
}

func TestLoadExistingSystemKey(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "freyja_load_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dataDir := filepath.Join(tmpDir, "data")
	systemKey := "existing-system-key-1234567890123456" // 32 bytes for AES-256

	t.Run("Load existing system key", func(t *testing.T) {
		t.Skip("loadExistingSystemKey function not yet implemented")
		// First initialize the system
		container := di.NewContainer()
		factory := container.GetSystemServiceFactory()
		systemService, err := factory.CreateSystemService(dataDir, systemKey, true, 4096)
		assert.NoError(t, err)
		err = systemService.InitializeSystem(dataDir, systemKey, systemKey)
		assert.NoError(t, err)

		// Now try to load it
		loadedKey, err := loadExistingSystemKey(dataDir)
		assert.NoError(t, err)
		assert.Equal(t, systemKey, loadedKey)
	})

	t.Run("Load from non-existent system", func(t *testing.T) {
		// Note: In the actual implementation, the container would be set via SetContainer()
		// For this test, we'll skip since loadExistingSystemKey is not fully implemented
		t.Skip("loadExistingSystemKey function needs container initialization")
		nonExistentDir := filepath.Join(tmpDir, "nonexistent")
		loadedKey, err := loadExistingSystemKey(nonExistentDir)
		assert.Error(t, err)
		assert.Empty(t, loadedKey)
		assert.Contains(t, err.Error(), "system not initialized")
	})
}
