---
name: avm-bicep-rules
description: |
  MANDATORY rules for Bicep infrastructure generation. Enforces Azure Verified Modules (AVM)
  from the Bicep public registry instead of raw resource declarations. Prefer azd pattern
  modules first, then resource modules, raw declarations only as last resort.
  USE FOR: Every Bicep generation, every infra/ folder creation, every IaC task.
  INVOKED AUTOMATICALLY by azure-manager, azure-architect, and azure-devops agents.
---

# AVM Bicep Rules

> **MANDATORY COMPLIANCE** — These rules apply to ALL Bicep infrastructure generation.
> When generating Bicep, you MUST use Azure Verified Modules (AVM) from the public registry
> instead of writing raw `resource` declarations. This document is the authoritative source
> for module selection priority and usage patterns.

## When to Apply

Apply these rules **every time** you:
- Generate Bicep infrastructure (`infra/main.bicep`, modules)
- Create new Azure resources in Bicep
- Modify existing Bicep templates
- Create `infra/` folder for any project

## Module Priority (Mandatory)

| Priority | Module Type | Registry Prefix | When to Use |
|----------|-------------|-----------------|-------------|
| 1️⃣ **First** | AZD Pattern Modules | `br/public:avm/ptn/azd/<name>:<version>` | Preferred for azd projects — pre-composed, opinionated, secure-by-default |
| 2️⃣ **Second** | Pattern Modules | `br/public:avm/ptn/<area>/<name>:<version>` | Multi-resource patterns (e.g., hub networking, AI platform) |
| 3️⃣ **Third** | Resource Modules | `br/public:avm/res/<provider>/<type>:<version>` | Individual resources with best-practice defaults |
| 4️⃣ **Last resort** | Raw `resource` | N/A | Only when NO AVM module exists for the resource type |

> ❌ **NEVER write raw `resource` declarations when an AVM module exists for that resource type.**

---

## Key AZD Pattern Modules

These are **preferred** for azd projects because they are pre-composed, opinionated stacks:

| Module | Registry Path | Use For |
|--------|---------------|---------|
| Container Apps Stack | `avm/ptn/azd/container-apps-stack` | Container Apps Environment + ACR + Log Analytics |
| ACR Container App | `avm/ptn/azd/acr-container-app` | Single Container App with ACR integration |
| Container App Upsert | `avm/ptn/azd/container-app-upsert` | Create or update a Container App |
| AKS | `avm/ptn/azd/aks` | AKS cluster for azd projects |
| AKS Automatic | `avm/ptn/azd/aks-automatic-cluster` | AKS Automatic cluster |
| Monitoring | `avm/ptn/azd/monitoring` | Log Analytics + App Insights for azd |
| Insights Dashboard | `avm/ptn/azd/insights-dashboard` | App Insights dashboard |
| APIM API | `avm/ptn/azd/apim-api` | API Management API for azd |

## Key Resource Modules

Use these when no AZD pattern module fits:

| Module | Registry Path | Use For |
|--------|---------------|---------|
| Container App | `avm/res/app/container-app` | Container App |
| Managed Environment | `avm/res/app/managed-environment` | Container Apps Environment |
| Web App / Function App | `avm/res/web/site` | App Service / Function App |
| App Service Plan | `avm/res/web/serverfarm` | App Service Plan |
| Container Registry | `avm/res/container-registry/registry` | ACR |
| Key Vault | `avm/res/key-vault/vault` | Key Vault |
| Storage Account | `avm/res/storage/storage-account` | Storage |
| Cognitive Services | `avm/res/cognitive-services/account` | Azure OpenAI / Cognitive Services |
| Cosmos DB | `avm/res/document-db/database-account` | Cosmos DB |
| SQL Server | `avm/res/sql/server` | Azure SQL Server |
| PostgreSQL | `avm/res/db-for-postgre-sql/flexible-server` | PostgreSQL Flexible Server |
| AKS | `avm/res/container-service/managed-cluster` | AKS |
| Log Analytics | `avm/res/operational-insights/workspace` | Log Analytics Workspace |
| App Insights | `avm/res/insights/component` | Application Insights |
| Redis | `avm/res/cache/redis` | Azure Cache for Redis |
| Service Bus | `avm/res/service-bus/namespace` | Service Bus Namespace |
| Event Grid | `avm/res/event-grid/topic` | Event Grid Topic |

---

## Usage Examples

### Container Apps (AZD Pattern — Preferred)

```bicep
// ✅ PREFERRED: AZD pattern module for the full Container Apps stack
module containerAppsStack 'br/public:avm/ptn/azd/container-apps-stack:0.1.0' = {
  name: 'container-apps-stack'
  scope: rg
  params: {
    containerAppsEnvironmentName: 'cae-${resourceToken}'
    containerRegistryName: 'cr${resourceToken}'
    logAnalyticsWorkspaceResourceId: monitoring.outputs.logAnalyticsWorkspaceResourceId
    location: location
  }
}
```

### Monitoring (AZD Pattern — Preferred)

