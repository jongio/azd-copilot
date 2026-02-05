---
name: faq
description: Generate FAQ content for Azure applications
agent: support
---

# FAQ Generation

## Purpose

Create comprehensive FAQ content that reduces support burden.

## FAQ Structure

### Format

```markdown
## Category Name

### Question?

Answer in 1-3 paragraphs. Include:
- Direct answer to the question
- Additional context if needed
- Link to detailed documentation

[Learn more →](/docs/topic)
```

## Standard FAQ Categories

### Getting Started

```markdown
## Getting Started

### What is AzureCopilot?

AzureCopilot is an AI-powered application generator that creates production-ready Azure applications from natural language descriptions. It uses 16 specialized AI agents that work as a virtual engineering team to generate backend APIs, frontend UIs, Azure infrastructure, CI/CD pipelines, and comprehensive documentation.

### How do I get started?

1. Install the CLI: `npm install -g @azurecopilot/cli`
2. Sign in: `azcopilot auth login`
3. Create a project: `azcopilot build "Create a todo app with React and PostgreSQL"`

[View full getting started guide →](/docs/getting-started)

### What Azure services does AzureCopilot use?

AzureCopilot generates applications using:
- **Compute**: Azure Container Apps, Azure Functions, Static Web Apps
- **Data**: Azure PostgreSQL, Cosmos DB, Redis Cache
- **Security**: Azure Key Vault, Entra ID (Azure AD)
- **Observability**: Application Insights, Log Analytics
- **DevOps**: Azure Container Registry, GitHub Actions

The specific services depend on your application requirements.
```

### Pricing & Billing

```markdown
## Pricing & Billing

### Is AzureCopilot free?

AzureCopilot offers a free tier with 3 projects per month. The free tier includes all features but is limited in the number of projects you can generate.

**Note**: Azure infrastructure costs are billed separately by Microsoft. AzureCopilot helps you estimate these costs before deployment.

[View pricing →](/pricing)

### What does "Azure infrastructure costs billed separately" mean?

When AzureCopilot deploys your application to Azure, the Azure resources (Container Apps, PostgreSQL, etc.) incur charges from Microsoft. These charges are based on your Azure subscription and pricing tier.

AzureCopilot provides cost estimates before deployment so you know what to expect.

### How do I upgrade my plan?

Visit your account settings and select "Upgrade Plan". You can upgrade at any time and will be prorated for the remainder of your billing period.

[Manage subscription →](/settings/subscription)
```

### Features & Capabilities

```markdown
## Features & Capabilities

### What programming languages does AzureCopilot support?

AzureCopilot generates applications in:
- **Backend**: Node.js (TypeScript), .NET (C#), Python
- **Frontend**: React (TypeScript), Vue.js
- **Infrastructure**: Bicep, Azure CLI

### Can I customize the generated code?

Yes! All generated code is yours to modify. AzureCopilot generates standard, well-structured code following Azure best practices. You can:
- Modify any file after generation
- Add custom features
- Integrate with existing systems

### Does AzureCopilot work with existing projects?

Currently, AzureCopilot creates new projects from scratch. We're working on features to integrate with existing codebases. Join our waitlist to be notified when this ships.

### What's included in a generated project?

Every AzureCopilot project includes:
- ✅ Backend API with authentication
- ✅ Frontend application (if requested)
- ✅ Azure infrastructure (Bicep)
- ✅ CI/CD pipelines (GitHub Actions)
- ✅ Unit and integration tests
- ✅ Comprehensive documentation
- ✅ Cost estimation
- ✅ Security hardening
```

### Security & Privacy

```markdown
## Security & Privacy

### Is my code secure?

AzureCopilot generates applications with security best practices built-in:
- Secrets stored in Azure Key Vault
- Authentication via Azure Entra ID
- HTTPS enforced everywhere
- Managed identities for service-to-service auth
- Input validation and output encoding

### Where is my code stored?

Your generated code is stored locally on your machine and optionally in your GitHub repository. AzureCopilot does not retain copies of your generated applications.

Prompts and usage metadata may be retained for quality improvement unless you opt out.

### Is AzureCopilot SOC 2 compliant?

AzureCopilot is built on Azure infrastructure which is SOC 2 compliant. Our service is currently undergoing independent SOC 2 audit. Contact us for our current security posture documentation.

[View security documentation →](/docs/security)
```

### Troubleshooting

```markdown
## Troubleshooting

### My deployment failed. What should I do?

1. Check the deployment logs: `azcopilot logs <project-id>`
2. Verify your Azure credentials: `azcopilot auth status`
3. Check Azure resource quotas: `azcopilot quota check`
4. Review common issues: [Deployment troubleshooting guide →](/docs/troubleshooting/deployment)

If the issue persists, contact support with your project ID and error message.

### I'm getting authentication errors. How do I fix them?

Authentication issues usually stem from:
1. **Expired session**: Run `azcopilot auth login` to refresh
2. **Wrong tenant**: Verify you're signed into the correct Azure tenant
3. **Missing permissions**: Ensure you have Contributor access to the resource group

[View auth troubleshooting guide →](/docs/troubleshooting/auth)

### The generated code has an error. What should I do?

1. Run the built-in tests: `npm test`
2. Check for known issues in the error message
3. If it's a genuine bug, report it: `azcopilot feedback --bug`

Our AI is constantly improving, and your feedback helps us generate better code.

### How do I delete a generated project?

To delete a project and all associated Azure resources:

```bash
# Delete Azure resources
azd down

# Remove local files
rm -rf my-project
```

**Warning**: This permanently deletes all data in the Azure resources.
```

### Integration

```markdown
## Integration

### Can I use AzureCopilot with GitHub Copilot?

Yes! AzureCopilot is built on the GitHub Copilot SDK. Your generated code works seamlessly with GitHub Copilot for ongoing development and modifications.

### Does AzureCopilot integrate with Azure DevOps?

Currently, AzureCopilot generates GitHub Actions pipelines. Azure DevOps support is on our roadmap. You can manually convert the generated workflows to Azure Pipelines.

### Can I use my own Azure subscription?

Yes, you deploy to your own Azure subscription. AzureCopilot doesn't provision any Azure resources on your behalf without your explicit approval. You maintain full control over your Azure environment.
```

## FAQ Maintenance

### Review Schedule

| Frequency | Action |
|-----------|--------|
| Weekly | Review support tickets for new FAQs |
| Monthly | Update answers based on product changes |
| Quarterly | Remove outdated FAQs |

### FAQ Performance Metrics

Track:
- Most viewed FAQs
- FAQs that lead to support tickets
- Search queries with no FAQ match
- Time on FAQ pages

### Template for New FAQs

```markdown
### [Question in user's voice?]

[1-2 sentence direct answer]

[Additional context or explanation if needed]

[Step-by-step instructions if applicable]

[Link to detailed documentation →](/docs/...)
```

## Output Location

FAQs should be published to:
- `docs/faq/index.md` - Full FAQ
- Marketing site `/faq` page
- Help center / knowledge base
- In-app help sections
