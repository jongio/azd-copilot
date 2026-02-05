---
name: azure-capabilities
description: Azure service capability matrix for feasibility assessment
agent: product
---

# Azure Capabilities

## Purpose

Evaluate Azure service capabilities to ensure requirements are feasible.

## Service Capability Matrix

### Compute Services

| Service | Use Case | Limits | Pricing Model |
|---------|----------|--------|---------------|
| **Container Apps** | Microservices, APIs | 100 replicas, 4 vCPU/8GB per replica | Consumption (vCPU-s, memory-s) |
| **App Service** | Web apps, APIs | Up to 100 instances | Dedicated (plan-based) |
| **Functions** | Event-driven, serverless | 5 min timeout (consumption) | Consumption or Premium |
| **AKS** | Complex orchestration | Varies by node size | Node-based |

### Database Services

| Service | Use Case | Limits | Pricing Model |
|---------|----------|--------|---------------|
| **PostgreSQL Flexible** | Relational data | 64 vCores, 512 GB RAM | vCore + storage |
| **Cosmos DB** | Global distribution | Unlimited (with partitioning) | RU/s + storage |
| **SQL Database** | Enterprise SQL | 128 vCores | DTU or vCore |
| **Redis Cache** | Caching, sessions | 120 GB (Premium) | Tier-based |

### AI Services

| Service | Use Case | Limits | Pricing Model |
|---------|----------|--------|---------------|
| **Azure OpenAI** | GPT models | TPM quotas vary by model | Per 1K tokens |
| **AI Search** | Vector + keyword search | 50 indexes (S1) | Tier + storage |
| **Document Intelligence** | Document processing | 15 pages/request | Per page |

### Integration Services

| Service | Use Case | Limits | Pricing Model |
|---------|----------|--------|---------------|
| **Service Bus** | Messaging | 100 MB message | Tier-based |
| **Event Grid** | Event routing | 10 MB event | Per operation |
| **Event Hubs** | Streaming | 1 MB event | Throughput units |

## Regional Availability Check

```bash
# Check service availability in a region
az provider show --namespace Microsoft.App --query "resourceTypes[?resourceType=='containerApps'].locations" -o table

# Check quota
az quota show --resource-name "standardDv3Family" --scope "/subscriptions/{sub}/providers/Microsoft.Compute/locations/eastus"
```

## Feasibility Assessment Template

```markdown
## Feasibility Assessment: [Feature Name]

### Requirements
- [List key requirements]

### Proposed Azure Services
| Requirement | Azure Service | Feasibility |
|-------------|---------------|-------------|
| | | ✅ / ⚠️ / ❌ |

### Concerns
- [ ] Regional availability
- [ ] Quota limits
- [ ] Pricing impact
- [ ] Preview vs GA status

### Alternatives Considered
| Option | Pros | Cons |
|--------|------|------|
| | | |

### Recommendation
[Go / No-Go / Needs clarification]
```

## Common Constraints

### Hard Limits (Cannot Exceed)
- Container Apps: 4 vCPU, 8 GB memory per replica
- Functions Consumption: 5 minute timeout
- Service Bus: 100 MB message size
- Event Grid: 1 MB event size
- Blob Storage: 4.75 TB single blob (block blob)

### Soft Limits (Can Request Increase)
- Subscription vCPU quotas
- Azure OpenAI TPM quotas
- Storage account count per subscription
- Cosmos DB RU/s

## Preview Feature Policy

⚠️ **Preview features should NOT be used in production unless:**
1. Business explicitly accepts the risk
2. Fallback plan exists
3. Feature is close to GA (public preview, not private)

Check feature status:
```bash
az feature list --namespace Microsoft.App --query "[?properties.state=='Registered']"
```

## Cost Estimation Quick Reference

| Service | Cost Driver | Estimation Formula |
|---------|-------------|-------------------|
| Container Apps | vCPU-seconds, memory-seconds | `replicas × hours × (vCPU_price + memory_price)` |
| PostgreSQL | vCores, storage | `vCores × hours + storage_GB × rate` |
| Azure OpenAI | Tokens | `(input_tokens + output_tokens) / 1000 × rate` |
| Blob Storage | Storage, transactions | `storage_GB × rate + transactions × rate` |

See finance role for detailed cost estimation.
