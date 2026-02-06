---
name: refactor
description: Code refactoring patterns and checklist for improving code quality
triggers:
  - refactor
  - cleanup
  - improve code
  - reduce duplication
---

# Refactor Skill

A systematic approach to code refactoring with safety checks and patterns.

## Pre-Refactor Checklist

Before any refactoring:

1. **Verify tests exist** - Never refactor without test coverage
2. **Run tests** - Confirm all tests pass before changes
3. **Identify scope** - Define exactly what will be changed
4. **Create checkpoint** - Ensure changes can be reverted

## Refactoring Patterns

### 1. Extract Function/Method

**When:** Code block is repeated or does one specific thing

```typescript
// Before
function processUser(user: User) {
    // validation logic
    if (!user.email || !user.email.includes('@')) {
        throw new Error('Invalid email');
    }
    if (!user.name || user.name.length < 2) {
        throw new Error('Invalid name');
    }
    // ... rest of processing
}

// After
function validateUser(user: User): void {
    if (!user.email || !user.email.includes('@')) {
        throw new Error('Invalid email');
    }
    if (!user.name || user.name.length < 2) {
        throw new Error('Invalid name');
    }
}

function processUser(user: User) {
    validateUser(user);
    // ... rest of processing
}
```

### 2. Replace Magic Numbers

**When:** Hardcoded values appear in code

```typescript
// Before
if (retryCount > 3) { ... }
await sleep(5000);

// After
const MAX_RETRIES = 3;
const RETRY_DELAY_MS = 5000;

if (retryCount > MAX_RETRIES) { ... }
await sleep(RETRY_DELAY_MS);
```

### 3. Simplify Conditionals

**When:** Complex nested if/else or switch statements

```typescript
// Before
function getDiscount(customer: Customer): number {
    if (customer.type === 'premium') {
        if (customer.yearsActive > 5) {
            return 0.25;
        } else {
            return 0.15;
        }
    } else if (customer.type === 'standard') {
        if (customer.yearsActive > 5) {
            return 0.10;
        } else {
            return 0.05;
        }
    }
    return 0;
}

// After
const DISCOUNT_RATES: Record<string, Record<string, number>> = {
    premium: { loyal: 0.25, regular: 0.15 },
    standard: { loyal: 0.10, regular: 0.05 },
};

function getDiscount(customer: Customer): number {
    const tier = customer.yearsActive > 5 ? 'loyal' : 'regular';
    return DISCOUNT_RATES[customer.type]?.[tier] ?? 0;
}
```

### 4. Remove Dead Code

**When:** Code is never executed or referenced

Signs of dead code:
- Unreachable code after return/throw
- Unused variables or imports
- Commented-out code blocks
- Functions never called

```typescript
// Remove unused imports
import { used } from './module'; // Keep
// import { unused } from './module'; // Remove

// Remove unreachable code
function example() {
    return result;
    console.log('This never runs'); // Remove
}
```

### 5. Consolidate Duplicate Logic

**When:** Similar code appears in multiple places

```typescript
// Before - duplicated in multiple files
function fetchUsers() {
    try {
        const res = await fetch('/api/users');
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        return res.json();
    } catch (err) {
        logger.error('Failed to fetch users', err);
        throw err;
    }
}

// After - shared utility
async function apiFetch<T>(path: string): Promise<T> {
    try {
        const res = await fetch(path);
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        return res.json();
    } catch (err) {
        logger.error(`Failed to fetch ${path}`, err);
        throw err;
    }
}

const users = await apiFetch<User[]>('/api/users');
```

## Code Smells to Address

| Smell | Description | Solution |
|-------|-------------|----------|
| Long Method | >50 lines | Extract methods |
| Long File | >500 lines | Split into modules |
| Deep Nesting | >3 levels | Early returns, extract |
| Too Many Parameters | >4 params | Use options object |
| Feature Envy | Accesses other class's data | Move method |
| Primitive Obsession | Strings for everything | Create types |
| Magic Numbers | Hardcoded values | Named constants |

## Post-Refactor Checklist

After refactoring:

1. **Run tests** - All tests must still pass
2. **Run linter** - No new lint errors
3. **Check types** - No TypeScript errors
4. **Review diff** - Verify changes are minimal and correct
5. **Document** - Update comments if behavior changed
