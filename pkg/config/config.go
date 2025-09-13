/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the FreyjaDB configuration
type Config struct {
	DataDir  string   `yaml:"data_dir"`
	Port     int      `yaml:"port"`
	Bind     string   `yaml:"bind"`
	Security Security `yaml:"security"`
	Logging  Logging  `yaml:"logging"`
}

// Security contains security-related configuration
type Security struct {
	SystemKey    string `yaml:"system_key"`
	SystemAPIKey string `yaml:"system_api_key"`
	ClientAPIKey string `yaml:"client_api_key"`
}

// Logging contains logging configuration
type Logging struct {
	Level string `yaml:"level"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		DataDir: "./data",
		Port:    8080,
		Bind:    "127.0.0.1",
		Security: Security{
			SystemKey:    "auto",
			SystemAPIKey: "auto",
			ClientAPIKey: "auto",
		},
		Logging: Logging{
			Level: "info",
		},
	}
}

// LoadConfig loads configuration from the specified path
func LoadConfig(configPath string) (*Config, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", configPath)
	}

	// Validate path to prevent directory traversal
	if !filepath.IsAbs(configPath) {
		absPath, err := filepath.Abs(configPath)
		if err != nil {
			return nil, fmt.Errorf("invalid config path: %w", err)
		}
		configPath = absPath
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves the configuration to the specified path with secure permissions
func SaveConfig(config *Config, configPath string) error {
	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write with secure permissions (0600)
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GenerateSecureKey generates a cryptographically secure random key
func GenerateSecureKey(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure key: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// BootstrapConfig creates a new configuration with generated keys if it doesn't exist
func BootstrapConfig(configPath string, dataDir string) (*Config, error) {
	config := DefaultConfig()
	if dataDir != "" {
		config.DataDir = dataDir
	}

	// Generate secure keys
	systemKey, err := GenerateSecureKey(32) // 256 bits
	if err != nil {
		return nil, fmt.Errorf("failed to generate system key: %w", err)
	}
	config.Security.SystemKey = systemKey

	systemAPIKey, err := GenerateSecureKey(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate system API key: %w", err)
	}
	config.Security.SystemAPIKey = systemAPIKey

	clientAPIKey, err := GenerateSecureKey(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate client API key: %w", err)
	}
	config.Security.ClientAPIKey = clientAPIKey

	// Save the configuration
	if err := SaveConfig(config, configPath); err != nil {
		return nil, fmt.Errorf("failed to save bootstrap config: %w", err)
	}

	return config, nil
}

// GetDefaultConfigPath returns the default configuration path for the current platform
func GetDefaultConfigPath() string {
	// Use OS-specific default locations
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./freyja.yaml"
	}

	// For Linux/macOS, use ~/.config/freyja/config.yaml
	configDir := filepath.Join(homeDir, ".config", "freyja")
	return filepath.Join(configDir, "config.yaml")
}

// ConfigExists checks if a configuration file exists
func ConfigExists(configPath string) bool {
	_, err := os.Stat(configPath)
	return !os.IsNotExist(err)
}
