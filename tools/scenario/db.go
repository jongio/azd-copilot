// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package scenario

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS runs (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    scenario        TEXT NOT NULL,
    session_id      TEXT NOT NULL,
    git_commit      TEXT,
    started_at      DATETIME NOT NULL,
    duration_sec    INTEGER,
    total_turns     INTEGER,
    azd_up_attempts INTEGER,
    bicep_edits     INTEGER,
    delegated       BOOLEAN,
    deployed        BOOLEAN,
    score           REAL,
    passed          BOOLEAN
);

CREATE TABLE IF NOT EXISTS run_skills (
    run_id   INTEGER REFERENCES runs(id),
    skill    TEXT NOT NULL,
    invoked  BOOLEAN NOT NULL DEFAULT 0,
    PRIMARY KEY (run_id, skill)
);

CREATE TABLE IF NOT EXISTS run_regressions (
    run_id        INTEGER REFERENCES runs(id),
    name          TEXT NOT NULL,
    occurrences   INTEGER,
    max_allowed   INTEGER,
    passed        BOOLEAN,
    PRIMARY KEY (run_id, name)
);
`

// Run holds the result of analyzing a scenario session.
type Run struct {
	Scenario     string
	SessionID    string
	GitCommit    string
	StartedAt    time.Time
	DurationSec  int
	TotalTurns   int
	AzdUpAttempts int
	BicepEdits   int
	Delegated    bool
	Deployed     bool
	Score        float64
	Passed       bool
	Skills       map[string]bool       // skill name -> was invoked
	Regressions  map[string]RegResult  // regression name -> result
}

// RegResult is the result of checking one regression pattern.
type RegResult struct {
	Occurrences int
	MaxAllowed  int
	Passed      bool
}

// DB wraps a SQLite connection for scenario results.
type DB struct {
	db *sql.DB
}

// OpenDB opens (or creates) the scenario results database.
func OpenDB(dbPath string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}

	return &DB{db: db}, nil
}

// Close closes the database connection.
func (d *DB) Close() error {
	return d.db.Close()
}

// InsertRun saves a run result and returns the inserted row ID.
func (d *DB) InsertRun(r *Run) (int64, error) {
	res, err := d.db.Exec(`
		INSERT INTO runs (scenario, session_id, git_commit, started_at, duration_sec,
			total_turns, azd_up_attempts, bicep_edits, delegated, deployed, score, passed)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.Scenario, r.SessionID, r.GitCommit, r.StartedAt, r.DurationSec,
		r.TotalTurns, r.AzdUpAttempts, r.BicepEdits, r.Delegated, r.Deployed,
		r.Score, r.Passed,
	)
	if err != nil {
		return 0, fmt.Errorf("insert run: %w", err)
	}

	runID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Insert skill records
	for skill, invoked := range r.Skills {
		if _, err := d.db.Exec(`INSERT INTO run_skills (run_id, skill, invoked) VALUES (?, ?, ?)`,
			runID, skill, invoked); err != nil {
			return runID, fmt.Errorf("insert skill %s: %w", skill, err)
		}
	}

	// Insert regression records
	for name, reg := range r.Regressions {
		if _, err := d.db.Exec(`INSERT INTO run_regressions (run_id, name, occurrences, max_allowed, passed) VALUES (?, ?, ?, ?, ?)`,
			runID, name, reg.Occurrences, reg.MaxAllowed, reg.Passed); err != nil {
			return runID, fmt.Errorf("insert regression %s: %w", name, err)
		}
	}

	return runID, nil
}

// ListRuns returns recent runs for a scenario (newest first).
func (d *DB) ListRuns(scenarioName string, limit int) ([]Run, error) {
	rows, err := d.db.Query(`
		SELECT id, scenario, session_id, git_commit, started_at, duration_sec,
			total_turns, azd_up_attempts, bicep_edits, delegated, deployed, score, passed
		FROM runs
		WHERE scenario = ? OR ? = ''
		ORDER BY started_at DESC
		LIMIT ?`, scenarioName, scenarioName, limit)
	if err != nil {
		return nil, fmt.Errorf("query runs: %w", err)
	}
	defer rows.Close()

	var runs []Run
	for rows.Next() {
		var r Run
		var id int64
		var gitCommit sql.NullString
		if err := rows.Scan(&id, &r.Scenario, &r.SessionID, &gitCommit,
			&r.StartedAt, &r.DurationSec, &r.TotalTurns, &r.AzdUpAttempts,
			&r.BicepEdits, &r.Delegated, &r.Deployed, &r.Score, &r.Passed); err != nil {
			return nil, fmt.Errorf("scan run: %w", err)
		}
		if gitCommit.Valid {
			r.GitCommit = gitCommit.String
		}
		runs = append(runs, r)
	}
	return runs, rows.Err()
}

// ListRunsWithDetails returns runs with their skills and regressions populated.
func (d *DB) ListRunsWithDetails(scenarioName string, limit int) ([]Run, error) {
	rows, err := d.db.Query(`
		SELECT id, scenario, session_id, git_commit, started_at, duration_sec,
			total_turns, azd_up_attempts, bicep_edits, delegated, deployed, score, passed
		FROM runs
		WHERE scenario = ? OR ? = ''
		ORDER BY started_at ASC
		LIMIT ?`, scenarioName, scenarioName, limit)
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
		idRuns = append(idRuns, ir)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load skills and regressions for each run
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

func (d *DB) loadRunDetails(runID int64, r *Run) error {
	// Load skills
	sRows, err := d.db.Query(`SELECT skill, invoked FROM run_skills WHERE run_id = ?`, runID)
	if err != nil {
		return fmt.Errorf("query skills: %w", err)
	}
	defer sRows.Close()
	for sRows.Next() {
		var skill string
		var invoked bool
		if err := sRows.Scan(&skill, &invoked); err != nil {
			return fmt.Errorf("scan skill: %w", err)
		}
		r.Skills[skill] = invoked
	}

	// Load regressions
	rRows, err := d.db.Query(`SELECT name, occurrences, max_allowed, passed FROM run_regressions WHERE run_id = ?`, runID)
	if err != nil {
		return fmt.Errorf("query regressions: %w", err)
	}
	defer rRows.Close()
	for rRows.Next() {
		var name string
		var reg RegResult
		if err := rRows.Scan(&name, &reg.Occurrences, &reg.MaxAllowed, &reg.Passed); err != nil {
			return fmt.Errorf("scan regression: %w", err)
		}
		r.Regressions[name] = reg
	}

	return nil
}

// ListScenarios returns distinct scenario names from the runs table.
func (d *DB) ListScenarios() ([]string, error) {
	rows, err := d.db.Query(`SELECT DISTINCT scenario FROM runs ORDER BY scenario`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, rows.Err()
}
