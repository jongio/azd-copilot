// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package commands

import (
	"fmt"

	"github.com/jongio/azd-copilot/cli/src/internal/copilot"
	"github.com/jongio/azd-core/cliout"

	"github.com/spf13/cobra"
)

func NewInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "AI-assisted project initialization",
		Long: `Analyze the current directory and suggest Azure configuration.

This command:
1. Scans the project structure
2. Detects languages, frameworks, and services
3. Suggests azure.yaml configuration
4. Generates Bicep infrastructure
5. Creates project documentation`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runQuickAction(cmd, "init", `Analyze this project directory and help me initialize it for Azure deployment.

1. Scan all subdirectories to detect:
   - Programming languages and frameworks
   - Service types (API, web app, worker, etc.)
   - Existing configuration files
   - Database dependencies

2. Suggest an azure.yaml configuration with:
   - Appropriate Azure services for each component
   - Container Apps for APIs
   - Static Web Apps for frontends
   - Appropriate database services

3. Generate Bicep infrastructure files

4. Create a README.md with deployment instructions

Start by analyzing the current directory structure.`)
		},
	}
}

func NewReviewCommand() *cobra.Command {
	var path string
	var focus string

	cmd := &cobra.Command{
		Use:   "review",
		Short: "AI-powered code review",
		Long: `Perform an AI-powered code review of the project.

Focus areas:
  security    - Security vulnerabilities, secrets, auth issues
  performance - Performance bottlenecks, inefficiencies
  quality     - Code quality, maintainability, best practices`,
		RunE: func(cmd *cobra.Command, args []string) error {
			prompt := fmt.Sprintf(`Perform a thorough code review of this project.

Focus area: %s

Review the code for:
1. Security vulnerabilities (secrets, injection, auth issues)
2. Performance bottlenecks
3. Code quality and maintainability
4. Azure best practices
5. Error handling and edge cases

For each issue found:
- Describe the problem
- Explain the risk/impact
- Provide a specific fix

Start by exploring the codebase structure, then dive into the code.`, focus)

			if path != "" && path != "." {
				prompt += fmt.Sprintf("\n\nFocus on the path: %s", path)
			}

			return runQuickAction(cmd, "review", prompt)
		},
	}

	cmd.Flags().StringVar(&path, "path", ".", "Path to review")
	cmd.Flags().StringVar(&focus, "focus", "all", "Focus area: security, performance, quality, all")

	return cmd
}

func NewFixCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "fix",
		Short: "AI-powered error fixing",
		Long: `Automatically fix build errors and test failures.

This command:
1. Runs the build command
2. Captures any errors
3. Analyzes errors with AI
4. Applies fixes
5. Re-runs until passing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runQuickAction(cmd, "fix", `Fix all build errors and test failures in this project.

1. First, detect the project type and find the build/test commands
2. Run the build command and capture any errors
3. For each error:
   - Analyze the root cause
   - Apply the fix
   - Verify the fix works
4. Run tests and fix any failures
5. Continue until build succeeds and all tests pass

Be thorough - fix one issue at a time and verify before moving on.`)
		},
	}
}

func NewOptimizeCommand() *cobra.Command {
	var focus string

	cmd := &cobra.Command{
		Use:   "optimize",
		Short: "Cost and performance optimization",
		Long: `Analyze and optimize Azure costs and performance.

Focus areas:
  cost        - Reduce Azure spending
  performance - Improve application performance
  both        - Optimize both cost and performance`,
		RunE: func(cmd *cobra.Command, args []string) error {
			prompt := fmt.Sprintf(`Optimize this Azure application for %s.

`, focus)

			if focus == "cost" || focus == "both" {
				prompt += `For COST optimization:
1. Analyze the Bicep/infrastructure files
2. Identify over-provisioned resources
3. Suggest right-sizing (smaller SKUs, reserved instances)
4. Find unused or orphaned resources
5. Recommend auto-scaling configurations
6. Calculate estimated monthly savings

`
			}

			if focus == "performance" || focus == "both" {
				prompt += `For PERFORMANCE optimization:
1. Review application code for bottlenecks
2. Check database queries and indexes
3. Analyze caching opportunities
4. Review network configuration
5. Check for N+1 query problems
6. Suggest async/parallel processing improvements

`
			}

			prompt += `Provide specific, actionable recommendations with code changes.`

			return runQuickAction(cmd, "optimize", prompt)
		},
	}

	cmd.Flags().StringVar(&focus, "focus", "both", "Focus: cost, performance, both")

	return cmd
}

func NewDiagnoseCommand() *cobra.Command {
	var resource string
	var logs bool

	cmd := &cobra.Command{
		Use:   "diagnose",
		Short: "Troubleshoot Azure issues",
		Long: `Diagnose and troubleshoot Azure deployment issues.

This command helps identify and fix:
- Deployment failures
- Runtime errors
- Connectivity issues
- Configuration problems`,
		RunE: func(cmd *cobra.Command, args []string) error {
			prompt := `Diagnose issues with this Azure deployment.

1. Check the deployment status and identify any failures
2. Analyze error messages and logs
3. Identify the root cause
4. Provide step-by-step fix instructions
5. Verify the fix resolves the issue

`
			if resource != "" {
				prompt += fmt.Sprintf("Focus on resource: %s\n", resource)
			}

			if logs {
				prompt += `Include log analysis:
- Check application logs
- Review Azure Monitor logs
- Analyze Container App/App Service logs
`
			}

			return runQuickAction(cmd, "diagnose", prompt)
		},
	}

	cmd.Flags().StringVar(&resource, "resource", "", "Specific resource to diagnose")
	cmd.Flags().BoolVar(&logs, "logs", false, "Include log analysis")

	return cmd
}

// runQuickAction executes a quick action by launching Copilot with a specific prompt
func runQuickAction(cmd *cobra.Command, action string, prompt string) error {
	cliout.Section("🤖", fmt.Sprintf("Azure Copilot CLI: %s", action))
	cliout.Newline()

	// Check if Copilot CLI is installed
	if !copilot.IsCopilotInstalled() {
		cliout.Error("GitHub Copilot CLI not found!")
		cliout.Newline()
		cliout.Hint("Install with: winget install GitHub.Copilot")
		return fmt.Errorf("copilot CLI not installed")
	}

	// Launch Copilot with the specific prompt
	return copilot.Launch(cmd.Context(), copilot.Options{
		Prompt: prompt,
		Agent:  "azure-manager",
	})
}
