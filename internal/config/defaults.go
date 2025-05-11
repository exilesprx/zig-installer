package config

import (
	"runtime"
)

// Common defaults that are the same across all platforms
const (
	defaultZigPubKey   = "RWSGOq2NVecA2UPNdBUZykf1CCb147pkmdtYxgb3Ti+JO/wCYvhbAb/U"
	defaultZigDownURL  = "https://ziglang.org/builds/"
	defaultZigIndexURL = "https://ziglang.org/download/index.json"
)

// DefaultConfig contains the default configuration values
type DefaultConfig struct {
	ZigDir      string
	ZLSDir      string
	BinDir      string
	ZigPubKey   string
	ZigDownURL  string
	ZigIndexURL string
}

// getPlatformPaths returns the platform-specific paths
func getPlatformPaths() (zigDir, zlsDir, binDir string) {
	switch runtime.GOOS {
	case "darwin":
		return "/usr/local/zig",
			"/usr/local/zls",
			"/usr/local/bin"
	default: // linux and others
		return "/opt/zig",
			"/opt/zls",
			"/usr/local/bin"
	}
}

// GetDefaults returns platform-specific default configuration values
func GetDefaults() *DefaultConfig {
	zigDir, zlsDir, binDir := getPlatformPaths()

	return &DefaultConfig{
		ZigDir:      zigDir,
		ZLSDir:      zlsDir,
		BinDir:      binDir,
		ZigPubKey:   defaultZigPubKey,
		ZigDownURL:  defaultZigDownURL,
		ZigIndexURL: defaultZigIndexURL,
	}
}
