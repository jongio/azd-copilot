// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package scenario

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	// perPromptTimeout is the max time a single prompt can run before being killed.
	perPromptTimeout = 15 * time.Minute
	// stuckRepeatThreshold is how many consecutive duplicate short lines trigger stuck detection.
	stuckRepeatThreshold = 5
	// stuckLineMaxLen — only lines shorter than this are checked for repeats (long lines are real output).
	stuckLineMaxLen = 20
	// idleTimeout kills the process if no output is produced for this long.
	idleTimeout = 3 * time.Minute
)

// RunScenario executes a scenario by launching azd copilot with each prompt.
// Returns the session ID of the resulting session.
func RunScenario(ctx context.Context, s *Scenario, azdBinary string) (string, error) {
	// Create a fresh temp directory
	tempDir, err := os.MkdirTemp("", "scenario-"+s.Name+"-*")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}
	fmt.Printf("📂 Working directory: %s\n", tempDir)

	// Parse overall scenario timeout
	timeout := 30 * time.Minute
	if s.Timeout != "" {
		if d, err := time.ParseDuration(s.Timeout); err == nil {
			timeout = d
		}
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for i, prompt := range s.Prompts {
		fmt.Printf("\n📝 Prompt %d/%d: %s\n", i+1, len(s.Prompts), truncate(prompt.Text, 80))

		err := runSinglePrompt(ctx, azdBinary, tempDir, prompt.Text, i > 0)
		if err != nil {
			fmt.Printf("⚠️  Prompt %d exited with error: %v\n", i+1, err)
			// Continue to next prompt — partial completion is still useful to analyze
		}

		fmt.Printf("✅ Prompt %d complete\n", i+1)
	}

	// Find the most recent session ID
	sessionID, err := findLatestSession()
	if err != nil {
		return "", fmt.Errorf("find session: %w", err)
	}

	fmt.Printf("\n📊 Session ID: %s\n", sessionID)
	return sessionID, nil
}

// runSinglePrompt runs one azd copilot invocation with completion detection.
// It watches the session's events.jsonl for task_complete tool calls, detects
// stuck output loops, and enforces per-prompt and idle timeouts.
func runSinglePrompt(ctx context.Context, azdBinary, workDir, promptText string, resume bool) error {
	promptCtx, promptCancel := context.WithTimeout(ctx, perPromptTimeout)
	defer promptCancel()

	args := []string{"copilot", "--yolo", "-p", promptText}
	if resume {
		args = append(args, "--resume")
	}

	cmd := exec.CommandContext(promptCtx, azdBinary, args...)
	cmd.Dir = workDir
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start: %w", err)
	}

	// Monitor stdout for stuck loops
	stuckCh := make(chan string, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		reason := monitorOutput(stdout, os.Stdout)
		if reason != "" {
			stuckCh <- reason
		}
	}()

	// Watch events.jsonl for task_complete
	taskDoneCh := make(chan struct{}, 1)
	stopWatching := make(chan struct{})
	go func() {
		watchEventsForCompletion(stopWatching, taskDoneCh)
	}()
	defer close(stopWatching)

	// Wait for: process exit, task_complete event, stuck detection, or timeout
	doneCh := make(chan error, 1)
	go func() {
		doneCh <- cmd.Wait()
	}()

	select {
	case err := <-doneCh:
		wg.Wait()
		return err
	case <-taskDoneCh:
		fmt.Printf("\n✅ task_complete detected in events.jsonl — moving to next prompt\n")
		_ = cmd.Process.Kill()
		<-doneCh
		wg.Wait()
		return nil
	case reason := <-stuckCh:
		fmt.Printf("\n🔄 Stuck loop detected: %s — killing process\n", reason)
		_ = cmd.Process.Kill()
		<-doneCh
		wg.Wait()
		return nil
	case <-promptCtx.Done():
		fmt.Printf("\n⏰ Per-prompt timeout reached — killing process\n")
		_ = cmd.Process.Kill()
		<-doneCh
		wg.Wait()
		return nil
	}
}

