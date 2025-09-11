package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ssargent/freyjadb/pkg/store"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a value for a key",
	Long: `Get a value for a key from the FreyjaDB store.

Example:
  freyja get mykey`,
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
		if err := kv.Open(); err != nil {
			fmt.Printf("Error opening store: %v\n", err)
			return
		}
		defer kv.Close()

		// Get value
		value, err := kv.Get(key)
		if err != nil {
			fmt.Printf("Error getting value: %v\n", err)
			return
		}

		fmt.Printf("%s\n", string(value))
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().StringVar(&dataDir, "data-dir", "./data", "Data directory for the store")
}
