// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package commands

import (
	"context"
	"fmt"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/mark3labs/mcp-go/mcp"
)

// registerGRPCTools registers MCP tools that wrap azd gRPC services.
func registerGRPCTools(builder *azdext.MCPServerBuilder) {
	registerEnvironmentTools(builder)
	registerDeploymentTools(builder)
	registerAccountTools(builder)
	registerWorkflowTools(builder)
	registerComposeTools(builder)
}

// newAzdClient creates a new azd gRPC client and returns it along with a context
// that includes the access token for authentication.
func newAzdClient(ctx context.Context) (context.Context, *azdext.AzdClient, error) {
	client, err := azdext.NewAzdClient()
	if err != nil {
		return ctx, nil, fmt.Errorf("failed to create azd client: %w", err)
	}
	return azdext.WithAccessToken(ctx), client, nil
}

func registerEnvironmentTools(builder *azdext.MCPServerBuilder) {
	// Tool: list_environments
	builder.AddTool("list_environments",
		func(ctx context.Context, args azdext.ToolArgs) (*mcp.CallToolResult, error) {
			ctx, client, err := newAzdClient(ctx)
			if err != nil {
				return azdext.MCPErrorResult("%s", err), nil
			}
			defer client.Close()

			resp, err := client.Environment().List(ctx, &azdext.EmptyRequest{})
			if err != nil {
				return azdext.MCPErrorResult("listing environments: %s", err), nil
			}

			type envInfo struct {
				Name    string `json:"name"`
				Local   bool   `json:"local"`
				Remote  bool   `json:"remote"`
				Default bool   `json:"default"`
			}

			envs := make([]envInfo, 0, len(resp.Environments))
			for _, e := range resp.Environments {
				envs = append(envs, envInfo{
					Name:    e.Name,
					Local:   e.Local,
					Remote:  e.Remote,
					Default: e.Default,
				})
			}

			return azdext.MCPJSONResult(envs), nil
		},
		azdext.MCPToolOptions{
			Description: "List all azd environments",
			ReadOnly:    true,
		},
	)

	// Tool: get_environment_values
	builder.AddTool("get_environment_values",
		func(ctx context.Context, args azdext.ToolArgs) (*mcp.CallToolResult, error) {
			name, err := args.RequireString("environment_name")
			if err != nil || name == "" {
				return azdext.MCPErrorResult("environment_name is required"), nil
			}

			ctx, client, err := newAzdClient(ctx)
			if err != nil {
				return azdext.MCPErrorResult("%s", err), nil
			}
			defer client.Close()

			resp, err := client.Environment().GetValues(ctx, &azdext.GetEnvironmentRequest{Name: name})
			if err != nil {
				return azdext.MCPErrorResult("getting environment values: %s", err), nil
			}

			kvMap := make(map[string]string, len(resp.KeyValues))
			for _, kv := range resp.KeyValues {
				kvMap[kv.Key] = kv.Value
			}

			return azdext.MCPJSONResult(kvMap), nil
		},
		azdext.MCPToolOptions{
			Description: "Get all key-value pairs for an azd environment",
			ReadOnly:    true,
		},
		mcp.WithString("environment_name", mcp.Required(), mcp.Description("Name of the environment")),
	)

	// Tool: set_environment_value
	builder.AddTool("set_environment_value",
		func(ctx context.Context, args azdext.ToolArgs) (*mcp.CallToolResult, error) {
			envName, err := args.RequireString("environment_name")
			if err != nil || envName == "" {
				return azdext.MCPErrorResult("environment_name is required"), nil
			}
			key, err := args.RequireString("key")
			if err != nil || key == "" {
				return azdext.MCPErrorResult("key is required"), nil
			}
			value, err := args.RequireString("value")
			if err != nil {
				return azdext.MCPErrorResult("value is required"), nil
			}

			ctx, client, err := newAzdClient(ctx)
			if err != nil {
				return azdext.MCPErrorResult("%s", err), nil
			}
			defer client.Close()

			_, err = client.Environment().SetValue(ctx, &azdext.SetEnvRequest{
				EnvName: envName,
				Key:     key,
				Value:   value,
			})
			if err != nil {
				return azdext.MCPErrorResult("setting environment value: %s", err), nil
			}

			return mcp.NewToolResultText(fmt.Sprintf("Successfully set %s in environment %s", key, envName)), nil
		},
		azdext.MCPToolOptions{
			Description: "Set a key-value pair in an azd environment",
			Destructive: true,
		},
		mcp.WithString("environment_name", mcp.Required(), mcp.Description("Name of the environment")),
		mcp.WithString("key", mcp.Required(), mcp.Description("Key to set")),
		mcp.WithString("value", mcp.Required(), mcp.Description("Value to set")),
	)
}

