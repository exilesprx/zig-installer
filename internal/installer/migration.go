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

// MigrationChoice represents user's choice when system installation is detected
type MigrationChoice string

const (
	MigrationChoiceMigrate  MigrationChoice = "migrate"
	MigrationChoiceKeepBoth MigrationChoice = "keep_both"
	MigrationChoiceCancel   MigrationChoice = "cancel"
)

// SystemVersion represents a Zig installation found in system directories
type SystemVersion struct {
	DirName    string // e.g., "zig-linux-x86_64-0.13.0"
	Version    string // e.g., "0.13.0" (extracted)
	SourcePath string // e.g., "/opt/zig/zig-linux-x86_64-0.13.0"
	Size       int64  // Directory size for verification
}

// DiscoverSystemVersions scans systemDir for all zig-* subdirectories
func DiscoverSystemVersions(systemDir string) ([]SystemVersion, error) {
	entries, err := os.ReadDir(systemDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read system directory: %w", err)
	}

	var versions []SystemVersion

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Only consider directories that start with "zig-"
		if !strings.HasPrefix(entry.Name(), "zig-") {
			continue
		}

		path := filepath.Join(systemDir, entry.Name())
		version := extractVersionFromPath(path)

		if version == "" {
			continue
		}

		// Calculate directory size
		size, err := CalculateDirectorySize(path)
		if err != nil {
			size = 0 // Continue even if we can't get size
		}

		versions = append(versions, SystemVersion{
			DirName:    entry.Name(),
			Version:    version,
			SourcePath: path,
			Size:       size,
		})
	}

	return versions, nil
}

// DetectActiveZigVersion checks system bin directories for zig symlinks
func DetectActiveZigVersion(binDirs []string) (string, error) {
	for _, binDir := range binDirs {
		linkPath := filepath.Join(binDir, "zig")

		// Check if symlink exists
		if _, err := os.Lstat(linkPath); err != nil {
			continue // Try next directory
		}

		// Read symlink target
		target, err := os.Readlink(linkPath)
		if err != nil {
			continue
		}

		// Check if target points to system path
		if strings.HasPrefix(target, "/opt/") || strings.HasPrefix(target, "/usr/local/") {
			// Extract version from target path
			dir := filepath.Dir(target)
			version := extractVersionFromPath(dir)
			if version != "" {
				return version, nil
			}
		}
	}

	return "", nil // No active system version found
}

// DetectActiveZLSVersion checks if ZLS symlink exists in system bin directories
func DetectActiveZLSVersion(binDirs []string) (bool, error) {
	for _, binDir := range binDirs {
		linkPath := filepath.Join(binDir, "zls")

		// Check if symlink exists
		if _, err := os.Lstat(linkPath); err != nil {
			continue
		}

		// Read symlink target
		target, err := os.Readlink(linkPath)
		if err != nil {
			continue
		}

		// Check if target points to system path
		if strings.HasPrefix(target, "/opt/") || strings.HasPrefix(target, "/usr/local/") {
			return true, nil
		}
	}

	return false, nil
}

