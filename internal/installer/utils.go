package installer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"

	"github.com/exilesprx/zig-install/internal/config"
	"github.com/exilesprx/zig-install/internal/tui"
)

// Global variable to hold styles for printing
var globalStyles *tui.Styles
var globalConfig *config.Config

// SetGlobalConfig sets the global config and styles for task printing
func SetGlobalConfig(config *config.Config, styles *tui.Styles) {
	globalConfig = config
	globalStyles = styles
}

// PrintTask prints a task completion message with optional detailed output
func PrintTask(name, status, output string) {
	if globalConfig == nil || globalStyles == nil {
		// Fallback to simple print if globals aren't set
		fmt.Printf("%s %s\n", status, name)
		if output != "" {
			fmt.Printf("  %s\n", output)
		}
		return
	}

	if globalConfig.NoColor {
		fmt.Printf("%s %s\n", status, name)
		if output != "" && globalConfig.Verbose {
			fmt.Printf("  %s\n", output)
		}
	} else {
		fmt.Println(globalStyles.Success.Render(fmt.Sprintf("%s %s", status, name)))
		if output != "" && globalConfig.Verbose {
			fmt.Println(globalStyles.Detail.Render(fmt.Sprintf("  %s", output)))
		}
	}
}

// getZigVersion fetches version information from ziglang.org
func getZigVersion(zigIndexURL string, requestedVersion string) (*ZigVersionInfo, error) {
	resp, err := http.Get(zigIndexURL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var versions map[string]ZigVersionInfo
	if err := json.Unmarshal(body, &versions); err != nil {
		return nil, err
	}

	version := requestedVersion
	if version == "" || version == "master" {
		version = "master"
	}

	versionInfo, ok := versions[version]
	if !ok {
		return nil, fmt.Errorf("version %s not found", version)
	}

	// For non-master versions, use the key from the JSON as the version
	// since these entries don't have a version field
	if version != "master" {
		versionInfo.Version = version
	} else if versionInfo.Version == "" {
		// Only check for empty version on master, since it should have one
		return nil, fmt.Errorf("could not determine master version")
	}

	return &versionInfo, nil
}

// isZigInstalled checks if the specified version is already installed
func isZigInstalled(version string) bool {
	cmd := exec.Command("zig", "version")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Check if version string contains the specified version
	// This handles cases like "0.11.0" matching "0.11.0-dev.3180+hash"
	installedVersion := strings.TrimSpace(string(output))

	// Handle master version specially
	if version == "master" {
		return strings.Contains(installedVersion, "-dev.")
	}

	return strings.HasPrefix(installedVersion, version)
}
