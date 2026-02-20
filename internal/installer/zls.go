package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/exilesprx/zig-installer/internal/config"
	"github.com/exilesprx/zig-installer/internal/logger"
)

// InstallZLS handles the ZLS installation process
func InstallZLS(p interface{}, config *config.Config, logger logger.ILogger, formatter OutputFormatter, zigVersion string) error {
	// Prepare directories
	if err := os.MkdirAll(config.ZLSDir, 0o755); err != nil {
		return fmt.Errorf("could not create directory %s: %w", config.ZLSDir, err)
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
		output, err := cmd.Output()
		if err == nil && strings.Contains(string(output), "zigtools/zls") {
			isRepoCloned = true
			formatter.PrintSuccess("Repository check", "ZLS repository already exists")
		}
	}

	if !isRepoCloned {
		formatter.PrintProgress("Repository clone", "Cloning ZLS repository...")

		// Clone the repository
		cmd := exec.Command("git", "clone", "https://github.com/zigtools/zls.git", config.ZLSDir)
		output, err := cmd.CombinedOutput()
		if err != nil {
			formatter.PrintError("Repository clone", fmt.Sprintf("Error cloning repository: %s", output))
			return fmt.Errorf("could not clone ZLS repository: %w", err)
		}

		formatter.PrintSuccess("ZLS clone", fmt.Sprintf("Cloned repository to %s", config.ZLSDir))
	}

	// Handle version-specific setup
	if isSpecificVersion {
		formatter.PrintProgress("Version setup", fmt.Sprintf("Setting up ZLS version %s...", version))

		// Fetch all tags
		cmd := exec.Command("git", "fetch", "--tags")
		cmd.Dir = config.ZLSDir
		if output, err := cmd.CombinedOutput(); err != nil {
			formatter.PrintError("Version setup", fmt.Sprintf("Error fetching tags: %s", output))
			return fmt.Errorf("could not fetch tags: %w", err)
		}

		// Verify the version exists
		cmd = exec.Command("git", "tag", "-l", version)
		cmd.Dir = config.ZLSDir
		output, err := cmd.Output()
		if err != nil || len(strings.TrimSpace(string(output))) == 0 {
			formatter.PrintError("Version setup", fmt.Sprintf("Version %s not found", version))
			return fmt.Errorf("version %s not found in ZLS repository", version)
		}

		// Checkout the specific version
		cmd = exec.Command("git", "checkout", version)
		cmd.Dir = config.ZLSDir
		if output, err := cmd.CombinedOutput(); err != nil {
			formatter.PrintError("Version setup", fmt.Sprintf("Error checking out version: %s", output))
			return fmt.Errorf("could not checkout version %s: %w", version, err)
		}

		formatter.PrintSuccess("ZLS version", fmt.Sprintf("Checked out version %s", version))
	} else {
		formatter.PrintProgress("Latest setup", "Setting up latest ZLS...")

		// For latest version, pull the latest changes
		if isRepoCloned {
			// Reset to ensure clean state
			cmd := exec.Command("git", "reset", "--hard", "HEAD")
			cmd.Dir = config.ZLSDir
			_ = cmd.Run() // Ignore errors for reset as it's a cleanup operation

			// Switch to master and pull latest
			cmd = exec.Command("git", "checkout", "master")
			cmd.Dir = config.ZLSDir
			_ = cmd.Run() // Ignore errors for checkout as pull will handle it

			cmd = exec.Command("git", "pull", "origin", "master")
			cmd.Dir = config.ZLSDir
			output, err := cmd.CombinedOutput()
			if err != nil {
				formatter.PrintError("Latest setup", fmt.Sprintf("Error pulling latest changes: %s", output))
				return fmt.Errorf("could not pull latest changes: %w", err)
			}
		}

		formatter.PrintSuccess("ZLS latest", "Updated to latest version")
	}

	// Build ZLS
	formatter.PrintProgress("ZLS build", "Building ZLS...")

	// Stream output in real-time if verbose mode is enabled
	if config.Verbose {
		formatter.PrintTask("Build details", "Info", fmt.Sprintf("Running: zig build -Doptimize=ReleaseSafe in %s", config.ZLSDir))

		cmd := exec.Command("zig", "build", "-Doptimize=ReleaseSafe", "--verbose")
		cmd.Dir = config.ZLSDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			formatter.PrintError("ZLS build", fmt.Sprintf("Error building ZLS: %v", err))
			return fmt.Errorf("could not build ZLS: %w", err)
		}
	} else {
		// Capture output for error reporting only
		cmd := exec.Command("zig", "build", "-Doptimize=ReleaseSafe")
		cmd.Dir = config.ZLSDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			formatter.PrintError("ZLS build", fmt.Sprintf("Error building ZLS: %s", output))
			return fmt.Errorf("could not build ZLS: %w", err)
		}
	}

	formatter.PrintSuccess("ZLS build", "ZLS built successfully")

	// Create symbolic link to ZLS binary
	zlsBinPath := filepath.Join(config.ZLSDir, "zig-out", "bin", "zls")
	linkPath := filepath.Join(config.BinDir, "zls")

	formatter.PrintProgress("ZLS symlink setup", fmt.Sprintf("Creating symlink from %s to %s", zlsBinPath, linkPath))

	// Remove existing symlink/file if it exists
	if _, err := os.Lstat(linkPath); err == nil {
		if err := os.Remove(linkPath); err != nil {
			formatter.PrintError("ZLS symlink setup", fmt.Sprintf("Failed to remove existing symlink: %v", err))
			return fmt.Errorf("could not remove existing symlink: %w", err)
		}
	}

	if err := os.Symlink(zlsBinPath, linkPath); err != nil {
		return fmt.Errorf("could not create symbolic link: %w", err)
	}

	formatter.PrintSuccess("ZLS symbolic link setup", fmt.Sprintf("Created symlink: %s -> %s", linkPath, zlsBinPath))

	return nil
}

// convertToSemanticVersion converts Zig version format to ZLS tag format
func convertToSemanticVersion(zigVersion string) string {
	if zigVersion == "" {
		return ""
	}

	// Handle master/dev versions specially - they should use master branch
	if zigVersion == "master" || strings.Contains(zigVersion, "-dev.") {
		return "master"
	}

	// Convert input version to semantic version format
	version := zigVersion

	// Strip pre-release suffix (after '-')
	if idx := strings.Index(version, "-"); idx != -1 {
		version = version[:idx]
	}

	// Strip build metadata (after '+')
	if idx := strings.Index(version, "+"); idx != -1 {
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
