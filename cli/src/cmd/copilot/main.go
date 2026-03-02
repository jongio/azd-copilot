package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/jongio/azd-copilot/cli/src/cmd/copilot/commands"
	"github.com/jongio/azd-copilot/cli/src/internal/assets"
	"github.com/jongio/azd-copilot/cli/src/internal/copilot"
	selfskills "github.com/jongio/azd-copilot/cli/src/internal/skills"
	"github.com/jongio/azd-core/cliout"
	"github.com/jongio/azd-core/logutil"

	"github.com/common-nighthawk/go-figure"
	"github.com/spf13/cobra"
)

var (
	structuredLogs bool

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

	// SDK extension context
	extCtx *azdext.ExtensionContext
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
	rootCmd, ec := azdext.NewExtensionRootCommand(azdext.ExtensionCommandOptions{
		Name:    "copilot",
		Version: commands.Version,
		Short:   "Azure Copilot CLI - AI-powered Azure development assistant",
		Long: fmt.Sprintf(`Azure Copilot CLI is an Azure Developer CLI extension that integrates GitHub Copilot CLI
with %d specialized Azure agents and %d focused skills for Azure development.

When run without subcommands, starts an interactive Copilot session with Azure context.`, assets.AgentCount(), assets.SkillCount()),
	})
	extCtx = ec

	rootCmd.Example = `  # Start interactive session
  azd copilot

  # Start with a specific prompt
  azd copilot -p "help me deploy this app to Azure"

  # Resume last session
  azd copilot --resume

  # Use a specific agent
  azd copilot --agent azure-architect

  # Auto-approve mode (careful!)
  azd copilot --yolo`

	// Chain extension-specific PersistentPreRunE after the SDK's
	sdkPreRunE := rootCmd.PersistentPreRunE
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if sdkPreRunE != nil {
			if err := sdkPreRunE(cmd, args); err != nil {
				return err
			}
		}

		// Handle force color
		if forceColor {
			cliout.ForceColor()
			_ = os.Setenv("FORCE_COLOR", "1")
		}

		// Set global output format and debug mode
		if extCtx.Debug {
			_ = os.Setenv("AZD_DEBUG", "true")
			_ = os.Setenv("AZD_COPILOT_DEBUG", "true")
			slog.SetLogLoggerLevel(slog.LevelDebug)
		}

		// Configure logging
		logutil.SetupLogger(extCtx.Debug, structuredLogs)

		// Install azd-copilot self-skill
		if err := selfskills.InstallSkill(); err != nil {
			if extCtx.Debug {
				slog.Debug("Failed to install copilot self-skill", "error", err)
			}
		}

		// Log startup in debug mode
		if extCtx.Debug {
			logutil.Debug("Starting azd copilot extension",
				"version", commands.Version,
				"command", cmd.Name(),
				"args", args,
				"cwd", extCtx.Cwd,
			)
		}

		return nil
	}

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Default behavior: start interactive Copilot session
		return runCopilotSession(cmd)
	}

	// Add extension-specific persistent flags
	rootCmd.PersistentFlags().BoolVar(&structuredLogs, "structured-logs", false, "Enable structured JSON logging to stderr")
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
		commands.NewVersionCommand(&extCtx.OutputFormat),
		commands.NewListenCommand(),
		commands.NewAgentsCommand(),
		commands.NewSkillsCommand(),
		commands.NewSessionsCommand(),
		commands.NewContextCommand(),
		commands.NewCheckpointsCommand(),
		commands.NewBuildCommand(),
		commands.NewSpecCommand(),
		commands.NewMCPCommand(),
		commands.NewMetadataCommand("1.0", "jongio.azd.copilot", newRootCmd),
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
		if extCtx.Debug {
			fmt.Fprintf(os.Stderr, "Warning: failed to configure MCP servers: %v\n", err)
		}
	}

	// Install agents and skills to ~/.azd/copilot/
	assetDirs, err := setupAgentsAndSkills()
	if err != nil {
		if extCtx.Debug {
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
		Debug:          extCtx.Debug,
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
