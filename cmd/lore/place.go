package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var placeCmd = &cobra.Command{
	Use:   "place",
	Short: "Manage places",
	Long:  `Create, read, update, and delete place entities.`,
}

var placeCreateCmd = &cobra.Command{
	Use:   "create <name> [flags]",
	Short: "Create a new place",
	Long: `Create a new place with the specified name.

Examples:
  lore place create "Winterfell" --summary "A great castle in the North"
  lore place create "King's Landing" --tags "capital,city"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Get flags
		summary, _ := cmd.Flags().GetString("summary")
		tagsStr, _ := cmd.Flags().GetString("tags")
		details, _ := cmd.Flags().GetString("details")

		var tags []string
		if tagsStr != "" {
			tags = strings.Split(tagsStr, ",")
			for i, t := range tags {
				tags[i] = strings.TrimSpace(t)
			}
		}

		// Create entity
		entity := NewEntity(EntityTypePlace, name)
		entity.Summary = summary
		entity.Tags = tags
		entity.Details = details

		// Store entity
		if err := loreStore.PutEntity(entity); err != nil {
			return fmt.Errorf("failed to create place: %w", err)
		}

		if !config.Quiet {
			fmt.Printf("Created place '%s' with ID '%s'\n", name, entity.ID)
		}

		return nil
	},
}

var placeGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a place by ID",
	Long:  `Retrieve and display a place by its ID.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		entity, err := loreStore.GetEntity(EntityTypePlace, id)
		if err != nil {
			return err
		}

		return outputEntity(entity)
	},
}

var placeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all places",
	Long:  `List all places in the project.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		entities, err := loreStore.ListEntities(EntityTypePlace)
		if err != nil {
			return err
		}

		return outputEntities(entities)
	},
}

var placeUpdateCmd = &cobra.Command{
	Use:   "update <id> [flags]",
	Short: "Update a place",
	Long: `Update an existing place with new information.

Examples:
  lore place update winterfell --summary "The ancient seat of House Stark"
  lore place update kings-landing --tags "capital,city,port"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		// Get existing entity
		entity, err := loreStore.GetEntity(EntityTypePlace, id)
		if err != nil {
			return err
		}

		// Get flags and update entity
		if summary, _ := cmd.Flags().GetString("summary"); summary != "" {
			entity.Summary = summary
		}
		if tagsStr, _ := cmd.Flags().GetString("tags"); tagsStr != "" {
			tags := strings.Split(tagsStr, ",")
			for i, t := range tags {
				tags[i] = strings.TrimSpace(t)
			}
			entity.Tags = tags
		}
		if details, _ := cmd.Flags().GetString("details"); details != "" {
			entity.Details = details
		}

		// Store updated entity
		if err := loreStore.PutEntity(entity); err != nil {
			return fmt.Errorf("failed to update place: %w", err)
		}

		if !config.Quiet {
			fmt.Printf("Updated place '%s'\n", id)
		}

		return nil
	},
}

var placeDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a place",
	Long:  `Delete a place by its ID.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		// Confirm deletion if not using --yes flag
		if !config.Yes {
			fmt.Printf("Are you sure you want to delete place '%s'? (y/N): ", id)
			var response string
			n, err := fmt.Scanln(&response)
			if err != nil || n != 1 {
				return fmt.Errorf("failed to read input: %w", err)
			}
			if strings.ToLower(response) != confirmYes && strings.ToLower(response) != confirmYesLong {
				fmt.Println("Deletion cancelled")
				return nil
			}
		}

		if err := loreStore.DeleteEntity(EntityTypePlace, id); err != nil {
			return err
		}

		if !config.Quiet {
			fmt.Printf("Deleted place '%s'\n", id)
		}

		return nil
	},
}

func setupPlaceCommands() {
	// Add flags to create command
	placeCreateCmd.Flags().String("summary", "", "Place summary")
	placeCreateCmd.Flags().String("tags", "", "Tags (comma-separated)")
	placeCreateCmd.Flags().String("details", "", "Detailed description")

	// Add flags to update command
	placeUpdateCmd.Flags().String("summary", "", "Place summary")
	placeUpdateCmd.Flags().String("tags", "", "Tags (comma-separated)")
	placeUpdateCmd.Flags().String("details", "", "Detailed description")

	// Add subcommands
	placeCmd.AddCommand(placeCreateCmd)
	placeCmd.AddCommand(placeGetCmd)
	placeCmd.AddCommand(placeListCmd)
	placeCmd.AddCommand(placeUpdateCmd)
	placeCmd.AddCommand(placeDeleteCmd)
}
