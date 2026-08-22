package main

import (
	"github.com/spf13/cobra"
)

var groupCmd *cobra.Command

//nolint:gochecknoinits // cobra requires init() to register subcommands
func init() {
	factory := NewEntityCommandFactory()
	groupCmd = factory.CreateEntityCommands(
		EntityTypeGroup,
		"group",
		"Manage groups",
		`Create, read, update, and delete group entities.`,
	)
}
