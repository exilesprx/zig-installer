package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/exilesprx/zig-installer/internal/config"
	"github.com/exilesprx/zig-installer/internal/installer"
	"github.com/exilesprx/zig-installer/internal/logger"
	"github.com/exilesprx/zig-installer/internal/tui"
	"github.com/spf13/cobra"
)

type InstallCommand struct {
	cmd        *cobra.Command
	options    *CommandOptions
	rootCmd    *RootCommand
	zigVersion string
}

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
			cfg, err := rootCmd.LoadConfig()
			styles := tui.LoadStyles()
			if err != nil {
				fmt.Printf("Error initializing: %v\n", err)
				os.Exit(1)
			}
			log, err := rootCmd.CreateLogger(cfg)
			if err != nil {
				fmt.Printf("Error initializing logger: %v\n", err)
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
				fmt.Println(tui.PrintWithStyles("Created .env file with default settings", styles.Success, cfg.NoColor))
				os.Exit(0)
			}

			if cfg.ShowSettings {
				cfg.PrintSettings()
				log.LogInfo("Displayed settings")
				os.Exit(0)
			}

			log.LogInfo("Starting installation process")

			if err := checkNotRoot(); err != nil {
				log.LogError("Root check failed: %v", err)
				fmt.Println(tui.PrintWithStyles(fmt.Sprintf("Error: %v", err), styles.Error, cfg.NoColor))
				os.Exit(1)
			}

			if runtime.GOOS == "darwin" {
				formatter := installer.NewTaskFormatter(cfg, styles)
				formatter.PrintWarning("macOS Support", "macOS support is experimental and may have issues")
				fmt.Println() // Blank line
			}

			if err := checkDependencies(); err != nil {
				log.LogError("Dependency check failed: %v", err)
				fmt.Println(tui.PrintWithStyles(fmt.Sprintf("Error: %v", err), styles.Error, cfg.NoColor))
				os.Exit(1)
			}

			formatter := installer.NewTaskFormatter(cfg, styles)

			migrationChoice, systemDir, err := installer.DetectAndPromptMigration(formatter, log)
			if err != nil {
				log.LogError("Migration prompt failed: %v", err)
				fmt.Println(tui.PrintWithStyles(fmt.Sprintf("Error: %v", err), styles.Error, cfg.NoColor))
				os.Exit(1)
			}

			switch migrationChoice {
			case installer.MigrationChoiceMigrate:
				formatter.PrintSection("Migration")
				if err := installer.PerformMigration(systemDir, formatter, log, false); err != nil {
					fmt.Println()
					formatter.PrintError("Migration Failed", fmt.Sprintf("%v", err))
					formatter.PrintTask("Next Steps", "Manual removal required",
						"Follow the instructions above, then run install again")
					os.Exit(1)
				}
				formatter.PrintSuccess("Migration Complete", "System installation migrated successfully")
				fmt.Println() // Blank line for readability

			case installer.MigrationChoiceKeepBoth:
				installer.WarnAboutPathConflict(systemDir, cfg.BinDir, formatter)
				fmt.Println() // Blank line

			case installer.MigrationChoiceCancel:
				formatter.PrintTask("Installation", "Cancelled", "User cancelled installation")
				os.Exit(0)

			default:
				// No system installation detected, proceed normally
			}

			if ic.zigVersion != "" && ic.zigVersion != "master" {
				if !isValidVersion(ic.zigVersion) {
					log.LogError("Invalid version format: %s", ic.zigVersion)
					fmt.Println(tui.PrintWithStyles(
						fmt.Sprintf("Error: invalid version format %q. Expected semver (e.g., 0.13.0) or \"master\"", ic.zigVersion),
						styles.Error, cfg.NoColor,
					))
					os.Exit(1)
				}
			}

			runInstallation(cfg, styles, log, ic.zigVersion)
		},
	}

	// Add version flag
	installCmd.Flags().StringVarP(&ic.zigVersion, "version", "v", "", "Specify Zig version to install (default: latest master)")

	ic.cmd = installCmd
	return ic
}

func (ic *InstallCommand) Command() *cobra.Command {
	return ic.cmd
}