```bicep
// ✅ PREFERRED: AZD pattern module for monitoring stack
module monitoring 'br/public:avm/ptn/azd/monitoring:0.1.0' = {
  name: 'monitoring'
  scope: rg
  params: {
    logAnalyticsName: 'log-${resourceToken}'
    applicationInsightsName: 'appi-${resourceToken}'
    location: location
    tags: tags
  }
}
```

### AKS (AZD Pattern — Preferred)

```bicep
// ✅ PREFERRED: AZD pattern module for AKS
module aks 'br/public:avm/ptn/azd/aks:0.1.0' = {
  name: 'aks-cluster'
  scope: rg
  params: {
    clusterName: 'aks-${resourceToken}'
    containerRegistryName: containerRegistry.outputs.name
    logAnalyticsName: monitoring.outputs.logAnalyticsWorkspaceName
    location: location
    tags: tags
  }
}
```

### Key Vault (Resource Module)

```bicep
// ✅ AVM resource module (when no pattern module fits)
module keyVault 'br/public:avm/res/key-vault/vault:0.11.0' = {
  name: 'key-vault'
  scope: rg
  params: {
    name: 'kv-${resourceToken}'
    location: location
    enableRbacAuthorization: true
    tags: tags
  }
}
```

### App Service (Resource Module)

```bicep
// ✅ AVM resource module for App Service Plan
module appServicePlan 'br/public:avm/res/web/serverfarm:0.4.0' = {
  name: 'app-service-plan'
  scope: rg
  params: {
    name: 'plan-${resourceToken}'
    location: location
    sku: { name: 'B1', tier: 'Basic' }
    reserved: true
    tags: tags
  }
}

// ✅ AVM resource module for Web App
module webApp 'br/public:avm/res/web/site:0.15.0' = {
  name: 'web-app'
  scope: rg
  params: {
    name: 'app-${resourceToken}'
    kind: 'app,linux'
    serverFarmResourceId: appServicePlan.outputs.resourceId
    location: location
    managedIdentities: { systemAssigned: true }
    httpsOnly: true
    tags: tags
  }
}
```

### ❌ WRONG: Raw resource when AVM exists

```bicep
// ❌ WRONG: Raw resource declaration when AVM module exists
resource keyVault 'Microsoft.KeyVault/vaults@2023-07-01' = {
  name: 'kv-${resourceToken}'
  location: location
  properties: {
    sku: { family: 'A', name: 'standard' }
    tenantId: subscription().tenantId
    enableRbacAuthorization: true
  }
}
```

---

## main.bicep Template

```bicep
targetScope = 'subscription'

@description('Name of the environment')
param environmentName string

@description('Location for all resources')
param location string

var resourceToken = take(uniqueString(subscription().id, environmentName, location), 6)
var tags = { 'azd-env-name': environmentName }

resource rg 'Microsoft.Resources/resourceGroups@2023-07-01' = {
  name: 'rg-${environmentName}'
  location: location
  tags: tags
}

// ✅ Use AZD pattern module for monitoring stack
module monitoring 'br/public:avm/ptn/azd/monitoring:0.1.0' = {
  name: 'monitoring'
  scope: rg
  params: {
    logAnalyticsName: 'log-${resourceToken}'
    applicationInsightsName: 'appi-${resourceToken}'
    location: location
    tags: tags
  }
}

// ✅ Use AZD pattern module for Container Apps stack
module containerAppsStack 'br/public:avm/ptn/azd/container-apps-stack:0.1.0' = {
  name: 'container-apps-stack'
  scope: rg
  params: {
    containerAppsEnvironmentName: 'cae-${resourceToken}'
    containerRegistryName: 'cr${resourceToken}'
    logAnalyticsWorkspaceResourceId: monitoring.outputs.logAnalyticsWorkspaceResourceId
    location: location
  }
}
```

## File Structure

```
infra/
├── main.bicep              # Main orchestration (uses AVM modules from Bicep registry)
├── main.parameters.json    # Environment parameters
├── abbreviations.json      # Resource naming conventions
└── modules/                # Only for custom logic not covered by AVM
    └── ...
```

> **Prefer AVM modules from `br/public:avm/...` over local modules in `./modules/`.** Only create local modules for custom orchestration logic that no AVM module covers.

---

## Discovering Modules

- **MCP tool**: Use `mcp_bicep_list_avm_metadata` to search for available modules
- **AZD patterns**: https://azure.github.io/Azure-Verified-Modules/indexes/bicep/bicep-pattern-modules/
- **Resource modules**: https://azure.github.io/Azure-Verified-Modules/indexes/bicep/bicep-resource-modules/
- **Registry**: All modules use `br/public:avm/...` prefix from the Bicep public registry

## Validation Checklist

Before proceeding to `azure-validate` or `azure-deploy`, verify:

- [ ] Every resource with an AVM module uses the AVM module (not raw `resource`)
- [ ] AZD pattern modules used where available (checked before `avm/res/*`)
- [ ] No local `./modules/` duplicating what AVM already provides
- [ ] Module versions are pinned (e.g., `:0.1.0`, not `:latest`)
