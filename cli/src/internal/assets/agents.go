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

// InstallAgents extracts embedded agents to ~/.azd/copilot/agents/
func InstallAgents() (string, int, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", 0, fmt.Errorf("failed to get home directory: %w", err)
	}

	destDir := filepath.Join(home, ".azd", "copilot", "agents")
	if err := fileutil.EnsureDir(destDir); err != nil {
		return "", 0, fmt.Errorf("failed to create agents directory: %w", err)
	}

	entries, err := fs.ReadDir(embeddedAgents, "agents")
	if err != nil {
		return "", 0, fmt.Errorf("failed to read embedded agents: %w", err)
	}

	installed := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		data, err := embeddedAgents.ReadFile("agents/" + entry.Name())
		if err != nil {
			continue
		}

		destPath := filepath.Join(destDir, entry.Name())
		if err := fileutil.AtomicWriteFile(destPath, data, 0644); err != nil {
			continue
		}
		installed++
	}

	return destDir, installed, nil
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
