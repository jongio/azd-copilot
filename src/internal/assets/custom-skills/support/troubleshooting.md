---
name: troubleshooting
description: Systematic troubleshooting guides for Azure applications
agent: support
---

# Troubleshooting Guides

## Purpose

Create systematic troubleshooting procedures for Azure-deployed applications.

## Troubleshooting Framework

### 1. Triage Template

```markdown
## Issue Triage

**Reported**: [Date/time]
**Reporter**: [Name/contact]
**Severity**: [Critical/High/Medium/Low]

### Symptoms
- [What the user observed]

### Impact
- [Number of users affected]
- [Business impact]

### Environment
- Azure region: [e.g., East US]
- Service: [e.g., Container Apps]
- Version: [e.g., v1.2.3]

### Recent Changes
- [Deployments, config changes, etc.]
```

### 2. Diagnostic Steps

```markdown
## Standard Diagnostic Workflow

### Step 1: Verify the Issue
- [ ] Reproduce the issue
- [ ] Confirm scope (one user vs all users)
- [ ] Note exact error messages

### Step 2: Check Azure Health
- [ ] [Azure Status](https://status.azure.com/)
- [ ] Service Health in portal
- [ ] Activity Log for recent events

### Step 3: Check Application Health
- [ ] Container App status/replicas
- [ ] Application Insights availability
- [ ] Recent deployments

### Step 4: Review Logs
- [ ] Container App console logs
- [ ] Application Insights exceptions
- [ ] Azure Activity Log

### Step 5: Check Dependencies
- [ ] Database connectivity
- [ ] Key Vault access
- [ ] External API status
```

## Common Issues & Solutions

### Container App Not Starting

```markdown
## Issue: Container App shows 0 replicas

### Symptoms
- App returns 502/503 errors
- No replicas running in portal

### Diagnostic Commands
```bash
# Check app status
az containerapp show -n myapp -g mygroup --query "properties.runningStatus"

# Check replica status
az containerapp replica list -n myapp -g mygroup

# Check logs
az containerapp logs show -n myapp -g mygroup --type console
```

### Common Causes

| Cause | Check | Fix |
|-------|-------|-----|
| Image pull failure | Logs show "ImagePullBackOff" | Verify ACR credentials |
| Startup crash | Logs show exception | Fix app code |
| Health check failure | Probe failing | Fix health endpoint |
| Resource limits | OOMKilled | Increase memory |
| Secret missing | "Key not found" | Add Key Vault reference |

### Resolution Steps
1. Check container logs for startup errors
2. Verify image exists in ACR
3. Check managed identity has ACR pull permission
4. Verify all secrets/env vars are set
5. Check resource limits aren't too low
```

### Database Connection Failures

```markdown
## Issue: Cannot connect to PostgreSQL

### Symptoms
- "Connection refused" or timeout errors
- App starts but fails on first DB query

### Diagnostic Commands
```bash
# Check PostgreSQL status
az postgres flexible-server show -n myserver -g mygroup --query "state"

# Check firewall rules
az postgres flexible-server firewall-rule list -n myserver -g mygroup

# Test connectivity from Cloud Shell
psql "host=myserver.postgres.database.azure.com dbname=mydb user=myuser"
```

### Common Causes

| Cause | Check | Fix |
|-------|-------|-----|
| Server stopped | Portal shows "Stopped" | Start the server |
| Firewall blocks | No rule for client IP | Add firewall rule |
| Wrong credentials | Auth errors in logs | Check Key Vault secret |
| Managed identity | RBAC not configured | Assign db role |
| SSL required | SSL error in logs | Add `sslmode=require` |

### Resolution Steps
1. Verify PostgreSQL server is running
2. Check firewall allows Container Apps subnet
3. For managed identity: verify AAD admin is set
4. Test connection string from local machine
5. Check Application Insights for detailed errors
```

### Authentication Failures

