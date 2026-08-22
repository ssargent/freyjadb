package main

import (
	"github.com/spf13/cobra"
)

var placeCmd *cobra.Command

//nolint:gochecknoinits // cobra requires init() to register subcommands
func init() {
	factory := NewEntityCommandFactory()
	placeCmd = factory.CreateEntityCommands(
		EntityTypePlace,
		"place",
		"Manage places",
		`Create, read, update, and delete place entities.`,
	)
}
