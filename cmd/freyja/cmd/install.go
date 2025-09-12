/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install freyja as a systemd service",
	Long: `Install freyja as a systemd service on Linux systems.

This command will:
- Check if running as root (required for installation)
- Stop any existing freyja service
- Build and install the latest binary
- Create systemd service configuration
- Enable and start the service

Example:
  sudo freyja install --api-key=mysecretkey --data-dir=/opt/mythicalcodelabs/freyja/data`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip the root command's store initialization for install command
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Check if running as root
		if os.Geteuid() != 0 {
			cmd.Printf("Error: freyja install must be run as root (sudo)\n")
			cmd.Printf("Usage: sudo freyja install [flags]\n")
			os.Exit(1)
		}

		// Get flags
		dataDir, _ := cmd.Flags().GetString("data-dir")
		apiKey, _ := cmd.Flags().GetString("api-key")
		systemKey, _ := cmd.Flags().GetString("system-key")
		port, _ := cmd.Flags().GetInt("port")
		force, _ := cmd.Flags().GetBool("force")

		if apiKey == "" {
			cmd.Printf("Error: --api-key is required\n")
			os.Exit(1)
		}

		if systemKey == "" {
			cmd.Printf("Error: --system-key is required for system administration\n")
			os.Exit(1)
		}

		cmd.Printf("Starting freyja installation...\n")

		// Create data directory
		if err := createDataDirectory(dataDir); err != nil {
			cmd.Printf("Error creating data directory: %v\n", err)
			os.Exit(1)
		}

		// Check if service is running
		isRunning, err := isServiceRunning()
		if err != nil {
			cmd.Printf("Warning: Could not check service status: %v\n", err)
		}

		// Stop service if running
		if isRunning {
			cmd.Printf("Stopping existing freyja service...\n")
			if err := stopService(); err != nil {
				cmd.Printf("Error stopping service: %v\n", err)
				if !force {
					os.Exit(1)
				}
			}
		}

		// Build and install binary
		if err := buildAndInstallBinary(); err != nil {
			cmd.Printf("Error building/installing binary: %v\n", err)
			os.Exit(1)
		}

		// Create systemd service
		if err := createSystemdService(dataDir, apiKey, systemKey, port); err != nil {
			cmd.Printf("Error creating systemd service: %v\n", err)
			os.Exit(1)
		}

		// Reload systemd
		if err := reloadSystemd(); err != nil {
			cmd.Printf("Error reloading systemd: %v\n", err)
			os.Exit(1)
		}

		// Enable and start service
		if err := enableAndStartService(); err != nil {
			cmd.Printf("Error enabling/starting service: %v\n", err)
			os.Exit(1)
		}

		cmd.Printf("✅ freyja installation completed successfully!\n")
		cmd.Printf("Service is running and will start automatically on boot.\n")
		cmd.Printf("Data directory: %s\n", dataDir)
		cmd.Printf("API endpoint: http://localhost:%d\n", port)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)

	installCmd.Flags().String("data-dir", "/opt/mythicalcodelabs/freyja/data", "Data directory for freyja")
	installCmd.Flags().String("api-key", "", "API key for user authentication (required)")
	installCmd.Flags().String("system-key", "", "System API key for administrative operations (required)")
	installCmd.Flags().Int("port", 8080, "Port for the API server")
	installCmd.Flags().Bool("force", false, "Force reinstall even if errors occur")
	installCmd.MarkFlagRequired("api-key")
	installCmd.MarkFlagRequired("system-key")
}

// createDataDirectory creates the data directory with proper permissions
func createDataDirectory(dataDir string) error {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory %s: %w", dataDir, err)
	}

	// Change ownership to freyja user if it exists, otherwise keep as root
	if _, err := exec.LookPath("id"); err == nil {
		if err := exec.Command("chown", "-R", "freyja:freyja", dataDir).Run(); err != nil {
			// If freyja user doesn't exist, that's okay - keep as root
			fmt.Printf("Warning: Could not change ownership to freyja user: %v\n", err)
		}
	}

	return nil
}

// isServiceRunning checks if the freyja service is currently running
func isServiceRunning() (bool, error) {
	cmd := exec.Command("systemctl", "is-active", "freyja")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	status := strings.TrimSpace(string(output))
	return status == "active", nil
}

// stopService stops the freyja service
func stopService() error {
	cmd := exec.Command("systemctl", "stop", "freyja")
	return cmd.Run()
}

// buildAndInstallBinary builds the latest binary and installs it
func buildAndInstallBinary() error {
	// Build Linux binary
	fmt.Printf("Building freyja binary...\n")
	buildCmd := exec.Command("make", "build-linux")
	buildCmd.Dir = "/Users/scott/source/github/ssargent/freyjadb" // Adjust path as needed
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("failed to build binary: %w", err)
	}

	// Install binary
	fmt.Printf("Installing binary to /usr/local/bin...\n")
	installCmd := exec.Command("cp", "bin/freyja_unix", "/usr/local/bin/freyja")
	installCmd.Dir = "/Users/scott/source/github/ssargent/freyjadb" // Adjust path as needed
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install binary: %w", err)
	}

	// Make executable
	if err := exec.Command("chmod", "+x", "/usr/local/bin/freyja").Run(); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	return nil
}

// createSystemdService creates the systemd service file
func createSystemdService(dataDir, apiKey, systemKey string, port int) error {
	serviceContent := fmt.Sprintf(`[Unit]
Description=FreyjaDB Key-Value Store
After=network.target

[Service]
Type=simple
User=freyja
Environment=DATA_DIR=%s
Environment=SYSTEM_KEY=%s
ExecStart=/usr/local/bin/freyja serve --data-dir=${DATA_DIR} --api-key=%s --system-key=${SYSTEM_KEY} --port=%d --enable-encryption
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
`, dataDir, systemKey, apiKey, port)

	servicePath := "/etc/systemd/system/freyja.service"
	file, err := os.Create(servicePath)
	if err != nil {
		return fmt.Errorf("failed to create service file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(serviceContent); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	fmt.Printf("Created systemd service file: %s\n", servicePath)
	return nil
}

// reloadSystemd reloads the systemd daemon
func reloadSystemd() error {
	cmd := exec.Command("systemctl", "daemon-reload")
	return cmd.Run()
}

// enableAndStartService enables and starts the freyja service
func enableAndStartService() error {
	// Enable service
	fmt.Printf("Enabling freyja service...\n")
	if err := exec.Command("systemctl", "enable", "freyja").Run(); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}

	// Start service
	fmt.Printf("Starting freyja service...\n")
	if err := exec.Command("systemctl", "start", "freyja").Run(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	return nil
}
