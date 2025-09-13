package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ssargent/freyjadb/pkg/config"
	"github.com/ssargent/freyjadb/pkg/di"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpCommand(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "freyja_up_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dataDir := filepath.Join(tmpDir, "data")
	configPath := filepath.Join(tmpDir, "config.yaml")

	t.Run("bootstrap and config creation", func(t *testing.T) {
		// Initialize dependency injection container
		container := di.NewContainer()
		SetContainer(container)

		// Bootstrap config
		cfg, err := config.BootstrapConfig(configPath, dataDir)
		require.NoError(t, err)

		// Verify config was created
		assert.True(t, config.ConfigExists(configPath))

		// Verify config content
		loadedConfig, err := config.LoadConfig(configPath)
		require.NoError(t, err)
		assert.Equal(t, dataDir, loadedConfig.DataDir)
		assert.Equal(t, cfg.Security.SystemKey, loadedConfig.Security.SystemKey)
		assert.Equal(t, cfg.Security.SystemAPIKey, loadedConfig.Security.SystemAPIKey)
		assert.Equal(t, cfg.Security.ClientAPIKey, loadedConfig.Security.ClientAPIKey)
	})

	t.Run("load existing config", func(t *testing.T) {
		// Create a config file
		existingConfig := &config.Config{
			DataDir: dataDir,
			Port:    9000,
			Bind:    "0.0.0.0",
			Security: config.Security{
				SystemKey:    "existing-system-key",
				SystemAPIKey: "existing-system-api-key",
				ClientAPIKey: "existing-client-api-key",
			},
			Logging: config.Logging{
				Level: "debug",
			},
		}

		err := config.SaveConfig(existingConfig, configPath)
		require.NoError(t, err)

		// Load the config
		loadedConfig, err := config.LoadConfig(configPath)
		require.NoError(t, err)
		assert.Equal(t, existingConfig, loadedConfig)
	})

	t.Run("initialize system if needed", func(t *testing.T) {
		// Initialize dependency injection container
		container := di.NewContainer()
		SetContainer(container)

		// Create config
		cfg := &config.Config{
			DataDir: dataDir,
			Port:    8080,
			Bind:    "127.0.0.1",
			Security: config.Security{
				SystemKey:    "test-system-key-1234567890123456",
				SystemAPIKey: "test-system-api-key-1234567890123456",
				ClientAPIKey: "test-client-api-key-1234567890123456",
			},
			Logging: config.Logging{
				Level: "info",
			},
		}

		// Test system initialization
		err := initializeSystemIfNeeded(cfg)
		require.NoError(t, err)

		// Verify system directory was created
		systemDir := filepath.Join(dataDir, "system")
		assert.DirExists(t, systemDir)

		// Verify system data file was created
		systemFile := filepath.Join(systemDir, "active.data")
		assert.FileExists(t, systemFile)
	})

	t.Run("system already initialized", func(t *testing.T) {
		// Initialize dependency injection container
		container := di.NewContainer()
		SetContainer(container)

		// Create config
		cfg := &config.Config{
			DataDir: dataDir,
			Port:    8080,
			Bind:    "127.0.0.1",
			Security: config.Security{
				SystemKey:    "test-system-key-1234567890123456",
				SystemAPIKey: "test-system-api-key-1234567890123456",
				ClientAPIKey: "test-client-api-key-1234567890123456",
			},
			Logging: config.Logging{
				Level: "info",
			},
		}

		// First initialization
		err := initializeSystemIfNeeded(cfg)
		require.NoError(t, err)

		// Second initialization should not fail
		err = initializeSystemIfNeeded(cfg)
		assert.NoError(t, err)
	})

	t.Run("invalid data directory", func(t *testing.T) {
		// Initialize dependency injection container
		container := di.NewContainer()
		SetContainer(container)

		// Create config with invalid data directory
		cfg := &config.Config{
			DataDir: "/invalid/path/that/does/not/exist",
			Port:    8080,
			Bind:    "127.0.0.1",
			Security: config.Security{
				SystemKey:    "test-system-key-1234567890123456",
				SystemAPIKey: "test-system-api-key-1234567890123456",
				ClientAPIKey: "test-client-api-key-1234567890123456",
			},
			Logging: config.Logging{
				Level: "info",
			},
		}

		// System initialization should fail with invalid path
		err := initializeSystemIfNeeded(cfg)
		assert.Error(t, err)
	})

	t.Run("container not initialized", func(t *testing.T) {
		// Reset container
		SetContainer(nil)

		cfg := config.DefaultConfig()
		err := initializeSystemIfNeeded(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "dependency container not initialized")
	})
}

func TestDefaultConfigPath(t *testing.T) {
	path := config.GetDefaultConfigPath()
	assert.NotEmpty(t, path)
	assert.Contains(t, path, ".config")
	assert.Contains(t, path, "freyja")
	assert.Contains(t, path, "config.yaml")
}

