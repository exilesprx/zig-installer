package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/exilesprx/zig-install/internal/config"
	"github.com/exilesprx/zig-install/internal/logger"
	"github.com/pterm/pterm"
)

// VersionInfo represents an installed Zig version
type VersionInfo struct {
	Version     string    // e.g., "0.13.0", "0.12.0-dev.123"
	Path        string    // Full path to installation directory
	Size        int64     // Size in bytes
	InstallDate time.Time // Extracted from directory mtime
	IsCurrent   bool      // Whether this is the currently symlinked version
}

// extractVersionFromPath extracts the version string from a directory path
// Example: "/opt/zig/zig-linux-x86_64-0.13.0" -> "0.13.0"
func extractVersionFromPath(path string) string {
	base := filepath.Base(path)

	// Pattern: zig-{os}-{arch}-{version}
	// We want everything after the last hyphen after "zig-"
	parts := strings.Split(base, "-")
	if len(parts) < 4 {
		return ""
	}

	// Join everything after the third hyphen (zig-os-arch-version...)
	version := strings.Join(parts[3:], "-")
	return version
}

// CalculateDirectorySize recursively calculates the total size of a directory
func CalculateDirectorySize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// FormatBytes converts bytes to human-readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.0f %s", float64(bytes)/float64(div), units[exp])
}

// GetCurrentVersion reads the symlink to determine the currently active version
func GetCurrentVersion(binDir string) (string, error) {
	linkPath := filepath.Join(binDir, "zig")

	target, err := os.Readlink(linkPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // No symlink exists
		}
		return "", err
	}

	// Extract version from the symlink target
	// e.g., "/opt/zig/zig-linux-x86_64-0.13.0/zig" -> "0.13.0"
	dir := filepath.Dir(target)
	return extractVersionFromPath(dir), nil
}

// ScanInstalledVersions scans the zig directory for installed versions
func ScanInstalledVersions(zigDir, binDir string) ([]VersionInfo, error) {
	entries, err := os.ReadDir(zigDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read zig directory: %w", err)
	}

	currentVersion, err := GetCurrentVersion(binDir)
	if err != nil {
		// Log but don't fail - we can still show versions
		currentVersion = ""
	}

	var versions []VersionInfo

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Only consider directories that start with "zig-"
		if !strings.HasPrefix(entry.Name(), "zig-") {
			continue
		}

		path := filepath.Join(zigDir, entry.Name())
		version := extractVersionFromPath(path)

		if version == "" {
			continue
		}

		// Get directory info for modification time
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Calculate directory size
		size, err := CalculateDirectorySize(path)
		if err != nil {
			size = 0 // Continue even if we can't get size
		}

		versions = append(versions, VersionInfo{
			Version:     version,
			Path:        path,
			Size:        size,
			InstallDate: info.ModTime(),
			IsCurrent:   version == currentVersion,
		})
	}

	// Sort by install date (newest first)
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].InstallDate.After(versions[j].InstallDate)
	})

	return versions, nil
}

// DisplayVersionsTable displays a table of installed versions using pterm
func DisplayVersionsTable(versions []VersionInfo, noColor bool) error {
	if len(versions) == 0 {
		return fmt.Errorf("no versions found")
	}

	// Build table data
	tableData := pterm.TableData{
		{"Version", "Size", "Install Date", "Current"},
	}

	var totalSize int64
	for _, v := range versions {
		totalSize += v.Size
		current := ""
		if v.IsCurrent {
			current = "✓"
		}

		tableData = append(tableData, []string{
			v.Version,
			FormatBytes(v.Size),
			v.InstallDate.Format("2006-01-02"),
			current,
		})
	}

	// Print table
	if noColor {
		pterm.DisableColor()
		defer func() { pterm.EnableColor() }()
	}

	if err := pterm.DefaultTable.WithHasHeader().WithData(tableData).Render(); err != nil {
		return err
	}

	// Print total
	pterm.Println()
	pterm.Printf("Total disk usage: %s\n", FormatBytes(totalSize))
	pterm.Println()

	return nil
}

// filterVersionsToKeep filters versions based on keep-last parameter
// Returns versions that should be REMOVED
func filterVersionsToKeep(versions []VersionInfo, keepLast int) []VersionInfo {
	if keepLast <= 0 {
		return nil
	}

	// Sort by install date (newest first) if not already sorted
	sorted := make([]VersionInfo, len(versions))
	copy(sorted, versions)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].InstallDate.After(sorted[j].InstallDate)
	})

	var toRemove []VersionInfo
	kept := 0

	for _, v := range sorted {
		// Always keep current version
		if v.IsCurrent {
			continue
		}

		if kept < keepLast {
			kept++
		} else {
			toRemove = append(toRemove, v)
		}
	}

	return toRemove
}

