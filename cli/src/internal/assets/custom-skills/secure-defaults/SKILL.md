---
name: secure-defaults
description: |
  MANDATORY security rules for Azure project creation. Enforces managed identity,
  RBAC role assignments, and credential-free authentication for all generated
  infrastructure and application code. This skill OVERRIDES insecure patterns
  from upstream service references (connection strings, account keys, admin passwords).
  USE FOR: Every new Azure project, every Bicep generation, every SDK code generation.
  INVOKED AUTOMATICALLY by azure-manager, azure-architect, and azure-dev agents.
---

# Secure Defaults

> **MANDATORY COMPLIANCE** — These rules override any conflicting patterns from upstream
> service references. When a service reference shows a connection string or key-based
> pattern, use the secure alternative from this document instead.

## When to Apply

Apply these rules **every time** you:
- Generate Bicep / Terraform infrastructure
- Write application code that connects to Azure services
- Configure environment variables for Azure service access
- Create Dockerfiles or container configurations

## Azure Verified Modules (AVM)

When generating Bicep, **always use AVM modules** from the Bicep public registry. The security patterns in this document apply ON TOP of AVM modules. Priority order:

1. **`avm/ptn/azd/*`** — AZD pattern modules (preferred, secure-by-default)
2. **`avm/ptn/*`** — Other pattern modules
3. **`avm/res/*`** — Resource modules
4. **Raw `resource`** — Last resort only

AVM modules include many security best practices by default, but you MUST still verify the banned patterns below are not present.

## Banned Patterns

> ⛔ **NEVER generate any of the following. These are security violations.**

| Banned Pattern | Why | Secure Alternative |
|---------------|-----|-------------------|
| `listKeys()` | Extracts storage account keys into Bicep outputs | Use `AzureWebJobsStorage__accountName` + RBAC roles |
| `listCredentials()` | Extracts ACR admin passwords | Use `identity: 'system'` in registry config + `AcrPull` role |
| `administratorLogin` / `administratorLoginPassword` | SQL password auth | Use `azureADOnlyAuthentication: true` with Entra ID |
| `COSMOS_CONNECTION_STRING` | Contains primary key | Use `COSMOS_ENDPOINT` with `DefaultAzureCredential` |
| `AZURE_STORAGE_CONNECTION_STRING` | Contains account key | Use `AZURE_STORAGE_ACCOUNT` with `DefaultAzureCredential` |
| `SQL_CONNECTION_STRING` with password | Contains SQL password | Use `SQL_SERVER` + `SQL_DATABASE` with managed identity |
| `SERVICEBUS_CONNECTION_STRING` | Contains SAS key | Use `SERVICEBUS_NAMESPACE` with `DefaultAzureCredential` |
| `from_connection_string()` / `fromConnectionString()` | Key-based SDK init | Use `DefaultAzureCredential` constructor |
| `AzureWebJobsStorage` with `AccountKey=` | Storage key in app setting | Use `AzureWebJobsStorage__accountName` (identity-based) |
| ACR `username` / `passwordSecretRef` | Admin credentials | Use managed identity with `AcrPull` role |

---

## Required Patterns

### 1. Managed Identity on ALL Compute

Every App Service, Function App, Container App, and VM **MUST** have:

```bicep
identity: {
  type: 'SystemAssigned'
}
```

### 2. RBAC Role Assignments for Data Access

Every service-to-service connection **MUST** have a corresponding role assignment.
Do NOT just enable managed identity — you must also grant it permissions.

#### Cosmos DB

```bicep
var cosmosDataContributorRoleId = '00000000-0000-0000-0000-000000000002'

resource cosmosRoleAssignment 'Microsoft.DocumentDB/databaseAccounts/sqlRoleAssignments@2023-04-15' = {
  parent: cosmosAccount
  name: guid(cosmosAccount.id, principalId, cosmosDataContributorRoleId)
  properties: {
    roleDefinitionId: '${cosmosAccount.id}/sqlRoleDefinitions/${cosmosDataContributorRoleId}'
    principalId: principalId
    scope: cosmosAccount.id
  }
}
```

App settings: `COSMOS_ENDPOINT` only (no connection string).

#### Storage (Blob)

```bicep
resource storageBlobRole 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(storageAccount.id, principalId, 'Storage Blob Data Contributor')
  scope: storageAccount
  properties: {
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', 'ba92f5b4-2d11-453d-a403-e96b0029c9fe')
    principalId: principalId
    principalType: 'ServicePrincipal'
  }
}
```

App settings: `AZURE_STORAGE_ACCOUNT` only (no connection string).

#### Storage for Azure Functions

Functions require three RBAC roles for identity-based storage:

```bicep
// Storage Blob Data Owner
resource storageBlobRole 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(storageAccount.id, functionApp.id, 'Storage Blob Data Owner')
  scope: storageAccount
  properties: {
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', 'b7e6dc6d-f1e8-4753-8033-0f276bb0955b')
    principalId: functionApp.identity.principalId
    principalType: 'ServicePrincipal'
  }
}

// Storage Queue Data Contributor
resource storageQueueRole 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(storageAccount.id, functionApp.id, 'Storage Queue Data Contributor')
  scope: storageAccount
  properties: {
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', '974c5e8b-45b9-4653-ba55-5f855dd0fb88')
    principalId: functionApp.identity.principalId
    principalType: 'ServicePrincipal'
  }
}

// Storage Table Data Contributor
resource storageTableRole 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(storageAccount.id, functionApp.id, 'Storage Table Data Contributor')
  scope: storageAccount
  properties: {
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', '0a9a7e1f-b9d0-4cc4-a60d-0319b160aaa3')
    principalId: functionApp.identity.principalId
    principalType: 'ServicePrincipal'
  }
}
```

