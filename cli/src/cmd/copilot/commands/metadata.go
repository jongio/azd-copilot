package commands

import (
	"encoding/json"
	"fmt"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/azure/azure-dev/cli/azd/pkg/extensions"
	"github.com/spf13/cobra"
)

// metadataSchemaVersion is the version of the extension metadata schema this
// command emits. It describes the metadata document, not azd-copilot.
const metadataSchemaVersion = "1.0"

// metadataExtensionID must match the id in extension.yaml.
const metadataExtensionID = "jongio.azd.copilot"

// ExtensionEnvironmentVariables returns the environment variables azd-copilot
// defines: the one it reads as input, and the ones it sets on the GitHub
// Copilot CLI process it launches.
//
// The set variables are the whole reason this list matters. azd-copilot ships
// agents and skills that read them to know which project, subscription and
// infrastructure they are operating against, so these names are a contract
// between this extension and its own prompt content, not incidental plumbing.
// Nothing else published that contract before, which meant a rename in
// launcher.go could silently strand every skill that depended on the old name.
//
// Only variables azd-copilot defines are listed. AZD_SERVER and AZD_DEBUG are
// read but belong to azd, and the AZURE_* values forwarded from the azd
// environment belong to the project, so claiming any of them here would assert
// ownership azd-copilot does not have. TestEveryOwnedEnvVarIsDocumented
// enforces that split against the source tree, so a new AZD_COPILOT_* variable
// fails the build until it is described here.
func ExtensionEnvironmentVariables() []extensions.EnvironmentVariable {
	return []extensions.EnvironmentVariable{
		{
			Name:        "AZD_COPILOT_DEBUG",
			Description: "Set to true to emit debug diagnostics and pass --log-level debug through to the GitHub Copilot CLI. Only the exact value true counts. The --debug flag sets this for the current process.",
			Default:     "false",
			Example:     "true",
		},
		{
			Name:        "AZD_COPILOT_EXTENSION",
			Description: "Set by azd copilot on the GitHub Copilot CLI process. Marks the session as launched by this extension so agents can tell it apart from a bare copilot invocation. Always true when present. Not read by azd copilot.",
			Example:     "true",
		},
		{
			Name:        "AZD_COPILOT_VERSION",
			Description: "Set by azd copilot on the GitHub Copilot CLI process. Version of this extension. Not read by azd copilot.",
			Example:     "0.6.0",
		},
		{
			Name:        "COPILOT_CUSTOM_INSTRUCTIONS_DIRS",
			Description: "Set by azd copilot on the GitHub Copilot CLI process. Comma separated directories holding the agents and skills this extension ships, which is how the Copilot CLI discovers them. Omitted when no directories are added. Not read by azd copilot.",
			Example:     "/home/user/.azd/extensions/jongio.azd.copilot/agents",
		},
		{
			Name:        "AZD_PROJECT_NAME",
			Description: "Set by azd copilot on the GitHub Copilot CLI process. Project name from azure.yaml. Omitted outside a project. Not read by azd copilot.",
			Example:     "myapp",
		},
		{
			Name:        "AZD_PROJECT_PATH",
			Description: "Set by azd copilot on the GitHub Copilot CLI process. Absolute path to the directory containing azure.yaml. Omitted outside a project. Not read by azd copilot.",
			Example:     "/home/user/myapp",
		},
		{
			Name:        "AZD_SERVICES",
			Description: "Set by azd copilot on the GitHub Copilot CLI process. Comma separated service names from azure.yaml. Omitted when the project defines no services. Not read by azd copilot.",
			Example:     "api,web",
		},
		{
			Name:        "AZD_SERVICE_DETAILS",
			Description: "Set by azd copilot on the GitHub Copilot CLI process. Semicolon separated services, each formatted name:language:host:path. Omitted when the project defines no services. Not read by azd copilot.",
			Example:     "api:python:containerapp:./src/api;web:ts:staticwebapp:./src/web",
		},
		{
			Name:        "AZD_SUBSCRIPTION_ID",
			Description: "Set by azd copilot on the GitHub Copilot CLI process. Azure subscription id of the selected azd environment. Omitted when no subscription is selected. Not read by azd copilot.",
			Example:     "00000000-0000-0000-0000-000000000000",
		},
		{
			Name:        "AZD_SUBSCRIPTION_NAME",
			Description: "Set by azd copilot on the GitHub Copilot CLI process. Display name of the subscription. Omitted when unavailable. Not read by azd copilot.",
			Example:     "My Subscription",
		},
		{
			Name:        "AZD_TENANT_ID",
			Description: "Set by azd copilot on the GitHub Copilot CLI process. Microsoft Entra tenant id of the selected environment. Omitted when unavailable. Not read by azd copilot.",
			Example:     "00000000-0000-0000-0000-000000000000",
		},
		{
			Name:        "AZD_USER",
			Description: "Set by azd copilot on the GitHub Copilot CLI process. Signed in azd user. Omitted when not signed in. Not read by azd copilot.",
			Example:     "user@example.com",
		},
		{
			Name:        "AZD_INFRA_PATH",
			Description: "Set by azd copilot on the GitHub Copilot CLI process. Path to the infrastructure directory. Omitted when the project defines none. Not read by azd copilot.",
			Example:     "./infra",
		},
		{
			Name:        "AZD_INFRA_MODULE",
			Description: "Set by azd copilot on the GitHub Copilot CLI process. Name of the entry point infrastructure module. Omitted when the project defines none. Not read by azd copilot.",
			Example:     "main",
		},
		{
			Name:        "AZD_HAS_BICEP",
			Description: "Set by azd copilot on the GitHub Copilot CLI process only when the project contains Bicep. Absent rather than false otherwise, so test for presence. Not read by azd copilot.",
			Example:     "true",
		},
	}
}

