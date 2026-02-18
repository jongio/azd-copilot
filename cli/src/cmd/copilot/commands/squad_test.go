// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package commands

import (
	"bytes"
	"testing"
)

func TestNewSquadCommand(t *testing.T) {
	cmd := NewSquadCommand()

	if cmd == nil {
		t.Fatal("NewSquadCommand() returned nil")
	}

	if cmd.Use != "squad" {
		t.Errorf("Command Use = %q, want %q", cmd.Use, "squad")
	}

	if cmd.Short == "" {
		t.Error("Command should have a Short description")
	}
}

func TestNewSquadCommand_Subcommands(t *testing.T) {
	cmd := NewSquadCommand()

	expectedSubcommands := []string{"init", "status", "members"}
	subCmds := cmd.Commands()

	subNames := make(map[string]bool)
	for _, sub := range subCmds {
		subNames[sub.Use] = true
	}

	for _, expected := range expectedSubcommands {
		if !subNames[expected] {
			t.Errorf("Missing subcommand %q", expected)
		}
	}
}

func TestNewSquadCommand_InitHelp(t *testing.T) {
	cmd := NewSquadCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"init", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("init --help failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("init --help produced no output")
	}
}

func TestNewSquadCommand_StatusHelp(t *testing.T) {
	cmd := NewSquadCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"status", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("status --help failed: %v", err)
	}
}

func TestNewSquadCommand_MembersHelp(t *testing.T) {
	cmd := NewSquadCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"members", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("members --help failed: %v", err)
	}
}
