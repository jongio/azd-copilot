// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package cache

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	corecache "github.com/jongio/azd-core/cache"
)

const (
	cacheKey      = "copilot-setup-cache"
	cacheDuration = 24 * time.Hour
)

// SetupCache stores the results of one-time setup checks
type SetupCache struct {
	AgentsInstalled   bool `json:"agentsInstalled"`
	SkillsInstalled   bool `json:"skillsInstalled"`
	MCPConfigured     bool `json:"mcpConfigured"`
	ExtensionsChecked bool `json:"extensionsChecked"`
}

var (
	manager     *corecache.Manager
	managerOnce sync.Once
	managerErr  error
)

func getManager() (*corecache.Manager, error) {
	managerOnce.Do(func() {
		home, err := os.UserHomeDir()
		if err != nil {
			managerErr = err
			return
		}
		manager = corecache.NewManager(corecache.Options{
			Dir:     filepath.Join(home, ".azd"),
			TTL:     cacheDuration,
			Version: "1",
		})
	})
	return manager, managerErr
}

// Load reads the cache from disk
func Load() (*SetupCache, error) {
	m, err := getManager()
	if err != nil {
		return nil, err
	}
	var c SetupCache
	ok, err := m.Get(cacheKey, &c)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil // Cache doesn't exist or is expired
	}
	return &c, nil
}

// Save writes the cache to disk
func Save(c *SetupCache) error {
	m, err := getManager()
	if err != nil {
		return err
	}
	return m.Set(cacheKey, c)
}

// Clear removes the cache file
func Clear() error {
	m, err := getManager()
	if err != nil {
		return err
	}
	return m.Invalidate(cacheKey)
}

// NeedsSetup returns true if any setup step needs to run
func (c *SetupCache) NeedsSetup() bool {
	if c == nil {
		return true
	}
	return !c.AgentsInstalled || !c.SkillsInstalled || !c.MCPConfigured || !c.ExtensionsChecked
}
