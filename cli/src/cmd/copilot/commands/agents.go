// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package commands

import (
	"fmt"
	"strings"

	"github.com/jongio/azd-copilot/cli/src/internal/assets"
	"github.com/jongio/azd-core/cliout"

	"github.com/spf13/cobra"
)

func NewAgentsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "List and manage Azure agents",
		Long: fmt.Sprintf(`List and manage the %d specialized Azure agents.

Each agent is an expert in a specific domain of Azure development:
- azure-manager: Orchestrates all agents, main entry point
- azure-architect: Infrastructure design, Bicep, networking
- azure-dev: Application code, APIs, frontend
- azure-data: Database schema, queries, migrations
- azure-ai: AI services, RAG, agent frameworks
- azure-security: Security audits, identity, compliance
- azure-devops: CI/CD, deployment, observability
- And more...`, assets.AgentCount()),
		RunE: func(cmd *cobra.Command, args []string) error {
			return listAgents()
		},
	}

	cmd.AddCommand(newAgentsListCommand())
	cmd.AddCommand(newAgentsShowCommand())

	return cmd
}

func newAgentsListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listAgents()
		},
	}
}

func newAgentsShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <agent-name>",
		Short: "Show details about a specific agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return showAgent(args[0])
		},
	}
}

func listAgents() error {
	agents, err := assets.ListAgents()
	if err != nil {
		return fmt.Errorf("failed to list agents: %w", err)
	}

	cliout.Section("🤖", fmt.Sprintf("Available Agents (%d)", len(agents)))
	cliout.Newline()

	// Calculate max name length for alignment
	maxLen := 0
	for _, agent := range agents {
		if len(agent.Name) > maxLen {
			maxLen = len(agent.Name)
		}
	}

	// Print agents
	for _, agent := range agents {
		desc := agent.Description
		if desc == "" {
			desc = "(no description)"
		}
		fmt.Printf("  %s%-*s%s  %s\n", cliout.Cyan, maxLen, agent.Name, cliout.Reset, desc)
	}

	cliout.Newline()
	cliout.Hint("Use 'azd copilot run --agent <name>' to use a specific agent")

	return nil
}

func showAgent(name string) error {
	agent, err := assets.GetAgent(name)
	if err != nil {
		return err
	}

	cliout.Section("🤖", fmt.Sprintf("Agent: %s", agent.Name))
	cliout.Newline()

	if agent.Description != "" {
		cliout.Label("Description", agent.Description)
	}

	if len(agent.Tools) > 0 {
		cliout.Label("Tools", strings.Join(agent.Tools, ", "))
	}

	cliout.Newline()
	cliout.Hint(fmt.Sprintf("Usage: azd copilot run --agent %s", name))

	return nil
}
