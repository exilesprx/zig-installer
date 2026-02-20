package installer

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/exilesprx/zig-installer/internal/config"
	"github.com/exilesprx/zig-installer/internal/tui"
)

func TestTaskFormatter_PrintSection(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create test configuration
	cfg := &config.Config{NoColor: true}
	styles := tui.LoadStyles()

	// Create formatter
	formatter := NewTaskFormatter(cfg, styles)

	// Test printing a section
	formatter.PrintSection("Test Section")

	// Restore stdout and read captured output
	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	if err != nil {
		t.Fatalf("Failed to copy output: %v", err)
	}
	output := buf.String()

	expected := "==> Test Section\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestTaskFormatter_PrintTask(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create test configuration
	cfg := &config.Config{NoColor: true, Verbose: false}
	styles := tui.LoadStyles()

	// Create formatter
	formatter := NewTaskFormatter(cfg, styles)

	// Test printing a task
	formatter.PrintSuccess("Task Name", "Additional output")

	// Restore stdout and read captured output
	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	if err != nil {
		t.Fatalf("Failed to copy output: %v", err)
	}
	output := buf.String()

	expected := "  --> Success: Task Name\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestTaskFormatter_PrintTaskWithVerbose(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create test configuration with verbose enabled
	cfg := &config.Config{NoColor: true, Verbose: true}
	styles := tui.LoadStyles()

	// Create formatter
	formatter := NewTaskFormatter(cfg, styles)

	// Test printing a task with additional output
	formatter.PrintSuccess("Task Name", "Additional verbose output")

	// Restore stdout and read captured output
	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	if err != nil {
		t.Fatalf("Failed to copy output: %v", err)
	}
	output := buf.String()

	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("Expected 2 lines of output, got %d: %v", len(lines), lines)
	}

	expectedFirstLine := "  --> Success: Task Name"
	if lines[0] != expectedFirstLine {
		t.Errorf("Expected first line %q, got %q", expectedFirstLine, lines[0])
	}

	expectedSecondLine := "    Additional verbose output"
	if lines[1] != expectedSecondLine {
		t.Errorf("Expected second line %q, got %q", expectedSecondLine, lines[1])
	}
}

func TestTaskFormatter_NilConfigAndStyles(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create formatter with nil config and styles
	formatter := NewTaskFormatter(nil, nil)

	// Test printing with nil dependencies (should use fallback behavior)
	formatter.PrintSection("Test Section")
	formatter.PrintSuccess("Task Name", "Additional output")

	// Restore stdout and read captured output
	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	if err != nil {
		t.Fatalf("Failed to copy output: %v", err)
	}
	output := buf.String()

	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines of output, got %d: %v", len(lines), lines)
	}

	expectedLines := []string{
		"==> Test Section",
		"  Success: Task Name",
		"    Additional output",
	}

	for i, expectedLine := range expectedLines {
		if lines[i] != expectedLine {
			t.Errorf("Expected line %d %q, got %q", i+1, expectedLine, lines[i])
		}
	}
}
