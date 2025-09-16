package installer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// sendDetailedOutputMsg sends detailed output messages to the program if verbose mode is enabled
func sendDetailedOutputMsg(p *tea.Program, msg string, verbose bool) {
	if !verbose || len(msg) == 0 {
		return
	}
	p.Send(msg)
}

// getZigVersion fetches version information from ziglang.org
func getZigVersion(zigIndexURL string, requestedVersion string) (*ZigVersionInfo, error) {
	resp, err := http.Get(zigIndexURL)
	if err != nil {
		return nil, err
	}
	defer func() { resp.Body.Close() }()

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

// convertToSemanticVersion converts a Zig version to a ZLS tag format
func convertToSemanticVersion(zigVersion string) string {
	if zigVersion == "" {
		return ""
	}

	// Handle master/dev versions specially
	if zigVersion == "master" || strings.Contains(zigVersion, "-dev.") {
		return "master"
	}

	// Convert input version to semantic version format
	version := zigVersion
	if idx := strings.Index(version, "-"); idx != -1 {
		version = version[:idx]
	}

	// Split into components
	parts := strings.Split(version, ".")
	if len(parts) >= 3 {
		return fmt.Sprintf("%s.%s.%s", parts[0], parts[1], parts[2])
	}
	if len(parts) == 2 {
		return fmt.Sprintf("%s.%s.0", parts[0], parts[1])
	}
	return ""
}
