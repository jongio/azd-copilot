# AVM Bicep Module Index

> Complete list of Azure Verified Modules from `br/public:avm/...`. Use this to find exact module paths instead of guessing.

## AZD Pattern Modules (`avm/ptn/azd/*`) — PREFERRED for azd projects

| Module Path | Use For |
|-------------|---------|
| `avm/ptn/azd/acr-container-app` | Single Container App with ACR |
| `avm/ptn/azd/aks` | AKS cluster for azd |
| `avm/ptn/azd/aks-automatic-cluster` | AKS Automatic cluster |
| `avm/ptn/azd/apim-api` | API Management API |
| `avm/ptn/azd/container-app-upsert` | Create/update Container App |
| `avm/ptn/azd/container-apps-stack` | Container Apps Env + ACR + Log Analytics |
| `avm/ptn/azd/insights-dashboard` | App Insights dashboard |
| `avm/ptn/azd/ml-ai-environment` | ML/AI environment |
| `avm/ptn/azd/ml-hub-dependencies` | ML Hub dependencies |
| `avm/ptn/azd/ml-project` | ML project |
| `avm/ptn/azd/monitoring` | Log Analytics + App Insights |

## Other Pattern Modules (`avm/ptn/*`)

### AI & ML
| Module Path | Use For |
|-------------|---------|
| `avm/ptn/ai-ml/ai-foundry` | AI Foundry |
| `avm/ptn/ai-ml/landing-zone` | AI/ML landing zone |
| `avm/ptn/ai-platform/baseline` | AI platform baseline |
| `avm/ptn/openai/cognitive-search` | OpenAI + Cognitive Search |
| `avm/ptn/openai/e2e-baseline` | OpenAI end-to-end baseline |
| `avm/ptn/sa/build-your-own-copilot` | Build your own copilot |
| `avm/ptn/sa/chat-with-your-data` | Chat with your data |
| `avm/ptn/sa/content-processing` | Content processing |
| `avm/ptn/sa/conversation-knowledge-mining` | Conversation knowledge mining |
| `avm/ptn/sa/customer-chatbot` | Customer chatbot |
| `avm/ptn/sa/document-knowledge-mining` | Document knowledge mining |
| `avm/ptn/sa/modernize-your-code` | Code modernization |
| `avm/ptn/sa/multi-agent-custom-automation-engine` | Multi-agent automation |

### App Hosting
| Module Path | Use For |
|-------------|---------|
| `avm/ptn/aca-lza/hosting-environment` | ACA landing zone hosting |
| `avm/ptn/app-service-lza/hosting-environment` | App Service landing zone |
| `avm/ptn/app/container-job-toolkit` | Container job toolkit |
| `avm/ptn/app/cosmos-db-account-container-app` | Cosmos DB + Container App |
| `avm/ptn/app/mongodb-cluster-container-app` | MongoDB + Container App |
| `avm/ptn/app/iaas-vm-cosmosdb-tier4` | IaaS VM + Cosmos DB tier4 |
| `avm/ptn/app/paas-ase-cosmosdb-tier4` | PaaS ASE + Cosmos DB tier4 |

### Authorization & Policy
| Module Path | Use For |
|-------------|---------|
| `avm/ptn/authorization/pim-role-assignment` | PIM role assignment |
| `avm/ptn/authorization/policy-assignment` | Policy assignment |
| `avm/ptn/authorization/policy-exemption` | Policy exemption |
| `avm/ptn/authorization/resource-role-assignment` | Resource-scoped role assignment |
| `avm/ptn/authorization/role-assignment` | Role assignment |
| `avm/ptn/authorization/role-definition` | Role definition |
| `avm/ptn/policy-insights/remediation` | Policy remediation |

### Networking
| Module Path | Use For |
|-------------|---------|
| `avm/ptn/network/hub-networking` | Hub-spoke networking |
| `avm/ptn/network/private-link-private-dns-zones` | Private Link DNS zones |
| `avm/ptn/network/virtual-wan` | Virtual WAN |
| `avm/ptn/network/vwan-connected-vnets` | VWAN connected VNets |

