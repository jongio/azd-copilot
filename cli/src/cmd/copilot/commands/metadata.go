// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package commands

import (
	"encoding/json"
	"os"

	"github.com/jongio/azd-copilot/cli/src/internal/assets"

	"github.com/spf13/cobra"
)

// NewMetadataCommand creates the hidden metadata command for IntelliSense support
func NewMetadataCommand() *cobra.Command {
	return &cobra.Command{
		Use:    "metadata",
		Short:  "Output extension metadata for IntelliSense",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return outputMetadata()
		},
	}
}

type extensionMetadata struct {
	Version     string             `json:"version"`
	Commands    []commandMetadata  `json:"commands"`
	Agents      []agentMetadata    `json:"agents"`
	Skills      []skillMetadata    `json:"skills"`
	Flags       []flagMetadata     `json:"flags"`
	Completions []completionOption `json:"completions"`
}

type commandMetadata struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Subcommands []string `json:"subcommands,omitempty"`
	Examples    []string `json:"examples,omitempty"`
}

type agentMetadata struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tools       []string `json:"tools,omitempty"`
}

type skillMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
}

type flagMetadata struct {
	Name        string   `json:"name"`
	Short       string   `json:"short,omitempty"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Default     string   `json:"default,omitempty"`
	Values      []string `json:"values,omitempty"`
}

type completionOption struct {
	Context string   `json:"context"`
	Values  []string `json:"values"`
}

func outputMetadata() error {
	metadata := extensionMetadata{
		Version: Version,
		Commands: []commandMetadata{
			{
				Name:        "build",
				Description: "Generate a complete Azure application from description",
				Subcommands: []string{},
				Examples: []string{
					"azd copilot build \"todo API with PostgreSQL\"",
					"azd copilot build --mode prototype \"chat app demo\"",
					"azd copilot build --approve",
				},
			},
			{
				Name:        "agents",
				Description: "List and manage Azure agents",
				Subcommands: []string{"list", "show"},
			},
			{
				Name:        "skills",
				Description: "List and manage Azure skills",
				Subcommands: []string{"list", "show"},
			},
			{
				Name:        "init",
				Description: "AI-assisted project initialization",
			},
			{
				Name:        "review",
				Description: "AI-powered code review",
			},
			{
				Name:        "fix",
				Description: "AI-powered error fixing",
			},
			{
				Name:        "optimize",
				Description: "Cost and performance optimization",
			},
			{
				Name:        "diagnose",
				Description: "Troubleshoot Azure issues",
			},
			{
				Name:        "checkpoints",
				Description: "Manage build checkpoints",
				Subcommands: []string{"list", "show", "resume", "create", "clear"},
			},
			{
				Name:        "sessions",
				Description: "Manage Copilot sessions",
				Subcommands: []string{"show", "delete"},
			},
			{
				Name:        "spec",
				Description: "View and manage project spec",
				Subcommands: []string{"show", "edit", "delete"},
			},
			{
				Name:        "context",
				Description: "Get azd project context",
			},
			{
				Name:        "mcp",
				Description: "MCP server management",
				Subcommands: []string{"serve", "configure"},
			},
			{
				Name:        "version",
				Description: "Show version information",
			},
		},
		Flags: []flagMetadata{
			{
				Name:        "prompt",
				Short:       "p",
				Description: "Run with a specific prompt",
				Type:        "string",
			},
			{
				Name:        "resume",
				Short:       "r",
				Description: "Resume the last session",
				Type:        "bool",
			},
			{
				Name:        "agent",
				Short:       "a",
				Description: "Use a specific agent",
				Type:        "string",
				Default:     "squad",
				Values:      []string{"squad", "azure-architect", "azure-dev", "azure-data", "azure-ai", "azure-security", "azure-devops"},
			},
			{
				Name:        "model",
				Short:       "m",
				Description: "Use a specific AI model",
				Type:        "string",
				Values:      []string{"claude-sonnet-4", "gpt-4o", "claude-3.5-sonnet"},
			},
			{
				Name:        "yolo",
				Short:       "y",
				Description: "Auto-approve all actions",
				Type:        "bool",
			},
			{
				Name:        "mode",
				Description: "Project mode",
				Type:        "string",
				Values:      []string{"prototype", "production"},
			},
			{
				Name:        "approve",
				Description: "Auto-approve spec",
				Type:        "bool",
			},
		},
		Completions: []completionOption{
			{
				Context: "agent",
				Values:  []string{"squad", "azure-architect", "azure-dev", "azure-data", "azure-ai", "azure-security", "azure-devops"},
			},
			{
				Context: "mode",
				Values:  []string{"prototype", "production"},
			},
			{
				Context: "focus",
				Values:  []string{"security", "performance", "quality", "cost", "all", "both"},
			},
		},
	}

	// Get agents from assets
	agents, err := assets.ListAgents()
	if err == nil {
		for _, agent := range agents {
			metadata.Agents = append(metadata.Agents, agentMetadata{
				Name:        agent.Name,
				Description: agent.Description,
				Tools:       agent.Tools,
			})
		}
	}

	// Get skills from assets
	skills, err := assets.ListSkills()
	if err == nil {
		for _, skill := range skills {
			metadata.Skills = append(metadata.Skills, skillMetadata{
				Name:        skill.Name,
				Description: skill.Description,
				Path:        skill.Path,
			})
		}
	}

	// Output as JSON
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(metadata)
}
