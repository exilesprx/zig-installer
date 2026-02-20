package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Common defaults that are the same across all platforms
const (
	defaultZigPubKey   = "RWSGOq2NVecA2UPNdBUZykf1CCb147pkmdtYxgb3Ti+JO/wCYvhbAb/U"
	defaultZigDownURL  = "https://ziglang.org/builds/"
	defaultZigIndexURL = "https://ziglang.org/download/index.json"
	defaultEnvFile     = ".env"
	defaultVerbose     = false
	defaultLogFile     = "zig-install.log"
	defaultEnableLog   = false
)

// DefaultConfig contains the default configuration values
type DefaultConfig struct {
	ZigDir      string
	ZLSDir      string
	BinDir      string
	ZigPubKey   string
	ZigDownURL  string
	ZigIndexURL string
	EnvFile     string
	Verbose     bool
	LogFile     string
	EnableLog   bool
}

// getUserLocalPaths returns user-local installation paths
func getUserLocalPaths() (zigDir, zlsDir, binDir string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", "", fmt.Errorf("could not determine home directory: %w", err)
	}

	return filepath.Join(home, ".local", "share", "zig"),
		filepath.Join(home, ".local", "share", "zls"),
		filepath.Join(home, ".local", "bin"),
		nil
}

// GetSystemZigDirs returns potential system installation directories for detection
func GetSystemZigDirs() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{"/usr/local/zig", "/opt/zig"}
	default: // linux
		return []string{"/opt/zig", "/usr/local/zig"}
	}
}

// DetectSystemInstallation checks if Zig is installed system-wide
func DetectSystemInstallation() (string, bool) {
	// Check for system directories with content
	for _, dir := range GetSystemZigDirs() {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			// Check if it actually contains Zig installations
			entries, err := os.ReadDir(dir)
			if err == nil && len(entries) > 0 {
				return dir, true
			}
		}
	}

	// Check for system symlinks (even if broken)
	systemBinLinks := []string{"/usr/local/bin/zig", "/usr/local/bin/zls"}
	for _, link := range systemBinLinks {
		if _, err := os.Lstat(link); err == nil {
			// Symlink exists - check if it points to a system path
			target, err := os.Readlink(link)
			if err == nil {
				// If symlink points to /opt or /usr/local, consider it a system installation
				if strings.HasPrefix(target, "/opt/") || strings.HasPrefix(target, "/usr/local/") {
					return filepath.Dir(target), true
				}
			}
		}
	}

	return "", false
}

// GetDefaults returns user-local default configuration values
func GetDefaults() *DefaultConfig {
	zigDir, zlsDir, binDir, err := getUserLocalPaths()
	if err != nil {
		// Fatal error - can't proceed without home directory
		panic(fmt.Sprintf("Cannot determine user home directory: %v", err))
	}

	return &DefaultConfig{
		ZigDir:      zigDir,
		ZLSDir:      zlsDir,
		BinDir:      binDir,
		ZigPubKey:   defaultZigPubKey,
		ZigDownURL:  defaultZigDownURL,
		ZigIndexURL: defaultZigIndexURL,
		EnvFile:     defaultEnvFile,
		Verbose:     defaultVerbose,
		LogFile:     defaultLogFile,
		EnableLog:   defaultEnableLog,
	}
}
