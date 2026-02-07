---
name: azure-marketplace
description: Create Azure Marketplace listings for SaaS applications
agent: marketing
---

# Azure Marketplace Listing

## Purpose

Create effective Azure Marketplace listings for SaaS applications.

## Listing Components

### 1. Offer Details

```yaml
# marketplace-offer.yaml

offer:
  name: "AzureCopilot"
  type: "SaaS"  # SaaS, VM, Container, Managed App
  categories:
    - "Developer Tools"
    - "AI + Machine Learning"
  industries:
    - "Professional Services"
    - "IT & Communications"
  
  summary: "AI-powered Azure application generator"  # 100 chars max
  
  description: |
    Generate production-ready Azure applications from natural language prompts.
    
    ## Key Features
    - 16 specialized AI agents
    - Complete infrastructure with Azure Verified Modules
    - Security hardened by default
    - Full test coverage included
    - One-command deployment with azd
    
    ## What You Get
    - Backend APIs (Node.js, .NET, Python)
    - Frontend applications (React, Vue)
    - Azure infrastructure (Bicep)
    - CI/CD pipelines (GitHub Actions)
    - Comprehensive documentation
    
    ## Azure Services Used
    - Azure Container Apps
    - Azure PostgreSQL
    - Azure Key Vault
    - Azure Application Insights
    - Azure Static Web Apps
```

### 2. Plan Configuration

```yaml
plans:
  - name: "Free"
    displayName: "Free Tier"
    description: "Get started with AzureCopilot"
    pricing:
      model: "flat-rate"
      price: 0
      billingTerm: "monthly"
    limits:
      projects: 3
      support: "community"
    
  - name: "Professional"
    displayName: "Professional"
    description: "For individual developers and small teams"
    pricing:
      model: "flat-rate"
      price: 29
      billingTerm: "monthly"
    limits:
      projects: "unlimited"
      support: "email"
    features:
      - "Unlimited projects"
      - "Email support"
      - "Priority generation queue"
    
  - name: "Team"
    displayName: "Team"
    description: "For development teams"
    pricing:
      model: "per-user"
      price: 15
      billingTerm: "monthly"
      minUsers: 5
    limits:
      projects: "unlimited"
      support: "priority"
    features:
      - "Everything in Professional"
      - "Team management"
      - "Shared templates"
      - "Priority support"
```

### 3. Technical Configuration

```yaml
technical:
  landingPage: "https://azurecopilot.dev/marketplace/landing"
  connectionWebhook: "https://api.azurecopilot.dev/marketplace/webhook"
  
  # SaaS fulfillment API integration
  fulfillment:
    activationUrl: "https://api.azurecopilot.dev/marketplace/activate"
    subscriptionApi: "v2"
  
  # Single sign-on
  sso:
    aadAppId: "00000000-0000-0000-0000-000000000000"
    tenantId: "common"  # Multi-tenant
```

### 4. Marketplace Assets

```yaml
assets:
  logo:
    small: "assets/logo-48x48.png"   # 48x48 PNG
    medium: "assets/logo-90x90.png"  # 90x90 PNG
    large: "assets/logo-216x216.png" # 216x216 PNG
    wide: "assets/logo-255x115.png"  # 255x115 PNG
    hero: "assets/hero-815x290.png"  # 815x290 PNG
  
  screenshots:  # 1280x720 recommended
    - path: "assets/screenshot-dashboard.png"
      caption: "Project dashboard showing AI agents at work"
    - path: "assets/screenshot-code.png"
      caption: "Generated code with full type safety"
    - path: "assets/screenshot-deploy.png"
      caption: "One-command deployment to Azure"
  
  videos:
    - url: "https://www.youtube.com/watch?v=..."
      thumbnail: "assets/video-thumbnail.png"
      title: "Getting Started with AzureCopilot"
```

### 5. Legal & Support

```yaml
legal:
  privacyPolicy: "https://azurecopilot.dev/privacy"
  termsOfUse: "https://azurecopilot.dev/terms"
  
support:
  helpUrl: "https://docs.azurecopilot.dev"
  supportUrl: "https://azurecopilot.dev/support"
  engineeringContact: "engineering@azurecopilot.dev"
  supportContact: "support@azurecopilot.dev"
  
# Required certifications
certifications:
  - type: "Microsoft 365 Certification"
    status: "not-required"  # Only for M365 apps
  - type: "Azure Certified"
    status: "in-progress"
```

## Listing Best Practices

### Title & Summary
- **Title**: Clear, searchable (include "Azure" if relevant)
- **Summary**: Lead with benefit, include key differentiator
- Keep under 100 characters

### Description
- Start with value proposition
- Use headers and bullets for scannability
- Include specific features and benefits
- Mention Azure services used
- Add call-to-action at the end

### Screenshots
- Show the product in action
- Include captions explaining value
- Use consistent branding
- Show Azure integration points

### Pricing
- Be transparent about all costs
- Explain what's included in each tier
- Note that Azure infrastructure is separate
- Offer free trial when possible

## SaaS Fulfillment Flow

```
1. Customer finds offer in Marketplace
2. Customer selects plan and subscribes
3. Azure sends webhook to your landing page
4. Customer completes account setup
5. You call Marketplace API to activate
6. Subscription is now active
7. You report usage (if metered)
8. Azure handles billing
```

### Landing Page Requirements

```html
<!-- Marketplace landing page must: -->
<!-- 1. Accept marketplace token -->
<!-- 2. Resolve token to get subscription details -->
<!-- 3. Collect any additional info needed -->
<!-- 4. Activate subscription via API -->

<script>
  // Token passed in URL
  const token = new URLSearchParams(location.search).get('token');
  
  // Resolve token to get subscription
  const subscription = await resolveMarketplaceToken(token);
  
  // Show user the plan details
  showPlanDetails(subscription);
  
  // On user confirmation
  await activateSubscription(subscription.id);
</script>
```

## Certification Requirements

For Azure Certified badge:
- [ ] Deployed on Azure infrastructure
- [ ] Uses Azure Active Directory for auth
- [ ] Follows Azure security best practices
- [ ] Passes security review
- [ ] Has SLA defined
- [ ] Support channels documented

## Output Files

```
marketplace/
├── offer.json           # Offer configuration
├── plans/
│   ├── free.json
│   ├── professional.json
│   └── team.json
├── assets/
│   ├── logos/
│   ├── screenshots/
│   └── videos/
├── landing/
│   └── index.html       # Landing page
└── docs/
    ├── getting-started.md
    └── faq.md
```
