package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/jongio/azd-copilot/src/internal/cache"
	"github.com/jongio/azd-copilot/src/internal/copilot"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/spf13/cobra"
)

// NewListenCommand creates a new listen command that establishes
// a connection with azd for extension framework operations.
func NewListenCommand() *cobra.Command {
	return &cobra.Command{
		Use:          "listen",
		Short:        "Start the extension server (required by azd framework)",
		Long:         `Internal command used by the azd CLI to communicate with this extension via JSON-RPC over stdio.`,
		Hidden:       true,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create a context with the AZD access token
			ctx := azdext.WithAccessToken(cmd.Context())

			// Create a new AZD client
			azdClient, err := azdext.NewAzdClient()
			if err != nil {
				return fmt.Errorf("failed to create azd client: %w", err)
			}
			defer azdClient.Close()

			// Create an extension host with lifecycle event handlers
			host := azdext.NewExtensionHost(azdClient).
				// Project-level events
				WithProjectEventHandler("preinit", handlePreInit).
				WithProjectEventHandler("preprovision", handlePreProvision).
				WithProjectEventHandler("postprovision", handlePostProvision).
				// Service-level events
				WithServiceEventHandler("predeploy", handlePreDeploy, &azdext.ServiceEventOptions{}).
				WithServiceEventHandler("postdeploy", handlePostDeploy, &azdext.ServiceEventOptions{})

			// Start the extension host
			// This blocks until azd closes the connection
			if err := host.Run(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "Extension host error: %v\n", err)
				return fmt.Errorf("failed to run extension: %w", err)
			}

			return nil
		},
	}
}

// handlePreInit is called before azd init completes
func handlePreInit(ctx context.Context, args *azdext.ProjectEventArgs) error {
	// Clear cached setup state when initializing a new project
	if err := cache.Clear(); err != nil {
		// Non-fatal, just log
		fmt.Fprintf(os.Stderr, "Warning: failed to clear cache: %v\n", err)
	}
	return nil
}

// handlePreProvision is called before azd provision starts
func handlePreProvision(ctx context.Context, args *azdext.ProjectEventArgs) error {
	// Ensure MCP servers are configured for Copilot CLI
	if err := copilot.ConfigureMCPServer(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to configure MCP servers: %v\n", err)
	}
	return nil
}

// handlePostProvision is called after azd provision completes
func handlePostProvision(ctx context.Context, args *azdext.ProjectEventArgs) error {
	// Could save provisioned resource info for checkpoints
	return nil
}

// handlePreDeploy is called before deploying a service
func handlePreDeploy(ctx context.Context, args *azdext.ServiceEventArgs) error {
	// Could run pre-deploy checks specific to the service
	return nil
}

// handlePostDeploy is called after deploying a service
func handlePostDeploy(ctx context.Context, args *azdext.ServiceEventArgs) error {
	// Could create a deployment checkpoint
	return nil
}
