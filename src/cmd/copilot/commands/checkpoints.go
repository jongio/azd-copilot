// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package commands

import (
	"fmt"
	"strings"

	"github.com/jongio/azd-copilot/src/internal/checkpoint"
	"github.com/jongio/azd-copilot/src/internal/copilot"
	"github.com/jongio/azd-core/cliout"

	"github.com/spf13/cobra"
)

var (
	checkpointPhase string
	checkpointType  string
	keepLatest      int
)

func NewCheckpointsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkpoints",
		Short: "Manage build checkpoints",
		Long: `Manage build checkpoints stored in docs/checkpoints/.

Checkpoints are saved after each build phase (spec, design, develop, quality, deploy).
You can resume from any checkpoint if a build is interrupted.

Checkpoint Types:
  phase     - Created after completing a build phase
  task      - Created after completing individual tasks  
  snapshot  - Full file backup before risky operations
  recovery  - Created when errors occur for debugging`,
		RunE: runCheckpointsList,
	}

	cmd.AddCommand(newCheckpointsListCommand())
	cmd.AddCommand(newCheckpointsShowCommand())
	cmd.AddCommand(newCheckpointsResumeCommand())
	cmd.AddCommand(newCheckpointsCreateCommand())
	cmd.AddCommand(newCheckpointsClearCommand())

	return cmd
}

func newCheckpointsListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all checkpoints",
		RunE:  runCheckpointsList,
	}

	cmd.Flags().StringVar(&checkpointPhase, "phase", "", "Filter by phase (spec, design, develop, quality, deploy)")
	cmd.Flags().StringVar(&checkpointType, "type", "", "Filter by type (phase, task, snapshot, recovery)")

	return cmd
}

func newCheckpointsShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <checkpoint-id>",
		Short: "Show checkpoint details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cp, err := checkpoint.Get(args[0])
			if err != nil {
				return err
			}

			cliout.Section("📍", fmt.Sprintf("Checkpoint: %s", cp.ID))
			cliout.Newline()

			// Basic info
			fmt.Printf("  Type:        %s\n", cp.Type)
			fmt.Printf("  Trigger:     %s\n", cp.Trigger)
			fmt.Printf("  Phase:       %s\n", cp.Phase)
			fmt.Printf("  Created:     %s\n", cp.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("  Can Resume:  %v\n", cp.CanResume)
			cliout.Newline()

			// Description
			fmt.Printf("  Description: %s\n", cp.Description)
			cliout.Newline()

			// Completed phases
			if len(cp.CompletedPhases) > 0 {
				fmt.Println("  Completed Phases:")
				for _, p := range cp.CompletedPhases {
					fmt.Printf("    ✓ %s\n", p)
				}
				cliout.Newline()
			}

			// Files
			allFiles := make([]string, 0, len(cp.Files.Created)+len(cp.Files.Modified))
			allFiles = append(allFiles, cp.Files.Created...)
			allFiles = append(allFiles, cp.Files.Modified...)
			if len(allFiles) > 0 {
				fmt.Printf("  Files (%d):\n", len(allFiles))
				for _, f := range allFiles {
					fmt.Printf("    • %s\n", f)
				}
				cliout.Newline()
			}

			// Context
			if cp.Context.ErrorMessage != "" {
				cliout.Warning("Error Context:")
				fmt.Printf("    %s\n", cp.Context.ErrorMessage)
				cliout.Newline()
			}

			if cp.Context.SessionID != "" {
				fmt.Printf("  Session ID: %s\n", cp.Context.SessionID)
			}

			return nil
		},
	}
}

func newCheckpointsResumeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "resume [checkpoint-id]",
		Short: "Resume from a checkpoint",
		Long: `Resume building from a specific checkpoint.

If no checkpoint ID is provided, resumes from the latest checkpoint.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var cp *checkpoint.Checkpoint
			var err error

			if len(args) > 0 {
				cp, err = checkpoint.Get(args[0])
				if err != nil {
					return err
				}
			} else {
				// Check for interrupted build first
				cp, err = checkpoint.DetectInterrupted()
				if err != nil {
					return err
				}
				if cp == nil {
					cp, err = checkpoint.Latest()
					if err != nil {
						return err
					}
				}
				if cp == nil {
					cliout.Warning("No checkpoints found.")
					cliout.Newline()
					fmt.Println("Run 'azd copilot build' to start a new build.")
					return nil
				}
			}

			cliout.Section("📍", fmt.Sprintf("Resuming from checkpoint: %s", cp.ID))
			fmt.Printf("   Type:        %s\n", cp.Type)
			fmt.Printf("   Phase:       %s\n", cp.Phase)
			fmt.Printf("   Description: %s\n", cp.Description)

			fileCount := len(cp.Files.Created) + len(cp.Files.Modified)
			fmt.Printf("   Files:       %d created/modified\n", fileCount)
			fmt.Println()

			// Generate resume prompt and launch copilot
			prompt := checkpoint.GenerateResumePrompt(cp)

			return copilot.Launch(cmd.Context(), copilot.Options{
				Prompt: prompt,
				Agent:  "azure-manager",
			})
		},
	}
}

func newCheckpointsCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create <description>",
		Short: "Create a manual checkpoint",
		Long: `Create a manual checkpoint with the current project state.

