// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package commands

import (
	"context"
	"fmt"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/mark3labs/mcp-go/mcp"
)

// Azure AI model and quota grounding tools.
//
// Agents scaffolding an Azure OpenAI application otherwise have to guess which
// model names exist, which regions they are available in, and whether the
// subscription has capacity there. Guessing produces a deploy that fails on a
// quota error, which the agent then has to diagnose from an opaque ARM
// message. These tools let it ask first.
//
// Every call needs the caller's Azure context (subscription and tenant), which
// azd resolves. There is no separate credential to manage here.

// optionalStringSlice reads an array-typed tool argument.
//
// azdext.ToolArgs has accessors for string, int, bool and float but none for
// arrays, so array arguments have to be read off Raw(). A JSON array decodes
// to []any, so that is the case that matters in practice; []string is handled
// as well for callers that build args in Go. Anything else, including a
// non-string element inside the array, is skipped rather than reported,
// because a malformed filter should narrow nothing rather than fail the call.
func optionalStringSlice(args azdext.ToolArgs, key string) []string {
	raw, ok := args.Raw()[key]
	if !ok || raw == nil {
		return nil
	}

	switch v := raw.(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	default:
		return nil
	}
}

// azureContext fetches the caller's Azure scope from azd. Every AI model call
// is subscription-scoped, so this runs first for all of them.
func azureContext(ctx context.Context, client *azdext.AzdClient) (*azdext.AzureContext, error) {
	resp, err := client.Deployment().GetDeploymentContext(ctx, &azdext.EmptyRequest{})
	if err != nil {
		return nil, fmt.Errorf("getting Azure context: %w", err)
	}
	if resp.AzureContext == nil || resp.AzureContext.Scope == nil {
		return nil, fmt.Errorf("no Azure context available; run azd auth login and select a subscription")
	}
	return resp.AzureContext, nil
}

// aiModelSku is the deployable unit of a model version, carrying the capacity
// range the agent needs to pick a deployment size.
type aiModelSku struct {
	Name            string `json:"name"`
	UsageName       string `json:"usage_name"`
	DefaultCapacity int32  `json:"default_capacity"`
	MinCapacity     int32  `json:"min_capacity"`
	MaxCapacity     int32  `json:"max_capacity"`
	CapacityStep    int32  `json:"capacity_step"`
}

type aiModelVersion struct {
	Version         string       `json:"version"`
	IsDefault       bool         `json:"is_default"`
	LifecycleStatus string       `json:"lifecycle_status"`
	Skus            []aiModelSku `json:"skus"`
}

type aiModelInfo struct {
	Name         string           `json:"name"`
	Format       string           `json:"format"`
	Capabilities []string         `json:"capabilities"`
	Locations    []string         `json:"locations"`
	Versions     []aiModelVersion `json:"versions"`
}

func toModelInfo(m *azdext.AiModel) aiModelInfo {
	versions := make([]aiModelVersion, 0, len(m.Versions))
	for _, v := range m.Versions {
		skus := make([]aiModelSku, 0, len(v.Skus))
		for _, s := range v.Skus {
			skus = append(skus, aiModelSku{
				Name:            s.Name,
				UsageName:       s.UsageName,
				DefaultCapacity: s.DefaultCapacity,
				MinCapacity:     s.MinCapacity,
				MaxCapacity:     s.MaxCapacity,
				CapacityStep:    s.CapacityStep,
			})
		}
		versions = append(versions, aiModelVersion{
			Version:         v.Version,
			IsDefault:       v.IsDefault,
			LifecycleStatus: v.LifecycleStatus,
			Skus:            skus,
		})
	}

	return aiModelInfo{
		Name:         m.Name,
		Format:       m.Format,
		Capabilities: m.Capabilities,
		Locations:    m.Locations,
		Versions:     versions,
	}
}

// locationQuota reports how much room a region has left for a model.
type locationQuota struct {
	Location          string  `json:"location"`
	DisplayName       string  `json:"display_name"`
	MaxRemainingQuota float64 `json:"max_remaining_quota"`
}

func toLocationQuotas(from []*azdext.ModelLocationQuota) []locationQuota {
	locations := make([]locationQuota, 0, len(from))
	for _, l := range from {
		if l == nil {
			continue
		}
		entry := locationQuota{MaxRemainingQuota: l.MaxRemainingQuota}
		// Location is a separate message and may be absent; the remaining
		// quota is still worth reporting when it is.
		if l.Location != nil {
			entry.Location = l.Location.Name
			entry.DisplayName = l.Location.DisplayName
		}
		locations = append(locations, entry)
	}

	return locations
}

// usageInfo reports one quota, with the remaining headroom precomputed so the
// agent does not have to subtract.
type usageInfo struct {
	Name      string  `json:"name"`
	Current   float64 `json:"current_value"`
	Limit     float64 `json:"limit"`
	Remaining float64 `json:"remaining"`
}

func toUsageInfos(from []*azdext.AiModelUsage) []usageInfo {
	usages := make([]usageInfo, 0, len(from))
	for _, u := range from {
		if u == nil {
			continue
		}
		usages = append(usages, usageInfo{
			Name:      u.Name,
			Current:   u.CurrentValue,
			Limit:     u.Limit,
			Remaining: u.Limit - u.CurrentValue,
		})
	}

	return usages
}

