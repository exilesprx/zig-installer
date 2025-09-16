package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/exilesprx/zig-install/internal/config"
	"github.com/exilesprx/zig-install/internal/logger"
	"github.com/exilesprx/zig-install/internal/tui"
)

// ZigBuildInfo represents information about a specific Zig build
type ZigBuildInfo struct {
	Tarball string `json:"tarball"`
	Shasum  string `json:"shasum"`
	Size    string `json:"size"`
}

// ZigVersionInfo represents version-specific information
type ZigVersionInfo struct {
	Version      string       `json:"version"`
	Date         string       `json:"date"`
	Docs         string       `json:"docs"`
	StdDocs      string       `json:"stdDocs"`
	Src          ZigBuildInfo `json:"src"`
	Bootstrap    ZigBuildInfo `json:"bootstrap"`
	X86_64MacOS  ZigBuildInfo `json:"x86_64-macos"`
	Aarch64MacOS ZigBuildInfo `json:"aarch64-macos"`
	X86_64Linux  ZigBuildInfo `json:"x86_64-linux"`
	Aarch64Linux ZigBuildInfo `json:"aarch64-linux"`
}

// getPlatformBuildInfo returns the appropriate build info for the current platform
func getPlatformBuildInfo(versionInfo *ZigVersionInfo) (*ZigBuildInfo, error) {
	arch := runtime.GOARCH
	switch runtime.GOOS {
	case "darwin":
		switch arch {
		case "amd64":
			return &versionInfo.X86_64MacOS, nil
		case "arm64":
			return &versionInfo.Aarch64MacOS, nil
		}
	case "linux":
		switch arch {
		case "amd64":
			return &versionInfo.X86_64Linux, nil
		case "arm64":
			return &versionInfo.Aarch64Linux, nil
		}
	}
	return nil, fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, arch)
}

// InstallZig handles the Zig installation process
func InstallZig(p *tea.Program, config *config.Config, logger logger.ILogger, requestedVersion string) (string, error) {
	// Get the version info
	msg := "Fetching latest Zig version..."
	if requestedVersion != "" && requestedVersion != "master" {
		msg = fmt.Sprintf("Fetching Zig version %s...", requestedVersion)
	}
	p.Send(tui.StatusMsg(msg))

	versionInfo, err := getZigVersion(config.ZigIndexURL, requestedVersion)
	if err != nil {
		return "", err
	}

	version := versionInfo.Version
	p.Send(tui.StatusMsg(fmt.Sprintf("Using Zig version: %s", version)))

	// Get platform-specific build info
	buildInfo, err := getPlatformBuildInfo(versionInfo)
	if err != nil {
		return "", err
	}

	// Check if already installed
	if isZigInstalled(version) {
		p.Send(tui.StatusMsg(fmt.Sprintf("Zig %s is already installed.", version)))
		return version, nil
	}

	// Prepare directories
	if err := os.MkdirAll(config.ZigDir, 0o755); err != nil {
		return "", fmt.Errorf("could not create directory %s: %w", config.ZigDir, err)
	}

	// Get the username to set ownership
	user := os.Getenv("SUDO_USER")
	if user == "" {
		user = os.Getenv("USER")
	}

	// Set ownership
	if user != "" {
		sendDetailedOutputMsg(p, fmt.Sprintf("Setting ownership of %s to %s", config.ZigDir, user), config.Verbose)
		cmd := exec.Command("chown", "-R", user+":"+user, config.ZigDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			sendDetailedOutputMsg(p, fmt.Sprintf("Error setting ownership: %s", output), config.Verbose)
			return "", fmt.Errorf("could not set ownership of %s: %w", config.ZigDir, err)
		} else {
			sendDetailedOutputMsg(p, string(output), config.Verbose)
		}
	}

	// Download Zig
	tarURL := buildInfo.Tarball
	tarFile := filepath.Base(tarURL)
	tarPath := filepath.Join(config.ZigDir, tarFile)
	sigPath := tarPath + ".minisig"

	p.Send(tui.StatusMsg(fmt.Sprintf("Downloading Zig %s...", version)))

	sendDetailedOutputMsg(p, fmt.Sprintf("Downloading from %s to %s", tarURL, tarPath), config.Verbose)

	cmd := exec.Command("wget", "-O", tarPath, tarURL)
	if output, err := cmd.CombinedOutput(); err != nil {
		sendDetailedOutputMsg(p, fmt.Sprintf("Error downloading: %s", output), config.Verbose)
		return "", fmt.Errorf("could not download Zig: %w", err)
	} else {
		sendDetailedOutputMsg(p, string(output), config.Verbose)
	}

	// Download signature
	p.Send(tui.StatusMsg("Downloading Zig signature..."))
	sendDetailedOutputMsg(p, fmt.Sprintf("Downloading signature from %s.minisig to %s", tarURL, sigPath), config.Verbose)

	cmd = exec.Command("wget", "-O", sigPath, tarURL+".minisig")
	if output, err := cmd.CombinedOutput(); err != nil {
		sendDetailedOutputMsg(p, fmt.Sprintf("Error downloading signature: %s", output), config.Verbose)
		return "", fmt.Errorf("could not download Zig signature: %w", err)
	} else {
		sendDetailedOutputMsg(p, string(output), config.Verbose)
	}

	// Verify signature
	p.Send(tui.StatusMsg("Verifying Zig download..."))
	sendDetailedOutputMsg(p, fmt.Sprintf("Verifying %s with key %s", tarPath, config.ZigPubKey), config.Verbose)

	output, err := exec.Command("minisign", "-Vm", tarPath, "-P", config.ZigPubKey).CombinedOutput()
	if err != nil {
		// Clean up files if verification fails
		_ = os.Remove(tarPath)
		_ = os.Remove(sigPath)
		sendDetailedOutputMsg(p, fmt.Sprintf("Verification failed: %s", output), config.Verbose)
		return "", fmt.Errorf("signature verification failed: %w: %s", err, output)
	}

	sendDetailedOutputMsg(p, string(output), config.Verbose)

	// Remove signature file after verification
	_ = os.Remove(sigPath)

	// Extract Zig
	p.Send(tui.StatusMsg("Extracting Zig..."))
	sendDetailedOutputMsg(p, fmt.Sprintf("Extracting %s to %s", tarPath, config.ZigDir), config.Verbose)

	output, err = exec.Command("tar", "-xf", tarPath, "-C", config.ZigDir).CombinedOutput()
	if err != nil {
		sendDetailedOutputMsg(p, fmt.Sprintf("Extraction failed: %s", output), config.Verbose)
		return "", fmt.Errorf("extraction failed: %w", err)
	} else {
		sendDetailedOutputMsg(p, string(output), config.Verbose)
	}

	// Remove tar file after extraction
	_ = os.Remove(tarPath)

	// The extracted directory name is the same as the tarball name without the .tar.xz extension
	extractedDir := strings.TrimSuffix(tarFile, ".tar.xz")

	// Create symbolic link
	zigBinPath := filepath.Join(config.ZigDir, extractedDir, "zig")
	linkPath := filepath.Join(config.BinDir, "zig")

	p.Send(tui.StatusMsg("Creating symbolic link..."))
	sendDetailedOutputMsg(p, fmt.Sprintf("Creating symlink from %s to %s", zigBinPath, linkPath), config.Verbose)

	if _, err := os.Stat(linkPath); err == nil {
		_ = os.Remove(linkPath)
	}
	if err := os.Symlink(zigBinPath, linkPath); err != nil {
		return "", fmt.Errorf("could not create symbolic link: %w", err)
	}

	return version, nil
}
