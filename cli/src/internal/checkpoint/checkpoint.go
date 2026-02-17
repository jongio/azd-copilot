// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package checkpoint

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jongio/azd-copilot/cli/src/internal/spec"
	"github.com/jongio/azd-core/fileutil"
)

// Phase represents a build phase
type Phase string

const (
	PhaseSpec    Phase = "spec"
	PhaseDesign  Phase = "design"
	PhaseDevelop Phase = "develop"
	PhaseQuality Phase = "quality"
	PhaseDeploy  Phase = "deploy"
)

// CheckpointType indicates what triggered the checkpoint
type CheckpointType string

const (
	TypePhase    CheckpointType = "phase"    // After phase completion
	TypeTask     CheckpointType = "task"     // After task completion
	TypeSnapshot CheckpointType = "snapshot" // Full file backup
	TypeRecovery CheckpointType = "recovery" // Error recovery
	TypeManual   CheckpointType = "manual"   // User-created
)

// Trigger indicates what caused the checkpoint to be created
type Trigger string

const (
	TriggerPhaseCompleted    Trigger = "phase_completed"
	TriggerTaskCompleted     Trigger = "task_completed"
	TriggerUserInterjection  Trigger = "user_interjection"
	TriggerBeforeDeployment  Trigger = "before_deployment"
	TriggerBeforeDestructive Trigger = "before_destructive"
	TriggerPeriodic          Trigger = "periodic"
	TriggerErrorRecovery     Trigger = "error_recovery"
	TriggerManual            Trigger = "manual"
)

// FileState tracks file changes
type FileState struct {
	Created  []string          `json:"created,omitempty"`
	Modified []string          `json:"modified,omitempty"`
	Deleted  []string          `json:"deleted,omitempty"`
	Hashes   map[string]string `json:"hashes,omitempty"` // path → SHA256
}

// TaskState tracks task execution state
type TaskState struct {
	CompletedTasks []string      `json:"completedTasks,omitempty"`
	PendingTasks   []string      `json:"pendingTasks,omitempty"`
	FailedTasks    []TaskFailure `json:"failedTasks,omitempty"`
}

// TaskFailure records a failed task
type TaskFailure struct {
	Task      string    `json:"task"`
	Error     string    `json:"error"`
	Retries   int       `json:"retries"`
	Timestamp time.Time `json:"timestamp"`
}

// Context provides information for resuming
type Context struct {
	SpecHash          string `json:"specHash,omitempty"`
	LastPrompt        string `json:"lastPrompt,omitempty"`
	LastAgentResponse string `json:"lastAgentResponse,omitempty"` // Truncated
	SessionID         string `json:"sessionId,omitempty"`         // Copilot CLI session
	ErrorMessage      string `json:"errorMessage,omitempty"`
	ErrorStack        string `json:"errorStack,omitempty"`
}

// Checkpoint represents a saved state
type Checkpoint struct {
	// Identity
	ID        string `json:"id"`
	ProjectID string `json:"projectId,omitempty"`
	SessionID string `json:"sessionId,omitempty"`

	// Classification
	Type    CheckpointType `json:"type"`
	Trigger Trigger        `json:"trigger"`
	Phase   Phase          `json:"phase"`

	// State
	CompletedPhases []Phase   `json:"completedPhases,omitempty"`
	Tasks           TaskState `json:"tasks,omitempty"`
	Files           FileState `json:"files,omitempty"`

	// Context for resume
	Context Context `json:"context,omitempty"`

	// Metadata
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	CanResume   bool      `json:"canResume"`
}

// GetCheckpointDir returns the checkpoint directory path
func GetCheckpointDir() string {
	m, _ := spec.LoadMetadata()
	return m.CheckpointDir
}

// List returns all checkpoints sorted by creation time (newest first)
func List() ([]Checkpoint, error) {
	checkpointDir := GetCheckpointDir()
	indexPath := filepath.Join(checkpointDir, "index.json")

	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return []Checkpoint{}, nil
	}

	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoint index: %w", err)
	}

	var checkpoints []Checkpoint
	if err := json.Unmarshal(data, &checkpoints); err != nil {
		return nil, fmt.Errorf("failed to parse checkpoint index: %w", err)
	}

	// Sort by creation time, newest first
	sort.Slice(checkpoints, func(i, j int) bool {
		return checkpoints[i].CreatedAt.After(checkpoints[j].CreatedAt)
	})

	return checkpoints, nil
}

// Get returns a specific checkpoint by ID
func Get(id string) (*Checkpoint, error) {
	checkpoints, err := List()
	if err != nil {
		return nil, err
	}

	for _, cp := range checkpoints {
		if cp.ID == id {
			return &cp, nil
		}
	}

	return nil, fmt.Errorf("checkpoint not found: %s", id)
}