### Landing Zones & Governance
| Module Path | Use For |
|-------------|---------|
| `avm/ptn/alz/ama` | ALZ Azure Monitor Agent |
| `avm/ptn/alz/empty` | ALZ empty template |
| `avm/ptn/lz/sub-vending` | Subscription vending |
| `avm/ptn/mgmt-groups/subscription-placement` | Mgmt group sub placement |
| `avm/ptn/subscription/service-health-alerts` | Service health alerts |

### Security & Monitoring
| Module Path | Use For |
|-------------|---------|
| `avm/ptn/security/security-center` | Security Center / Defender |
| `avm/ptn/security/sentinel` | Microsoft Sentinel |
| `avm/ptn/monitoring/amba` | Azure Monitor baseline alerts |
| `avm/ptn/monitoring/amba-alz` | AMBA for ALZ |
| `avm/ptn/finops-toolkit/finops-hub` | FinOps hub |

### Data & Other
| Module Path | Use For |
|-------------|---------|
| `avm/ptn/data/private-analytical-workspace` | Private analytical workspace |
| `avm/ptn/deployment-script/import-image-to-acr` | Import image to ACR |
| `avm/ptn/deployment-script/create-kv-ssh-key-pair` | Create KV SSH key pair |
| `avm/ptn/deployment-script/private` | Private deployment script |
| `avm/ptn/dev-ops/cicd-agents-and-runners` | CI/CD agents and runners |
| `avm/ptn/dev-center/dev-box` | Dev Center Dev Box |
| `avm/ptn/virtual-machine-images/azure-image-builder` | Azure Image Builder |
| `avm/ptn/avd-lza/insights` | AVD insights |
| `avm/ptn/avd-lza/management-plane` | AVD management plane |
| `avm/ptn/avd-lza/networking` | AVD networking |
| `avm/ptn/avd-lza/session-hosts` | AVD session hosts |
| `avm/ptn/maintenance/azure-update-manager` | Azure Update Manager |
| `avm/ptn/lza-shared/data-services` | LZA shared data services |

---

## Resource Modules (`avm/res/*`)

### Compute
| Module Path | Use For |
|-------------|---------|
| `avm/res/compute/virtual-machine` | Virtual Machine |
| `avm/res/compute/virtual-machine-scale-set` | VM Scale Set |
| `avm/res/compute/availability-set` | Availability Set |
| `avm/res/compute/disk` | Managed Disk |
| `avm/res/compute/disk-encryption-set` | Disk Encryption Set |
| `avm/res/compute/gallery` | Compute Gallery |
| `avm/res/compute/image` | VM Image |
| `avm/res/compute/proximity-placement-group` | Proximity Placement Group |
| `avm/res/compute/ssh-public-key` | SSH Public Key |

### Containers & Kubernetes
| Module Path | Use For |
|-------------|---------|
| `avm/res/app/container-app` | Container App |
| `avm/res/app/job` | Container App Job |
| `avm/res/app/managed-environment` | Container Apps Environment |
| `avm/res/app/session-pool` | Container App Session Pool |
| `avm/res/container-instance/container-group` | Container Instance |
| `avm/res/container-registry/registry` | Container Registry (ACR) |
| `avm/res/container-service/managed-cluster` | AKS Managed Cluster |

### Web & App Service
| Module Path | Use For |
|-------------|---------|
| `avm/res/web/site` | Web App / Function App |
| `avm/res/web/site/config` | Web App config |
| `avm/res/web/site/slot` | Web App slot |
| `avm/res/web/serverfarm` | App Service Plan |
| `avm/res/web/hosting-environment` | App Service Environment |
| `avm/res/web/static-site` | Static Web App |
| `avm/res/web/connection` | API Connection |

