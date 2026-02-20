package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetUserLocalPaths(t *testing.T) {
	zigDir, zlsDir, binDir, err := getUserLocalPaths()
	if err != nil {
		t.Fatalf("getUserLocalPaths() failed: %v", err)
	}

	// Get user home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home directory: %v", err)
	}

	// Test zigDir
	expectedZigDir := filepath.Join(homeDir, ".local", "share", "zig")
	if zigDir != expectedZigDir {
		t.Errorf("zigDir = %q, want %q", zigDir, expectedZigDir)
	}

	// Test zlsDir
	expectedZlsDir := filepath.Join(homeDir, ".local", "share", "zls")
	if zlsDir != expectedZlsDir {
		t.Errorf("zlsDir = %q, want %q", zlsDir, expectedZlsDir)
	}

	// Test binDir
	expectedBinDir := filepath.Join(homeDir, ".local", "bin")
	if binDir != expectedBinDir {
		t.Errorf("binDir = %q, want %q", binDir, expectedBinDir)
	}

	// Verify paths contain .local
	if !strings.Contains(zigDir, ".local") {
		t.Error("zigDir should contain .local")
	}
	if !strings.Contains(zlsDir, ".local") {
		t.Error("zlsDir should contain .local")
	}
	if !strings.Contains(binDir, ".local") {
		t.Error("binDir should contain .local")
	}
}

func TestGetDefaults(t *testing.T) {
	cfg := GetDefaults()

	// Test that ZigDir contains .local
	if !strings.Contains(cfg.ZigDir, ".local") {
		t.Errorf("ZigDir should contain .local, got: %s", cfg.ZigDir)
	}

	// Test that ZLSDir contains .local
	if !strings.Contains(cfg.ZLSDir, ".local") {
		t.Errorf("ZLSDir should contain .local, got: %s", cfg.ZLSDir)
	}

	// Test that BinDir contains .local
	if !strings.Contains(cfg.BinDir, ".local") {
		t.Errorf("BinDir should contain .local, got: %s", cfg.BinDir)
	}

	// Test that ZigDir ends with /zig
	if !strings.HasSuffix(cfg.ZigDir, "zig") {
		t.Errorf("ZigDir should end with 'zig', got: %s", cfg.ZigDir)
	}

	// Test that ZLSDir ends with /zls
	if !strings.HasSuffix(cfg.ZLSDir, "zls") {
		t.Errorf("ZLSDir should end with 'zls', got: %s", cfg.ZLSDir)
	}

	// Test that BinDir ends with /bin
	if !strings.HasSuffix(cfg.BinDir, "bin") {
		t.Errorf("BinDir should end with 'bin', got: %s", cfg.BinDir)
	}

	// Test other default values
	if cfg.ZigPubKey == "" {
		t.Error("ZigPubKey should not be empty")
	}

	if cfg.ZigDownURL == "" {
		t.Error("ZigDownURL should not be empty")
	}

	if cfg.ZigIndexURL == "" {
		t.Error("ZigIndexURL should not be empty")
	}

	if cfg.LogFile == "" {
		t.Error("LogFile should not be empty")
	}
}

func TestGetSystemZigDirs(t *testing.T) {
	dirs := GetSystemZigDirs()

	// Should return at least one directory
	if len(dirs) == 0 {
		t.Fatal("GetSystemZigDirs should return at least one directory")
	}

	switch runtime.GOOS {
	case "linux":
		// Linux should have /opt/zig and /usr/local/zig
		expectedDirs := []string{"/opt/zig", "/usr/local/zig"}
		if len(dirs) != len(expectedDirs) {
			t.Errorf("Linux should return %d directories, got %d", len(expectedDirs), len(dirs))
		}
		for _, expected := range expectedDirs {
			found := false
			for _, dir := range dirs {
				if dir == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected directory %q not found in results", expected)
			}
		}

	case "darwin":
		// macOS should have /usr/local/zig, /opt/zig
		expectedDirs := []string{"/usr/local/zig", "/opt/zig"}
		if len(dirs) != len(expectedDirs) {
			t.Errorf("macOS should return %d directories, got %d", len(expectedDirs), len(dirs))
		}
		for _, expected := range expectedDirs {
			found := false
			for _, dir := range dirs {
				if dir == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected directory %q not found in results", expected)
			}
		}

	default:
		t.Logf("Unsupported platform: %s", runtime.GOOS)
	}

	// Verify none of the system directories contain .local
	for _, dir := range dirs {
		if strings.Contains(dir, ".local") {
			t.Errorf("System directory should not contain .local: %s", dir)
		}
	}
}

func TestDetectSystemInstallation(t *testing.T) {
	// Note: This test depends on the actual filesystem state
	// It won't fail if no system installation exists, just verify it returns correct values

	dir, hasSystem := DetectSystemInstallation()

	// Should return a string and boolean
	// We can't assert the value since it depends on the system state
	t.Logf("System installation detected: %v at %q", hasSystem, dir)

	// If true, at least one system directory should exist
	if hasSystem {
		dirs := GetSystemZigDirs()
		found := false
		for _, dir := range dirs {
			if _, err := os.Stat(dir); err == nil {
				found = true
				t.Logf("Found system installation at: %s", dir)
				break
			}
		}
		if !found {
			t.Error("DetectSystemInstallation returned true but no system directories exist")
		}
	}
}

func TestSystemPathsDoNotOverlapUserPaths(t *testing.T) {
	systemDirs := GetSystemZigDirs()
	zigDir, zlsDir, _, err := getUserLocalPaths()
	if err != nil {
		t.Fatalf("getUserLocalPaths() failed: %v", err)
	}

	// Ensure system directories don't overlap with user-local directories
	for _, sysDir := range systemDirs {
		if sysDir == zigDir {
			t.Errorf("System directory %q overlaps with user zigDir %q", sysDir, zigDir)
		}
		if sysDir == zlsDir {
			t.Errorf("System directory %q overlaps with user zlsDir %q", sysDir, zlsDir)
		}
	}
}

func TestUserLocalPathsAreAbsolute(t *testing.T) {
	zigDir, zlsDir, binDir, err := getUserLocalPaths()
	if err != nil {
		t.Fatalf("getUserLocalPaths() failed: %v", err)
	}

	if !filepath.IsAbs(zigDir) {
		t.Errorf("zigDir should be absolute path, got: %s", zigDir)
	}

	if !filepath.IsAbs(zlsDir) {
		t.Errorf("zlsDir should be absolute path, got: %s", zlsDir)
	}

	if !filepath.IsAbs(binDir) {
		t.Errorf("binDir should be absolute path, got: %s", binDir)
	}
}
