package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/exilesprx/zig-installer/internal/config"
	"github.com/exilesprx/zig-installer/internal/installer"
	"github.com/exilesprx/zig-installer/internal/tui"
	"github.com/spf13/cobra"
)

// MigrateCommand represents the migrate command
type MigrateCommand struct {
	options *CommandOptions
	rootCmd *RootCommand
	dryRun  bool
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
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate system-wide installation to user-local",
		Long: `Detect and migrate existing system-wide Zig installation to user-local.

zig-installer v4.0.0+ uses user-local installation only (~/.local).

This command:
  1. Detects system installation in /opt/zig or /usr/local/zig
  2. Moves Zig and ZLS installations to ~/.local/share (preserving all versions)
  3. Recreates symlinks in ~/.local/bin
  4. Removes old system symlinks from /usr/local/bin
  5. Cleans up empty system directories

After migration, your existing installations will be ready to use immediately.

Flags:
  --dry-run    Show what would be migrated without making changes`,
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

			if mc.dryRun {
				formatter.PrintSection("System to User-Local Migration [DRY RUN]")
			} else {
				formatter.PrintSection("System to User-Local Migration")
			}

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
			if mc.dryRun {
				formatter.PrintSection("Migration Preview [DRY RUN]")
			} else {
				formatter.PrintSection("Migrating Installation")
			}

			if err := installer.PerformMigration(systemDir, formatter, log, mc.dryRun); err != nil {
				fmt.Println()
				formatter.PrintError("Migration Failed", fmt.Sprintf("%v", err))

				// If permission error, provide manual instructions
				if os.IsPermission(err) {
					fmt.Println()
					formatter.PrintTask("Manual Migration Steps", "Run these commands",
						fmt.Sprintf("Migration requires sudo access to move files from system directories.\n\n"+
							"To manually migrate, run:\n"+
							"  sudo mv /opt/zig/zig-* ~/.local/share/zig/\n"+
							"  sudo mv /opt/zls ~/.local/share/zls/\n"+
							"  sudo rm -f /usr/local/bin/zig /usr/local/bin/zls\n\n"+
							"Or to remove system installations instead:\n"+
							"  sudo rm -rf %s /opt/zls /usr/local/zls\n"+
							"  sudo rm -f /usr/local/bin/zig /usr/local/bin/zls\n\n"+
							"Then run: zig-installer install", systemDir))
				}

				return err
			}

			if mc.dryRun {
				fmt.Println()
				formatter.PrintSuccess("Dry Run Complete", "No changes were made")
				fmt.Println()
				formatter.PrintTask("To apply changes", "Run without --dry-run",
					"Run: zig-installer migrate")
			} else {
				fmt.Println()
				formatter.PrintSuccess("Migration Complete", "System installation migrated to user-local successfully")
				fmt.Println()

				formatter.PrintTask("Ready to use", "Migration complete",
					"Your Zig installation is ready. Run 'zig version' to verify.")
			}

			return nil
		},
	}

	// Add dry-run flag
	cmd.Flags().BoolVar(&mc.dryRun, "dry-run", false, "Show what would be migrated without making any changes")

	return cmd
}