// Latest returns the most recent checkpoint
func Latest() (*Checkpoint, error) {
	checkpoints, err := List()
	if err != nil {
		return nil, err
	}

	if len(checkpoints) == 0 {
		return nil, nil
	}

	return &checkpoints[0], nil
}

// LatestForPhase returns the most recent checkpoint for a specific phase
func LatestForPhase(phase Phase) (*Checkpoint, error) {
	checkpoints, err := List()
	if err != nil {
		return nil, err
	}

	for _, cp := range checkpoints {
		if cp.Phase == phase {
			return &cp, nil
		}
	}

	return nil, nil
}

// SaveOptions configures checkpoint creation
type SaveOptions struct {
	Phase           Phase
	Type            CheckpointType
	Trigger         Trigger
	Description     string
	Files           FileState
	Tasks           TaskState
	CompletedPhases []Phase
	Context         Context
	ProjectID       string
	SessionID       string
}

// Save creates a new checkpoint (legacy signature for compatibility)
func Save(phase Phase, description string, files []string) (*Checkpoint, error) {
	return SaveWithOptions(SaveOptions{
		Phase:       phase,
		Type:        TypePhase,
		Trigger:     TriggerPhaseCompleted,
		Description: description,
		Files:       FileState{Created: files},
	})
}

// SaveWithOptions creates a checkpoint with full options
func SaveWithOptions(opts SaveOptions) (*Checkpoint, error) {
	checkpointDir := GetCheckpointDir()

	// Ensure directory exists
	if err := fileutil.EnsureDir(checkpointDir); err != nil {
		return nil, fmt.Errorf("failed to create checkpoint directory: %w", err)
	}

	// Generate ID based on type and timestamp
	id := fmt.Sprintf("%s-%s-%d", opts.Type, opts.Phase, time.Now().Unix())

	// Compute file hashes if files provided
	if opts.Files.Hashes == nil && len(opts.Files.Created) > 0 {
		opts.Files.Hashes = computeFileHashes(opts.Files.Created)
	}

	// Get spec hash for context
	if opts.Context.SpecHash == "" {
		opts.Context.SpecHash = getSpecHash()
	}

	checkpoint := Checkpoint{
		ID:              id,
		ProjectID:       opts.ProjectID,
		SessionID:       opts.SessionID,
		Type:            opts.Type,
		Trigger:         opts.Trigger,
		Phase:           opts.Phase,
		CompletedPhases: opts.CompletedPhases,
		Tasks:           opts.Tasks,
		Files:           opts.Files,
		Context:         opts.Context,
		Description:     opts.Description,
		CreatedAt:       time.Now(),
		CanResume:       true,
	}

	// Load existing checkpoints
	checkpoints, err := List()
	if err != nil {
		checkpoints = []Checkpoint{}
	}

	// Add new checkpoint
	checkpoints = append(checkpoints, checkpoint)

	// Save index
	if err := saveIndex(checkpoints); err != nil {
		return nil, err
	}

	// Save checkpoint detail using atomic write
	detailPath := filepath.Join(checkpointDir, checkpoint.ID+".json")
	if err := fileutil.AtomicWriteJSON(detailPath, checkpoint); err != nil {
		return nil, fmt.Errorf("failed to save checkpoint: %w", err)
	}

	// Update latest.json
	latestPath := filepath.Join(checkpointDir, "latest.json")
	_ = fileutil.AtomicWriteJSON(latestPath, checkpoint)

	return &checkpoint, nil
}

// SavePhaseCheckpoint creates a checkpoint after phase completion
func SavePhaseCheckpoint(phase Phase, description string, files []string, completedPhases []Phase) (*Checkpoint, error) {
	return SaveWithOptions(SaveOptions{
		Phase:           phase,
		Type:            TypePhase,
		Trigger:         TriggerPhaseCompleted,
		Description:     description,
		Files:           FileState{Created: files, Hashes: computeFileHashes(files)},
		CompletedPhases: completedPhases,
	})
}

// SaveTaskCheckpoint creates a checkpoint after task completion
func SaveTaskCheckpoint(phase Phase, taskName string, files []string) (*Checkpoint, error) {
	return SaveWithOptions(SaveOptions{
		Phase:       phase,
		Type:        TypeTask,
		Trigger:     TriggerTaskCompleted,
		Description: fmt.Sprintf("Task completed: %s", taskName),
		Files:       FileState{Created: files},
		Tasks:       TaskState{CompletedTasks: []string{taskName}},
	})
}

