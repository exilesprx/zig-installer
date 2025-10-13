package installer

import (
	"testing"
)

func TestConvertToSemanticVersion(t *testing.T) {
	tests := []struct {
		name        string
		zigVersion  string
		expected    string
		description string
	}{
		{
			name:        "empty version",
			zigVersion:  "",
			expected:    "",
			description: "Empty string should return empty string",
		},
		{
			name:        "master version",
			zigVersion:  "master",
			expected:    "master",
			description: "Master version should return master",
		},
		{
			name:        "development version with dev",
			zigVersion:  "0.16.0-dev.728+87c18945c",
			expected:    "master",
			description: "Development versions containing -dev. should map to master",
		},
		{
			name:        "another development version",
			zigVersion:  "0.15.0-dev.1234+abcdef12",
			expected:    "master",
			description: "Any version with -dev. should map to master branch",
		},
		{
			name:        "release version three parts",
			zigVersion:  "0.15.0",
			expected:    "0.15.0",
			description: "Clean semantic version should be returned as-is",
		},
		{
			name:        "release version with pre-release",
			zigVersion:  "0.14.0-rc1",
			expected:    "0.14.0",
			description: "Version with pre-release should strip pre-release part",
		},
		{
			name:        "version two parts",
			zigVersion:  "0.15",
			expected:    "0.15.0",
			description: "Two-part version should get .0 appended",
		},
		{
			name:        "version with build metadata",
			zigVersion:  "0.13.0+build123",
			expected:    "0.13.0",
			description: "Version with build metadata should strip metadata",
		},
		{
			name:        "single digit version",
			zigVersion:  "1",
			expected:    "",
			description: "Single component version should return empty (invalid)",
		},
		{
			name:        "complex pre-release",
			zigVersion:  "0.12.0-alpha.1",
			expected:    "0.12.0",
			description: "Complex pre-release should be stripped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToSemanticVersion(tt.zigVersion)
			if result != tt.expected {
				t.Errorf("convertToSemanticVersion(%q) = %q, expected %q\nDescription: %s",
					tt.zigVersion, result, tt.expected, tt.description)
			}
		})
	}
}