func registerAiModelTools(builder *azdext.MCPServerBuilder) {
	// Tool: list_ai_models
	builder.AddTool("list_ai_models",
		func(ctx context.Context, args azdext.ToolArgs) (*mcp.CallToolResult, error) {
			ctx, client, err := newAzdClient(ctx)
			if err != nil {
				return azdext.MCPErrorResult("%s", err), nil
			}
			defer client.Close()

			azCtx, err := azureContext(ctx, client)
			if err != nil {
				return azdext.MCPErrorResult("%s", err), nil
			}

			filter := &azdext.AiModelFilterOptions{
				Locations:    optionalStringSlice(args, "locations"),
				Capabilities: optionalStringSlice(args, "capabilities"),
			}

			resp, err := client.Ai().ListModels(ctx, &azdext.ListModelsRequest{
				AzureContext: azCtx,
				Filter:       filter,
			})
			if err != nil {
				return azdext.MCPErrorResult("listing AI models: %s", err), nil
			}

			models := make([]aiModelInfo, 0, len(resp.Models))
			for _, m := range resp.Models {
				models = append(models, toModelInfo(m))
			}

			return azdext.MCPJSONResult(models), nil
		},
		azdext.MCPToolOptions{
			Description: "List the Azure AI models available to the current subscription, with their versions, " +
				"deployment SKUs, capacity ranges and the regions they are offered in. Use this before writing " +
				"infrastructure that names a model, so the name, version and SKU are real rather than assumed.",
			ReadOnly: true,
		},
		mcp.WithArray("locations",
			mcp.Description("Optional Azure regions to restrict results to, for example eastus"),
			mcp.WithStringItems(),
		),
		mcp.WithArray("capabilities",
			mcp.Description("Optional capabilities to filter on, for example chat or embeddings"),
			mcp.WithStringItems(),
		),
	)

	// Tool: find_ai_model_locations_with_quota
	builder.AddTool("find_ai_model_locations_with_quota",
		func(ctx context.Context, args azdext.ToolArgs) (*mcp.CallToolResult, error) {
			modelName, err := args.RequireString("model_name")
			if err != nil || modelName == "" {
				return azdext.MCPErrorResult("model_name is required"), nil
			}

			ctx, client, err := newAzdClient(ctx)
			if err != nil {
				return azdext.MCPErrorResult("%s", err), nil
			}
			defer client.Close()

			azCtx, err := azureContext(ctx, client)
			if err != nil {
				return azdext.MCPErrorResult("%s", err), nil
			}

			request := &azdext.ListModelLocationsWithQuotaRequest{
				AzureContext:     azCtx,
				ModelName:        modelName,
				AllowedLocations: optionalStringSlice(args, "allowed_locations"),
				Quota: &azdext.QuotaCheckOptions{
					MinRemainingCapacity: args.OptionalFloat("min_remaining_capacity", 0),
				},
			}

			resp, err := client.Ai().ListModelLocationsWithQuota(ctx, request)
			if err != nil {
				return azdext.MCPErrorResult("finding locations with quota for %s: %s", modelName, err), nil
			}

			locations := toLocationQuotas(resp.Locations)

			if len(locations) == 0 {
				return azdext.MCPErrorResult(
					"no region has remaining quota for %s under the requested capacity; "+
						"request a quota increase or choose a different model", modelName), nil
			}

			return azdext.MCPJSONResult(locations), nil
		},
		azdext.MCPToolOptions{
			Description: "Find the Azure regions where the current subscription still has quota to deploy a " +
				"named AI model, with the remaining capacity in each. Use this to pick a region before " +
				"provisioning, instead of deploying and reading a quota failure afterwards.",
			ReadOnly: true,
		},
		mcp.WithString("model_name", mcp.Required(),
			mcp.Description("Model to check, for example gpt-4o")),
		mcp.WithArray("allowed_locations",
			mcp.Description("Optional Azure regions to restrict the search to"),
			mcp.WithStringItems(),
		),
		mcp.WithNumber("min_remaining_capacity",
			mcp.Description("Minimum remaining capacity a region must have. Zero means any remaining quota.")),
	)

	// Tool: list_ai_quota_usage
	builder.AddTool("list_ai_quota_usage",
		func(ctx context.Context, args azdext.ToolArgs) (*mcp.CallToolResult, error) {
			location, err := args.RequireString("location")
			if err != nil || location == "" {
				return azdext.MCPErrorResult("location is required"), nil
			}

			ctx, client, err := newAzdClient(ctx)
			if err != nil {
				return azdext.MCPErrorResult("%s", err), nil
			}
			defer client.Close()

			azCtx, err := azureContext(ctx, client)
			if err != nil {
				return azdext.MCPErrorResult("%s", err), nil
			}

			resp, err := client.Ai().ListUsages(ctx, &azdext.ListUsagesRequest{
				AzureContext: azCtx,
				Location:     location,
			})
			if err != nil {
				return azdext.MCPErrorResult("listing AI quota usage in %s: %s", location, err), nil
			}

			return azdext.MCPJSONResult(toUsageInfos(resp.Usages)), nil
		},
		azdext.MCPToolOptions{
			Description: "List Azure AI quota usage in a region, showing the current value, limit and " +
				"remaining capacity for each quota. Use this to explain why a deployment was rejected " +
				"for capacity.",
			ReadOnly: true,
		},
		mcp.WithString("location", mcp.Required(),
			mcp.Description("Azure region to report on, for example eastus")),
	)
}
