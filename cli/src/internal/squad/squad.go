// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package squad

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jongio/azd-copilot/cli/src/internal/assets"
	"github.com/jongio/azd-core/fileutil"
)

// DetectTeam checks if .ai-team/ exists in the given project path by looking
// for a team.md file.
func DetectTeam(projectPath string) bool {
	_, err := os.Stat(filepath.Join(TeamDir(projectPath), "team.md"))
	return err == nil
}

// TeamDir returns the .ai-team/ directory path for the given project.
func TeamDir(projectPath string) string {
	return filepath.Join(projectPath, ".ai-team")
}

// charterRoles maps squad role keys to existing agent names (from assets/agents/).
// The Scribe has no corresponding agent — it uses an embedded template.
var charterRoles = []struct {
	key       string
	agentName string // maps to assets/agents/{agentName}.md; empty means use embedded template
}{
	{"architect", "azure-architect"},
	{"developer", "azure-dev"},
	{"data", "azure-data"},
	{"security", "azure-security"},
	{"devops", "azure-devops"},
	{"quality", "azure-quality"},
	{"ai", "azure-ai"},
	{"analytics", "azure-analytics"},
	{"compliance", "azure-compliance"},
	{"design", "azure-design"},
	{"docs", "azure-docs"},
	{"finance", "azure-finance"},
	{"marketing", "azure-marketing"},
	{"product", "azure-product"},
	{"support", "azure-support"},
	{"scribe", ""}, // Squad-specific role, uses embedded charter-scribe.md
}

// InitTeam creates the .ai-team/ directory structure with Azure-specialized
// agent charters, routing.md, team.md templates.
func InitTeam(projectPath string, opts InitOptions) error {
	td := TeamDir(projectPath)

	// Create top-level directories
	for _, sub := range []string{"", "agents", "decisions", "decisions/inbox", "log"} {
		if err := fileutil.EnsureDir(filepath.Join(td, sub)); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", sub, err)
		}
	}

	// Write team.md with placeholders replaced
	teamContent := replaceTemplatePlaceholders(readTemplate("team.md"), opts)
	if err := fileutil.AtomicWriteFile(filepath.Join(td, "team.md"), []byte(teamContent), 0644); err != nil {
		return fmt.Errorf("failed to write team.md: %w", err)
	}

	// Write routing.md
	if err := fileutil.AtomicWriteFile(filepath.Join(td, "routing.md"), readTemplate("routing.md"), 0644); err != nil {
		return fmt.Errorf("failed to write routing.md: %w", err)
	}

	// Write decisions.md
	if err := fileutil.AtomicWriteFile(filepath.Join(td, "decisions.md"), readTemplate("decisions.md"), 0644); err != nil {
		return fmt.Errorf("failed to write decisions.md: %w", err)
	}

	// Write charter files for each role
	for _, cr := range charterRoles {
		agentDir := filepath.Join(td, "agents", cr.key)
		if err := fileutil.EnsureDir(agentDir); err != nil {
			return fmt.Errorf("failed to create agent directory %s: %w", cr.key, err)
		}

		var data []byte
		if cr.agentName != "" {
			// Read the agent definition from assets/agents/ — no duplication
			agentContent, err := assets.GetAgentContent(cr.agentName)
			if err != nil {
				return fmt.Errorf("failed to read agent %s: %w", cr.agentName, err)
			}
			// Wrap with a Squad collaboration header
			data = []byte(wrapAgentAsCharter(cr.key, agentContent))
		} else {
			// Squad-specific role (Scribe) — use embedded template
			data = readTemplate("charter-scribe.md")
		}

		if err := fileutil.AtomicWriteFile(filepath.Join(agentDir, "charter.md"), data, 0644); err != nil {
			return fmt.Errorf("failed to write charter for %s: %w", cr.key, err)
		}
	}

	return nil
}

// wrapAgentAsCharter wraps an existing agent definition with Squad collaboration
// protocol, avoiding duplication of agent expertise.
func wrapAgentAsCharter(castName string, agentContent string) string {
	collab := fmt.Sprintf(`## Squad Collaboration

Before starting work, read `+"`"+`.ai-team/decisions.md`+"`"+` for team decisions that affect me.
After making a decision, write it to `+"`"+`.ai-team/decisions/inbox/%s-{brief-slug}.md`+"`"+`.
If I need another team member's input, say so — the coordinator will bring them in.

---

`, castName)

	return collab + agentContent
}

