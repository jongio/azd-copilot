# Scenario Runner

Automated testing and quality tracking for azd-copilot. Define repeatable scenarios as YAML, replay them through `azd copilot`, score the results, detect regressions, and track improvements over time.

## Quick Start

```bash
cd tools/scenario

# 1. Run an existing scenario
mage scenario:run ../../scenarios/dog-breed-lookup.yaml

# 2. Analyze the session that was produced
mage scenario:analyze <session-id> ../../scenarios/dog-breed-lookup.yaml

# 3. View results
mage scenario:history dog-breed-lookup
mage scenario:dashboard
```

## Concepts

| Concept | Description |
|---------|-------------|
| **Scenario** | A YAML file defining prompts to send, scoring limits, regression patterns, and optional Playwright verification steps. |
| **Run** | One execution of a scenario — produces a session ID, metrics, and a composite score. |
| **Scoring** | Each metric (duration, turns, azd up attempts, Bicep edits, skill invocations, regressions) contributes to a weighted 0–100% composite score. |
| **Improvement Loop** | Automated cycle: run scenario → analyze → send failures to copilot for fixes → rebuild extension → repeat. |

## Scenario YAML Format

```yaml
name: my-scenario
description: "What this scenario tests"
timeout: 35m                       # overall timeout for the entire run

prompts:
  - text: "build a todo app"       # first prompt sent to azd copilot
    success_criteria:               # optional — what should exist after this prompt
      files_exist:
        - azure.yaml
        - infra/main.bicep
      deployed: true
      endpoint_responds: true

  - text: "add a database"         # subsequent prompts resume the session

scoring:
  max_duration_minutes: 25         # score degrades proportionally above limit
  max_turns: 40                    # agent turn count
  max_azd_up_attempts: 4           # how many times `azd up` was called
  max_bicep_edits: 5               # edits to main.bicep
  must_delegate: true              # requires task() delegation to sub-agents
  must_invoke_skills:              # skills that must be invoked
    - avm-bicep-rules
    - azure-functions
  regressions:                     # regex patterns to watch for in assistant output
    - name: "ACR auth spiral"
      pattern: "ACR.*auth|can't pull|registry.*credential"
      max_occurrences: 2

verification:                      # optional Playwright steps run against the deployed app
  - name: "homepage loads"
    action: navigate
    url: "{{endpoint}}"            # {{endpoint}} is replaced with the discovered URL
    status_code: 200

  - name: "page has content"
    action: check
    selector: "body"
    value: "expected text"
```

### Verification Actions

| Action | Fields | Description |
|--------|--------|-------------|
| `navigate` | `url`, `status_code`, `value` | Navigate to URL, check status code and optional body text |
| `click` | `selector` | Click an element |
| `type` | `selector`, `value` | Type text into an input |
| `wait` | `selector` | Wait for an element to appear |
| `check` | `selector`, `value` | Assert element is visible and optionally contains text |
| `check_not_empty` | `selector` | Assert at least one matching element exists |
| `screenshot` | — | Take a full-page screenshot |

The `{{endpoint}}` placeholder is replaced with the deployed app URL, discovered from `azd env get-values`.

## Mage Commands

All commands are in the `scenario` namespace. Run from `tools/scenario/`:

| Command | Description |
|---------|-------------|
| `mage scenario:extract <session-id>` | Generate a scenario YAML from an existing copilot session log |
| `mage scenario:run <scenario.yaml>` | Execute a scenario (launches `azd copilot` for each prompt) |
| `mage scenario:analyze <session-id> <scenario.yaml>` | Score a session against a scenario and save to DB |
| `mage scenario:history [scenario-name]` | Show recent runs (all scenarios if name omitted) |
| `mage scenario:dashboard` | Generate and serve an interactive HTML dashboard |
| `mage scenario:loop <scenario.yaml>` | Run the full improvement loop (3 iterations) |
| `mage scenario:export` | Export results from SQLite to `results.json` (for git) |
| `mage scenario:import` | Import results from `results.json` into SQLite |

## How It Works

### Running a Scenario

