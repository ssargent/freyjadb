/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/ssargent/freyjadb/pkg/store"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "freyja",
	Short: "FreyjaDB - Embeddable KV Store",
	Long: `FreyjaDB is a Bitcask-style embeddable key-value store with
optional partitioning and sort keys.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		dataDir, _ := cmd.Flags().GetString("data-dir")
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return fmt.Errorf("failed to create data dir: %w", err)
		}
		kvStore, err := store.NewKVStore(store.KVStoreConfig{DataDir: dataDir})
		if err != nil {
			return fmt.Errorf("failed to create store: %w", err)
		}
		recovery, err := kvStore.Open()
		if err != nil {
			return fmt.Errorf("failed to open store: %w", err)
		}
		if recovery.RecordsTruncated > 0 {
			fmt.Printf("Recovered from corruption: %d records truncated\n", recovery.RecordsTruncated)
		}
		// Store in command context
		cmd.SetContext(context.WithValue(cmd.Context(), "store", kvStore))
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global data directory flag
	rootCmd.PersistentFlags().StringP("data-dir", "d", "./data", "Data directory for the store")
}
