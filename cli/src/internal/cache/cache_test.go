// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package cache

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetupCache_NeedsSetup(t *testing.T) {
	tests := []struct {
		name  string
		cache *SetupCache
		want  bool
	}{
		{
			name:  "nil cache needs setup",
			cache: nil,
			want:  true,
		},
		{
			name: "empty cache needs setup",
			cache: &SetupCache{
				AgentsInstalled:   false,
				SkillsInstalled:   false,
				MCPConfigured:     false,
				ExtensionsChecked: false,
			},
			want: true,
		},
		{
			name: "partial setup needs completion",
			cache: &SetupCache{
				AgentsInstalled:   true,
				SkillsInstalled:   true,
				MCPConfigured:     false,
				ExtensionsChecked: true,
			},
			want: true,
		},
		{
			name: "complete setup",
			cache: &SetupCache{
				AgentsInstalled:   true,
				SkillsInstalled:   true,
				MCPConfigured:     true,
				ExtensionsChecked: true,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cache.NeedsSetup()
			if got != tt.want {
				t.Errorf("SetupCache.NeedsSetup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Use a temporary directory to avoid interfering with real cache
	tmpDir := t.TempDir()

	// Override the manager to use temp dir
	oldManager := manager
	defer func() { manager = nil; manager = oldManager }()

	manager = nil
	// We need to set the manager manually for testing
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)

	// Ensure .azd directory exists
	if err := os.MkdirAll(filepath.Join(tmpDir, ".azd"), 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	original := &SetupCache{
		AgentsInstalled:   true,
		SkillsInstalled:   true,
		MCPConfigured:     false,
		ExtensionsChecked: true,
	}

	// Test Save
	err := Save(original)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Test Load
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded == nil {
		t.Fatal("Load() returned nil")
	}

	if loaded.AgentsInstalled != original.AgentsInstalled {
		t.Errorf("AgentsInstalled = %v, want %v", loaded.AgentsInstalled, original.AgentsInstalled)
	}
	if loaded.SkillsInstalled != original.SkillsInstalled {
		t.Errorf("SkillsInstalled = %v, want %v", loaded.SkillsInstalled, original.SkillsInstalled)
	}
	if loaded.MCPConfigured != original.MCPConfigured {
		t.Errorf("MCPConfigured = %v, want %v", loaded.MCPConfigured, original.MCPConfigured)
	}
	if loaded.ExtensionsChecked != original.ExtensionsChecked {
		t.Errorf("ExtensionsChecked = %v, want %v", loaded.ExtensionsChecked, original.ExtensionsChecked)
	}
}

func TestClear(t *testing.T) {
	// Use a temporary directory
	tmpDir := t.TempDir()

	oldManager := manager
	defer func() { manager = nil; manager = oldManager }()

	manager = nil
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)

	if err := os.MkdirAll(filepath.Join(tmpDir, ".azd"), 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Save then clear
	err := Save(&SetupCache{AgentsInstalled: true})
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	err = Clear()
	if err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	// Load should return nil after clear
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded != nil {
		t.Errorf("Load() should return nil after Clear(), got %+v", loaded)
	}
}
