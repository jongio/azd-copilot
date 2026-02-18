# Work Routing — Azure Squad

How to decide who handles what.

## Routing Table

| Work Type | Route To | Examples |
|-----------|----------|----------|
| Full app build (multi-service) | architect + dev + data | Fan-out: phased orchestration via coordinator |
| Infrastructure / Bicep / IaC | architect | Bicep modules, azure.yaml, networking, identity, AVM |
| Application code / APIs | developer | Backend APIs, frontend UI, app logic, middleware |
| Database / data layer | data | Schema design, migrations, queries, Cosmos/Postgres |
| Security / RBAC / identity | security | Vulnerability scanning, RBAC, managed identity, compliance |
| CI/CD / monitoring / observability | devops | GitHub Actions, App Insights, alerts, SKU selection |
| Testing / quality / code review | quality | Unit tests, integration tests, reviews, refactoring |
| AI / ML / RAG / agents | ai | Azure OpenAI, AI Search, Foundry, prompt flows, embeddings |
| Analytics / dashboards / metrics | analytics | Usage tracking, KQL queries, workbooks, reporting |
| Compliance / auditing / governance | compliance | GDPR, SOC2, HIPAA assessment, policy compliance |
| UX / accessibility / design | design | WCAG compliance, accessibility audits, UI review |
| Documentation / ADRs / runbooks | docs | README, API docs, architecture decisions, runbooks |
| Cost / pricing / optimization | finance | Cost estimation, TCO analysis, waste identification |
| Marketing / positioning / comms | marketing | Landing pages, feature communication, positioning |
| Product / requirements / specs | product | User needs, acceptance criteria, feature specs |
| Troubleshooting / support / FAQs | support | Error messages, FAQ generation, onboarding |
| Deployment (azd up/deploy) | Coordinator | NEVER delegated — coordinator runs deployment |
| Session logging | Scribe | Automatic — never needs routing |

## Azure Skill References

| Skill | When Used | By Whom |
|-------|-----------|---------|
| avm-bicep-rules | Before ANY Bicep generation | architect, devops |
| secure-defaults | Before ANY code or Bicep generation | All agents |
| azure-prepare | Initialize project for Azure | Coordinator |
| azure-functions | When app uses Azure Functions | developer |
| container-app-acr-auth | Container App + ACR Bicep | architect |

## Rules

1. **Eager by default** — spawn all agents who could usefully start work, including anticipatory downstream work.
2. **Scribe always runs** after substantial work, always as background. Never blocks.
3. **Quick facts → coordinator answers directly.** Don't spawn an agent for simple questions.
4. **When two agents could handle it**, pick the one whose domain is the primary concern.
5. **"Team, ..." → fan-out.** Spawn all relevant agents in parallel.
6. **Anticipate downstream work.** If a feature is being built, spawn quality to write test cases simultaneously.
7. **Azure skills are mandatory.** Before generating infrastructure or code, reference the relevant skill.
