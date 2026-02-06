---
name: code-review
description: Code review checklist and best practices for high-quality reviews
triggers:
  - code review
  - review code
  - pr review
---

# Code Review Skill

A systematic approach to code reviews that catches real issues without nitpicking.

## Review Philosophy

**Focus on what matters:**
- 🐛 Bugs and logic errors
- 🔒 Security vulnerabilities
- 💥 Breaking changes
- 🏗️ Architectural issues

**Don't waste time on:**
- Formatting (let tools handle it)
- Personal style preferences
- Minor naming debates

## 8-Category Review Checklist

### 1. Security (Critical)

- [ ] No hardcoded secrets or credentials
- [ ] Input validation on all external data
- [ ] SQL/NoSQL injection prevention (parameterized queries)
- [ ] XSS prevention (proper escaping/encoding)
- [ ] Authentication checks on protected routes
- [ ] Authorization checks (user can access this resource?)
- [ ] Sensitive data not logged
- [ ] HTTPS enforced for sensitive operations

```typescript
// ❌ Bad - SQL injection risk
const query = `SELECT * FROM users WHERE id = '${userId}'`;

// ✅ Good - Parameterized query
const query = await db.query('SELECT * FROM users WHERE id = $1', [userId]);
```

### 2. Logic Errors (Critical)

- [ ] Edge cases handled (null, empty, zero, negative)
- [ ] Boundary conditions correct (off-by-one)
- [ ] Error handling complete (try/catch, .catch())
- [ ] Async/await used correctly (no floating promises)
- [ ] Race conditions prevented
- [ ] State mutations are intentional

```typescript
// ❌ Bad - floating promise
async function save() {
    updateDatabase(data); // Promise not awaited!
    return 'saved';
}

// ✅ Good - properly awaited
async function save() {
    await updateDatabase(data);
    return 'saved';
}
```

### 3. Type Safety

- [ ] No `any` types (unless justified)
- [ ] Null/undefined handled (optional chaining, nullish coalescing)
- [ ] Type guards used for narrowing
- [ ] Generic types are appropriate
- [ ] Return types explicit on public functions

```typescript
// ❌ Bad - any type
function process(data: any) { ... }

// ✅ Good - proper typing
function process(data: ProcessInput): ProcessResult { ... }
```

### 4. Error Handling

- [ ] Errors are caught and handled
- [ ] Error messages are helpful (not generic)
- [ ] Errors don't leak internal details to users
- [ ] Failed operations have proper cleanup
- [ ] Retry logic has backoff and limits

```typescript
// ❌ Bad - swallowed error
try {
    await riskyOperation();
} catch {
    // Silent failure!
}

// ✅ Good - proper handling
try {
    await riskyOperation();
} catch (error) {
    logger.error('Operation failed', { error, context });
    throw new UserFacingError('Unable to complete operation');
}
```

### 5. Code Quality

- [ ] Functions do one thing (single responsibility)
- [ ] No copy-pasted code (DRY)
- [ ] Names are descriptive and consistent
- [ ] Comments explain "why", not "what"
- [ ] Complexity is manageable (no 10-level nesting)

### 6. Test Coverage

- [ ] New code has tests
- [ ] Edge cases are tested
- [ ] Error paths are tested
- [ ] Tests are meaningful (not just for coverage)
- [ ] No flaky tests introduced

### 7. Performance

- [ ] No N+1 queries
- [ ] Large lists are paginated
- [ ] Expensive operations are cached/memoized
- [ ] No memory leaks (cleanup subscriptions, listeners)
- [ ] Async operations don't block

```typescript
// ❌ Bad - N+1 query
for (const user of users) {
    const orders = await getOrdersForUser(user.id); // Query per user!
}

// ✅ Good - batch query
const orders = await getOrdersForUsers(users.map(u => u.id));
```

### 8. Architecture

- [ ] Changes fit the existing patterns
- [ ] Dependencies flow in correct direction
- [ ] No circular dependencies
- [ ] Proper separation of concerns
- [ ] Breaking changes are documented

## Review Response Templates

### Approval
```
LGTM! ✅

Minor suggestions (non-blocking):
- [optional improvements]
```

### Request Changes
```
Good progress! A few things to address:

**Must fix:**
- [ ] [Critical issue]

**Should fix:**
- [ ] [Important but not blocking]

Let me know if you have questions!
```

### Conditional Approval
```
Looks good with one condition:

- [ ] [One thing that must be done]

Once that's addressed, feel free to merge.
```

## Language-Specific Idioms

### TypeScript
- Use `const` by default, `let` only when reassigning
- Prefer `interface` over `type` for object shapes
- Use strict mode (`strict: true` in tsconfig)
- Avoid `enum`, prefer const objects or union types

### Python
- Use type hints on all public functions
- Follow PEP 8 naming conventions
- Use dataclasses or Pydantic for data models
- Prefer `pathlib` over string path manipulation

### C#/.NET
- Use `var` when type is obvious from RHS
- Prefer records for immutable data
- Use `async`/`await` all the way up
- Follow Microsoft naming conventions
