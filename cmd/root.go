package cmd

import (
	"fmt"
	"os"

	"github.com/exilesprx/zig-install/internal/config"
	"github.com/exilesprx/zig-install/internal/logger"
	"github.com/exilesprx/zig-install/internal/tui"
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
	GenerateEnv  bool
	ShowSettings bool
	LogFile      string
	EnableLog    bool
}

// RootCommand encapsulates the root command and its dependencies
type RootCommand struct {
	cmd       *cobra.Command
	options   *CommandOptions
	viperInst *viper.Viper
}

// NewRootCommand creates a new instance of the root command
func NewRootCommand() *RootCommand {
	options := &CommandOptions{
		CfgFile:   ".env",
		LogFile:   "zig-install.log",
		EnableLog: true,
	}

	rootCmd := &cobra.Command{
		Use:   "zig-install",
		Short: "Install Zig and ZLS (Zig Language Server)",
		Long: `A tool to install Zig and ZLS (Zig Language Server).
This program must be run as root or with sudo.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Check for environment variables with ZIG_INSTALL prefix and apply them
			processEnvVars(options)
			return nil
		},
	}

	// Main flags - moved to PersistentFlags so they're available to subcommands
	rootCmd.PersistentFlags().BoolVar(&options.ZigOnly, "zig-only", false, "Install only Zig")
	rootCmd.PersistentFlags().BoolVar(&options.ZlsOnly, "zls-only", false, "Install only ZLS (Zig Language Server)")
	rootCmd.PersistentFlags().BoolVar(&options.Verbose, "verbose", false, "Show detailed output during installation")

	// Configuration flags
	rootCmd.PersistentFlags().StringVar(&options.CfgFile, "env", ".env", "Path to environment file")
	rootCmd.PersistentFlags().BoolVar(&options.GenerateEnv, "generate-env", false, "Generate a template .env file")
	rootCmd.PersistentFlags().BoolVar(&options.ShowSettings, "settings", false, "Show current settings")
	rootCmd.PersistentFlags().BoolVar(&options.NoColor, "no-color", false, "Disable colored output")

	// Logging flags
	rootCmd.PersistentFlags().StringVar(&options.LogFile, "log-file", "zig-install.log", "File to log errors to")
	rootCmd.PersistentFlags().BoolVar(&options.EnableLog, "enable-log", true, "Enable logging to file")

	return &RootCommand{
		cmd:       rootCmd,
		options:   options,
		viperInst: viper.New(),
	}
}

// processEnvVars checks for environment variables with ZIG_INSTALL prefix and applies them to options
func processEnvVars(options *CommandOptions) {
	// Check for ZIG_INSTALL environment variables and apply them to options
	// Note: .env file's specific values (ZIG_DIR, etc.) are handled by Viper, not here
	if val := os.Getenv("ZIG_INSTALL_LOG_FILE"); val != "" {
		options.LogFile = val
	}
	if val := os.Getenv("ZIG_INSTALL_ENV"); val != "" {
		options.CfgFile = val
	}
	// TODO: clean up the environment variables, vs .env, vs build values. right now they're a mess
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

	// Add version command
	rc.cmd.AddCommand(NewVersionCommand().cmd)

	// Add other commands as needed
}

// LoadLoggerAndConfig prepares the logger and config for commands
func (rc *RootCommand) LoadLoggerAndConfig() (*config.Config, logger.ILogger, *tui.Styles, error) {
	// Initialize a fresh Viper instance that will ONLY handle .env file settings
	v := config.InitViper()

	// Load only .env configurable settings using Viper
	cfg, err := config.LoadEnvConfig(v, rc.options.CfgFile)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load .env configuration: %w", err)
	}

	// Set all Cobra-managed config values from command-line flags
	cfg.EnvFile = rc.options.CfgFile
	cfg.ZigOnly = rc.options.ZigOnly
	cfg.ZLSOnly = rc.options.ZlsOnly
	cfg.Verbose = rc.options.Verbose
	cfg.NoColor = rc.options.NoColor
	cfg.GenerateEnv = rc.options.GenerateEnv
	cfg.ShowSettings = rc.options.ShowSettings
	cfg.LogFile = rc.options.LogFile
	cfg.EnableLog = rc.options.EnableLog

	// Initialize logger
	log, err := logger.NewFileLogger(cfg.LogFile, cfg.EnableLog)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize styles
	colors := tui.NewMochaColors()
	styles := tui.NewStyles(colors)

	return cfg, log, styles, nil
}
