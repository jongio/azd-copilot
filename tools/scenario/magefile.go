//go:build mage

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"

	scenario "github.com/jongio/azd-copilot/tools/scenario"
)

var _ mg.Namespace // ensure mage import is used

const (
	scenariosDir = "../../scenarios"
	dbFile       = "results.db"
	jsonFile     = "results.json"
)

func scenarioDir() string {
	return scenariosDir
}

func dbPath() string {
	return filepath.Join(scenariosDir, dbFile)
}

func jsonPath() string {
	return filepath.Join(scenariosDir, jsonFile)
}

type Scenario mg.Namespace

// Extract parses a copilot session log and generates a scenario YAML file.
// Usage: mage scenario:extract <session-id>
func (Scenario) Extract(sessionID string) error {
	fmt.Printf("📋 Extracting scenario from session %s...\n", sessionID)

	s, err := scenario.Extract(sessionID)
	if err != nil {
		return err
	}

	outPath := filepath.Join(scenariosDir, s.Name+".yaml")
	if err := scenario.SaveScenario(s, outPath); err != nil {
		return err
	}

	fmt.Printf("✅ Scenario saved: %s\n", outPath)
	fmt.Printf("   Name: %s\n", s.Name)
	fmt.Printf("   Prompts: %d\n", len(s.Prompts))
	fmt.Printf("   Max duration: %dm\n", s.Scoring.MaxDurationMin)
	return nil
}

// Analyze scores a copilot session against a scenario's criteria.
// Usage: mage scenario:analyze <session-id> <scenario-file>
func (Scenario) Analyze(sessionID, scenarioFile string) error {
	fmt.Printf("📊 Analyzing session %s against %s...\n", sessionID, scenarioFile)

	s, err := scenario.LoadScenario(scenarioFile)
	if err != nil {
		return err
	}

	// Get git commit for tracking
	commit := "unknown"
	if out, gitErr := sh.Output("git", "rev-parse", "HEAD"); gitErr == nil && out != "" {
		commit = out[:min(len(out), 12)]
	}

	run, err := scenario.Analyze(sessionID, s, commit)
	if err != nil {
		return err
	}

	// Save to DB
	db, err := scenario.OpenDB(dbPath())
	if err != nil {
		return err
	}
	defer db.Close()

	id, err := db.InsertRun(run)
	if err != nil {
		return err
	}

	// Print report
	report := scenario.FormatReport(run, s)
	fmt.Println(report)
	fmt.Printf("💾 Saved as run #%d in %s\n", id, dbPath())
	return nil
}

// Run executes a scenario by launching azd copilot with each prompt.
// Usage: mage scenario:run <scenario-file>
func (Scenario) Run(scenarioFile string) error {
	s, err := scenario.LoadScenario(scenarioFile)
	if err != nil {
		return err
	}

	fmt.Printf("🚀 Running scenario: %s\n", s.Name)
	fmt.Printf("   Prompts: %d\n", len(s.Prompts))
	fmt.Printf("   Timeout: %s\n", s.Timeout)

	azdBinary := "azd"
	runResult, err := scenario.RunScenario(context.Background(), s, azdBinary)
	if err != nil {
		return err
	}

	fmt.Printf("\n📊 Run complete. Analyze with:\n")
	fmt.Printf("   mage scenario:analyze %s %s\n", runResult.SessionID, scenarioFile)
	return nil
}

// History shows recent run results for a scenario (or all scenarios).
// Usage: mage scenario:history [scenario-name]
func (Scenario) History(scenarioName string) error {
	db, err := scenario.OpenDB(dbPath())
	if err != nil {
		return err
	}
	defer db.Close()

	runs, err := db.ListRuns(scenarioName, 20)
	if err != nil {
		return err
	}

	if len(runs) == 0 {
		fmt.Println("No runs found.")
		return nil
	}

	fmt.Printf("%-25s %-8s %-6s %-6s %-6s %-8s\n", "SCENARIO", "SCORE", "PASS", "TURNS", "AZD↑", "DURATION")
	fmt.Println(strings.Repeat("-", 70))
	for _, r := range runs {
		status := "✅"
		if !r.Passed {
			status = "❌"
		}
		fmt.Printf("%-25s %5.0f%%  %s    %-6d %-6d %ds\n",
			r.Scenario, r.Score*100, status, r.TotalTurns, r.AzdUpAttempts, r.DurationSec)
	}
	return nil
}

