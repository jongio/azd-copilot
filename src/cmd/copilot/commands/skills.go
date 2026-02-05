// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package commands

import (
	"fmt"

	"github.com/jongio/azd-copilot/src/internal/assets"
	"github.com/jongio/azd-core/cliout"

	"github.com/spf13/cobra"
)

func NewSkillsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skills",
		Short: "List and manage Azure skills",
		Long: fmt.Sprintf(`List and manage the %d specialized Azure skills.

Skills provide focused expertise for specific Azure tasks:
- azure-prepare: Initialize project for Azure hosting
- azure-deploy: Deployment patterns and best practices
- azure-functions: Azure Functions development
- azure-security: Security hardening and best practices
- azure-cost-estimation: Estimate deployment costs
- And more...`, assets.SkillCount()),
		RunE: func(cmd *cobra.Command, args []string) error {
			return listSkills()
		},
	}

	cmd.AddCommand(newSkillsListCommand())
	cmd.AddCommand(newSkillsShowCommand())

	return cmd
}

func newSkillsListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listSkills()
		},
	}
}

func newSkillsShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <skill-name>",
		Short: "Show details about a specific skill",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return showSkill(args[0])
		},
	}
}

func listSkills() error {
	skills, err := assets.ListSkills()
	if err != nil {
		return fmt.Errorf("failed to list skills: %w", err)
	}

	cliout.Section("⚡", fmt.Sprintf("Available Skills (%d)", len(skills)))
	cliout.Newline()

	// Calculate max name length for alignment
	maxLen := 0
	for _, skill := range skills {
		if len(skill.Name) > maxLen {
			maxLen = len(skill.Name)
		}
	}

	// Print skills
	for _, skill := range skills {
		desc := skill.Description
		if desc == "" {
			desc = "(no description)"
		}
		// Truncate long descriptions
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		fmt.Printf("  %s%-*s%s  %s\n", cliout.Cyan, maxLen, skill.Name, cliout.Reset, desc)
	}

	cliout.Newline()
	cliout.Info("Skills are automatically available during copilot sessions.")

	return nil
}

func showSkill(name string) error {
	skill, err := assets.GetSkill(name)
	if err != nil {
		return err
	}

	cliout.Section("⚡", fmt.Sprintf("Skill: %s", skill.Name))
	cliout.Newline()

	if skill.Description != "" {
		cliout.Label("Description", skill.Description)
	}

	cliout.Label("Path", skill.Path)

	cliout.Newline()
	cliout.Info("Skills are automatically available during copilot sessions.")

	return nil
}
