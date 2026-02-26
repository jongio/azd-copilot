// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package commands

import (
	"bytes"
	"testing"
)

func TestVersionConstants(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
	if BuildTime == "" {
		t.Error("BuildTime should not be empty")
	}
	if Commit == "" {
		t.Error("Commit should not be empty")
	}
	t.Logf("Version: %s, BuildTime: %s, Commit: %s", Version, BuildTime, Commit)
}

func TestVersionInfo(t *testing.T) {
	if VersionInfo == nil {
		t.Fatal("VersionInfo should not be nil")
	}
	if VersionInfo.ExtensionID != "jongio.azd.copilot" {
		t.Errorf("VersionInfo.ExtensionID = %q, want %q", VersionInfo.ExtensionID, "jongio.azd.copilot")
	}
	if VersionInfo.Name != "azd copilot" {
		t.Errorf("VersionInfo.Name = %q, want %q", VersionInfo.Name, "azd copilot")
	}
	if VersionInfo.Version == "" {
		t.Error("VersionInfo.Version should not be empty")
	}
}

func TestNewVersionCommand(t *testing.T) {
	outputFormat := "default"
	cmd := NewVersionCommand("jongio.azd.copilot", Version, &outputFormat)

	if cmd == nil {
		t.Fatal("NewVersionCommand() returned nil")
	}
	if cmd.Use != "version" {
		t.Errorf("cmd.Use = %q, want %q", cmd.Use, "version")
	}
	if cmd.Short == "" {
		t.Error("cmd.Short should not be empty")
	}
}

func TestVersionCommand_DefaultOutput(t *testing.T) {
	outputFormat := "default"
	cmd := NewVersionCommand("jongio.azd.copilot", Version, &outputFormat)

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("cmd.Execute() error = %v", err)
	}
}
