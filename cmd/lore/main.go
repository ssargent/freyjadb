package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// Global configuration
type Config struct {
	ProjectDir string
	Format     string
	Quiet      bool
	Yes        bool
}

// Global variables
var (
	config    Config
	loreStore *LoreStore
	rootCmd   *cobra.Command
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd = &cobra.Command{
		Use:   "lore",
		Short: "Book Lore CLI - Manage writing reference notes",
		Long: `A command-line tool for writers to create, browse, and update
reference notes about characters, places, and groups.

Examples:
  lore character create "John Doe" --summary "A brave knight"
  lore place list
  lore group get merchants-guild`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Initialize store
			var err error
			loreStore, err = NewLoreStore(config.ProjectDir)
			if err != nil {
				return fmt.Errorf("failed to initialize store: %w", err)
			}

			// Open the store
			if err := loreStore.Open(); err != nil {
				return fmt.Errorf("failed to open store: %w", err)
			}

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			// Close store
			if loreStore != nil {
				return loreStore.Close()
			}
			return nil
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&config.ProjectDir, "project", "p", ".", "path to project directory")
	rootCmd.PersistentFlags().StringVarP(&config.Format, "format", "o", "table", "output format (table or json)")
	rootCmd.PersistentFlags().BoolVarP(&config.Quiet, "quiet", "q", false, "suppress non-essential messages")
	rootCmd.PersistentFlags().BoolVarP(&config.Yes, "yes", "y", false, "assume 'yes' for prompts")

	// Set default project directory to current working directory
	if cwd, err := os.Getwd(); err == nil {
		config.ProjectDir = cwd
	}

	// Add subcommands
	rootCmd.AddCommand(characterCmd)
	rootCmd.AddCommand(placeCmd)
	rootCmd.AddCommand(groupCmd)
	rootCmd.AddCommand(relationshipCmd)
}

// getProjectDir returns the absolute path to the project directory
func getProjectDir() string {
	if filepath.IsAbs(config.ProjectDir) {
		return config.ProjectDir
	}
	abs, err := filepath.Abs(config.ProjectDir)
	if err != nil {
		return config.ProjectDir
	}
	return abs
}
