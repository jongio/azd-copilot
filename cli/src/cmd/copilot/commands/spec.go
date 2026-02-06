// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package commands

import (
	"fmt"
	"os"

	"github.com/jongio/azd-copilot/cli/src/internal/spec"
	"github.com/jongio/azd-core/cliout"

	"github.com/spf13/cobra"
)

func NewSpecCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spec",
		Short: "View and manage the project spec",
		Long: `View and manage the project specification (docs/spec.md).

The spec is generated during 'azd copilot build' and contains:
- Project name and description
- Services to build
- Azure resources needed
- Architecture decisions
- Estimated costs`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return viewSpec()
		},
	}

	cmd.AddCommand(newSpecShowCommand())
	cmd.AddCommand(newSpecEditCommand())
	cmd.AddCommand(newSpecDeleteCommand())

	return cmd
}

func newSpecShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show the current spec",
		RunE: func(cmd *cobra.Command, args []string) error {
			return viewSpec()
		},
	}
}

func newSpecEditCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Open spec in default editor",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !spec.Exists() {
				cliout.Warning("No spec found.")
				cliout.Newline()
				cliout.Hint("Run 'azd copilot build \"description\"' to generate a spec.")
				return nil
			}

			return spec.OpenInEditor()
		},
	}
}

func newSpecDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete the spec",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !spec.Exists() {
				cliout.Warning("No spec found.")
				return nil
			}

			specPath := spec.Path()
			if err := os.Remove(specPath); err != nil {
				return fmt.Errorf("failed to delete spec: %w", err)
			}

			cliout.Success("Spec deleted")
			return nil
		},
	}
}

func viewSpec() error {
	if !spec.Exists() {
		cliout.Warning("No spec found.")
		cliout.Newline()
		cliout.Hint("Run 'azd copilot build \"description\"' to generate a spec.")
		return nil
	}

	content, err := spec.Read()
	if err != nil {
		return err
	}

	fmt.Println(content)
	return nil
}
