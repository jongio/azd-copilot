# azd copilot

[![CI](https://github.com/jongio/azd-copilot/actions/workflows/ci.yml/badge.svg)](https://github.com/jongio/azd-copilot/actions/workflows/ci.yml)
[![CodeQL](https://github.com/jongio/azd-copilot/actions/workflows/codeql.yml/badge.svg)](https://github.com/jongio/azd-copilot/actions/workflows/codeql.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jongio/azd-copilot/cli)](https://goreportcard.com/report/github.com/jongio/azd-copilot/cli)
[![Go Reference](https://pkg.go.dev/badge/github.com/jongio/azd-copilot/cli.svg)](https://pkg.go.dev/github.com/jongio/azd-copilot/cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![govulncheck](https://img.shields.io/badge/govulncheck-passing-brightgreen)](https://github.com/jongio/azd-copilot/actions/workflows/govulncheck.yml)
[![golangci-lint](https://img.shields.io/badge/golangci--lint-enabled-blue)](https://github.com/jongio/azd-copilot/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/go-1.26.0-blue)](https://go.dev/)
[![Platform Support](https://img.shields.io/badge/platform-linux%20%7C%20macOS%20%7C%20windows-lightgrey)](https://github.com/jongio/azd-copilot)

**Describe what you want. Copilot builds it and deploys it to Azure.**

[**🌐 Visit the Website →**](https://jongio.github.io/azd-copilot/) | [**📦 azd Extensions Hub →**](https://jongio.github.io/azd-extensions/)

`azd copilot` is an [Azure Developer CLI](https://learn.microsoft.com/azure/developer/azure-developer-cli/) extension that turns GitHub Copilot CLI into a full-stack Azure development assistant — 16 specialized agents, 28 Azure skills, and 6 MCP servers working together so you can go from idea to deployed app in one command.

```
   _                        ___            _ _       _
  /_\  ___ _  _ _ _ ___    / __|___ _ __  (_) |___  | |_
 / _ \|_ / || | '_/ -_ )  | (__/ _ \ '_ \ | | / _ \ |  _|
/_/ \_\/__|\_,_|_| \___|   \___\___/ .__/ |_|_\___/  \__|
                                   |_|
         AI-powered Azure development assistant
          Version 0.1.0 • 16 agents • 28 skills
```

---

## Quick Start

### 1. Install GitHub Copilot CLI

```bash
npm install -g @github/copilot
```

Or see the [GitHub Copilot CLI installation guide](https://docs.github.com/en/copilot/how-tos/copilot-cli/install-copilot-cli) for more options including Homebrew and WinGet.

### 2. Install Azure Developer CLI

- **Windows**: `winget install microsoft.azd`
- **macOS**: `brew tap azure/azd && brew install azd`
- **Linux**: `curl -fsSL https://aka.ms/install-azd.sh | bash`

### 3. Install the Extension

```bash
# Add the azd-copilot extension registry
azd extension source add -n jongio -t url -l https://jongio.github.io/azd-extensions/registry.json

# Install the extension
azd extension install jongio.azd.copilot
```

### 4. Run It

```bash
# Start an interactive session
azd copilot

# Or describe what you want
azd copilot -p "build me a REST API with PostgreSQL and deploy it"

# Build a complete app from a description
azd copilot build "todo app with React frontend and Azure Functions backend"
```

## What It Does

When you run `azd copilot`, the extension:

1. **Installs 16 specialized agents** to `~/.copilot/agents/` — each an expert in a domain (architecture, security, DevOps, data, AI, etc.)
2. **Installs 28 focused skills** to `~/.copilot/skills/` — step-by-step playbooks for Azure tasks (deploy, prepare, validate, cost-optimize, etc.)
3. **Configures 6 MCP servers** in `~/.copilot/mcp-config.json` — Azure, azd, Microsoft Learn, Context7, Playwright, and the extension's own MCP server
4. **Launches GitHub Copilot CLI** with all of the above pre-loaded

The result: Copilot understands Azure deeply and can build, test, deploy, diagnose, and optimize your applications.

## Commands

### Core

| Command | Description |
|---------|-------------|
| `azd copilot` | Start an interactive AI session |
| `azd copilot -p "prompt"` | Run a single prompt non-interactively |
| `azd copilot --resume` | Resume the previous session |
| `azd copilot --agent azure-security` | Use a specific agent |
| `azd copilot --yolo` | Auto-approve all tool executions |

### Build

| Command | Description |
|---------|-------------|
| `azd copilot build "description"` | Generate a complete app from a description |
| `azd copilot build --mode prototype "demo chat app"` | Quick prototype with free tiers |
| `azd copilot build --approve` | Build from an approved spec |

The build process generates a spec → waits for approval → generates code, infra, tests, docs → runs preflight → deploys.

### Quick Actions

| Command | Description |
|---------|-------------|
| `azd copilot init` | Scan a project and generate Azure config |
| `azd copilot review` | AI-powered code review |
| `azd copilot fix` | Auto-fix build errors and test failures |
| `azd copilot optimize` | Cost and performance optimization |
| `azd copilot diagnose` | Troubleshoot Azure deployment issues |

### Management

| Command | Description |
|---------|-------------|
| `azd copilot agents` | List all available agents |
| `azd copilot skills` | List all available skills |
| `azd copilot sessions` | List and manage Copilot sessions |
| `azd copilot checkpoints` | Manage build checkpoints |
| `azd copilot spec` | View or edit the project spec |
| `azd copilot context` | Show current azd project context |
| `azd copilot version` | Show version info |
| `azd copilot mcp configure` | Configure MCP servers |

## Agents

16 specialized agents, each an expert in a specific domain:

| Agent | Role |
|-------|------|
| `azure-manager` | **Coordinator** — plans work, delegates to other agents |
| `azure-architect` | Infrastructure design, Bicep, networking |
| `azure-dev` | Application code, APIs, frontend |
| `azure-data` | Database schema, queries, migrations |
| `azure-ai` | AI services, RAG, agent frameworks |
| `azure-security` | Security audits, identity, compliance |
| `azure-devops` | CI/CD, deployment, observability |
| `azure-finance` | Cost estimation, optimization, TCO analysis |
| `azure-quality` | Testing, code review, refactoring |
| `azure-docs` | README, API docs, ADRs, runbooks |
| `azure-product` | Requirements, acceptance criteria, specs |
| `azure-design` | WCAG compliance, accessibility, UI review |
| `azure-analytics` | Usage analytics, dashboards, metrics |
| `azure-marketing` | Positioning, landing pages, competitive analysis |
| `azure-support` | Troubleshooting, FAQ, error messages |
| `azure-compliance` | Framework assessment (GDPR, SOC2, HIPAA) |

## Skills

28 curated skills providing deep expertise. Skills come from two sources:

- **Upstream skills** (18) — synced from [microsoft/GitHub-Copilot-for-Azure](https://github.com/microsoft/GitHub-Copilot-for-Azure) via `mage SyncSkills`
- **Custom skills** (10) — maintained in this repo

Skills cover: preparing apps for Azure, deploying, validating, cost optimization, diagnostics, networking, observability, AI services, storage, PostgreSQL, Cosmos DB, Functions, Container Apps, security, Entra ID, and more.

Run `azd copilot skills` to see the full list.

## MCP Servers

The extension auto-configures these MCP servers for Copilot CLI:

| Server | Purpose |
|--------|---------|
| `azure` | Azure resource operations via `@azure/mcp` |
| `azd` | Azure Developer CLI operations |
| `azd-app` | This extension's own MCP server (agents, skills, checkpoints) |
| `microsoft-learn` | Microsoft Learn documentation search |
| `context7` | Library documentation lookup |
| `playwright` | Browser automation and E2E testing |

## Architecture

### How It Works

When you run `azd copilot`, the Go extension bootstraps a full AI development environment and hands off to GitHub Copilot CLI:

```mermaid
flowchart LR
    User["👤 User"] --> AZD["azd copilot"]

    subgraph Extension ["Go CLI Extension"]
        AZD --> Context["Detect Project Context\nazure.yaml, env, services"]
        Context --> Install["Install Agents & Skills\n~/.azd/copilot/"]
        Install --> MCP["Configure MCP Servers\n~/.copilot/mcp-config.json"]
    end

    MCP --> Copilot["GitHub Copilot CLI"]

    subgraph Session ["Copilot Session"]
        Copilot --> Agents["16 Agents"]
        Copilot --> Skills["28 Skills"]
        Copilot --> Servers["6 MCP Servers"]
    end

    Servers --> Azure["☁️ Azure"]

    style Extension fill:#0078D4,color:#fff
    style Session fill:#24292f,color:#fff
```

### GitHub Copilot CLI + MCP Servers

Copilot CLI acts as the AI brain. MCP servers give it tools to interact with Azure, your project, documentation, and the browser:

```mermaid
flowchart TB
    Copilot["GitHub Copilot CLI\n(AI reasoning + tool calling)"]

    Copilot --> azure["@azure/mcp\nAzure resource operations"]
    Copilot --> azd_mcp["azd mcp server\nAzure Developer CLI"]
    Copilot --> app["azd copilot mcp serve\nExtension MCP server"]
    Copilot --> learn["microsoft-learn\nDocumentation search"]
    Copilot --> c7["context7\nLibrary docs lookup"]
    Copilot --> pw["playwright\nBrowser automation"]

    azure --> Azure["☁️ Azure\nResources, subscriptions,\ndeployments"]
    azd_mcp --> AZD["Azure Developer CLI\nazure.yaml, environments,\nprovision & deploy"]
    app --> Ext["Extension Data\ncheckpoints, specs,\nagent/skill metadata"]

    style Copilot fill:#24292f,color:#fff
    style Azure fill:#0078D4,color:#fff
```

### Agents & Skills

Agents are domain experts that Copilot delegates to. Skills are step-by-step playbooks that agents follow. The **azure-manager** agent coordinates all others:

```mermaid
flowchart TB
    Manager["azure-manager\n(Coordinator)"]

    Manager --> Arch["azure-architect\nInfra & Bicep"]
    Manager --> Dev["azure-dev\nApp code & APIs"]
    Manager --> Data["azure-data\nDatabases & schemas"]
    Manager --> AI["azure-ai\nAI services & RAG"]
    Manager --> Sec["azure-security\nSecurity & identity"]
    Manager --> DevOps["azure-devops\nCI/CD & observability"]
    Manager --> Fin["azure-finance\nCost optimization"]
    Manager --> Quality["azure-quality\nTesting & review"]
    Manager --> More["+ 7 more agents\ndocs, product, design,\nanalytics, marketing,\nsupport, compliance"]

    Dev --> S1["Skills: azure-prepare\nazure-deploy, azure-validate"]
    Sec --> S2["Skills: azure-compliance\nentra-app-registration"]
    AI --> S3["Skills: azure-ai\nmicrosoft-foundry"]
    DevOps --> S4["Skills: azure-diagnostics\nazure-observability"]

    style Manager fill:#0078D4,color:#fff
```

### Build Flow

The `azd copilot build` command goes from description to deployed app in two phases:

```mermaid
flowchart LR
    Desc["📝 Description"] --> Spec

    subgraph Phase1 ["Phase 1: Spec Generation"]
        Spec["Generate Spec\n(azure-manager agent)"] --> Review["👤 User Review\ndocs/spec.md"]
    end

    Review -->|"azd copilot build --approve"| Build

    subgraph Phase2 ["Phase 2: Build & Deploy"]
        Build["Design\nazure.yaml, architecture"] --> Develop["Develop\nbackend, frontend, DB"]
        Develop --> Test["Quality\ntests, lint, security"]
        Test --> Infra["Infrastructure\nBicep, CI/CD, docs"]
        Infra --> Deploy["Deploy\nazd up"]
    end

    Deploy --> Live["🚀 Live on Azure"]

    style Phase1 fill:#24292f,color:#fff
    style Phase2 fill:#0078D4,color:#fff
```

### Upstream Skill Sync

Skills from [microsoft/GitHub-Copilot-for-Azure](https://github.com/microsoft/GitHub-Copilot-for-Azure) are synced into this repo. Local source uses exact sync; remote uses smart merge to preserve local edits:

```mermaid
flowchart LR
    subgraph Sources ["Source"]
        Local["📂 Local Clone\n(exact sync)"]
        Remote["🌐 GitHub Repo\n(smart merge)"]
    end

    Local --> Sync
    Remote --> Sync

    Sync["mage SyncSkills"] --> Target["ghcp4a-skills/\n18 upstream skills"]
    Target --> Embed["go:embed\nCompiled into binary"]
    Embed --> Install["~/.azd/copilot/skills/\nInstalled at runtime"]

    Custom["skills/\n10 custom skills"] --> Embed

    style Sources fill:#24292f,color:#fff
```

### Project Structure

```
cli/
├── src/
│   ├── cmd/copilot/           # CLI commands (cobra)
│   │   ├── main.go            # Entry point, root command, session launcher
│   │   └── commands/          # Subcommands (build, agents, skills, etc.)
│   └── internal/
│       ├── assets/            # Embedded agents & skills
│       │   ├── agents/        # 16 agent definitions (*.md)
│       │   ├── ghcp4a-skills/ # Upstream skills (synced from GitHub)
│       │   └── skills/        # Custom skills
│       ├── copilot/           # Copilot CLI launcher & MCP config
│       ├── checkpoint/        # Build checkpoint management
│       ├── spec/              # Project spec generation
│       ├── cache/             # Caching utilities
│       └── logging/           # Structured logging
├── magefile.go                # Build automation (mage)
└── extension.yaml             # azd extension manifest
```

## Development

### Prerequisites

- [Go 1.25+](https://golang.org/dl/)
- [Azure Developer CLI (azd)](https://learn.microsoft.com/azure/developer/azure-developer-cli/install-azd)
- [Mage](https://magefile.org/) (optional, for build automation)

### Build & Test

```bash
cd cli

# Build, lint, test, and install
mage           # runs All (fmt → lint → test → build → install)

# Or individually
mage build     # build + install to ~/.azd/extensions/
mage test      # run tests
mage lint      # run golangci-lint
mage fmt       # format code

# Without mage
go build ./...
go test ./...
```

### Upstream Skill Workflow

```bash
# Pull latest skills from upstream main (default)
mage SyncSkills

# Sync from a local clone of the upstream repo
mage SyncSkills C:\code\GitHub-Copilot-for-Azure

# Sync from a different repo or branch
mage SyncSkills https://github.com/user/fork.git
mage SyncSkills https://github.com/user/fork.git@my-branch

# Contribute your skill changes back upstream
mage ContributeSkills
```

The sync uses **smart merge** — your local changes are preserved, new upstream files are added, and only unmodified files are updated.

### Adding a Custom Skill

Create a directory in `cli/src/internal/assets/skills/` with a `SKILL.md`:

```
skills/my-skill/
├── SKILL.md           # Required — YAML frontmatter (name, description) + instructions
├── references/        # Optional — additional docs
└── assets/            # Optional — templates
```

Then run `mage UpdateCounts` to update agent/skill counts across the project.

## 🔗 azd Extensions

azd copilot is part of a suite of Azure Developer CLI extensions by [Jon Gallant](https://github.com/jongio).

| Extension | Description | Website |
|-----------|-------------|---------|
| **[azd app](https://github.com/jongio/azd-app)** | Run Azure apps locally with auto-dependencies, dashboard, and AI debugging | [jongio.github.io/azd-app](https://jongio.github.io/azd-app/) |
| **[azd copilot](https://github.com/jongio/azd-copilot)** | AI-powered Azure development with 16 agents and 28 skills | [jongio.github.io/azd-copilot](https://jongio.github.io/azd-copilot/) |
| **[azd exec](https://github.com/jongio/azd-exec)** | Execute scripts with azd environment context and Key Vault integration | [jongio.github.io/azd-exec](https://jongio.github.io/azd-exec/) |
| **[azd rest](https://github.com/jongio/azd-rest)** | Authenticated REST API calls with automatic scope detection | [jongio.github.io/azd-rest](https://jongio.github.io/azd-rest/) |

🌐 **Extension Hub**: [jongio.github.io/azd-extensions](https://jongio.github.io/azd-extensions/) — Browse all extensions, quick install, and registry info.

## License

MIT