// MoveZigVersion moves a single Zig version directory using copy-verify-delete pattern
func MoveZigVersion(source, destParent, dirName string, formatter OutputFormatter, dryRun bool) error {
	destPath := filepath.Join(destParent, dirName)

	// Check if destination already exists
	if _, err := os.Stat(destPath); err == nil {
		formatter.PrintWarning("Skipping", fmt.Sprintf("%s (already exists in user-local)", dirName))
		return nil // Not an error - skip means prefer existing
	}

	if dryRun {
		sourceSize, _ := CalculateDirectorySize(source)
		formatter.PrintTask("Dry run", "Would move",
			fmt.Sprintf("%s (%s) -> %s", dirName, FormatBytes(sourceSize), destParent))
		return nil
	}

	// Calculate source size before copy
	sourceSize, err := CalculateDirectorySize(source)
	if err != nil {
		return fmt.Errorf("failed to calculate source size: %w", err)
	}

	formatter.PrintProgress("Copying", fmt.Sprintf("%s (%s)", dirName, FormatBytes(sourceSize)))

	// Execute: sudo cp -r <source> <destPath>
	cmd := exec.Command("sudo", "cp", "-r", source, destPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to copy: %w\nOutput: %s", err, string(output))
	}

	// Verify copy
	destSize, err := CalculateDirectorySize(destPath)
	if err != nil {
		return fmt.Errorf("failed to verify copy: %w", err)
	}

	// Compare sizes (allow small discrepancy)
	tolerance := int64(1024 * 1024) // 1MB tolerance
	if absInt64(sourceSize-destSize) > tolerance {
		return fmt.Errorf("size mismatch: source=%s, dest=%s",
			FormatBytes(sourceSize), FormatBytes(destSize))
	}

	// Check that key file exists
	zigBinary := filepath.Join(destPath, "zig")
	if _, err := os.Stat(zigBinary); err != nil {
		return fmt.Errorf("zig binary not found after copy: %w", err)
	}

	formatter.PrintSuccess("Copied", fmt.Sprintf("%s -> %s", dirName, destParent))

	// Fix ownership
	formatter.PrintProgress("Fixing ownership", dirName)
	cmd = exec.Command("sudo", "chown", "-R", fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()), destPath)
	if _, err := cmd.CombinedOutput(); err != nil {
		formatter.PrintWarning("Ownership", fmt.Sprintf("Failed to fix ownership: %v", err))
		// Not a fatal error, continue
	} else {
		formatter.PrintSuccess("Ownership", "Fixed")
	}

	// Delete source
	formatter.PrintProgress("Removing original", source)
	cmd = exec.Command("sudo", "rm", "-rf", source)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove source: %w\nOutput: %s", err, string(output))
	}

	// Verify source is gone
	if _, err := os.Stat(source); err == nil {
		return fmt.Errorf("source still exists after removal")
	}

	formatter.PrintSuccess("Removed", source)

	return nil
}

// MoveZLSInstallation moves ZLS directory using copy-verify-delete pattern
func MoveZLSInstallation(source, dest string, formatter OutputFormatter, dryRun bool) error {
	// Check if destination already exists
	if _, err := os.Stat(dest); err == nil {
		formatter.PrintWarning("Skipping ZLS", "ZLS already exists in user-local directory")
		return nil // Not an error - prefer existing installation
	}

	if dryRun {
		sourceSize, _ := CalculateDirectorySize(source)
		formatter.PrintTask("Dry run", "Would move",
			fmt.Sprintf("ZLS (%s) -> %s", FormatBytes(sourceSize), dest))
		return nil
	}

	// Calculate source size
	sourceSize, err := CalculateDirectorySize(source)
	if err != nil {
		return fmt.Errorf("failed to calculate source size: %w", err)
	}

	formatter.PrintProgress("Copying", fmt.Sprintf("ZLS (%s)", FormatBytes(sourceSize)))

	// Execute: sudo cp -r <source> <dest>
	cmd := exec.Command("sudo", "cp", "-r", source, dest)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to copy ZLS: %w\nOutput: %s", err, string(output))
	}

	// Verify copy - check .git directory exists
	gitDir := filepath.Join(dest, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return fmt.Errorf(".git directory not found after copy (ZLS should be a git repo): %w", err)
	}

	// Compare sizes
	destSize, err := CalculateDirectorySize(dest)
	if err != nil {
		return fmt.Errorf("failed to verify copy: %w", err)
	}

	tolerance := int64(1024 * 1024) // 1MB tolerance
	if absInt64(sourceSize-destSize) > tolerance {
		return fmt.Errorf("size mismatch: source=%s, dest=%s",
			FormatBytes(sourceSize), FormatBytes(destSize))
	}

	formatter.PrintSuccess("Copied", fmt.Sprintf("ZLS -> %s", dest))

	// Fix ownership
	formatter.PrintProgress("Fixing ownership", "ZLS")
	cmd = exec.Command("sudo", "chown", "-R", fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()), dest)
	if _, err := cmd.CombinedOutput(); err != nil {
		formatter.PrintWarning("Ownership", fmt.Sprintf("Failed to fix ownership: %v", err))
		// Not a fatal error, continue
	} else {
		formatter.PrintSuccess("Ownership", "Fixed")
	}

	// Delete source
	formatter.PrintProgress("Removing original", source)
	cmd = exec.Command("sudo", "rm", "-rf", source)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove source: %w\nOutput: %s", err, string(output))
	}

	// Verify source is gone
	if _, err := os.Stat(source); err == nil {
		return fmt.Errorf("source still exists after removal")
	}

	formatter.PrintSuccess("Removed", source)

	return nil
}

