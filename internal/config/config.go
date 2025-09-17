package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
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
	defaults := GetDefaults()
	return &Config{
		// Default values for .env configurable settings
		ZigDir:      defaults.ZigDir,
		ZLSDir:      defaults.ZLSDir,
		BinDir:      defaults.BinDir,
		ZigPubKey:   defaults.ZigPubKey,
		ZigDownURL:  defaults.ZigDownURL,
		ZigIndexURL: defaults.ZigIndexURL,

		// Default values for CLI options
		EnvFile:   defaults.EnvFile,
		Verbose:   defaults.Verbose,
		LogFile:   defaults.LogFile,
		EnableLog: defaults.EnableLog,
	}
}

// InitViper initializes Viper with platform-specific defaults
func InitViper() *viper.Viper {
	v := viper.New()
	defaults := GetDefaults()

	// Set default values
	v.SetDefault("zig_dir", defaults.ZigDir)
	v.SetDefault("zls_dir", defaults.ZLSDir)
	v.SetDefault("bin_dir", defaults.BinDir)
	v.SetDefault("zig_pub_key", defaults.ZigPubKey)
	v.SetDefault("zig_down_url", defaults.ZigDownURL)
	v.SetDefault("zig_index_url", defaults.ZigIndexURL)

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

			// Only override values if they are set in the .env file
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

	return config, nil
}

// GenerateEnvFile creates a template .env file with default values
func (c *Config) GenerateEnvFile() error {
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

	f, err := os.Create(c.EnvFile)
	if err != nil {
		return fmt.Errorf("could not create .env file: %w", err)
	}
	defer func() { _ = f.Close() }()

	for _, line := range defaults {
		if _, err := f.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("could not write to .env file: %w", err)
		}
	}

	return nil
}

// PrintSettings displays the current configuration
func (c *Config) PrintSettings() {
	fmt.Println("Current Settings:")
	fmt.Printf("ZIG_DIR: %s\n", c.ZigDir)
	fmt.Printf("ZLS_DIR: %s\n", c.ZLSDir)
	fmt.Printf("BIN_DIR: %s\n", c.BinDir)
	fmt.Printf("ZIG_PUB_KEY: %s\n", c.ZigPubKey)
	fmt.Printf("ZIG_DOWN_URL: %s\n", c.ZigDownURL)
	fmt.Printf("ZIG_INDEX_URL: %s\n", c.ZigIndexURL)
	fmt.Printf("Environment file: %s\n", c.EnvFile)
}

// EnsureDirectories creates the necessary directories if they don't exist
func (c *Config) EnsureDirectories() error {
	dirs := []string{c.ZigDir, c.ZLSDir}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	if _, err := os.Stat(c.BinDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(c.BinDir, 0o755); err != nil {
				return fmt.Errorf("failed to create bin directory %s: %w", c.BinDir, err)
			}
		} else {
			return fmt.Errorf("error checking bin directory: %w", err)
		}
	}

	return nil
}
