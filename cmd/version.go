package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// BuildInfo contains version information
type BuildInfo struct {
	// Version is the application version (set during build)
	Version string
	// Commit is the git commit hash (set during build)
	Commit string
	// BuildDate is the build date (set during build)
	BuildDate string
}

// Default build information
var (
	Version   = "4.0.0"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// VersionCommand encapsulates the version command
type VersionCommand struct {
	cmd       *cobra.Command
	buildInfo BuildInfo
}

// NewVersionCommand creates a new version command instance
func NewVersionCommand() *VersionCommand {
	buildInfo := BuildInfo{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  `Display the version, commit hash, and build date of the Zig installer tool.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Zig Installer v%s\n", buildInfo.Version)
			fmt.Printf("Commit: %s\n", buildInfo.Commit)
			fmt.Printf("Built on: %s\n", buildInfo.BuildDate)
		},
	}

	return &VersionCommand{
		cmd:       versionCmd,
		buildInfo: buildInfo,
	}
}