// RecreateZigSymlink creates symlink in user-local bin directory
func RecreateZigSymlink(binDir, zigDir, version string, formatter OutputFormatter, dryRun bool) error {
	// Find the directory matching this version
	versions, err := ScanInstalledVersions(zigDir, binDir)
	if err != nil {
		return fmt.Errorf("failed to scan versions: %w", err)
	}

	var targetVersionInfo *VersionInfo
	for i := range versions {
		if versions[i].Version == version {
			targetVersionInfo = &versions[i]
			break
		}
	}

	if targetVersionInfo == nil {
		return fmt.Errorf("version %s not found in %s", version, zigDir)
	}

	// Build symlink target
	zigBinPath := filepath.Join(targetVersionInfo.Path, "zig")
	linkPath := filepath.Join(binDir, "zig")

	// Verify target exists
	if _, err := os.Stat(zigBinPath); err != nil {
		return fmt.Errorf("zig binary not found at %s: %w", zigBinPath, err)
	}

	if dryRun {
		formatter.PrintTask("Dry run", "Would create symlink",
			fmt.Sprintf("%s -> %s", linkPath, zigBinPath))
		return nil
	}

	formatter.PrintProgress("Creating symlink", fmt.Sprintf("zig -> %s", version))

	// Remove existing symlink if present
	if _, err := os.Lstat(linkPath); err == nil {
		if err := os.Remove(linkPath); err != nil {
			return fmt.Errorf("could not remove existing symlink: %w", err)
		}
	}

	// Create new symlink
	if err := os.Symlink(zigBinPath, linkPath); err != nil {
		return fmt.Errorf("could not create symbolic link: %w", err)
	}

	formatter.PrintSuccess("Symlink created", fmt.Sprintf("%s -> %s", linkPath, zigBinPath))

	// Verify symlink works
	if err := VerifySwitch(binDir, version, formatter, false); err != nil {
		return fmt.Errorf("symlink verification failed: %w", err)
	}

	return nil
}

// RecreateZLSSymlink creates symlink for ZLS in user-local bin directory
func RecreateZLSSymlink(binDir, zlsDir string, formatter OutputFormatter, dryRun bool) error {
	// Build symlink target
	zlsBinPath := filepath.Join(zlsDir, "zig-out", "bin", "zls")
	linkPath := filepath.Join(binDir, "zls")

	// Verify target exists
	if _, err := os.Stat(zlsBinPath); err != nil {
		return fmt.Errorf("zls binary not found at %s: %w", zlsBinPath, err)
	}

	if dryRun {
		formatter.PrintTask("Dry run", "Would create symlink",
			fmt.Sprintf("%s -> %s", linkPath, zlsBinPath))
		return nil
	}

	formatter.PrintProgress("Creating symlink", "zls")

	// Remove existing symlink if present
	if _, err := os.Lstat(linkPath); err == nil {
		if err := os.Remove(linkPath); err != nil {
			return fmt.Errorf("could not remove existing symlink: %w", err)
		}
	}

	// Create new symlink
	if err := os.Symlink(zlsBinPath, linkPath); err != nil {
		return fmt.Errorf("could not create symbolic link: %w", err)
	}

	formatter.PrintSuccess("Symlink created", fmt.Sprintf("%s -> %s", linkPath, zlsBinPath))

	return nil
}

// CleanupSystemSymlinks removes system symlinks
func CleanupSystemSymlinks(formatter OutputFormatter, dryRun bool) error {
	systemBinLinks := []string{"/usr/local/bin/zig", "/usr/local/bin/zls", "/usr/bin/zig", "/usr/bin/zls"}

	removedCount := 0
	for _, link := range systemBinLinks {
		// Check if exists
		if _, err := os.Lstat(link); err != nil {
			continue // Doesn't exist, skip
		}

		if dryRun {
			formatter.PrintTask("Dry run", "Would remove", link)
			removedCount++
			continue
		}

		// Remove with sudo
		cmd := exec.Command("sudo", "rm", "-f", link)
		_, err := cmd.CombinedOutput()
		if err != nil {
			formatter.PrintWarning("Cleanup", fmt.Sprintf("Failed to remove %s: %v", link, err))
			continue
		}

		// Verify removal
		if _, err := os.Lstat(link); err == nil {
			formatter.PrintWarning("Cleanup", fmt.Sprintf("Still exists: %s", link))
			continue
		}

		formatter.PrintSuccess("Removed", link)
		removedCount++
	}

	if removedCount > 0 {
		formatter.PrintSuccess("Cleanup", fmt.Sprintf("Removed %d system symlink(s)", removedCount))
	}

	return nil
}

