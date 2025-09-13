/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/ssargent/freyjadb/pkg/config"
)

// serviceCmd represents the service command
var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage FreyjaDB as a systemd service",
	Long: `Manage FreyjaDB as a systemd service. This command provides
native integration with systemd for production deployments.

The service will be installed with proper security settings and
automatic restart on failure.`,
}

// installServiceCmd represents the service install command
var installServiceCmd = &cobra.Command{
	Use:   "install",
	Short: "Install FreyjaDB as a systemd service",
	Long: `Install FreyjaDB as a systemd service with proper configuration.

This will:
- Create or use existing configuration
- Generate systemd unit file
- Enable and optionally start the service

Examples:
  freyja service install
  freyja service install --data-dir /var/lib/freyjadb --user freyja`,
	Run: func(cmd *cobra.Command, args []string) {
		dataDir, _ := cmd.Flags().GetString("data-dir")
		configPath, _ := cmd.Flags().GetString("config")
		user, _ := cmd.Flags().GetString("user")
		port, _ := cmd.Flags().GetInt("port")
		startNow, _ := cmd.Flags().GetBool("start")

		// Use default config path if not specified
		if configPath == "" {
			configPath = config.GetDefaultConfigPath()
		}

		// Check if running as root (required for systemd operations)
		if os.Geteuid() != 0 {
			cmd.Printf("Error: service install requires root privileges\n")
			cmd.Printf("Run with: sudo freyja service install\n")
			os.Exit(1)
		}

		cmd.Printf("🔧 Installing FreyjaDB systemd service...\n")

		// Ensure config exists
		var cfg *config.Config
		var err error

		if config.ConfigExists(configPath) {
			cfg, err = config.LoadConfig(configPath)
			if err != nil {
				cmd.Printf("Error loading config: %v\n", err)
				os.Exit(1)
			}
			cmd.Printf("✅ Loaded existing configuration\n")
		} else {
			// Bootstrap config
			cfg, err = config.BootstrapConfig(configPath, dataDir)
			if err != nil {
				cmd.Printf("Error bootstrapping config: %v\n", err)
				os.Exit(1)
			}
			cmd.Printf("✅ Created new configuration at %s\n", configPath)
		}

		// Override config with flags
		if dataDir != "" {
			cfg.DataDir = dataDir
		}
		if port != 8080 {
			cfg.Port = port
		}

		// Save updated config
		if err := config.SaveConfig(cfg, configPath); err != nil {
			cmd.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}

		// Create systemd unit file
		if err := createSystemdUnit(cfg, configPath, user); err != nil {
			cmd.Printf("Error creating systemd unit: %v\n", err)
			os.Exit(1)
		}

		// Reload systemd
		if err := runSystemctlCommand("daemon-reload"); err != nil {
			cmd.Printf("Error reloading systemd: %v\n", err)
			os.Exit(1)
		}

		// Enable service
		if err := runSystemctlCommand("enable", "freyja.service"); err != nil {
			cmd.Printf("Error enabling service: %v\n", err)
			os.Exit(1)
		}

		cmd.Printf("✅ Service enabled successfully\n")

		// Start service if requested
		if startNow {
			if err := runSystemctlCommand("start", "freyja.service"); err != nil {
				cmd.Printf("Error starting service: %v\n", err)
				os.Exit(1)
			}
			cmd.Printf("✅ Service started successfully\n")
		}

		cmd.Printf("\n🎉 FreyjaDB service installed!\n")
		cmd.Printf("Service: freyja.service\n")
		cmd.Printf("Config: %s\n", configPath)
		cmd.Printf("Data: %s\n", cfg.DataDir)
		cmd.Printf("Port: %d\n", cfg.Port)

		if !startNow {
			cmd.Printf("\nTo start the service: sudo systemctl start freyja.service\n")
		}
		cmd.Printf("To check status: sudo systemctl status freyja.service\n")
		cmd.Printf("To view logs: sudo journalctl -u freyja.service -f\n")
	},
}

// startCmd represents the service start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the FreyjaDB service",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runSystemctlCommand("start", "freyja.service"); err != nil {
			cmd.Printf("Error starting service: %v\n", err)
			os.Exit(1)
		}
		cmd.Printf("✅ FreyjaDB service started\n")
	},
}

// stopCmd represents the service stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the FreyjaDB service",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runSystemctlCommand("stop", "freyja.service"); err != nil {
			cmd.Printf("Error stopping service: %v\n", err)
			os.Exit(1)
		}
		cmd.Printf("✅ FreyjaDB service stopped\n")
	},
}

