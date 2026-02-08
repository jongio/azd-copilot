// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package scenario

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// exportRun is the JSON-serializable form of a single run with all details.
type exportRun struct {
	Scenario      string              `json:"scenario"`
	SessionID     string              `json:"session_id"`
	GitCommit     string              `json:"git_commit,omitempty"`
	StartedAt     time.Time           `json:"started_at"`
	DurationSec   int                 `json:"duration_sec"`
	TotalTurns    int                 `json:"total_turns"`
	AzdUpAttempts int                 `json:"azd_up_attempts"`
	BicepEdits    int                 `json:"bicep_edits"`
	Delegated     bool                `json:"delegated"`
	Deployed      bool                `json:"deployed"`
	Score         float64             `json:"score"`
	Passed        bool                `json:"passed"`
	Skills        map[string]bool     `json:"skills,omitempty"`
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
		exported[i] = exportRun{
			Scenario:      r.Scenario,
			SessionID:     r.SessionID,
			GitCommit:     r.GitCommit,
			StartedAt:     r.StartedAt,
			DurationSec:   r.DurationSec,
			TotalTurns:    r.TotalTurns,
			AzdUpAttempts: r.AzdUpAttempts,
			BicepEdits:    r.BicepEdits,
			Delegated:     r.Delegated,
			Deployed:      r.Deployed,
			Score:         r.Score,
			Passed:        r.Passed,
			Skills:        r.Skills,
			Regressions:   r.Regressions,
			Verification:  r.Verification,
		}
	}

	data, err := json.MarshalIndent(exported, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

// ImportJSON reads runs from a JSON file and inserts any that are not already
// present (matched by session_id).
func (d *DB) ImportJSON(path string) (int, error) {
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
		if err := d.db.QueryRow(`SELECT COUNT(*) FROM runs WHERE session_id = ?`, rec.SessionID).Scan(&count); err != nil {
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
	rows, err := d.db.Query(`
		SELECT id, scenario, session_id, git_commit, started_at, duration_sec,
			total_turns, azd_up_attempts, bicep_edits, delegated, deployed, score, passed
		FROM runs
		ORDER BY started_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("query runs: %w", err)
	}
	defer rows.Close()

	type idRun struct {
		id  int64
		run Run
	}
	var idRuns []idRun
	for rows.Next() {
		var ir idRun
		var gitCommit sql.NullString
		if err := rows.Scan(&ir.id, &ir.run.Scenario, &ir.run.SessionID, &gitCommit,
			&ir.run.StartedAt, &ir.run.DurationSec, &ir.run.TotalTurns, &ir.run.AzdUpAttempts,
			&ir.run.BicepEdits, &ir.run.Delegated, &ir.run.Deployed, &ir.run.Score, &ir.run.Passed); err != nil {
			return nil, fmt.Errorf("scan run: %w", err)
		}
		if gitCommit.Valid {
			ir.run.GitCommit = gitCommit.String
		}
		ir.run.Skills = make(map[string]bool)
		ir.run.Regressions = make(map[string]RegResult)
		ir.run.Verification = make(map[string]VerifyResult)
		idRuns = append(idRuns, ir)
	}
	if err := rows.Err(); err != nil {
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
