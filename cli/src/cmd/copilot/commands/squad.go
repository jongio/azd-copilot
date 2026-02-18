// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jongio/azd-copilot/cli/src/internal/squad"
	"github.com/jongio/azd-core/cliout"
	"github.com/spf13/cobra"
)

// NewSquadCommand creates the squad command group
func NewSquadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "squad",
		Short: "Manage Azure Squad team",
		Long: `Manage the Azure Squad — a persistent team of AI agents that live in your repo.

The Squad framework provides:
  • Persistent memory across sessions (decisions.md, history.md)
  • Parallel agent fan-out for faster builds
  • Git-native state (.ai-team/ committed to repo)
  • Specialized Azure roles (Architect, Developer, Security, etc.)`,
	}

	cmd.AddCommand(
		newSquadInitCommand(),
		newSquadStatusCommand(),
		newSquadMembersCommand(),
	)

	return cmd
}

func newSquadInitCommand() *cobra.Command {
	var techStack string
	var projectName string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize an Azure Squad team in this project",
		Long: `Create a .ai-team/ directory with Azure-specialized agent charters,
routing configuration, and shared decision log.

This sets up:
  • Agent charters for Architect, Developer, Data Engineer, Security, DevOps, Quality
  • Scribe for session logging and memory management
  • Routing table for Azure work domains
  • Shared decisions.md for persistent team memory`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			// Check if team already exists
			if squad.DetectTeam(cwd) {
				cliout.Warning("Squad team already exists in this project")
				cliout.Hint("Use 'azd copilot squad status' to see team info")
				return nil
			}

			// Get project name from flag or directory name
			if projectName == "" {
				projectName = getProjectName(cwd)
			}

			cliout.Section("🏢", "Initializing Azure Squad")
			cliout.Newline()

			opts := squad.InitOptions{
				ProjectName: projectName,
				TechStack:   techStack,
			}

			if err := squad.InitTeam(cwd, opts); err != nil {
				return fmt.Errorf("failed to initialize squad: %w", err)
			}

			cliout.Success("Azure Squad initialized!")
			cliout.Newline()

			members, err := squad.ListMembers(cwd)
			if err == nil && len(members) > 0 {
				fmt.Println("Team members:")
				for _, m := range members {
					fmt.Printf("  %s %s — %s\n", m.Emoji, m.Name, m.Role)
				}
			}

			cliout.Newline()
			cliout.Hint("Run 'azd copilot' to start working with your squad")
			cliout.Hint("The .ai-team/ directory should be committed to git")

			return nil
		},
	}

	cmd.Flags().StringVar(&techStack, "stack", "", "Tech stack description (e.g., 'Go API with React frontend')")
	cmd.Flags().StringVar(&projectName, "name", "", "Project name (defaults to directory name)")

	return cmd
}

func newSquadStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show Azure Squad team status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			if !squad.DetectTeam(cwd) {
				cliout.Warning("No Squad team found in this project")
				cliout.Hint("Run 'azd copilot squad init' to create one")
				return nil
			}

			cliout.Section("🏢", "Azure Squad Status")
			cliout.Newline()

			members, err := squad.ListMembers(cwd)
			if err != nil {
				return fmt.Errorf("failed to list members: %w", err)
			}

			fmt.Printf("Team: %d members\n", len(members))
			cliout.Newline()
			for _, m := range members {
				fmt.Printf("  %s %-20s %-25s %s\n", m.Emoji, m.Name, m.Role, m.Status)
			}

			cliout.Newline()

			// Show decisions summary
			decisions, err := squad.GetDecisions(cwd)
			if err == nil && decisions != "" {
				lines := 0
				for _, c := range decisions {
					if c == '\n' {
						lines++
					}
				}
				fmt.Printf("Decisions: %d lines in decisions.md\n", lines)
			}

			return nil
		},
	}
}

func newSquadMembersCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "members",
		Short: "List Azure Squad team members",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			if !squad.DetectTeam(cwd) {
				cliout.Warning("No Squad team found in this project")
				cliout.Hint("Run 'azd copilot squad init' to create one")
				return nil
			}

			members, err := squad.ListMembers(cwd)
			if err != nil {
				return fmt.Errorf("failed to list members: %w", err)
			}

			cliout.Section("🏢", "Azure Squad Members")
			cliout.Newline()

			for _, m := range members {
				fmt.Printf("%s %s\n", m.Emoji, m.Name)
				fmt.Printf("  Role: %s\n", m.Role)
				fmt.Printf("  Charter: %s\n", m.CharterPath)
				fmt.Printf("  Status: %s\n", m.Status)
				cliout.Newline()
			}

			return nil
		},
	}
}

// getProjectName extracts the project name from the directory path.
func getProjectName(dir string) string {
	if content, err := os.ReadFile(filepath.Join(dir, "azure.yaml")); err == nil {
		for _, line := range strings.Split(string(content), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "name:") {
				return strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			}
		}
	}
	return filepath.Base(dir)
}