App setting: `AzureWebJobsStorage__accountName` (NOT `AzureWebJobsStorage` with key).

#### ACR (Container Registry)

```bicep
resource acrPullRole 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(containerRegistry.id, containerApp.id, 'AcrPull')
  scope: containerRegistry
  properties: {
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', '7f951dda-4ed3-4680-a7ca-43fe172d538d')
    principalId: containerApp.identity.principalId
    principalType: 'ServicePrincipal'
  }
}
```

Registry config uses `identity: 'system'` (NOT username/password).

#### Service Bus

```bicep
resource serviceBusSenderRole 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(serviceBus.id, principalId, 'Azure Service Bus Data Sender')
  scope: serviceBus
  properties: {
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', '69a216fc-b8fb-44d8-bc22-1f3c2cd27a39')
    principalId: principalId
    principalType: 'ServicePrincipal'
  }
}
```

App settings: `SERVICEBUS_NAMESPACE` only (no connection string).

#### SQL Database

```bicep
resource sqlServer 'Microsoft.Sql/servers@2022-05-01-preview' = {
  name: serverName
  location: location
  properties: {
    minimalTlsVersion: '1.2'
    administrators: {
      administratorType: 'ActiveDirectory'
      principalType: 'Group'
      login: entraAdminGroupName
      sid: entraAdminGroupObjectId
      tenantId: subscription().tenantId
      azureADOnlyAuthentication: true
    }
  }
}
```

After deployment, grant the app's managed identity access:
```sql
CREATE USER [my-container-app] FROM EXTERNAL PROVIDER;
ALTER ROLE db_datareader ADD MEMBER [my-container-app];
ALTER ROLE db_datawriter ADD MEMBER [my-container-app];
```

App settings: `SQL_SERVER` + `SQL_DATABASE` only (no connection string with password).

### 3. Container Apps Registry — Managed Identity

```bicep
resource containerApp 'Microsoft.App/containerApps@2023-05-01' = {
  identity: {
    type: 'SystemAssigned'
  }
  properties: {
    configuration: {
      registries: [
        {
          server: containerRegistry.properties.loginServer
          identity: 'system'
        }
      ]
      // NO secrets array with registry-password
    }
  }
}
```

### 4. Key Vault — RBAC Only, External Secrets Only

```bicep
resource keyVault 'Microsoft.KeyVault/vaults@2023-07-01' = {
  properties: {
    enableRbacAuthorization: true    // RBAC, not access policies
    enableSoftDelete: true
    enablePurgeProtection: true
  }
}
```

Key Vault is for **third-party secrets only** (external API keys, webhook tokens). For Azure
service-to-service access, use managed identity — do NOT store Azure connection strings in Key Vault.

---

## SDK Connection Patterns

When writing application code, always use `DefaultAzureCredential`:

### Cosmos DB

```typescript
import { CosmosClient } from "@azure/cosmos";
import { DefaultAzureCredential } from "@azure/identity";

const client = new CosmosClient({
  endpoint: process.env.COSMOS_ENDPOINT!,
  aadCredentials: new DefaultAzureCredential(),
});
```

### Storage

```typescript
import { BlobServiceClient } from "@azure/storage-blob";
import { DefaultAzureCredential } from "@azure/identity";

const client = new BlobServiceClient(
  `https://${process.env.AZURE_STORAGE_ACCOUNT}.blob.core.windows.net`,
  new DefaultAzureCredential()
);
```

### Service Bus

```typescript
import { ServiceBusClient } from "@azure/service-bus";
import { DefaultAzureCredential } from "@azure/identity";

const client = new ServiceBusClient(
  `${process.env.SERVICEBUS_NAMESPACE}.servicebus.windows.net`,
  new DefaultAzureCredential()
);
```

### SQL Database

```typescript
import sql from "mssql";

const config: sql.config = {
  server: process.env.SQL_SERVER!,
  database: process.env.SQL_DATABASE!,
  authentication: { type: "azure-active-directory-default" },
  options: { encrypt: true },
};
```

---

## Validation Checklist

Before proceeding to `azure-validate` or `azure-deploy`, verify:

- [ ] Every compute resource has `identity: { type: 'SystemAssigned' }`
- [ ] Every service-to-service connection has a `roleAssignments` resource
- [ ] No `listKeys()` or `listCredentials()` anywhere in Bicep
- [ ] No connection strings with embedded keys in app settings
- [ ] No `administratorLogin` / `administratorLoginPassword` on SQL servers
- [ ] ACR uses `identity: 'system'`, not username/password
- [ ] Functions use `AzureWebJobsStorage__accountName`, not key-based `AzureWebJobsStorage`
- [ ] Key Vault uses `enableRbacAuthorization: true`
- [ ] All SDK code uses `DefaultAzureCredential`, not `fromConnectionString()`
