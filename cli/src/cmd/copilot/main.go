package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/jongio/azd-copilot/cli/src/cmd/copilot/commands"
	"github.com/jongio/azd-copilot/cli/src/internal/assets"
	"github.com/jongio/azd-copilot/cli/src/internal/copilot"
	selfskills "github.com/jongio/azd-copilot/cli/src/internal/skills"
	"github.com/jongio/azd-core/cliout"
	"github.com/jongio/azd-core/logutil"

	"github.com/common-nighthawk/go-figure"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/propagation"
)

var (
	outputFormat   string
	debugMode      bool
	structuredLogs bool
	cwdFlag        string

	// Root command flags for copilot session
	prompt     string
	resume     bool
	yolo       bool
	agent      string
	model      string
	addDirs    []string
	verbose    bool
	noBanner   bool
	forceColor bool
)

func main() {
	// Set version in copilot package
	copilot.Version = commands.Version

	rootCmd := newRootCmd()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "copilot",
		Short: "Azure Copilot CLI - AI-powered Azure development assistant",
		Long: fmt.Sprintf(`Azure Copilot CLI is an Azure Developer CLI extension that integrates GitHub Copilot CLI
with %d specialized Azure agents and %d focused skills for Azure development.

When run without subcommands, starts an interactive Copilot session with Azure context.`, assets.AgentCount(), assets.SkillCount()),
		Example: `  # Start interactive session
  azd copilot

  # Start with a specific prompt
  azd copilot -p "help me deploy this app to Azure"

  # Resume last session
  azd copilot --resume

  # Use a specific agent
  azd copilot --agent azure-architect

  # Auto-approve mode (careful!)
  azd copilot --yolo`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Inject OTel trace context from env vars while preserving cobra's signal handling
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			if parent := os.Getenv("TRACEPARENT"); parent != "" {
				tc := propagation.TraceContext{}
				ctx = tc.Extract(ctx, propagation.MapCarrier{
					"traceparent": parent,
					"tracestate":  os.Getenv("TRACESTATE"),
				})
			}
			cmd.SetContext(ctx)

			// Change working directory if --cwd is specified
			if cwdFlag != "" {
				if err := os.Chdir(cwdFlag); err != nil {
					return fmt.Errorf("failed to change to directory '%s': %w", cwdFlag, err)
				}
			}

			// Handle force color
			if forceColor {
				cliout.ForceColor()
				_ = os.Setenv("FORCE_COLOR", "1")
			}

			// Set global output format and debug mode
			if debugMode {
				_ = os.Setenv("AZD_DEBUG", "true")
				_ = os.Setenv("AZD_COPILOT_DEBUG", "true")
				slog.SetLogLoggerLevel(slog.LevelDebug)
			}

			// Configure logging
			logutil.SetupLogger(debugMode, structuredLogs)

			// Install azd-copilot self-skill
			if err := selfskills.InstallSkill(); err != nil {
				if debugMode {
					slog.Debug("Failed to install copilot self-skill", "error", err)
				}
			}

			// Log startup in debug mode
			if debugMode {
				logutil.Debug("Starting azd copilot extension",
					"version", commands.Version,
					"command", cmd.Name(),
					"args", args,
					"cwd", cwdFlag,
				)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default behavior: start interactive Copilot session
			return runCopilotSession(cmd)
		},
	}

	// Add global flags
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "default", "Output format (default, json)")
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().BoolVar(&structuredLogs, "structured-logs", false, "Enable structured JSON logging to stderr")
	rootCmd.PersistentFlags().StringVarP(&cwdFlag, "cwd", "C", "", "Sets the current working directory")
	rootCmd.PersistentFlags().BoolVar(&forceColor, "color", false, "Force colored output")

	// Add root command flags for copilot session
	rootCmd.Flags().StringVarP(&prompt, "prompt", "p", "", "Run with a specific prompt")
	rootCmd.Flags().BoolVarP(&resume, "resume", "r", false, "Resume the last session")
	rootCmd.Flags().BoolVarP(&yolo, "yolo", "y", false, "Auto-approve all actions (use with caution)")
	rootCmd.Flags().StringVarP(&agent, "agent", "a", "", "Use a specific agent (default: azure-manager)")
	rootCmd.Flags().StringVarP(&model, "model", "m", "", "Use a specific AI model")
	rootCmd.Flags().StringSliceVar(&addDirs, "add-dir", nil, "Additional directories to include")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.Flags().BoolVar(&noBanner, "no-banner", false, "Skip the banner")

	// Register all commands
	rootCmd.AddCommand(
		commands.NewVersionCommand(&outputFormat),
		commands.NewListenCommand(),
		commands.NewAgentsCommand(),
		commands.NewSkillsCommand(),
		commands.NewSessionsCommand(),
		commands.NewContextCommand(),
		commands.NewCheckpointsCommand(),
		commands.NewBuildCommand(),
		commands.NewSpecCommand(),
		commands.NewMCPCommand(),
		commands.NewMetadataCommand(newRootCmd),
		// Quick actions
		commands.NewInitCommand(),
		commands.NewReviewCommand(),
		commands.NewFixCommand(),
		commands.NewOptimizeCommand(),
		commands.NewDiagnoseCommand(),
	)

	return rootCmd
}

