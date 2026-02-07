---
name: container-app-acr-auth
description: |
  Bicep patterns for Container App + ACR authentication. Covers managed identity
  with AcrPull role assignment, dependency ordering to avoid circular deps, and
  how azd deploy interacts with ACR image pull.
  USE FOR: Container App ACR auth, ACR pull authentication, managed identity ACR,
  AcrPull role assignment, container app image pull, ACR admin credentials,
  container app can't pull image, ACR authentication Bicep.
  DO NOT USE FOR: AKS ACR integration (use azure-networking), general RBAC (use azure-security).
---

# Container App + ACR Authentication

> Reference guide for configuring Container App authentication with Azure Container Registry.
> Use this when generating Bicep that involves Container Apps pulling images from ACR.

## How azd Deploy Works with ACR

Understanding this flow prevents most ACR auth issues:

1. `azd provision` creates infrastructure (ACR, Container App Environment, Container App)
2. `azd deploy` builds the Docker image, pushes it to ACR, then updates the Container App's image reference
3. **azd deploy does NOT configure ACR pull auth** — the Container App must already have permission to pull from ACR

This means your Bicep must configure ACR pull authentication during provisioning.

## Recommended Approach: System-Assigned Managed Identity + AcrPull Role

This is the **preferred, secure approach**. The key challenge is dependency ordering:
the Container App's principal ID is only known after creation, but the role assignment
needs that principal ID.

**Solution: Use `dependsOn` — create the Container App first with a placeholder image,
then assign the role.**

```bicep
// 1. Container Apps Environment (via AVM)
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

// 2. Container App with system-assigned identity (starts with a placeholder image)
module api 'br/public:avm/ptn/azd/acr-container-app:0.1.0' = {
  name: 'api'
  scope: rg
  params: {
    name: 'ca-api-${resourceToken}'
    containerAppsEnvironmentName: containerAppsStack.outputs.environmentName
    containerRegistryName: containerAppsStack.outputs.registryName
    location: location
    identityType: 'SystemAssigned'
    // Do NOT set registryIdentity here yet — role assignment happens below
    exists: false
    containerName: 'api'
    targetPort: 3000
    env: []
    tags: union(tags, { 'azd-service-name': 'api' })
  }
}

// 3. AcrPull role assignment AFTER Container App exists
// This uses the Container App's system-assigned identity principal ID
module acrPullRole 'br/public:avm/ptn/authorization/resource-role-assignment:0.1.1' = {
  name: 'acr-pull-role'
  scope: rg
  params: {
    principalId: api.outputs.systemAssignedMIPrincipalId
    roleDefinitionId: '7f951dda-4ed3-4680-a7ca-43fe172d538d' // AcrPull
    resourceId: containerAppsStack.outputs.registryId
    principalType: 'ServicePrincipal'
  }
}
```

### Key Outputs from AVM Modules

These are the **correct** output names (do NOT guess):

| Module | Output | Description |
|--------|--------|-------------|
| `container-apps-stack` | `.outputs.environmentName` | Container Apps Environment name |
| `container-apps-stack` | `.outputs.registryName` | ACR name |
| `container-apps-stack` | `.outputs.registryLoginServer` | ACR login server URL |
| `container-apps-stack` | `.outputs.registryId` | ACR resource ID |
| `acr-container-app` | `.outputs.uri` | Container App URI (**NOT** `fqdn`) |
| `acr-container-app` | `.outputs.systemAssignedMIPrincipalId` | System identity principal ID |
| `acr-container-app` | `.outputs.name` | Container App name |

> ⚠️ The output is `.outputs.uri` — NOT `.outputs.fqdn`. Using `fqdn` will cause a Bicep compile error.

## Alternative: ACR Admin Credentials (Simpler but Less Secure)

For prototypes or simple apps where managed identity overhead isn't worth it:

```bicep
module containerAppsStack 'br/public:avm/ptn/azd/container-apps-stack:0.1.0' = {
  name: 'container-apps-stack'
  scope: rg
  params: {
    containerAppsEnvironmentName: 'cae-${resourceToken}'
    containerRegistryName: 'cr${resourceToken}'
    containerRegistryAdminUserEnabled: true  // Enable admin credentials
    logAnalyticsWorkspaceResourceId: monitoring.outputs.logAnalyticsWorkspaceResourceId
    location: location
  }
}
```

When `containerRegistryAdminUserEnabled: true` is set on the stack, `azd deploy` can
use the admin credentials to push and the Container App can pull using those same credentials.

> ⚠️ Admin credentials are less secure. Prefer managed identity for production apps.

## Common Pitfalls

### 1. Zone Redundancy Requires a Subnet
The `container-apps-stack` module defaults to `zoneRedundant: true`, which requires
a delegated subnet. For simple apps without a VNet:

```bicep
params: {
  zoneRedundant: false  // Disable if no subnet/VNet is configured
}
```

### 2. Circular Dependency: Identity → Role → Pull
You **cannot** configure `registryIdentity: 'system'` on the Container App AND
assign the AcrPull role in the same Bicep deployment if the Container App doesn't
exist yet. The principal ID isn't known until after creation.

**Solutions (pick one):**
- Use the two-step pattern above (create CA first, then role assignment with `dependsOn`)
- Use ACR admin credentials instead of managed identity
- Use a user-assigned managed identity (create identity first, assign role, then reference in CA)

### 3. Stuck Container App Provisioning
If a Container App gets stuck in `InProgress` state after a failed image pull:
- Wait up to 5 minutes — Azure will eventually timeout the failed revision
- If still stuck: `az containerapp delete -n <name> -g <rg> --yes` then re-provision
- Do NOT keep retrying `azd deploy` on a stuck Container App — it will queue behind the failed operation

### 4. main.parameters.json Location
`main.parameters.json` MUST be in the `infra/` directory (not project root).
The standard azd convention is `infra/main.parameters.json`.

## Bicep Outputs for azd Deploy

azd needs certain outputs from your Bicep to deploy services. Add these to `main.bicep`:

```bicep
// Required for azd deploy to find the Container App
output AZURE_CONTAINER_REGISTRY_ENDPOINT string = containerAppsStack.outputs.registryLoginServer
output API_BASE_URL string = api.outputs.uri
```
