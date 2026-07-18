package installer

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pterm/pterm"
)

func DisplayVersionsTable(versions []VersionInfo, noColor bool) error {
	if len(versions) == 0 {
		return errors.New("no versions found")
	}

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

	if err := func() error {
		if noColor {
			pterm.DisableColor()
			defer pterm.EnableColor()
		}
		if err := pterm.DefaultTable.WithHasHeader().WithData(tableData).Render(); err != nil {
			return err
		}
		pterm.Println()
		pterm.Printf("Total disk usage: %s\n", FormatBytes(totalSize))
		pterm.Println()
		return nil
	}(); err != nil {
		return err
	}

	return nil
}

func PromptVersionSelection(versions []VersionInfo) ([]string, error) {
	if len(versions) == 0 {
		return nil, errors.New("no versions available for selection")
	}

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

	var selected []string
	prompt := &survey.MultiSelect{
		Message: "Select versions to remove (space to select, enter to confirm):",
		Options: options,
	}

	if err := survey.AskOne(prompt, &selected); err != nil {
		return nil, err
	}

	var cleaned []string
	for _, s := range selected {
		parts := strings.Split(s, " ")
		cleaned = append(cleaned, parts[0])
	}
	return cleaned, nil
}

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
