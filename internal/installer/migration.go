package installer

import (
	"fmt"
	"os"
	"os/exec"
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

// PerformMigration removes system installation using sudo
func PerformMigration(systemDir string, formatter OutputFormatter, logger logger.ILogger) error {
	formatter.PrintProgress("Migration", "Preparing to remove system installation")

	// Build list of all paths to remove
	pathsToRemove := []string{systemDir}

	// Check for ZLS directories
	systemZLSDirs := []string{"/opt/zls", "/usr/local/zls"}
	for _, zlsDir := range systemZLSDirs {
		if _, err := os.Stat(zlsDir); err == nil {
			pathsToRemove = append(pathsToRemove, zlsDir)
		}
	}

	// Check for symlinks
	systemBinLinks := []string{"/usr/local/bin/zig", "/usr/local/bin/zls"}
	for _, link := range systemBinLinks {
		if _, err := os.Lstat(link); err == nil {
			pathsToRemove = append(pathsToRemove, link)
		}
	}

	// Show what will be removed
	formatter.PrintTask("Will remove", "Paths", fmt.Sprintf("%d items", len(pathsToRemove)))
	for _, path := range pathsToRemove {
		logger.LogInfo("  - %s", path)
	}

	// Execute sudo rm command
	formatter.PrintTask("Executing", "sudo rm", "You may be prompted for your password")

	// Build sudo command
	args := []string{"rm", "-rf"}
	args = append(args, pathsToRemove...)

	// Execute with proper error handling
	cmd := exec.Command("sudo", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if any paths still exist
		remainingPaths := []string{}
		for _, path := range pathsToRemove {
			if _, statErr := os.Lstat(path); statErr == nil {
				remainingPaths = append(remainingPaths, path)
			}
		}

		if len(remainingPaths) > 0 {
			formatter.PrintWarning("Partial Removal", "Some paths still exist")
			formatter.PrintTask("Manual Removal", "Required",
				fmt.Sprintf("Run this command:\n      sudo rm -rf %s", strings.Join(remainingPaths, " ")))
			return fmt.Errorf("failed to remove all paths: %w\nOutput: %s", err, string(output))
		}

		// If paths don't exist anymore, the removal worked despite the error
		logger.LogInfo("Removal completed despite error: %v", err)
	}

	// Verify removal
	removedCount := 0
	stillExist := []string{}

	for _, path := range pathsToRemove {
		if _, err := os.Lstat(path); err != nil {
			removedCount++
		} else {
			stillExist = append(stillExist, path)
		}
	}

	if len(stillExist) > 0 {
		formatter.PrintWarning("Incomplete", fmt.Sprintf("%d paths still exist", len(stillExist)))
		formatter.PrintTask("Manual Cleanup", "Required",
			fmt.Sprintf("Run this command:\n      sudo rm -rf %s", strings.Join(stillExist, " ")))
		return fmt.Errorf("failed to remove all system installations")
	}

	formatter.PrintSuccess("Migration", fmt.Sprintf("Removed %d system paths successfully", removedCount))
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
