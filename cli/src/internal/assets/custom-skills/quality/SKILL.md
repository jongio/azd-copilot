---
name: quality
description: Code quality skills including refactoring patterns and code review guidelines
---

# Quality Skills

Skills for maintaining and improving code quality.

## Available Skills

| Skill | Description | When to Use |
|-------|-------------|-------------|
| [refactor](refactor.md) | Systematic refactoring patterns | When improving existing code |
| [code-review](code-review.md) | Review checklist and guidelines | When reviewing PRs or code |
| [e2e-playwright](e2e-playwright.md) | Playwright E2E testing patterns | **When project has a frontend component** |

## Quick Reference

### Before Refactoring
1. Ensure tests exist
2. Run tests (must pass)
3. Create checkpoint

### During Code Review
Focus on:
- Security vulnerabilities
- Logic errors
- Type safety
- Error handling

Don't nitpick:
- Formatting (let tools handle it)
- Personal style preferences

### E2E Testing (Frontends)
If the project has a frontend (SPA, SSR, SWA):
1. Install Playwright: `npm init playwright@latest`
2. Create tests in `tests/e2e/`
3. Test core user flows, navigation, forms
4. Use the **Playwright MCP server** for browser interaction during test development