// SaveRecoveryCheckpoint creates a checkpoint for error recovery
func SaveRecoveryCheckpoint(phase Phase, err error, context string) (*Checkpoint, error) {
	return SaveWithOptions(SaveOptions{
		Phase:       phase,
		Type:        TypeRecovery,
		Trigger:     TriggerErrorRecovery,
		Description: fmt.Sprintf("Recovery checkpoint: %v", err),
		Context: Context{
			ErrorMessage: err.Error(),
			LastPrompt:   context,
		},
	})
}

// SaveBeforeDeployment creates a checkpoint before deployment
func SaveBeforeDeployment(files []string) (*Checkpoint, error) {
	return SaveWithOptions(SaveOptions{
		Phase:           PhaseDeploy,
		Type:            TypeSnapshot,
		Trigger:         TriggerBeforeDeployment,
		Description:     "Pre-deployment snapshot",
		Files:           FileState{Created: files, Hashes: computeFileHashes(files)},
		CompletedPhases: []Phase{PhaseSpec, PhaseDesign, PhaseDevelop, PhaseQuality},
	})
}

// computeFileHashes computes SHA256 hashes for files
func computeFileHashes(files []string) map[string]string {
	hashes := make(map[string]string)
	for _, path := range files {
		hash, err := hashFile(path)
		if err == nil {
			hashes[path] = hash
		}
	}
	return hashes
}

// hashFile computes SHA256 of a file
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// getSpecHash returns hash of the spec file
func getSpecHash() string {
	specPath := spec.GetSpecPath()
	hash, err := hashFile(specPath)
	if err != nil {
		return ""
	}
	return hash
}

// Delete removes a checkpoint
func Delete(id string) error {
	checkpointDir := GetCheckpointDir()

	checkpoints, err := List()
	if err != nil {
		return err
	}

	var updated []Checkpoint
	found := false
	for _, cp := range checkpoints {
		if cp.ID == id {
			found = true
			// Delete detail file
			detailPath := filepath.Join(checkpointDir, cp.ID+".json")
			_ = os.Remove(detailPath) // ignore error
		} else {
			updated = append(updated, cp)
		}
	}

	if !found {
		return fmt.Errorf("checkpoint not found: %s", id)
	}

	return saveIndex(updated)
}

// Clear removes all checkpoints
func Clear() error {
	return os.RemoveAll(GetCheckpointDir())
}

func saveIndex(checkpoints []Checkpoint) error {
	checkpointDir := GetCheckpointDir()
	indexPath := filepath.Join(checkpointDir, "index.json")
	return fileutil.AtomicWriteJSON(indexPath, checkpoints)
}

