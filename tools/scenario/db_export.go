// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package scenario

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// exportRun is the JSON-serializable form of a single run with all details.
type exportRun struct {
	Scenario      string                  `json:"scenario"`
	SessionID     string                  `json:"session_id"`
	GitCommit     string                  `json:"git_commit,omitempty"`
	StartedAt     time.Time               `json:"started_at"`
	DurationSec   int                     `json:"duration_sec"`
	TotalTurns    int                     `json:"total_turns"`
	AzdUpAttempts int                     `json:"azd_up_attempts"`
	BicepEdits    int                     `json:"bicep_edits"`
	Delegated     bool                    `json:"delegated"`
	Deployed      bool                    `json:"deployed"`
	Score         float64                 `json:"score"`
	Passed        bool                    `json:"passed"`
	Skills        map[string]bool         `json:"skills,omitempty"`
	Regressions   map[string]RegResult    `json:"regressions,omitempty"`
	Verification  map[string]VerifyResult `json:"verification,omitempty"`
}

// ExportJSON writes all runs to a JSON file.
func (d *DB) ExportJSON(path string) error {
	runs, err := d.listAllRunsWithDetails()
	if err != nil {
		return fmt.Errorf("list runs: %w", err)
	}

	exported := make([]exportRun, len(runs))
	for i, r := range runs {
		exported[i] = exportRun(r)
	}

	data, err := json.MarshalIndent(exported, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

// ImportJSON reads runs from a JSON file and inserts any that are not already
// present (matched by session_id).
func (d *DB) ImportJSON(path string) (int, error) {
	// #nosec G304 -- Import paths are explicit user input for this local scenario tool.
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("read: %w", err)
	}

	var records []exportRun
	if err := json.Unmarshal(data, &records); err != nil {
		return 0, fmt.Errorf("unmarshal: %w", err)
	}

	imported := 0
	for _, rec := range records {
		// Skip if session already exists
		var count int
		if err := d.db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM runs WHERE session_id = ?`, rec.SessionID).Scan(&count); err != nil {
			return imported, fmt.Errorf("check existing: %w", err)
		}
		if count > 0 {
			continue
		}

		run := &Run{
			Scenario:      rec.Scenario,
			SessionID:     rec.SessionID,
			GitCommit:     rec.GitCommit,
			StartedAt:     rec.StartedAt,
			DurationSec:   rec.DurationSec,
			TotalTurns:    rec.TotalTurns,
			AzdUpAttempts: rec.AzdUpAttempts,
			BicepEdits:    rec.BicepEdits,
			Delegated:     rec.Delegated,
			Deployed:      rec.Deployed,
			Score:         rec.Score,
			Passed:        rec.Passed,
			Skills:        rec.Skills,
			Regressions:   rec.Regressions,
			Verification:  rec.Verification,
		}

		if _, err := d.InsertRun(run); err != nil {
			return imported, fmt.Errorf("insert run %s: %w", rec.SessionID, err)
		}
		imported++
	}

	return imported, nil
}

// listAllRunsWithDetails returns all runs with full details, ordered by started_at ASC.
func (d *DB) listAllRunsWithDetails() ([]Run, error) {
	rows, err := d.db.QueryContext(context.Background(), `
		SELECT id, scenario, session_id, git_commit, started_at, duration_sec,
			total_turns, azd_up_attempts, bicep_edits, delegated, deployed, score, passed
		FROM runs
		ORDER BY started_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("query runs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	idRuns, err := scanIDRuns(rows)
	if err != nil {
		return nil, err
	}

	for i := range idRuns {
		if err := d.loadRunDetails(idRuns[i].id, &idRuns[i].run); err != nil {
			return nil, err
		}
	}

	runs := make([]Run, len(idRuns))
	for i, ir := range idRuns {
		runs[i] = ir.run
	}
	return runs, nil
}
