---
name: support
description: Support and troubleshooting skills for Azure applications
---

# Support Skills

Skills for customer support, troubleshooting, and documentation.

## Available Skills

| Skill | Description | When to Use |
|-------|-------------|-------------|
| [troubleshooting](troubleshooting.md) | Diagnostic guides and procedures | Investigating issues |
| [error-messages](error-messages.md) | User-friendly error messages | Error handling |
| [faq](faq.md) | FAQ content generation | Documentation |

## Azure Specialization

The support role understands Azure-specific troubleshooting:

- **Azure Portal**: Resource navigation, Activity Log, Diagnose & Solve
- **Azure CLI**: `az` commands for diagnostics
- **Azure Monitor**: Log Analytics queries, metrics
- **Azure Support**: Creating support tickets, severity levels

## Quick Reference

### Troubleshooting Workflow

```
1. Reproduce
   └─ Confirm the issue and gather details

2. Isolate
   └─ Identify affected components

3. Diagnose
   └─ Check logs, metrics, and configuration

4. Resolve
   └─ Apply fix or workaround

5. Document
   └─ Update runbooks and knowledge base
```

### Common Azure Diagnostics

| Issue | First Check | Tool |
|-------|-------------|------|
| App not responding | Container App status | Portal / `az containerapp show` |
| 5xx errors | Application Insights | KQL: `requests \| where resultCode >= 500` |
| Slow performance | Metrics | App Insights Performance blade |
| Auth failures | Entra ID sign-in logs | Portal: AAD > Sign-in logs |
| Missing secrets | Key Vault access | Portal: Key Vault > Access policies |
