// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package commands

import (
"context"
"encoding/json"
"fmt"

"github.com/azure/azure-dev/cli/azd/pkg/azdext"
"github.com/jongio/azd-core/azdextutil"
"github.com/mark3labs/mcp-go/mcp"
"github.com/mark3labs/mcp-go/server"
)

var grpcRateLimiter = azdextutil.NewRateLimiter(10, 1.0)

// registerGRPCTools registers MCP tools that wrap azd gRPC services.
func registerGRPCTools(s *server.MCPServer) {
registerEnvironmentTools(s)
registerDeploymentTools(s)
registerAccountTools(s)
registerWorkflowTools(s)
registerComposeTools(s)
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

func registerEnvironmentTools(s *server.MCPServer) {
// Tool: list_environments
s.AddTool(
mcp.NewTool("list_environments",
mcp.WithDescription("List all azd environments"),
mcp.WithReadOnlyHintAnnotation(true),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
if err := grpcRateLimiter.CheckRateLimit("list_environments"); err != nil {
return mcp.NewToolResultError(err.Error()), nil
}

ctx, client, err := newAzdClient(ctx)
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
defer client.Close()

resp, err := client.Environment().List(ctx, &azdext.EmptyRequest{})
if err != nil {
return mcp.NewToolResultError(fmt.Sprintf("listing environments: %s", err)), nil
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

data, err := json.MarshalIndent(envs, "", "  ")
if err != nil {
return mcp.NewToolResultError(fmt.Sprintf("marshaling response: %s", err)), nil
}
return mcp.NewToolResultText(string(data)), nil
},
)

// Tool: get_environment_values
s.AddTool(
mcp.NewTool("get_environment_values",
mcp.WithDescription("Get all key-value pairs for an azd environment"),
mcp.WithString("environment_name", mcp.Required(), mcp.Description("Name of the environment")),
mcp.WithReadOnlyHintAnnotation(true),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
if err := grpcRateLimiter.CheckRateLimit("get_environment_values"); err != nil {
return mcp.NewToolResultError(err.Error()), nil
}

args, ok := req.Params.Arguments.(map[string]interface{})
if !ok {
return mcp.NewToolResultError("invalid arguments"), nil
}
name, _ := args["environment_name"].(string)
if name == "" {
return mcp.NewToolResultError("environment_name is required"), nil
}

ctx, client, err := newAzdClient(ctx)
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
defer client.Close()

resp, err := client.Environment().GetValues(ctx, &azdext.GetEnvironmentRequest{Name: name})
if err != nil {
return mcp.NewToolResultError(fmt.Sprintf("getting environment values: %s", err)), nil
}

kvMap := make(map[string]string, len(resp.KeyValues))
for _, kv := range resp.KeyValues {
kvMap[kv.Key] = kv.Value
}

data, err := json.MarshalIndent(kvMap, "", "  ")
if err != nil {
return mcp.NewToolResultError(fmt.Sprintf("marshaling response: %s", err)), nil
}
return mcp.NewToolResultText(string(data)), nil
},
)

// Tool: set_environment_value
s.AddTool(
mcp.NewTool("set_environment_value",
mcp.WithDescription("Set a key-value pair in an azd environment"),
mcp.WithString("environment_name", mcp.Required(), mcp.Description("Name of the environment")),
mcp.WithString("key", mcp.Required(), mcp.Description("Key to set")),
mcp.WithString("value", mcp.Required(), mcp.Description("Value to set")),
mcp.WithDestructiveHintAnnotation(true),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
if err := grpcRateLimiter.CheckRateLimit("set_environment_value"); err != nil {
return mcp.NewToolResultError(err.Error()), nil
}

args, ok := req.Params.Arguments.(map[string]interface{})
if !ok {
return mcp.NewToolResultError("invalid arguments"), nil
}
envName, _ := args["environment_name"].(string)
key, _ := args["key"].(string)
value, _ := args["value"].(string)
if envName == "" || key == "" {
return mcp.NewToolResultError("environment_name and key are required"), nil
}

ctx, client, err := newAzdClient(ctx)
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
defer client.Close()

_, err = client.Environment().SetValue(ctx, &azdext.SetEnvRequest{
EnvName: envName,
Key:     key,
Value:   value,
})
if err != nil {
return mcp.NewToolResultError(fmt.Sprintf("setting environment value: %s", err)), nil
}

return mcp.NewToolResultText(fmt.Sprintf("Successfully set %s in environment %s", key, envName)), nil
},
)
}

func registerDeploymentTools(s *server.MCPServer) {
// Tool: get_deployment_info
s.AddTool(
mcp.NewTool("get_deployment_info",
mcp.WithDescription("Get the latest Azure deployment info including ID, location, outputs, and resources"),
mcp.WithReadOnlyHintAnnotation(true),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
if err := grpcRateLimiter.CheckRateLimit("get_deployment_info"); err != nil {
return mcp.NewToolResultError(err.Error()), nil
}

ctx, client, err := newAzdClient(ctx)
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
defer client.Close()

resp, err := client.Deployment().GetDeployment(ctx, &azdext.EmptyRequest{})
if err != nil {
return mcp.NewToolResultError(fmt.Sprintf("getting deployment: %s", err)), nil
}

d := resp.Deployment
if d == nil {
return mcp.NewToolResultError("no deployment found"), nil
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

data, err := json.MarshalIndent(info, "", "  ")
if err != nil {
return mcp.NewToolResultError(fmt.Sprintf("marshaling response: %s", err)), nil
}
return mcp.NewToolResultText(string(data)), nil
},
)

// Tool: get_deployment_context
s.AddTool(
mcp.NewTool("get_deployment_context",
mcp.WithDescription("Get current Azure deployment context including subscription, tenant, location, resource group, and resources"),
mcp.WithReadOnlyHintAnnotation(true),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
if err := grpcRateLimiter.CheckRateLimit("get_deployment_context"); err != nil {
return mcp.NewToolResultError(err.Error()), nil
}

ctx, client, err := newAzdClient(ctx)
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
defer client.Close()

resp, err := client.Deployment().GetDeploymentContext(ctx, &azdext.EmptyRequest{})
if err != nil {
return mcp.NewToolResultError(fmt.Sprintf("getting deployment context: %s", err)), nil
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

data, err := json.MarshalIndent(info, "", "  ")
if err != nil {
return mcp.NewToolResultError(fmt.Sprintf("marshaling response: %s", err)), nil
}
return mcp.NewToolResultText(string(data)), nil
},
)
}

func registerAccountTools(s *server.MCPServer) {
// Tool: list_subscriptions
s.AddTool(
mcp.NewTool("list_subscriptions",
mcp.WithDescription("List Azure subscriptions accessible to the current account"),
mcp.WithReadOnlyHintAnnotation(true),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
if err := grpcRateLimiter.CheckRateLimit("list_subscriptions"); err != nil {
return mcp.NewToolResultError(err.Error()), nil
}

ctx, client, err := newAzdClient(ctx)
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
defer client.Close()

resp, err := client.Account().ListSubscriptions(ctx, &azdext.ListSubscriptionsRequest{})
if err != nil {
return mcp.NewToolResultError(fmt.Sprintf("listing subscriptions: %s", err)), nil
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

data, err := json.MarshalIndent(subs, "", "  ")
if err != nil {
return mcp.NewToolResultError(fmt.Sprintf("marshaling response: %s", err)), nil
}
return mcp.NewToolResultText(string(data)), nil
},
)
}

func registerWorkflowTools(s *server.MCPServer) {
// Tool: run_workflow
s.AddTool(
mcp.NewTool("run_workflow",
mcp.WithDescription("Execute an azd workflow with the given name and steps"),
mcp.WithString("workflow_name", mcp.Required(), mcp.Description("Name of the workflow to run")),
mcp.WithArray("steps", mcp.Required(), mcp.Description("Array of step objects, each with an 'args' array of command arguments")),
mcp.WithDestructiveHintAnnotation(true),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
if err := grpcRateLimiter.CheckRateLimit("run_workflow"); err != nil {
return mcp.NewToolResultError(err.Error()), nil
}

args, ok := req.Params.Arguments.(map[string]interface{})
if !ok {
return mcp.NewToolResultError("invalid arguments"), nil
}
workflowName, _ := args["workflow_name"].(string)
if workflowName == "" {
return mcp.NewToolResultError("workflow_name is required"), nil
}

stepsRaw, ok := args["steps"].([]interface{})
if !ok || len(stepsRaw) == 0 {
return mcp.NewToolResultError("steps array is required and must not be empty"), nil
}

var steps []*azdext.WorkflowStep
for i, stepRaw := range stepsRaw {
stepMap, ok := stepRaw.(map[string]interface{})
if !ok {
return mcp.NewToolResultError(fmt.Sprintf("step %d is not a valid object", i)), nil
}
argsRaw, _ := stepMap["args"].([]interface{})
cmdArgs := make([]string, 0, len(argsRaw))
for j, a := range argsRaw {
if a == nil {
return mcp.NewToolResultError(fmt.Sprintf("step %d arg %d: null values not allowed", i, j)), nil
}
switch v := a.(type) {
case string:
cmdArgs = append(cmdArgs, v)
default:
cmdArgs = append(cmdArgs, fmt.Sprint(v))
}
}
if len(cmdArgs) == 0 {
return mcp.NewToolResultError(fmt.Sprintf("step %d has no command arguments", i)), nil
}
steps = append(steps, &azdext.WorkflowStep{
Command: &azdext.WorkflowCommand{Args: cmdArgs},
})
}

if len(steps) == 0 {
return mcp.NewToolResultError("no valid workflow steps found"), nil
}

ctx, client, err := newAzdClient(ctx)
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
defer client.Close()

_, err = client.Workflow().Run(ctx, &azdext.RunWorkflowRequest{
Workflow: &azdext.Workflow{
Name:  workflowName,
Steps: steps,
},
})
if err != nil {
return mcp.NewToolResultError(fmt.Sprintf("running workflow: %s", err)), nil
}

return mcp.NewToolResultText(fmt.Sprintf("Workflow '%s' completed successfully", workflowName)), nil
},
)
}

func registerComposeTools(s *server.MCPServer) {
// Tool: list_compose_resources
s.AddTool(
mcp.NewTool("list_compose_resources",
mcp.WithDescription("List composability resources defined in the azd project"),
mcp.WithReadOnlyHintAnnotation(true),
),
func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
if err := grpcRateLimiter.CheckRateLimit("list_compose_resources"); err != nil {
return mcp.NewToolResultError(err.Error()), nil
}

ctx, client, err := newAzdClient(ctx)
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
defer client.Close()

resp, err := client.Compose().ListResources(ctx, &azdext.EmptyRequest{})
if err != nil {
return mcp.NewToolResultError(fmt.Sprintf("listing compose resources: %s", err)), nil
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

data, err := json.MarshalIndent(resources, "", "  ")
if err != nil {
return mcp.NewToolResultError(fmt.Sprintf("marshaling response: %s", err)), nil
}
return mcp.NewToolResultText(string(data)), nil
},
)
}
