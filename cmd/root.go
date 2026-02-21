package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/exilesprx/zig-installer/internal/config"
	"github.com/exilesprx/zig-installer/internal/logger"
	"github.com/exilesprx/zig-installer/internal/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
		Long:  `Installs Zig and the Zig Language Server (ZLS) to your user-local directory (~/.local).`,
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

	// Add switch command
	rc.cmd.AddCommand(NewSwitchCommand(rc.options, rc).GetCommand())

	// Add version command
	rc.cmd.AddCommand(NewVersionCommand().cmd)

	// Add env command
	rc.cmd.AddCommand(NewEnvCommand(rc.options, rc).cmd)

	// Setup custom help template with colors (must be done AFTER adding subcommands)
	rc.setupHelpTemplate()
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

// setupHelpTemplate configures a custom help template with Catppuccin colors for all commands
func (rc *RootCommand) setupHelpTemplate() {
	// Apply to root command
	rc.cmd.SetHelpFunc(rc.customHelpFunc)

	// Apply to all subcommands
	for _, cmd := range rc.cmd.Commands() {
		cmd.SetHelpFunc(rc.customHelpFunc)
	}
}

// customHelpFunc is the custom help function that renders colorized help output
func (rc *RootCommand) customHelpFunc(cmd *cobra.Command, args []string) {
	// Check NoColor flag
	noColor := rc.options.NoColor

	// Load Catppuccin theme styles
	styles := tui.LoadStyles()

	// Build colorized help message
	helpText := buildHelpMessage(cmd, styles, noColor)

	// Print to stdout
	fmt.Fprint(cmd.OutOrStdout(), helpText)
}