// Dashboard generates the dashboard HTML and serves it via a local HTTP server.
// The dashboard reads results.db live via sql.js — no regeneration needed after new runs.
// Usage: mage scenario:dashboard
func (Scenario) Dashboard() error {
	db, err := scenario.OpenDB(dbPath())
	if err != nil {
		return err
	}
	defer db.Close()

	outPath := filepath.Join(scenariosDir, "dashboard.html")
	if err := scenario.GenerateDashboard(db, outPath); err != nil {
		return err
	}

	fmt.Printf("✅ Dashboard ready at: %s\n", outPath)

	// Serve via local HTTP server (sql.js needs HTTP, not file://)
	absDir, _ := filepath.Abs(scenariosDir)
	fmt.Printf("🌐 Starting server at http://localhost:8086/dashboard.html\n")
	fmt.Println("   Press Ctrl+C to stop")

	// Open in browser
	go func() {
		// Small delay so server starts first
		<-make(chan struct{})
	}()
	if runtime.GOOS == "windows" {
		_ = exec.Command("cmd", "/c", "start", "http://localhost:8086/dashboard.html").Start()
	} else if runtime.GOOS == "darwin" {
		_ = exec.Command("open", "http://localhost:8086/dashboard.html").Start()
	} else {
		_ = exec.Command("xdg-open", "http://localhost:8086/dashboard.html").Start()
	}

	return http.ListenAndServe(":8086", http.FileServer(http.Dir(absDir)))
}

// Regenerate regenerates the dashboard from the existing DB without analyzing new sessions.
func (Scenario) Regenerate() error {
	return Scenario{}.Dashboard()
}

// Loop runs the full improvement loop: run scenario → analyze → fix → rebuild → repeat.
// Runs 3 iterations. Usage: mage scenario:loop <scenario-file>
func (Scenario) Loop(scenarioFile string) error {
	iters := 3

	repoRoot, _ := filepath.Abs(filepath.Join(scenariosDir, ".."))

	cfg := scenario.LoopConfig{
		ScenarioFile: scenarioFile,
		AzdBinary:    "azd",
		MaxIters:     iters,
		RepoRoot:     repoRoot,
		DBPath:       dbPath(),
	}

	results, err := scenario.RunLoop(context.Background(), cfg)
	if err != nil {
		return err
	}

	// Summary
	fmt.Printf("\n%s\n", strings.Repeat("═", 70))
	fmt.Println("  LOOP SUMMARY")
	fmt.Printf("%s\n\n", strings.Repeat("═", 70))
	for _, r := range results {
		status := "❌"
		if r.Run.Passed {
			status = "✅"
		}
		fmt.Printf("  Iteration %d: %s %.0f%% | %d turns | %d azd ups | %ds\n",
			r.Iteration, status, r.Run.Score*100,
			r.Run.TotalTurns, r.Run.AzdUpAttempts, r.Run.DurationSec)
	}

	// Open dashboard
	dashPath, _ := filepath.Abs(filepath.Join(scenariosDir, "dashboard.html"))
	if runtime.GOOS == "windows" {
		_ = exec.Command("cmd", "/c", "start", dashPath).Start()
	} else if runtime.GOOS == "darwin" {
		_ = exec.Command("open", dashPath).Start()
	} else {
		_ = exec.Command("xdg-open", dashPath).Start()
	}

	return nil
}

// Export saves all results from the SQLite database to results.json (committed to git).
// Usage: mage scenario:export
func (Scenario) Export() error {
	db, err := scenario.OpenDB(dbPath())
	if err != nil {
		return err
	}
	defer db.Close()

	jp := jsonPath()
	if err := db.ExportJSON(jp); err != nil {
		return err
	}

	fmt.Printf("✅ Exported results to %s\n", jp)
	return nil
}

// Import loads results from results.json into the SQLite database, skipping duplicates.
// Usage: mage scenario:import
func (Scenario) Import() error {
	jp := jsonPath()
	if _, err := os.Stat(jp); os.IsNotExist(err) {
		return fmt.Errorf("no results file found at %s — nothing to import", jp)
	}

	db, err := scenario.OpenDB(dbPath())
	if err != nil {
		return err
	}
	defer db.Close()

	n, err := db.ImportJSON(jp)
	if err != nil {
		return err
	}

	fmt.Printf("✅ Imported %d new run(s) from %s\n", n, jp)
	return nil
}

func init() {
	// Ensure scenarios directory exists
	os.MkdirAll(scenariosDir, 0755)
}
