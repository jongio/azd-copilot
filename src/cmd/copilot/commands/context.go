// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package commands

import (
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/jongio/azd-core/cliout"
	"github.com/spf13/cobra"
)

func NewContextCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "context",
		Short: "Get the context of the AZD project & environment.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create a new context that includes the AZD access token
			ctx := azdext.WithAccessToken(cmd.Context())

			// Create a new AZD client
			azdClient, err := azdext.NewAzdClient()
			if err != nil {
				return fmt.Errorf("failed to create azd client: %w", err)
			}

			defer azdClient.Close()

			hasEnv := false

			getConfigResponse, err := azdClient.UserConfig().Get(ctx, &azdext.GetUserConfigRequest{
				Path: "",
			})
			if err == nil {
				if getConfigResponse.Found {
					fmt.Printf("%sUser Config%s\n", cliout.Bold, cliout.Reset)
					var userConfig map[string]string
					err := json.Unmarshal(getConfigResponse.Value, &userConfig)
					if err == nil {
						jsonBytes, err := json.MarshalIndent(userConfig, "", "  ")
						if err == nil {
							fmt.Println(string(jsonBytes))
						}
					}
				}
			}

			getProjectResponse, err := azdClient.Project().Get(ctx, &azdext.EmptyRequest{})
			if err == nil {
				cliout.Section("📁", "Project")

				projectValues := map[string]string{
					"Name": getProjectResponse.Project.Name,
					"Path": getProjectResponse.Project.Path,
				}

				for key, value := range projectValues {
					cliout.Label(key, value)
				}
				cliout.Newline()
			} else {
				cliout.Warning("No azd project found in current working directory")
				cliout.Hint("Run 'azd init' to create a new project.")
				return nil
			}

			var currentEnvName string

			getEnvResponse, err := azdClient.Environment().GetCurrent(ctx, &azdext.EmptyRequest{})
			if err == nil {
				currentEnvName = getEnvResponse.Environment.Name
				hasEnv = true
			} else {
				cliout.Warning("No azd environment(s) found.")
				cliout.Hint("Run 'azd env new' to create a new environment.")
				return nil
			}

			environments := []string{}
			envListResponse, err := azdClient.Environment().List(ctx, &azdext.EmptyRequest{})
			if err == nil {
				for _, env := range envListResponse.Environments {
					environments = append(environments, env.Name)
				}
			}

			if len(environments) == 0 {
				fmt.Println("No environments found")
			}

			if hasEnv {
				cliout.Section("🌐", "Environments")
				for _, env := range environments {
					envLine := env
					if env == currentEnvName {
						envLine = fmt.Sprintf("%s%s (selected)%s", cliout.Bold, env, cliout.Reset)
					}

					fmt.Printf("- %s\n", envLine)
				}

				cliout.Newline()

				getValuesResponse, err := azdClient.Environment().GetValues(ctx, &azdext.GetEnvironmentRequest{
					Name: currentEnvName,
				})
				if err == nil {
					cliout.Section("📝", "Environment values")
					for _, pair := range getValuesResponse.KeyValues {
						cliout.Label(pair.Key, pair.Value)
					}
					cliout.Newline()
				}

				deploymentContextResponse, err := azdClient.Deployment().GetDeploymentContext(ctx, &azdext.EmptyRequest{})
				if err == nil {
					scopeMap := map[string]string{
						"Tenant ID":       deploymentContextResponse.AzureContext.Scope.TenantId,
						"Subscription ID": deploymentContextResponse.AzureContext.Scope.SubscriptionId,
						"Location":        deploymentContextResponse.AzureContext.Scope.Location,
						"Resource Group":  deploymentContextResponse.AzureContext.Scope.ResourceGroup,
					}

					cliout.Section("☁️", "Deployment Context")
					for key, value := range scopeMap {
						if value == "" {
							value = "N/A"
						}

						cliout.Label(key, value)
					}
					cliout.Newline()

					cliout.Section("📦", "Provisioned Azure Resources")
					for _, resourceId := range deploymentContextResponse.AzureContext.Resources {
						resource, err := arm.ParseResourceID(resourceId)
						if err == nil {
							fmt.Printf("- %s (%s)\n", resource.Name, resource.ResourceType.String())
						}
					}
					cliout.Newline()
				}
			}

			return nil
		},
	}
}
