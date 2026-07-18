// Package cmd defines the root command and shared functionality for the zig-installer CLI application. It sets up the main command structure, global flags, configuration loading, logging, and a custom help template with Catppuccin colors. The RootCommand struct encapsulates the Cobra command and its dependencies, providing a clean interface for subcommands to access shared options and utilities.
package cmd

import (
	"fmt"
	"os"

	"github.com/exilesprx/zig-installer/internal/config"
	"github.com/exilesprx/zig-installer/internal/logger"
	"github.com/exilesprx/zig-installer/internal/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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

type RootCommand struct {
	cmd       *cobra.Command
	options   *CommandOptions
	viperInst *viper.Viper
}

func NewRootCommand() *RootCommand {
	options := &CommandOptions{}

	rootCmd := &cobra.Command{
		Use:   "zig-installer",
		Short: "Install Zig and ZLS (Zig Language Server)",
		Long:  `Installs Zig and the Zig Language Server (ZLS) to your user-local directory (~/.local).`,
	}

	rootCmd.PersistentFlags().BoolVar(&options.ZigOnly, "zig-only", false, "Install only Zig")
	rootCmd.PersistentFlags().BoolVar(&options.ZlsOnly, "zls-only", false, "Install only ZLS (Zig Language Server)")
	rootCmd.PersistentFlags().BoolVar(&options.Verbose, "verbose", false, "Show detailed output during installation")

	rootCmd.PersistentFlags().StringVar(&options.CfgFile, "env", ".env", "Path to environment file")
	rootCmd.PersistentFlags().BoolVar(&options.ShowSettings, "settings", false, "Show current settings")
	rootCmd.PersistentFlags().BoolVar(&options.NoColor, "no-color", false, "Disable colored output")

	rootCmd.PersistentFlags().StringVar(&options.LogFile, "log-file", "zig-install.log", "File to log errors to")
	rootCmd.PersistentFlags().BoolVar(&options.EnableLog, "enable-log", true, "Enable logging to file")

	rootCmd.PersistentFlags().BoolVar(&options.AutoCleanup, "auto-cleanup", false, "Automatically cleanup old versions after install without prompting")
	rootCmd.PersistentFlags().BoolVar(&options.NoCleanup, "no-cleanup", false, "Disable auto-cleanup prompt after install")
	rootCmd.PersistentFlags().IntVar(&options.KeepLast, "keep-last", 0, "Keep last N versions when cleaning up")

	return &RootCommand{
		cmd:       rootCmd,
		options:   options,
		viperInst: viper.New(),
	}
}

func Execute() {
	rootCommand := NewRootCommand()

	rootCommand.AddCommands()

	if err := rootCommand.cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (rc *RootCommand) AddCommands() {
	rc.cmd.AddCommand(NewInstallCommand(rc.options, rc).Command())
	rc.cmd.AddCommand(NewCleanupCommand(rc.options, rc).Command())
	rc.cmd.AddCommand(NewMigrateCommand(rc.options, rc).Command())
	rc.cmd.AddCommand(NewSwitchCommand(rc.options, rc).Command())
	rc.cmd.AddCommand(NewVersionCommand().Command())
	rc.cmd.AddCommand(NewEnvCommand(rc.options, rc).Command())

	rc.setupHelpTemplate()
}

func (rc *RootCommand) LoadConfig() (*config.Config, error) {
	v := config.InitViper()
	cfg, err := config.LoadEnvConfig(v, rc.options.CfgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load .env configuration: %w", err)
	}

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

	return cfg, nil
}

func (rc *RootCommand) CreateLogger(cfg *config.Config) (logger.Logger, error) {
	fileLog, err := logger.NewFileLogger(cfg.LogFile, cfg.EnableLog)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	var log logger.Logger = fileLog

	if cfg.Verbose {
		styles := tui.LoadStyles()
		consoleLog := logger.NewConsoleLogger(styles, cfg.NoColor)
		log = logger.NewMultiLogger(fileLog, consoleLog)
	}

	return log, nil
}

func (rc *RootCommand) setupHelpTemplate() {
	rc.cmd.SetHelpFunc(rc.customHelpFunc)
	for _, cmd := range rc.cmd.Commands() {
		cmd.SetHelpFunc(rc.customHelpFunc)
	}
}

func (rc *RootCommand) customHelpFunc(cmd *cobra.Command, args []string) {
	noColor := rc.options.NoColor
	styles := tui.LoadStyles()
	helpText := buildHelpMessage(cmd, styles, noColor)
	_, _ = fmt.Fprint(cmd.OutOrStdout(), helpText)
}