Use this before making significant changes or experiments.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			description := strings.Join(args, " ")

			// Get current files (simplified - just docs folder)
			files := []string{}

			cp, err := checkpoint.SaveWithOptions(checkpoint.SaveOptions{
				Phase:       checkpoint.Phase(checkpointPhase),
				Type:        checkpoint.TypeManual,
				Trigger:     checkpoint.TriggerManual,
				Description: description,
				Files:       checkpoint.FileState{Created: files},
			})
			if err != nil {
				return fmt.Errorf("failed to create checkpoint: %w", err)
			}

			cliout.Success("Checkpoint created: %s", cp.ID)
			cliout.Newline()
			fmt.Printf("  Description: %s\n", cp.Description)
			fmt.Printf("  Phase:       %s\n", cp.Phase)
			cliout.Newline()
			cliout.Hint(fmt.Sprintf("Resume with: azd copilot checkpoints resume %s", cp.ID))

			return nil
		},
	}
}

func newCheckpointsClearCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Delete checkpoints",
		Long: `Delete checkpoints. By default deletes all checkpoints.

Use --keep-latest N to keep the N most recent checkpoints.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if keepLatest > 0 {
				if err := checkpoint.KeepLatest(keepLatest); err != nil {
					return fmt.Errorf("failed to clean up checkpoints: %w", err)
				}
				cliout.Success("Kept %d most recent checkpoints", keepLatest)
				return nil
			}

			if err := checkpoint.Clear(); err != nil {
				return fmt.Errorf("failed to clear checkpoints: %w", err)
			}
			cliout.Success("All checkpoints cleared")
			return nil
		},
	}

	cmd.Flags().IntVar(&keepLatest, "keep-latest", 0, "Keep N most recent checkpoints")

	return cmd
}

func runCheckpointsList(cmd *cobra.Command, args []string) error {
	checkpoints, err := checkpoint.List()
	if err != nil {
		return err
	}

	// Apply filters
	var filtered []checkpoint.Checkpoint
	for _, cp := range checkpoints {
		if checkpointPhase != "" && string(cp.Phase) != checkpointPhase {
			continue
		}
		if checkpointType != "" && string(cp.Type) != checkpointType {
			continue
		}
		filtered = append(filtered, cp)
	}
	checkpoints = filtered

	if len(checkpoints) == 0 {
		cliout.Warning("No checkpoints found.")
		cliout.Newline()
		fmt.Println("Checkpoints are created during 'azd copilot build' after each phase.")
		return nil
	}

	cliout.Section("📍", fmt.Sprintf("Checkpoints (%d)", len(checkpoints)))
	cliout.Newline()

	for i, cp := range checkpoints {
		marker := "  "
		if i == 0 {
			marker = "→ " // latest
		}

		// Type indicator
		typeIcon := "◆"
		switch cp.Type {
		case checkpoint.TypePhase:
			typeIcon = "●"
		case checkpoint.TypeTask:
			typeIcon = "○"
		case checkpoint.TypeSnapshot:
			typeIcon = "◈"
		case checkpoint.TypeRecovery:
			typeIcon = "⚠"
		case checkpoint.TypeManual:
			typeIcon = "★"
		}

		fmt.Printf("%s%s %s%s%s\n", marker, typeIcon, cliout.Cyan, cp.ID, cliout.Reset)
		fmt.Printf("     Phase: %-10s  Type: %s\n", cp.Phase, cp.Type)
		fmt.Printf("     %s\n", cp.Description)

		fileCount := len(cp.Files.Created) + len(cp.Files.Modified)
		fmt.Printf("     %s | %d files\n", cp.CreatedAt.Format("2006-01-02 15:04"), fileCount)
		cliout.Newline()
	}

	fmt.Println("Commands:")
	fmt.Println("  Show details:  azd copilot checkpoints show <id>")
	fmt.Println("  Resume:        azd copilot checkpoints resume [id]")
	return nil
}
