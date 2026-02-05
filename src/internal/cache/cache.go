// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package cache

import (
	"os"
	"path/filepath"
	"time"

	"github.com/jongio/azd-core/fileutil"
)

const (
	cacheFile     = "copilot-setup-cache.json"
	cacheDuration = 24 * time.Hour
)

// SetupCache stores the results of one-time setup checks
type SetupCache struct {
	Metadata          fileutil.CacheMetadata `json:"_cache"`
	AgentsInstalled   bool                   `json:"agentsInstalled"`
	SkillsInstalled   bool                   `json:"skillsInstalled"`
	MCPConfigured     bool                   `json:"mcpConfigured"`
	ExtensionsChecked bool                   `json:"extensionsChecked"`
}

// getCachePath returns the path to the cache file
func getCachePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".azd", cacheFile), nil
}

// Load reads the cache from disk
func Load() (*SetupCache, error) {
	path, err := getCachePath()
	if err != nil {
		return nil, err
	}

	var cache SetupCache
	valid, err := fileutil.LoadCacheJSON(path, &cache, fileutil.CacheOptions{
		TTL: cacheDuration,
	})
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, nil // Cache doesn't exist or is expired
	}

	return &cache, nil
}

// Save writes the cache to disk
func Save(cache *SetupCache) error {
	path, err := getCachePath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := fileutil.EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}

	// Update metadata
	cache.Metadata.CachedAt = time.Now()

	return fileutil.AtomicWriteJSON(path, cache)
}

// Clear removes the cache file
func Clear() error {
	path, err := getCachePath()
	if err != nil {
		return err
	}
	return fileutil.ClearCache(path)
}

// IsValid checks if the cache is still valid
func (c *SetupCache) IsValid(currentVersion string) bool {
	if c == nil {
		return false
	}

	// Invalidate if version changed
	if c.Metadata.Version != currentVersion {
		return false
	}

	// Invalidate if cache is too old
	if time.Since(c.Metadata.CachedAt) > cacheDuration {
		return false
	}

	return true
}

// NeedsSetup returns true if any setup step needs to run
func (c *SetupCache) NeedsSetup() bool {
	if c == nil {
		return true
	}
	return !c.AgentsInstalled || !c.SkillsInstalled || !c.MCPConfigured || !c.ExtensionsChecked
}
