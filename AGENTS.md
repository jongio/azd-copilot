# AGENTS.md

## Project Structure

This is the `azd-copilot` CLI extension — a Go project that provides Azure-focused agents, skills, and MCP server integration for GitHub Copilot CLI.

```
cli/src/internal/assets/
├── agents/          # Agent definitions (*.md) — editable
├── ghcp4a-skills/   # Upstream skills from microsoft/GitHub-Copilot-for-Azure — editable, synced via smart merge
├── skills/          # Custom skills maintained in this repo — editable
├── agents.go        # Go embed for agents
└── skills.go        # Go embed for ghcp4a-skills + skills
```

## Critical Rules

### Editing upstream skills (`ghcp4a-skills/`)

The `ghcp4a-skills/` directory is synced from upstream ([microsoft/GitHub-Copilot-for-Azure](https://github.com/microsoft/github-copilot-for-azure)) via `mage SyncSkills`. Unlike a destructive wipe, the sync uses **smart merge**: your local changes are preserved, new upstream files are added, and only unmodified files are updated.

To contribute your changes back upstream, run `mage ContributeSkills` — this creates a branch in a clone of the upstream repo with your diffs applied, ready for a PR.

### ✅ Editable directories

| Directory | What | Notes |
|-----------|------|-------|
| `cli/src/internal/assets/agents/` | Agent definitions | Freely editable |
| `cli/src/internal/assets/ghcp4a-skills/` | Upstream skills | Editable — changes survive sync, contribute back via `mage ContributeSkills` |
| `cli/src/internal/assets/skills/` | Custom skills | Freely editable, not synced |
| `cli/src/internal/copilot/` | Go launcher code | Freely editable |
| `cli/src/cmd/` | CLI commands | Freely editable |

## Build & Test

```bash
cd cli
go build ./...
go test ./...
```

## Adding a Custom Skill

Create a directory in `skills/` with a `SKILL.md`:

```
skills/my-skill/
├── SKILL.md           # Required: YAML frontmatter + instructions
├── references/        # Optional: Additional docs
└── assets/            # Optional: Templates
```

The `SKILL.md` must have YAML frontmatter with `name` and `description`. After adding, run `mage updateCounts`.

## Upstream Skill Workflow

| Command | What it does |
|---------|-------------|
| `mage SyncSkills` | Pull latest upstream skills into `ghcp4a-skills/` (smart merge — keeps your changes) |
| `mage SyncSkills /path/to/clone` | Sync from a local clone instead of cloning remotely |
| `mage SyncSkills url@branch` | Sync from a custom repo/branch (e.g. a fork) |
| `mage ContributeSkills` | Create a branch with your `ghcp4a-skills/` changes for a PR to upstream |

## MCP Server Configuration

MCP servers are auto-configured in `~/.copilot/mcp-config.json` by `ConfigureMCPServer()` in `launcher.go`. To add a new MCP server, add it to the `requiredServers` map and the full config template.
