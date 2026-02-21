package installer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateZigSymlink(t *testing.T) {
	// Create temporary directories for testing
	tmpDir := t.TempDir()
	versionDir := filepath.Join(tmpDir, "zig-linux-x86_64-0.13.0")
	binDir := filepath.Join(tmpDir, "bin")

	// Create directories
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		t.Fatalf("Failed to create version dir: %v", err)
	}
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("Failed to create bin dir: %v", err)
	}

	// Create a fake zig binary
	zigBinary := filepath.Join(versionDir, "zig")
	if err := os.WriteFile(zigBinary, []byte("#!/bin/sh\necho 0.13.0"), 0o755); err != nil {
		t.Fatalf("Failed to create fake zig binary: %v", err)
	}

	// Create a mock formatter
	formatter := &MockFormatter{}

	// Test creating symlink
	err := UpdateZigSymlink(versionDir, binDir, "0.13.0", formatter)
	if err != nil {
		t.Errorf("UpdateZigSymlink failed: %v", err)
	}

	// Verify symlink was created
	linkPath := filepath.Join(binDir, "zig")
	target, err := os.Readlink(linkPath)
	if err != nil {
		t.Errorf("Failed to read symlink: %v", err)
	}

	expectedTarget := zigBinary
	if target != expectedTarget {
		t.Errorf("Symlink target mismatch: expected %s, got %s", expectedTarget, target)
	}

	// Test updating existing symlink
	versionDir2 := filepath.Join(tmpDir, "zig-linux-x86_64-0.14.0")
	if err := os.MkdirAll(versionDir2, 0o755); err != nil {
		t.Fatalf("Failed to create version dir 2: %v", err)
	}
	zigBinary2 := filepath.Join(versionDir2, "zig")
	if err := os.WriteFile(zigBinary2, []byte("#!/bin/sh\necho 0.14.0"), 0o755); err != nil {
		t.Fatalf("Failed to create fake zig binary 2: %v", err)
	}

	err = UpdateZigSymlink(versionDir2, binDir, "0.14.0", formatter)
	if err != nil {
		t.Errorf("UpdateZigSymlink (update) failed: %v", err)
	}

	// Verify symlink was updated
	target, err = os.Readlink(linkPath)
	if err != nil {
		t.Errorf("Failed to read updated symlink: %v", err)
	}

	expectedTarget = zigBinary2
	if target != expectedTarget {
		t.Errorf("Updated symlink target mismatch: expected %s, got %s", expectedTarget, target)
	}
}

func TestUpdateZigSymlink_BinaryNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	versionDir := filepath.Join(tmpDir, "zig-linux-x86_64-0.13.0")
	binDir := filepath.Join(tmpDir, "bin")

	// Create directories but no zig binary
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		t.Fatalf("Failed to create version dir: %v", err)
	}
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("Failed to create bin dir: %v", err)
	}

	formatter := &MockFormatter{}

	// Test should fail because zig binary doesn't exist
	err := UpdateZigSymlink(versionDir, binDir, "0.13.0", formatter)
	if err == nil {
		t.Error("Expected error when zig binary doesn't exist, got nil")
	}
}

func TestPromptVersionSwitch_SingleVersion(t *testing.T) {
	versions := []VersionInfo{
		{Version: "0.13.0", Path: "/fake/path", IsCurrent: true},
	}

	_, err := PromptVersionSwitch(versions)
	if err == nil {
		t.Error("Expected error with single version, got nil")
	}
}

func TestPromptVersionSwitch_NoVersions(t *testing.T) {
	versions := []VersionInfo{}

	_, err := PromptVersionSwitch(versions)
	if err == nil {
		t.Error("Expected error with no versions, got nil")
	}
}

// MockFormatter is a test implementation of OutputFormatter
type MockFormatter struct {
	Sections  []string
	Tasks     []string
	Successes []string
	Errors    []string
	Warnings  []string
	Progress  []string
}

func (m *MockFormatter) PrintSection(sectionName string) {
	m.Sections = append(m.Sections, sectionName)
}

func (m *MockFormatter) PrintProgress(name, output string) {
	m.Progress = append(m.Progress, name+": "+output)
}

func (m *MockFormatter) PrintSuccess(name, output string) {
	m.Successes = append(m.Successes, name+": "+output)
}

func (m *MockFormatter) PrintError(name, output string) {
	m.Errors = append(m.Errors, name+": "+output)
}

func (m *MockFormatter) PrintWarning(name, output string) {
	m.Warnings = append(m.Warnings, name+": "+output)
}

func (m *MockFormatter) PrintTask(section, name, output string) {
	m.Tasks = append(m.Tasks, section+": "+name+": "+output)
}
