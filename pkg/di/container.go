// Package di provides dependency injection container
package di

import (
	"github.com/ssargent/freyjadb/pkg/api" //nolint:depguard
)

// Container holds all the dependencies for the application
type Container struct {
	systemServiceFactory api.SystemServiceFactory
	serverFactory        api.ServerFactory
}

// NewContainer creates a new dependency injection container
func NewContainer() *Container {
	return &Container{
		systemServiceFactory: api.NewSystemServiceFactory(),
		serverFactory:        api.NewServerFactory(),
	}
}

// GetSystemServiceFactory returns the system service factory
func (c *Container) GetSystemServiceFactory() api.SystemServiceFactory {
	return c.systemServiceFactory
}

// GetServerFactory returns the server factory
func (c *Container) GetServerFactory() api.ServerFactory {
	return c.serverFactory
}

// SetSystemServiceFactory allows overriding the system service factory (for testing)
func (c *Container) SetSystemServiceFactory(factory api.SystemServiceFactory) {
	c.systemServiceFactory = factory
}
