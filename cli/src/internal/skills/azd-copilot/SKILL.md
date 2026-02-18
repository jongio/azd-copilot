---
name: azd-copilot
description: |
  AI-powered Azure development assistant built on GitHub Copilot CLI. Provides interactive
  sessions, automated app building, code review, optimization, diagnostics, and 16 specialized
  Azure agents. USE FOR: build Azure apps from descriptions, AI code review, cost optimization,
  deployment troubleshooting, security review, interactive Azure development, copilot sessions,
  checkpoint resume. DO NOT USE FOR: non-Azure projects, manual CLI operations (use azd directly),
  simple git operations.
---

# azd-copilot Extension

AI-powered Azure development assistant that extends the Azure Developer CLI (`azd`) with
GitHub Copilot CLI integration, 16 specialized Azure agents, and 29+ focused skills.

## When to Use

- Building Azure applications from natural language descriptions
- AI-assisted code review, security audits, and optimization
- Diagnosing deployment failures and production issues
- Interactive Azure development sessions with checkpoint/resume
- Generating infrastructure-as-code (Bicep/Terraform) with best practices

## Commands

### Interactive Session (default)

```bash
azd copilot                          # Start interactive session
azd copilot -p "deploy a Python API" # Start with a prompt
azd copilot --resume                 # Resume last session
azd copilot --agent azure-architect  # Use a specific agent
azd copilot --yolo                   # Auto-approve all actions
azd copilot -m claude-sonnet-4       # Use a specific model
```

### Build (2-Phase App Creation)

```bash
azd copilot build -p "React + FastAPI on Azure Container Apps"
azd copilot build --resume           # Resume from checkpoint
```

The build command runs two phases:
1. **Spec phase**: Generates a detailed application specification for review/approval
2. **Build phase**: Implements the spec with infrastructure, application code, and deployment config

### Quick Actions

| Command | Description |
|---------|-------------|
| `azd copilot init` | Initialize a new Azure project with AI guidance |
| `azd copilot review` | AI code review of current changes |
| `azd copilot fix` | Diagnose and fix issues in the codebase |
| `azd copilot optimize` | Suggest performance and cost optimizations |
| `azd copilot diagnose` | Troubleshoot deployment or runtime issues |

### Management Commands

| Command | Description |
|---------|-------------|
| `azd copilot agents` | List available agents |
| `azd copilot skills` | List available skills |
| `azd copilot sessions` | Manage copilot sessions |
| `azd copilot checkpoints` | List and manage build checkpoints |
| `azd copilot spec` | View or edit the current build spec |
| `azd copilot context` | Show project context information |
| `azd copilot mcp` | Manage MCP server configuration |
| `azd copilot version` | Show version information |

## 16 Specialized Agents

| Agent | Specialization |
|-------|---------------|
| `squad` | Azure Squad coordinator — routes work to specialized agents (default) |
| `azure-architect` | Infrastructure design and architecture decisions |
| `azure-dev` | Application code — backend, frontend, data layer |
| `azure-devops` | CI/CD, deployment, reliability, observability |
| `azure-data` | Database selection, schema design, query optimization |
| `azure-ai` | AI services, agent frameworks, RAG, model deployment |
| `azure-security` | Code security, infrastructure security, identity & auth |
| `azure-finance` | Cost estimation, optimization, waste identification |
| `azure-quality` | Testing, code review, refactoring, package evaluation |
| `azure-docs` | README, API docs, ADRs, runbooks |
| `azure-design` | WCAG compliance, accessibility audits, UI review |
| `azure-product` | Requirements, user stories, acceptance criteria |
| `azure-marketing` | Positioning, landing pages, competitive analysis |
| `azure-analytics` | Usage analytics, dashboards, metrics design |
| `azure-compliance` | Framework assessment (GDPR, SOC2, HIPAA), gap analysis |
| `azure-support` | Troubleshooting, FAQ generation, error messages |

## Build Flow

1. Run `azd copilot build -p "description"`
2. Spec is generated and presented for review
3. Approve or edit the spec
4. Build proceeds through phases with automatic checkpoints
5. Resume anytime with `azd copilot build --resume`

Checkpoints save progress at key milestones so builds can survive interruptions.

## MCP Servers

azd-copilot configures these MCP servers for Copilot CLI:

- **azure** — Azure resource management and queries
- **azd** — Azure Developer CLI operations
- **azd-app** — Application-level project context
- **microsoft-learn** — Official Microsoft documentation search
- **context7** — Library documentation lookup
- **playwright** — Browser automation for testing

## Key Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--prompt` | `-p` | Run with a specific prompt |
| `--resume` | `-r` | Resume the last session |
| `--yolo` | `-y` | Auto-approve all actions |
| `--agent` | `-a` | Use a specific agent |
| `--model` | `-m` | Use a specific AI model |
| `--add-dir` | | Additional directories to include |
| `--verbose` | `-v` | Verbose output |
| `--debug` | | Enable debug logging |
| `--cwd` | `-C` | Set working directory |

## Common Workflows

### New Azure App from Scratch

```bash
azd copilot build -p "Node.js API with Cosmos DB on Container Apps"
```

### Review and Harden Existing App

```bash
azd copilot review
azd copilot --agent azure-security -p "audit this app for vulnerabilities"
azd copilot optimize
```

### Troubleshoot a Failed Deployment

```bash
azd copilot diagnose
azd copilot --agent azure-devops -p "fix the CI/CD pipeline"
```

### Cost Optimization

```bash
azd copilot --agent azure-finance -p "analyze costs and suggest savings"
```
