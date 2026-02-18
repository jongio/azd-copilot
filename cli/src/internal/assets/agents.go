// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package assets

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/jongio/azd-core/fileutil"
	"gopkg.in/yaml.v3"
)

//go:embed agents/*.md
var embeddedAgents embed.FS

// AgentInfo contains metadata about an agent
type AgentInfo struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Tools       []string `yaml:"tools"`
	FilePath    string
}

// InstallAgents extracts embedded agents to ~/.copilot/agents/ (primary, where Copilot CLI reads)
// and ~/.azd/copilot/agents/ (legacy). It also removes stale agent files that no longer exist
// in the embedded set to prevent ghost agents from appearing in the agent list.
func InstallAgents() (string, int, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", 0, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Primary: ~/.copilot/agents/ (where Copilot CLI discovers agents natively)
	primaryDir := filepath.Join(home, ".copilot", "agents")
	// Legacy: ~/.azd/copilot/agents/ (used via --add-dir)
	legacyDir := filepath.Join(home, ".azd", "copilot", "agents")

	for _, destDir := range []string{primaryDir, legacyDir} {
		if err := fileutil.EnsureDir(destDir); err != nil {
			return "", 0, fmt.Errorf("failed to create agents directory %s: %w", destDir, err)
		}
	}

	entries, err := fs.ReadDir(embeddedAgents, "agents")
	if err != nil {
		return "", 0, fmt.Errorf("failed to read embedded agents: %w", err)
	}

	// Build set of embedded agent filenames for stale cleanup
	embeddedNames := make(map[string]bool, len(entries))
	installed := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		embeddedNames[entry.Name()] = true

		data, err := embeddedAgents.ReadFile("agents/" + entry.Name())
		if err != nil {
			continue
		}

		written := false
		for _, destDir := range []string{primaryDir, legacyDir} {
			destPath := filepath.Join(destDir, entry.Name())
			if err := fileutil.AtomicWriteFile(destPath, data, 0644); err == nil {
				written = true
			}
		}
		if written {
			installed++
		}
	}

	// Remove stale agent files that no longer exist in the embedded set
	for _, destDir := range []string{primaryDir, legacyDir} {
		dirEntries, err := os.ReadDir(destDir)
		if err != nil {
			continue
		}
		for _, de := range dirEntries {
			if de.IsDir() || !strings.HasSuffix(de.Name(), ".md") {
				continue
			}
			if !embeddedNames[de.Name()] {
				_ = os.Remove(filepath.Join(destDir, de.Name()))
			}
		}
	}

	return primaryDir, installed, nil
}

// ListAgents returns information about all embedded agents
func ListAgents() ([]AgentInfo, error) {
	entries, err := fs.ReadDir(embeddedAgents, "agents")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded agents: %w", err)
	}

	var agents []AgentInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		data, err := embeddedAgents.ReadFile("agents/" + entry.Name())
		if err != nil {
			continue
		}

		agent := parseAgentInfo(entry.Name(), data)
		agents = append(agents, agent)
	}

	return agents, nil
}

// GetAgent returns information about a specific agent
func GetAgent(name string) (*AgentInfo, error) {
	agents, err := ListAgents()
	if err != nil {
		return nil, err
	}

	for _, agent := range agents {
		if agent.Name == name {
			return &agent, nil
		}
	}

	return nil, fmt.Errorf("agent not found: %s", name)
}

// AgentCount returns the number of available agents.
func AgentCount() int {
	agents, err := ListAgents()
	if err != nil {
		return 0
	}
	return len(agents)
}

// GetAgentContent returns the raw markdown content of an embedded agent by name.
func GetAgentContent(name string) (string, error) {
	data, err := embeddedAgents.ReadFile("agents/" + name + ".md")
	if err != nil {
		return "", fmt.Errorf("agent not found: %s", name)
	}
	return string(data), nil
}

// parseAgentInfo extracts agent info from markdown with YAML frontmatter
func parseAgentInfo(filename string, data []byte) AgentInfo {
	content := string(data)
	agent := AgentInfo{
		Name:     strings.TrimSuffix(filename, ".md"),
		FilePath: filename,
	}

	// Parse YAML frontmatter
	if strings.HasPrefix(content, "---") {
		parts := strings.SplitN(content, "---", 3)
		if len(parts) >= 3 {
			var frontmatter struct {
				Name        string   `yaml:"name"`
				Description string   `yaml:"description"`
				Tools       []string `yaml:"tools"`
			}
			if err := yaml.Unmarshal([]byte(parts[1]), &frontmatter); err == nil {
				if frontmatter.Name != "" {
					agent.Name = frontmatter.Name
				}
				agent.Description = frontmatter.Description
				agent.Tools = frontmatter.Tools
			}
		}
	}

	return agent
}
