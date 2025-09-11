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

		// Get store from context
		kv, ok := cmd.Context().Value("store").(*store.KVStore)
		if !ok {
			fmt.Printf("Error: store not found in context\n")
			return
		}

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
}
