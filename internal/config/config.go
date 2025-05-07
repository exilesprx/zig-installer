package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Build-time configurable defaults that can be set with linker flags
var (
	// DefaultZigPubKey must be set with -ldflags="-X 'github.com/exilesprx/zig-install/internal/config.DefaultZigPubKey=key'"
	DefaultZigPubKey string

	// DefaultZigDir must be set with -ldflags="-X 'github.com/exilesprx/zig-install/internal/config.DefaultZigDir=/custom/path'"
	DefaultZigDir string

	// DefaultZLSDir must be set with -ldflags="-X 'github.com/exilesprx/zig-install/internal/config.DefaultZLSDir=/custom/path'"
	DefaultZLSDir string

	// DefaultBinDir must be set with -ldflags="-X 'github.com/exilesprx/zig-install/internal/config.DefaultBinDir=/custom/path'"
	DefaultBinDir string

	// DefaultZigDownURL must be set with -ldflags="-X 'github.com/exilesprx/zig-install/internal/config.DefaultZigDownURL=url'"
	DefaultZigDownURL string

	// DefaultZigIndexURL must be set with -ldflags="-X 'github.com/exilesprx/zig-install/internal/config.DefaultZigIndexURL=url'"
	DefaultZigIndexURL string

	// DefaultEnvFile the default name of the environment file
	DefaultEnvFile = ".env"

	// DefaultVerbose the default verbosity level
	DefaultVerbose = false

	// DefaultLogFile the default log file name
	DefaultLogFile = "zig-install.log"

	// DefaultEnableLog the default log enable flag
	DefaultEnableLog = false
)

// Config contains the application configuration
type Config struct {
	// .env configurable values (via Viper)
	ZigDir      string
	ZLSDir      string
	BinDir      string
	ZigPubKey   string
	ZigDownURL  string
	ZigIndexURL string

	// CLI options and flags (via Cobra)
	EnvFile      string
	ZigOnly      bool
	ZLSOnly      bool
	NoColor      bool
	GenerateEnv  bool
	ShowSettings bool
	Verbose      bool
	LogFile      string
	EnableLog    bool
}

// NewConfig creates a new configuration with default values
func NewConfig() *Config {
	return &Config{
		// Default values for .env configurable settings
		ZigDir:      DefaultZigDir,
		ZLSDir:      DefaultZLSDir,
		BinDir:      DefaultBinDir,
		ZigPubKey:   DefaultZigPubKey,
		ZigDownURL:  DefaultZigDownURL,
		ZigIndexURL: DefaultZigIndexURL,

		// Default values for CLI options
		EnvFile:   DefaultEnvFile,
		Verbose:   DefaultVerbose,
		LogFile:   DefaultLogFile,
		EnableLog: DefaultEnableLog,
	}
}

// InitViper initializes Viper with only the .env configurable values
func InitViper() *viper.Viper {
	v := viper.New()

	// Set default values for .env configurable settings only
	v.SetDefault("zig_dir", DefaultZigDir)
	v.SetDefault("zls_dir", DefaultZLSDir)
	v.SetDefault("bin_dir", DefaultBinDir)
	v.SetDefault("zig_pub_key", DefaultZigPubKey)
	v.SetDefault("zig_down_url", DefaultZigDownURL)
	v.SetDefault("zig_index_url", DefaultZigIndexURL)

	return v
}

