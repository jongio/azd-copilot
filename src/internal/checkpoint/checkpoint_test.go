// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package checkpoint

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPhaseConstants(t *testing.T) {
	// Verify phase constants are defined correctly
	phases := []Phase{PhaseSpec, PhaseDesign, PhaseDevelop, PhaseQuality, PhaseDeploy}
	expected := []string{"spec", "design", "develop", "quality", "deploy"}

	for i, phase := range phases {
		if string(phase) != expected[i] {
			t.Errorf("Phase %d = %q, want %q", i, phase, expected[i])
		}
	}
}

func TestCheckpointTypeConstants(t *testing.T) {
	types := []CheckpointType{TypePhase, TypeTask, TypeSnapshot, TypeRecovery, TypeManual}
	expected := []string{"phase", "task", "snapshot", "recovery", "manual"}

	for i, cpType := range types {
		if string(cpType) != expected[i] {
			t.Errorf("CheckpointType %d = %q, want %q", i, cpType, expected[i])
		}
	}
}

func TestTriggerConstants(t *testing.T) {
	triggers := []Trigger{
		TriggerPhaseCompleted,
		TriggerTaskCompleted,
		TriggerUserInterjection,
		TriggerBeforeDeployment,
		TriggerBeforeDestructive,
		TriggerPeriodic,
		TriggerErrorRecovery,
		TriggerManual,
	}

	for _, trigger := range triggers {
		if trigger == "" {
			t.Error("Trigger should not be empty")
		}
	}
}

func TestNextPhase(t *testing.T) {
	tests := []struct {
		input Phase
		want  Phase
	}{
		{PhaseSpec, PhaseDesign},
		{PhaseDesign, PhaseDevelop},
		{PhaseDevelop, PhaseQuality},
		{PhaseQuality, PhaseDeploy},
		{PhaseDeploy, PhaseDeploy}, // Terminal
		{"unknown", PhaseDevelop},  // Default
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			got := NextPhase(tt.input)
			if got != tt.want {
				t.Errorf("NextPhase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestComputeFileHashes(t *testing.T) {
	// Create a temp file for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	content := []byte("test content for hashing")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	hashes := computeFileHashes([]string{testFile})

	if len(hashes) != 1 {
		t.Errorf("computeFileHashes() returned %d hashes, want 1", len(hashes))
	}

	hash, ok := hashes[testFile]
	if !ok {
		t.Error("computeFileHashes() did not include test file")
	}

	// SHA256 hash should be 64 hex characters
	if len(hash) != 64 {
		t.Errorf("Hash length = %d, want 64", len(hash))
	}
}

func TestComputeFileHashes_NonExistent(t *testing.T) {
	hashes := computeFileHashes([]string{"/nonexistent/file.txt"})

	// Should return empty map for non-existent files (no error)
	if len(hashes) != 0 {
		t.Errorf("computeFileHashes() for non-existent file returned %d hashes, want 0", len(hashes))
	}
}

func TestHashFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "hash_test.txt")

	content := []byte("hello world")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	hash, err := hashFile(testFile)
	if err != nil {
		t.Fatalf("hashFile() error = %v", err)
	}

	// Verify hash format (64 hex chars)
	if len(hash) != 64 {
		t.Errorf("hashFile() returned hash of length %d, want 64", len(hash))
	}

	// Known SHA256 of "hello world"
	expectedHash := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if hash != expectedHash {
		t.Errorf("hashFile() = %q, want %q", hash, expectedHash)
	}
}

func TestHashFile_NonExistent(t *testing.T) {
	_, err := hashFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("hashFile() should error for non-existent file")
	}
}

func TestCheckpoint_Fields(t *testing.T) {
	cp := Checkpoint{
		ID:              "test-id-123",
		ProjectID:       "project-1",
		SessionID:       "session-1",
		Type:            TypePhase,
		Trigger:         TriggerPhaseCompleted,
		Phase:           PhaseDevelop,
		CompletedPhases: []Phase{PhaseSpec, PhaseDesign},
		Description:     "Test checkpoint",
		CreatedAt:       time.Now(),
		CanResume:       true,
	}

	if cp.ID != "test-id-123" {
		t.Errorf("Checkpoint.ID = %q, want %q", cp.ID, "test-id-123")
	}
	if cp.Type != TypePhase {
		t.Errorf("Checkpoint.Type = %q, want %q", cp.Type, TypePhase)
	}
	if cp.Phase != PhaseDevelop {
		t.Errorf("Checkpoint.Phase = %q, want %q", cp.Phase, PhaseDevelop)
	}
	if len(cp.CompletedPhases) != 2 {
		t.Errorf("Checkpoint.CompletedPhases length = %d, want 2", len(cp.CompletedPhases))
	}
	if !cp.CanResume {
		t.Error("Checkpoint.CanResume should be true")
	}
}

func TestFileState_Fields(t *testing.T) {
	fs := FileState{
		Created:  []string{"file1.go", "file2.go"},
		Modified: []string{"file3.go"},
		Deleted:  []string{"old.go"},
		Hashes:   map[string]string{"file1.go": "abc123"},
	}

	if len(fs.Created) != 2 {
		t.Errorf("FileState.Created length = %d, want 2", len(fs.Created))
	}
	if len(fs.Modified) != 1 {
		t.Errorf("FileState.Modified length = %d, want 1", len(fs.Modified))
	}
	if len(fs.Deleted) != 1 {
		t.Errorf("FileState.Deleted length = %d, want 1", len(fs.Deleted))
	}
	if fs.Hashes["file1.go"] != "abc123" {
		t.Errorf("FileState.Hashes[file1.go] = %q, want %q", fs.Hashes["file1.go"], "abc123")
	}
}

