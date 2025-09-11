package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ssargent/freyjadb/pkg/store"
)

var dataDir string

// putCmd represents the put command
var putCmd = &cobra.Command{
	Use:   "put <key> <value>",
	Short: "Put a key-value pair",
	Long: `Put a key-value pair into the FreyjaDB store.

Example:
  freyja put mykey myvalue`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := []byte(args[0])
		value := []byte(args[1])

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

		// Put key-value pair
		if err := kv.Put(key, value); err != nil {
			fmt.Printf("Error putting key-value: %v\n", err)
			return
		}

		fmt.Printf("Successfully put key '%s' with value '%s'\n", string(key), string(value))
	},
}

func init() {
	rootCmd.AddCommand(putCmd)
	putCmd.Flags().StringVar(&dataDir, "data-dir", "./data", "Data directory for the store")
}
