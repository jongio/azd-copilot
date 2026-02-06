// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package commands

import (
	"fmt"
	"strings"

	"github.com/jongio/azd-copilot/cli/src/internal/copilot"
	"github.com/jongio/azd-copilot/cli/src/internal/spec"
	"github.com/jongio/azd-core/cliout"

	"github.com/spf13/cobra"
)

var (
	buildMode    string
	buildApprove bool
)

func NewBuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build [description]",
		Short: "Generate a complete Azure application from description",
		Long: `Generate a complete Azure application from a natural language description.

This is the core AzureCopilot feature: describe what you want, and Copilot builds it.

The build process:
1. Analyzes your description
2. Generates a spec (docs/spec.md)
3. Waits for your approval (or use --approve to skip)
4. Generates code, infrastructure, tests, and documentation
5. Runs preflight checks
6. Deploys to Azure (if approved)

Project Modes:
  prototype   - Fast, minimal setup, free tiers (POC, demo, experiment)
  production  - Full quality gates, security, testing, monitoring (default)`,
		Example: `  # Build a todo API (generates spec first)
  azd copilot build "create a todo API with PostgreSQL and React frontend"

  # Build a quick prototype
  azd copilot build --mode prototype "demo chat app with Azure OpenAI"

  # Skip spec approval and proceed directly
  azd copilot build --approve "REST API for inventory management"
  
  # Continue after editing spec
  azd copilot build --approve`,
		RunE: func(cmd *cobra.Command, args []string) error {
			description := strings.Join(args, " ")

			// If --approve and no description, check for existing spec
			if description == "" && buildApprove {
				if spec.Exists() {
					return runBuildFromSpec(cmd)
				}
				return fmt.Errorf("no spec found. Run 'azd copilot build \"description\"' first")
			}

			if description == "" {
				return fmt.Errorf("please provide a description of what you want to build")
			}
			return runBuild(cmd, description)
		},
	}

	cmd.Flags().StringVar(&buildMode, "mode", "", "Project mode: prototype or production (auto-detected if not specified)")
	cmd.Flags().BoolVar(&buildApprove, "approve", false, "Auto-approve spec and proceed with generation")

	return cmd
}

func runBuild(cmd *cobra.Command, description string) error {
	cliout.Section("🏗️", "Azure Copilot: Building your application")
	cliout.Newline()

	// Check if Copilot CLI is installed
	if !copilot.IsCopilotInstalled() {
		cliout.Error("GitHub Copilot CLI not found!")
		cliout.Newline()
		cliout.Hint("Install with: winget install GitHub.Copilot")
		return fmt.Errorf("copilot CLI not installed")
	}

	// Detect project mode from description if not specified
	detectedMode := buildMode
	if detectedMode == "" {
		detectedMode = detectProjectMode(description)
		cliout.Warning("📋 Detected mode: %s", detectedMode)
	} else {
		cliout.Warning("📋 Mode: %s", detectedMode)
	}
	cliout.Newline()

	// If not auto-approve, generate spec first
	if !buildApprove {
		cliout.Info("📝 Generating spec...")
		fmt.Println("   Spec will be saved to: docs/spec.md")
		cliout.Newline()

		prompt := spec.GeneratePrompt(description, detectedMode)

		// Launch Copilot to generate spec
		if err := copilot.Launch(cmd.Context(), copilot.Options{
			Prompt: prompt,
			Agent:  "azure-manager",
		}); err != nil {
			return err
		}

		// Check if spec was created
		if spec.Exists() {
			cliout.Newline()
			cliout.Success("Spec generated at docs/spec.md")
			cliout.Newline()
			fmt.Println("Next steps:")
			fmt.Println("  1. Review the spec: azd copilot spec")
			fmt.Println("  2. Edit if needed:  azd copilot spec edit")
			fmt.Println("  3. Proceed:         azd copilot build --approve")
		}
		return nil
	}

	// --approve: Build using the spec or description
	return runBuildFromSpec(cmd)
}

func runBuildFromSpec(cmd *cobra.Command) error {
	cliout.Section("🚀", "Building from spec...")
	cliout.Newline()

	var prompt string
	if spec.Exists() {
		content, err := spec.Read()
		if err != nil {
			return err
		}
		prompt = buildFromSpecPrompt(content)
	} else {
		return fmt.Errorf("no spec found. Run 'azd copilot build \"description\"' first")
	}

	// Launch Copilot with the build prompt
	return copilot.Launch(cmd.Context(), copilot.Options{
		Prompt: prompt,
		Agent:  "azure-manager",
	})
}

func buildFromSpecPrompt(specContent string) string {
	var sb strings.Builder

	sb.WriteString("# Build Application from Spec\n\n")
	sb.WriteString("The user has approved the following specification. Now generate the complete application.\n\n")
	sb.WriteString("## Approved Spec\n\n")
	sb.WriteString(specContent)
	sb.WriteString("\n\n")
	sb.WriteString("## Instructions\n\n")
	sb.WriteString("Generate all code, infrastructure, tests, and documentation according to this spec.\n\n")
	sb.WriteString("Follow these phases:\n\n")
	sb.WriteString("### Phase 1: Design\n")
	sb.WriteString("- Finalize architecture decisions\n")
	sb.WriteString("- Create azure.yaml\n\n")
	sb.WriteString("### Phase 2: Develop\n")
	sb.WriteString("- Generate backend code\n")
	sb.WriteString("- Generate frontend code (if applicable)\n")
	sb.WriteString("- Generate database schema and migrations\n\n")
	sb.WriteString("### Phase 3: Quality\n")
	sb.WriteString("- Generate tests\n")
	sb.WriteString("- Run linter\n")
	sb.WriteString("- Security review\n\n")
	sb.WriteString("### Phase 4: Infrastructure\n")
	sb.WriteString("- Generate Bicep files in /infra\n")
	sb.WriteString("- Generate CI/CD pipeline\n")
	sb.WriteString("- Generate documentation\n\n")
	sb.WriteString("Save checkpoints to `docs/checkpoints/` after each phase.\n")

	return sb.String()
}

func detectProjectMode(description string) string {
	desc := strings.ToLower(description)

	// Prototype signals
	prototypeSignals := []string{
		"prototype", "poc", "proof of concept", "demo", "spike",
		"experiment", "hackathon", "quick", "fast", "simple",
		"basic", "minimal", "mvp", "test", "try", "explore",
		"throwaway", "temporary", "just want to see", "learning",
		"tutorial", "example", "sample", "for myself", "personal",
		"side project", "deadline tomorrow", "need it today", "asap",
	}

	// Production signals
	productionSignals := []string{
		"production", "prod", "enterprise", "commercial",
		"customer-facing", "user-facing", "public", "launch",
		"release", "go-live", "ship", "real users", "customers",
		"scale", "scalable", "high availability", "ha",
		"compliance", "gdpr", "hipaa", "soc2", "pci",
		"secure", "security-first", "zero-trust", "business",
		"company", "organization", "team", "paying customers",
		"revenue", "monetize", "sla", "uptime", "reliability",
	}

	prototypeScore := 0
	productionScore := 0

	for _, signal := range prototypeSignals {
		if strings.Contains(desc, signal) {
			prototypeScore++
		}
	}

	for _, signal := range productionSignals {
		if strings.Contains(desc, signal) {
			productionScore++
		}
	}

	if prototypeScore > productionScore {
		return "prototype"
	}

	// Default to production (safer)
	return "production"
}
