// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package squad

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectTeam_NoTeam(t *testing.T) {
	dir := t.TempDir()
	if DetectTeam(dir) {
		t.Error("DetectTeam() should return false for empty directory")
	}
}

func TestDetectTeam_WithTeam(t *testing.T) {
	dir := t.TempDir()
	teamDir := filepath.Join(dir, ".ai-team")
	if err := os.MkdirAll(teamDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(teamDir, "team.md"), []byte("# Team"), 0644); err != nil {
		t.Fatal(err)
	}
	if !DetectTeam(dir) {
		t.Error("DetectTeam() should return true when .ai-team/team.md exists")
	}
}

func TestTeamDir(t *testing.T) {
	got := TeamDir("/project")
	want := filepath.Join("/project", ".ai-team")
	if got != want {
		t.Errorf("TeamDir() = %q, want %q", got, want)
	}
}

func TestInitTeam(t *testing.T) {
	dir := t.TempDir()

	opts := InitOptions{
		ProjectName: "test-project",
		TechStack:   "Go + React",
		UserName:    "Test User",
		UserEmail:   "test@example.com",
	}

	if err := InitTeam(dir, opts); err != nil {
		t.Fatalf("InitTeam() error = %v", err)
	}

	// Verify team.md exists and contains project name
	teamContent, err := os.ReadFile(filepath.Join(dir, ".ai-team", "team.md"))
	if err != nil {
		t.Fatalf("team.md not created: %v", err)
	}
	if !strings.Contains(string(teamContent), "test-project") {
		t.Error("team.md should contain project name")
	}

	// Verify routing.md exists
	if _, err := os.Stat(filepath.Join(dir, ".ai-team", "routing.md")); err != nil {
		t.Error("routing.md not created")
	}

	// Verify decisions.md exists
	if _, err := os.Stat(filepath.Join(dir, ".ai-team", "decisions.md")); err != nil {
		t.Error("decisions.md not created")
	}

	// Verify agent charters exist
	expectedAgents := []string{"architect", "developer", "data", "security", "devops", "quality", "ai", "analytics", "compliance", "design", "docs", "finance", "marketing", "product", "support", "scribe"}
	for _, agent := range expectedAgents {
		charterPath := filepath.Join(dir, ".ai-team", "agents", agent, "charter.md")
		if _, err := os.Stat(charterPath); err != nil {
			t.Errorf("Charter for %s not created at %s", agent, charterPath)
		}
	}

	// Verify decisions/inbox directory exists
	inboxDir := filepath.Join(dir, ".ai-team", "decisions", "inbox")
	info, err := os.Stat(inboxDir)
	if err != nil || !info.IsDir() {
		t.Error("decisions/inbox directory not created")
	}

	// Verify log directory exists
	logDir := filepath.Join(dir, ".ai-team", "log")
	info, err = os.Stat(logDir)
	if err != nil || !info.IsDir() {
		t.Error("log directory not created")
	}
}

func TestInitTeam_DetectAfterInit(t *testing.T) {
	dir := t.TempDir()

	if DetectTeam(dir) {
		t.Error("DetectTeam() should be false before InitTeam")
	}

	opts := InitOptions{ProjectName: "test"}
	if err := InitTeam(dir, opts); err != nil {
		t.Fatalf("InitTeam() error = %v", err)
	}

	if !DetectTeam(dir) {
		t.Error("DetectTeam() should be true after InitTeam")
	}
}

func TestListMembers(t *testing.T) {
	dir := t.TempDir()

	// Initialize team first
	opts := InitOptions{
		ProjectName: "test-project",
		TechStack:   "Go",
	}
	if err := InitTeam(dir, opts); err != nil {
		t.Fatalf("InitTeam() error = %v", err)
	}

	members, err := ListMembers(dir)
	if err != nil {
		t.Fatalf("ListMembers() error = %v", err)
	}

	if len(members) == 0 {
		t.Fatal("ListMembers() returned empty list")
	}

	// Check that we have the expected roles
	roleNames := make(map[string]bool)
	for _, m := range members {
		roleNames[m.Role] = true
	}

	expectedRoles := []string{"Azure Architect", "Azure Developer", "Azure Data Engineer", "Azure Security", "Azure DevOps", "Azure Quality", "Azure AI/ML Engineer", "Azure Analytics", "Azure Compliance", "Azure UX/Accessibility", "Azure Technical Writer", "Azure FinOps", "Azure Marketing", "Azure Product Manager", "Azure Customer Success", "Session Logger"}
	for _, role := range expectedRoles {
		if !roleNames[role] {
			t.Errorf("Expected role %q not found in members", role)
		}
	}

	// Verify Scribe has "silent" status
	for _, m := range members {
		if m.Name == "Scribe" && m.Status != "silent" {
			t.Errorf("Scribe status = %q, want %q", m.Status, "silent")
		}
	}

	// Check that emojis are populated
	for _, m := range members {
		if m.Emoji == "" {
			t.Errorf("Member %q has empty Emoji", m.Name)
		}
	}
}

