package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/exilesprx/zig-installer/internal/config"
	"github.com/exilesprx/zig-installer/internal/logger"
)

// SwitchToVersion switches the active Zig version by updating the symlink
func SwitchToVersion(cfg *config.Config, log logger.ILogger, formatter OutputFormatter, targetVersion string) error {
	formatter.PrintSection("Switching Zig Version")

	// Scan installed versions
	formatter.PrintProgress("Scanning versions", "Checking installed Zig versions")
	versions, err := ScanInstalledVersions(cfg.ZigDir, cfg.BinDir)
	if err != nil {
		formatter.PrintError("Scanning versions", err.Error())
		return fmt.Errorf("failed to scan installed versions: %w", err)
	}

	if len(versions) == 0 {
		formatter.PrintError("No versions found", "No Zig versions installed")
		return fmt.Errorf("no Zig versions installed in %s", cfg.ZigDir)
	}

	if len(versions) == 1 {
		formatter.PrintWarning("Only one version", fmt.Sprintf("Only %s is installed", versions[0].Version))
		return fmt.Errorf("only one version installed - install another version first")
	}

	formatter.PrintSuccess("Scanning versions", fmt.Sprintf("Found %d installed version(s)", len(versions)))

	// Find the target version
	var targetVersionInfo *VersionInfo
	for i := range versions {
		if versions[i].Version == targetVersion {
			targetVersionInfo = &versions[i]
			break
		}
	}

	if targetVersionInfo == nil {
		formatter.PrintError("Version not found", fmt.Sprintf("Version %s not found", targetVersion))
		return fmt.Errorf("version %s is not installed", targetVersion)
	}

	// Check if already using this version
	currentVersion, _ := GetCurrentVersion(cfg.BinDir)
	if currentVersion == targetVersion {
		formatter.PrintWarning("Already active", fmt.Sprintf("Already using Zig %s", targetVersion))
		formatter.PrintTask("Action", "Recreating symlink", "Will recreate the symlink to ensure it's correct")
	}

	// Update symlink
	if err := UpdateZigSymlink(targetVersionInfo.Path, cfg.BinDir, targetVersion, formatter); err != nil {
		log.LogError("Failed to update symlink: %v", err)
		return err
	}

	// Verify the switch
	if err := VerifySwitch(cfg.BinDir, targetVersion, formatter, cfg.NoColor); err != nil {
		log.LogError("Failed to verify switch: %v", err)
		return err
	}

	formatter.PrintSuccess("Switch complete", fmt.Sprintf("Now using Zig %s", targetVersion))
	return nil
}

// PromptVersionSwitch prompts the user to select a version to switch to
func PromptVersionSwitch(versions []VersionInfo) (string, error) {
	if len(versions) == 0 {
		return "", fmt.Errorf("no versions available for selection")
	}

	if len(versions) == 1 {
		return "", fmt.Errorf("only one version installed - cannot switch")
	}

	// Build options list
	var options []string
	versionMap := make(map[string]string) // Display string -> actual version

	for _, v := range versions {
		display := v.Version
		if v.IsCurrent {
			display = fmt.Sprintf("→ %s (current)", v.Version)
		}
		options = append(options, display)
		versionMap[display] = v.Version
	}

	// Create single-select prompt
	var selected string
	prompt := &survey.Select{
		Message: "Select Zig version to switch to:",
		Options: options,
	}

	if err := survey.AskOne(prompt, &selected); err != nil {
		return "", err
	}

	// Return the actual version string
	return versionMap[selected], nil
}

// UpdateZigSymlink updates the zig symlink to point to the specified version
func UpdateZigSymlink(versionPath, binDir, version string, formatter OutputFormatter) error {
	formatter.PrintProgress("Updating symlink", fmt.Sprintf("Switching to version %s", version))

	// Find the zig binary in the version directory
	zigBinPath := filepath.Join(versionPath, "zig")

	// Verify the binary exists
	if _, err := os.Stat(zigBinPath); err != nil {
		formatter.PrintError("Binary not found", fmt.Sprintf("Zig binary not found at %s", zigBinPath))
		return fmt.Errorf("zig binary not found at %s: %w", zigBinPath, err)
	}

	linkPath := filepath.Join(binDir, "zig")

	// Remove existing symlink/file if it exists
	if _, err := os.Lstat(linkPath); err == nil {
		if err := os.Remove(linkPath); err != nil {
			formatter.PrintError("Removing old symlink", fmt.Sprintf("Failed to remove existing symlink: %v", err))
			return fmt.Errorf("could not remove existing symlink: %w", err)
		}
		formatter.PrintSuccess("Removing old symlink", fmt.Sprintf("Removed old symlink at %s", linkPath))
	}

	// Create new symlink
	if err := os.Symlink(zigBinPath, linkPath); err != nil {
		formatter.PrintError("Creating symlink", fmt.Sprintf("Failed to create symlink: %v", err))
		return fmt.Errorf("could not create symbolic link: %w", err)
	}

	formatter.PrintSuccess("Creating symlink", fmt.Sprintf("Created symlink: %s -> %s", linkPath, zigBinPath))
	return nil
}

// VerifySwitch verifies that the switch was successful by running zig version
func VerifySwitch(binDir, expectedVersion string, formatter OutputFormatter, noColor bool) error {
	formatter.PrintProgress("Verifying switch", "Running 'zig version' to confirm")

	zigPath := filepath.Join(binDir, "zig")

	// Check if zig binary/symlink exists
	if _, err := os.Lstat(zigPath); err != nil {
		formatter.PrintError("Verification failed", "Zig symlink not found")
		return fmt.Errorf("zig symlink not found at %s", zigPath)
	}

	// Run zig version
	cmd := exec.Command(zigPath, "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		formatter.PrintError("Verification failed", fmt.Sprintf("Failed to run 'zig version': %v", err))
		return fmt.Errorf("failed to run 'zig version': %w", err)
	}

	actualVersion := strings.TrimSpace(string(output))

	// Check if the version matches (allowing for some flexibility in format)
	if !strings.Contains(actualVersion, expectedVersion) {
		formatter.PrintError("Verification failed",
			fmt.Sprintf("Version mismatch - expected %s, got %s", expectedVersion, actualVersion))
		return fmt.Errorf("version mismatch: expected %s, got %s", expectedVersion, actualVersion)
	}

	formatter.PrintSuccess("Verification successful", fmt.Sprintf("Zig version: %s", actualVersion))
	return nil
}
