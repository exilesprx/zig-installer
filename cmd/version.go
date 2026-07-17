package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Default build information (set via ldflags at build time)
var (
	Version   = "unknown"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// VersionCommand encapsulates the version command
type VersionCommand struct {
	cmd *cobra.Command
}

// NewVersionCommand creates a new version command instance
func NewVersionCommand() *VersionCommand {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  `Display the version, commit hash, and build date of the Zig installer tool.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Zig Installer %s\n", Version)
			fmt.Printf("Commit: %s\n", Commit)
			fmt.Printf("Built on: %s\n", BuildDate)
		},
	}

	return &VersionCommand{
		cmd: versionCmd,
	}
}

// Command returns the cobra command
func (vc *VersionCommand) Command() *cobra.Command {
	return vc.cmd
}