### Databases
| Module Path | Use For |
|-------------|---------|
| `avm/res/document-db/database-account` | Cosmos DB |
| `avm/res/document-db/database-account/sql-database` | Cosmos SQL Database |
| `avm/res/document-db/database-account/sql-role-assignment` | Cosmos SQL role assignment |
| `avm/res/document-db/database-account/sql-role-definition` | Cosmos SQL role definition |
| `avm/res/document-db/database-account/table` | Cosmos Table |
| `avm/res/document-db/mongo-cluster` | MongoDB cluster (vCore) |
| `avm/res/sql/server` | Azure SQL Server |
| `avm/res/sql/server/database` | Azure SQL Database |
| `avm/res/sql/managed-instance` | SQL Managed Instance |
| `avm/res/sql/instance-pool` | SQL Instance Pool |
| `avm/res/db-for-postgre-sql/flexible-server` | PostgreSQL Flexible Server |
| `avm/res/db-for-my-sql/flexible-server` | MySQL Flexible Server |
| `avm/res/cache/redis` | Azure Cache for Redis |
| `avm/res/cache/redis-enterprise` | Redis Enterprise |
| `avm/res/kusto/cluster` | Azure Data Explorer (Kusto) |

### AI & Cognitive Services
| Module Path | Use For |
|-------------|---------|
| `avm/res/cognitive-services/account` | Azure OpenAI / Cognitive Services |
| `avm/res/search/search-service` | Azure AI Search |
| `avm/res/machine-learning-services/workspace` | ML Workspace (AI Foundry) |
| `avm/res/machine-learning-services/registry` | ML Registry |

### Storage
| Module Path | Use For |
|-------------|---------|
| `avm/res/storage/storage-account` | Storage Account |
| `avm/res/storage/storage-account/blob-service/container` | Blob Container |
| `avm/res/storage/storage-account/file-service/share` | File Share |
| `avm/res/storage/storage-account/queue-service/queue` | Queue |
| `avm/res/storage/storage-account/table-service/table` | Table |
| `avm/res/storage/storage-account/local-user` | Storage local user |
| `avm/res/storage/storage-account/management-policy` | Lifecycle management |

### Networking
| Module Path | Use For |
|-------------|---------|
| `avm/res/network/virtual-network` | Virtual Network |
| `avm/res/network/virtual-network/subnet` | Subnet |
| `avm/res/network/network-security-group` | NSG |
| `avm/res/network/public-ip-address` | Public IP |
| `avm/res/network/public-ip-prefix` | Public IP Prefix |
| `avm/res/network/private-endpoint` | Private Endpoint |
| `avm/res/network/private-link-service` | Private Link Service |
| `avm/res/network/private-dns-zone` | Private DNS Zone |
| `avm/res/network/load-balancer` | Load Balancer |
| `avm/res/network/application-gateway` | Application Gateway |
| `avm/res/network/application-gateway-web-application-firewall-policy` | App Gateway WAF Policy |
| `avm/res/network/application-security-group` | Application Security Group |
| `avm/res/network/azure-firewall` | Azure Firewall |
| `avm/res/network/firewall-policy` | Firewall Policy |
| `avm/res/network/bastion-host` | Bastion Host |
| `avm/res/network/nat-gateway` | NAT Gateway |
| `avm/res/network/network-interface` | Network Interface |
| `avm/res/network/network-manager` | Network Manager |
| `avm/res/network/network-watcher` | Network Watcher |
| `avm/res/network/network-security-perimeter` | Network Security Perimeter |
| `avm/res/network/route-table` | Route Table |
| `avm/res/network/dns-zone` | DNS Zone |
| `avm/res/network/dns-resolver` | DNS Resolver |
| `avm/res/network/dns-forwarding-ruleset` | DNS Forwarding Ruleset |
| `avm/res/network/virtual-network-gateway` | VPN/ExpressRoute Gateway |
| `avm/res/network/virtual-wan` | Virtual WAN |
| `avm/res/network/virtual-hub` | Virtual Hub |
| `avm/res/network/vpn-gateway` | VPN Gateway |
| `avm/res/network/vpn-site` | VPN Site |
| `avm/res/network/vpn-server-configuration` | VPN Server Config |
| `avm/res/network/p2s-vpn-gateway` | P2S VPN Gateway |
| `avm/res/network/express-route-circuit` | ExpressRoute Circuit |
| `avm/res/network/express-route-gateway` | ExpressRoute Gateway |
| `avm/res/network/express-route-port` | ExpressRoute Port |
| `avm/res/network/connection` | VPN/ER Connection |
| `avm/res/network/local-network-gateway` | Local Network Gateway |
| `avm/res/network/ddos-protection-plan` | DDoS Protection Plan |
| `avm/res/network/ip-group` | IP Group |
| `avm/res/network/front-door` | Front Door (classic) |
| `avm/res/network/front-door-web-application-firewall-policy` | Front Door WAF Policy |
| `avm/res/network/service-endpoint-policy` | Service Endpoint Policy |
| `avm/res/network/trafficmanagerprofile` | Traffic Manager |
| `avm/res/cdn/profile` | CDN / Front Door Profile |
| `avm/res/service-networking/traffic-controller` | Traffic Controller |