// absInt64 returns the absolute value of an int64
func absInt64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

// DetectAndPromptMigration checks for system installation and prompts user
func DetectAndPromptMigration(formatter OutputFormatter, logger logger.ILogger) (MigrationChoice, string, error) {
	systemDir, found := config.DetectSystemInstallation()
	if !found {
		return "", "", nil
	}

	formatter.PrintWarning("System Installation Detected",
		fmt.Sprintf("Found existing system-wide Zig installation at: %s", systemDir))

	formatter.PrintTask("Migration Notice", "Important",
		"This installer now uses user-local installation (~/.local)")

	var choice string
	prompt := &survey.Select{
		Message: "What would you like to do?",
		Options: []string{
			"Migrate to user-local (recommended - removes system installation)",
			"Keep both (may cause PATH conflicts)",
			"Cancel installation",
		},
		Default: "Migrate to user-local (recommended - removes system installation)",
	}

	if err := survey.AskOne(prompt, &choice); err != nil {
		return MigrationChoiceCancel, "", err
	}

	switch {
	case strings.Contains(choice, "Migrate"):
		return MigrationChoiceMigrate, systemDir, nil
	case strings.Contains(choice, "Keep both"):
		return MigrationChoiceKeepBoth, systemDir, nil
	default:
		return MigrationChoiceCancel, systemDir, nil
	}
}

