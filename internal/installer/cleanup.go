package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/exilesprx/zig-installer/internal/config"
	"github.com/exilesprx/zig-installer/internal/logger"
	"github.com/pterm/pterm"
)

func RemoveVersions(zigDir string, versions []string, formatter OutputFormatter) error {
	formatter.PrintSection("Removing versions")

	for _, version := range versions {
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

		size, _ := CalculateDirectorySize(dirToRemove)
		formatter.PrintProgress("Version removal", fmt.Sprintf("Removing %s", version))
		if err := os.RemoveAll(dirToRemove); err != nil {
			formatter.PrintError("Version removal", fmt.Sprintf("Failed to remove %s: %v", version, err))
			return fmt.Errorf("failed to remove %s: %w", version, err)
		}
		formatter.PrintSuccess("Version removal", fmt.Sprintf("Removed %s (%s)", version, FormatBytes(size)))
	}

	return nil
}

func AutoCleanupPrompt(cfg *config.Config, log logger.Logger, formatter OutputFormatter, currentVersion string) error {
	versions, err := ScanInstalledVersions(cfg.ZigDir, cfg.BinDir)
	if err != nil {
		return fmt.Errorf("failed to scan versions: %w", err)
	}

	var otherVersions []VersionInfo
	var totalSize int64
	for _, v := range versions {
		if !v.IsCurrent {
			otherVersions = append(otherVersions, v)
			totalSize += v.Size
		}
	}

	if len(otherVersions) == 0 {
		log.LogInfo("No other versions found for cleanup")
		return nil
	}

	pterm.Println()
	pterm.Info.Printf("Found %d other installed version(s) (%s)\n", len(otherVersions), FormatBytes(totalSize))

	if cfg.AutoCleanup && cfg.KeepLast > 0 {
		return autoCleanupWithKeepLast(cfg, log, formatter, versions)
	}

	if cfg.AutoCleanup {
		return interactiveCleanup(cfg, log, formatter, versions)
	}

	var wantsCleanup bool
	prompt := &survey.Confirm{
		Message: "Clean up old versions?",
		Default: true,
	}

	if err := survey.AskOne(prompt, &wantsCleanup); err != nil {
		return err
	}

	if !wantsCleanup {
		log.LogInfo("User declined cleanup")
		return nil
	}

	return interactiveCleanup(cfg, log, formatter, versions)
}

func autoCleanupWithKeepLast(cfg *config.Config, log logger.Logger, formatter OutputFormatter, versions []VersionInfo) error {
	pterm.Println()
	formatter.PrintSection(fmt.Sprintf("Auto-cleanup (keeping last %d versions)", cfg.KeepLast))

	toRemove := filterVersionsToKeep(versions, cfg.KeepLast)
	if len(toRemove) == 0 {
		formatter.PrintSuccess("Auto-cleanup", "No versions to remove")
		return nil
	}

	var totalSize int64
	var versionNames []string
	for _, v := range toRemove {
		totalSize += v.Size
		versionNames = append(versionNames, v.Version)
	}

	formatter.PrintProgress("Auto-cleanup", fmt.Sprintf("Found %d version(s) to remove: %s (%s)",
		len(toRemove), strings.Join(versionNames, ", "), FormatBytes(totalSize)))

	if err := RemoveVersions(cfg.ZigDir, versionNames, formatter); err != nil {
		return err
	}

	log.LogInfo("Auto-cleanup completed: removed %d versions, freed %s", len(versionNames), FormatBytes(totalSize))
	pterm.Println()
	pterm.Success.Printf("Freed %s of disk space\n", FormatBytes(totalSize))
	return nil
}

func interactiveCleanup(cfg *config.Config, log logger.Logger, formatter OutputFormatter, versions []VersionInfo) error {
	pterm.Println()

	if err := DisplayVersionsTable(versions, cfg.NoColor); err != nil {
		return err
	}

	selected, err := PromptVersionSelection(versions)
	if err != nil {
		return fmt.Errorf("failed to get selection: %w", err)
	}

	if len(selected) == 0 {
		log.LogInfo("No versions selected for removal")
		pterm.Info.Println("No versions selected")
		return nil
	}

	var totalSize int64
	for _, v := range versions {
		for _, s := range selected {
			if v.Version == s {
				totalSize += v.Size
				break
			}
		}
	}

	confirmed, err := ConfirmRemoval(selected, totalSize)
	if err != nil {
		return fmt.Errorf("failed to get confirmation: %w", err)
	}

	if !confirmed {
		log.LogInfo("User cancelled cleanup")
		pterm.Info.Println("Cleanup cancelled")
		return nil
	}

	pterm.Println()
	if err := RemoveVersions(cfg.ZigDir, selected, formatter); err != nil {
		return err
	}

	log.LogInfo("Interactive cleanup completed: removed %d versions, freed %s", len(selected), FormatBytes(totalSize))
	pterm.Println()
	pterm.Success.Printf("Freed %s of disk space\n", FormatBytes(totalSize))
	return nil
}

func CleanupCommand(cfg *config.Config, log logger.Logger, formatter OutputFormatter, dryRun bool, autoYes bool, keepLast int) error {
	formatter.PrintSection("Scanning for installed Zig versions")

	versions, err := ScanInstalledVersions(cfg.ZigDir, cfg.BinDir)
	if err != nil {
		return fmt.Errorf("failed to scan versions: %w", err)
	}

	if len(versions) == 0 {
		formatter.PrintSuccess("Scan", "No Zig versions found")
		return nil
	}

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

	if keepLast > 0 {
		return cleanupWithKeepLast(cfg, log, formatter, versions, keepLast, dryRun, autoYes)
	}

	if err := DisplayVersionsTable(versions, cfg.NoColor); err != nil {
		return err
	}

	selected, err := PromptVersionSelection(versions)
	if err != nil {
		return fmt.Errorf("failed to get selection: %w", err)
	}

	if len(selected) == 0 {
		log.LogInfo("No versions selected for removal")
		pterm.Info.Println("No versions selected")
		return nil
	}

	var totalSize int64
	for _, v := range versions {
		for _, s := range selected {
			if v.Version == s {
				totalSize += v.Size
				break
			}
		}
	}

	if dryRun {
		pterm.Info.Printf("Dry run mode: would remove %d version(s) and free %s\n", len(selected), FormatBytes(totalSize))
		for _, s := range selected {
			pterm.Printf("  - %s\n", s)
		}
		return nil
	}

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

func cleanupWithKeepLast(cfg *config.Config, log logger.Logger, formatter OutputFormatter, versions []VersionInfo, keepLast int, dryRun bool, autoYes bool) error {
	formatter.PrintSection(fmt.Sprintf("Auto-cleanup mode (keeping last %d versions)", keepLast))

	toRemove := filterVersionsToKeep(versions, keepLast)
	if len(toRemove) == 0 {
		formatter.PrintSuccess("Auto-cleanup", "No versions to remove")
		return nil
	}

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

	var totalSize int64
	var versionNames []string
	for _, v := range toRemove {
		totalSize += v.Size
		versionNames = append(versionNames, v.Version)
	}

	pterm.Info.Printf("Keeping: %s\n", strings.Join(keptVersions, ", "))
	pterm.Info.Printf("Removing: %s\n", strings.Join(versionNames, ", "))
	pterm.Println()

	if dryRun {
		pterm.Info.Printf("Dry run mode: would remove %d version(s) and free %s\n", len(toRemove), FormatBytes(totalSize))
		return nil
	}

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
