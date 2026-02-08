// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package scenario

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// LoopConfig configures the scenario improvement loop.
type LoopConfig struct {
	ScenarioFile string
	AzdBinary    string
	MaxIters     int
	RepoRoot     string // path to azd-copilot repo root
	DBPath       string
}

// LoopResult holds the outcome of one loop iteration.
type LoopResult struct {
	Iteration int
	SessionID string
	Run       *Run
	Report    string
}

// RunLoop executes the full scenario → analyze → fix → rebuild → repeat loop.
// Returns results from each iteration.
func RunLoop(ctx context.Context, cfg LoopConfig) ([]LoopResult, error) {
	s, err := LoadScenario(cfg.ScenarioFile)
	if err != nil {
		return nil, fmt.Errorf("load scenario: %w", err)
	}

	db, err := OpenDB(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	var results []LoopResult

	for i := 1; i <= cfg.MaxIters; i++ {
		fmt.Printf("\n%s\n", strings.Repeat("═", 70))
		fmt.Printf("  ITERATION %d/%d — %s\n", i, cfg.MaxIters, s.Name)
		fmt.Printf("%s\n\n", strings.Repeat("═", 70))

		// Step 1: Run the scenario
		fmt.Println("▶ Step 1: Running scenario...")
		runResult, err := RunScenario(ctx, s, cfg.AzdBinary)
		if err != nil {
			return results, fmt.Errorf("iteration %d run: %w", i, err)
		}
		sessionID := runResult.SessionID

		// Step 2: Analyze
		fmt.Println("\n▶ Step 2: Analyzing session...")
		commit := gitCommit(cfg.RepoRoot)
		run, err := Analyze(sessionID, s, commit)
		if err != nil {
			return results, fmt.Errorf("iteration %d analyze: %w", i, err)
		}

		if _, err := db.InsertRun(run); err != nil {
			return results, fmt.Errorf("iteration %d save: %w", i, err)
		}

		report := FormatReport(run, s)
		fmt.Println(report)

		result := LoopResult{
			Iteration: i,
			SessionID: sessionID,
			Run:       run,
			Report:    report,
		}
		results = append(results, result)

		// Step 3: Check if passed
		if run.Passed {
			fmt.Printf("🎉 PASSED on iteration %d! Score: %.0f%%\n", i, run.Score*100)
			break
		}

		if i == cfg.MaxIters {
			fmt.Printf("⏹️  Max iterations reached. Best score: %.0f%%\n", run.Score*100)
			break
		}

		// Step 4: Generate fix prompt and send to copilot
		fmt.Println("\n▶ Step 3: Sending analysis to copilot for fixes...")
		fixPrompt := buildFixPrompt(run, s)
		if err := runCopilotFix(ctx, cfg, fixPrompt); err != nil {
			fmt.Printf("⚠️  Fix step failed: %v\n", err)
			fmt.Println("   Continuing to next iteration anyway...")
		}

		// Step 5: Rebuild the extension
		fmt.Println("\n▶ Step 4: Rebuilding extension...")
		if err := rebuildExtension(cfg.RepoRoot); err != nil {
			return results, fmt.Errorf("iteration %d rebuild: %w", i, err)
		}

		fmt.Printf("\n✅ Iteration %d complete. Score: %.0f%% → running again...\n", i, run.Score*100)
	}

	// Generate dashboard
	fmt.Println("\n▶ Generating dashboard...")
	dashPath := filepath.Join(filepath.Dir(cfg.DBPath), "dashboard.html")
	if err := GenerateDashboard(db, dashPath); err != nil {
		fmt.Printf("⚠️  Dashboard generation failed: %v\n", err)
	}

	return results, nil
}

func buildFixPrompt(run *Run, s *Scenario) string {
	var b strings.Builder
	b.WriteString("The scenario test '")
	b.WriteString(s.Name)
	b.WriteString("' failed. Here are the issues to fix in our skills and agents:\n\n")

	if s.Scoring.MaxTurns > 0 && run.TotalTurns > s.Scoring.MaxTurns {
		fmt.Fprintf(&b, "- Too many agent turns: %d (limit: %d). Improve agent efficiency.\n",
			run.TotalTurns, s.Scoring.MaxTurns)
	}
	if s.Scoring.MaxAzdUpAttempts > 0 && run.AzdUpAttempts > s.Scoring.MaxAzdUpAttempts {
		fmt.Fprintf(&b, "- Too many azd up attempts: %d (limit: %d). Fix deployment issues so it succeeds on fewer tries.\n",
			run.AzdUpAttempts, s.Scoring.MaxAzdUpAttempts)
	}
	if s.Scoring.MaxBicepEdits > 0 && run.BicepEdits > s.Scoring.MaxBicepEdits {
		fmt.Fprintf(&b, "- Too many Bicep edits: %d (limit: %d). Get infrastructure right the first time.\n",
			run.BicepEdits, s.Scoring.MaxBicepEdits)
	}
	if s.Scoring.MaxDurationMin > 0 && run.DurationSec > s.Scoring.MaxDurationMin*60 {
		fmt.Fprintf(&b, "- Took too long: %ds (limit: %ds).\n",
			run.DurationSec, s.Scoring.MaxDurationMin*60)
	}
	if s.Scoring.MustDelegate && !run.Delegated {
		b.WriteString("- Agent did not delegate to specialized agents. Use task() to delegate.\n")
	}
	for skill, invoked := range run.Skills {
		if !invoked {
			fmt.Fprintf(&b, "- Required skill '%s' was not invoked.\n", skill)
		}
	}
	for name, reg := range run.Regressions {
		if !reg.Passed {
			fmt.Fprintf(&b, "- Regression '%s': %d occurrences (max: %d). Fix the root cause.\n",
				name, reg.Occurrences, reg.MaxAllowed)
		}
	}

	b.WriteString("\nAnalyze the session log at ~/.copilot/session-state/")
	b.WriteString(run.SessionID)
	b.WriteString("/events.jsonl to understand what went wrong, then update the skills and agents in cli/src/internal/assets/ to fix these issues. ")
	b.WriteString("Only edit files in cli/src/internal/assets/agents/ and cli/src/internal/assets/skills/ (custom skills). ")
	b.WriteString("Do NOT edit files in cli/src/internal/assets/ghcp4a-skills/ (upstream). ")
	b.WriteString("After making changes, run 'cd cli && go build ./... && go test ./...' to verify.")

	return b.String()
}

func runCopilotFix(ctx context.Context, cfg LoopConfig, prompt string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	args := []string{"copilot", "--yolo", "-p", prompt}
	cmd := exec.CommandContext(ctx, cfg.AzdBinary, args...)
	cmd.Dir = cfg.RepoRoot
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// Monitor for stuck loops and task_complete
	stuckCh := make(chan string, 1)
	go func() {
		reason := monitorOutput(stdout, os.Stdout)
		if reason != "" {
			stuckCh <- reason
		}
	}()

	taskDoneCh := make(chan struct{}, 1)
	stopWatching := make(chan struct{})
	go func() {
		watchEventsForCompletion(stopWatching, taskDoneCh)
	}()
	defer close(stopWatching)

	doneCh := make(chan error, 1)
	go func() {
		doneCh <- cmd.Wait()
	}()

	select {
	case err := <-doneCh:
		return err
	case <-taskDoneCh:
		fmt.Printf("\n✅ task_complete detected — fix step done\n")
		_ = cmd.Process.Kill()
		<-doneCh
		return nil
	case reason := <-stuckCh:
		fmt.Printf("\n🔄 Fix step stuck: %s — killing\n", reason)
		_ = cmd.Process.Kill()
		<-doneCh
		return nil
	case <-ctx.Done():
		fmt.Printf("\n⏰ Fix step timeout — killing\n")
		_ = cmd.Process.Kill()
		<-doneCh
		return nil
	}
}

func rebuildExtension(repoRoot string) error {
	cliDir := filepath.Join(repoRoot, "cli")

	// Build
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = cliDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// Test
	cmd = exec.Command("go", "test", "./...")
	cmd.Dir = cliDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}

	// Install via mage
	cmd = exec.Command("mage", "build")
	cmd.Dir = cliDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("⚠️  mage build failed (non-fatal): %v\n", err)
	}

	return nil
}

func gitCommit(repoRoot string) string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	s := strings.TrimSpace(string(out))
	if len(s) > 12 {
		return s[:12]
	}
	return s
}
