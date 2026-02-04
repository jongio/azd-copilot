package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags.
var Version = "dev"

// BuildTime is set at build time via -ldflags.
var BuildTime = "unknown"

// Commit is set at build time via -ldflags.
var Commit = "unknown"

// VersionInfo represents version information for JSON output.
type VersionInfo struct {
	Version   string `json:"version"`
	BuildTime string `json:"buildTime"`
}

// NewVersionCommand creates the version command.
func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:          "version",
		Short:        "Show version information",
		Long:         `Display the version of the azd copilot extension.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Version: %s\n", Version)
			fmt.Printf("Built: %s\n", BuildTime)
			return nil
		},
	}
}
