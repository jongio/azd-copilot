// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package assets

import (
	"io/fs"
	"strings"
	"testing"
)

// TestAllAgents_HaveValidFrontmatter verifies every embedded agent has valid YAML
// frontmatter with required 'name' and 'description' fields.
func TestAllAgents_HaveValidFrontmatter(t *testing.T) {
	agents, err := ListAgents()
	if err != nil {
		t.Fatalf("ListAgents() error = %v", err)
	}

	for _, agent := range agents {
		if agent.Name == "" {
			t.Errorf("Agent from file %q has empty name", agent.FilePath)
		}
		if agent.Description == "" {
			t.Errorf("Agent %q has empty description", agent.Name)
		}
	}
}

// TestAllSkills_HaveValidFrontmatter verifies every embedded skill has valid YAML
// frontmatter with required 'name' and 'description' fields.
func TestAllSkills_HaveValidFrontmatter(t *testing.T) {
	skills, err := ListSkills()
	if err != nil {
		t.Fatalf("ListSkills() error = %v", err)
	}

	for _, skill := range skills {
		if skill.Name == "" {
			t.Errorf("Skill at path %q has empty name", skill.Path)
		}
		if skill.Description == "" {
			t.Errorf("Skill %q has empty description", skill.Name)
		}
	}
}

// TestAllAgents_FrontmatterStartsWithDelimiter verifies all agent markdown files
// start with YAML frontmatter delimiters.
func TestAllAgents_FrontmatterStartsWithDelimiter(t *testing.T) {
	entries, err := fs.ReadDir(embeddedAgents, "agents")
	if err != nil {
		t.Fatalf("Failed to read embedded agents: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		data, err := embeddedAgents.ReadFile("agents/" + entry.Name())
		if err != nil {
			t.Errorf("Failed to read agent %q: %v", entry.Name(), err)
			continue
		}

		content := string(data)
		if !strings.HasPrefix(content, "---") {
			t.Errorf("Agent %q does not start with YAML frontmatter delimiter '---'", entry.Name())
		}
	}
}

// TestAllSkills_FrontmatterStartsWithDelimiter verifies all skill SKILL.md files
// start with YAML frontmatter delimiters.
func TestAllSkills_FrontmatterStartsWithDelimiter(t *testing.T) {
	for _, src := range allSkillSources() {
		entries, err := fs.ReadDir(src.fs, src.prefix)
		if err != nil {
			t.Fatalf("Failed to read skills from %s: %v", src.prefix, err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			skillPath := src.prefix + "/" + entry.Name() + "/SKILL.md"
			data, err := src.fs.ReadFile(skillPath)
			if err != nil {
				continue // Missing SKILL.md tested elsewhere
			}

			content := string(data)
			if !strings.HasPrefix(content, "---") {
				t.Errorf("Skill %q (%s) SKILL.md does not start with YAML frontmatter delimiter '---'",
					entry.Name(), src.prefix)
			}
		}
	}
}

// TestDeploymentConfirmation_AzureManager verifies the azure-manager agent
// requires user confirmation before deploying. This is a regression test to
// prevent re-introducing auto-deploy behavior.
func TestDeploymentConfirmation_AzureManager(t *testing.T) {
	agent, err := GetAgent("azure-manager")
	if err != nil {
		t.Fatalf("GetAgent(azure-manager) error = %v", err)
	}

	data, err := embeddedAgents.ReadFile("agents/" + agent.FilePath)
	if err != nil {
		t.Fatalf("Failed to read azure-manager: %v", err)
	}

	content := string(data)

	// Must NOT contain the old auto-deploy rule
	forbiddenPhrases := []string{
		`NEVER ask "should I deploy this?"`,
		"NEVER ask should I deploy",
	}

	for _, phrase := range forbiddenPhrases {
		if strings.Contains(content, phrase) {
			t.Errorf("azure-manager.md still contains forbidden auto-deploy rule: %q", phrase)
		}
	}

	// Must contain confirmation requirement
	requiredPhrases := []string{
		"ask_user",
		"confirm before deploying",
	}

	for _, phrase := range requiredPhrases {
		if !strings.Contains(content, phrase) {
			t.Errorf("azure-manager.md missing required deployment confirmation phrase: %q", phrase)
		}
	}
}
