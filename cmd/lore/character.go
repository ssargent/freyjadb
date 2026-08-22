package main

import (
	"github.com/spf13/cobra"
)

var characterCmd *cobra.Command

//nolint:gochecknoinits // cobra requires init() to register subcommands
func init() {
	factory := NewEntityCommandFactory()
	characterCmd = factory.CreateEntityCommands(
		EntityTypeCharacter,
		"character",
		"Manage characters",
		`Create, read, update, and delete character entities.`,
	)

	// Add character-specific flags
	characterCreateCmd := characterCmd.Commands()[0] // create command
	characterCreateCmd.Flags().String("aka", "", "Alternative names (comma-separated)")

	characterUpdateCmd := characterCmd.Commands()[3] // update command
	characterUpdateCmd.Flags().String("aka", "", "Alternative names (comma-separated)")
}