### Security & Identity
| Module Path | Use For |
|-------------|---------|
| `avm/res/key-vault/vault` | Key Vault |
| `avm/res/key-vault/vault/key` | Key Vault key |
| `avm/res/key-vault/vault/secret` | Key Vault secret |
| `avm/res/key-vault/vault/access-policy` | Key Vault access policy |
| `avm/res/key-vault/managed-hsm` | Managed HSM |
| `avm/res/managed-identity/user-assigned-identity` | User-Assigned Managed Identity |
| `avm/res/authorization/role-assignment` | Role Assignment (sub scope) |
| `avm/res/authorization/role-assignment/rg-scope` | Role Assignment (RG scope) |
| `avm/res/authorization/role-assignment/mg-scope` | Role Assignment (MG scope) |
| `avm/res/authorization/policy-assignment` | Policy Assignment |
| `avm/res/aad/domain-service` | Azure AD Domain Services |

### Monitoring & Observability
| Module Path | Use For |
|-------------|---------|
| `avm/res/operational-insights/workspace` | Log Analytics Workspace |
| `avm/res/operational-insights/cluster` | Log Analytics Cluster |
| `avm/res/insights/component` | Application Insights |
| `avm/res/insights/action-group` | Alert Action Group |
| `avm/res/insights/activity-log-alert` | Activity Log Alert |
| `avm/res/insights/metric-alert` | Metric Alert |
| `avm/res/insights/scheduled-query-rule` | Scheduled Query Rule |
| `avm/res/insights/diagnostic-setting` | Diagnostic Setting |
| `avm/res/insights/data-collection-endpoint` | Data Collection Endpoint |
| `avm/res/insights/data-collection-rule` | Data Collection Rule |
| `avm/res/insights/private-link-scope` | Monitor Private Link Scope |
| `avm/res/insights/webtest` | Availability Webtest |
| `avm/res/insights/autoscale-setting` | Autoscale Setting |
| `avm/res/portal/dashboard` | Portal Dashboard |
| `avm/res/dashboard/grafana` | Managed Grafana |

### Messaging & Events
| Module Path | Use For |
|-------------|---------|
| `avm/res/service-bus/namespace` | Service Bus Namespace |
| `avm/res/service-bus/namespace/queue` | Service Bus Queue |
| `avm/res/service-bus/namespace/topic` | Service Bus Topic |
| `avm/res/event-hub/namespace` | Event Hubs Namespace |
| `avm/res/event-hub/namespace/eventhub` | Event Hub |
| `avm/res/event-grid/domain` | Event Grid Domain |
| `avm/res/event-grid/namespace` | Event Grid Namespace |
| `avm/res/event-grid/topic` | Event Grid Topic |
| `avm/res/event-grid/system-topic` | Event Grid System Topic |
| `avm/res/signal-r-service/signal-r` | SignalR Service |
| `avm/res/signal-r-service/web-pub-sub` | Web PubSub |
| `avm/res/communication/communication-service` | Communication Services |
| `avm/res/communication/email-service` | Email Service |

### API Management
| Module Path | Use For |
|-------------|---------|
| `avm/res/api-management/service` | API Management |
| `avm/res/api-management/service/api` | APIM API |
| `avm/res/api-management/service/api-version-set` | APIM API Version Set |
| `avm/res/api-management/service/backend` | APIM Backend |
| `avm/res/api-management/service/policy` | APIM Policy |
| `avm/res/api-management/service/product` | APIM Product |
| `avm/res/api-management/service/named-value` | APIM Named Value |
| `avm/res/api-management/service/logger` | APIM Logger |
| `avm/res/api-management/service/subscription` | APIM Subscription |
| `avm/res/api-management/service/identity-provider` | APIM Identity Provider |
| `avm/res/api-management/service/authorization-server` | APIM Auth Server |
| `avm/res/api-management/service/cache` | APIM Cache |

