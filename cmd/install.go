package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/exilesprx/zig-install/internal/config"
	"github.com/exilesprx/zig-install/internal/installer"
	"github.com/exilesprx/zig-install/internal/logger"
	"github.com/exilesprx/zig-install/internal/tui"
	"github.com/spf13/cobra"
)

// InstallCommand encapsulates the install command
type InstallCommand struct {
	cmd        *cobra.Command
	options    *CommandOptions
	rootCmd    *RootCommand
	zigVersion string
}

// NewInstallCommand creates a new install command instance
func NewInstallCommand(options *CommandOptions, rootCmd *RootCommand) *InstallCommand {
	ic := &InstallCommand{
		options: options,
		rootCmd: rootCmd,
	}

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install Zig and ZLS",
		Long: `Install the Zig compiler and ZLS language server.
By default, both Zig and ZLS will be installed unless --zig-only or --zls-only is specified.
You can specify a version to install using --version, otherwise the latest master version will be used.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Use the provided root command instead of creating a new one
			cfg, log, err := rootCmd.LoadLoggerAndConfig()
			styles := tui.LoadStyles()
			if err != nil {
				fmt.Printf("Error initializing: %v\n", err)
				os.Exit(1)
			}
			defer func() { _ = log.Close() }()

			if cfg.GenerateEnv {
				if err := cfg.GenerateEnvFile(); err != nil {
					log.LogError("Failed to generate .env file: %v", err)
					fmt.Println(tui.PrintWithStyles(fmt.Sprintf("Error: %v", err), styles.Error, cfg.NoColor))
					os.Exit(1)
				}
				log.LogInfo("Created .env file with default settings")
				fmt.Println(tui.PrintWithStyles("✓ Created .env file with default settings", styles.Success, cfg.NoColor))
				os.Exit(0)
			}

			if cfg.ShowSettings {
				cfg.PrintSettings()
				log.LogInfo("Displayed settings")
				os.Exit(0)
			}

			log.LogInfo("Starting installation process")

			// First check for root privileges
			if err := checkIsRoot(); err != nil {
				log.LogError("Root check failed: %v", err)
				fmt.Println(tui.PrintWithStyles(fmt.Sprintf("Error: %v", err), styles.Error, cfg.NoColor))
				os.Exit(1)
			}

			// Then check for dependencies
			if err := checkDependencies(); err != nil {
				log.LogError("Dependency check failed: %v", err)
				fmt.Println(tui.PrintWithStyles(fmt.Sprintf("Error: %v", err), styles.Error, cfg.NoColor))
				os.Exit(1)
			}

			// Run the TUI installer
			runInstallation(cfg, styles, log, ic.zigVersion)
		},
	}

	// Add version flag
	installCmd.Flags().StringVarP(&ic.zigVersion, "version", "v", "", "Specify Zig version to install (default: latest master)")

	ic.cmd = installCmd
	return ic
}

// runInstallation starts the installation process with simple, clean output
func runInstallation(config *config.Config, styles *tui.Styles, logger logger.ILogger, zigVersion string) {
	// Set global config for installers to use
	installer.SetGlobalConfig(config, styles)

	// System check
	installer.PrintTask("System check", "✓ Success", "Dependencies verified, ready to install")

	// Store Zig version to match with ZLS
	if !config.ZLSOnly {
		logger.LogInfo("Starting Zig installation")
		installer.PrintTask("Zig installation start", "→ Starting", "Beginning Zig installation process")

		var err error
		zigVersion, err = installer.InstallZig(nil, config, logger, zigVersion)
		if err != nil {
			logger.LogError("Zig installation failed: %v", err)
			fmt.Println(styles.Error.Render(fmt.Sprintf("Error: %v", err)))
			return
		}
		logger.LogInfo("Zig installation completed successfully")
		installer.PrintTask("Zig installation", "✓ Success", "Zig compiler installed and configured")
	} else {
		// If only installing ZLS, get the current Zig version
		zigCmd := exec.Command("zig", "version")
		output, err := zigCmd.Output()
		if err != nil {
			logger.LogError("Failed to get Zig version: %v", err)
			fmt.Println(styles.Error.Render(fmt.Sprintf("Error: failed to get Zig version: %v", err)))
			return
		}
		zigVersion = strings.TrimSpace(string(output))
	}

	if !config.ZigOnly {
		logger.LogInfo("Starting ZLS installation")
		installer.PrintTask("ZLS installation start", "→ Starting", "Beginning ZLS installation process")

		if err := installer.InstallZLS(nil, config, logger, zigVersion); err != nil {
			logger.LogError("ZLS installation failed: %v", err)
			fmt.Println(styles.Error.Render(fmt.Sprintf("Error: %v", err)))
			return
		}
		logger.LogInfo("ZLS installation completed successfully")
		installer.PrintTask("ZLS installation", "✓ Success", "ZLS language server installed and configured")
	}

	logger.LogInfo("Installation process completed successfully")
	fmt.Println()
	fmt.Println(styles.Success.Render("Installation completed successfully! 🎉"))
	fmt.Println(styles.Separator.Render(strings.Repeat("─", 40)))
}

// checkIsRoot verifies the script is running as root
func checkIsRoot() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("this program must be run as root. Please use 'sudo' or log in as root")
	}
	return nil
}

// checkDependencies verifies all required tools are installed
func checkDependencies() error {
	var missingDeps []string
	requiredDeps := []string{"git", "wget", "jq", "minisign", "xz"}

	for _, dep := range requiredDeps {
		if _, err := exec.LookPath(dep); err != nil {
			missingDeps = append(missingDeps, dep)
		}
	}

	if len(missingDeps) > 0 {
		if len(missingDeps) == 1 {
			return fmt.Errorf("missing dependency: %s. Please install it", missingDeps[0])
		} else {
			return fmt.Errorf("missing dependencies: %s. Please install them", strings.Join(missingDeps, ", "))
		}
	}
	return nil
}
