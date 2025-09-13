/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/spf13/cobra"
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
		if err := os.MkdirAll(dataDir, 0750); err != nil {
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

		// Initialize system store using dependency injection
		if container == nil {
			cmd.Printf("Error: dependency container not initialized\n")
			os.Exit(1)
		}

		factory := container.GetSystemServiceFactory()

		systemService, err := factory.CreateSystemService(systemDataDir, systemKey, true, 4096)
		if err != nil {
			cmd.Printf("Error creating system service: %v\n", err)
			os.Exit(1)
		}

		if err := systemService.InitializeSystem(systemDataDir, systemKey, systemAPIKey); err != nil {
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
	initCmd.Flags().String("system-api-key", "",
		"System API key for administrative operations (optional, will be generated if not provided)")
	initCmd.Flags().String("data-dir", "./data", "Data directory for freyja")
	initCmd.Flags().Bool("force", false, "Force reinitialization even if system already exists")
	if err := initCmd.MarkFlagRequired("system-key"); err != nil {
		panic(err)
	}
}

// generateSystemAPIKey generates a secure random API key
func generateSystemAPIKey() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random API key: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
