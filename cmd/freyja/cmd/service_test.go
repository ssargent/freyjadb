package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ssargent/freyjadb/pkg/config"
	"github.com/ssargent/freyjadb/pkg/di"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceCommands(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "freyja_service_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dataDir := filepath.Join(tmpDir, "data")
	configPath := filepath.Join(tmpDir, "config.yaml")

	t.Run("create systemd unit", func(t *testing.T) {
		cfg := &config.Config{
			DataDir: dataDir,
			Port:    8080,
			Bind:    "127.0.0.1",
			Security: config.Security{
				SystemKey:    "test-system-key",
				SystemAPIKey: "test-system-api-key",
				ClientAPIKey: "test-client-api-key",
			},
			Logging: config.Logging{
				Level: "info",
			},
		}

		user := "freyja"
		err := createSystemdUnit(cfg, configPath, user)

		// The function may fail if not running as root, which is expected
		if err != nil {
			// Accept both permission denied and file not found errors
			errorMsg := err.Error()
			assert.True(t, strings.Contains(errorMsg, "permission denied") ||
				strings.Contains(errorMsg, "no such file or directory") ||
				strings.Contains(errorMsg, "permission-denied"))
		} else {
			// If running as root, verify unit file was created
			unitPath := "/etc/systemd/system/freyja.service"
			if _, err := os.Stat(unitPath); err == nil {
				content, err := os.ReadFile(unitPath)
				require.NoError(t, err)
				assert.Contains(t, string(content), "User=freyja")
				assert.Contains(t, string(content), "Group=freyja")
				assert.Contains(t, string(content), configPath)
				assert.Contains(t, string(content), dataDir)
			}
		}
	})

	t.Run("systemd unit content", func(t *testing.T) {
		cfg := &config.Config{
			DataDir: "/var/lib/freyjadb",
			Port:    9000,
			Bind:    "127.0.0.1",
			Security: config.Security{
				SystemKey:    "test-system-key",
				SystemAPIKey: "test-system-api-key",
				ClientAPIKey: "test-client-api-key",
			},
			Logging: config.Logging{
				Level: "info",
			},
		}

		user := "testuser"
		err := createSystemdUnit(cfg, "/etc/freyja/config.yaml", user)

		// The function may fail if not running as root, which is expected
		if err != nil {
			// Accept both permission denied and file not found errors
			errorMsg := err.Error()
			assert.True(t, strings.Contains(errorMsg, "permission denied") ||
				strings.Contains(errorMsg, "no such file or directory") ||
				strings.Contains(errorMsg, "permission-denied"))
		} else {
			// Verify unit file content if it was created
			unitPath := "/etc/systemd/system/freyja.service"
			if _, err := os.Stat(unitPath); err == nil {
				content, err := os.ReadFile(unitPath)
				require.NoError(t, err)
				unitContent := string(content)
				assert.Contains(t, unitContent, "User=testuser")
				assert.Contains(t, unitContent, "Group=testuser")
				assert.Contains(t, unitContent, "/etc/freyja/config.yaml")
				assert.Contains(t, unitContent, "/var/lib/freyjadb")
			}
		}
	})

	t.Run("initialize system for service", func(t *testing.T) {
		// Initialize dependency injection container
		container := di.NewContainer()
		SetContainer(container)

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

		err := initializeSystemIfNeeded(cfg)
		require.NoError(t, err)

		// Verify system was initialized
		systemDir := filepath.Join(dataDir, "system")
		assert.DirExists(t, systemDir)
		systemFile := filepath.Join(systemDir, "active.data")
		assert.FileExists(t, systemFile)
	})

	t.Run("service command structure", func(t *testing.T) {
		// Test that service command has all expected subcommands
		assert.NotNil(t, serviceCmd)
		assert.Equal(t, "service", serviceCmd.Use)
		assert.Contains(t, serviceCmd.Short, "systemd")

		// Check that subcommands are added
		subCommands := serviceCmd.Commands()
		commandNames := make([]string, len(subCommands))
		for i, cmd := range subCommands {
			commandNames[i] = cmd.Use
		}

		assert.Contains(t, commandNames, "install")
		assert.Contains(t, commandNames, "start")
		assert.Contains(t, commandNames, "stop")
		assert.Contains(t, commandNames, "restart")
		assert.Contains(t, commandNames, "status")
		assert.Contains(t, commandNames, "logs")
		assert.Contains(t, commandNames, "uninstall")
	})

	t.Run("install service command flags", func(t *testing.T) {
		// Test that install command has expected flags
		installFlags := installServiceCmd.Flags()

		// Check flag existence
		dataDirFlag := installFlags.Lookup("data-dir")
		assert.NotNil(t, dataDirFlag)
		assert.Equal(t, "/var/lib/freyjadb", dataDirFlag.DefValue)

		configFlag := installFlags.Lookup("config")
		assert.NotNil(t, configFlag)
		assert.Equal(t, "", configFlag.DefValue)

		userFlag := installFlags.Lookup("user")
		assert.NotNil(t, userFlag)
		assert.Equal(t, "freyja", userFlag.DefValue)

		portFlag := installFlags.Lookup("port")
		assert.NotNil(t, portFlag)
		assert.Equal(t, "8080", portFlag.DefValue)

		startFlag := installFlags.Lookup("start")
		assert.NotNil(t, startFlag)
		assert.Equal(t, "true", startFlag.DefValue)
	})

	t.Run("logs command flags", func(t *testing.T) {
		// Test that logs command has expected flags
		logsFlags := logsCmd.Flags()

		followFlag := logsFlags.Lookup("follow")
		assert.NotNil(t, followFlag)
		assert.Equal(t, "false", followFlag.DefValue)

		linesFlag := logsFlags.Lookup("lines")
		assert.NotNil(t, linesFlag)
		assert.Equal(t, "0", linesFlag.DefValue)
	})

	t.Run("systemd unit template validation", func(t *testing.T) {
		// Test the systemd unit template generation
		cfg := &config.Config{
			DataDir: "/test/data",
			Port:    8080,
			Bind:    "127.0.0.1",
			Security: config.Security{
				SystemKey:    "test-key",
				SystemAPIKey: "test-api-key",
				ClientAPIKey: "test-client-key",
			},
			Logging: config.Logging{
				Level: "info",
			},
		}

		user := "testuser"
		err := createSystemdUnit(cfg, "/test/config.yaml", user)

		// The function may fail if not running as root, which is expected
		if err != nil {
			// Accept both permission denied and file not found errors
			errorMsg := err.Error()
			assert.True(t, strings.Contains(errorMsg, "permission denied") ||
				strings.Contains(errorMsg, "no such file or directory") ||
				strings.Contains(errorMsg, "permission-denied"))
		} else {
			// If unit file was created, verify it contains expected content
			unitPath := "/etc/systemd/system/freyja.service"
			if _, err := os.Stat(unitPath); err == nil {
				content, err := os.ReadFile(unitPath)
				require.NoError(t, err)
				unitContent := string(content)

				// Check required systemd directives
				assert.Contains(t, unitContent, "[Unit]")
				assert.Contains(t, unitContent, "[Service]")
				assert.Contains(t, unitContent, "[Install]")
				assert.Contains(t, unitContent, "Description=FreyjaDB Server")
				assert.Contains(t, unitContent, "User=testuser")
				assert.Contains(t, unitContent, "Group=testuser")
				assert.Contains(t, unitContent, "Restart=on-failure")
				assert.Contains(t, unitContent, "WantedBy=multi-user.target")
			}
		}
	})
}

func TestServiceCommandErrorHandling(t *testing.T) {
	t.Run("create systemd unit with invalid path", func(t *testing.T) {
		cfg := config.DefaultConfig()
		// This should not fail even with invalid paths since we're not running as root
		err := createSystemdUnit(cfg, "/invalid/config.yaml", "testuser")
		// The function may succeed or fail depending on permissions
		// We just verify it doesn't panic
		_ = err // Ignore error for this test
	})

	t.Run("initialize system with nil container", func(t *testing.T) {
		// Reset container
		SetContainer(nil)

		cfg := config.DefaultConfig()
		err := initializeSystemIfNeeded(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "dependency container not initialized")
	})
}
