package commands

import (
	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
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

// NewVersionCommand creates the version command.
func NewVersionCommand(extensionId, version string, outputFormat *string) *cobra.Command {
	return azdext.NewVersionCommand(extensionId, version, outputFormat)
}