// buildHelpMessage constructs the colorized help output
func buildHelpMessage(cmd *cobra.Command, styles *tui.Styles, noColor bool) string {
	var b strings.Builder

	// Title (Command name + short description)
	if cmd.Short != "" {
		b.WriteString(tui.PrintWithStyles(cmd.Short, styles.Header, noColor))
		b.WriteString("\n\n")
	}

	// Long description (if available)
	if cmd.Long != "" {
		b.WriteString(tui.PrintWithStyles(cmd.Long, styles.Detail, noColor))
		b.WriteString("\n\n")
	}

	// Usage section
	b.WriteString(tui.PrintWithStyles("Usage:", styles.Subtitle, noColor))
	b.WriteString("\n")
	b.WriteString(tui.PrintWithStyles(fmt.Sprintf("  %s", cmd.UseLine()), styles.Info, noColor))
	b.WriteString("\n\n")

	// Aliases (if any)
	if len(cmd.Aliases) > 0 {
		b.WriteString(tui.PrintWithStyles("Aliases:", styles.Subtitle, noColor))
		b.WriteString("\n")
		b.WriteString(tui.PrintWithStyles(fmt.Sprintf("  %s", strings.Join(cmd.Aliases, ", ")), styles.Detail, noColor))
		b.WriteString("\n\n")
	}

	// Examples (if any)
	if cmd.Example != "" {
		b.WriteString(tui.PrintWithStyles("Examples:", styles.Subtitle, noColor))
		b.WriteString("\n")
		b.WriteString(tui.PrintWithStyles(cmd.Example, styles.Detail, noColor))
		b.WriteString("\n\n")
	}

	// Available Commands (only show if this command has subcommands)
	if cmd.HasAvailableSubCommands() {
		b.WriteString(tui.PrintWithStyles("Available Commands:", styles.Subtitle, noColor))
		b.WriteString("\n")
		for _, c := range cmd.Commands() {
			if !c.Hidden && c.IsAvailableCommand() {
				// Command name in Info style, description in Detail style
				b.WriteString("  ")
				b.WriteString(tui.PrintWithStyles(fmt.Sprintf("%-15s", c.Name()), styles.Info, noColor))
				b.WriteString(tui.PrintWithStyles(c.Short, styles.Detail, noColor))
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")
	}

	// Flags
	if cmd.HasAvailableFlags() {
		b.WriteString(tui.PrintWithStyles("Flags:", styles.Subtitle, noColor))
		b.WriteString("\n")
		b.WriteString(formatFlags(cmd.Flags(), styles, noColor))
		b.WriteString("\n")
	}

	// Global Flags (if this is a subcommand and has inherited flags)
	if cmd.HasParent() && cmd.HasAvailableInheritedFlags() {
		b.WriteString(tui.PrintWithStyles("Global Flags:", styles.Subtitle, noColor))
		b.WriteString("\n")
		b.WriteString(formatFlags(cmd.InheritedFlags(), styles, noColor))
		b.WriteString("\n")
	}

	// Footer
	b.WriteString(tui.PrintWithStyles(fmt.Sprintf("Use \"%s [command] --help\" for more information about a command.", cmd.Root().Name()), styles.Footer, noColor))
	b.WriteString("\n")

	return b.String()
}

// usageParts holds the split components of a flag's usage string
type usageParts struct {
	description  string
	defaultValue string
}

// splitUsageAndDefault splits usage text and default value for separate coloring
// Handles both "(default value)" and "(default: value)" patterns
func splitUsageAndDefault(usage string) usageParts {
	// Look for "(default" pattern (with or without colon)
	idx := strings.Index(usage, "(default")
	if idx == -1 {
		// Try with capital D (defensive)
		idx = strings.Index(usage, "(Default")
	}

	if idx != -1 {
		return usageParts{
			description:  strings.TrimSpace(usage[:idx]),
			defaultValue: " " + usage[idx:], // Keep leading space for visual separation
		}
	}

	// No default value found
	return usageParts{
		description:  usage,
		defaultValue: "",
	}
}

// formatDefaultValue formats a flag's default value according to pflag conventions
// Returns empty string for implicit defaults (false bools, empty strings)
// Returns formatted string like "(default value)" for explicit defaults
func formatDefaultValue(flag *pflag.Flag) string {
	defValue := flag.DefValue

	// No default value set
	if defValue == "" {
		// Empty string is implicit default for string types
		if flag.Value.Type() == "string" {
			return ""
		}
		// For non-string types, empty DefValue means not set
		return ""
	}

	// Boolean flags: only show "true", hide "false"
	if flag.Value.Type() == "bool" {
		if defValue == "false" {
			return "" // Implicit default, don't show
		}
		return fmt.Sprintf("(default %s)", defValue)
	}

	// String flags: show with quotes
	if flag.Value.Type() == "string" {
		return fmt.Sprintf("(default %q)", defValue)
	}

	// All other types (int, float, duration, etc.): show as-is
	// This includes zero values per user preference
	return fmt.Sprintf("(default %s)", defValue)
}

// calculateFlagNameLength returns the display length of flag name + type
// Used for calculating column alignment
func calculateFlagNameLength(flag *pflag.Flag) int {
	length := 0

	// Account for shorthand: "-v, " (4 chars) or "    " (4 spaces)
	if flag.Shorthand != "" && flag.ShorthandDeprecated == "" {
		length += 4 // "-x, "
	} else {
		length += 4 // "    "
	}

	// Account for long flag name: "--version"
	length += 2 + len(flag.Name) // "--" + name

	// Account for type annotation if not boolean: " string"
	if flag.Value.Type() != "bool" {
		length += 1 + len(flag.Value.Type()) // " " + type
	}

	return length
}

// formatFlagLine formats a single flag line with all components colorized
func formatFlagLine(flag *pflag.Flag, maxNameLen int, styles *tui.Styles, noColor bool) string {
	var b strings.Builder

	// Start with indentation (2 spaces)
	b.WriteString("  ")

	// 1. Format shorthand (if exists)
	if flag.Shorthand != "" && flag.ShorthandDeprecated == "" {
		shorthand := fmt.Sprintf("-%s, ", flag.Shorthand)
		b.WriteString(tui.PrintWithStyles(shorthand, styles.FlagName, noColor))
	} else {
		// No shorthand, add spacing to align with flags that have shorthand
		b.WriteString("    ")
	}

	// 2. Format long flag name
	longFlag := fmt.Sprintf("--%s", flag.Name)
	b.WriteString(tui.PrintWithStyles(longFlag, styles.FlagName, noColor))

	// 3. Format type annotation (if not boolean)
	if flag.Value.Type() != "bool" {
		typeStr := " " + flag.Value.Type()
		b.WriteString(tui.PrintWithStyles(typeStr, styles.FlagType, noColor))
	}

	// 4. Calculate padding for description alignment
	currentLen := calculateFlagNameLength(flag)
	padding := maxNameLen - currentLen + 4 // +4 for spacing between columns
	if padding < 4 {
		padding = 4 // Minimum spacing
	}
	b.WriteString(strings.Repeat(" ", padding))

	// 5. Format description and default value with different colors
	parts := splitUsageAndDefault(flag.Usage)
	b.WriteString(tui.PrintWithStyles(parts.description, styles.Detail, noColor))

	// Check if Usage string already has a default value
	if parts.defaultValue != "" {
		// Use default from Usage string (e.g., custom formats like "default: latest master")
		b.WriteString(tui.PrintWithStyles(parts.defaultValue, styles.FlagDefault, noColor))
	} else {
		// No default in Usage string, check DefValue field
		defaultStr := formatDefaultValue(flag)
		if defaultStr != "" {
			// Add space before default for visual separation
			b.WriteString(tui.PrintWithStyles(" "+defaultStr, styles.FlagDefault, noColor))
		}
	}

	b.WriteString("\n")

	return b.String()
}

// formatFlags formats a flag set with colorized output
// Uses two-pass approach: first calculates alignment, then formats with colors
func formatFlags(flagSet *pflag.FlagSet, styles *tui.Styles, noColor bool) string {
	var b strings.Builder

	// Collect all non-hidden flags for processing
	var flags []*pflag.Flag
	var maxNameLen int

	// First pass: collect flags and calculate max name width for alignment
	flagSet.VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}
		flags = append(flags, flag)

		// Calculate flag name + type length for alignment
		nameLen := calculateFlagNameLength(flag)
		if nameLen > maxNameLen {
			maxNameLen = nameLen
		}
	})

	// Second pass: format each flag with colors
	for _, flag := range flags {
		b.WriteString(formatFlagLine(flag, maxNameLen, styles, noColor))
	}

	return b.String()
}