// PromptVersionSelection prompts the user to select versions to remove
func PromptVersionSelection(versions []VersionInfo) ([]string, error) {
	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions available for selection")
	}

	// Build options list
	var options []string
	disabledOptions := make(map[string]bool)

	for _, v := range versions {
		option := v.Version
		if v.IsCurrent {
			option += " (current version - cannot be removed)"
			disabledOptions[option] = true
		}
		options = append(options, option)
	}

	// Create multi-select prompt
	var selected []string
	prompt := &survey.MultiSelect{
		Message: "Select versions to remove (space to select, enter to confirm):",
		Options: options,
	}

	if err := survey.AskOne(prompt, &selected); err != nil {
		return nil, err
	}

	// Clean up selected versions (remove the " (current...)" suffix if present)
	var cleaned []string
	for _, s := range selected {
		// Extract just the version number
		parts := strings.Split(s, " ")
		cleaned = append(cleaned, parts[0])
	}

	return cleaned, nil
}

// ConfirmRemoval asks the user to confirm the removal
func ConfirmRemoval(versions []string, totalSize int64) (bool, error) {
	message := fmt.Sprintf("Remove %d version(s) and free %s?", len(versions), FormatBytes(totalSize))

	var confirmed bool
	prompt := &survey.Confirm{
		Message: message,
		Default: true,
	}

	if err := survey.AskOne(prompt, &confirmed); err != nil {
		return false, err
	}

	return confirmed, nil
}

// RemoveVersions removes the specified versions
func RemoveVersions(zigDir string, versions []string, formatter OutputFormatter) error {
	formatter.PrintSection("Removing versions")

	for _, version := range versions {
		// Find the directory matching this version
		entries, err := os.ReadDir(zigDir)
		if err != nil {
			return fmt.Errorf("failed to read zig directory: %w", err)
		}

		var dirToRemove string
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			path := filepath.Join(zigDir, entry.Name())
			v := extractVersionFromPath(path)

			if v == version {
				dirToRemove = path
				break
			}
		}

		if dirToRemove == "" {
			formatter.PrintError("Version removal", fmt.Sprintf("Could not find directory for version %s", version))
			continue
		}

		// Get size before removal for reporting
		size, _ := CalculateDirectorySize(dirToRemove)

		// Remove the directory
		formatter.PrintProgress("Version removal", fmt.Sprintf("Removing %s", version))
		if err := os.RemoveAll(dirToRemove); err != nil {
			formatter.PrintError("Version removal", fmt.Sprintf("Failed to remove %s: %v", version, err))
			return fmt.Errorf("failed to remove %s: %w", version, err)
		}

		formatter.PrintSuccess("Version removal", fmt.Sprintf("Removed %s (%s)", version, FormatBytes(size)))
	}

	return nil
}

// AutoCleanupPrompt prompts for cleanup after installation
func AutoCleanupPrompt(cfg *config.Config, log logger.ILogger, formatter OutputFormatter, currentVersion string) error {
	// Scan for installed versions
	versions, err := ScanInstalledVersions(cfg.ZigDir, cfg.BinDir)
	if err != nil {
		return fmt.Errorf("failed to scan versions: %w", err)
	}

	// Filter out the current version for counting other versions
	var otherVersions []VersionInfo
	var totalSize int64
	for _, v := range versions {
		if !v.IsCurrent {
			otherVersions = append(otherVersions, v)
			totalSize += v.Size
		}
	}

	// If no other versions, nothing to clean up
	if len(otherVersions) == 0 {
		log.LogInfo("No other versions found for cleanup")
		return nil
	}

	pterm.Println()
	pterm.Info.Printf("Found %d other installed version(s) (%s)\n", len(otherVersions), FormatBytes(totalSize))

	// If auto-cleanup with keep-last is set, handle automatically
	if cfg.AutoCleanup && cfg.KeepLast > 0 {
		return autoCleanupWithKeepLast(cfg, log, formatter, versions)
	}

	// If auto-cleanup without keep-last, go straight to interactive
	if cfg.AutoCleanup {
		return interactiveCleanup(cfg, log, formatter, versions)
	}

	// Default: Prompt user if they want to clean up
	var wantsCleanup bool
	prompt := &survey.Confirm{
		Message: "Clean up old versions?",
		Default: true, // Default is Yes
	}

	if err := survey.AskOne(prompt, &wantsCleanup); err != nil {
		return err
	}

	if !wantsCleanup {
		log.LogInfo("User declined cleanup")
		return nil
	}

	// User said yes, proceed with interactive cleanup
	return interactiveCleanup(cfg, log, formatter, versions)
}