// watchEventsForCompletion tails the most recent session's events.jsonl
// and signals on taskDoneCh when a task_complete tool call is detected.
func watchEventsForCompletion(stop <-chan struct{}, taskDoneCh chan<- struct{}) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	sessDir := filepath.Join(home, ".copilot", "session-state")

	// Wait for events.jsonl to appear (session may not exist yet)
	var eventsPath string
	for {
		select {
		case <-stop:
			return
		default:
		}

		// Find the most recently modified session directory
		latest, err := findLatestSessionDir(sessDir)
		if err == nil {
			candidate := filepath.Join(latest, "events.jsonl")
			if _, err := os.Stat(candidate); err == nil {
				eventsPath = candidate
				break
			}
		}
		time.Sleep(2 * time.Second)
	}

	// Tail the file, watching for task_complete
	var lastSize int64
	for {
		select {
		case <-stop:
			return
		default:
		}

		info, err := os.Stat(eventsPath)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		currentSize := info.Size()
		if currentSize > lastSize {
			// Read only the new bytes
			f, err := os.Open(eventsPath)
			if err != nil {
				time.Sleep(1 * time.Second)
				continue
			}
			if lastSize > 0 {
				f.Seek(lastSize, io.SeekStart)
			}
			newData := make([]byte, currentSize-lastSize)
			n, _ := f.Read(newData)
			f.Close()
			lastSize = currentSize

			// Check each new line for task_complete
			for _, line := range strings.Split(string(newData[:n]), "\n") {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				if strings.Contains(line, `"task_complete"`) {
					taskDoneCh <- struct{}{}
					return
				}
			}
		}

		time.Sleep(1 * time.Second)
	}
}

// findLatestSessionDir returns the path of the most recently modified session directory.
func findLatestSessionDir(sessDir string) (string, error) {
	entries, err := os.ReadDir(sessDir)
	if err != nil {
		return "", err
	}

	var latest string
	var latestTime time.Time
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(latestTime) {
			latestTime = info.ModTime()
			latest = filepath.Join(sessDir, e.Name())
		}
	}
	if latest == "" {
		return "", fmt.Errorf("no sessions found")
	}
	return latest, nil
}

// monitorOutput reads from r line-by-line, writes to w, and detects stuck loops.
// Returns a reason string if stuck, empty string if EOF reached normally.
func monitorOutput(r io.Reader, w io.Writer) string {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var lastLine string
	repeatCount := 0
	lastOutput := time.Now()

	// Idle timeout checker runs alongside the scanner
	idleCh := make(chan struct{})
	go func() {
		for {
			time.Sleep(30 * time.Second)
			if time.Since(lastOutput) > idleTimeout {
				close(idleCh)
				return
			}
		}
	}()

	lineCh := make(chan string)
	eofCh := make(chan struct{})
	go func() {
		for scanner.Scan() {
			lineCh <- scanner.Text()
		}
		close(eofCh)
	}()

	for {
		select {
		case line, ok := <-lineCh:
			if !ok {
				return ""
			}
			lastOutput = time.Now()
			fmt.Fprintln(w, line)

			// Check for repeated short lines (stuck loop pattern)
			trimmed := strings.TrimSpace(line)
			if len(trimmed) > 0 && len(trimmed) <= stuckLineMaxLen {
				if trimmed == lastLine {
					repeatCount++
					if repeatCount >= stuckRepeatThreshold {
						return fmt.Sprintf("%q repeated %d times", trimmed, repeatCount+1)
					}
				} else {
					lastLine = trimmed
					repeatCount = 0
				}
			} else {
				// Long line or empty — reset repeat counter
				lastLine = ""
				repeatCount = 0
			}
		case <-eofCh:
			return ""
		case <-idleCh:
			return fmt.Sprintf("no output for %v", idleTimeout)
		}
	}
}

// findLatestSession returns the most recently modified session ID.
func findLatestSession() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	sessDir := filepath.Join(home, ".copilot", "session-state")
	entries, err := os.ReadDir(sessDir)
	if err != nil {
		return "", fmt.Errorf("read session-state dir: %w", err)
	}

	type sessionEntry struct {
		name    string
		modTime time.Time
	}

	var sessions []sessionEntry
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		sessions = append(sessions, sessionEntry{name: e.Name(), modTime: info.ModTime()})
	}

	if len(sessions) == 0 {
		return "", fmt.Errorf("no sessions found")
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].modTime.After(sessions[j].modTime)
	})

	return sessions[0].name, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