func TestConfigOverride(t *testing.T) {
	// Test that command line flags override config values
	tmpDir, err := os.MkdirTemp("", "freyja_override_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create base config
	baseConfig := &config.Config{
		DataDir: "./data",
		Port:    8080,
		Bind:    "127.0.0.1",
		Security: config.Security{
			SystemKey:    "base-system-key",
			SystemAPIKey: "base-system-api-key",
			ClientAPIKey: "base-client-api-key",
		},
		Logging: config.Logging{
			Level: "info",
		},
	}

	err = config.SaveConfig(baseConfig, configPath)
	require.NoError(t, err)

	// Load and modify as if command line flags were provided
	loadedConfig, err := config.LoadConfig(configPath)
	require.NoError(t, err)

	// Simulate flag overrides
	customDataDir := "/custom/data/dir"
	customPort := 9000
	customBind := "0.0.0.0"

	if customDataDir != "" {
		loadedConfig.DataDir = customDataDir
	}
	if customPort != 8080 {
		loadedConfig.Port = customPort
	}
	if customBind != "127.0.0.1" {
		loadedConfig.Bind = customBind
	}

	// Verify overrides
	assert.Equal(t, customDataDir, loadedConfig.DataDir)
	assert.Equal(t, customPort, loadedConfig.Port)
	assert.Equal(t, customBind, loadedConfig.Bind)
}

func TestUpCommandFlagHandling(t *testing.T) {
	// Test flag parsing and default values
	t.Run("default config path", func(t *testing.T) {
		// Test that empty config path uses default
		testConfigPath := ""
		if testConfigPath == "" {
			testConfigPath = config.GetDefaultConfigPath()
		}
		assert.NotEmpty(t, testConfigPath)
		assert.Contains(t, testConfigPath, "freyja")
	})

	t.Run("custom config path", func(t *testing.T) {
		// Test custom config path
		customPath := "/custom/config/path.yaml"
		testConfigPath := customPath
		if testConfigPath == "" {
			testConfigPath = config.GetDefaultConfigPath()
		}
		assert.Equal(t, customPath, testConfigPath)
	})

	t.Run("flag override logic", func(t *testing.T) {
		// Test the flag override logic from the Run function
		cfg := &config.Config{
			DataDir: "./data",
			Port:    8080,
			Bind:    "127.0.0.1",
		}

		// Simulate flag values
		dataDir := "/flag/data/dir"
		port := 9000
		bind := "0.0.0.0"

		// Apply overrides (same logic as in Run function)
		if dataDir != "" {
			cfg.DataDir = dataDir
		}
		if port != 8080 {
			cfg.Port = port
		}
		if bind != "127.0.0.1" {
			cfg.Bind = bind
		}

		assert.Equal(t, "/flag/data/dir", cfg.DataDir)
		assert.Equal(t, 9000, cfg.Port)
		assert.Equal(t, "0.0.0.0", cfg.Bind)
	})

	t.Run("no overrides", func(t *testing.T) {
		// Test when no flags are provided (should keep config values)
		cfg := &config.Config{
			DataDir: "/config/data",
			Port:    8080,
			Bind:    "127.0.0.1",
		}

		// Simulate empty/default flag values
		dataDir := ""
		port := 8080
		bind := "127.0.0.1"

		// Apply overrides
		if dataDir != "" {
			cfg.DataDir = dataDir
		}
		if port != 8080 {
			cfg.Port = port
		}
		if bind != "127.0.0.1" {
			cfg.Bind = bind
		}

		// Should remain unchanged
		assert.Equal(t, "/config/data", cfg.DataDir)
		assert.Equal(t, 8080, cfg.Port)
		assert.Equal(t, "127.0.0.1", cfg.Bind)
	})
}

func TestUpCommandErrorHandling(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "freyja_error_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Run("invalid config file", func(t *testing.T) {
		// Create invalid config file
		invalidConfigPath := filepath.Join(tmpDir, "invalid.yaml")
		err := os.WriteFile(invalidConfigPath, []byte("invalid: yaml: content: ["), 0600)
		require.NoError(t, err)

		// Try to load invalid config
		_, err = config.LoadConfig(invalidConfigPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse config file")
	})

	t.Run("config bootstrap failure", func(t *testing.T) {
		// Try to bootstrap to invalid path
		invalidPath := "/invalid/path/config.yaml"
		_, err := config.BootstrapConfig(invalidPath, "/some/data")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create config directory")
	})

	t.Run("config save failure", func(t *testing.T) {
		// Try to save config to invalid path
		cfg := config.DefaultConfig()
		invalidPath := "/invalid/path/config.yaml"
		err := config.SaveConfig(cfg, invalidPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create config directory")
	})
}
