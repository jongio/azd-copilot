---
name: requirements
description: Gather and document requirements with Azure feasibility validation
agent: product
---

# Requirements Gathering

## Purpose

Systematically gather, document, and validate requirements for Azure applications.

## Process

### 1. Stakeholder Identification

```markdown
## Stakeholders

| Role | Name | Responsibilities | Contact |
|------|------|------------------|---------|
| Product Owner | | Final decisions | |
| Technical Lead | | Architecture approval | |
| End Users | | Acceptance testing | |
| Operations | | Support & maintenance | |
```

### 2. Functional Requirements Template

```markdown
## Functional Requirements

### FR-001: [Short Name]
- **Description**: [What the system must do]
- **Priority**: [Must Have | Should Have | Nice to Have]
- **User Story**: As a [persona], I want [action] so that [benefit]
- **Azure Services**: [Relevant Azure services]
- **Acceptance Criteria**:
  - [ ] Criterion 1
  - [ ] Criterion 2
```

### 3. Non-Functional Requirements

| Category | Requirement | Azure Solution |
|----------|-------------|----------------|
| **Performance** | Response time < 200ms | Azure Front Door, Redis Cache |
| **Scalability** | Handle 10K concurrent users | Container Apps autoscaling |
| **Availability** | 99.9% uptime | Multi-region deployment |
| **Security** | RBAC, encryption at rest | Entra ID, Key Vault |
| **Compliance** | SOC 2, GDPR | Azure Policy, Purview |

### 4. Azure Feasibility Check

Before finalizing requirements, validate:

- [ ] All required Azure services available in target region
- [ ] Service quotas sufficient for expected load
- [ ] Pricing model understood and budgeted
- [ ] No preview features in production path (unless approved)
- [ ] Compliance requirements achievable

## Output Artifacts

1. `docs/requirements/functional.md` - Functional requirements
2. `docs/requirements/non-functional.md` - Non-functional requirements
3. `docs/requirements/azure-services.md` - Azure service mapping
4. `docs/requirements/constraints.md` - Known constraints and limitations

## Integration with Other Roles

| Role | Handoff |
|------|---------|
| architect | Validated requirements → architecture design |
| dev | User stories → implementation |
| quality | Acceptance criteria → test cases |
| finance | Requirements → cost estimation |
