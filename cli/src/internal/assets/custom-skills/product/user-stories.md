---
name: user-stories
description: Create well-formed user stories with Azure context
agent: product
---

# User Stories

## Purpose

Create user stories that capture user needs and map to Azure implementation patterns.

## User Story Format

```markdown
## US-001: [Short Title]

**As a** [persona/role]
**I want to** [action/capability]
**So that** [benefit/value]

### Details
[Additional context, constraints, or notes]

### Azure Services
- [Primary service]
- [Supporting services]

### Acceptance Criteria
- [ ] Given/When/Then 1
- [ ] Given/When/Then 2

### Dependencies
- [Blocked by / Blocks]

### Estimate
[T-shirt size: XS, S, M, L, XL]
```

## Persona Templates

### Developer Persona
```markdown
**Alex - Backend Developer**
- Uses Azure daily for deployment
- Comfortable with CLI and APIs
- Values: Automation, clear documentation, fast feedback
- Pain points: Manual processes, unclear errors, slow builds
```

### End User Persona
```markdown
**Jordan - Business User**
- Accesses the app via web browser
- Limited technical knowledge
- Values: Simplicity, reliability, speed
- Pain points: Complex UIs, downtime, data loss
```

## Story Examples

### Authentication Story
```markdown
## US-AUTH-001: Secure Login

**As a** registered user
**I want to** log in with my Microsoft account
**So that** I can access my personalized dashboard securely

### Azure Services
- Azure Entra ID (authentication)
- Azure Key Vault (token signing keys)
- Application Insights (login metrics)

### Acceptance Criteria
- [ ] Given I'm on the login page, when I click "Sign in with Microsoft", then I'm redirected to Entra ID
- [ ] Given I complete Entra ID login, when authentication succeeds, then I'm redirected to my dashboard
- [ ] Given I'm authenticated, when I refresh the page, then I remain logged in
- [ ] Given my session expires, when I try to access protected content, then I'm prompted to re-authenticate
```

### Data Story
```markdown
## US-DATA-001: View My Orders

**As a** customer
**I want to** see my order history
**So that** I can track past purchases

### Azure Services
- Azure PostgreSQL (order data)
- Azure Redis Cache (session/recent orders)
- Azure CDN (static assets)

### Acceptance Criteria
- [ ] Given I'm logged in, when I navigate to "My Orders", then I see a paginated list of orders
- [ ] Given I have 100+ orders, when I load the page, then the first page loads in under 500ms
- [ ] Given an order exists, when I click on it, then I see order details including items and status
```

## Story Splitting Guidelines

Large stories should be split by:

1. **Workflow steps**: Login → View → Edit → Save
2. **User types**: Admin vs Regular user
3. **Data operations**: Create, Read, Update, Delete
4. **Happy path vs error handling**: Success first, then errors
5. **Azure services**: Core service first, then enhancements

## INVEST Criteria

Good stories are:
- **I**ndependent - Can be developed in any order
- **N**egotiable - Details can be discussed
- **V**aluable - Delivers user value
- **E**stimable - Team can estimate effort
- **S**mall - Fits in one sprint
- **T**estable - Has clear acceptance criteria

## Output Location

Store user stories in:
- `docs/requirements/user-stories/` - Organized by epic
- Or project management tool (GitHub Issues, Azure DevOps)
