package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/exilesprx/zig-install/internal/config"
	"github.com/exilesprx/zig-install/internal/installer"
	"github.com/exilesprx/zig-install/internal/logger"
	"github.com/exilesprx/zig-install/internal/tui"
	"github.com/spf13/cobra"
)

// InstallCommand encapsulates the install command
type InstallCommand struct {
	cmd     *cobra.Command
	options *CommandOptions
	rootCmd *RootCommand // Add reference to root command
}

// NewInstallCommand creates a new install command instance
func NewInstallCommand(options *CommandOptions, rootCmd *RootCommand) *InstallCommand {
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install Zig and ZLS",
		Long: `Install the Zig compiler and ZLS language server.
By default, both Zig and ZLS will be installed unless --zig-only or --zls-only is specified.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Use the provided root command instead of creating a new one
			cfg, log, styles, err := rootCmd.LoadLoggerAndConfig()
			if err != nil {
				fmt.Printf("Error initializing: %v\n", err)
				os.Exit(1)
			}
			defer log.Close()

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
				if cfg.NoColor {
					cfg.PrintSettings(cfg.NoColor)
				} else {
					cfg.PrintSettings(cfg.NoColor)
				}
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
			runInstallation(cfg, styles, log)
		},
	}

	return &InstallCommand{
		cmd:     installCmd,
		options: options,
		rootCmd: rootCmd, // Set the root command reference
	}
}

// runInstallation starts the TUI installation process
func runInstallation(config *config.Config, styles *tui.Styles, logger logger.ILogger) {
	initialModel := tui.NewModel(config, styles, logger)
	p := tea.NewProgram(initialModel)

	go func() {
		logger.LogInfo("Starting installation process")

		// Store Zig version to match with ZLS
		var zigVersion string

		if !config.ZLSOnly {
			logger.LogInfo("Starting Zig installation")
			p.Send(tui.StatusMsg("Installing Zig..."))
			var err error
			zigVersion, err = installer.InstallZig(p, config, logger)
			if err != nil {
				logger.LogError("Zig installation failed: %v", err)
				p.Send(tui.ErrorMsg(err))
				return
			}
			logger.LogInfo("Zig installation completed successfully")
			p.Send(tui.ZigDoneMsg{})
		} else {
			// If only installing ZLS, get the current Zig version
			zigCmd := exec.Command("zig", "version")
			output, err := zigCmd.Output()
			if err != nil {
				logger.LogError("Failed to get Zig version: %v", err)
				p.Send(tui.ErrorMsg(fmt.Errorf("failed to get Zig version: %w", err)))
				return
			}
			zigVersion = strings.TrimSpace(string(output))
		}

		if !config.ZigOnly {
			logger.LogInfo("Starting ZLS installation")
			p.Send(tui.StatusMsg("Installing ZLS..."))
			if err := installer.InstallZLS(p, config, logger, zigVersion); err != nil {
				logger.LogError("ZLS installation failed: %v", err)
				p.Send(tui.ErrorMsg(err))
				return
			}
			logger.LogInfo("ZLS installation completed successfully")
			p.Send(tui.ZLSDoneMsg{})
		}

		logger.LogInfo("Installation process completed successfully")
		p.Send(tui.InstallCompleteMsg("Installation completed successfully! 🎉"))
	}()

	if _, err := p.Run(); err != nil {
		logger.LogError("Error running program: %v", err)
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
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