// autoCleanupWithKeepLast handles automatic cleanup with keep-last parameter
func autoCleanupWithKeepLast(cfg *config.Config, log logger.ILogger, formatter OutputFormatter, versions []VersionInfo) error {
	pterm.Println()
	formatter.PrintSection(fmt.Sprintf("Auto-cleanup (keeping last %d versions)", cfg.KeepLast))

	// Filter versions to remove
	toRemove := filterVersionsToKeep(versions, cfg.KeepLast)

	if len(toRemove) == 0 {
		formatter.PrintSuccess("Auto-cleanup", "No versions to remove")
		return nil
	}

	// Calculate total size
	var totalSize int64
	var versionNames []string
	for _, v := range toRemove {
		totalSize += v.Size
		versionNames = append(versionNames, v.Version)
	}

	formatter.PrintProgress("Auto-cleanup", fmt.Sprintf("Found %d version(s) to remove: %s (%s)",
		len(toRemove), strings.Join(versionNames, ", "), FormatBytes(totalSize)))

	// Remove versions
	if err := RemoveVersions(cfg.ZigDir, versionNames, formatter); err != nil {
		return err
	}

	log.LogInfo("Auto-cleanup completed: removed %d versions, freed %s", len(versionNames), FormatBytes(totalSize))

	pterm.Println()
	pterm.Success.Printf("Freed %s of disk space\n", FormatBytes(totalSize))

	return nil
}

// interactiveCleanup handles interactive version selection and removal
func interactiveCleanup(cfg *config.Config, log logger.ILogger, formatter OutputFormatter, versions []VersionInfo) error {
	pterm.Println()

	// Display versions table
	if err := DisplayVersionsTable(versions, cfg.NoColor); err != nil {
		return err
	}

	// Prompt for selection
	selected, err := PromptVersionSelection(versions)
	if err != nil {
		return fmt.Errorf("failed to get selection: %w", err)
	}

	if len(selected) == 0 {
		log.LogInfo("No versions selected for removal")
		pterm.Info.Println("No versions selected")
		return nil
	}

	// Calculate total size of selected versions
	var totalSize int64
	for _, v := range versions {
		for _, s := range selected {
			if v.Version == s {
				totalSize += v.Size
				break
			}
		}
	}

	// Confirm removal
	confirmed, err := ConfirmRemoval(selected, totalSize)
	if err != nil {
		return fmt.Errorf("failed to get confirmation: %w", err)
	}

	if !confirmed {
		log.LogInfo("User cancelled cleanup")
		pterm.Info.Println("Cleanup cancelled")
		return nil
	}

	// Remove selected versions
	pterm.Println()
	if err := RemoveVersions(cfg.ZigDir, selected, formatter); err != nil {
		return err
	}

	log.LogInfo("Interactive cleanup completed: removed %d versions, freed %s", len(selected), FormatBytes(totalSize))

	pterm.Println()
	pterm.Success.Printf("Freed %s of disk space\n", FormatBytes(totalSize))

	return nil
}

