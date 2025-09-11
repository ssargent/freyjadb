package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "Manage groups",
	Long:  `Create, read, update, and delete group entities.`,
}

var groupCreateCmd = &cobra.Command{
	Use:   "create <name> [flags]",
	Short: "Create a new group",
	Long: `Create a new group with the specified name.

Examples:
  lore group create "House Stark" --summary "An ancient noble house from the North"
  lore group create "Night's Watch" --tags "military,order"`,
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
		entity := NewEntity(EntityTypeGroup, name)
		entity.Summary = summary
		entity.Tags = tags
		entity.Details = details

		// Store entity
		if err := loreStore.PutEntity(entity); err != nil {
			return fmt.Errorf("failed to create group: %w", err)
		}

		if !config.Quiet {
			fmt.Printf("Created group '%s' with ID '%s'\n", name, entity.ID)
		}

		return nil
	},
}

var groupGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a group by ID",
	Long:  `Retrieve and display a group by its ID.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		entity, err := loreStore.GetEntity(EntityTypeGroup, id)
		if err != nil {
			return err
		}

		return outputEntity(entity)
	},
}

var groupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all groups",
	Long:  `List all groups in the project.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		entities, err := loreStore.ListEntities(EntityTypeGroup)
		if err != nil {
			return err
		}

		return outputEntities(entities)
	},
}

var groupUpdateCmd = &cobra.Command{
	Use:   "update <id> [flags]",
	Short: "Update a group",
	Long: `Update an existing group with new information.

Examples:
  lore group update house-stark --summary "The ancient and honorable House Stark"
  lore group update nights-watch --tags "military,order,defense"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		// Get existing entity
		entity, err := loreStore.GetEntity(EntityTypeGroup, id)
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
			return fmt.Errorf("failed to update group: %w", err)
		}

		if !config.Quiet {
			fmt.Printf("Updated group '%s'\n", id)
		}

		return nil
	},
}

var groupDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a group",
	Long:  `Delete a group by its ID.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		// Confirm deletion if not using --yes flag
		if !config.Yes {
			fmt.Printf("Are you sure you want to delete group '%s'? (y/N): ", id)
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				fmt.Println("Deletion cancelled")
				return nil
			}
		}

		if err := loreStore.DeleteEntity(EntityTypeGroup, id); err != nil {
			return err
		}

		if !config.Quiet {
			fmt.Printf("Deleted group '%s'\n", id)
		}

		return nil
	},
}

func init() {
	// Add flags to create command
	groupCreateCmd.Flags().String("summary", "", "Group summary")
	groupCreateCmd.Flags().String("tags", "", "Tags (comma-separated)")
	groupCreateCmd.Flags().String("details", "", "Detailed description")

	// Add flags to update command
	groupUpdateCmd.Flags().String("summary", "", "Group summary")
	groupUpdateCmd.Flags().String("tags", "", "Tags (comma-separated)")
	groupUpdateCmd.Flags().String("details", "", "Detailed description")

	// Add subcommands
	groupCmd.AddCommand(groupCreateCmd)
	groupCmd.AddCommand(groupGetCmd)
	groupCmd.AddCommand(groupListCmd)
	groupCmd.AddCommand(groupUpdateCmd)
	groupCmd.AddCommand(groupDeleteCmd)
}
