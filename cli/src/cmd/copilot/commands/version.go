package commands

import (
	coreversion "github.com/jongio/azd-core/version"
	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags.
var Version = "0.1.0"

// BuildTime is set at build time via -ldflags.
var BuildTime = "unknown"

// Commit is set at build time via -ldflags.
var Commit = "unknown"

// VersionInfo provides the shared version information for this extension.
var VersionInfo = coreversion.New("jongio.azd.copilot", "azd copilot")

func init() {
	VersionInfo.Version = Version
	VersionInfo.BuildDate = BuildTime
	VersionInfo.GitCommit = Commit
}

// NewVersionCommand creates the version command using azd-core's version
// command which provides --quiet flag and full JSON output (version, buildDate,
// gitCommit, extensionId, name) for backward compatibility.
func NewVersionCommand(outputFormat *string) *cobra.Command {
	return coreversion.NewCommand(VersionInfo, outputFormat)
}