// PerformMigration migrates system installation to user-local using copy-verify-delete pattern
func PerformMigration(systemDir string, formatter OutputFormatter, logger logger.ILogger, dryRun bool) error {
	if dryRun {
		formatter.PrintProgress("Migration", "Starting migration preview (dry run)")
	} else {
		formatter.PrintProgress("Migration", "Starting migration to user-local installation")
	}

	// Get user-local paths
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	userZigDir := filepath.Join(home, ".local", "share", "zig")
	userZLSDir := filepath.Join(home, ".local", "share", "zls")
	userBinDir := filepath.Join(home, ".local", "bin")

	// Ensure user-local directories exist (except in dry-run mode)
	if !dryRun {
		if err := os.MkdirAll(userZigDir, 0755); err != nil {
			return fmt.Errorf("could not create %s: %w", userZigDir, err)
		}
		if err := os.MkdirAll(userBinDir, 0755); err != nil {
			return fmt.Errorf("could not create %s: %w", userBinDir, err)
		}
	}

	// === Step 1: Discover what needs to be migrated ===
	formatter.PrintSection("Discovering installations")

	// Discover Zig versions
	zigVersions, err := DiscoverSystemVersions(systemDir)
	if err != nil {
		return fmt.Errorf("could not discover Zig versions: %w", err)
	}
	formatter.PrintSuccess("Discovery", fmt.Sprintf("Found %d Zig version(s)", len(zigVersions)))

	// Log discovered versions
	for _, v := range zigVersions {
		logger.LogInfo("  - %s (%s)", v.Version, FormatBytes(v.Size))
	}

	// Detect active version
	activeVersion, err := DetectActiveZigVersion([]string{"/usr/local/bin", "/usr/bin"})
	if err != nil {
		logger.LogInfo("Could not detect active Zig version: %v", err)
	}
	if activeVersion != "" {
		formatter.PrintTask("Active version", "Detected", activeVersion)
	}

	// Check for ZLS
	hasZLS := false
	systemZLSDirs := []string{"/opt/zls", "/usr/local/zls"}
	var systemZLSDir string
	for _, dir := range systemZLSDirs {
		if _, err := os.Stat(dir); err == nil {
			hasZLS = true
			systemZLSDir = dir
			break
		}
	}
	if hasZLS {
		zlsSize, _ := CalculateDirectorySize(systemZLSDir)
		formatter.PrintSuccess("Discovery", fmt.Sprintf("Found ZLS installation: %s (%s)", systemZLSDir, FormatBytes(zlsSize)))
	}

	// === Step 2: Migrate Zig versions ===
	if len(zigVersions) > 0 {
		if dryRun {
			formatter.PrintSection(fmt.Sprintf("Would migrate %d Zig version(s) [DRY RUN]", len(zigVersions)))
		} else {
			formatter.PrintSection(fmt.Sprintf("Migrating %d Zig version(s)", len(zigVersions)))
		}

		movedCount := 0
		skippedCount := 0

		for _, ver := range zigVersions {
			destPath := filepath.Join(userZigDir, ver.DirName)

			// Check if destination exists
			if _, err := os.Stat(destPath); err == nil {
				formatter.PrintWarning("Skipping",
					fmt.Sprintf("%s (already exists in user-local)", ver.Version))
				logger.LogInfo("Skipped %s - destination exists", ver.Version)
				skippedCount++
				continue
			}

			// Move this version
			if !dryRun {
				formatter.PrintProgress("Migrating", ver.Version)
			}
			if err := MoveZigVersion(ver.SourcePath, userZigDir, ver.DirName, formatter, dryRun); err != nil {
				return fmt.Errorf("failed to migrate %s: %w", ver.Version, err)
			}
			if !dryRun {
				formatter.PrintSuccess("Migrated", ver.Version)
			}
			movedCount++
		}

		if dryRun {
			formatter.PrintSuccess("Preview",
				fmt.Sprintf("Would move %d version(s), skip %d", movedCount, skippedCount))
		} else {
			formatter.PrintSuccess("Zig migration",
				fmt.Sprintf("Moved %d version(s), skipped %d", movedCount, skippedCount))
		}
	}

	// === Step 3: Migrate ZLS ===
	zlsMigrated := false
	if hasZLS {
		if dryRun {
			formatter.PrintSection("Would migrate ZLS [DRY RUN]")
		} else {
			formatter.PrintSection("Migrating ZLS")
		}

		// Check if destination exists
		if _, err := os.Stat(userZLSDir); err == nil {
			formatter.PrintWarning("Skipping ZLS",
				"ZLS already exists in user-local directory")
			logger.LogInfo("Skipped ZLS - destination exists")
		} else {
			if err := MoveZLSInstallation(systemZLSDir, userZLSDir, formatter, dryRun); err != nil {
				// Don't fail entire migration if just ZLS fails
				formatter.PrintError("ZLS migration", fmt.Sprintf("%v", err))
				logger.LogInfo("ZLS migration failed: %v", err)
			} else {
				if dryRun {
					formatter.PrintSuccess("Preview", "Would migrate ZLS successfully")
				} else {
					formatter.PrintSuccess("ZLS migration", "Successfully migrated")
					zlsMigrated = true
				}
			}
		}
	}

	// === Step 4: Recreate symlinks ===
	if dryRun {
		formatter.PrintSection("Would recreate symlinks [DRY RUN]")
	} else {
		formatter.PrintSection("Recreating symlinks")
	}

	// Recreate Zig symlink
	if activeVersion != "" && len(zigVersions) > 0 {
		if dryRun {
			formatter.PrintTask("Preview", "Would create symlink", fmt.Sprintf("zig -> version %s", activeVersion))
		} else {
			formatter.PrintProgress("Zig symlink", fmt.Sprintf("Creating symlink for version %s", activeVersion))
		}

		if err := RecreateZigSymlink(userBinDir, userZigDir, activeVersion, formatter, dryRun); err != nil {
			// If active version fails, try the first available version
			logger.LogInfo("Could not recreate symlink for %s: %v", activeVersion, err)
			formatter.PrintWarning("Active version unavailable",
				"Will use first available version")

			if len(zigVersions) > 0 {
				fallbackVersion := zigVersions[0].Version
				if err := RecreateZigSymlink(userBinDir, userZigDir, fallbackVersion, formatter, dryRun); err != nil {
					return fmt.Errorf("could not create Zig symlink: %w", err)
				}
				if !dryRun {
					formatter.PrintSuccess("Zig symlink", fmt.Sprintf("Created for version %s", fallbackVersion))
				}
			}
		} else {
			if !dryRun {
				formatter.PrintSuccess("Zig symlink", fmt.Sprintf("Created for version %s", activeVersion))
			}
		}
	} else if len(zigVersions) > 0 {
		// No active version detected, use first available
		firstVersion := zigVersions[0].Version
		if dryRun {
			formatter.PrintTask("Preview", "Would create symlink", fmt.Sprintf("zig -> version %s (first available)", firstVersion))
		} else {
			formatter.PrintProgress("Zig symlink", fmt.Sprintf("Creating symlink for version %s", firstVersion))
		}

		if err := RecreateZigSymlink(userBinDir, userZigDir, firstVersion, formatter, dryRun); err != nil {
			return fmt.Errorf("could not create Zig symlink: %w", err)
		}
		if !dryRun {
			formatter.PrintSuccess("Zig symlink", fmt.Sprintf("Created for version %s", firstVersion))
		}
	}

	// Recreate ZLS symlink
	if zlsMigrated || (dryRun && hasZLS) {
		if dryRun {
			formatter.PrintTask("Preview", "Would create symlink", "zls -> ~/.local/share/zls/zig-out/bin/zls")
		} else {
			formatter.PrintProgress("ZLS symlink", "Creating symlink")
		}

		if err := RecreateZLSSymlink(userBinDir, userZLSDir, formatter, dryRun); err != nil {
			// Don't fail entire migration if just ZLS symlink fails
			formatter.PrintWarning("ZLS symlink", fmt.Sprintf("Failed: %v", err))
			logger.LogInfo("ZLS symlink creation failed: %v", err)
		} else {
			if !dryRun {
				formatter.PrintSuccess("ZLS symlink", "Created successfully")
			}
		}
	}

	// === Step 5: Cleanup system symlinks ===
	if dryRun {
		formatter.PrintSection("Would clean up system symlinks [DRY RUN]")
	} else {
		formatter.PrintSection("Cleaning up system symlinks")
	}

	if err := CleanupSystemSymlinks(formatter, dryRun); err != nil {
		// Don't fail migration if cleanup fails - log warning
		formatter.PrintWarning("Cleanup", fmt.Sprintf("Could not remove all system symlinks: %v", err))
		logger.LogInfo("System symlink cleanup incomplete: %v", err)
	} else {
		if !dryRun {
			formatter.PrintSuccess("Cleanup", "Removed system symlinks")
		}
	}

	// === Step 6: Remove empty system directories ===
	if !dryRun {
		formatter.PrintProgress("Final cleanup", "Removing empty system directories")

		// Try to remove system Zig dir (will only succeed if empty)
		if err := exec.Command("sudo", "rmdir", systemDir).Run(); err != nil {
			logger.LogInfo("Could not remove %s (may not be empty): %v", systemDir, err)
		} else {
			formatter.PrintSuccess("Cleanup", fmt.Sprintf("Removed %s", systemDir))
		}

		// Try to remove system ZLS dir (will only succeed if empty)
		if hasZLS {
			if err := exec.Command("sudo", "rmdir", systemZLSDir).Run(); err != nil {
				logger.LogInfo("Could not remove %s (may not be empty): %v", systemZLSDir, err)
			} else {
				formatter.PrintSuccess("Cleanup", fmt.Sprintf("Removed %s", systemZLSDir))
			}
		}
	} else {
		formatter.PrintTask("Preview", "Would remove empty directories", fmt.Sprintf("%s, %s", systemDir, systemZLSDir))
	}

	return nil
}

// WarnAboutPathConflict warns user about keeping both installations
func WarnAboutPathConflict(systemDir string, userLocalBin string, formatter OutputFormatter) {
	formatter.PrintWarning("PATH Conflict Warning",
		"You have both system and user-local installations")

	formatter.PrintTask("Recommendation", "Update PATH",
		"Ensure your user-local installation takes precedence")

	formatter.PrintTask("PATH Configuration", "Add to shell config",
		fmt.Sprintf("Add this line BEFORE other PATH entries:\n"+
			"      export PATH=\"%s:$PATH\"\n\n"+
			"    System installation: %s\n"+
			"    User-local installation: %s",
			userLocalBin, systemDir, userLocalBin))
}
