// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jongio/azd-core/cliout"
	"github.com/spf13/cobra"
)

func NewSessionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sessions",
		Short: "Manage Copilot sessions",
		Long:  `List, show, and manage GitHub Copilot CLI sessions.`,
		RunE:  listSessions,
	}

	var limit int
	cmd.Flags().IntVarP(&limit, "limit", "n", 10, "Maximum sessions to show")

	cmd.AddCommand(newSessionsShowCommand())
	cmd.AddCommand(newSessionsDeleteCommand())

	return cmd
}

func newSessionsShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <session-id>",
		Short: "Show session details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			return showSession(sessionID)
		},
	}
}

func newSessionsDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <session-id>",
		Short: "Delete a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			return deleteSession(sessionID, force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation")

	return cmd
}

type sessionInfo struct {
	ID      string
	Path    string
	ModTime time.Time
	Agent   string
	HasPlan bool
}

func listSessions(cmd *cobra.Command, args []string) error {
	limit, _ := cmd.Flags().GetInt("limit")

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	sessionsDir := filepath.Join(home, ".copilot", "session-state")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No sessions found.")
			return nil
		}
		return fmt.Errorf("failed to read sessions directory: %w", err)
	}

	var sessions []sessionInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		sessionPath := filepath.Join(sessionsDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		session := sessionInfo{
			ID:      entry.Name(),
			Path:    sessionPath,
			ModTime: info.ModTime(),
		}

		// Check for plan.md
		planPath := filepath.Join(sessionPath, "plan.md")
		if _, err := os.Stat(planPath); err == nil {
			session.HasPlan = true
		}

		// Try to detect agent from checkpoints
		checkpointsDir := filepath.Join(sessionPath, "checkpoints")
		if cpEntries, err := os.ReadDir(checkpointsDir); err == nil && len(cpEntries) > 0 {
			session.Agent = "azure-manager" // Default assumption
		}

		sessions = append(sessions, session)
	}

	// Sort by modification time (newest first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].ModTime.After(sessions[j].ModTime)
	})

	// Limit results
	if limit > 0 && len(sessions) > limit {
		sessions = sessions[:limit]
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions found.")
		return nil
	}

	// Print header
	fmt.Println()
	header := fmt.Sprintf("%-38s  %-15s  %-10s  %s", "SESSION ID", "AGENT", "AGE", "STATUS")
	fmt.Printf("%s%s%s\n", cliout.Bold, header, cliout.Reset)
	fmt.Println(strings.Repeat("─", 80))

	for _, session := range sessions {
		age := formatAge(time.Since(session.ModTime))
		agent := session.Agent
		if agent == "" {
			agent = "-"
		}

		status := "completed"
		if session.HasPlan {
			status = "has plan"
		}

		fmt.Printf("%-38s  %-15s  %-10s  %s\n",
			session.ID,
			agent,
			age,
			status,
		)
	}

	fmt.Println()
	fmt.Printf("Showing %d session(s). Use 'azd copilot sessions show <id>' for details.\n", len(sessions))

	return nil
}

func showSession(sessionID string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	sessionPath := filepath.Join(home, ".copilot", "session-state", sessionID)
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	cliout.Newline()
	cliout.Section("📋", fmt.Sprintf("Session: %s", sessionID))
	cliout.Newline()

	// Show plan if exists
	planPath := filepath.Join(sessionPath, "plan.md")
	if content, err := os.ReadFile(planPath); err == nil {
		cliout.Warning("Plan:")
		fmt.Println(string(content))
	}

	// Show checkpoints
	checkpointsDir := filepath.Join(sessionPath, "checkpoints")
	if entries, err := os.ReadDir(checkpointsDir); err == nil && len(entries) > 0 {
		cliout.Newline()
		cliout.Warning("Checkpoints:")
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".md") {
				fmt.Printf("  • %s\n", strings.TrimSuffix(entry.Name(), ".md"))
			}
		}
	}

	// Show files
	filesDir := filepath.Join(sessionPath, "files")
	if entries, err := os.ReadDir(filesDir); err == nil && len(entries) > 0 {
		cliout.Newline()
		cliout.Warning("Files:")
		for _, entry := range entries {
			fmt.Printf("  • %s\n", entry.Name())
		}
	}

	cliout.Newline()
	cliout.Label("Path", sessionPath)

	return nil
}

func deleteSession(sessionID string, force bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	sessionPath := filepath.Join(home, ".copilot", "session-state", sessionID)
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	if !force {
		fmt.Printf("Delete session %s? [y/N] ", sessionID)
		var response string
		_, _ = fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	if err := os.RemoveAll(sessionPath); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	cliout.Success("Session deleted: %s", sessionID)
	return nil
}

func formatAge(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	days := int(d.Hours() / 24)
	if days == 1 {
		return "1d ago"
	}
	return fmt.Sprintf("%dd ago", days)
}