func runInstallation(config *config.Config, styles *tui.Styles, logger logger.Logger, zigVersion string) {
	formatter := installer.NewTaskFormatter(config, styles)

	formatter.PrintSection("System Check")
	formatter.PrintSuccess("Dependencies verified, ready to install", "")

	if !config.ZLSOnly {
		logger.LogInfo("Starting Zig installation")
		formatter.PrintSection("Zig Installation")

		var err error
		zigVersion, err = installer.InstallZig(config, logger, formatter, zigVersion)
		if err != nil {
			logger.LogError("Zig installation failed: %v", err)
			fmt.Println(styles.Error.Render(fmt.Sprintf("Error: %v", err)))
			return
		}
		logger.LogInfo("Zig installation completed successfully")
		formatter.PrintSuccess("Zig compiler installed and configured", "")

		fmt.Println() // Blank line for readability
		checkPathConfiguration(config.BinDir, formatter)
		fmt.Println() // Blank line

		if !config.NoCleanup {
			logger.LogInfo("Checking for auto-cleanup opportunity")
			if err := installer.AutoCleanupPrompt(config, logger, formatter, zigVersion); err != nil {
				logger.LogError("Auto-cleanup failed: %v", err)
				formatter.PrintError("Auto-cleanup", fmt.Sprintf("Cleanup failed: %v", err))
			}
		}
	} else {
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
		formatter.PrintSection("ZLS Installation")

		if err := installer.InstallZLS(config, logger, formatter, zigVersion); err != nil {
			logger.LogError("ZLS installation failed: %v", err)
			fmt.Println(styles.Error.Render(fmt.Sprintf("Error: %v", err)))
			return
		}
		logger.LogInfo("ZLS installation completed successfully")
		formatter.PrintSuccess("ZLS language server installed and configured", "")
	}

	logger.LogInfo("Installation process completed successfully")
	fmt.Println()
	fmt.Println(styles.Success.Render("Installation completed successfully!"))
	fmt.Println(styles.Separator.Render(strings.Repeat("─", 40)))
}

func checkNotRoot() error {
	if os.Geteuid() == 0 {
		return fmt.Errorf(`❌ ERROR: This installer should NOT be run with sudo.

zig-installer v4.0.0 uses user-local installation only.

Installation directory: ~/.local/share/zig
Binary symlinks: ~/.local/bin/zig and ~/.local/bin/zls

If you have an existing system-wide installation (/opt/zig or /usr/local/zig):
  1. Run without sudo: ./zig-installer migrate
  2. Then install: ./zig-installer install

To manually remove system installation:
  sudo rm -rf /opt/zig /opt/zls /usr/local/zig /usr/local/zls
  sudo rm -f /usr/local/bin/zig /usr/local/bin/zls

Then run: ./zig-installer install (without sudo)`)
	}
	return nil
}

func checkPathConfiguration(binDir string, formatter installer.OutputFormatter) {
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return
	}

	pathDirs := strings.Split(pathEnv, string(os.PathListSeparator))

	for _, dir := range pathDirs {
		if dir == binDir {
			return
		}
		if absDir, err := filepath.EvalSymlinks(dir); err == nil {
			if absBin, err := filepath.EvalSymlinks(binDir); err == nil {
				if absDir == absBin {
					return
				}
			}
		}
	}

	formatter.PrintWarning("PATH Configuration Required",
		fmt.Sprintf("%s is not in your PATH", binDir))

	shell := os.Getenv("SHELL")
	var configFile, exportCmd string

	if strings.Contains(shell, "fish") {
		configFile = "~/.config/fish/config.fish"
		exportCmd = fmt.Sprintf("set -gx PATH %s $PATH", binDir)
	} else if strings.Contains(shell, "zsh") {
		configFile = "~/.zshrc"
		exportCmd = fmt.Sprintf("export PATH=\"%s:$PATH\"", binDir)
	} else {
		configFile = "~/.bashrc"
		exportCmd = fmt.Sprintf("export PATH=\"%s:$PATH\"", binDir)
	}

	formatter.PrintTask("Step 1", "Add to PATH",
		fmt.Sprintf("Add this line to your %s:\n      %s", configFile, exportCmd))
	formatter.PrintTask("Step 2", "Reload configuration",
		fmt.Sprintf("Run: source %s", configFile))
	formatter.PrintTask("Alternative", "Start new terminal",
		"Or simply open a new terminal session")
}

func checkDependencies() error {
	var missingDeps []string
	requiredDeps := []string{"git", "wget", "minisign", "xz"}

	for _, dep := range requiredDeps {
		if _, err := exec.LookPath(dep); err != nil {
			missingDeps = append(missingDeps, dep)
		}
	}

	if len(missingDeps) > 0 {
		if len(missingDeps) == 1 {
			return fmt.Errorf("missing dependency: %s. Please install it", missingDeps[0])
		}
		return fmt.Errorf("missing dependencies: %s. Please install them", strings.Join(missingDeps, ", "))
	}
	return nil
}

func isValidVersion(version string) bool {
	if version == "master" {
		return true
	}
	// Match patterns like 0.13.0, 0.12.0-dev.123, 0.13.0-rc.1
	for _, c := range version {
		if c >= '0' && c <= '9' || c == '.' || c == '-' {
			continue
		}
		return false
	}
	// Must contain at least one dot (e.g., 0.13)
	return strings.Contains(version, ".")
}
