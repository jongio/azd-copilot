---
name: error-messages
description: Create user-friendly error messages for Azure applications
agent: support
---

# Error Messages

## Purpose

Create clear, actionable error messages that help users resolve issues.

## Error Message Framework

### Structure

```
[What happened] + [Why it happened (if known)] + [What to do next]
```

### Good vs Bad Examples

❌ **Bad**: `Error: ECONNREFUSED`
✅ **Good**: `Unable to connect to the database. Please check that PostgreSQL is running and your connection string is correct.`

❌ **Bad**: `403 Forbidden`
✅ **Good**: `You don't have permission to access this resource. Contact your administrator to request access.`

❌ **Bad**: `Something went wrong`
✅ **Good**: `We couldn't save your changes because the session expired. Please sign in again and try once more.`

## Error Message Templates

### Authentication Errors

```typescript
const authErrors = {
  SESSION_EXPIRED: {
    title: "Session expired",
    message: "Your session has expired. Please sign in again to continue.",
    action: "Sign in",
    actionUrl: "/login"
  },
  
  UNAUTHORIZED: {
    title: "Access denied",
    message: "You don't have permission to perform this action. Contact your administrator if you need access.",
    action: "Request access",
    actionUrl: "/request-access"
  },
  
  MFA_REQUIRED: {
    title: "Additional verification required",
    message: "This action requires multi-factor authentication. Complete the verification to continue.",
    action: "Verify",
    actionUrl: "/mfa"
  },
  
  ACCOUNT_LOCKED: {
    title: "Account locked",
    message: "Your account has been temporarily locked due to too many failed sign-in attempts. Try again in 15 minutes or reset your password.",
    action: "Reset password",
    actionUrl: "/reset-password"
  }
};
```

### Resource Errors

```typescript
const resourceErrors = {
  NOT_FOUND: {
    title: "Not found",
    message: "The requested resource could not be found. It may have been deleted or you may have followed an outdated link.",
    action: "Go back",
    actionUrl: "/"
  },
  
  QUOTA_EXCEEDED: {
    title: "Quota exceeded",
    message: "You've reached the limit for this resource. Upgrade your plan or delete unused resources to continue.",
    action: "View usage",
    actionUrl: "/settings/usage"
  },
  
  CONFLICT: {
    title: "Conflict",
    message: "This resource was modified by someone else. Refresh the page to see the latest version.",
    action: "Refresh",
    actionUrl: null  // JavaScript refresh
  },
  
  RATE_LIMITED: {
    title: "Too many requests",
    message: "You're making requests too quickly. Please wait a moment and try again.",
    retryAfter: 30  // seconds
  }
};
```

### Azure-Specific Errors

```typescript
const azureErrors = {
  KEYVAULT_ACCESS_DENIED: {
    title: "Secret access denied",
    message: "Unable to retrieve configuration from Azure Key Vault. Check that the managed identity has 'Key Vault Secrets User' permissions.",
    docs: "https://docs.microsoft.com/azure/key-vault/general/rbac-guide"
  },
  
  DATABASE_UNAVAILABLE: {
    title: "Database unavailable",
    message: "Unable to connect to Azure PostgreSQL. The database may be starting up or undergoing maintenance.",
    action: "Check Azure status",
    actionUrl: "https://status.azure.com"
  },
  
  STORAGE_QUOTA: {
    title: "Storage limit reached",
    message: "Your Azure Blob Storage account has reached its capacity. Delete unused files or request a quota increase.",
    action: "Manage storage",
    actionUrl: "https://portal.azure.com"
  },
  
  REGION_UNAVAILABLE: {
    title: "Service unavailable in region",
    message: "This Azure service is not available in the selected region. Choose a different region or contact support.",
    action: "Select region",
    actionUrl: "/settings/region"
  },
  
  DEPLOYMENT_FAILED: {
    title: "Deployment failed",
    message: "Unable to deploy to Azure Container Apps. Check the deployment logs for details.",
    action: "View logs",
    actionUrl: "/deployments/latest/logs"
  }
};
```