func runCopilotSession(cmd *cobra.Command) error {
	// Print banner unless --no-banner or --prompt
	if !noBanner && prompt == "" {
		printBanner()
	}

	// Check if Copilot CLI is installed
	if !copilot.IsCopilotInstalled() {
		cliout.Error("GitHub Copilot CLI not found!")
		cliout.Newline()
		fmt.Println("Install with one of:")
		fmt.Println("  • winget install GitHub.Copilot")
		fmt.Println("  • npm install -g @github/copilot")
		cliout.Newline()
		return fmt.Errorf("copilot CLI not installed")
	}

	// Configure MCP servers for Copilot CLI
	if err := copilot.ConfigureMCPServer(); err != nil {
		if debugMode {
			fmt.Fprintf(os.Stderr, "Warning: failed to configure MCP servers: %v\n", err)
		}
	}

	// Install agents and skills to ~/.azd/copilot/
	assetDirs, err := setupAgentsAndSkills()
	if err != nil {
		if debugMode {
			fmt.Fprintf(os.Stderr, "Warning: failed to install agents/skills: %v\n", err)
		}
	}

	// Build project context
	projectContext := buildProjectContext()

	// Launch Copilot CLI
	return copilot.Launch(cmd.Context(), copilot.Options{
		Prompt:         prompt,
		Resume:         resume,
		Yolo:           yolo,
		Agent:          agent,
		Model:          model,
		AddDirs:        append(addDirs, assetDirs...),
		Verbose:        verbose,
		Debug:          debugMode,
		ProjectContext: projectContext,
	})
}

func printBanner() {
	banner := figure.NewFigure("Azure Copilot CLI", "small", true)
	fmt.Printf("%s%s%s", cliout.Cyan, banner.String(), cliout.Reset)
	cliout.Newline()
	fmt.Printf("%sAI-powered Azure development assistant%s\n", cliout.Bold, cliout.Reset)
	fmt.Printf("Built on Copilot SDK • Azure MCP • Azure Developer CLI • Azure Agents & Skills\n")
	agentCount := 0
	if agents, err := assets.ListAgents(); err == nil {
		agentCount = len(agents)
	}
	skillCount := 0
	if skills, err := assets.ListSkills(); err == nil {
		skillCount = len(skills)
	}
	fmt.Printf("Version %s • %d agents • %d skills\n", commands.Version, agentCount, skillCount)
	cliout.Newline()
}

func setupAgentsAndSkills() ([]string, error) {
	var dirs []string

	// Install agents
	agentsDir, _, err := assets.InstallAgents()
	if err != nil {
		return nil, fmt.Errorf("failed to install agents: %w", err)
	}
	dirs = append(dirs, agentsDir)

	// Install skills
	skillsDir, _, err := assets.InstallSkills()
	if err != nil {
		return nil, fmt.Errorf("failed to install skills: %w", err)
	}
	dirs = append(dirs, skillsDir)

	return dirs, nil
}

func buildProjectContext() *copilot.ProjectContext {
	// Try to detect azd project context
	// This is a simplified version - full implementation would use azdext client
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}

	// Check for azure.yaml
	if _, err := os.Stat("azure.yaml"); os.IsNotExist(err) {
		return nil
	}

	// Basic project context
	ctx := &copilot.ProjectContext{
		Path: cwd,
	}

	// Try to extract project name from azure.yaml
	if content, err := os.ReadFile("azure.yaml"); err == nil {
		// Simple extraction - look for name: line
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "name:") {
				ctx.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
				break
			}
		}
	}

	return ctx
}
