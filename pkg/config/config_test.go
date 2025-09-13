package config

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "./data", config.DataDir)
	assert.Equal(t, 8080, config.Port)
	assert.Equal(t, "127.0.0.1", config.Bind)
	assert.Equal(t, "auto", config.Security.SystemKey)
	assert.Equal(t, "auto", config.Security.SystemAPIKey)
	assert.Equal(t, "auto", config.Security.ClientAPIKey)
	assert.Equal(t, 4096, config.Security.MaxRecordSize)
	assert.Equal(t, "info", config.Logging.Level)
}

func TestGenerateSecureKey(t *testing.T) {
	t.Run("generate 32 byte key", func(t *testing.T) {
		key, err := GenerateSecureKey(32)
		require.NoError(t, err)
		assert.Len(t, key, 64) // 32 bytes = 64 hex characters

		// Verify it's valid hex
		_, err = hex.DecodeString(key)
		assert.NoError(t, err)
	})

	t.Run("generate different keys", func(t *testing.T) {
		key1, err := GenerateSecureKey(16)
		require.NoError(t, err)
		key2, err := GenerateSecureKey(16)
		require.NoError(t, err)

		assert.NotEqual(t, key1, key2)
	})

	t.Run("zero length", func(t *testing.T) {
		key, err := GenerateSecureKey(0)
		require.NoError(t, err)
		assert.Empty(t, key)
	})
}

func TestLoadConfig(t *testing.T) {
	t.Run("load existing config", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "freyja_config_test")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		configPath := filepath.Join(tmpDir, "config.yaml")
		expectedConfig := &Config{
			DataDir: "/custom/data",
			Port:    9000,
			Bind:    "0.0.0.0",
			Security: Security{
				SystemKey:     "test-system-key",
				SystemAPIKey:  "test-system-api-key",
				ClientAPIKey:  "test-client-api-key",
				MaxRecordSize: 4096,
			},
			Logging: Logging{
				Level: "debug",
			},
		}

		err = SaveConfig(expectedConfig, configPath)
		require.NoError(t, err)

		loadedConfig, err := LoadConfig(configPath)
		require.NoError(t, err)
		assert.Equal(t, expectedConfig, loadedConfig)
	})

	t.Run("load non-existent config", func(t *testing.T) {
		_, err := LoadConfig("/non/existent/config.yaml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config file does not exist")
	})

	t.Run("load invalid yaml", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "freyja_config_test")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		configPath := filepath.Join(tmpDir, "invalid.yaml")
		err = os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644)
		require.NoError(t, err)

		_, err = LoadConfig(configPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse config file")
	})
}

func TestSaveConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "freyja_config_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.yaml")
	config := DefaultConfig()

	err = SaveConfig(config, configPath)
	require.NoError(t, err)

	// Verify file exists
	info, err := os.Stat(configPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	// Verify content
	loadedConfig, err := LoadConfig(configPath)
	require.NoError(t, err)
	assert.Equal(t, config, loadedConfig)
}

func TestBootstrapConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "freyja_config_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.yaml")
	dataDir := "/custom/data/dir"

	config, err := BootstrapConfig(configPath, dataDir)
	require.NoError(t, err)

	// Verify config values
	assert.Equal(t, dataDir, config.DataDir)
	assert.Equal(t, 8080, config.Port)
	assert.Equal(t, "127.0.0.1", config.Bind)
	assert.Equal(t, "info", config.Logging.Level)

	// Verify keys are generated and not "auto"
	assert.NotEqual(t, "auto", config.Security.SystemKey)
	assert.NotEqual(t, "auto", config.Security.SystemAPIKey)
	assert.NotEqual(t, "auto", config.Security.ClientAPIKey)

	// Verify keys are valid hex
	_, err = hex.DecodeString(config.Security.SystemKey)
	assert.NoError(t, err)
	_, err = hex.DecodeString(config.Security.SystemAPIKey)
	assert.NoError(t, err)
	_, err = hex.DecodeString(config.Security.ClientAPIKey)
	assert.NoError(t, err)

	// Verify file was created
	assert.True(t, ConfigExists(configPath))

	// Verify we can load it back
	loadedConfig, err := LoadConfig(configPath)
	require.NoError(t, err)
	assert.Equal(t, config, loadedConfig)
}

func TestGetDefaultConfigPath(t *testing.T) {
	path := GetDefaultConfigPath()
	assert.NotEmpty(t, path)
	assert.Contains(t, path, "freyja")
	assert.Contains(t, path, "config.yaml")
}

func TestConfigExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "freyja_config_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	existingPath := filepath.Join(tmpDir, "exists.yaml")
	nonExistentPath := filepath.Join(tmpDir, "does-not-exist.yaml")

	// Create a file
	err = os.WriteFile(existingPath, []byte("test"), 0644)
	require.NoError(t, err)

	assert.True(t, ConfigExists(existingPath))
	assert.False(t, ConfigExists(nonExistentPath))
}

func TestConfigYAMLMarshalling(t *testing.T) {
	config := &Config{
		DataDir: "/test/data",
		Port:    9999,
		Bind:    "localhost",
		Security: Security{
			SystemKey:     "system-key-123",
			SystemAPIKey:  "system-api-key-456",
			ClientAPIKey:  "client-api-key-789",
			MaxRecordSize: 4096,
		},
		Logging: Logging{
			Level: "warn",
		},
	}

	data, err := yaml.Marshal(config)
	require.NoError(t, err)

	var unmarshalled Config
	err = yaml.Unmarshal(data, &unmarshalled)
	require.NoError(t, err)

	assert.Equal(t, config, &unmarshalled)
}

func TestSaveConfigErrorHandling(t *testing.T) {
	config := DefaultConfig()

	// Try to save to a directory that can't be created
	invalidPath := "/invalid/path/that/cannot/be/created/config.yaml"

	err := SaveConfig(config, invalidPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create config directory")
}