// GenerateResumePrompt generates a prompt for resuming from a checkpoint
func GenerateResumePrompt(checkpoint *Checkpoint) string {
	nextPhase := NextPhase(checkpoint.Phase)

	var sb strings.Builder

	sb.WriteString("# Resume Build from Checkpoint\n\n")

	// Summary section
	sb.WriteString("## Previous Progress\n\n")
	fmt.Fprintf(&sb, "- **Checkpoint ID:** `%s`\n", checkpoint.ID)
	fmt.Fprintf(&sb, "- **Phase Completed:** %s\n", checkpoint.Phase)
	fmt.Fprintf(&sb, "- **Type:** %s\n", checkpoint.Type)
	fmt.Fprintf(&sb, "- **Created:** %s\n", checkpoint.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(&sb, "- **Description:** %s\n\n", checkpoint.Description)

	// Completed phases
	if len(checkpoint.CompletedPhases) > 0 {
		sb.WriteString("### Completed Phases\n")
		for _, p := range checkpoint.CompletedPhases {
			fmt.Fprintf(&sb, "- [x] %s\n", p)
		}
		fmt.Fprintf(&sb, "- [ ] %s ← **Continue here**\n\n", nextPhase)
	}

	// Files created
	allFiles := make([]string, 0, len(checkpoint.Files.Created)+len(checkpoint.Files.Modified))
	allFiles = append(allFiles, checkpoint.Files.Created...)
	allFiles = append(allFiles, checkpoint.Files.Modified...)
	if len(allFiles) > 0 {
		sb.WriteString("### Files Already Created\n\n")
		sb.WriteString("```\n")
		for _, f := range allFiles {
			fmt.Fprintf(&sb, "%s\n", f)
		}
		sb.WriteString("```\n\n")
	}

	// Error context if recovery checkpoint
	if checkpoint.Type == TypeRecovery && checkpoint.Context.ErrorMessage != "" {
		sb.WriteString("### Previous Error\n\n")
		sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", checkpoint.Context.ErrorMessage))
		sb.WriteString("The previous attempt failed with the error above. Please try a different approach.\n\n")
	}

	// Instructions
	sb.WriteString("## Instructions\n\n")

	if checkpoint.Type == TypeRecovery {
		sb.WriteString("This is a **recovery checkpoint**. The previous attempt encountered an error.\n\n")
		sb.WriteString("1. Read the spec at `docs/spec.md`\n")
		sb.WriteString("2. Review existing files listed above\n")
		sb.WriteString("3. Identify what went wrong and try a different approach\n")
		sb.WriteString(fmt.Sprintf("4. Continue with the **%s** phase\n\n", checkpoint.Phase))
	} else {
		sb.WriteString(fmt.Sprintf("The **%s** phase completed successfully. Continue with **%s**.\n\n", checkpoint.Phase, nextPhase))
		sb.WriteString("1. Read the spec at `docs/spec.md`\n")
		sb.WriteString("2. Review existing files - do NOT regenerate them unless changes are needed\n")
		sb.WriteString(fmt.Sprintf("3. Proceed with the **%s** phase\n", nextPhase))
		sb.WriteString(fmt.Sprintf("4. Save a checkpoint when %s is complete\n\n", nextPhase))
	}

	// Phase-specific guidance
	sb.WriteString(getPhaseGuidance(nextPhase))

	return sb.String()
}

// NextPhase returns the phase that follows the given phase
func NextPhase(phase Phase) Phase {
	switch phase {
	case PhaseSpec:
		return PhaseDesign
	case PhaseDesign:
		return PhaseDevelop
	case PhaseDevelop:
		return PhaseQuality
	case PhaseQuality:
		return PhaseDeploy
	case PhaseDeploy:
		return PhaseDeploy // Terminal
	default:
		return PhaseDevelop
	}
}

// getPhaseGuidance returns guidance text for a phase
func getPhaseGuidance(phase Phase) string {
	switch phase {
	case PhaseDesign:
		return `### Design Phase Tasks
- Finalize architecture decisions
- Create azure.yaml configuration
- Design database schema
- Plan API endpoints
`
	case PhaseDevelop:
		return `### Develop Phase Tasks
- Generate backend code
- Generate frontend code (if applicable)
- Implement database migrations
- Create API endpoints
`
	case PhaseQuality:
		return `### Quality Phase Tasks
- Generate unit tests
- Generate integration tests
- Run linter and fix issues
- Perform security review
`
	case PhaseDeploy:
		return `### Deploy Phase Tasks
- Generate Bicep infrastructure files
- Create CI/CD pipeline
- Generate documentation
- Run preflight checks
- Deploy to Azure (with user approval)
`
	default:
		return ""
	}
}

// ListByType returns checkpoints filtered by type
func ListByType(cpType CheckpointType) ([]Checkpoint, error) {
	all, err := List()
	if err != nil {
		return nil, err
	}

	var filtered []Checkpoint
	for _, cp := range all {
		if cp.Type == cpType {
			filtered = append(filtered, cp)
		}
	}
	return filtered, nil
}

// KeepLatest removes old checkpoints, keeping only the N most recent
func KeepLatest(n int) error {
	checkpoints, err := List()
	if err != nil {
		return err
	}

	if len(checkpoints) <= n {
		return nil
	}

	checkpointDir := GetCheckpointDir()

	// Delete older checkpoints (list is already sorted newest first)
	for i := n; i < len(checkpoints); i++ {
		detailPath := filepath.Join(checkpointDir, checkpoints[i].ID+".json")
		_ = os.Remove(detailPath)
	}

	// Update index with only the kept checkpoints
	return saveIndex(checkpoints[:n])
}

// DetectInterrupted checks if there's an incomplete build that can be resumed
func DetectInterrupted() (*Checkpoint, error) {
	latest, err := Latest()
	if err != nil || latest == nil {
		return nil, err
	}

	// If the latest checkpoint is not for deploy phase, build was interrupted
	if latest.Phase != PhaseDeploy && latest.CanResume {
		return latest, nil
	}

	return nil, nil
}

// GetProjectFiles returns all files tracked across checkpoints
func GetProjectFiles() ([]string, error) {
	checkpoints, err := List()
	if err != nil {
		return nil, err
	}

	fileSet := make(map[string]bool)
	for _, cp := range checkpoints {
		for _, f := range cp.Files.Created {
			fileSet[f] = true
		}
		for _, f := range cp.Files.Modified {
			fileSet[f] = true
		}
	}

	var files []string
	for f := range fileSet {
		files = append(files, f)
	}
	sort.Strings(files)
	return files, nil
}
