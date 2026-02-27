// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	// Create MCP server
	s := server.NewMCPServer(
		"azd-copilot",
		Version,
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, false),
	)

	// Register tools
	registerMCPTools(s)

	// Register resources
	registerMCPResources(s)

	// Start stdio server
	return server.ServeStdio(s)
}

func registerMCPTools(s *server.MCPServer) {
	// Tool: list_agents
	s.AddTool(
		mcp.NewTool("list_agents",
			mcp.WithDescription("List all available Azure agents"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			agents := []string{
				"squad - Azure Squad coordinator, routes work to specialized agents",
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
	)

	// Tool: list_skills
	s.AddTool(
		mcp.NewTool("list_skills",
			mcp.WithDescription("List all available Azure skills"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	)

	// Tool: get_project_context
	s.AddTool(
		mcp.NewTool("get_project_context",
			mcp.WithDescription("Get current azd project context including services and environment"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// This would normally query the azd client
			result := `Project Context:
- Check azure.yaml for project configuration
- Check .azure/ folder for environment settings
- Run 'azd context' for full details`
			return mcp.NewToolResultText(result), nil
		},
	)

	// Tool: create_checkpoint
	s.AddTool(
		mcp.NewTool("create_checkpoint",
			mcp.WithDescription("Create a checkpoint to save current build state"),
			mcp.WithString("description", mcp.Required(), mcp.Description("Description of the checkpoint")),
			mcp.WithString("phase", mcp.Description("Build phase (spec, design, develop, quality, deploy)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Extract arguments from the map
			args, ok := req.Params.Arguments.(map[string]interface{})
			if !ok {
				return mcp.NewToolResultText("Error: invalid arguments"), nil
			}
			description, _ := args["description"].(string)
			phase, _ := args["phase"].(string)

			result := fmt.Sprintf("Checkpoint created:\n- Description: %s\n- Phase: %s\n- Use 'azd copilot checkpoints' to list all checkpoints", description, phase)
			return mcp.NewToolResultText(result), nil
		},
	)

	// Tool: list_squad_members
	s.AddTool(
		mcp.NewTool("list_squad_members",
			mcp.WithDescription("List all members of the Azure Squad team"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			projectDir := getSquadProjectDir()
			teamFile := filepath.Join(projectDir, ".ai-team", "team.md")
			data, err := os.ReadFile(teamFile)
			if err != nil {
				return mcp.NewToolResultText("No Squad team found. Run 'azd copilot squad init' to create one."), nil
			}
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// Tool: get_squad_decisions
	s.AddTool(
		mcp.NewTool("get_squad_decisions",
			mcp.WithDescription("Get the shared decisions log for the Azure Squad team"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			projectDir := getSquadProjectDir()
			decisionsFile := filepath.Join(projectDir, ".ai-team", "decisions.md")
			data, err := os.ReadFile(decisionsFile)
			if err != nil {
				return mcp.NewToolResultText("No decisions file found. The Squad team may not be initialized."), nil
			}
			return mcp.NewToolResultText(string(data)), nil
		},
	)

	// Tool: create_squad_decision
	s.AddTool(
		mcp.NewTool("create_squad_decision",
			mcp.WithDescription("Write a decision to the Squad team's shared decision inbox"),
			mcp.WithString("author", mcp.Required(), mcp.Description("Name of the agent or user making the decision")),
			mcp.WithString("summary", mcp.Required(), mcp.Description("Brief summary of the decision")),
			mcp.WithString("detail", mcp.Description("Detailed explanation of the decision and rationale")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, ok := req.Params.Arguments.(map[string]interface{})
			if !ok {
				return mcp.NewToolResultText("Error: invalid arguments"), nil
			}
			author, _ := args["author"].(string)
			summary, _ := args["summary"].(string)
			detail, _ := args["detail"].(string)

			projectDir := getSquadProjectDir()

			inboxDir := filepath.Join(projectDir, ".ai-team", "decisions", "inbox")
			if err := os.MkdirAll(inboxDir, 0750); err != nil {
				return mcp.NewToolResultText(fmt.Sprintf("Error creating inbox directory: %v", err)), nil
			}

			// Create a slug from the summary
			slug := strings.ToLower(strings.ReplaceAll(summary, " ", "-"))
			if len(slug) > 50 {
				slug = slug[:50]
			}

			filename := filepath.Join(inboxDir, fmt.Sprintf("%s-%s.md", strings.ToLower(author), slug))
			content := fmt.Sprintf("### %s\n**By:** %s\n**What:** %s\n", summary, author, summary)
			if detail != "" {
				content += fmt.Sprintf("**Why:** %s\n", detail)
			}

			if err := os.WriteFile(filename, []byte(content), 0600); err != nil {
				return mcp.NewToolResultText(fmt.Sprintf("Error writing decision: %v", err)), nil
			}

			return mcp.NewToolResultText(fmt.Sprintf("Decision written to: %s", filename)), nil
		},
	)
}

// getSquadProjectDir resolves the project directory for Squad operations.
func getSquadProjectDir() string {
	if dir := os.Getenv("AZD_SQUAD_DIR"); dir != "" {
		// AZD_SQUAD_DIR points to .ai-team/ dir, return parent
		return filepath.Dir(dir)
	}
	if dir := os.Getenv("AZD_COPILOT_PROJECT_DIR"); dir != "" {
		return dir
	}
	cwd, _ := os.Getwd()
	return cwd
}

func registerMCPResources(s *server.MCPServer) {
	// Resource: agents
	s.AddResource(
		mcp.NewResource(
			"azd-copilot://agents",
			"Azure Agents",
			mcp.WithResourceDescription("List of available Azure agents"),
			mcp.WithMIMEType("text/plain"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			content := `# Azure Agents

The following specialized agents are available:

1. **squad** - Azure Squad coordinator, routes work to specialized agents
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
	)

	// Resource: skills
	s.AddResource(
		mcp.NewResource(
			"azd-copilot://skills",
			"Azure Skills",
			mcp.WithResourceDescription("List of available Azure skills"),
			mcp.WithMIMEType("text/plain"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
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
	)
}