// ExtensionConfiguration describes azd-copilot's configuration surface.
//
// Global, Project and Service are deliberately left nil. azd-copilot persists
// no user configuration subtree of its own: the context command reads azd's
// entire user config for display but writes none of it, and the extension
// defines no azure.yaml keys at either project or service scope. State it does
// own lives in files rather than configuration (.copilot.json beside the
// project, checkpoints under the session directory, and the Copilot CLI's own
// mcp-config.json), none of which azd reads. Publishing empty schemas for
// those scopes would claim configuration that does not exist, so the honest
// document carries environment variables alone. TestExtensionConfiguration
// OwnsNoConfigScopes pins that, and fails the day this extension starts
// writing configuration, which is the point at which a schema becomes owed.
func ExtensionConfiguration() *extensions.ConfigurationMetadata {
	return &extensions.ConfigurationMetadata{
		EnvironmentVariables: ExtensionEnvironmentVariables(),
	}
}

// ExtensionMetadata builds the full metadata document for azd-copilot.
//
// azdext.NewMetadataCommand cannot be used directly: it marshals the result of
// GenerateExtensionMetadata immediately and exposes no hook for setting the
// Configuration field, so the command is rebuilt here around the same
// generator.
func ExtensionMetadata(root *cobra.Command) *extensions.ExtensionCommandMetadata {
	metadata := azdext.GenerateExtensionMetadata(metadataSchemaVersion, metadataExtensionID, root)
	metadata.Configuration = ExtensionConfiguration()
	return metadata
}

// NewMetadataCommand creates the hidden metadata command azd uses to discover
// this extension's commands and configuration. rootCmdProvider returns the
// root command for introspection.
func NewMetadataCommand(rootCmdProvider func() *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:    "metadata",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			metadata := ExtensionMetadata(rootCmdProvider())

			jsonBytes, err := json.MarshalIndent(metadata, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal metadata: %w", err)
			}

			if _, err := fmt.Fprintln(cmd.OutOrStdout(), string(jsonBytes)); err != nil {
				return fmt.Errorf("failed to write metadata: %w", err)
			}

			return nil
		},
	}
}
