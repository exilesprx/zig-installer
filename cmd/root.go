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
			// Initialize Viper with the command's viper instance
			v := viper.GetViper()

			// Bind config flags to viper and check for errors
			if err := v.BindPFlag("zig_only", cmd.PersistentFlags().Lookup("zig-only")); err != nil {
				return fmt.Errorf("error binding zig-only flag: %w", err)
			}
			if err := v.BindPFlag("zls_only", cmd.PersistentFlags().Lookup("zls-only")); err != nil {
				return fmt.Errorf("error binding zls-only flag: %w", err)
			}
			if err := v.BindPFlag("verbose", cmd.PersistentFlags().Lookup("verbose")); err != nil {
				return fmt.Errorf("error binding verbose flag: %w", err)
			}
			if err := v.BindPFlag("no_color", cmd.PersistentFlags().Lookup("no-color")); err != nil {
				return fmt.Errorf("error binding no-color flag: %w", err)
			}
			if err := v.BindPFlag("generate_env", cmd.PersistentFlags().Lookup("generate-env")); err != nil {
				return fmt.Errorf("error binding generate-env flag: %w", err)
			}
			if err := v.BindPFlag("show_settings", cmd.PersistentFlags().Lookup("settings")); err != nil {
				return fmt.Errorf("error binding settings flag: %w", err)
			}
			if err := v.BindPFlag("log_file", cmd.PersistentFlags().Lookup("log-file")); err != nil {
				return fmt.Errorf("error binding log-file flag: %w", err)
			}
			if err := v.BindPFlag("enable_log", cmd.PersistentFlags().Lookup("enable-log")); err != nil {
				return fmt.Errorf("error binding enable-log flag: %w", err)
			}

			// Set config file if provided
			if options.CfgFile != "" {
				v.SetConfigFile(options.CfgFile)
			} else {
				// Look for .env in the current directory
				v.SetConfigFile(".env")
			}

			// Read in environment variables with prefix
			v.SetEnvPrefix("ZIG_INSTALL")
			v.AutomaticEnv()

			// Load the config file if it exists
			if err := v.ReadInConfig(); err == nil {
				if options.Verbose {
					fmt.Println("Using config file:", v.ConfigFileUsed())
				}
			} else if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				// Config file was found but another error was produced
				return fmt.Errorf("error reading config file: %w", err)
			}

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
		viperInst: viper.GetViper(),
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
	// Add install command
	rc.cmd.AddCommand(NewInstallCommand(rc.options).cmd)

	// Add version command
	rc.cmd.AddCommand(NewVersionCommand().cmd)

	// Add other commands as needed
}

// LoadLoggerAndConfig prepares the logger and config for commands
func (rc *RootCommand) LoadLoggerAndConfig() (*config.Config, logger.ILogger, *tui.Styles, error) {
	// Initialize Viper and load configuration
	v := config.InitViper()
	cfg, err := config.LoadConfig(v, rc.options.CfgFile)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override config with CLI flags
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
