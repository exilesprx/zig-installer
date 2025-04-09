package installer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/exilesprx/zig-install/internal/config"
	"github.com/exilesprx/zig-install/internal/logger"
	"github.com/exilesprx/zig-install/internal/tui"
)

// ZigJSON represents the JSON response from ziglang.org/download/index.json
type ZigJSON struct {
	Master struct {
		Version string `json:"version"`
	} `json:"master"`
}

// InstallZig handles the Zig installation process
func InstallZig(p *tea.Program, config *config.Config, logger logger.ILogger) (string, error) {
	// Get the latest version
	p.Send(tui.StatusMsg("Fetching latest Zig version..."))
	version, err := getLatestZigVersion(config.ZigIndexURL)
	if err != nil {
		return "", err
	}
	p.Send(tui.StatusMsg(fmt.Sprintf("Found latest Zig version: %s", version)))

	// Check if already installed
	if isZigInstalled(version) {
		p.Send(tui.StatusMsg(fmt.Sprintf("Zig %s is already installed.", version)))
		return version, nil
	}

	// Prepare directories
	if err := os.MkdirAll(config.ZigDir, 0755); err != nil {
		return "", fmt.Errorf("could not create directory %s: %w", config.ZigDir, err)
	}

	// Get the username to set ownership
	user := os.Getenv("SUDO_USER")
	if user == "" {
		user = os.Getenv("USER")
	}

	// Set ownership
	if user != "" {
		if config.Verbose {
			p.Send(tui.DetailOutputMsg(fmt.Sprintf("Setting ownership of %s to %s", config.ZigDir, user)))
		}
		cmd := exec.Command("chown", "-R", user+":"+user, config.ZigDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			if config.Verbose {
				p.Send(tui.DetailOutputMsg(fmt.Sprintf("Error: %s", output)))
			}
			return "", fmt.Errorf("could not set ownership of %s: %w", config.ZigDir, err)
		} else if config.Verbose && len(output) > 0 {
			p.Send(tui.DetailOutputMsg(string(output)))
		}
	}

	// Download Zig
	tarFile := fmt.Sprintf("zig-linux-x86_64-%s.tar.xz", version)
	tarPath := filepath.Join(config.ZigDir, tarFile)
	sigPath := tarPath + ".minisig"

	p.Send(tui.StatusMsg(fmt.Sprintf("Downloading Zig %s...", version)))

	if config.Verbose {
		p.Send(tui.DetailOutputMsg(fmt.Sprintf("Downloading from %s to %s", config.ZigDownURL+tarFile, tarPath)))
	}

	cmd := exec.Command("wget", "-O", tarPath, config.ZigDownURL+tarFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		if config.Verbose {
			p.Send(tui.DetailOutputMsg(fmt.Sprintf("Error downloading: %s", output)))
		}
		return "", fmt.Errorf("could not download Zig: %w", err)
	} else if config.Verbose {
		p.Send(tui.DetailOutputMsg(string(output)))
	}

	// Download signature
	p.Send(tui.StatusMsg("Downloading Zig signature..."))
	if config.Verbose {
		p.Send(tui.DetailOutputMsg(fmt.Sprintf("Downloading signature from %s to %s", config.ZigDownURL+tarFile+".minisig", sigPath)))
	}

	cmd = exec.Command("wget", "-O", sigPath, config.ZigDownURL+tarFile+".minisig")
	if output, err := cmd.CombinedOutput(); err != nil {
		if config.Verbose {
			p.Send(tui.DetailOutputMsg(fmt.Sprintf("Error downloading signature: %s", output)))
		}
		return "", fmt.Errorf("could not download Zig signature: %w", err)
	} else if config.Verbose {
		p.Send(tui.DetailOutputMsg(string(output)))
	}

	// Verify signature
	p.Send(tui.StatusMsg("Verifying Zig download..."))
	if config.Verbose {
		p.Send(tui.DetailOutputMsg(fmt.Sprintf("Verifying %s with key %s", tarPath, config.ZigPubKey)))
	}

	output, err := exec.Command("minisign", "-Vm", tarPath, "-P", config.ZigPubKey).CombinedOutput()
	if err != nil {
		// Clean up files if verification fails
		os.Remove(tarPath)
		os.Remove(sigPath)
		if config.Verbose {
			p.Send(tui.DetailOutputMsg(fmt.Sprintf("Verification failed: %s", output)))
		}
		return "", fmt.Errorf("signature verification failed: %w: %s", err, output)
	}

	if config.Verbose {
		p.Send(tui.DetailOutputMsg(string(output)))
	}

	// Remove signature file after verification
	os.Remove(sigPath)

	// Extract Zig
	p.Send(tui.StatusMsg("Extracting Zig..."))
	if config.Verbose {
		p.Send(tui.DetailOutputMsg(fmt.Sprintf("Extracting %s to %s", tarPath, config.ZigDir)))
	}

	output, err = exec.Command("tar", "-xf", tarPath, "-C", config.ZigDir).CombinedOutput()
	if err != nil {
		if config.Verbose {
			p.Send(tui.DetailOutputMsg(fmt.Sprintf("Extraction failed: %s", output)))
		}
		return "", fmt.Errorf("extraction failed: %w", err)
	} else if config.Verbose && len(output) > 0 {
		p.Send(tui.DetailOutputMsg(string(output)))
	}

	// Remove tar file after extraction
	os.Remove(tarPath)

	// Create symbolic link
	zigBinPath := filepath.Join(config.ZigDir, fmt.Sprintf("zig-linux-x86_64-%s", version), "zig")
	linkPath := filepath.Join(config.BinDir, "zig")

	p.Send(tui.StatusMsg("Creating symbolic link..."))
	if config.Verbose {
		p.Send(tui.DetailOutputMsg(fmt.Sprintf("Creating symlink from %s to %s", zigBinPath, linkPath)))
	}

	if _, err := os.Stat(linkPath); err == nil {
		os.Remove(linkPath)
	}
	if err := os.Symlink(zigBinPath, linkPath); err != nil {
		return "", fmt.Errorf("could not create symbolic link: %w", err)
	}

	return version, nil
}

// Helper functions

// getLatestZigVersion fetches the latest version from ziglang.org
func getLatestZigVersion(zigIndexURL string) (string, error) {
	resp, err := http.Get(zigIndexURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var zigJSON ZigJSON
	if err := json.Unmarshal(body, &zigJSON); err != nil {
		return "", err
	}

	if zigJSON.Master.Version == "" {
		return "", fmt.Errorf("could not determine latest Zig version")
	}

	return zigJSON.Master.Version, nil
}

// isZigInstalled checks if the specified version is already installed
func isZigInstalled(version string) bool {
	cmd := exec.Command("zig", "version")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == version
}
