package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var characterCmd = &cobra.Command{
	Use:   "character",
	Short: "Manage characters",
	Long:  `Create, read, update, and delete character entities.`,
}

var characterCreateCmd = &cobra.Command{
	Use:   "create <name> [flags]",
	Short: "Create a new character",
	Long: `Create a new character with the specified name.

Examples:
  lore character create "John Doe" --summary "A brave knight"
  lore character create "Jane Smith" --aka "Lady J" --tags "noble,warrior"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Get flags
		summary, _ := cmd.Flags().GetString("summary")
		akaStr, _ := cmd.Flags().GetString("aka")
		tagsStr, _ := cmd.Flags().GetString("tags")
		details, _ := cmd.Flags().GetString("details")

		// Parse aka and tags
		var aka []string
		if akaStr != "" {
			aka = strings.Split(akaStr, ",")
			for i, a := range aka {
				aka[i] = strings.TrimSpace(a)
			}
		}

		var tags []string
		if tagsStr != "" {
			tags = strings.Split(tagsStr, ",")
			for i, t := range tags {
				tags[i] = strings.TrimSpace(t)
			}
		}

		// Create entity
		entity := NewEntity(EntityTypeCharacter, name)
		entity.Summary = summary
		entity.Aka = aka
		entity.Tags = tags
		entity.Details = details

		// Store entity
		if err := loreStore.PutEntity(entity); err != nil {
			return fmt.Errorf("failed to create character: %w", err)
		}

		if !config.Quiet {
			fmt.Printf("Created character '%s' with ID '%s'\n", name, entity.ID)
		}

		return nil
	},
}

var characterGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a character by ID",
	Long:  `Retrieve and display a character by its ID.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		entity, err := loreStore.GetEntity(EntityTypeCharacter, id)
		if err != nil {
			return err
		}

		return outputEntity(entity)
	},
}

var characterListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all characters",
	Long:  `List all characters in the project.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		entities, err := loreStore.ListEntities(EntityTypeCharacter)
		if err != nil {
			return err
		}

		return outputEntities(entities)
	},
}

var characterUpdateCmd = &cobra.Command{
	Use:   "update <id> [flags]",
	Short: "Update a character",
	Long: `Update an existing character with new information.

Examples:
  lore character update john-doe --summary "A legendary knight"
  lore character update jane-smith --tags "noble,warrior,mage"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		// Get existing entity
		entity, err := loreStore.GetEntity(EntityTypeCharacter, id)
		if err != nil {
			return err
		}

		// Get flags and update entity
		if summary, _ := cmd.Flags().GetString("summary"); summary != "" {
			entity.Summary = summary
		}
		if akaStr, _ := cmd.Flags().GetString("aka"); akaStr != "" {
			aka := strings.Split(akaStr, ",")
			for i, a := range aka {
				aka[i] = strings.TrimSpace(a)
			}
			entity.Aka = aka
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
			return fmt.Errorf("failed to update character: %w", err)
		}

		if !config.Quiet {
			fmt.Printf("Updated character '%s'\n", id)
		}

		return nil
	},
}

var characterDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a character",
	Long:  `Delete a character by its ID.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		// Confirm deletion if not using --yes flag
		if !config.Yes {
			fmt.Printf("Are you sure you want to delete character '%s'? (y/N): ", id)
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				fmt.Println("Deletion cancelled")
				return nil
			}
		}

		if err := loreStore.DeleteEntity(EntityTypeCharacter, id); err != nil {
			return err
		}

		if !config.Quiet {
			fmt.Printf("Deleted character '%s'\n", id)
		}

		return nil
	},
}

func init() {
	// Add flags to create command
	characterCreateCmd.Flags().String("summary", "", "Character summary")
	characterCreateCmd.Flags().String("aka", "", "Alternative names (comma-separated)")
	characterCreateCmd.Flags().String("tags", "", "Tags (comma-separated)")
	characterCreateCmd.Flags().String("details", "", "Detailed description")

	// Add flags to update command
	characterUpdateCmd.Flags().String("summary", "", "Character summary")
	characterUpdateCmd.Flags().String("aka", "", "Alternative names (comma-separated)")
	characterUpdateCmd.Flags().String("tags", "", "Tags (comma-separated)")
	characterUpdateCmd.Flags().String("details", "", "Detailed description")

	// Add subcommands
	characterCmd.AddCommand(characterCreateCmd)
	characterCmd.AddCommand(characterGetCmd)
	characterCmd.AddCommand(characterListCmd)
	characterCmd.AddCommand(characterUpdateCmd)
	characterCmd.AddCommand(characterDeleteCmd)
}
