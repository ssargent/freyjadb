package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var relationshipCmd = &cobra.Command{
	Use:   "relationship",
	Short: "Manage relationships between entities",
	Long:  `Create, list, and manage relationships between Lore entities.`,
}

var relationshipCreateCmd = &cobra.Command{
	Use:   "create <from_type>:<from_id> <relation> <to_type>:<to_id>",
	Short: "Create a relationship between two entities",
	Long: `Create a relationship between two Lore entities.

Examples:
  lore relationship create character:john-doe friend character:jane-smith
  lore relationship create character:john-doe located_in place:winterfell
  lore relationship create character:john-doe member_of group:stark-family`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		fromSpec := args[0]
		relation := args[1]
		toSpec := args[2]

		fromType, fromID, err := parseEntitySpec(fromSpec)
		if err != nil {
			return fmt.Errorf("invalid from entity: %w", err)
		}

		toType, toID, err := parseEntitySpec(toSpec)
		if err != nil {
			return fmt.Errorf("invalid to entity: %w", err)
		}

		// Validate that both entities exist
		if !loreStore.EntityExists(fromType, fromID) {
			return fmt.Errorf("source entity %s:%s does not exist", fromType, fromID)
		}
		if !loreStore.EntityExists(toType, toID) {
			return fmt.Errorf("target entity %s:%s does not exist", toType, toID)
		}

		// Create the relationship
		err = loreStore.PutRelationship(fromType, fromID, toType, toID, relation)
		if err != nil {
			return fmt.Errorf("failed to create relationship: %w", err)
		}

		if !config.Quiet {
			fmt.Printf("Created relationship: %s:%s --[%s]--> %s:%s\n",
				fromType, fromID, relation, toType, toID)
		}

		return nil
	},
}

var relationshipListCmd = &cobra.Command{
	Use:   "list <entity_type>:<entity_id>",
	Short: "List all relationships for an entity",
	Long: `List all relationships (incoming and outgoing) for a given entity.

Examples:
  lore relationship list character:john-doe
  lore relationship list place:winterfell`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		entitySpec := args[0]

		entityType, entityID, err := parseEntitySpec(entitySpec)
		if err != nil {
			return fmt.Errorf("invalid entity: %w", err)
		}

		// Get entity with relationships
		entityWithRels, err := loreStore.GetEntityWithRelationships(entityType, entityID)
		if err != nil {
			return err
		}

		return outputEntityWithRelationships(entityWithRels)
	},
}

var relationshipDeleteCmd = &cobra.Command{
	Use:   "delete <from_type>:<from_id> <relation> <to_type>:<to_id>",
	Short: "Delete a relationship between two entities",
	Long: `Delete a relationship between two Lore entities.

Examples:
  lore relationship delete character:john-doe friend character:jane-smith
  lore relationship delete character:john-doe located_in place:winterfell`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		fromSpec := args[0]
		relation := args[1]
		toSpec := args[2]

		fromType, fromID, err := parseEntitySpec(fromSpec)
		if err != nil {
			return fmt.Errorf("invalid from entity: %w", err)
		}

		toType, toID, err := parseEntitySpec(toSpec)
		if err != nil {
			return fmt.Errorf("invalid to entity: %w", err)
		}

		// Confirm deletion if not using --yes flag
		if !config.Yes {
			fmt.Printf("Are you sure you want to delete the relationship %s:%s --[%s]--> %s:%s? (y/N): ",
				fromType, fromID, relation, toType, toID)
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				fmt.Println("Deletion cancelled")
				return nil
			}
		}

		// Delete the relationship
		err = loreStore.DeleteRelationship(fromType, fromID, toType, toID, relation)
		if err != nil {
			return fmt.Errorf("failed to delete relationship: %w", err)
		}

		if !config.Quiet {
			fmt.Printf("Deleted relationship: %s:%s --[%s]--> %s:%s\n",
				fromType, fromID, relation, toType, toID)
		}

		return nil
	},
}

// parseEntitySpec parses an entity specification like "character:john-doe"
func parseEntitySpec(spec string) (EntityType, string, error) {
	parts := strings.SplitN(spec, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid entity specification: %s (expected format: type:id)", spec)
	}

	entityType := EntityType(parts[0])
	id := parts[1]

	// Validate entity type
	switch entityType {
	case EntityTypeCharacter, EntityTypePlace, EntityTypeGroup:
		return entityType, id, nil
	default:
		return "", "", fmt.Errorf("unknown entity type: %s", entityType)
	}
}

func init() {
	// Add subcommands
	relationshipCmd.AddCommand(relationshipCreateCmd)
	relationshipCmd.AddCommand(relationshipListCmd)
	relationshipCmd.AddCommand(relationshipDeleteCmd)
}
