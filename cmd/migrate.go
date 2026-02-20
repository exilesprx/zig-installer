package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/exilesprx/zig-install/internal/config"
	"github.com/exilesprx/zig-install/internal/installer"
	"github.com/exilesprx/zig-install/internal/tui"
	"github.com/spf13/cobra"
)

// MigrateCommand represents the migrate command
type MigrateCommand struct {
	options *CommandOptions
	rootCmd *RootCommand
}

// NewMigrateCommand creates a new migrate command
func NewMigrateCommand(options *CommandOptions, rootCmd *RootCommand) *MigrateCommand {
	return &MigrateCommand{
		options: options,
		rootCmd: rootCmd,
	}
}

// Command returns the cobra command
func (mc *MigrateCommand) Command() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Migrate system-wide installation to user-local",
		Long: `Detect and migrate existing system-wide Zig installation to user-local.

zig-installer v4.0.0+ uses user-local installation only (~/.local).

This command:
  1. Detects system installation in /opt/zig or /usr/local/zig
  2. Removes the system installation (may require sudo password)
  3. Removes system symlinks from /usr/local/bin
  4. Prepares for user-local installation

After migration, run: zig-installer install`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config and logger
			cfg, log, err := mc.rootCmd.LoadLoggerAndConfig()
			if err != nil {
				return err
			}

			// Check NOT root
			if os.Geteuid() == 0 {
				return fmt.Errorf("do not run 'migrate' with sudo.\n\nRun as regular user: ./zig-installer migrate\nYou will be prompted for sudo password if needed")
			}

			// macOS warning
			if runtime.GOOS == "darwin" {
				fmt.Println("⚠️  Warning: macOS support is experimental")
				fmt.Println()
			}

			// Create formatter
			styles := tui.LoadStyles()
			formatter := installer.NewTaskFormatter(cfg, styles)

			formatter.PrintSection("System to User-Local Migration")

			// Detect system installation
			systemDir, found := config.DetectSystemInstallation()
			if !found {
				formatter.PrintTask("Detection", "Not Found", "No system-wide installation detected")
				fmt.Println()
				formatter.PrintTask("Next Step", "Install user-local",
					"Run: zig-installer install")
				return nil
			}

			formatter.PrintSuccess("Detection", fmt.Sprintf("Found system installation: %s", systemDir))
			fmt.Println()

			// Perform migration
			formatter.PrintSection("Removing System Installation")
			if err := installer.PerformMigration(systemDir, formatter, log); err != nil {
				fmt.Println()
				formatter.PrintError("Migration Failed", fmt.Sprintf("%v", err))

				// If permission error, provide manual instructions
				if os.IsPermission(err) {
					fmt.Println()
					formatter.PrintTask("Manual Removal Required", "Run these commands",
						fmt.Sprintf("sudo rm -rf %s /opt/zls /usr/local/zls\nsudo rm -f /usr/local/bin/zig /usr/local/bin/zls", systemDir))
					fmt.Println()
					formatter.PrintTask("Then", "Install user-local",
						"zig-installer install")
				}

				return err
			}

			fmt.Println()
			formatter.PrintSuccess("Migration Complete", "System installation removed successfully")
			fmt.Println()

			formatter.PrintTask("Next Step", "Install user-local version",
				"Run: zig-installer install")

			return nil
		},
	}
}
