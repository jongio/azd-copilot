// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package commands

import (
	"context"
	"fmt"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/jongio/azd-copilot/cli/src/internal/copilot"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

func NewMCPCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "MCP server management",
		Long:  `Manage Model Context Protocol (MCP) server for AI tool integration.`,
	}

	cmd.AddCommand(newMCPServeCommand())
	cmd.AddCommand(newMCPConfigureCommand())

	return cmd
}

func newMCPServeCommand() *cobra.Command {
	return &cobra.Command{
		Use:    "serve",
		Short:  "Start MCP server (for Copilot CLI integration)",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return serveMCP(cmd.Context())
		},
	}
}

func newMCPConfigureCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "configure",
		Short: "Configure external MCP servers for Copilot CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := copilot.ConfigureMCPServer(); err != nil {
				return fmt.Errorf("failed to configure MCP servers: %w", err)
			}

			fmt.Println("MCP servers configured. Available servers:")
			fmt.Println("  • azure         - Azure resource operations via @azure/mcp")
			fmt.Println("  • azd           - Azure Developer CLI operations")
			fmt.Println("  • microsoft-learn - Microsoft Learn documentation")
			fmt.Println("  • context7      - Context7 memory")
			fmt.Println("  • playwright    - Playwright browser automation and E2E testing")
			fmt.Println()
			fmt.Println("These servers are automatically started by Copilot CLI.")

			return nil
		},
	}
}

// serveMCP starts the MCP server for azd-copilot extension
func serveMCP(ctx context.Context) error {
	builder := azdext.NewMCPServerBuilder("azd-copilot", Version).
		WithRateLimit(10, 1.0).
		WithResourceCapabilities(true, false)

	// Register tools
	registerMCPTools(builder)

	// Register resources
	registerMCPResources(builder)

	// Build and start stdio server
	s := builder.Build()
	return server.ServeStdio(s)
}

func registerMCPTools(builder *azdext.MCPServerBuilder) {
	// Register gRPC service tools (environments, deployments, accounts, workflows, compose)
	registerGRPCTools(builder)

	// Tool: list_agents
	builder.AddTool("list_agents",
		func(ctx context.Context, args azdext.ToolArgs) (*mcp.CallToolResult, error) {
			agents := []string{
				"azure-manager - Orchestrates all agents, main entry point",
				"azure-architect - Infrastructure design, Bicep, networking",
				"azure-dev - Application code, APIs, frontend",
				"azure-data - Database schema, queries, migrations",
				"azure-ai - AI services, RAG, agent frameworks",
				"azure-security - Security audits, identity, compliance",
				"azure-devops - CI/CD, deployment, observability",
			}

			result := "Available Azure Agents:\n"
			for _, agent := range agents {
				result += "• " + agent + "\n"
			}
			return mcp.NewToolResultText(result), nil
		},
		azdext.MCPToolOptions{
			Description: "List all available Azure agents",
			ReadOnly:    true,
		},
	)

	// Tool: list_skills
	builder.AddTool("list_skills",
		func(ctx context.Context, args azdext.ToolArgs) (*mcp.CallToolResult, error) {
			skills := []string{
				"azure-prepare - Initialize project for Azure hosting",
				"azure-deploy - Deployment patterns and best practices",
				"azure-functions - Azure Functions development",
				"azure-security - Security hardening",
				"azure-cost-estimation - Estimate deployment costs",
			}

			result := "Available Azure Skills:\n"
			for _, skill := range skills {
				result += "• " + skill + "\n"
			}
			return mcp.NewToolResultText(result), nil
		},
		azdext.MCPToolOptions{
			Description: "List all available Azure skills",
			ReadOnly:    true,
		},
	)

	// Tool: get_project_context
	builder.AddTool("get_project_context",
		func(ctx context.Context, args azdext.ToolArgs) (*mcp.CallToolResult, error) {
			result := `Project Context:
- Check azure.yaml for project configuration
- Check .azure/ folder for environment settings
- Run 'azd context' for full details`
			return mcp.NewToolResultText(result), nil
		},
		azdext.MCPToolOptions{
			Description: "Get current azd project context including services and environment",
			ReadOnly:    true,
		},
	)

	// Tool: create_checkpoint
	builder.AddTool("create_checkpoint",
		func(ctx context.Context, args azdext.ToolArgs) (*mcp.CallToolResult, error) {
			description, err := args.RequireString("description")
			if err != nil {
				return azdext.MCPErrorResult("missing description: %s", err), nil
			}
			phase := args.OptionalString("phase", "")

			result := fmt.Sprintf("Checkpoint created:\n- Description: %s\n- Phase: %s\n- Use 'azd copilot checkpoints' to list all checkpoints", description, phase)
			return mcp.NewToolResultText(result), nil
		},
		azdext.MCPToolOptions{
			Description: "Create a checkpoint to save current build state",
		},
		mcp.WithString("description", mcp.Required(), mcp.Description("Description of the checkpoint")),
		mcp.WithString("phase", mcp.Description("Build phase (spec, design, develop, quality, deploy)")),
	)
}

func registerMCPResources(builder *azdext.MCPServerBuilder) {
	builder.AddResources(
		server.ServerResource{
			Resource: mcp.NewResource(
				"azd-copilot://agents",
				"Azure Agents",
				mcp.WithResourceDescription("List of available Azure agents"),
				mcp.WithMIMEType("text/plain"),
			),
			Handler: func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
				content := `# Azure Agents

The following specialized agents are available:

1. **azure-manager** - Orchestrates all agents, main entry point
2. **azure-architect** - Infrastructure design, Bicep, networking
3. **azure-dev** - Application code, APIs, frontend
4. **azure-data** - Database schema, queries, migrations
5. **azure-ai** - AI services, RAG, agent frameworks
6. **azure-security** - Security audits, identity, compliance
7. **azure-devops** - CI/CD, deployment, observability
8. **azure-cost** - Cost optimization and estimation
9. **azure-storage** - Storage solutions and data management
10. **azure-networking** - VNets, firewalls, load balancers

Use 'azd copilot agents' for full details.`

				return []mcp.ResourceContents{
					mcp.TextResourceContents{
						URI:      req.Params.URI,
						MIMEType: "text/plain",
						Text:     content,
					},
				}, nil
			},
		},
		server.ServerResource{
			Resource: mcp.NewResource(
				"azd-copilot://skills",
				"Azure Skills",
				mcp.WithResourceDescription("List of available Azure skills"),
				mcp.WithMIMEType("text/plain"),
			),
			Handler: func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
				content := `# Azure Skills

Skills provide focused expertise for specific tasks:

- azure-prepare - Initialize project for Azure hosting
- azure-deploy - Deployment patterns and best practices
- azure-functions - Azure Functions development
- azure-security - Security hardening and best practices
- azure-cost-estimation - Estimate deployment costs
- azure-container-apps - Container Apps configuration
- azure-app-service - App Service setup
- azure-database - Database selection and setup
- azure-ai-services - AI service integration
- azure-monitoring - Observability and alerting

Use 'azd copilot skills' for full details.`

				return []mcp.ResourceContents{
					mcp.TextResourceContents{
						URI:      req.Params.URI,
						MIMEType: "text/plain",
						Text:     content,
					},
				}, nil
			},
		},
	)
}
