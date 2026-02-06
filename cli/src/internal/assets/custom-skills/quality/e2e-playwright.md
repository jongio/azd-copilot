# Playwright E2E Testing

End-to-end testing patterns using Playwright for Azure-hosted frontends.

## When to Use

Write Playwright E2E tests whenever the project includes a frontend:
- React / Vue / Svelte / Angular SPA
- Static Web App (SWA)
- SSR Web App (Next.js, Nuxt, SvelteKit)
- Any web UI hosted on Azure

## Playwright MCP Server

The **Playwright MCP server** (`@playwright/mcp`) is auto-configured with azd-copilot. Use it during test development to:
- Navigate and interact with the app in a real browser
- Inspect page state, elements, and accessibility
- Debug test failures visually

## Setup

### New Project

```bash
npm init playwright@latest
```

This creates:
- `playwright.config.ts` — Configuration
- `tests/` — Test directory
- `.github/workflows/playwright.yml` — CI workflow (optional)

### Existing Project

```bash
npm install -D @playwright/test
npx playwright install
```

## Configuration

```typescript
// playwright.config.ts
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  reporter: process.env.CI ? 'github' : 'html',
  use: {
    baseURL: process.env.PLAYWRIGHT_BASE_URL || 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'mobile',
      use: { ...devices['iPhone 14'] },
    },
  ],
  webServer: {
    command: 'npm run dev',
    port: 3000,
    reuseExistingServer: !process.env.CI,
  },
});
```

## Test Patterns

### Page Navigation

```typescript
import { test, expect } from '@playwright/test';

test('should load home page', async ({ page }) => {
  await page.goto('/');
  await expect(page).toHaveTitle(/My App/);
  await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
});

test('should navigate between pages', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('link', { name: /about/i }).click();
  await expect(page).toHaveURL(/.*about/);
});
```

### Form Interaction

```typescript
test('should submit contact form', async ({ page }) => {
  await page.goto('/contact');

  await page.getByLabel('Name').fill('Test User');
  await page.getByLabel('Email').fill('test@example.com');
  await page.getByLabel('Message').fill('Hello from Playwright');
  await page.getByRole('button', { name: /submit/i }).click();

  await expect(page.getByText(/thank you/i)).toBeVisible();
});

test('should show validation errors', async ({ page }) => {
  await page.goto('/contact');
  await page.getByRole('button', { name: /submit/i }).click();

  await expect(page.getByText(/required/i)).toBeVisible();
});
```

### Authentication Flow

```typescript
test('should login and access dashboard', async ({ page }) => {
  await page.goto('/login');

  await page.getByLabel('Email').fill('user@example.com');
  await page.getByLabel('Password').fill('password123');
  await page.getByRole('button', { name: /sign in/i }).click();

  await expect(page).toHaveURL(/.*dashboard/);
  await expect(page.getByText(/welcome/i)).toBeVisible();
});
```

### API-Driven UI

```typescript
test('should display data from API', async ({ page }) => {
  await page.goto('/items');

  // Wait for data to load
  await expect(page.getByRole('list')).toBeVisible();
  const items = page.getByRole('listitem');
  await expect(items).toHaveCount(await items.count());
  expect(await items.count()).toBeGreaterThan(0);
});
```

### Accessibility

```typescript
test('should be keyboard navigable', async ({ page }) => {
  await page.goto('/');

  // Tab through interactive elements
  await page.keyboard.press('Tab');
  const focused = page.locator(':focus');
  await expect(focused).toBeVisible();
});
```

## What to Test

| Priority | Category | Examples |
|----------|----------|---------|
| 🔴 Must | **Core Flows** | Login, signup, CRUD operations |
| 🔴 Must | **Navigation** | All pages load, links work, routing |
| 🟡 Should | **Forms** | Validation, submit, error states |
| 🟡 Should | **Responsive** | Mobile viewport, tablet viewport |
| 🟢 Nice | **Accessibility** | Keyboard nav, ARIA labels, focus management |
| 🟢 Nice | **Visual** | Screenshot comparisons (via `toHaveScreenshot()`) |

## CI Integration

### GitHub Actions

```yaml
# .github/workflows/e2e.yml
name: E2E Tests

on:
  push:
    branches: [main]
  pull_request:

jobs:
  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
      - run: npm ci
      - run: npx playwright install --with-deps
      - run: npx playwright test
      - uses: actions/upload-artifact@v4
        if: ${{ !cancelled() }}
        with:
          name: playwright-report
          path: playwright-report/
          retention-days: 30
```

## Azure-Specific Patterns

### Testing Deployed App

After `azd up`, test the live deployment:

```bash
# Get the deployed URL
ENDPOINT=$(azd env get-value SERVICE_WEB_ENDPOINT_URL)

# Run E2E tests against deployment
PLAYWRIGHT_BASE_URL=$ENDPOINT npx playwright test
```

### Testing with SWA CLI

For Static Web Apps with API:

```typescript
// playwright.config.ts
webServer: {
  command: 'swa start',
  port: 4280,
  reuseExistingServer: !process.env.CI,
},
```

## File Structure

```
tests/
├── e2e/
│   ├── home.spec.ts          # Home page tests
│   ├── auth.spec.ts          # Authentication tests
│   ├── dashboard.spec.ts     # Dashboard tests
│   └── fixtures/
│       └── test-data.ts      # Shared test data
├── unit/                     # Unit tests (Vitest/Jest)
└── integration/              # API integration tests
```
