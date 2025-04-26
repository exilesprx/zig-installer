package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Build-time configurable defaults that can be set with linker flags
var (
	// DefaultZigPubKey can be set with -ldflags="-X 'github.com/exilesprx/zig-install/internal/config.DefaultZigPubKey=key'"
	DefaultZigPubKey = "RWSGOq2NVecA2UPNdBUZykf1CCb147pkmdtYxgb3Ti+JO/wCYvhbAb/U"

	// DefaultZigDownURL can be set with linker flags if needed
	DefaultZigDownURL = "https://ziglang.org/builds/"

	// DefaultZigIndexURL can be set with linker flags if needed
	DefaultZigIndexURL = "https://ziglang.org/download/index.json"
)

// Config contains the application configuration
type Config struct {
	ZigDir       string
	ZLSDir       string
	BinDir       string
	ZigPubKey    string
	ZigDownURL   string
	ZigIndexURL  string
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
		ZigDir:      "/opt/zig",
		ZLSDir:      "/opt/zls",
		BinDir:      "/usr/local/bin",
		ZigPubKey:   DefaultZigPubKey,   // Use the linker-configurable default
		ZigDownURL:  DefaultZigDownURL,  // Use the linker-configurable default
		ZigIndexURL: DefaultZigIndexURL, // Use the linker-configurable default
		EnvFile:     ".env",
		Verbose:     false,
		LogFile:     "zig-install.log",
		EnableLog:   true,
	}
}

// InitViper initializes Viper with config values
func InitViper() *viper.Viper {
	v := viper.New()

	// Set default values
	v.SetDefault("zig_dir", "/opt/zig")
	v.SetDefault("zls_dir", "/opt/zls")
	v.SetDefault("bin_dir", "/usr/local/bin")
	v.SetDefault("zig_pub_key", DefaultZigPubKey)     // Use the linker-configurable default
	v.SetDefault("zig_down_url", DefaultZigDownURL)   // Use the linker-configurable default
	v.SetDefault("zig_index_url", DefaultZigIndexURL) // Use the linker-configurable default
	v.SetDefault("verbose", false)
	v.SetDefault("log_file", "zig-install.log")
	v.SetDefault("enable_log", true)

	// Set environment variable prefix
	v.SetEnvPrefix("ZIG_INSTALL")
	v.AutomaticEnv()

	return v
}

// LoadConfig loads configuration from environment file and environment variables
func LoadConfig(v *viper.Viper, envFile string) (*Config, error) {
	if envFile != "" {
		// Check if the env file exists
		if _, err := os.Stat(envFile); err == nil {
			v.SetConfigFile(envFile)
			if err := v.ReadInConfig(); err != nil {
				return nil, fmt.Errorf("error reading config file: %w", err)
			}
		}
	}

	config := &Config{
		ZigDir:       v.GetString("zig_dir"),
		ZLSDir:       v.GetString("zls_dir"),
		BinDir:       v.GetString("bin_dir"),
		ZigPubKey:    v.GetString("zig_pub_key"),
		ZigDownURL:   v.GetString("zig_down_url"),
		ZigIndexURL:  v.GetString("zig_index_url"),
		EnvFile:      envFile,
		ZigOnly:      v.GetBool("zig_only"),
		ZLSOnly:      v.GetBool("zls_only"),
		NoColor:      v.GetBool("no_color"),
		Verbose:      v.GetBool("verbose"),
		LogFile:      v.GetString("log_file"),
		EnableLog:    v.GetBool("enable_log"),
		GenerateEnv:  v.GetBool("generate_env"),
		ShowSettings: v.GetBool("show_settings"),
	}

	return config, nil
}

// GenerateEnvFile creates a template .env file with default values
func (c *Config) GenerateEnvFile() error {
	// Default configuration values
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
