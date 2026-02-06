package commands

import (
	"encoding/json"
	"os"

	"github.com/jongio/azd-core/cliout"
	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags.
var Version = "0.1.0"

// BuildTime is set at build time via -ldflags.
var BuildTime = "unknown"

// Commit is set at build time via -ldflags.
var Commit = "unknown"

// VersionInfo represents version information for JSON output.
type VersionInfo struct {
	Version   string `json:"version"`
	BuildTime string `json:"buildTime"`
	Commit    string `json:"commit"`
}

// NewVersionCommand creates the version command.
func NewVersionCommand() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:          "version",
		Short:        "Show version information",
		Long:         `Display the version of the azd copilot extension.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonOutput {
				info := VersionInfo{
					Version:   Version,
					BuildTime: BuildTime,
					Commit:    Commit,
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(info)
			}

			cliout.Section("🤖", "Azure Copilot Extension")
			cliout.Newline()
			cliout.Label("Version", Version)
			cliout.Label("Built", BuildTime)
			if Commit != "unknown" {
				cliout.Label("Commit", Commit)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output version as JSON")

	return cmd
}
