/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ssargent/freyjadb/pkg/api"
	"github.com/ssargent/freyjadb/pkg/store"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the REST API server",
	Long: `Start the FreyjaDB REST API server with authentication and system data management.

The server supports both user data (via standard APIs) and system data (API keys, configuration)
stored in separate encrypted databases for enhanced security.

Examples:
  freyja serve --api-key=mysecretkey --port=8080
  freyja serve --api-key=mysecretkey --data-dir=./data --enable-encryption --system-encryption-key=my32bytekey`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")
		apiKey, _ := cmd.Flags().GetString("api-key")
		systemKey, _ := cmd.Flags().GetString("system-key")
		dataDir, _ := cmd.Flags().GetString("data-dir")
		systemEncryptionKey, _ := cmd.Flags().GetString("system-encryption-key")
		enableEncryption, _ := cmd.Flags().GetBool("enable-encryption")

		if apiKey == "" {
			cmd.Println("Error: --api-key is required")
			return
		}

		// If system key is not provided, try to load it from existing system store
		if systemKey == "" {
			var err error
			systemKey, err = loadExistingSystemKey(dataDir)
			if err != nil {
				cmd.Printf("Error: --system-key is required (or run 'freyja init' first): %v\n", err)
				return
			}
			cmd.Printf("Loaded existing system key from initialized system\n")
		}

		if dataDir == "" {
			dataDir = "./data" // Default data directory
		}

		// Use system key as encryption key if not specified
		if systemEncryptionKey == "" && enableEncryption {
			systemEncryptionKey = systemKey[:32] // Use first 32 chars of system key
			if len(systemEncryptionKey) < 32 {
				systemEncryptionKey = systemKey + strings.Repeat("0", 32-len(systemKey))
			}
		}

		// Get store from context
		kv, ok := cmd.Context().Value("store").(*store.KVStore)
		if !ok {
			cmd.Println("Error: store not found in context")
			return
		}

		// Start API server
		config := api.ServerConfig{
			Port:                port,
			APIKey:              apiKey,
			SystemKey:           systemKey,
			DataDir:             dataDir,
			SystemDataDir:       dataDir, // Use same base directory for system data
			SystemEncryptionKey: systemEncryptionKey,
			EnableEncryption:    enableEncryption,
		}

		if err := api.StartServer(kv, config); err != nil {
			cmd.Printf("Error starting server: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntP("port", "p", 8080, "Port to listen on")
	serveCmd.Flags().String("api-key", "", "API key for user authentication (required)")
	serveCmd.Flags().String("system-key", "", "System API key for administrative operations (required)")
	serveCmd.Flags().String("data-dir", "./data", "Data directory for storing databases")
	serveCmd.Flags().String("system-encryption-key", "", "Encryption key for system data (32 bytes recommended)")
	serveCmd.Flags().Bool("enable-encryption", false, "Enable encryption for system data")
	serveCmd.MarkFlagRequired("api-key")
	serveCmd.MarkFlagRequired("system-key")
}

// loadExistingSystemKey attempts to load the system API key from an existing initialized system
func loadExistingSystemKey(dataDir string) (string, error) {
	// Create system service config
	systemConfig := api.SystemConfig{
		DataDir:          dataDir,
		EncryptionKey:    "dummy-key-for-loading", // Will be overridden when we load the actual key
		EnableEncryption: false,                   // Disable encryption temporarily to load the key
	}

	systemService, err := api.NewSystemService(systemConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create system service: %w", err)
	}

	// Open system service
	if err := systemService.Open(); err != nil {
		return "", fmt.Errorf("failed to open system service (run 'freyja init' first): %w", err)
	}
	defer systemService.Close()

	// Try to get the system API key
	systemAPIKey, err := systemService.GetAPIKey("system-root")
	if err != nil {
		return "", fmt.Errorf("system not initialized (run 'freyja init' first): %w", err)
	}

	return systemAPIKey.Key, nil
}
