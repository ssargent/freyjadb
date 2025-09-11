package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ssargent/freyjadb/pkg/store"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete <key>",
	Short: "Delete a key-value pair",
	Long: `Delete a key-value pair from the FreyjaDB store.

Example:
  freyja delete mykey`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := []byte(args[0])

		// Create KV store
		config := store.KVStoreConfig{
			DataDir:       dataDir,
			FsyncInterval: 0,
		}

		kv, err := store.NewKVStore(config)
		if err != nil {
			fmt.Printf("Error creating store: %v\n", err)
			return
		}

		// Open store
		_, err = kv.Open()
		if err != nil {
			fmt.Printf("Error opening store: %v\n", err)
			return
		}
		defer kv.Close()

		// Delete key
		if err := kv.Delete(key); err != nil {
			fmt.Printf("Error deleting key: %v\n", err)
			return
		}

		fmt.Printf("Successfully deleted key '%s'\n", string(key))
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringVar(&dataDir, "data-dir", "./data", "Data directory for the store")
}
