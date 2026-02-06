---
name: acceptance-criteria
description: Define clear, testable acceptance criteria for features
agent: product
---

# Acceptance Criteria

## Purpose

Define clear, testable acceptance criteria that verify features work correctly on Azure.

## Format: Given-When-Then

```gherkin
Feature: [Feature Name]

  Scenario: [Scenario Name]
    Given [initial context/state]
    When [action performed]
    Then [expected outcome]
    And [additional outcome]
```

## Examples

### API Endpoint

```gherkin
Feature: Create User API

  Scenario: Successfully create a new user
    Given the API is running on Azure Container Apps
    And the user has a valid authentication token
    When a POST request is made to /api/users with valid user data
    Then the response status code is 201
    And the response body contains the created user ID
    And the user is persisted in Azure PostgreSQL
    And an audit log entry is created in Log Analytics

  Scenario: Reject duplicate email
    Given a user with email "test@example.com" already exists
    When a POST request is made with the same email
    Then the response status code is 409
    And the response body contains error code "USER_EXISTS"
```

### Azure-Specific Criteria

```gherkin
Feature: Secure Configuration

  Scenario: Secrets loaded from Key Vault
    Given the application is deployed to Azure Container Apps
    When the application starts
    Then all secrets are loaded from Azure Key Vault
    And no secrets appear in environment variables
    And no secrets are logged to Application Insights

  Scenario: Managed Identity authentication
    Given the Container App has a managed identity
    When the app connects to PostgreSQL
    Then authentication uses the managed identity
    And no connection string password is used
```

## Acceptance Criteria Checklist

### Functional
- [ ] Describes observable behavior
- [ ] Includes success criteria
- [ ] Includes error scenarios
- [ ] Specifies edge cases

### Azure-Specific
- [ ] Verifies correct Azure service integration
- [ ] Includes security requirements (Key Vault, Managed Identity)
- [ ] Includes observability (logs, metrics)
- [ ] Includes cost-relevant behaviors (caching, batching)

### Testable
- [ ] Can be automated
- [ ] Has clear pass/fail criteria
- [ ] Doesn't depend on manual inspection
- [ ] Includes performance thresholds where relevant

## Anti-Patterns

❌ **Vague criteria**
```
Then the system should be fast
```

✅ **Specific criteria**
```
Then the API response time is less than 200ms at p95
```

❌ **Implementation details**
```
Then the code uses async/await
```

✅ **Behavior focus**
```
Then the API can handle 100 concurrent requests without errors
```

## Output

Acceptance criteria should be stored in:
- `docs/requirements/acceptance-criteria/` - By feature
- Or inline in user stories

Format: Markdown with Gherkin syntax for test automation compatibility.
