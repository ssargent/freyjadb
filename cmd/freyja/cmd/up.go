/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ssargent/freyjadb/pkg/config"
	"github.com/ssargent/freyjadb/pkg/store"
)

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Bootstrap and start FreyjaDB server",
	Long: `Bootstrap FreyjaDB by creating configuration and keys if they don't exist,
then start the REST API server. This is the recommended way to get FreyjaDB running.

The command will:
- Create configuration file with secure keys if missing
- Initialize the system store
- Start the REST API server

Examples:
  freyja up
  freyja up --data-dir ./mydata --port 9000
  freyja up --config ./custom-config.yaml --non-interactive`,
	Run: func(cmd *cobra.Command, args []string) {
		dataDir, _ := cmd.Flags().GetString("data-dir")
		port, _ := cmd.Flags().GetInt("port")
		bind, _ := cmd.Flags().GetString("bind")
		configPath, _ := cmd.Flags().GetString("config")
		printKeys, _ := cmd.Flags().GetBool("print-keys")

		// Use default config path if not specified
		if configPath == "" {
			configPath = config.GetDefaultConfigPath()
		}

		var cfg *config.Config
		var err error

		// Check if config exists
		if config.ConfigExists(configPath) {
			// Load existing config
			cfg, err = config.LoadConfig(configPath)
			if err != nil {
				cmd.Printf("Error loading existing config: %v\n", err)
				os.Exit(1)
			}
			cmd.Printf("✅ Loaded existing configuration from %s\n", configPath)
		} else {
			// Bootstrap new config
			cmd.Printf("🔧 First run detected. Bootstrapping FreyjaDB...\n")

			cfg, err = config.BootstrapConfig(configPath, dataDir)
			if err != nil {
				cmd.Printf("Error bootstrapping config: %v\n", err)
				os.Exit(1)
			}

			cmd.Printf("✅ Configuration created at %s\n", configPath)

			if printKeys {
				cmd.Printf("\n🔑 Generated Keys:\n")
				cmd.Printf("System Key: %s\n", cfg.Security.SystemKey)
				cmd.Printf("System API Key: %s\n", cfg.Security.SystemAPIKey)
				cmd.Printf("Client API Key: %s\n", cfg.Security.ClientAPIKey)
				cmd.Printf("\n⚠️  Store these keys securely! They are also saved in %s\n", configPath)
			}
		}

		// Override config with command line flags if provided
		if dataDir != "" {
			cfg.DataDir = dataDir
		}
		if port != 8080 { // Only override if explicitly set
			cfg.Port = port
		}
		if bind != "127.0.0.1" { // Only override if explicitly set
			cfg.Bind = bind
		}

		// Initialize system if needed
		if err := initializeSystemIfNeeded(cfg); err != nil {
			cmd.Printf("Error initializing system: %v\n", err)
			os.Exit(1)
		}

		// Start the server
		cmd.Printf("🚀 Starting FreyjaDB server on %s:%d\n", cfg.Bind, cfg.Port)
		cmd.Printf("📁 Data directory: %s\n", cfg.DataDir)

		if container == nil {
			cmd.Printf("Error: dependency container not initialized\n")
			os.Exit(1)
		}

		serverFactory := container.GetServerFactory()
		serverStarter := serverFactory.CreateServerStarter()

		// Get store from context (created by root command)
		kv, ok := cmd.Context().Value("store").(*store.KVStore)
		if !ok {
			cmd.Printf("Error: store not found in context\n")
			os.Exit(1)
		}

		if err := serverStarter.StartServer(kv, cfg.Port, cfg.Security.ClientAPIKey,
			cfg.Security.SystemKey, cfg.DataDir, cfg.Security.SystemKey, true); err != nil {
			cmd.Printf("Error starting server: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(upCmd)

	upCmd.Flags().StringP("data-dir", "d", "./data", "Data directory for the store")
	upCmd.Flags().IntP("port", "p", 8080, "Port to listen on")
	upCmd.Flags().String("bind", "127.0.0.1", "Address to bind server to")
	upCmd.Flags().String("config", "", "Path to config file (default: OS-specific location)")
	upCmd.Flags().Bool("non-interactive", false, "Skip prompts and use defaults")
	upCmd.Flags().Bool("print-keys", false, "Print generated API keys to console")
}

// initializeSystemIfNeeded initializes the system store if it doesn't exist
func initializeSystemIfNeeded(cfg *config.Config) error {
	if container == nil {
		return fmt.Errorf("dependency container not initialized")
	}

	// Check if system is already initialized
	systemDataDir := cfg.DataDir
	systemStorePath := fmt.Sprintf("%s/system/active.data", systemDataDir)

	if _, err := os.Stat(systemStorePath); err == nil {
		// System already exists
		return nil
	}

	// Initialize system store
	factory := container.GetSystemServiceFactory()
	systemService, err := factory.CreateSystemService(systemDataDir, cfg.Security.SystemKey, true)
	if err != nil {
		return fmt.Errorf("failed to create system service: %w", err)
	}

	if err := systemService.InitializeSystem(systemDataDir, cfg.Security.SystemKey,
		cfg.Security.SystemAPIKey); err != nil {
		return fmt.Errorf("failed to initialize system store: %w", err)
	}

	return nil
}
