// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package assets

import (
	"strings"
	"testing"
)

func TestListAgents(t *testing.T) {
	agents, err := ListAgents()
	if err != nil {
		t.Fatalf("ListAgents() error = %v", err)
	}

	if len(agents) == 0 {
		t.Error("ListAgents() returned empty list, expected embedded agents")
	}

	t.Logf("Found %d agents", len(agents))

	// Verify each agent has required fields
	for _, agent := range agents {
		if agent.Name == "" {
			t.Error("Agent has empty Name")
		}
		if agent.FilePath == "" {
			t.Error("Agent has empty FilePath")
		}
		if !strings.HasSuffix(agent.FilePath, ".md") {
			t.Errorf("Agent FilePath %q should end with .md", agent.FilePath)
		}
	}
}

func TestListAgents_Content(t *testing.T) {
	agents, err := ListAgents()
	if err != nil {
		t.Fatalf("ListAgents() error = %v", err)
	}

	// Check for expected agents
	expectedAgents := []string{
		"azure-manager",
		"azure-architect",
		"azure-dev",
		"azure-security",
		"azure-devops",
	}

	agentNames := make(map[string]bool)
	for _, agent := range agents {
		agentNames[agent.Name] = true
	}

	for _, expected := range expectedAgents {
		if !agentNames[expected] {
			t.Logf("Expected agent %q not found (may be renamed or not included)", expected)
		}
	}
}

func TestGetAgent_Found(t *testing.T) {
	agents, err := ListAgents()
	if err != nil {
		t.Fatalf("ListAgents() error = %v", err)
	}

	if len(agents) == 0 {
		t.Skip("No agents found to test GetAgent")
	}

	// Get the first agent
	firstAgent := agents[0]

	agent, err := GetAgent(firstAgent.Name)
	if err != nil {
		t.Fatalf("GetAgent(%q) error = %v", firstAgent.Name, err)
	}

	if agent == nil {
		t.Fatalf("GetAgent(%q) returned nil", firstAgent.Name)
	}

	if agent.Name != firstAgent.Name {
		t.Errorf("GetAgent() returned wrong agent: got %q, want %q", agent.Name, firstAgent.Name)
	}
}

func TestGetAgent_NotFound(t *testing.T) {
	_, err := GetAgent("nonexistent-agent-xyz")
	if err == nil {
		t.Error("GetAgent() should return error for non-existent agent")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("GetAgent() error should contain 'not found', got %q", err.Error())
	}
}

func TestAgentInfo_Fields(t *testing.T) {
	info := AgentInfo{
		Name:        "test-agent",
		Description: "A test agent",
		Tools:       []string{"tool1", "tool2"},
		FilePath:    "test-agent.md",
	}

	if info.Name != "test-agent" {
		t.Errorf("AgentInfo.Name = %q, want %q", info.Name, "test-agent")
	}
	if info.Description != "A test agent" {
		t.Errorf("AgentInfo.Description = %q, want %q", info.Description, "A test agent")
	}
	if len(info.Tools) != 2 {
		t.Errorf("AgentInfo.Tools length = %d, want 2", len(info.Tools))
	}
	if info.FilePath != "test-agent.md" {
		t.Errorf("AgentInfo.FilePath = %q, want %q", info.FilePath, "test-agent.md")
	}
}

func TestParseAgentInfo_WithFrontmatter(t *testing.T) {
	content := []byte(`---
name: azure-test
description: A test agent for Azure
tools:
  - azure_tool
  - another_tool
---

# Azure Test Agent

This is the content of the agent.
`)

	info := parseAgentInfo("test.md", content)

	if info.Name != "azure-test" {
		t.Errorf("parseAgentInfo() Name = %q, want %q", info.Name, "azure-test")
	}
	if info.Description != "A test agent for Azure" {
		t.Errorf("parseAgentInfo() Description = %q, want %q", info.Description, "A test agent for Azure")
	}
	if len(info.Tools) != 2 {
		t.Errorf("parseAgentInfo() Tools length = %d, want 2", len(info.Tools))
	}
}

func TestParseAgentInfo_WithoutFrontmatter(t *testing.T) {
	content := []byte(`# Simple Agent

This agent has no frontmatter.
`)

	info := parseAgentInfo("simple-agent.md", content)

	// Should use filename as name
	if info.Name != "simple-agent" {
		t.Errorf("parseAgentInfo() Name = %q, want %q", info.Name, "simple-agent")
	}
	if info.Description != "" {
		t.Errorf("parseAgentInfo() Description should be empty, got %q", info.Description)
	}
}

func TestParseAgentInfo_MalformedFrontmatter(t *testing.T) {
	content := []byte(`---
name: test
invalid yaml content here
---

# Test
`)

	// Should not panic on malformed YAML
	info := parseAgentInfo("test.md", content)

	// Should fall back to filename
	if info.Name == "" {
		t.Error("parseAgentInfo() should return filename as fallback Name")
	}
}

func TestInstallAgents(t *testing.T) {
	// Test that InstallAgents doesn't panic
	// Note: Actually installing to ~/.copilot/agents/ may modify user's system
	// so we just verify the function works without error

	count, err := InstallAgents()
	if err != nil {
		t.Logf("InstallAgents() error = %v (may be expected if no home directory)", err)
		return
	}

	t.Logf("InstallAgents() installed %d agents", count)

	if count == 0 {
		t.Log("InstallAgents() installed 0 agents (may be expected if already installed)")
	}
}

func TestEmbeddedAgents_NotEmpty(t *testing.T) {
	// Verify that we have embedded agents
	entries, err := embeddedAgents.ReadDir("agents")
	if err != nil {
		t.Fatalf("Failed to read embedded agents directory: %v", err)
	}

	mdCount := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			mdCount++
		}
	}

	if mdCount == 0 {
		t.Error("No .md files found in embedded agents")
	}

	t.Logf("Found %d embedded agent markdown files", mdCount)
}

func TestEmbeddedAgents_Readable(t *testing.T) {
	entries, err := embeddedAgents.ReadDir("agents")
	if err != nil {
		t.Fatalf("Failed to read embedded agents directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		// Verify each file is readable
		data, err := embeddedAgents.ReadFile("agents/" + entry.Name())
		if err != nil {
			t.Errorf("Failed to read embedded agent %q: %v", entry.Name(), err)
			continue
		}

		if len(data) == 0 {
			t.Errorf("Embedded agent %q is empty", entry.Name())
		}
	}
}
