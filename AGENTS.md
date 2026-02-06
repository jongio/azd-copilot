# AGENTS.md

## Project Structure

This is the `azd-copilot` CLI extension — a Go project that provides Azure-focused agents, skills, and MCP server integration for GitHub Copilot CLI.

```
cli/src/internal/assets/
├── agents/          # Agent definitions (*.md) — editable
├── skills/          # ⛔ UPSTREAM skills — DO NOT EDIT
├── custom-skills/   # ✅ Custom skills — editable
├── agents.go        # Go embed for agents
└── skills.go        # Go embed for skills + custom-skills
```

## Critical Rules

### ⛔ NEVER edit files in `cli/src/internal/assets/skills/`

The `skills/` directory is **synced from upstream** ([microsoft/GitHub-Copilot-for-Azure](https://github.com/microsoft/github-copilot-for-azure)) and will be **overwritten** by `mage syncSkills`. Any changes made here will be lost.

If you need to add or modify skill behavior:
1. Create a new skill in `cli/src/internal/assets/custom-skills/` instead
2. Or modify an existing custom skill in that directory
3. Reference the new custom skill from agent files

### ✅ Editable directories

| Directory | What | Notes |
|-----------|------|-------|
| `cli/src/internal/assets/agents/` | Agent definitions | Freely editable |
| `cli/src/internal/assets/custom-skills/` | Custom skills | Freely editable, not synced |
| `cli/src/internal/copilot/` | Go launcher code | Freely editable |
| `cli/src/cmd/` | CLI commands | Freely editable |

## Build & Test

```bash
cd cli
go build ./...
go test ./...
```

## Adding a Custom Skill

Create a directory in `custom-skills/` with a `SKILL.md`:

```
custom-skills/my-skill/
├── SKILL.md           # Required: YAML frontmatter + instructions
├── references/        # Optional: Additional docs
└── assets/            # Optional: Templates
```

The `SKILL.md` must have YAML frontmatter with `name` and `description`. After adding, run `mage updateCounts`.

## MCP Server Configuration

MCP servers are auto-configured in `~/.copilot/mcp-config.json` by `ConfigureMCPServer()` in `launcher.go`. To add a new MCP server, add it to the `requiredServers` map and the full config template.