func registerDeploymentTools(builder *azdext.MCPServerBuilder) {
	// Tool: get_deployment_info
	builder.AddTool("get_deployment_info",
		func(ctx context.Context, args azdext.ToolArgs) (*mcp.CallToolResult, error) {
			ctx, client, err := newAzdClient(ctx)
			if err != nil {
				return azdext.MCPErrorResult("%s", err), nil
			}
			defer client.Close()

			resp, err := client.Deployment().GetDeployment(ctx, &azdext.EmptyRequest{})
			if err != nil {
				return azdext.MCPErrorResult("getting deployment: %s", err), nil
			}

			d := resp.Deployment
			if d == nil {
				return azdext.MCPErrorResult("no deployment found"), nil
			}
			info := map[string]interface{}{
				"id":            d.Id,
				"deployment_id": d.DeploymentId,
				"name":          d.Name,
				"type":          d.Type,
				"location":      d.Location,
				"tags":          d.Tags,
				"outputs":       d.Outputs,
				"resources":     d.Resources,
			}

			return azdext.MCPJSONResult(info), nil
		},
		azdext.MCPToolOptions{
			Description: "Get the latest Azure deployment info including ID, location, outputs, and resources",
			ReadOnly:    true,
		},
	)

	// Tool: get_deployment_context
	builder.AddTool("get_deployment_context",
		func(ctx context.Context, args azdext.ToolArgs) (*mcp.CallToolResult, error) {
			ctx, client, err := newAzdClient(ctx)
			if err != nil {
				return azdext.MCPErrorResult("%s", err), nil
			}
			defer client.Close()

			resp, err := client.Deployment().GetDeploymentContext(ctx, &azdext.EmptyRequest{})
			if err != nil {
				return azdext.MCPErrorResult("getting deployment context: %s", err), nil
			}

			info := map[string]interface{}{}
			if resp.AzureContext != nil {
				info["resources"] = resp.AzureContext.Resources
				if resp.AzureContext.Scope != nil {
					info["subscription_id"] = resp.AzureContext.Scope.SubscriptionId
					info["tenant_id"] = resp.AzureContext.Scope.TenantId
					info["location"] = resp.AzureContext.Scope.Location
					info["resource_group"] = resp.AzureContext.Scope.ResourceGroup
				}
			}

			return azdext.MCPJSONResult(info), nil
		},
		azdext.MCPToolOptions{
			Description: "Get current Azure deployment context including subscription, tenant, location, resource group, and resources",
			ReadOnly:    true,
		},
	)
}

func registerAccountTools(builder *azdext.MCPServerBuilder) {
	// Tool: list_subscriptions
	builder.AddTool("list_subscriptions",
		func(ctx context.Context, args azdext.ToolArgs) (*mcp.CallToolResult, error) {
			ctx, client, err := newAzdClient(ctx)
			if err != nil {
				return azdext.MCPErrorResult("%s", err), nil
			}
			defer client.Close()

			resp, err := client.Account().ListSubscriptions(ctx, &azdext.ListSubscriptionsRequest{})
			if err != nil {
				return azdext.MCPErrorResult("listing subscriptions: %s", err), nil
			}

			type subInfo struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				TenantID  string `json:"tenant_id"`
				IsDefault bool   `json:"is_default"`
			}

			subs := make([]subInfo, 0, len(resp.Subscriptions))
			for _, s := range resp.Subscriptions {
				subs = append(subs, subInfo{
					ID:        s.Id,
					Name:      s.Name,
					TenantID:  s.TenantId,
					IsDefault: s.IsDefault,
				})
			}

			return azdext.MCPJSONResult(subs), nil
		},
		azdext.MCPToolOptions{
			Description: "List Azure subscriptions accessible to the current account",
			ReadOnly:    true,
		},
	)
}

