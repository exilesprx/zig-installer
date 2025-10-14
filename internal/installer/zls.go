package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/exilesprx/zig-install/internal/config"
	"github.com/exilesprx/zig-install/internal/logger"
)

// InstallZLS handles the ZLS installation process
func InstallZLS(p interface{}, config *config.Config, logger logger.ILogger, formatter OutputFormatter, zigVersion string) error {
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
		formatter.PrintTask("Directory setup", "In progress", fmt.Sprintf("Setting ownership of %s to %s", config.ZLSDir, user))
		cmd := exec.Command("chown", "-R", user+":"+user, config.ZLSDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			formatter.PrintTask("Directory setup", "Failed", fmt.Sprintf("Error setting ownership: %s", output))
			return fmt.Errorf("could not set ownership of %s: %w", config.ZLSDir, err)
		} else {
			formatter.PrintTask("Directory setup", "Success", "Directory ownership configured")
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
		output, err := cmd.Output()
		if err == nil && strings.Contains(string(output), "zigtools/zls") {
			isRepoCloned = true
			formatter.PrintTask("Repository check", "Success", "ZLS repository already exists")
		}
	}

	if !isRepoCloned {
		formatter.PrintTask("Repository clone", "In progress", "Cloning ZLS repository...")

		// Clone the repository
		cmd := exec.Command("git", "clone", "https://github.com/zigtools/zls.git", config.ZLSDir)
		output, err := cmd.CombinedOutput()
		if err != nil {
			formatter.PrintTask("Repository clone", "Failed", fmt.Sprintf("Error cloning repository: %s", output))
			return fmt.Errorf("could not clone ZLS repository: %w", err)
		}

		formatter.PrintTask("ZLS clone", "Success", fmt.Sprintf("Cloned repository to %s", config.ZLSDir))
	}

	// Handle version-specific setup
	if isSpecificVersion {
		formatter.PrintTask("Version setup", "In progress", fmt.Sprintf("Setting up ZLS version %s...", version))

		// Fetch all tags
		cmd := exec.Command("git", "fetch", "--tags")
		cmd.Dir = config.ZLSDir
		if output, err := cmd.CombinedOutput(); err != nil {
			formatter.PrintTask("Version setup", "Failed", fmt.Sprintf("Error fetching tags: %s", output))
			return fmt.Errorf("could not fetch tags: %w", err)
		}

		// Verify the version exists
		cmd = exec.Command("git", "tag", "-l", version)
		cmd.Dir = config.ZLSDir
		output, err := cmd.Output()
		if err != nil || len(strings.TrimSpace(string(output))) == 0 {
			formatter.PrintTask("Version setup", "Failed", fmt.Sprintf("Version %s not found", version))
			return fmt.Errorf("version %s not found in ZLS repository", version)
		}

		// Checkout the specific version
		cmd = exec.Command("git", "checkout", version)
		cmd.Dir = config.ZLSDir
		if output, err := cmd.CombinedOutput(); err != nil {
			formatter.PrintTask("Version setup", "Failed", fmt.Sprintf("Error checking out version: %s", output))
			return fmt.Errorf("could not checkout version %s: %w", version, err)
		}

		formatter.PrintTask("ZLS version", "Success", fmt.Sprintf("Checked out version %s", version))
	} else {
		formatter.PrintTask("Latest setup", "In progress", "Setting up latest ZLS...")

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
				formatter.PrintTask("Latest setup", "Failed", fmt.Sprintf("Error pulling latest changes: %s", output))
				return fmt.Errorf("could not pull latest changes: %w", err)
			}
		}

		formatter.PrintTask("ZLS latest", "Success", "Updated to latest version")
	}

	// Set ownership after git operations
	if user != "" {
		formatter.PrintTask("Ownership update", "In progress", fmt.Sprintf("Setting ownership after git operations for %s", config.ZLSDir))
		cmd := exec.Command("chown", "-R", user+":"+user, config.ZLSDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			formatter.PrintTask("Ownership update", "Failed", fmt.Sprintf("Error: %s", output))
			return fmt.Errorf("could not set ownership after git operations: %w", err)
		} else {
			formatter.PrintTask("Ownership update", "Success", "Repository ownership updated")
		}
	}

	// Build ZLS
	formatter.PrintTask("ZLS build", "In progress", "Building ZLS...")
	formatter.PrintTask("Build details", "Info", fmt.Sprintf("Running: zig build -Doptimize=ReleaseSafe in %s", config.ZLSDir))

	// Stream output in real-time if verbose mode is enabled
	if config.Verbose {
		cmd := exec.Command("zig", "build", "-Doptimize=ReleaseSafe", "--verbose")
		cmd.Dir = config.ZLSDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			formatter.PrintTask("ZLS build", "Failed", fmt.Sprintf("Error building ZLS: %v", err))
			return fmt.Errorf("could not build ZLS: %w", err)
		}
	} else {
		// Capture output for error reporting only
		cmd := exec.Command("zig", "build", "-Doptimize=ReleaseSafe")
		cmd.Dir = config.ZLSDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			formatter.PrintTask("ZLS build", "Failed", fmt.Sprintf("Error building ZLS: %s", output))
			return fmt.Errorf("could not build ZLS: %w", err)
		}
	}

	formatter.PrintTask("ZLS build", "Success", "ZLS built successfully")

	// Set ownership of the build output
	buildOutDir := filepath.Join(config.ZLSDir, "zig-out")
	if user != "" && isDirectory(buildOutDir) {
		formatter.PrintTask("Build ownership", "In progress", fmt.Sprintf("Setting ownership of build output in %s", buildOutDir))
		cmd := exec.Command("chown", "-R", user+":"+user, buildOutDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			formatter.PrintTask("Build ownership", "Failed", fmt.Sprintf("Error setting ownership: %s", output))
			return fmt.Errorf("could not set ownership of build output: %w", err)
		} else {
			formatter.PrintTask("Build ownership", "Success", "Build output ownership configured")
		}
	}

	// Create symbolic link to ZLS binary
	zlsBinPath := filepath.Join(config.ZLSDir, "zig-out", "bin", "zls")
	linkPath := filepath.Join(config.BinDir, "zls")

	formatter.PrintTask("ZLS symlink", "In progress", fmt.Sprintf("Creating symlink from %s to %s", zlsBinPath, linkPath))

	// Remove existing symlink/file if it exists
	if _, err := os.Lstat(linkPath); err == nil {
		if err := os.Remove(linkPath); err != nil {
			formatter.PrintTask("ZLS symlink", "Failed", fmt.Sprintf("Failed to remove existing symlink: %v", err))
			return fmt.Errorf("could not remove existing symlink: %w", err)
		}
	}

	if err := os.Symlink(zlsBinPath, linkPath); err != nil {
		return fmt.Errorf("could not create symbolic link: %w", err)
	}

	formatter.PrintTask("ZLS symbolic link", "Success", fmt.Sprintf("Created symlink: %s -> %s", linkPath, zlsBinPath))

	return nil
}

// isDirectory checks if the given path is a directory
func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
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
