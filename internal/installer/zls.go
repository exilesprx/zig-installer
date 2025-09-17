package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/exilesprx/zig-install/internal/config"
	"github.com/exilesprx/zig-install/internal/logger"
	"github.com/exilesprx/zig-install/internal/tui"
)

// InstallZLS handles the ZLS installation process
func InstallZLS(p *tea.Program, config *config.Config, logger logger.ILogger, zigVersion string) error {
	// Prepare directories
	if err := os.MkdirAll(config.ZLSDir, 0o755); err != nil {
		return fmt.Errorf("could not create directory %s: %w", config.ZLSDir, err)
	}

	// Get the username to set ownership
	user := os.Getenv("SUDO_USER")
	if user == "" {
		user = os.Getenv("USER")
	}

	// Set initial directory ownership
	if user != "" {
		sendDetailedOutputMsg(p, fmt.Sprintf("Setting ownership of %s to %s", config.ZLSDir, user), config.Verbose)
		cmd := exec.Command("chown", "-R", user+":"+user, config.ZLSDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			sendDetailedOutputMsg(p, fmt.Sprintf("Error setting ownership: %s", output), config.Verbose)
			return fmt.Errorf("could not set ownership of %s: %w", config.ZLSDir, err)
		} else {
			sendDetailedOutputMsg(p, string(output), config.Verbose)
		}
	}

	// First determine if we're installing a specific version
	version := convertToSemanticVersion(zigVersion)
	logger.LogInfo("Zig version detected: %s, converted to ZLS version: %s", zigVersion, version)
	isSpecificVersion := version != "" && version != "master"

	// Check if repo already exists
	isRepoCloned := false
	if _, err := os.Stat(filepath.Join(config.ZLSDir, ".git")); err == nil {
		// Verify it's the correct repo
		cmd := exec.Command("git", "config", "--get", "remote.origin.url")
		cmd.Dir = config.ZLSDir
		if output, err := cmd.CombinedOutput(); err == nil && strings.Contains(string(output), "zigtools/zls") {
			isRepoCloned = true
		}
	}

	// Clone repository if it doesn't exist
	if !isRepoCloned {
		p.Send(tui.StatusMsg("Cloning ZLS repository..."))

		// Clean directory if it exists but isn't a valid repo
		if err := os.RemoveAll(config.ZLSDir); err != nil {
			return fmt.Errorf("could not clean ZLS directory: %w", err)
		}

		cmd := exec.Command("git", "clone", "https://github.com/zigtools/zls.git", config.ZLSDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			sendDetailedOutputMsg(p, fmt.Sprintf("Error cloning repository: %s", output), config.Verbose)
			return fmt.Errorf("could not clone ZLS repository: %w", err)
		}
	}

	// Handle version-specific installation
	if isSpecificVersion {
		p.Send(tui.StatusMsg(fmt.Sprintf("Setting up ZLS version %s...", version)))

		// Fetch tags
		fetchCmd := exec.Command("git", "fetch", "--tags")
		fetchCmd.Dir = config.ZLSDir
		if output, err := fetchCmd.CombinedOutput(); err != nil {
			sendDetailedOutputMsg(p, fmt.Sprintf("Error fetching tags: %s", output), config.Verbose)
			return fmt.Errorf("could not fetch tags: %w", err)
		}

		// Check if tag exists
		checkCmd := exec.Command("git", "rev-parse", "--verify", "refs/tags/"+version)
		checkCmd.Dir = config.ZLSDir
		if err := checkCmd.Run(); err != nil {
			sendDetailedOutputMsg(p, fmt.Sprintf("Error checking out version: %s", err), config.Verbose)
			return fmt.Errorf("ZLS version %s not found", version)
		}

		// Checkout the specific version
		checkoutCmd := exec.Command("git", "checkout", version)
		checkoutCmd.Dir = config.ZLSDir
		if output, err := checkoutCmd.CombinedOutput(); err != nil {
			sendDetailedOutputMsg(p, fmt.Sprintf("Error checking out version: %s", output), config.Verbose)
			return fmt.Errorf("could not checkout version %s: %w", version, err)
		}
	} else {
		// Just update to latest master
		p.Send(tui.StatusMsg("Setting up latest ZLS..."))

		// First try to checkout master/main
		checkoutCmd := exec.Command("git", "checkout", "master")
		checkoutCmd.Dir = config.ZLSDir
		if err := checkoutCmd.Run(); err != nil {
			// Try main if master fails
			checkoutCmd = exec.Command("git", "checkout", "main")
			checkoutCmd.Dir = config.ZLSDir
			if err := checkoutCmd.Run(); err != nil {
				return fmt.Errorf("could not checkout master/main branch")
			}
		}

		// Pull latest changes
		pullCmd := exec.Command("git", "pull")
		pullCmd.Dir = config.ZLSDir
		if output, err := pullCmd.CombinedOutput(); err != nil {
			sendDetailedOutputMsg(p, fmt.Sprintf("Error pulling latest changes: %s", output), config.Verbose)
			return fmt.Errorf("could not pull latest changes: %w", err)
		}
	}

	// Set ownership after git operations if not root user
	if user != "" {
		sendDetailedOutputMsg(p, fmt.Sprintf("Setting ownership after git operations for %s", config.ZLSDir), config.Verbose)
		cmd := exec.Command("chown", "-R", user+":"+user, config.ZLSDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			sendDetailedOutputMsg(p, fmt.Sprintf("Error: %s", output), config.Verbose)
			return fmt.Errorf("could not set ownership of %s: %w", config.ZLSDir, err)
		} else {
			sendDetailedOutputMsg(p, string(output), config.Verbose)
		}
	}

	// Build ZLS
	p.Send(tui.StatusMsg("Building ZLS..."))

	sendDetailedOutputMsg(p, fmt.Sprintf("Running: zig build -Doptimize=ReleaseSafe in %s", config.ZLSDir), config.Verbose)

	cmd := exec.Command("zig", "build", "-Doptimize=ReleaseSafe")
	cmd.Dir = config.ZLSDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		sendDetailedOutputMsg(p, fmt.Sprintf("Error building ZLS: %s", output), config.Verbose)
		return fmt.Errorf("could not build ZLS: %w", err)
	}

	sendDetailedOutputMsg(p, string(output), config.Verbose)

	// Set ownership after building if not root user
	if user != "" {
		buildOutDir := filepath.Join(config.ZLSDir, "zig-out")
		sendDetailedOutputMsg(p, fmt.Sprintf("Setting ownership of build output in %s", buildOutDir), config.Verbose)
		cmd := exec.Command("chown", "-R", user+":"+user, buildOutDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			sendDetailedOutputMsg(p, fmt.Sprintf("Error setting ownership: %s", output), config.Verbose)
			return fmt.Errorf("could not set ownership of build output: %w", err)
		} else {
			sendDetailedOutputMsg(p, string(output), config.Verbose)
		}
	}

	// Create symbolic link
	zlsBinPath := filepath.Join(config.ZLSDir, "zig-out", "bin", "zls")
	linkPath := filepath.Join(config.BinDir, "zls")

	p.Send(tui.StatusMsg("Creating symbolic link..."))

	sendDetailedOutputMsg(p, fmt.Sprintf("Creating symlink from %s to %s", zlsBinPath, linkPath), config.Verbose)

	if _, err := os.Stat(linkPath); err == nil {
		_ = os.Remove(linkPath)
		sendDetailedOutputMsg(p, fmt.Sprintf("Removed existing symlink at %s", linkPath), config.Verbose)
	}

	if err := os.Symlink(zlsBinPath, linkPath); err != nil {
		return fmt.Errorf("could not create symbolic link: %w", err)
	}

	sendDetailedOutputMsg(p, fmt.Sprintf("Symlink created at %s", linkPath), config.Verbose)

	return nil
}