// restartCmd represents the service restart command
var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the FreyjaDB service",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runSystemctlCommand("restart", "freyja.service"); err != nil {
			cmd.Printf("Error restarting service: %v\n", err)
			os.Exit(1)
		}
		cmd.Printf("✅ FreyjaDB service restarted\n")
	},
}

// statusCmd represents the service status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show FreyjaDB service status",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runSystemctlCommand("status", "freyja.service"); err != nil {
			cmd.Printf("Error getting service status: %v\n", err)
			os.Exit(1)
		}
	},
}

// logsCmd represents the service logs command
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show FreyjaDB service logs",
	Long: `Show FreyjaDB service logs using journalctl.

Examples:
  freyja service logs
  freyja service logs -f  # Follow logs`,
	Run: func(cmd *cobra.Command, args []string) {
		follow, _ := cmd.Flags().GetBool("follow")
		lines, _ := cmd.Flags().GetInt("lines")

		journalArgs := []string{"-u", "freyja.service"}
		if follow {
			journalArgs = append(journalArgs, "-f")
		}
		if lines > 0 {
			journalArgs = append(journalArgs, fmt.Sprintf("-n%d", lines))
		}

		if err := runCommand("journalctl", journalArgs...); err != nil {
			cmd.Printf("Error getting service logs: %v\n", err)
			os.Exit(1)
		}
	},
}

// uninstallCmd represents the service uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the FreyjaDB service",
	Run: func(cmd *cobra.Command, args []string) {
		// Check if running as root
		if os.Geteuid() != 0 {
			cmd.Printf("Error: service uninstall requires root privileges\n")
			cmd.Printf("Run with: sudo freyja service uninstall\n")
			os.Exit(1)
		}

		cmd.Printf("🗑️  Uninstalling FreyjaDB service...\n")

		// Stop service first
		_ = runSystemctlCommand("stop", "freyja.service") // Ignore errors if already stopped

		// Disable service
		if err := runSystemctlCommand("disable", "freyja.service"); err != nil {
			cmd.Printf("Warning: could not disable service: %v\n", err)
		}

		// Remove unit file
		unitPath := "/etc/systemd/system/freyja.service"
		if _, err := os.Stat(unitPath); err == nil {
			if err := os.Remove(unitPath); err != nil {
				cmd.Printf("Error removing unit file: %v\n", err)
				os.Exit(1)
			}
		}

		// Reload systemd
		if err := runSystemctlCommand("daemon-reload"); err != nil {
			cmd.Printf("Error reloading systemd: %v\n", err)
			os.Exit(1)
		}

		cmd.Printf("✅ FreyjaDB service uninstalled\n")
		cmd.Printf("Note: Configuration and data files were not removed\n")
	},
}

func init() {
	rootCmd.AddCommand(serviceCmd)

	// Add subcommands
	serviceCmd.AddCommand(installServiceCmd)
	serviceCmd.AddCommand(startCmd)
	serviceCmd.AddCommand(stopCmd)
	serviceCmd.AddCommand(restartCmd)
	serviceCmd.AddCommand(statusCmd)
	serviceCmd.AddCommand(logsCmd)
	serviceCmd.AddCommand(uninstallCmd)

	// Install command flags
	installServiceCmd.Flags().String("data-dir", "/var/lib/freyjadb", "Data directory for the service")
	installServiceCmd.Flags().String("config", "", "Path to config file")
	installServiceCmd.Flags().String("user", "freyja", "User to run the service as")
	installServiceCmd.Flags().Int("port", 8080, "Port for the service")
	installServiceCmd.Flags().Bool("start", true, "Start the service after installation")

	// Logs command flags
	logsCmd.Flags().BoolP("follow", "f", false, "Follow log output")
	logsCmd.Flags().IntP("lines", "n", 0, "Number of lines to show")
}

// createSystemdUnit creates the systemd unit file
func createSystemdUnit(cfg *config.Config, configPath, user string) error {
	unitContent := fmt.Sprintf(`[Unit]
Description=FreyjaDB Server
After=network-online.target
Wants=network-online.target

[Service]
User=%s
Group=%s
ExecStart=/usr/local/bin/freyja up --config %s
Restart=on-failure
NoNewPrivileges=true
UMask=0077
ReadWritePaths=%s
ReadWritePaths=%s

[Install]
WantedBy=multi-user.target
`, user, user, configPath, cfg.DataDir, filepath.Dir(configPath))

	unitPath := "/etc/systemd/system/freyja.service"
	return os.WriteFile(unitPath, []byte(unitContent), 0600)
}

// runSystemctlCommand runs a systemctl command
func runSystemctlCommand(args ...string) error {
	return runCommand("systemctl", args...)
}

// runCommand runs a system command and returns its error
func runCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