`scenario:run` launches `azd copilot --yolo -p "<prompt>"` for each prompt in a fresh temp directory. Subsequent prompts use `--resume` to continue the same session. The runner watches for:

- **task_complete** events in `~/.copilot/session-state/<id>/events.jsonl` — signals the prompt is done
- **Stuck loops** — kills the process if the same short line repeats 5+ times
- **Idle timeout** — kills after 3 minutes of no output
- **Per-prompt timeout** — kills after 15 minutes per prompt

### Analyzing & Scoring

`scenario:analyze` reads the session's `events.jsonl` and computes:

| Metric | Weight | Scoring |
|--------|--------|---------|
| Duration | 25 pts | Full points at/below limit, proportional reduction above (limit/actual) |
| Turns | 20 pts | Same proportional scoring |
| azd up attempts | 20 pts | Same proportional scoring |
| Bicep edits | 10 pts | Same proportional scoring |
| Delegation | 10 pts | Binary — did the agent use `task()` |
| Skills (per skill) | 5 pts | Binary — was each required skill invoked |
| Regressions (per check) | 5 pts | Binary — pattern occurrences within limit |

A run **passes** only if all metrics are within their limits. The composite score (0–100%) uses proportional credit — even a run 2× over the turn limit gets 50% credit for turns.

### Extracting Scenarios from Sessions

`scenario:extract` creates a scenario YAML from a real copilot session by:

1. Extracting user messages as prompts
2. Computing scoring baselines from actual metrics (with ~30-50% headroom)
3. Adding default regression patterns (ACR auth, zone redundancy, npm lockfile)

### Improvement Loop

`scenario:loop` automates the full cycle:

1. **Run** the scenario → get a session ID
2. **Analyze** the session → compute score, save to DB
3. If failed: **Generate a fix prompt** describing what went wrong and send it to `azd copilot` targeting `cli/src/internal/assets/`
4. **Rebuild** the extension (`go build && go test && mage build`)
5. **Repeat** (default: 3 iterations, stops early on pass)

## Results Storage

| File | Format | Purpose |
|------|--------|---------|
| `scenarios/results.db` | SQLite | Primary storage — gitignored, used by dashboard |
| `scenarios/results.json` | JSON | Portable export — committed to git for sharing |

Use `scenario:export` to save DB → JSON and `scenario:import` to load JSON → DB (skips duplicates by session ID).

### Database Schema

- **`runs`** — one row per scenario execution (score, metrics, pass/fail, git commit)
- **`run_skills`** — which required skills were invoked per run
- **`run_regressions`** — regression pattern match counts per run
- **`run_verification`** — Playwright verification step results per run

## Dashboard

The dashboard is a self-contained HTML file that loads `results.db` client-side via [sql.js](https://github.com/sql-js/sql.js) (WASM SQLite). It shows:

- Summary cards (total runs, scenarios, latest score)
- Per-scenario tabs with charts (score, duration, turns, azd up over time)
- Run history table with expandable skill/regression details
- Comparison view (first run vs latest)

The dashboard reads the DB live — no regeneration needed after new runs. Served at `http://localhost:8086` via `mage scenario:dashboard`.

## File Layout

```
tools/scenario/
├── scenario.go       # YAML types, Load/Save, session event parsing
├── runner.go         # RunScenario — prompt execution, stuck detection, event watching
├── analyze.go        # Extract, Analyze, scoring, FormatReport
├── loop.go           # RunLoop — the improvement cycle
├── verify.go         # Playwright verification step execution
├── db.go             # SQLite schema, InsertRun, ListRuns
├── db_export.go      # ExportJSON / ImportJSON
├── dashboard.go      # HTML dashboard generation
├── magefile.go       # Mage commands (extract, run, analyze, history, dashboard, loop)
├── scenario_test.go  # Unit tests
└── go.mod

scenarios/
├── dog-breed-lookup.yaml       # Example: SWA + Container App
├── functions-todo-api.yaml     # Example: Azure Functions + Cosmos DB
├── results.db                  # SQLite results (gitignored)
├── results.json                # JSON export (committed)
└── dashboard.html              # Generated dashboard (gitignored)
```