// replaceTemplatePlaceholders replaces {{.Field}} tokens in the template.
func replaceTemplatePlaceholders(data []byte, opts InitOptions) string {
	s := string(data)
	s = strings.ReplaceAll(s, "{{.ProjectName}}", opts.ProjectName)
	s = strings.ReplaceAll(s, "{{.TechStack}}", opts.TechStack)
	s = strings.ReplaceAll(s, "{{.UserName}}", opts.UserName)
	s = strings.ReplaceAll(s, "{{.UserEmail}}", opts.UserEmail)
	s = strings.ReplaceAll(s, "{{.CreatedDate}}", time.Now().Format("2006-01-02"))
	return s
}

// ListMembers reads .ai-team/team.md and returns the squad members by parsing
// the Members table.
func ListMembers(projectPath string) ([]Member, error) {
	teamPath := filepath.Join(TeamDir(projectPath), "team.md")
	f, err := os.Open(teamPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open team.md: %w", err)
	}
	defer f.Close()

	var members []Member
	scanner := bufio.NewScanner(f)
	inMembers := false
	headerSkipped := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Detect the Members section
		if line == "## Members" {
			inMembers = true
			headerSkipped = false
			continue
		}

		// Stop at next section
		if inMembers && strings.HasPrefix(line, "## ") {
			break
		}

		if !inMembers {
			continue
		}

		// Skip the table header and separator rows
		if !headerSkipped {
			if strings.HasPrefix(line, "|") && strings.Contains(line, "---") {
				headerSkipped = true
			}
			continue
		}

		// Parse table rows: | Emoji | Name | Role | Charter | Status |
		if !strings.HasPrefix(line, "|") {
			continue
		}

		cols := strings.Split(line, "|")
		// Trim empty first/last from leading/trailing |
		var parts []string
		for _, c := range cols {
			c = strings.TrimSpace(c)
			if c != "" {
				parts = append(parts, c)
			}
		}
		if len(parts) < 5 {
			continue
		}

		status := "active"
		if strings.Contains(parts[4], "Silent") {
			status = "silent"
		} else if strings.Contains(parts[4], "Monitor") {
			status = "monitor"
		}

		members = append(members, Member{
			Emoji:       parts[0],
			Name:        parts[1],
			Role:        parts[2],
			CharterPath: strings.Trim(parts[3], "`"),
			Status:      status,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read team.md: %w", err)
	}

	return members, nil
}

// GetDecisions reads .ai-team/decisions.md content.
func GetDecisions(projectPath string) (string, error) {
	data, err := os.ReadFile(filepath.Join(TeamDir(projectPath), "decisions.md"))
	if err != nil {
		return "", fmt.Errorf("failed to read decisions.md: %w", err)
	}
	return string(data), nil
}

// WriteDecision writes a decision to .ai-team/decisions/inbox/{slug}.md.
func WriteDecision(projectPath string, decision Decision) error {
	inboxDir := filepath.Join(TeamDir(projectPath), "decisions", "inbox")
	if err := fileutil.EnsureDir(inboxDir); err != nil {
		return fmt.Errorf("failed to create inbox directory: %w", err)
	}

	slug := decision.Slug
	if slug == "" {
		slug = strings.ReplaceAll(strings.ToLower(decision.Summary), " ", "-")
	}
	// Sanitize slug to prevent path traversal
	slug = strings.ReplaceAll(slug, "/", "-")
	slug = strings.ReplaceAll(slug, "\\", "-")
	slug = strings.ReplaceAll(slug, "..", "")
	slug = filepath.Base(slug)

	content := fmt.Sprintf("# %s\n\n**By:** %s\n**Date:** %s\n\n## Summary\n%s\n\n## Detail\n%s\n",
		decision.Summary,
		decision.Author,
		time.Now().Format("2006-01-02"),
		decision.Summary,
		decision.Detail,
	)

	destPath := filepath.Join(inboxDir, slug+".md")
	if err := fileutil.AtomicWriteFile(destPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write decision %s: %w", slug, err)
	}

	return nil
}
