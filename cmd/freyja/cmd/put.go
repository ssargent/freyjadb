package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ssargent/freyjadb/pkg/store"
)

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

		// Get store from context
		kv, ok := cmd.Context().Value("store").(*store.KVStore)
		if !ok {
			fmt.Printf("Error: store not found in context\n")
			return
		}

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
}