// LoadEnvConfig loads only the .env configurable settings using Viper
func LoadEnvConfig(v *viper.Viper, envFile string) (*Config, error) {
	// Start with default configuration
	config := NewConfig()

	if envFile != "" {
		// Check if the env file exists
		if _, err := os.Stat(envFile); err == nil {
			v.SetConfigFile(envFile)
			if err := v.ReadInConfig(); err != nil {
				return nil, fmt.Errorf("error reading config file: %w", err)
			}

			// Only override the six specific values if they are set in the .env file
			if v.IsSet("zig_dir") {
				config.ZigDir = v.GetString("zig_dir")
			}
			if v.IsSet("zls_dir") {
				config.ZLSDir = v.GetString("zls_dir")
			}
			if v.IsSet("bin_dir") {
				config.BinDir = v.GetString("bin_dir")
			}
			if v.IsSet("zig_pub_key") {
				config.ZigPubKey = v.GetString("zig_pub_key")
			}
			if v.IsSet("zig_down_url") {
				config.ZigDownURL = v.GetString("zig_down_url")
			}
			if v.IsSet("zig_index_url") {
				config.ZigIndexURL = v.GetString("zig_index_url")
			}
		}
	}

	// The rest of the config values should be set by Cobra, not here

	return config, nil
}

// GenerateEnvFile creates a template .env file with only the six specific values
func (c *Config) GenerateEnvFile() error {
	// Only include the six specific values in the .env file
	defaults := []string{
		"# Zig & ZLS Installer Configuration",
		"# Directories",
		fmt.Sprintf("ZIG_DIR=%s", c.ZigDir),
		fmt.Sprintf("ZLS_DIR=%s", c.ZLSDir),
		fmt.Sprintf("BIN_DIR=%s", c.BinDir),
		"",
		"# Zig download and verification",
		fmt.Sprintf("ZIG_PUB_KEY=%s", c.ZigPubKey),
		fmt.Sprintf("ZIG_DOWN_URL=%s", c.ZigDownURL),
		fmt.Sprintf("ZIG_INDEX_URL=%s", c.ZigIndexURL),
	}

	// Create or overwrite the file
	f, err := os.Create(c.EnvFile)
	if err != nil {
		return fmt.Errorf("could not create .env file: %w", err)
	}
	defer f.Close()

	for _, line := range defaults {
		if _, err := f.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("could not write to .env file: %w", err)
		}
	}

	return nil
}

// PrintSettings displays the current configuration
func (c *Config) PrintSettings(noColor bool) {
	if noColor {
		fmt.Println("Current Settings:")
		fmt.Printf("ZIG_DIR: %s\n", c.ZigDir)
		fmt.Printf("ZLS_DIR: %s\n", c.ZLSDir)
		fmt.Printf("BIN_DIR: %s\n", c.BinDir)
		fmt.Printf("ZIG_PUB_KEY: %s\n", c.ZigPubKey)
		fmt.Printf("ZIG_DOWN_URL: %s\n", c.ZigDownURL)
		fmt.Printf("ZIG_INDEX_URL: %s\n", c.ZigIndexURL)
		fmt.Printf("Environment file: %s\n", c.EnvFile)
	} else {
		// In the actual implementation, this would use styles from the theme package
		fmt.Println("Current Settings (colored output would be here):")
		fmt.Printf("ZIG_DIR: %s\n", c.ZigDir)
		fmt.Printf("ZLS_DIR: %s\n", c.ZLSDir)
		fmt.Printf("BIN_DIR: %s\n", c.BinDir)
		fmt.Printf("ZIG_PUB_KEY: %s\n", c.ZigPubKey)
		fmt.Printf("ZIG_DOWN_URL: %s\n", c.ZigDownURL)
		fmt.Printf("ZIG_INDEX_URL: %s\n", c.ZigIndexURL)
		fmt.Printf("Environment file: %s\n", c.EnvFile)
	}
}

// EnsureDirectories creates the necessary directories if they don't exist
func (c *Config) EnsureDirectories() error {
	dirs := []string{c.ZigDir, c.ZLSDir}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Ensure bin directory exists
	if _, err := os.Stat(c.BinDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(c.BinDir, 0755); err != nil {
				return fmt.Errorf("failed to create bin directory %s: %w", c.BinDir, err)
			}
		} else {
			return fmt.Errorf("error checking bin directory: %w", err)
		}
	}

	return nil
}
