// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package scenario

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScenarioYAMLRoundTrip(t *testing.T) {
	s := &Scenario{
		Name:        "test-scenario",
		Description: "A test scenario",
		Timeout:     "15m",
		Prompts: []Prompt{
			{Text: "build a todo app"},
			{Text: "add a database"},
		},
		Scoring: Scoring{
			MaxDurationMin:   10,
			MaxTurns:         20,
			MaxAzdUpAttempts: 3,
			MaxBicepEdits:    4,
			MustDelegate:     true,
			MustInvokeSkills: []string{"avm-bicep-rules"},
			Regressions: []Regression{
				{Name: "ACR auth", Pattern: "ACR.*auth", MaxOccurrences: 2},
			},
		},
	}

	tmp := filepath.Join(t.TempDir(), "scenario.yaml")
	if err := SaveScenario(s, tmp); err != nil {
		t.Fatalf("SaveScenario: %v", err)
	}

	loaded, err := LoadScenario(tmp)
	if err != nil {
		t.Fatalf("LoadScenario: %v", err)
	}

	if loaded.Name != s.Name {
		t.Errorf("Name = %q, want %q", loaded.Name, s.Name)
	}
	if len(loaded.Prompts) != 2 {
		t.Errorf("Prompts count = %d, want 2", len(loaded.Prompts))
	}
	if loaded.Scoring.MaxTurns != 20 {
		t.Errorf("MaxTurns = %d, want 20", loaded.Scoring.MaxTurns)
	}
	if !loaded.Scoring.MustDelegate {
		t.Error("MustDelegate should be true")
	}
	if len(loaded.Scoring.Regressions) != 1 {
		t.Errorf("Regressions count = %d, want 1", len(loaded.Scoring.Regressions))
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"build a todo app", "build-a-todo-app"},
		{"Hello World!!!", "hello-world"},
		{"site that uses dog photo api to show photos of dogs by breed", "site-that-uses-dog-photo-api-to-show-photos-of-dog"},
	}
	for _, tt := range tests {
		got := slugify(tt.in)
		if got != tt.want {
			t.Errorf("slugify(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestDBInsertAndList(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	defer db.Close()

	r := &Run{
		Scenario:      "test-scenario",
		SessionID:     "abc-123",
		DurationSec:   120,
		TotalTurns:    10,
		AzdUpAttempts: 2,
		BicepEdits:    3,
		Delegated:     true,
		Deployed:      true,
		Score:         0.85,
		Passed:        true,
		Skills: map[string]bool{
			"avm-bicep-rules": true,
		},
		Regressions: map[string]RegResult{
			"ACR auth": {Occurrences: 1, MaxAllowed: 2, Passed: true},
		},
	}

	id, err := db.InsertRun(r)
	if err != nil {
		t.Fatalf("InsertRun: %v", err)
	}
	if id < 1 {
		t.Errorf("InsertRun returned id %d, want >= 1", id)
	}

	runs, err := db.ListRuns("test-scenario", 10)
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("ListRuns returned %d runs, want 1", len(runs))
	}
	if runs[0].SessionID != "abc-123" {
		t.Errorf("SessionID = %q, want %q", runs[0].SessionID, "abc-123")
	}
	if runs[0].Score != 0.85 {
		t.Errorf("Score = %f, want 0.85", runs[0].Score)
	}
}

func TestDBListAllScenarios(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	defer db.Close()

	// Insert runs for two different scenarios
	for _, name := range []string{"scenario-a", "scenario-b"} {
		_, err := db.InsertRun(&Run{
			Scenario:  name,
			SessionID: name + "-sess",
			Score:     0.9,
			Skills:    map[string]bool{},
			Regressions: map[string]RegResult{},
		})
		if err != nil {
			t.Fatalf("InsertRun: %v", err)
		}
	}

	// List all (empty scenario name)
	runs, err := db.ListRuns("", 10)
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if len(runs) != 2 {
		t.Errorf("ListRuns('') returned %d runs, want 2", len(runs))
	}
}

func TestFormatReport(t *testing.T) {
	r := &Run{
		Scenario:      "test",
		SessionID:     "abc",
		DurationSec:   300,
		TotalTurns:    15,
		AzdUpAttempts: 2,
		BicepEdits:    3,
		Delegated:     true,
		Score:         0.92,
		Passed:        true,
		Skills:        map[string]bool{"avm-bicep-rules": true},
		Regressions:   map[string]RegResult{"ACR auth": {Occurrences: 0, MaxAllowed: 2, Passed: true}},
	}
	s := &Scenario{
		Name: "test",
		Scoring: Scoring{
			MaxDurationMin:   10,
			MaxTurns:         20,
			MaxAzdUpAttempts: 3,
			MaxBicepEdits:    4,
			MustDelegate:     true,
		},
	}

	report := FormatReport(r, s)
	if report == "" {
		t.Error("FormatReport returned empty string")
	}
	if !contains(report, "PASSED") {
		t.Error("Report should contain PASSED")
	}
	if !contains(report, "92%") {
		t.Errorf("Report should contain 92%%, got:\n%s", report)
	}
}

func TestSessionEventsCountMethods(t *testing.T) {
	// Create a minimal events.jsonl in a temp session
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot get home dir")
	}
	sessID := "test-scenario-" + t.Name()
	sessDir := filepath.Join(home, ".copilot", "session-state", sessID)
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		t.Fatalf("create session dir: %v", err)
	}
	defer os.RemoveAll(sessDir)

	eventsContent := `{"type":"session.start","data":{},"id":"1","timestamp":"2026-01-01T00:00:00Z","parentId":null}
{"type":"user.message","data":{"content":"hello world"},"id":"2","timestamp":"2026-01-01T00:00:01Z","parentId":"1"}
{"type":"assistant.turn_start","data":{"turnId":"0"},"id":"3","timestamp":"2026-01-01T00:00:02Z","parentId":"2"}
{"type":"assistant.message","data":{"content":"I will build it"},"id":"4","timestamp":"2026-01-01T00:00:03Z","parentId":"3"}
{"type":"tool.execution_start","data":{"toolName":"powershell","arguments":{"command":"azd up --no-prompt"}},"id":"5","timestamp":"2026-01-01T00:00:04Z","parentId":"4"}
{"type":"tool.execution_start","data":{"toolName":"edit","arguments":{"path":"infra/main.bicep"}},"id":"6","timestamp":"2026-01-01T00:00:05Z","parentId":"4"}
{"type":"skill.invoked","data":{"name":"avm-bicep-rules"},"id":"7","timestamp":"2026-01-01T00:00:06Z","parentId":"4"}
{"type":"assistant.message","data":{"content":"Done deploying"},"id":"8","timestamp":"2026-01-01T00:00:10Z","parentId":"3"}`

	if err := os.WriteFile(filepath.Join(sessDir, "events.jsonl"), []byte(eventsContent), 0644); err != nil {
		t.Fatalf("write events.jsonl: %v", err)
	}

	se, err := LoadSessionEvents(sessID)
	if err != nil {
		t.Fatalf("LoadSessionEvents: %v", err)
	}

	if msgs := se.UserMessages(); len(msgs) != 1 || msgs[0] != "hello world" {
		t.Errorf("UserMessages = %v, want [hello world]", msgs)
	}
	if n := se.TurnCount(); n != 1 {
		t.Errorf("TurnCount = %d, want 1", n)
	}
	if n := se.CountToolCallsMatching("powershell", "azd up"); n != 1 {
		t.Errorf("CountToolCallsMatching(powershell, azd up) = %d, want 1", n)
	}
	if n := se.CountToolCallsMatching("edit", `main\.bicep`); n != 1 {
		t.Errorf("CountToolCallsMatching(edit, main.bicep) = %d, want 1", n)
	}
	if skills := se.SkillsInvoked(); len(skills) != 1 || skills[0] != "avm-bicep-rules" {
		t.Errorf("SkillsInvoked = %v, want [avm-bicep-rules]", skills)
	}
	if d := se.Duration(); d != 10*1e9 {
		t.Errorf("Duration = %v, want 10s", d)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestListRunsWithDetails(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	defer db.Close()

	_, err = db.InsertRun(&Run{
		Scenario:  "test",
		SessionID: "s1",
		Score:     0.8,
		Skills:    map[string]bool{"avm-bicep-rules": true, "container-app-acr-auth": false},
		Regressions: map[string]RegResult{
			"ACR auth": {Occurrences: 1, MaxAllowed: 2, Passed: true},
		},
	})
	if err != nil {
		t.Fatalf("InsertRun: %v", err)
	}

	runs, err := db.ListRunsWithDetails("test", 10)
	if err != nil {
		t.Fatalf("ListRunsWithDetails: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("got %d runs, want 1", len(runs))
	}
	r := runs[0]
	if len(r.Skills) != 2 {
		t.Errorf("Skills count = %d, want 2", len(r.Skills))
	}
	if !r.Skills["avm-bicep-rules"] {
		t.Error("avm-bicep-rules should be true")
	}
	if r.Skills["container-app-acr-auth"] {
		t.Error("container-app-acr-auth should be false")
	}
	if len(r.Regressions) != 1 {
		t.Errorf("Regressions count = %d, want 1", len(r.Regressions))
	}
	if r.Regressions["ACR auth"].Occurrences != 1 {
		t.Errorf("ACR auth occurrences = %d, want 1", r.Regressions["ACR auth"].Occurrences)
	}
}

func TestGenerateDashboard(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	defer db.Close()

	// Insert two runs for comparison
	for i, sid := range []string{"sess-1", "sess-2"} {
		_, err := db.InsertRun(&Run{
			Scenario:      "test-scenario",
			SessionID:     sid,
			DurationSec:   600 - i*200,
			TotalTurns:    20 - i*5,
			AzdUpAttempts: 5 - i*2,
			BicepEdits:    4 - i,
			Delegated:     i > 0,
			Score:         0.5 + float64(i)*0.3,
			Passed:        i > 0,
			Skills:        map[string]bool{"avm-bicep-rules": true},
			Regressions:   map[string]RegResult{"ACR auth": {Occurrences: 3 - i*2, MaxAllowed: 2, Passed: i > 0}},
		})
		if err != nil {
			t.Fatalf("InsertRun: %v", err)
		}
	}

	outPath := filepath.Join(t.TempDir(), "dashboard.html")
	if err := GenerateDashboard(db, outPath); err != nil {
		t.Fatalf("GenerateDashboard: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read dashboard: %v", err)
	}

	html := string(data)
	if !contains(html, "azd-copilot Scenario Dashboard") {
		t.Error("missing title")
	}
	if !contains(html, "sql.js") {
		t.Error("missing sql.js reference")
	}
	if !contains(html, "results.db") {
		t.Error("missing results.db fetch")
	}
	if !contains(html, "chart-score") {
		t.Error("missing chart elements")
	}
	if !contains(html, "switchTab") {
		t.Error("missing tab switching function")
	}
}