```markdown
## Issue: Users cannot sign in

### Symptoms
- Redirect loop on login
- "AADSTS" error codes
- 401 Unauthorized errors

### Diagnostic Steps
```bash
# Check app registration
az ad app show --id <app-id> --query "web.redirectUris"

# Check sign-in logs
# Portal: Azure AD > Sign-in logs > Filter by app
```

### Common AADSTS Errors

| Error Code | Meaning | Fix |
|------------|---------|-----|
| AADSTS50011 | Reply URL mismatch | Add correct redirect URI |
| AADSTS65001 | Consent required | Grant admin consent |
| AADSTS700016 | App not found | Check tenant and app ID |
| AADSTS7000218 | Invalid client secret | Rotate secret |

### Resolution Steps
1. Verify redirect URIs match exactly (including trailing slashes)
2. Check client ID and tenant ID in app config
3. Verify client secret hasn't expired
4. Check if admin consent is required
5. Test in incognito window (clear cookies)
```

### Key Vault Access Issues

```markdown
## Issue: Cannot retrieve secrets from Key Vault

### Symptoms
- "Forbidden" or 403 errors
- App fails to start with "secret not found"

### Diagnostic Commands
```bash
# Check managed identity
az containerapp identity show -n myapp -g mygroup

# Check Key Vault access policy
az keyvault show -n myvault --query "properties.accessPolicies"

# Test access
az keyvault secret show --vault-name myvault -n mysecret
```

### Common Causes

| Cause | Check | Fix |
|-------|-------|-----|
| No managed identity | Identity is null | Enable managed identity |
| No access policy | Policy missing | Add access policy |
| RBAC not assigned | No role assignment | Assign Key Vault Secrets User |
| Firewall blocks | Network rules | Allow Container Apps subnet |
| Wrong vault name | Typo in config | Fix vault reference |

### Resolution Steps
1. Verify Container App has managed identity enabled
2. For access policies: add GET permission for secrets
3. For RBAC: assign "Key Vault Secrets User" role
4. Check Key Vault network rules allow access
5. Verify secret exists in Key Vault
```

## Runbook Template

```markdown
# Runbook: [Issue Name]

## Overview
[Brief description of the issue and when this runbook applies]

## Prerequisites
- Azure CLI installed
- Contributor access to resource group
- Access to Application Insights

## Detection
How to identify this issue:
1. [Alert that fires]
2. [Symptoms to look for]

## Impact Assessment
| Severity | Criteria |
|----------|----------|
| Critical | [All users affected] |
| High | [Major feature broken] |
| Medium | [Minor feature affected] |
| Low | [Cosmetic/edge case] |

## Resolution Steps

### Step 1: [First action]
```bash
[Command to run]
```
Expected outcome: [What you should see]

### Step 2: [Second action]
[Instructions]

### Step 3: Verify Resolution
```bash
[Verification command]
```

## Escalation
If not resolved within [X minutes], escalate to:
- [Team/person]
- [Contact method]

## Post-Incident
1. Update this runbook if needed
2. Create follow-up ticket for root cause
3. Update alerts if detection was delayed
```

## Azure Support Integration

### Creating Support Tickets

```bash
# Via CLI
az support tickets create \
  --ticket-name "Container App not starting" \
  --severity "moderate" \
  --contact-first-name "John" \
  --contact-last-name "Doe" \
  --contact-email "john@example.com" \
  --contact-method "email" \
  --contact-timezone "Pacific Standard Time" \
  --description "Container App myapp in resource group mygroup showing 0 replicas since 2024-01-15 10:00 UTC" \
  --problem-classification "/providers/Microsoft.Support/services/xxx/problemClassifications/yyy"
```

### Severity Levels

| Azure Severity | Response Time | Use When |
|----------------|---------------|----------|
| Critical (Sev A) | 1 hour | Business-critical workload down |
| Severe (Sev B) | 4 hours | Significant impact, degraded |
| Moderate (Sev C) | 8 hours | Moderate impact |
| Minimal (Sev D) | 24 hours | General questions |
