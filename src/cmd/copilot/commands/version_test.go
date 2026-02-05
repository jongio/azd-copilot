// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package commands

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestVersionConstants(t *testing.T) {
	// Version should be set (either to "dev" or a version number)
	if Version == "" {
		t.Error("Version should not be empty")
	}

	// BuildTime should be set
	if BuildTime == "" {
		t.Error("BuildTime should not be empty")
	}

	// Commit should be set
	if Commit == "" {
		t.Error("Commit should not be empty")
	}

	t.Logf("Version: %s, BuildTime: %s, Commit: %s", Version, BuildTime, Commit)
}

func TestVersionInfo_Fields(t *testing.T) {
	info := VersionInfo{
		Version:   "1.0.0",
		BuildTime: "2024-01-15T10:30:00Z",
		Commit:    "abc123def",
	}

	if info.Version != "1.0.0" {
		t.Errorf("VersionInfo.Version = %q, want %q", info.Version, "1.0.0")
	}
	if info.BuildTime != "2024-01-15T10:30:00Z" {
		t.Errorf("VersionInfo.BuildTime = %q, want %q", info.BuildTime, "2024-01-15T10:30:00Z")
	}
	if info.Commit != "abc123def" {
		t.Errorf("VersionInfo.Commit = %q, want %q", info.Commit, "abc123def")
	}
}

func TestVersionInfo_JSON(t *testing.T) {
	info := VersionInfo{
		Version:   "1.0.0",
		BuildTime: "2024-01-15T10:30:00Z",
		Commit:    "abc123",
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Verify JSON contains expected fields
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if parsed["version"] != "1.0.0" {
		t.Errorf("JSON version = %v, want %q", parsed["version"], "1.0.0")
	}
	if parsed["buildTime"] != "2024-01-15T10:30:00Z" {
		t.Errorf("JSON buildTime = %v, want %q", parsed["buildTime"], "2024-01-15T10:30:00Z")
	}
	if parsed["commit"] != "abc123" {
		t.Errorf("JSON commit = %v, want %q", parsed["commit"], "abc123")
	}
}

func TestNewVersionCommand(t *testing.T) {
	cmd := NewVersionCommand()

	if cmd == nil {
		t.Fatal("NewVersionCommand() returned nil")
	}

	// Check command metadata
	if cmd.Use != "version" {
		t.Errorf("cmd.Use = %q, want %q", cmd.Use, "version")
	}
	if cmd.Short == "" {
		t.Error("cmd.Short should not be empty")
	}
	if cmd.Long == "" {
		t.Error("cmd.Long should not be empty")
	}

	// Check that --json flag exists
	flag := cmd.Flags().Lookup("json")
	if flag == nil {
		t.Error("--json flag should exist")
	}
}

func TestVersionCommand_DefaultOutput(t *testing.T) {
	cmd := NewVersionCommand()

	// Capture output
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	// Execute command
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("cmd.Execute() error = %v", err)
	}

	// Note: Output verification depends on cliout implementation
	// Just verify no error occurred
}

func TestVersionCommand_JSONOutput(t *testing.T) {
	cmd := NewVersionCommand()

	// Set --json flag
	if err := cmd.Flags().Set("json", "true"); err != nil {
		t.Fatalf("Failed to set flag: %v", err)
	}

	// Capture output
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	// Execute command - note this writes to os.Stdout in the current implementation
	// This test verifies the command doesn't error with --json flag
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("cmd.Execute() with --json error = %v", err)
	}
}

func TestVersionCommand_SilenceUsage(t *testing.T) {
	cmd := NewVersionCommand()

	if !cmd.SilenceUsage {
		t.Error("cmd.SilenceUsage should be true")
	}
}
