// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package commands

import (
	"testing"
	"time"
)

func TestFormatAge(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "just now (seconds)",
			duration: 30 * time.Second,
			want:     "just now",
		},
		{
			name:     "zero duration",
			duration: 0,
			want:     "just now",
		},
		{
			name:     "minutes",
			duration: 5 * time.Minute,
			want:     "5m ago",
		},
		{
			name:     "single minute",
			duration: 1 * time.Minute,
			want:     "1m ago",
		},
		{
			name:     "hours",
			duration: 3 * time.Hour,
			want:     "3h ago",
		},
		{
			name:     "single hour",
			duration: 1 * time.Hour,
			want:     "1h ago",
		},
		{
			name:     "one day",
			duration: 24 * time.Hour,
			want:     "1d ago",
		},
		{
			name:     "multiple days",
			duration: 72 * time.Hour,
			want:     "3d ago",
		},
		{
			name:     "just under a minute",
			duration: 59 * time.Second,
			want:     "just now",
		},
		{
			name:     "just under an hour",
			duration: 59 * time.Minute,
			want:     "59m ago",
		},
		{
			name:     "just under a day",
			duration: 23 * time.Hour,
			want:     "23h ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatAge(tt.duration)
			if got != tt.want {
				t.Errorf("formatAge(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

func TestNewSessionsCommand(t *testing.T) {
	cmd := NewSessionsCommand()

	if cmd == nil {
		t.Fatal("NewSessionsCommand() returned nil")
	}
	if cmd.Use != "sessions" {
		t.Errorf("cmd.Use = %q, want %q", cmd.Use, "sessions")
	}
	if cmd.Short == "" {
		t.Error("cmd.Short should not be empty")
	}

	// Check subcommands exist
	subCmds := cmd.Commands()
	if len(subCmds) < 2 {
		t.Errorf("expected at least 2 subcommands (show, delete), got %d", len(subCmds))
	}

	// Verify show subcommand
	showCmd, _, err := cmd.Find([]string{"show"})
	if err != nil {
		t.Errorf("'show' subcommand not found: %v", err)
	}
	if showCmd == nil {
		t.Error("'show' subcommand is nil")
	}

	// Verify delete subcommand
	deleteCmd, _, err := cmd.Find([]string{"delete"})
	if err != nil {
		t.Errorf("'delete' subcommand not found: %v", err)
	}
	if deleteCmd == nil {
		t.Error("'delete' subcommand is nil")
	}

	// Verify limit flag
	flag := cmd.Flags().Lookup("limit")
	if flag == nil {
		t.Error("--limit flag should exist")
	}
	if flag != nil && flag.DefValue != "10" {
		t.Errorf("--limit default = %q, want %q", flag.DefValue, "10")
	}
}
