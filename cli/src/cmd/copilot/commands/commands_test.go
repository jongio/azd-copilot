// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package commands

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewMCPCommand(t *testing.T) {
	cmd := NewMCPCommand()

	if cmd == nil {
		t.Fatal("NewMCPCommand() returned nil")
	}
	if cmd.Use != "mcp" {
		t.Errorf("cmd.Use = %q, want %q", cmd.Use, "mcp")
	}
	if cmd.Short == "" {
		t.Error("cmd.Short should not be empty")
	}

	// Should have serve and configure subcommands
	subCmds := cmd.Commands()
	if len(subCmds) < 2 {
		t.Errorf("expected at least 2 subcommands, got %d", len(subCmds))
	}

	// Check serve subcommand
	serveCmd, _, err := cmd.Find([]string{"serve"})
	if err != nil {
		t.Errorf("'serve' subcommand not found: %v", err)
	}
	if serveCmd != nil && !serveCmd.Hidden {
		t.Error("'serve' subcommand should be hidden")
	}

	// Check configure subcommand
	configCmd, _, err := cmd.Find([]string{"configure"})
	if err != nil {
		t.Errorf("'configure' subcommand not found: %v", err)
	}
	if configCmd == nil {
		t.Error("'configure' subcommand is nil")
	}
}

func TestNewListenCommand(t *testing.T) {
	cmd := NewListenCommand()

	if cmd == nil {
		t.Fatal("NewListenCommand() returned nil")
	}
	if cmd.Use != "listen" {
		t.Errorf("cmd.Use = %q, want %q", cmd.Use, "listen")
	}
}

func TestNewAgentsCommand(t *testing.T) {
	cmd := NewAgentsCommand()

	if cmd == nil {
		t.Fatal("NewAgentsCommand() returned nil")
	}
	if cmd.Use != "agents" {
		t.Errorf("cmd.Use = %q, want %q", cmd.Use, "agents")
	}
	if cmd.Short == "" {
		t.Error("cmd.Short should not be empty")
	}
}

func TestNewSkillsCommand(t *testing.T) {
	cmd := NewSkillsCommand()

	if cmd == nil {
		t.Fatal("NewSkillsCommand() returned nil")
	}
	if cmd.Use != "skills" {
		t.Errorf("cmd.Use = %q, want %q", cmd.Use, "skills")
	}
	if cmd.Short == "" {
		t.Error("cmd.Short should not be empty")
	}
}

func TestNewContextCommand(t *testing.T) {
	cmd := NewContextCommand()

	if cmd == nil {
		t.Fatal("NewContextCommand() returned nil")
	}
	if cmd.Use != "context" {
		t.Errorf("cmd.Use = %q, want %q", cmd.Use, "context")
	}
	if cmd.Short == "" {
		t.Error("cmd.Short should not be empty")
	}
}

func TestNewCheckpointsCommand(t *testing.T) {
	cmd := NewCheckpointsCommand()

	if cmd == nil {
		t.Fatal("NewCheckpointsCommand() returned nil")
	}
	if cmd.Use != "checkpoints" {
		t.Errorf("cmd.Use = %q, want %q", cmd.Use, "checkpoints")
	}
	if cmd.Short == "" {
		t.Error("cmd.Short should not be empty")
	}
}

func TestNewBuildCommand(t *testing.T) {
	cmd := NewBuildCommand()

	if cmd == nil {
		t.Fatal("NewBuildCommand() returned nil")
	}
	if cmd.Use == "" {
		t.Error("cmd.Use should not be empty")
	}
	if cmd.Short == "" {
		t.Error("cmd.Short should not be empty")
	}
}

func TestNewSpecCommand(t *testing.T) {
	cmd := NewSpecCommand()

	if cmd == nil {
		t.Fatal("NewSpecCommand() returned nil")
	}
	if cmd.Use != "spec" {
		t.Errorf("cmd.Use = %q, want %q", cmd.Use, "spec")
	}
	if cmd.Short == "" {
		t.Error("cmd.Short should not be empty")
	}
}

func TestQuickActionCommands(t *testing.T) {
	tests := []struct {
		name string
		use  string
		cmd  *cobra.Command
	}{
		{"init", "init", NewInitCommand()},
		{"review", "review", NewReviewCommand()},
		{"fix", "fix", NewFixCommand()},
		{"optimize", "optimize", NewOptimizeCommand()},
		{"diagnose", "diagnose", NewDiagnoseCommand()},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.cmd == nil {
				t.Fatalf("New%sCommand() returned nil", tc.name)
			}
			if tc.cmd.Use != tc.use {
				t.Errorf("cmd.Use = %q, want %q", tc.cmd.Use, tc.use)
			}
			if tc.cmd.Short == "" {
				t.Error("cmd.Short should not be empty")
			}
		})
	}
}
