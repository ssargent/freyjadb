/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/ssargent/freyjadb/cmd/freyja/cmd"
	"github.com/ssargent/freyjadb/pkg/di"
)

func main() {
	// Initialize dependency injection container
	container := di.NewContainer()

	// Inject dependencies into cmd package
	cmd.SetContainer(container)

	cmd.Execute()
}
