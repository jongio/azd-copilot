// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package commands

import (
	"encoding/json"
	"fmt"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/spf13/cobra"
)

// NewMetadataCommand creates a metadata command using the official azdext SDK.
func NewMetadataCommand(rootCmdProvider func() *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:    "metadata",
		Short:  "Output extension metadata for IntelliSense",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			root := rootCmdProvider()
			metadata := azdext.GenerateExtensionMetadata("1.0", "jongio.azd.copilot", root)
			data, err := json.MarshalIndent(metadata, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal metadata: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		},
	}
}