func TestTaskState_Fields(t *testing.T) {
	ts := TaskState{
		CompletedTasks: []string{"task1", "task2"},
		PendingTasks:   []string{"task3"},
		FailedTasks: []TaskFailure{
			{
				Task:      "task4",
				Error:     "some error",
				Retries:   2,
				Timestamp: time.Now(),
			},
		},
	}

	if len(ts.CompletedTasks) != 2 {
		t.Errorf("TaskState.CompletedTasks length = %d, want 2", len(ts.CompletedTasks))
	}
	if len(ts.PendingTasks) != 1 {
		t.Errorf("TaskState.PendingTasks length = %d, want 1", len(ts.PendingTasks))
	}
	if len(ts.FailedTasks) != 1 {
		t.Errorf("TaskState.FailedTasks length = %d, want 1", len(ts.FailedTasks))
	}
	if ts.FailedTasks[0].Retries != 2 {
		t.Errorf("TaskFailure.Retries = %d, want 2", ts.FailedTasks[0].Retries)
	}
}

func TestContext_Fields(t *testing.T) {
	ctx := Context{
		SpecHash:          "hash123",
		LastPrompt:        "build the app",
		LastAgentResponse: "I'll help you build...",
		SessionID:         "sess-123",
		ErrorMessage:      "",
		ErrorStack:        "",
	}

	if ctx.SpecHash != "hash123" {
		t.Errorf("Context.SpecHash = %q, want %q", ctx.SpecHash, "hash123")
	}
	if ctx.LastPrompt != "build the app" {
		t.Errorf("Context.LastPrompt = %q, want %q", ctx.LastPrompt, "build the app")
	}
	if ctx.SessionID != "sess-123" {
		t.Errorf("Context.SessionID = %q, want %q", ctx.SessionID, "sess-123")
	}
}

func TestGenerateResumePrompt(t *testing.T) {
	cp := &Checkpoint{
		ID:              "test-checkpoint-1",
		Type:            TypePhase,
		Phase:           PhaseDevelop,
		CompletedPhases: []Phase{PhaseSpec, PhaseDesign},
		Description:     "Completed develop phase",
		CreatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Files: FileState{
			Created: []string{"src/main.go", "src/handler.go"},
		},
	}

	prompt := GenerateResumePrompt(cp)

	// Verify key sections are present
	if !contains(prompt, "Resume Build from Checkpoint") {
		t.Error("Prompt should contain 'Resume Build from Checkpoint' header")
	}
	if !contains(prompt, "test-checkpoint-1") {
		t.Error("Prompt should contain checkpoint ID")
	}
	if !contains(prompt, "develop") {
		t.Error("Prompt should contain phase name")
	}
	if !contains(prompt, "quality") {
		t.Error("Prompt should reference next phase (quality)")
	}
	if !contains(prompt, "src/main.go") {
		t.Error("Prompt should list created files")
	}
}

func TestGenerateResumePrompt_RecoveryCheckpoint(t *testing.T) {
	cp := &Checkpoint{
		ID:          "recovery-1",
		Type:        TypeRecovery,
		Phase:       PhaseDevelop,
		Description: "Recovery from error",
		CreatedAt:   time.Now(),
		Context: Context{
			ErrorMessage: "compilation failed",
		},
	}

	prompt := GenerateResumePrompt(cp)

	if !contains(prompt, "recovery checkpoint") {
		t.Error("Prompt should mention recovery checkpoint")
	}
	if !contains(prompt, "compilation failed") {
		t.Error("Prompt should include error message")
	}
}

func TestGetPhaseGuidance(t *testing.T) {
	tests := []struct {
		phase    Phase
		contains string
	}{
		{PhaseDesign, "Design Phase Tasks"},
		{PhaseDevelop, "Develop Phase Tasks"},
		{PhaseQuality, "Quality Phase Tasks"},
		{PhaseDeploy, "Deploy Phase Tasks"},
		{PhaseSpec, ""}, // No guidance for spec phase
	}

	for _, tt := range tests {
		t.Run(string(tt.phase), func(t *testing.T) {
			guidance := getPhaseGuidance(tt.phase)
			if tt.contains != "" && !contains(guidance, tt.contains) {
				t.Errorf("getPhaseGuidance(%q) should contain %q", tt.phase, tt.contains)
			}
		})
	}
}

func TestSaveOptions_Fields(t *testing.T) {
	opts := SaveOptions{
		Phase:           PhaseDevelop,
		Type:            TypeTask,
		Trigger:         TriggerTaskCompleted,
		Description:     "Task completed",
		ProjectID:       "proj-1",
		SessionID:       "sess-1",
		CompletedPhases: []Phase{PhaseSpec},
	}

	if opts.Phase != PhaseDevelop {
		t.Errorf("SaveOptions.Phase = %q, want %q", opts.Phase, PhaseDevelop)
	}
	if opts.Type != TypeTask {
		t.Errorf("SaveOptions.Type = %q, want %q", opts.Type, TypeTask)
	}
}

// Helper function for string contains
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsString(s, substr)))
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
