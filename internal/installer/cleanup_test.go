package installer

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestExtractVersionFromPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "standard version",
			path:     "/opt/zig/zig-linux-x86_64-0.13.0",
			expected: "0.13.0",
		},
		{
			name:     "dev version",
			path:     "/opt/zig/zig-linux-x86_64-0.12.0-dev.123+abc",
			expected: "0.12.0-dev.123+abc",
		},
		{
			name:     "master version",
			path:     "/opt/zig/zig-linux-x86_64-0.13.0-dev.46+3648d7df1",
			expected: "0.13.0-dev.46+3648d7df1",
		},
		{
			name:     "macos aarch64",
			path:     "/usr/local/zig/zig-macos-aarch64-0.11.0",
			expected: "0.11.0",
		},
		{
			name:     "invalid path",
			path:     "/opt/zig",
			expected: "",
		},
		{
			name:     "too few parts",
			path:     "/opt/zig/zig-linux",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractVersionFromPath(tt.path)
			if result != tt.expected {
				t.Errorf("extractVersionFromPath(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{
			name:     "bytes",
			bytes:    512,
			expected: "512 B",
		},
		{
			name:     "kilobytes",
			bytes:    1024,
			expected: "1 KB",
		},
		{
			name:     "megabytes",
			bytes:    1024 * 1024,
			expected: "1 MB",
		},
		{
			name:     "megabytes with decimal",
			bytes:    1024 * 1024 * 150,
			expected: "150 MB",
		},
		{
			name:     "gigabytes",
			bytes:    1024 * 1024 * 1024,
			expected: "1 GB",
		},
		{
			name:     "gigabytes with decimal",
			bytes:    1024 * 1024 * 1024 * 2,
			expected: "2 GB",
		},
		{
			name:     "zero bytes",
			bytes:    0,
			expected: "0 B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatBytes(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestCalculateDirectorySize(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create some test files
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	subdir := filepath.Join(tmpDir, "subdir")
	file3 := filepath.Join(subdir, "file3.txt")

	// Write files with known sizes
	if err := os.WriteFile(file1, make([]byte, 100), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2, make([]byte, 200), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}
	if err := os.WriteFile(file3, make([]byte, 300), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Calculate size
	size, err := CalculateDirectorySize(tmpDir)
	if err != nil {
		t.Fatalf("CalculateDirectorySize failed: %v", err)
	}

	// Expected size: 100 + 200 + 300 = 600 bytes
	expected := int64(600)
	if size != expected {
		t.Errorf("CalculateDirectorySize(%q) = %d, want %d", tmpDir, size, expected)
	}
}

func TestFilterVersionsToKeep(t *testing.T) {
	// Create test versions with different install dates
	now := time.Now()
	versions := []VersionInfo{
		{
			Version:     "0.13.0",
			Path:        "/opt/zig/zig-linux-x86_64-0.13.0",
			Size:        1000,
			InstallDate: now.Add(-1 * time.Hour),
			IsCurrent:   true,
		},
		{
			Version:     "0.12.0",
			Path:        "/opt/zig/zig-linux-x86_64-0.12.0",
			Size:        900,
			InstallDate: now.Add(-24 * time.Hour),
			IsCurrent:   false,
		},
		{
			Version:     "0.11.0",
			Path:        "/opt/zig/zig-linux-x86_64-0.11.0",
			Size:        800,
			InstallDate: now.Add(-48 * time.Hour),
			IsCurrent:   false,
		},
		{
			Version:     "0.10.1",
			Path:        "/opt/zig/zig-linux-x86_64-0.10.1",
			Size:        700,
			InstallDate: now.Add(-72 * time.Hour),
			IsCurrent:   false,
		},
	}

	tests := []struct {
		name          string
		keepLast      int
		expectedCount int
		shouldInclude []string
	}{
		{
			name:          "keep last 2 (excluding current)",
			keepLast:      2,
			expectedCount: 1,
			shouldInclude: []string{"0.10.1"},
		},
		{
			name:          "keep last 1 (excluding current)",
			keepLast:      1,
			expectedCount: 2,
			shouldInclude: []string{"0.11.0", "0.10.1"},
		},
		{
			name:          "keep last 3 (excluding current)",
			keepLast:      3,
			expectedCount: 0,
			shouldInclude: []string{},
		},
		{
			name:          "keep none (0)",
			keepLast:      0,
			expectedCount: 0,
			shouldInclude: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterVersionsToKeep(versions, tt.keepLast)

			if len(result) != tt.expectedCount {
				t.Errorf("filterVersionsToKeep() returned %d versions, want %d", len(result), tt.expectedCount)
			}

			// Check that current version is never included
			for _, v := range result {
				if v.IsCurrent {
					t.Error("filterVersionsToKeep() included current version, which should never be removed")
				}
			}

			// Check specific versions
			for _, expectedVersion := range tt.shouldInclude {
				found := false
				for _, v := range result {
					if v.Version == expectedVersion {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("filterVersionsToKeep() missing expected version %s", expectedVersion)
				}
			}
		})
	}
}

func TestScanInstalledVersions(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()
	binDir := t.TempDir()

	// Create test version directories
	dirs := []string{
		"zig-linux-x86_64-0.13.0",
		"zig-linux-x86_64-0.12.0",
		"zig-linux-x86_64-0.11.0",
		"not-zig-folder", // Should be ignored
	}

	for _, dir := range dirs {
		path := filepath.Join(tmpDir, dir)
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatalf("failed to create test directory: %v", err)
		}
		// Create a dummy file in each directory
		testFile := filepath.Join(path, "zig")
		if err := os.WriteFile(testFile, []byte("test"), 0o644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	// Create a symlink to simulate current version
	currentPath := filepath.Join(tmpDir, "zig-linux-x86_64-0.13.0", "zig")
	linkPath := filepath.Join(binDir, "zig")
	if err := os.Symlink(currentPath, linkPath); err != nil {
		t.Fatalf("failed to create test symlink: %v", err)
	}

	// Scan versions
	versions, err := ScanInstalledVersions(tmpDir, binDir)
	if err != nil {
		t.Fatalf("ScanInstalledVersions failed: %v", err)
	}

	// Should find 3 versions (not the "not-zig-folder")
	if len(versions) != 3 {
		t.Errorf("ScanInstalledVersions found %d versions, want 3", len(versions))
	}

	// Check that 0.13.0 is marked as current
	var foundCurrent bool
	for _, v := range versions {
		if v.Version == "0.13.0" {
			if !v.IsCurrent {
				t.Error("version 0.13.0 should be marked as current")
			}
			foundCurrent = true
		} else if v.IsCurrent {
			t.Errorf("version %s should not be marked as current", v.Version)
		}
	}

	if !foundCurrent {
		t.Error("no version marked as current")
	}

	// Verify versions are sorted by install date (newest first)
	// Since we created them in order, they should be in reverse order
	if len(versions) >= 2 {
		if versions[0].InstallDate.Before(versions[1].InstallDate) {
			t.Error("versions are not sorted by install date (newest first)")
		}
	}
}

func TestGetCurrentVersion(t *testing.T) {
	tmpDir := t.TempDir()
	binDir := t.TempDir()

	// Test case 1: Valid symlink
	targetPath := filepath.Join(tmpDir, "zig-linux-x86_64-0.13.0", "zig")
	linkPath := filepath.Join(binDir, "zig")

	// Create target directory
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatalf("failed to create target directory: %v", err)
	}
	if err := os.WriteFile(targetPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create target file: %v", err)
	}

	// Create symlink
	if err := os.Symlink(targetPath, linkPath); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	version, err := GetCurrentVersion(binDir)
	if err != nil {
		t.Fatalf("GetCurrentVersion failed: %v", err)
	}

	if version != "0.13.0" {
		t.Errorf("GetCurrentVersion() = %q, want %q", version, "0.13.0")
	}

	// Test case 2: No symlink exists
	emptyBinDir := t.TempDir()
	version, err = GetCurrentVersion(emptyBinDir)
	if err != nil {
		t.Fatalf("GetCurrentVersion failed when no symlink exists: %v", err)
	}

	if version != "" {
		t.Errorf("GetCurrentVersion() = %q, want empty string when no symlink exists", version)
	}
}