### Validation Errors

```typescript
const validationErrors = {
  REQUIRED_FIELD: (field: string) => ({
    field,
    message: `${field} is required`
  }),
  
  INVALID_FORMAT: (field: string, expected: string) => ({
    field,
    message: `${field} must be ${expected}`
  }),
  
  TOO_SHORT: (field: string, min: number) => ({
    field,
    message: `${field} must be at least ${min} characters`
  }),
  
  TOO_LONG: (field: string, max: number) => ({
    field,
    message: `${field} must be no more than ${max} characters`
  }),
  
  INVALID_EMAIL: {
    field: "email",
    message: "Please enter a valid email address"
  },
  
  WEAK_PASSWORD: {
    field: "password",
    message: "Password must include uppercase, lowercase, number, and special character"
  }
};
```

## Error Response Format

### API Error Response

```typescript
interface ApiError {
  error: {
    code: string;           // Machine-readable code
    message: string;        // User-friendly message
    target?: string;        // Field or resource that caused error
    details?: ApiError[];   // Nested errors (e.g., validation)
    innererror?: {
      code: string;         // More specific error code
      message: string;      // Technical details (dev only)
    };
  };
  requestId: string;        // For support correlation
}

// Example
{
  "error": {
    "code": "ValidationError",
    "message": "The request contains invalid data",
    "details": [
      {
        "code": "RequiredField",
        "message": "Email is required",
        "target": "email"
      },
      {
        "code": "InvalidFormat",
        "message": "Phone number must be in format +1-555-123-4567",
        "target": "phone"
      }
    ]
  },
  "requestId": "abc-123-def"
}
```

### UI Error Display

```tsx
// React error component
function ErrorMessage({ error }: { error: ApiError }) {
  return (
    <div role="alert" className="error-message">
      <h3>{error.title || "Something went wrong"}</h3>
      <p>{error.message}</p>
      
      {error.action && (
        <a href={error.actionUrl} className="error-action">
          {error.action}
        </a>
      )}
      
      <details>
        <summary>Technical details</summary>
        <p>Error code: {error.code}</p>
        <p>Request ID: {error.requestId}</p>
      </details>
    </div>
  );
}
```

## Error Message Guidelines

### Do:
- Use plain language (avoid jargon)
- Tell users what to do next
- Provide a way to recover
- Include error codes for support
- Be specific about what failed

### Don't:
- Blame the user ("You entered invalid data")
- Expose technical details to end users
- Use negative language ("Failed", "Error", "Invalid")
- Leave users stuck with no next step
- Show stack traces in production

### Tone Examples

❌ "Invalid input"
✅ "Please enter your email address"

❌ "Operation failed"
✅ "We couldn't complete your request. Please try again."

❌ "Error 500"
✅ "Something unexpected happened on our end. Our team has been notified."

## Logging vs Display

```typescript
// What to log (for developers)
logger.error('Database connection failed', {
  error: err.message,
  stack: err.stack,
  connectionString: '[REDACTED]',
  host: 'myserver.postgres.database.azure.com',
  database: 'mydb',
  duration: 5000
});

// What to show (for users)
throw new UserError(
  'DATABASE_UNAVAILABLE',
  'Unable to connect to the database. Please try again in a few moments.'
);
```

## Error Code Catalog

Maintain a catalog of all error codes:

```markdown
## Error Code Catalog

| Code | HTTP | Message Template | Resolution |
|------|------|------------------|------------|
| AUTH_001 | 401 | Session expired | Sign in again |
| AUTH_002 | 403 | Access denied | Contact admin |
| VAL_001 | 400 | Required field missing | Fill in field |
| VAL_002 | 400 | Invalid format | Correct format |
| RES_001 | 404 | Resource not found | Check URL |
| RES_002 | 409 | Resource conflict | Refresh page |
| SYS_001 | 500 | Internal error | Retry later |
| SYS_002 | 503 | Service unavailable | Check status |
```