func TestListMembers_NoTeam(t *testing.T) {
	dir := t.TempDir()
	_, err := ListMembers(dir)
	if err == nil {
		t.Error("ListMembers() should return error when no team exists")
	}
}

func TestGetDecisions(t *testing.T) {
	dir := t.TempDir()

	opts := InitOptions{ProjectName: "test"}
	if err := InitTeam(dir, opts); err != nil {
		t.Fatalf("InitTeam() error = %v", err)
	}

	decisions, err := GetDecisions(dir)
	if err != nil {
		t.Fatalf("GetDecisions() error = %v", err)
	}

	if !strings.Contains(decisions, "Team Decisions") {
		t.Error("GetDecisions() should contain header text")
	}
}

func TestGetDecisions_NoTeam(t *testing.T) {
	dir := t.TempDir()
	_, err := GetDecisions(dir)
	if err == nil {
		t.Error("GetDecisions() should return error when no team exists")
	}
}

func TestWriteDecision(t *testing.T) {
	dir := t.TempDir()

	opts := InitOptions{ProjectName: "test"}
	if err := InitTeam(dir, opts); err != nil {
		t.Fatalf("InitTeam() error = %v", err)
	}

	decision := Decision{
		Author:  "architect",
		Summary: "Use Container Apps",
		Detail:  "Container Apps chosen over AKS for simplicity",
		Slug:    "use-container-apps",
	}

	if err := WriteDecision(dir, decision); err != nil {
		t.Fatalf("WriteDecision() error = %v", err)
	}

	// Verify the decision file was written
	expectedPath := filepath.Join(dir, ".ai-team", "decisions", "inbox", "use-container-apps.md")
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Decision file not created at %s: %v", expectedPath, err)
	}

	if !strings.Contains(string(content), "Use Container Apps") {
		t.Error("Decision file should contain summary")
	}
	if !strings.Contains(string(content), "architect") {
		t.Error("Decision file should contain author")
	}
}

func TestWriteDecision_AutoSlug(t *testing.T) {
	dir := t.TempDir()

	opts := InitOptions{ProjectName: "test"}
	if err := InitTeam(dir, opts); err != nil {
		t.Fatalf("InitTeam() error = %v", err)
	}

	decision := Decision{
		Author:  "developer",
		Summary: "Use Go for backend",
		Detail:  "Go chosen for performance",
	}

	if err := WriteDecision(dir, decision); err != nil {
		t.Fatalf("WriteDecision() error = %v", err)
	}

	// Verify auto-generated slug
	expectedPath := filepath.Join(dir, ".ai-team", "decisions", "inbox", "use-go-for-backend.md")
	if _, err := os.Stat(expectedPath); err != nil {
		t.Errorf("Decision file with auto-slug not created at %s", expectedPath)
	}
}

func TestDefaultAzureRoles(t *testing.T) {
	roles := DefaultAzureRoles()

	if len(roles) != 16 {
		t.Errorf("DefaultAzureRoles() returned %d roles, want 16", len(roles))
	}

	// Verify each role has required fields
	for _, role := range roles {
		if role.Role == "" {
			t.Error("AzureRole has empty Role")
		}
		if role.Emoji == "" {
			t.Error("AzureRole has empty Emoji")
		}
		if role.Expertise == "" {
			t.Error("AzureRole has empty Expertise")
		}
		if len(role.Owns) == 0 {
			t.Errorf("AzureRole %q has no Owns", role.Role)
		}
	}
}

func TestMember_Fields(t *testing.T) {
	m := Member{
		Name:        "architect",
		Role:        "Azure Architect",
		Emoji:       "🏗️",
		CharterPath: ".ai-team/agents/architect/charter.md",
		Status:      "active",
	}

	if m.Name != "architect" {
		t.Errorf("Member.Name = %q, want %q", m.Name, "architect")
	}
	if m.Status != "active" {
		t.Errorf("Member.Status = %q, want %q", m.Status, "active")
	}
}

func TestInitOptions_Fields(t *testing.T) {
	opts := InitOptions{
		ProjectName: "my-app",
		TechStack:   "Go + React",
		UserName:    "developer",
		UserEmail:   "dev@example.com",
	}

	if opts.ProjectName != "my-app" {
		t.Errorf("InitOptions.ProjectName = %q, want %q", opts.ProjectName, "my-app")
	}
}

func TestDecision_Fields(t *testing.T) {
	d := Decision{
		Author:  "security",
		Summary: "Enable managed identity",
		Detail:  "All services should use managed identity instead of connection strings",
		Slug:    "enable-managed-identity",
	}

	if d.Author != "security" {
		t.Errorf("Decision.Author = %q, want %q", d.Author, "security")
	}
	if d.Slug != "enable-managed-identity" {
		t.Errorf("Decision.Slug = %q, want %q", d.Slug, "enable-managed-identity")
	}
}
