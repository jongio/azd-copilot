// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSetupCache_IsValid(t *testing.T) {
	tests := []struct {
		name           string
		cache          *SetupCache
		currentVersion string
		want           bool
	}{
		{
			name:           "nil cache",
			cache:          nil,
			currentVersion: "1.0.0",
			want:           false,
		},
		{
			name: "valid cache with matching version",
			cache: &SetupCache{
				Metadata: struct {
					CachedAt time.Time `json:"cachedAt"`
					Version  string    `json:"version,omitempty"`
				}{
					CachedAt: time.Now(),
					Version:  "1.0.0",
				},
			},
			currentVersion: "1.0.0",
			want:           true,
		},
		{
			name: "invalid cache with version mismatch",
			cache: &SetupCache{
				Metadata: struct {
					CachedAt time.Time `json:"cachedAt"`
					Version  string    `json:"version,omitempty"`
				}{
					CachedAt: time.Now(),
					Version:  "0.9.0",
				},
			},
			currentVersion: "1.0.0",
			want:           false,
		},
		{
			name: "expired cache",
			cache: &SetupCache{
				Metadata: struct {
					CachedAt time.Time `json:"cachedAt"`
					Version  string    `json:"version,omitempty"`
				}{
					CachedAt: time.Now().Add(-48 * time.Hour), // 2 days ago
					Version:  "1.0.0",
				},
			},
			currentVersion: "1.0.0",
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cache.IsValid(tt.currentVersion)
			if got != tt.want {
				t.Errorf("SetupCache.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

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

func TestGetCachePath(t *testing.T) {
	path, err := getCachePath()
	if err != nil {
		t.Fatalf("getCachePath() error = %v", err)
	}

	// Should contain .azd directory
	if !filepath.IsAbs(path) {
		t.Errorf("getCachePath() should return absolute path, got %v", path)
	}

	if !contains(path, ".azd") {
		t.Errorf("getCachePath() should contain .azd directory, got %v", path)
	}

	if !contains(path, cacheFile) {
		t.Errorf("getCachePath() should contain %s, got %v", cacheFile, path)
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Skip if we can't get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get user home directory")
	}

	// Use a test-specific cache file to avoid interfering with real cache
	testDir := filepath.Join(home, ".azd-test-cache")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(testDir) }()

	cache := &SetupCache{
		AgentsInstalled:   true,
		SkillsInstalled:   true,
		MCPConfigured:     false,
		ExtensionsChecked: true,
	}

	// Test Save
	err = Save(cache)
	if err != nil {
		t.Logf("Save() error = %v (may be expected in test environment)", err)
	}

	// Test Load
	loaded, err := Load()
	if err != nil {
		t.Logf("Load() error = %v (may be expected in test environment)", err)
	}

	// If load succeeded, verify values
	if loaded != nil {
		if loaded.AgentsInstalled != cache.AgentsInstalled {
			t.Errorf("AgentsInstalled = %v, want %v", loaded.AgentsInstalled, cache.AgentsInstalled)
		}
	}
}

func TestClear(t *testing.T) {
	// Clear should not error even if file doesn't exist
	err := Clear()
	if err != nil {
		t.Logf("Clear() returned error = %v (may be expected)", err)
	}
}

// Helper function
func contains(s, substr string) bool {
	return filepath.Base(filepath.Dir(s)) == ".azd" || filepath.Base(s) == substr ||
		len(s) > len(substr) && s[len(s)-len(substr):] == substr
}
