// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package scenario

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// VerificationResult holds the outcome of running all verification steps.
type VerificationResult struct {
	Steps   map[string]VerifyResult
	Passed  bool
	Summary string
}

// RunVerification executes the Playwright verification steps for a scenario.
// The endpoint URL is discovered from the azd deployment output or provided directly.
func RunVerification(ctx context.Context, s *Scenario, workDir string, endpoint string) (*VerificationResult, error) {
	if len(s.Verification) == 0 {
		return &VerificationResult{Passed: true, Summary: "no verification steps defined"}, nil
	}

	if endpoint == "" {
		// Try to discover endpoint from azd env
		endpoint = discoverEndpoint(workDir)
	}

	if endpoint == "" {
		return &VerificationResult{
			Passed:  false,
			Summary: "no endpoint URL found — cannot run verification",
			Steps:   map[string]VerifyResult{"endpoint_discovery": {Passed: false, Error: "no endpoint URL"}},
		}, nil
	}

	fmt.Printf("🧪 Running %d verification steps against %s\n", len(s.Verification), endpoint)

	// Generate Playwright test file
	testDir, err := os.MkdirTemp("", "scenario-verify-*")
	if err != nil {
		return nil, fmt.Errorf("create verify dir: %w", err)
	}
	defer os.RemoveAll(testDir)

	testCode := generatePlaywrightTest(s, endpoint)
	testFile := filepath.Join(testDir, "verify.spec.js")
	if err := os.WriteFile(testFile, []byte(testCode), 0644); err != nil {
		return nil, fmt.Errorf("write test file: %w", err)
	}

	// Write minimal playwright config
	configCode := `
import { defineConfig } from '@playwright/test';
export default defineConfig({
  timeout: 30000,
  use: { headless: true },
  reporter: [['json', { outputFile: 'results.json' }]],
});
`
	configFile := filepath.Join(testDir, "playwright.config.ts")
	if err := os.WriteFile(configFile, []byte(configCode), 0644); err != nil {
		return nil, fmt.Errorf("write config: %w", err)
	}

	// Run Playwright
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "npx", "playwright", "test", "--config", configFile, testFile)
	cmd.Dir = testDir
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	fmt.Println(outputStr)

	// Parse results from output
	result := parseVerificationResults(s, outputStr, err)
	return result, nil
}

// generatePlaywrightTest creates a Playwright test file from scenario verification steps.
func generatePlaywrightTest(s *Scenario, endpoint string) string {
	var b strings.Builder

	b.WriteString("import { test, expect } from '@playwright/test';\n\n")
	b.WriteString(fmt.Sprintf("const ENDPOINT = '%s';\n\n", endpoint))

	b.WriteString(fmt.Sprintf("test.describe('%s verification', () => {\n", s.Name))

	for i, step := range s.Verification {
		testName := step.Name
		if testName == "" {
			testName = fmt.Sprintf("step-%d", i+1)
		}

		b.WriteString(fmt.Sprintf("  test('%s', async ({ page }) => {\n", testName))

		switch step.Action {
		case "navigate":
			url := resolveURL(step.URL, endpoint)
			b.WriteString(fmt.Sprintf("    const response = await page.goto('%s');\n", url))
			if step.StatusCode > 0 {
				b.WriteString(fmt.Sprintf("    expect(response.status()).toBe(%d);\n", step.StatusCode))
			} else {
				b.WriteString("    expect(response.status()).toBeLessThan(400);\n")
			}
			if step.Value != "" {
				b.WriteString(fmt.Sprintf("    await expect(page.locator('body')).toContainText('%s');\n", escapeJS(step.Value)))
			}

		case "click":
			b.WriteString(fmt.Sprintf("    await page.locator('%s').click();\n", escapeJS(step.Selector)))

		case "type":
			b.WriteString(fmt.Sprintf("    await page.locator('%s').fill('%s');\n", escapeJS(step.Selector), escapeJS(step.Value)))

		case "wait":
			if step.Selector != "" {
				b.WriteString(fmt.Sprintf("    await page.waitForSelector('%s', { timeout: 10000 });\n", escapeJS(step.Selector)))
			} else {
				b.WriteString("    await page.waitForTimeout(2000);\n")
			}

		case "check":
			if step.Selector != "" {
				b.WriteString(fmt.Sprintf("    await expect(page.locator('%s')).toBeVisible();\n", escapeJS(step.Selector)))
				if step.Value != "" {
					b.WriteString(fmt.Sprintf("    await expect(page.locator('%s')).toContainText('%s');\n", escapeJS(step.Selector), escapeJS(step.Value)))
				}
			}

		case "check_not_empty":
			b.WriteString(fmt.Sprintf("    const count = await page.locator('%s').count();\n", escapeJS(step.Selector)))
			b.WriteString("    expect(count).toBeGreaterThan(0);\n")

		case "screenshot":
			b.WriteString("    await page.screenshot({ path: 'verification.png', fullPage: true });\n")
		}

		b.WriteString("  });\n\n")
	}

	b.WriteString("});\n")
	return b.String()
}

// resolveURL replaces {{endpoint}} placeholder with the actual endpoint.
func resolveURL(url, endpoint string) string {
	if url == "" {
		return endpoint
	}
	return strings.ReplaceAll(url, "{{endpoint}}", strings.TrimRight(endpoint, "/"))
}

func escapeJS(s string) string {
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

// parseVerificationResults maps Playwright output back to step results.
func parseVerificationResults(s *Scenario, output string, runErr error) *VerificationResult {
	result := &VerificationResult{
		Steps:  make(map[string]VerifyResult),
		Passed: runErr == nil,
	}

	passCount := 0
	for i, step := range s.Verification {
		name := step.Name
		if name == "" {
			name = fmt.Sprintf("step-%d", i+1)
		}

		// Check if this specific test passed or failed in the output
		passed := !strings.Contains(output, fmt.Sprintf("✘") ) && runErr == nil
		errMsg := ""

		// Look for specific test failure
		if strings.Contains(output, name) {
			if strings.Contains(output, "✓ "+name) || strings.Contains(output, "✓  "+name) ||
				strings.Contains(output, "passed") {
				passed = true
			}
			// Extract error for failed tests
			if strings.Contains(output, "✘ "+name) || strings.Contains(output, "✘  "+name) {
				passed = false
				// Try to find the error message after the test name
				idx := strings.Index(output, name)
				if idx >= 0 {
					snippet := output[idx:]
					if nl := strings.Index(snippet, "\n"); nl > 0 {
						errMsg = strings.TrimSpace(snippet[len(name):nl])
					}
				}
			}
		}

		if passed {
			passCount++
		}

		result.Steps[name] = VerifyResult{Passed: passed, Error: errMsg}
	}

	result.Passed = passCount == len(s.Verification)
	result.Summary = fmt.Sprintf("%d/%d verification steps passed", passCount, len(s.Verification))
	return result
}

// discoverEndpoint tries to find the deployed app URL from azd env output.
func discoverEndpoint(workDir string) string {
	// Try azd env get-values
	cmd := exec.Command("azd", "env", "get-values")
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Look for common endpoint variables
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		for _, prefix := range []string{"AZURE_STATIC_WEB_APP_URL=", "SERVICE_WEB_ENDPOINT_URL=", "WEBSITE_URL=", "AZURE_WEBAPP_URL="} {
			if strings.HasPrefix(line, prefix) {
				url := strings.TrimPrefix(line, prefix)
				url = strings.Trim(url, "\"' ")
				if url != "" {
					return url
				}
			}
		}
	}

	return ""
}
