// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package commands

import (
	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/spf13/cobra"
)

// NewMetadataCommand creates a metadata command using the official azdext SDK.
func NewMetadataCommand(schemaVersion, extensionId string, rootCmdProvider func() *cobra.Command) *cobra.Command {
	return azdext.NewMetadataCommand(schemaVersion, extensionId, rootCmdProvider)
}