func registerWorkflowTools(builder *azdext.MCPServerBuilder) {
	// Tool: run_workflow
	builder.AddTool("run_workflow",
		func(ctx context.Context, args azdext.ToolArgs) (*mcp.CallToolResult, error) {
			workflowName, err := args.RequireString("workflow_name")
			if err != nil || workflowName == "" {
				return azdext.MCPErrorResult("workflow_name is required"), nil
			}

			raw := args.Raw()
			stepsRaw, ok := raw["steps"].([]interface{})
			if !ok || len(stepsRaw) == 0 {
				return azdext.MCPErrorResult("steps array is required and must not be empty"), nil
			}

			var steps []*azdext.WorkflowStep
			for i, stepRaw := range stepsRaw {
				stepMap, ok := stepRaw.(map[string]interface{})
				if !ok {
					return azdext.MCPErrorResult("step %d is not a valid object", i), nil
				}
				argsRaw, _ := stepMap["args"].([]interface{})
				cmdArgs := make([]string, 0, len(argsRaw))
				for j, a := range argsRaw {
					if a == nil {
						return azdext.MCPErrorResult("step %d arg %d: null values not allowed", i, j), nil
					}
					switch v := a.(type) {
					case string:
						cmdArgs = append(cmdArgs, v)
					default:
						cmdArgs = append(cmdArgs, fmt.Sprint(v))
					}
				}
				if len(cmdArgs) == 0 {
					return azdext.MCPErrorResult("step %d has no command arguments", i), nil
				}
				steps = append(steps, &azdext.WorkflowStep{
					Command: &azdext.WorkflowCommand{Args: cmdArgs},
				})
			}

			if len(steps) == 0 {
				return azdext.MCPErrorResult("no valid workflow steps found"), nil
			}

			ctx, client, err := newAzdClient(ctx)
			if err != nil {
				return azdext.MCPErrorResult("%s", err), nil
			}
			defer client.Close()

			_, err = client.Workflow().Run(ctx, &azdext.RunWorkflowRequest{
				Workflow: &azdext.Workflow{
					Name:  workflowName,
					Steps: steps,
				},
			})
			if err != nil {
				return azdext.MCPErrorResult("running workflow: %s", err), nil
			}

			return mcp.NewToolResultText(fmt.Sprintf("Workflow '%s' completed successfully", workflowName)), nil
		},
		azdext.MCPToolOptions{
			Description: "Execute an azd workflow with the given name and steps",
			Destructive: true,
		},
		mcp.WithString("workflow_name", mcp.Required(), mcp.Description("Name of the workflow to run")),
		mcp.WithArray("steps", mcp.Required(), mcp.Description("Array of step objects, each with an 'args' array of command arguments")),
	)
}

func registerComposeTools(builder *azdext.MCPServerBuilder) {
	// Tool: list_compose_resources
	builder.AddTool("list_compose_resources",
		func(ctx context.Context, args azdext.ToolArgs) (*mcp.CallToolResult, error) {
			ctx, client, err := newAzdClient(ctx)
			if err != nil {
				return azdext.MCPErrorResult("%s", err), nil
			}
			defer client.Close()

			resp, err := client.Compose().ListResources(ctx, &azdext.EmptyRequest{})
			if err != nil {
				return azdext.MCPErrorResult("listing compose resources: %s", err), nil
			}

			type resourceInfo struct {
				Name       string   `json:"name"`
				Type       string   `json:"type"`
				Uses       []string `json:"uses,omitempty"`
				ResourceID string   `json:"resource_id,omitempty"`
			}

			resources := make([]resourceInfo, 0, len(resp.Resources))
			for _, r := range resp.Resources {
				resources = append(resources, resourceInfo{
					Name:       r.Name,
					Type:       r.Type,
					Uses:       r.Uses,
					ResourceID: r.ResourceId,
				})
			}

			return azdext.MCPJSONResult(resources), nil
		},
		azdext.MCPToolOptions{
			Description: "List composability resources defined in the azd project",
			ReadOnly:    true,
		},
	)
}
