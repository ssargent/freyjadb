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

		// Get store from context
		kv, ok := cmd.Context().Value("store").(*store.KVStore)
		if !ok {
			fmt.Printf("Error: store not found in context\n")
			return
		}

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
}
