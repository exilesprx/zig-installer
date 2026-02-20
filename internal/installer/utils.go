package installer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/exilesprx/zig-install/internal/config"
	"github.com/exilesprx/zig-install/internal/tui"
)

// OutputFormatter handles formatted output for installation tasks
type OutputFormatter interface {
	PrintSection(sectionName string)
	PrintProgress(name, output string)
	PrintSuccess(name, output string)
	PrintError(name, output string)
	PrintTask(name, status, output string)
	PrintWarning(name, output string)
}

// TaskFormatter implements OutputFormatter with styling support
type TaskFormatter struct {
	config *config.Config
	styles *tui.Styles
}

// NewTaskFormatter creates a new TaskFormatter instance
func NewTaskFormatter(config *config.Config, styles *tui.Styles) *TaskFormatter {
	return &TaskFormatter{
		config: config,
		styles: styles,
	}
}

// PrintSection prints a top-level section header
func (tf *TaskFormatter) PrintSection(sectionName string) {
	if tf.config == nil || tf.styles == nil {
		fmt.Printf("==> %s\n", sectionName)
		return
	}

	if tf.config.NoColor {
		fmt.Printf("==> %s\n", sectionName)
	} else {
		fmt.Printf("%s %s\n",
			tf.styles.TopLevel.Render("==>"),
			tf.styles.Grey.Render(sectionName))
	}
}

// PrintTask prints a task completion message with optional detailed output as a sub-level item
func (tf *TaskFormatter) PrintTask(name, status, output string) {
	if tf.config == nil || tf.styles == nil {
		// Fallback to simple print if config/styles aren't set
		fmt.Printf("  %s %s\n", status, name)
		if output != "" {
			fmt.Printf("    %s\n", output)
		}
		return
	}

	if tf.config.NoColor {
		// Always print as sub-level with --> prefix
		fmt.Printf("  --> %s %s\n", status, name)
		if output != "" && tf.config.Verbose {
			fmt.Printf("    %s\n", output)
		}
	} else {
		// Color the status based on content
		var statusColor lipgloss.Style

		if strings.Contains(status, "Success") || strings.Contains(status, "Already installed") {
			statusColor = tf.styles.Success // Green for success
		} else if strings.Contains(status, "Failed") || strings.Contains(status, "Error") {
			statusColor = tf.styles.Error // Red for errors
		} else if strings.Contains(status, "Warning") {
			statusColor = tf.styles.Warning // Yellow for warnings
		} else {
			statusColor = tf.styles.Grey // Grey for other statuses
		}

		textColor := tf.styles.Grey // Grey for task names

		// Print sub-level message with --> prefix
		fmt.Printf("%s %s %s\n",
			tf.styles.SubLevel.Render("  -->"),
			statusColor.Render(status),
			textColor.Render(name))

		// Print additional output if present with more indentation
		if output != "" && tf.config.Verbose {
			fmt.Printf("    %s\n", tf.styles.Grey.Render(output))
		}
	}
}

func (tf *TaskFormatter) PrintProgress(name, output string) {
	tf.PrintTask(name, "Processing...", output)
}

func (tf *TaskFormatter) PrintSuccess(name, output string) {
	tf.PrintTask(name, "Success:", output)
}

func (tf *TaskFormatter) PrintError(name, output string) {
	tf.PrintTask(name, "Failed:", output)
}

func (tf *TaskFormatter) PrintWarning(name, output string) {
	tf.PrintTask(name, "Warning:", output)
}

// getZigVersion fetches version information from ziglang.org
func getZigVersion(zigIndexURL string, requestedVersion string) (*ZigVersionInfo, error) {
	resp, err := http.Get(zigIndexURL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var versions map[string]ZigVersionInfo
	if err := json.Unmarshal(body, &versions); err != nil {
		return nil, err
	}

	version := requestedVersion
	if version == "" || version == "master" {
		version = "master"
	}

	versionInfo, ok := versions[version]
	if !ok {
		return nil, fmt.Errorf("version %s not found", version)
	}

	// For non-master versions, use the key from the JSON as the version
	// since these entries don't have a version field
	if version != "master" {
		versionInfo.Version = version
	} else if versionInfo.Version == "" {
		// Only check for empty version on master, since it should have one
		return nil, fmt.Errorf("could not determine master version")
	}

	return &versionInfo, nil
}

// isZigInstalled checks if the specified version is already installed
func isZigInstalled(version string) bool {
	cmd := exec.Command("zig", "version")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Check if version string contains the specified version
	// This handles cases like "0.11.0" matching "0.11.0-dev.3180+hash"
	installedVersion := strings.TrimSpace(string(output))

	// Handle master version specially
	if version == "master" {
		return strings.Contains(installedVersion, "-dev.")
	}

	return strings.HasPrefix(installedVersion, version)
}
