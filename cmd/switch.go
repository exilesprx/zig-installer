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

// SwitchCommand encapsulates the switch command
type SwitchCommand struct {
	cmd     *cobra.Command
	options *CommandOptions
	rootCmd *RootCommand
	version string
}

// NewSwitchCommand creates a new switch command instance
func NewSwitchCommand(options *CommandOptions, rootCmd *RootCommand) *SwitchCommand {
	sc := &SwitchCommand{
		options: options,
		rootCmd: rootCmd,
	}

	switchCmd := &cobra.Command{
		Use:   "switch [version]",
		Short: "Switch between installed Zig versions",
		Long: `Switch between installed Zig versions by updating the symlink.

This command allows you to quickly switch between different Zig versions 
that you have already installed. It works by updating the symlink in 
~/.local/bin/zig to point to the selected version.

Usage:
  # Interactive selection - shows list of installed versions
  zig-installer switch

  # Switch to specific version directly
  zig-installer switch 0.13.0

Requirements:
  • Multiple Zig versions must be installed
  • Only works with user-local installations (~/.local)

Note: This only affects the Zig binary. ZLS versions are managed separately.
If you need a different ZLS version, reinstall it with the matching Zig version.`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Load configuration and logger
			cfg, log, err := rootCmd.LoadLoggerAndConfig()
			styles := tui.LoadStyles()
			if err != nil {
				fmt.Printf("Error initializing: %v\n", err)
				os.Exit(1)
			}
			defer func() { _ = log.Close() }()

			// Ensure we're operating on user-local installation
			if !strings.Contains(cfg.ZigDir, ".local") {
				log.LogError("Switch only works with user-local installations")
				fmt.Println(tui.PrintWithStyles(
					fmt.Sprintf("Error: switch only works with user-local installations.\n\nExpected path: ~/.local/share/zig\nGot: %s", cfg.ZigDir),
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
				formatter.PrintTask("Note", "Switch scope",
					"This command only affects user-local installations in ~/.local")
				fmt.Println()
			}

			// Determine target version
			var targetVersion string

			// Check if version provided as positional argument
			if len(args) > 0 {
				targetVersion = args[0]
			} else if sc.version != "" {
				// Check if provided via --version flag
				targetVersion = sc.version
			}

			// If no version specified, prompt interactively
			if targetVersion == "" {
				// Scan installed versions first
				versions, err := installer.ScanInstalledVersions(cfg.ZigDir, cfg.BinDir)
				if err != nil {
					log.LogError("Failed to scan versions: %v", err)
					fmt.Println(styles.Error.Render(fmt.Sprintf("Error: failed to scan installed versions: %v", err)))
					os.Exit(1)
				}

				if len(versions) == 0 {
					log.LogError("No Zig versions installed")
					fmt.Println(styles.Error.Render("Error: No Zig versions installed"))
					fmt.Println(styles.Info.Render("\nInstall a Zig version first:"))
					fmt.Println(styles.Info.Render("  zig-installer install 0.13.0"))
					os.Exit(1)
				}

				if len(versions) == 1 {
					log.LogError("Only one version installed")
					fmt.Println(styles.Error.Render("Error: Only one Zig version installed"))
					fmt.Println(styles.Info.Render(fmt.Sprintf("\nCurrently installed: %s", versions[0].Version)))
					fmt.Println(styles.Info.Render("\nInstall another version first:"))
					fmt.Println(styles.Info.Render("  zig-installer install <version>"))
					os.Exit(1)
				}

				// Prompt for version selection
				selected, err := installer.PromptVersionSwitch(versions)
				if err != nil {
					log.LogError("Version selection failed: %v", err)
					fmt.Println(styles.Error.Render(fmt.Sprintf("Error: %v", err)))
					os.Exit(1)
				}
				targetVersion = selected
			}

			// Perform the switch
			if err := installer.SwitchToVersion(cfg, log, formatter, targetVersion); err != nil {
				log.LogError("Switch failed: %v", err)
				fmt.Println(styles.Error.Render(fmt.Sprintf("Error: %v", err)))
				os.Exit(1)
			}

			fmt.Println()
			fmt.Println(styles.Success.Render("✓ Successfully switched to Zig " + targetVersion))
		},
	}

	// Add flags
	switchCmd.Flags().StringVarP(&sc.version, "version", "v", "", "Specific version to switch to (skips interactive prompt)")

	sc.cmd = switchCmd
	return sc
}

// GetCommand returns the cobra command
func (sc *SwitchCommand) GetCommand() *cobra.Command {
	return sc.cmd
}