### Data & Analytics
| Module Path | Use For |
|-------------|---------|
| `avm/res/data-factory/factory` | Data Factory |
| `avm/res/databricks/workspace` | Databricks Workspace |
| `avm/res/databricks/access-connector` | Databricks Access Connector |
| `avm/res/synapse/workspace` | Synapse Workspace |
| `avm/res/synapse/private-link-hub` | Synapse Private Link Hub |
| `avm/res/analysis-services/server` | Analysis Services |
| `avm/res/purview/account` | Purview Account |
| `avm/res/fabric/capacity` | Fabric Capacity |
| `avm/res/power-bi-dedicated/capacity` | Power BI Embedded |
| `avm/res/stream-analytics/streaming-job` | Stream Analytics |

### DevOps & Dev Tools
| Module Path | Use For |
|-------------|---------|
| `avm/res/dev-center/devcenter` | Dev Center |
| `avm/res/dev-center/project` | Dev Center Project |
| `avm/res/dev-center/network-connection` | Dev Center Network Connection |
| `avm/res/dev-ops-infrastructure/pool` | DevOps Managed Pool |
| `avm/res/dev-test-lab/lab` | DevTest Lab |

### Backup & Recovery
| Module Path | Use For |
|-------------|---------|
| `avm/res/recovery-services/vault` | Recovery Services Vault |
| `avm/res/data-protection/backup-vault` | Backup Vault |
| `avm/res/data-protection/resource-guard` | Resource Guard |

### Other Services
| Module Path | Use For |
|-------------|---------|
| `avm/res/automation/automation-account` | Automation Account |
| `avm/res/batch/batch-account` | Batch Account |
| `avm/res/bot-service/bot-service` | Bot Service |
| `avm/res/consumption/budget` | Budget |
| `avm/res/digital-twins/digital-twins-instance` | Digital Twins |
| `avm/res/elastic-san/elastic-san` | Elastic SAN |
| `avm/res/health-bot/health-bot` | Health Bot |
| `avm/res/healthcare-apis/workspace` | Healthcare APIs |
| `avm/res/load-test-service/load-test` | Azure Load Testing |
| `avm/res/logic/workflow` | Logic App |
| `avm/res/logic/integration-account` | Integration Account |
| `avm/res/maintenance/maintenance-configuration` | Maintenance Configuration |
| `avm/res/maintenance/configuration-assignment` | Maintenance Assignment |
| `avm/res/managed-services/registration-definition` | Lighthouse Registration |
| `avm/res/management/management-group` | Management Group |
| `avm/res/maps/account` | Azure Maps |
| `avm/res/net-app/net-app-account` | NetApp Files |
| `avm/res/relay/namespace` | Relay Namespace |
| `avm/res/resources/deployment-script` | Deployment Script |
| `avm/res/resources/resource-group` | Resource Group |
| `avm/res/resource-graph/query` | Resource Graph Query |
| `avm/res/virtual-machine-images/image-template` | Image Template |
| `avm/res/desktop-virtualization/host-pool` | AVD Host Pool |
| `avm/res/desktop-virtualization/application-group` | AVD Application Group |
| `avm/res/desktop-virtualization/workspace` | AVD Workspace |
| `avm/res/desktop-virtualization/scaling-plan` | AVD Scaling Plan |
| `avm/res/service-fabric/cluster` | Service Fabric |
| `avm/res/app-configuration/configuration-store` | App Configuration |
| `avm/res/alerts-management/action-rule` | Alert Action Rule |

---

> **Usage**: `module myResource 'br/public:<module-path>:<version>' = { ... }`
> **Versions**: Pin to a specific version. Check the [Bicep registry](https://github.com/Azure/bicep-registry-modules) for latest.
