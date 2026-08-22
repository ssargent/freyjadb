package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// EntityCommandFactory creates entity commands for different types
type EntityCommandFactory struct{}

// NewEntityCommandFactory creates a new factory
func NewEntityCommandFactory() *EntityCommandFactory {
	return &EntityCommandFactory{}
}

// parseTagsFlag splits a comma-separated tags flag value into a trimmed slice.
func parseTagsFlag(tagsStr string) []string {
	if tagsStr == "" {
		return nil
	}
	tags := strings.Split(tagsStr, ",")
	for i, t := range tags {
		tags[i] = strings.TrimSpace(t)
	}
	return tags
}

// runCreateEntity implements the "create" subcommand for an entity type.
func runCreateEntity(cmd *cobra.Command, args []string, entityType EntityType, use string) error {
	name := args[0]

	summary, _ := cmd.Flags().GetString("summary")
	tagsStr, _ := cmd.Flags().GetString("tags")
	details, _ := cmd.Flags().GetString("details")

	entity := NewEntity(entityType, name)
	entity.Summary = summary
	entity.Tags = parseTagsFlag(tagsStr)
	entity.Details = details

	if err := loreStore.PutEntity(entity); err != nil {
		return fmt.Errorf("failed to create %s: %w", use, err)
	}
	if !config.Quiet {
		fmt.Printf("Created %s '%s' with ID '%s'\n", use, name, entity.ID)
	}
	return nil
}

// runGetEntity implements the "get" subcommand for an entity type.
func runGetEntity(args []string, entityType EntityType) error {
	id := args[0]
	entity, err := loreStore.GetEntity(entityType, id)
	if err != nil {
		return err
	}
	return outputEntity(entity)
}

// runListEntities implements the "list" subcommand for an entity type.
func runListEntities(entityType EntityType) error {
	entities, err := loreStore.ListEntities(entityType)
	if err != nil {
		return err
	}
	return outputEntities(entities)
}

// runUpdateEntity implements the "update" subcommand for an entity type.
func runUpdateEntity(cmd *cobra.Command, args []string, entityType EntityType, use string) error {
	id := args[0]
	entity, err := loreStore.GetEntity(entityType, id)
	if err != nil {
		return err
	}

	summary, _ := cmd.Flags().GetString("summary")
	tagsStr, _ := cmd.Flags().GetString("tags")
	details, _ := cmd.Flags().GetString("details")

	if cmd.Flags().Changed("summary") {
		entity.Summary = summary
	}
	if cmd.Flags().Changed("details") {
		entity.Details = details
	}
	if cmd.Flags().Changed("tags") && tagsStr != "" {
		entity.Tags = parseTagsFlag(tagsStr)
	}

	if err := loreStore.PutEntity(entity); err != nil {
		return fmt.Errorf("failed to update %s: %w", use, err)
	}
	if !config.Quiet {
		fmt.Printf("Updated %s '%s'\n", use, entity.Name)
	}
	return nil
}

// runDeleteEntity implements the "delete" subcommand for an entity type.
func runDeleteEntity(args []string, entityType EntityType, use string) error {
	id := args[0]
	entity, err := loreStore.GetEntity(entityType, id)
	if err != nil {
		return err
	}

	if err := loreStore.DeleteEntity(entityType, id); err != nil {
		return fmt.Errorf("failed to delete %s: %w", use, err)
	}
	if !config.Quiet {
		fmt.Printf("Deleted %s '%s' (%s)\n", use, entity.Name, id)
	}
	return nil
}

// CreateEntityCommands creates all CRUD commands for an entity type
func (f *EntityCommandFactory) CreateEntityCommands(entityType EntityType, use, short, long string) *cobra.Command {
	entityCmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
	}

	createCmd := &cobra.Command{
		Use:   "create <name> [flags]",
		Short: fmt.Sprintf("Create a new %s", use),
		Long: fmt.Sprintf(`Create a new %s with the specified name.
Examples:
  lore %s create "Example %s" --summary "Description of the %s" --tags "tag1,tag2"`,
			use, use, use, use),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreateEntity(cmd, args, entityType, use)
		},
	}

	getCmd := &cobra.Command{
		Use:   "get <id>",
		Short: fmt.Sprintf("Get a %s by ID", use),
		Long:  fmt.Sprintf("Retrieve and display a %s by its ID.", use),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetEntity(args, entityType)
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("List all %ss", use),
		Long:  fmt.Sprintf("List all %ss in the lore database.", use),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListEntities(entityType)
		},
	}

	updateCmd := &cobra.Command{
		Use:   "update <id> [flags]",
		Short: fmt.Sprintf("Update a %s", use),
		Long:  fmt.Sprintf("Update an existing %s by ID.", use),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdateEntity(cmd, args, entityType, use)
		},
	}

	deleteCmd := &cobra.Command{
		Use:   "delete <id>",
		Short: fmt.Sprintf("Delete a %s", use),
		Long:  fmt.Sprintf("Delete a %s by its ID.", use),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeleteEntity(args, entityType, use)
		},
	}

	// Add flags to create command
	createCmd.Flags().String("summary", "", fmt.Sprintf("%s summary", use))
	createCmd.Flags().String("tags", "", "Tags (comma-separated)")
	createCmd.Flags().String("details", "", "Detailed description")

	// Add flags to update command
	updateCmd.Flags().String("summary", "", fmt.Sprintf("%s summary", use))
	updateCmd.Flags().String("tags", "", "Tags (comma-separated)")
	updateCmd.Flags().String("details", "", "Detailed description")

	// Add subcommands
	entityCmd.AddCommand(createCmd)
	entityCmd.AddCommand(getCmd)
	entityCmd.AddCommand(listCmd)
	entityCmd.AddCommand(updateCmd)
	entityCmd.AddCommand(deleteCmd)

	return entityCmd
}
