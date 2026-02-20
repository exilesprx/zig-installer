package cmd

import (
	"fmt"
	"os"

	"github.com/exilesprx/zig-install/internal/config"
	"github.com/exilesprx/zig-install/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CommandOptions holds configuration options shared by commands
type CommandOptions struct {
	CfgFile      string
	ZigOnly      bool
	ZlsOnly      bool
	Verbose      bool
	NoColor      bool
	ShowSettings bool
	LogFile      string
	EnableLog    bool
	AutoCleanup  bool
	NoCleanup    bool
	KeepLast     int
}

// RootCommand encapsulates the root command and its dependencies
type RootCommand struct {
	cmd       *cobra.Command
	options   *CommandOptions
	viperInst *viper.Viper
}

// NewRootCommand creates a new instance of the root command
func NewRootCommand() *RootCommand {
	options := &CommandOptions{}

	rootCmd := &cobra.Command{
		Use:   "zig-installer",
		Short: "Install Zig and ZLS (Zig Language Server)",
		Long: `Installs Zig and the Zig Language Server (ZLS) to your user-local directory.

⚠️  BREAKING CHANGE (v4.0.0): This installer now uses user-local installation ONLY.
   System-wide installation is no longer supported.

Installation directories:
  • Zig versions: ~/.local/share/zig/
  • ZLS repository: ~/.local/share/zls/
  • Binary symlinks: ~/.local/bin/zig and ~/.local/bin/zls

❌ Do NOT run with sudo. This installer uses user-local installation only.

Migration: If you have an existing system installation, run:
  zig-installer migrate

macOS: Support is experimental and may have issues.`,
	}

	// Main flags - moved to PersistentFlags so they're available to subcommands
	rootCmd.PersistentFlags().BoolVar(&options.ZigOnly, "zig-only", false, "Install only Zig")
	rootCmd.PersistentFlags().BoolVar(&options.ZlsOnly, "zls-only", false, "Install only ZLS (Zig Language Server)")
	rootCmd.PersistentFlags().BoolVar(&options.Verbose, "verbose", false, "Show detailed output during installation")

	// Configuration flags
	rootCmd.PersistentFlags().StringVar(&options.CfgFile, "env", ".env", "Path to environment file")
	rootCmd.PersistentFlags().BoolVar(&options.ShowSettings, "settings", false, "Show current settings")
	rootCmd.PersistentFlags().BoolVar(&options.NoColor, "no-color", false, "Disable colored output")

	// Logging flags
	rootCmd.PersistentFlags().StringVar(&options.LogFile, "log-file", "zig-install.log", "File to log errors to")
	rootCmd.PersistentFlags().BoolVar(&options.EnableLog, "enable-log", true, "Enable logging to file")

	// Cleanup flags
	rootCmd.PersistentFlags().BoolVar(&options.AutoCleanup, "auto-cleanup", false, "Automatically cleanup old versions after install without prompting")
	rootCmd.PersistentFlags().BoolVar(&options.NoCleanup, "no-cleanup", false, "Disable auto-cleanup prompt after install")
	rootCmd.PersistentFlags().IntVar(&options.KeepLast, "keep-last", 0, "Keep last N versions when cleaning up")

	return &RootCommand{
		cmd:       rootCmd,
		options:   options,
		viperInst: viper.New(),
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// It is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCommand := NewRootCommand()

	// Add subcommands
	rootCommand.AddCommands()

	if err := rootCommand.cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// AddCommands adds all child commands to the root command
func (rc *RootCommand) AddCommands() {
	// Add install command and pass this root command instance
	rc.cmd.AddCommand(NewInstallCommand(rc.options, rc).cmd)

	// Add cleanup command
	rc.cmd.AddCommand(NewCleanupCommand(rc.options, rc).cmd)

	// Add migrate command
	rc.cmd.AddCommand(NewMigrateCommand(rc.options, rc).Command())

	// Add version command
	rc.cmd.AddCommand(NewVersionCommand().cmd)

	// Add env command
	rc.cmd.AddCommand(NewEnvCommand(rc.options, rc).cmd)
}

// LoadLoggerAndConfig prepares the logger and config for commands
func (rc *RootCommand) LoadLoggerAndConfig() (*config.Config, logger.ILogger, error) {
	// Initialize a fresh Viper instance that will ONLY handle .env file settings
	v := config.InitViper()

	// Load only .env configurable settings using Viper
	cfg, err := config.LoadEnvConfig(v, rc.options.CfgFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load .env configuration: %w", err)
	}

	// Set all Cobra-managed config values from command-line flags
	cfg.EnvFile = rc.options.CfgFile
	cfg.ZigOnly = rc.options.ZigOnly
	cfg.ZLSOnly = rc.options.ZlsOnly
	cfg.Verbose = rc.options.Verbose
	cfg.NoColor = rc.options.NoColor
	cfg.ShowSettings = rc.options.ShowSettings
	cfg.LogFile = rc.options.LogFile
	cfg.EnableLog = rc.options.EnableLog
	cfg.AutoCleanup = rc.options.AutoCleanup
	cfg.NoCleanup = rc.options.NoCleanup
	cfg.KeepLast = rc.options.KeepLast

	// Initialize logger
	log, err := logger.NewFileLogger(cfg.LogFile, cfg.EnableLog)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	return cfg, log, nil
}
