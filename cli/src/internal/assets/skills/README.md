# Skills

This directory contains custom skills maintained in this repo. These are NOT synced from upstream.

For upstream Azure skills (synced from [GitHub-Copilot-for-Azure](https://github.com/microsoft/github-copilot-for-azure)), see `../ghcp4a-skills/`.

## Custom Skills

| Skill | Description |
|-------|-------------|
| `analytics` | Analytics and observability for Azure monitoring |
| `avm-bicep-rules` | **Mandatory** AVM module rules for Bicep — enforces Azure Verified Modules from registry, azd patterns preferred |
| `azure-functions` | Azure Functions development (HTTP, timer, queue triggers) |
| `copilot-docs-updater` | Update Copilot CLI & SDK documentation |
| `marketing` | Product positioning and marketplace listings |
| `product` | Requirements, user stories, acceptance criteria |
| `quality` | Code review, refactoring patterns, and Playwright E2E testing |
| `secure-defaults` | **Mandatory** security rules enforcing managed identity and RBAC for all generated infrastructure |
| `support` | Troubleshooting, FAQs, error messages |

## Adding a New Custom Skill

Create a new directory here with a `SKILL.md` file containing YAML frontmatter:

```
my-skill/
├── SKILL.md           # Required: YAML frontmatter + instructions
├── references/        # Optional: Additional documentation
└── assets/            # Optional: Templates, data files
```

After adding, run `mage updateCounts` to update counts across the project.
