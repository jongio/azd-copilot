---
name: copilot-docs-updater
description: Update and maintain Copilot CLI & SDK documentation for azurecopilot. Use when checking for new Copilot versions, updating reference docs, or getting recommendations for better Azure integration.
license: MIT
---

# Copilot Documentation Updater

This skill helps maintain up-to-date documentation for GitHub Copilot CLI and SDK, and provides recommendations for improving azurecopilot's integration with these tools.

## When to Use This Skill

Invoke this skill when:
- Checking for new versions of Copilot CLI or SDK
- Updating reference documentation after a new release
- Getting recommendations for Azure-specific Copilot features
- Analyzing feature gaps between CLI and SDK
- Planning Copilot integration improvements

## Workflow

### Step 1: Run the Analysis Script

First, run the Node.js analysis script to gather version information:

```bash
node scripts/update-copilot-docs.mjs
```

Or for JSON output suitable for further processing:

```bash
node scripts/update-copilot-docs.mjs --output-json > /tmp/copilot-analysis.json
```

### Step 2: Review Analysis Results

The script outputs:
- **Version Status**: Current vs latest versions for CLI and SDK
- **Recent Changes**: Parsed release notes with categorized features
- **Feature Comparison**: Features unique to CLI vs shared with SDK
- **Azure Recommendations**: Prioritized action items for azurecopilot

### Step 3: Update Documentation

If updates are needed, create new documentation files:

1. For CLI updates, create: `docs/copilot/reference/copilot-cli-v{VERSION}.md`
2. For SDK updates, create: `docs/copilot/reference/copilot-cli-sdk-v{VERSION}.md`

Use the existing documentation structure as a template.

## Documentation Structure

### CLI Documentation Template

```markdown
# GitHub Copilot CLI Reference

> **CLI Version:** {VERSION}
> **Last Updated:** {DATE}
> **Status:** Public Preview
> **Source:** [github/copilot-cli](https://github.com/github/copilot-cli)

## Overview
## Installation
## Authentication
## Modes of Use
## Permissions System
## Command Line Options
## Slash Commands
## Keyboard Shortcuts
## Custom Instructions
## Custom Agents
## Agent Skills
## Hooks
## MCP Servers
## Copilot Memory
## Context Management
## Model Selection
## Configuration Files
## Session Management
## Security Considerations
## Release History
```

### SDK Documentation Template

```markdown
# GitHub Copilot CLI SDK Reference

> **SDK Versions:** Node.js {VERSION} | Python {VERSION}
> **CLI Version:** {CLI_VERSION}
> **Last Updated:** {DATE}
> **Source:** [github/copilot-sdk](https://github.com/github/copilot-sdk)

## Overview
## Installation
## Core Concepts
## Client API
## Session API
## Event Types
## Custom Tools
## Custom Agents
## Session Hooks
## User Input Handling
## Permission Handling
## Custom Providers (BYOK)
## MCP Server Integration
## Available Models
## Configuration Files
## Examples
```

## Azure Integration Recommendations

Based on analysis, focus on these integration areas:

### High Priority

1. **Azure MCP Server Configuration**
   - Pre-configure Azure MCP server in default setup
   - Document Azure authentication requirements
   - Create azd integration for automatic configuration

2. **Azure-Specialized Custom Agents**
   - Create `.github/agents/azure-devops.agent.md`
   - Create `.github/agents/azure-security.agent.md`
   - Create `.github/agents/azure-cost.agent.md`
   - Create `.github/agents/azure-architect.agent.md`

3. **Migrate Existing Skills**
   - Convert `skills/azure-deploy` to SKILL.md format
   - Convert `skills/azure-diagnostics` to SKILL.md format
   - Ensure compatibility with both CLI and coding agent

### Medium Priority

4. **Security Hooks for Azure**
   - Create `.github/hooks/azure-security.json`
   - Implement credential validation in preToolUse
   - Add audit logging for compliance

5. **SDK Integration in azcopilot CLI**
   - Use SDK for programmatic Copilot access
   - Implement custom Azure tools via SDK
   - Enable session management for complex workflows

6. **Copilot Memory Enablement**
   - Document how to enable Copilot Memory
   - Create initial custom instructions
   - Track improvements in agent quality

## Feature Gap Analysis

### CLI Features Not Fully in SDK

These CLI features should be exposed or documented for SDK users:

| Feature | CLI Status | SDK Status | Recommendation |
|---------|-----------|------------|----------------|
| Plan Mode | ✅ Built-in | ❌ Missing | Document workaround via prompts |
| Skills | ✅ Native | ⚠️ Custom tools | Map skills to custom tools |
| Hooks | ✅ Native | ⚠️ Callbacks | Use session hooks in SDK |
| Copilot Memory | ✅ Automatic | ❌ N/A | Server-side only |
| Trusted Directories | ✅ Built-in | ⚠️ Manual | Implement in SDK wrapper |

### SDK Features Unique Value

| Feature | Value for azurecopilot |
|---------|------------------------|
| Programmatic sessions | Automated Azure workflows |
| Custom tools | Azure resource management |
| BYOK providers | Use Azure OpenAI |
| Event streaming | Real-time deployment feedback |
| Multi-language | Node.js, Python, Go, .NET |

## Versioning Strategy

- Name documentation files with version: `copilot-cli-v0.0.402.md`
- Keep previous versions for reference
- Update version in file header and filename together
- Run analysis script weekly or before major releases

## Sources to Monitor

1. **CLI Releases**: https://github.com/github/copilot-cli/releases
2. **CLI Docs**: https://docs.github.com/en/copilot/concepts/agents/about-copilot-cli
3. **SDK npm**: https://www.npmjs.com/package/@github/copilot-sdk
4. **SDK PyPI**: https://pypi.org/project/github-copilot-sdk/
5. **SDK GitHub**: https://github.com/github/copilot-sdk

## Troubleshooting

### Script fails to fetch versions

```bash
# Check network access
curl -I https://api.github.com/repos/github/copilot-cli/releases

# Set GitHub token if rate limited
export GITHUB_TOKEN=your_token
```

### Analysis file not found

The script saves analysis to `docs/copilot/reference/.copilot-docs-analysis.json`.
Ensure the `docs/copilot/reference` directory exists.

### Version comparison issues

Versions must be semantic (X.Y.Z). Pre-release versions (X.Y.Z-N) are tracked separately.
