package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// VersionInfo represents an installed Zig version
type VersionInfo struct {
	Version     string
	Path        string
	Size        int64
	InstallDate time.Time
	IsCurrent   bool
}

func extractVersionFromPath(path string) string {
	base := filepath.Base(path)
	parts := strings.Split(base, "-")
	if len(parts) < 4 {
		return ""
	}
	return strings.Join(parts[3:], "-")
}

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
	units := []string{"KB", "MB", "GB", "TB", "PB"}
	if exp >= len(units) {
		exp = len(units) - 1
		div = 1 << (10 * uint(exp))
	}
	return fmt.Sprintf("%.0f %s", float64(bytes)/float64(div), units[exp])
}

func GetCurrentVersion(binDir string) (string, error) {
	linkPath := filepath.Join(binDir, "zig")
	target, err := os.Readlink(linkPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	dir := filepath.Dir(target)
	return extractVersionFromPath(dir), nil
}

func ScanInstalledVersions(zigDir, binDir string) ([]VersionInfo, error) {
	entries, err := os.ReadDir(zigDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read zig directory: %w", err)
	}

	currentVersion, err := GetCurrentVersion(binDir)
	if err != nil {
		currentVersion = ""
	}

	var versions []VersionInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if !strings.HasPrefix(entry.Name(), "zig-") {
			continue
		}

		path := filepath.Join(zigDir, entry.Name())
		version := extractVersionFromPath(path)
		if version == "" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		size, err := CalculateDirectorySize(path)
		if err != nil {
			size = 0
		}

		versions = append(versions, VersionInfo{
			Version:     version,
			Path:        path,
			Size:        size,
			InstallDate: info.ModTime(),
			IsCurrent:   version == currentVersion,
		})
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i].InstallDate.After(versions[j].InstallDate)
	})

	return versions, nil
}

// filterVersionsToKeep returns versions that should be REMOVED
func filterVersionsToKeep(versions []VersionInfo, keepLast int) []VersionInfo {
	if keepLast <= 0 {
		return nil
	}

	sorted := make([]VersionInfo, len(versions))
	copy(sorted, versions)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].InstallDate.After(sorted[j].InstallDate)
	})

	var toRemove []VersionInfo
	kept := 0
	for _, v := range sorted {
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
