/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/ssargent/freyjadb/pkg/api"
	"github.com/ssargent/freyjadb/pkg/store"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the REST API server",
	Long: `Start the FreyjaDB REST API server with authentication.

Example:
  freyja serve --api-key=mysecretkey --port=8080`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")
		apiKey, _ := cmd.Flags().GetString("api-key")

		if apiKey == "" {
			cmd.Println("Error: --api-key is required")
			return
		}

		// Get store from context
		kv, ok := cmd.Context().Value("store").(*store.KVStore)
		if !ok {
			cmd.Println("Error: store not found in context")
			return
		}

		// Start API server
		config := api.ServerConfig{
			Port:   port,
			APIKey: apiKey,
		}

		if err := api.StartServer(kv, config); err != nil {
			cmd.Printf("Error starting server: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntP("port", "p", 8080, "Port to listen on")
	serveCmd.Flags().String("api-key", "", "API key for authentication (required)")
	serveCmd.MarkFlagRequired("api-key")
}
