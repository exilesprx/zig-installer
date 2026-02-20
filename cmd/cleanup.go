package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/exilesprx/zig-installer/internal/config"
	"github.com/exilesprx/zig-installer/internal/installer"
	"github.com/exilesprx/zig-installer/internal/tui"
	"github.com/spf13/cobra"
)

// CleanupCommand encapsulates the cleanup command
type CleanupCommand struct {
	cmd      *cobra.Command
	options  *CommandOptions
	rootCmd  *RootCommand
	dryRun   bool
	autoYes  bool
	keepLast int
}

// NewCleanupCommand creates a new cleanup command instance
func NewCleanupCommand(options *CommandOptions, rootCmd *RootCommand) *CleanupCommand {
	cc := &CleanupCommand{
		options: options,
		rootCmd: rootCmd,
	}

	cleanupCmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up old Zig versions",
		Long: `Interactively remove old Zig versions to free disk space.
Shows a list of installed versions and allows you to select which to remove.
The currently active version cannot be removed.`,
		Run: func(cmd *cobra.Command, args []string) {
			cfg, log, err := rootCmd.LoadLoggerAndConfig()
			styles := tui.LoadStyles()
			if err != nil {
				fmt.Printf("Error initializing: %v\n", err)
				os.Exit(1)
			}
			defer func() { _ = log.Close() }()

			// Ensure we're operating on user-local installation
			if !strings.Contains(cfg.ZigDir, ".local") {
				log.LogError("Cleanup only manages user-local installations")
				fmt.Println(tui.PrintWithStyles(
					fmt.Sprintf("Error: cleanup only manages user-local installations.\n\nExpected path: ~/.local/share/zig\nGot: %s", cfg.ZigDir),
					styles.Error, cfg.NoColor))
				os.Exit(1)
			}

			// Create formatter
			formatter := installer.NewTaskFormatter(cfg, styles)

			// Warn if system installation exists
			if systemDir, found := config.DetectSystemInstallation(); found {
				fmt.Println()
				formatter.PrintWarning("System Installation Detected",
					fmt.Sprintf("Found system installation at: %s", systemDir))
				formatter.PrintTask("Note", "Cleanup scope",
					"This command only manages user-local installations in ~/.local")
				formatter.PrintTask("Manual Removal", "If needed",
					fmt.Sprintf("To remove system installation:\n      sudo rm -rf %s /usr/local/bin/zig /usr/local/bin/zls", systemDir))
				fmt.Println()
			}

			// Run cleanup
			if err := installer.CleanupCommand(cfg, log, formatter, cc.dryRun, cc.autoYes, cc.keepLast); err != nil {
				log.LogError("Cleanup failed: %v", err)
				fmt.Println(styles.Error.Render(fmt.Sprintf("Error: %v", err)))
				os.Exit(1)
			}
		},
	}

	// Add flags
	cleanupCmd.Flags().BoolVar(&cc.dryRun, "dry-run", false, "Show what would be removed without actually removing")
	cleanupCmd.Flags().BoolVarP(&cc.autoYes, "yes", "y", false, "Skip confirmation prompts")
	cleanupCmd.Flags().IntVar(&cc.keepLast, "keep-last", 0, "Keep the last N versions (0 = interactive selection)")

	cc.cmd = cleanupCmd
	return cc
}
