/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/ssargent/freyjadb/pkg/api"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize FreyjaDB system for local development",
	Long: `Initialize FreyjaDB system store and set up system API key for local development.

This command will:
- Create the system data directory
- Initialize the system key-value store
- Set up the system API key for administrative operations
- Enable encryption for system data

This is required before running the server in development mode.

Examples:
	  freyja init --system-key=my-system-secret --data-dir=./data
	  freyja init --system-key=my-system-secret --system-api-key=my-api-key --data-dir=./data`,
	Run: func(cmd *cobra.Command, args []string) {
		systemKey, _ := cmd.Flags().GetString("system-key")
		systemAPIKey, _ := cmd.Flags().GetString("system-api-key")
		dataDir, _ := cmd.Flags().GetString("data-dir")
		force, _ := cmd.Flags().GetBool("force")

		if systemKey == "" {
			cmd.Printf("Error: --system-key is required\n")
			os.Exit(1)
		}

		if dataDir == "" {
			dataDir = "./data"
		}

		// Generate API key if not provided
		if systemAPIKey == "" {
			var err error
			systemAPIKey, err = generateSystemAPIKey()
			if err != nil {
				cmd.Printf("Error generating system API key: %v\n", err)
				os.Exit(1)
			}
		}

		cmd.Printf("Initializing FreyjaDB system...\n")
		cmd.Printf("Data directory: %s\n", dataDir)
		cmd.Printf("System key: %s\n", systemKey[:8]+"...")

		// Create data directory
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			cmd.Printf("Error creating data directory: %v\n", err)
			os.Exit(1)
		}

		// Check if system is already initialized
		systemDataDir := dataDir
		systemStorePath := fmt.Sprintf("%s/system/active.data", systemDataDir)

		if _, err := os.Stat(systemStorePath); err == nil && !force {
			cmd.Printf("System already initialized. Use --force to reinitialize.\n")
			cmd.Printf("System data location: %s\n", systemStorePath)
			return
		}

		// Initialize system store
		if err := initializeSystemStore(systemDataDir, systemKey, systemAPIKey); err != nil {
			cmd.Printf("Error initializing system store: %v\n", err)
			os.Exit(1)
		}

		cmd.Printf("✅ FreyjaDB system initialization completed successfully!\n")
		cmd.Printf("System API key: %s\n", systemAPIKey)
		cmd.Printf("Data directory: %s\n", dataDir)
		cmd.Printf("\nYou can now start the server with:\n")
		cmd.Printf("  freyja serve --api-key=your-user-key --system-key=%s --data-dir=%s\n", systemKey, dataDir)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().String("system-key", "", "System encryption key for data protection (required)")
	initCmd.Flags().String("system-api-key", "", "System API key for administrative operations (optional, will be generated if not provided)")
	initCmd.Flags().String("data-dir", "./data", "Data directory for freyja")
	initCmd.Flags().Bool("force", false, "Force reinitialization even if system already exists")
	initCmd.MarkFlagRequired("system-key")
}

// generateSystemAPIKey generates a secure random API key
func generateSystemAPIKey() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random API key: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// initializeSystemStore sets up the system key-value store and stores the system API key
func initializeSystemStore(dataDir, systemKey, systemAPIKey string) error {
	// Create system service
	encryptionKey := systemKey
	if len(encryptionKey) < 32 {
		// Pad with zeros if key is too short
		padding := make([]byte, 32-len(encryptionKey))
		encryptionKey = encryptionKey + string(padding)
	} else if len(encryptionKey) > 32 {
		// Truncate if key is too long
		encryptionKey = encryptionKey[:32]
	}

	systemConfig := api.SystemConfig{
		DataDir:          dataDir,
		EncryptionKey:    encryptionKey,
		EnableEncryption: true,
	}

	systemService, err := api.NewSystemService(systemConfig)
	if err != nil {
		return fmt.Errorf("failed to create system service: %w", err)
	}

	// Open system service
	if err := systemService.Open(); err != nil {
		return fmt.Errorf("failed to open system service: %w", err)
	}
	defer systemService.Close()

	// Store system API key
	apiKey := api.APIKey{
		ID:          "system-root",
		Key:         systemAPIKey,
		Description: "System root API key for administrative operations",
		CreatedAt:   time.Now(),
		IsActive:    true,
	}

	if err := systemService.StoreAPIKey(apiKey); err != nil {
		return fmt.Errorf("failed to store system API key: %w", err)
	}

	// Store some default system configuration
	defaultConfig := map[string]interface{}{
		"initialized_at":     time.Now().Format(time.RFC3339),
		"version":            "1.0.0",
		"encryption_enabled": true,
	}

	if err := systemService.StoreSystemConfig("system-info", defaultConfig); err != nil {
		return fmt.Errorf("failed to store system configuration: %w", err)
	}

	return nil
}