// CleanupCommand is the main entry point for the cleanup command
func CleanupCommand(cfg *config.Config, log logger.ILogger, formatter OutputFormatter, dryRun bool, autoYes bool, keepLast int) error {
	formatter.PrintSection("Scanning for installed Zig versions")

	// Scan for versions
	versions, err := ScanInstalledVersions(cfg.ZigDir, cfg.BinDir)
	if err != nil {
		return fmt.Errorf("failed to scan versions: %w", err)
	}

	if len(versions) == 0 {
		formatter.PrintSuccess("Scan", "No Zig versions found")
		return nil
	}

	// Count non-current versions
	var removableVersions []VersionInfo
	for _, v := range versions {
		if !v.IsCurrent {
			removableVersions = append(removableVersions, v)
		}
	}

	if len(removableVersions) == 0 {
		formatter.PrintSuccess("Scan", "Only the current version is installed (nothing to clean up)")
		return nil
	}

	formatter.PrintSuccess("Scan", fmt.Sprintf("Found %d installed version(s)", len(versions)))

	pterm.Println()

	// Handle keep-last mode (auto-remove without prompting)
	if keepLast > 0 {
		return cleanupWithKeepLast(cfg, log, formatter, versions, keepLast, dryRun, autoYes)
	}

	// Interactive mode: display table and let user select
	if err := DisplayVersionsTable(versions, cfg.NoColor); err != nil {
		return err
	}

	// Prompt for selection
	selected, err := PromptVersionSelection(versions)
	if err != nil {
		return fmt.Errorf("failed to get selection: %w", err)
	}

	if len(selected) == 0 {
		log.LogInfo("No versions selected for removal")
		pterm.Info.Println("No versions selected")
		return nil
	}

	// Calculate total size
	var totalSize int64
	for _, v := range versions {
		for _, s := range selected {
			if v.Version == s {
				totalSize += v.Size
				break
			}
		}
	}

	// Dry run mode
	if dryRun {
		pterm.Info.Printf("Dry run mode: would remove %d version(s) and free %s\n", len(selected), FormatBytes(totalSize))
		for _, s := range selected {
			pterm.Printf("  - %s\n", s)
		}
		return nil
	}

	// Confirm removal (unless --yes flag)
	if !autoYes {
		confirmed, err := ConfirmRemoval(selected, totalSize)
		if err != nil {
			return fmt.Errorf("failed to get confirmation: %w", err)
		}

		if !confirmed {
			log.LogInfo("User cancelled cleanup")
			pterm.Info.Println("Cleanup cancelled")
			return nil
		}
	}

	// Remove versions
	pterm.Println()
	if err := RemoveVersions(cfg.ZigDir, selected, formatter); err != nil {
		return err
	}

	log.LogInfo("Cleanup completed: removed %d versions, freed %s", len(selected), FormatBytes(totalSize))

	pterm.Println()
	pterm.Success.Printf("Cleanup completed successfully!\n")
	pterm.Success.Printf("Freed %s of disk space\n", FormatBytes(totalSize))

	return nil
}

// cleanupWithKeepLast handles cleanup with --keep-last parameter
func cleanupWithKeepLast(cfg *config.Config, log logger.ILogger, formatter OutputFormatter, versions []VersionInfo, keepLast int, dryRun bool, autoYes bool) error {
	formatter.PrintSection(fmt.Sprintf("Auto-cleanup mode (keeping last %d versions)", keepLast))

	// Filter versions to remove
	toRemove := filterVersionsToKeep(versions, keepLast)

	if len(toRemove) == 0 {
		formatter.PrintSuccess("Auto-cleanup", "No versions to remove")
		return nil
	}

	// Show what will be kept
	var keptVersions []string
	for _, v := range versions {
		isInRemoveList := false
		for _, r := range toRemove {
			if v.Version == r.Version {
				isInRemoveList = true
				break
			}
		}
		if !isInRemoveList {
			keptVersions = append(keptVersions, v.Version)
		}
	}

	// Calculate total size
	var totalSize int64
	var versionNames []string
	for _, v := range toRemove {
		totalSize += v.Size
		versionNames = append(versionNames, v.Version)
	}

	pterm.Info.Printf("Keeping: %s\n", strings.Join(keptVersions, ", "))
	pterm.Info.Printf("Removing: %s\n", strings.Join(versionNames, ", "))
	pterm.Println()

	// Dry run mode
	if dryRun {
		pterm.Info.Printf("Dry run mode: would remove %d version(s) and free %s\n", len(toRemove), FormatBytes(totalSize))
		return nil
	}

	// Confirm unless --yes flag
	if !autoYes {
		confirmed, err := ConfirmRemoval(versionNames, totalSize)
		if err != nil {
			return fmt.Errorf("failed to get confirmation: %w", err)
		}

		if !confirmed {
			log.LogInfo("User cancelled cleanup")
			pterm.Info.Println("Cleanup cancelled")
			return nil
		}
	}

	// Remove versions
	pterm.Println()
	if err := RemoveVersions(cfg.ZigDir, versionNames, formatter); err != nil {
		return err
	}

	log.LogInfo("Cleanup completed: removed %d versions, freed %s", len(versionNames), FormatBytes(totalSize))

	pterm.Println()
	pterm.Success.Printf("Cleanup completed successfully!\n")
	pterm.Success.Printf("Freed %s of disk space\n", FormatBytes(totalSize))

	return nil
}
