// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/mark3labs/mcp-go/mcp"
)

// The AI model tools are the only grounding this extension gives an agent for
// Azure OpenAI model names, versions, SKUs and regional quota. Everything they
// return comes from azd over gRPC, so these tests do not exercise the network.
// They cover the pure translation layer and the registration wiring, which is
// where a regression would silently drop a tool from the agent's toolset.

func TestOptionalStringSliceReadsJSONArray(t *testing.T) {
	// A JSON array arrives as []any, which is the case that actually happens
	// when an agent calls the tool.
	args := toolArgs(map[string]any{"locations": []any{"eastus", "westus2"}})

	got := optionalStringSlice(args, "locations")

	if len(got) != 2 || got[0] != "eastus" || got[1] != "westus2" {
		t.Fatalf("got %#v, want [eastus westus2]", got)
	}
}

func TestOptionalStringSliceHandlesGoNativeSlice(t *testing.T) {
	args := toolArgs(map[string]any{"locations": []string{"eastus"}})

	got := optionalStringSlice(args, "locations")

	if len(got) != 1 || got[0] != "eastus" {
		t.Fatalf("got %#v, want [eastus]", got)
	}
}

func TestOptionalStringSliceIgnoresUnusableValues(t *testing.T) {
	cases := map[string]map[string]any{
		"missing key":     {},
		"nil value":       {"locations": nil},
		"wrong type":      {"locations": "eastus"},
		"empty array":     {"locations": []any{}},
		"non-string item": {"locations": []any{42, true}},
		"empty strings":   {"locations": []any{"", ""}},
	}

	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			if got := optionalStringSlice(toolArgs(args), "locations"); got != nil {
				t.Fatalf("got %#v, want nil so the filter narrows nothing", got)
			}
		})
	}
}

func TestOptionalStringSliceKeepsUsableItemsFromMixedArray(t *testing.T) {
	args := toolArgs(map[string]any{"locations": []any{"eastus", 42, "", "westus2"}})

	got := optionalStringSlice(args, "locations")

	if len(got) != 2 || got[0] != "eastus" || got[1] != "westus2" {
		t.Fatalf("got %#v, want the two usable entries", got)
	}
}

// TestToolArgsStillLacksArrayAccessor is the trigger to delete
// optionalStringSlice. It exists only because azdext.ToolArgs has no array
// accessor; if the SDK grows one, this local helper is duplicate surface and
// should go.
func TestToolArgsStillLacksArrayAccessor(t *testing.T) {
	var args azdext.ToolArgs

	if _, ok := any(args).(interface{ GetStringSlice(string) []string }); ok {
		t.Error("azdext.ToolArgs now has GetStringSlice. Replace optionalStringSlice with it and " +
			"delete this test as part of that change")
	}
	if _, ok := any(args).(interface {
		OptionalStringSlice(string, []string) []string
	}); ok {
		t.Error("azdext.ToolArgs now has OptionalStringSlice. Replace optionalStringSlice with it and " +
			"delete this test as part of that change")
	}
}

func TestToModelInfoCarriesEverySkuCapacityField(t *testing.T) {
	model := &azdext.AiModel{
		Name:         "gpt-4o",
		Format:       "OpenAI",
		Capabilities: []string{"chat"},
		Locations:    []string{"eastus"},
		Versions: []*azdext.AiModelVersion{{
			Version:         "2024-08-06",
			IsDefault:       true,
			LifecycleStatus: "GenerallyAvailable",
			Skus: []*azdext.AiModelSku{{
				Name:            "GlobalStandard",
				UsageName:       "OpenAI.GlobalStandard.gpt-4o",
				DefaultCapacity: 30,
				MinCapacity:     1,
				MaxCapacity:     450,
				CapacityStep:    1,
			}},
		}},
	}

	got := toModelInfo(model)

	if got.Name != "gpt-4o" || got.Format != "OpenAI" {
		t.Fatalf("model identity not carried: %+v", got)
	}
	if len(got.Versions) != 1 || len(got.Versions[0].Skus) != 1 {
		t.Fatalf("version or sku dropped: %+v", got)
	}

	sku := got.Versions[0].Skus[0]
	// The capacity range is the reason this tool exists: an agent picking a
	// deployment size needs all four bounds, not just the default.
	if sku.DefaultCapacity != 30 || sku.MinCapacity != 1 || sku.MaxCapacity != 450 || sku.CapacityStep != 1 {
		t.Fatalf("capacity range not carried: %+v", sku)
	}
	if sku.UsageName != "OpenAI.GlobalStandard.gpt-4o" {
		t.Fatalf("usage name is the join key to the quota API and must survive: %q", sku.UsageName)
	}
	if !got.Versions[0].IsDefault || got.Versions[0].LifecycleStatus != "GenerallyAvailable" {
		t.Fatalf("version metadata not carried: %+v", got.Versions[0])
	}
}

func TestToModelInfoHandlesEmptyCollections(t *testing.T) {
	got := toModelInfo(&azdext.AiModel{Name: "gpt-4o"})

	if got.Versions == nil {
		t.Error("versions should serialize as an empty array, not null")
	}

	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(encoded), `"versions":[]`) {
		t.Errorf("want an empty array so the agent sees no versions rather than a null: %s", encoded)
	}
}

func TestToLocationQuotasCarriesRegionAndHeadroom(t *testing.T) {
	got := toLocationQuotas([]*azdext.ModelLocationQuota{
		{Location: &azdext.Location{Name: "eastus", DisplayName: "East US"}, MaxRemainingQuota: 150},
		{Location: &azdext.Location{Name: "westus2", DisplayName: "West US 2"}, MaxRemainingQuota: 0},
	})

	if len(got) != 2 {
		t.Fatalf("got %d entries, want 2", len(got))
	}
	if got[0].Location != "eastus" || got[0].DisplayName != "East US" || got[0].MaxRemainingQuota != 150 {
		t.Errorf("first entry not carried: %+v", got[0])
	}
	// A region reporting zero headroom is a real answer and must survive, so
	// the agent can see the region exists but is full.
	if got[1].Location != "westus2" || got[1].MaxRemainingQuota != 0 {
		t.Errorf("zero-quota region not carried: %+v", got[1])
	}
}

func TestToLocationQuotasToleratesMissingLocation(t *testing.T) {
	got := toLocationQuotas([]*azdext.ModelLocationQuota{
		nil,
		{MaxRemainingQuota: 42},
	})

	if len(got) != 1 {
		t.Fatalf("got %d entries, want the nil entry skipped and the other kept", len(got))
	}
	if got[0].Location != "" || got[0].MaxRemainingQuota != 42 {
		t.Errorf("quota should survive an absent Location message: %+v", got[0])
	}
}

func TestToLocationQuotasReturnsEmptySliceNotNil(t *testing.T) {
	encoded, err := json.Marshal(toLocationQuotas(nil))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(encoded) != "[]" {
		t.Errorf("got %s, want [] so the agent sees no regions rather than a null", encoded)
	}
}

func TestToUsageInfosComputesRemaining(t *testing.T) {
	got := toUsageInfos([]*azdext.AiModelUsage{
		{Name: "OpenAI.GlobalStandard.gpt-4o", CurrentValue: 30, Limit: 450},
		{Name: "OpenAI.Standard.gpt-4o", CurrentValue: 100, Limit: 100},
	})

	if len(got) != 2 {
		t.Fatalf("got %d entries, want 2", len(got))
	}
	// Precomputing the headroom is the whole point: an agent reading current
	// and limit separately has to do arithmetic it often gets wrong.
	if got[0].Remaining != 420 {
		t.Errorf("got remaining %v, want 420", got[0].Remaining)
	}
	if got[1].Remaining != 0 {
		t.Errorf("an exhausted quota should report 0 remaining, got %v", got[1].Remaining)
	}
	if got[0].Name != "OpenAI.GlobalStandard.gpt-4o" {
		t.Errorf("quota name not carried: %q", got[0].Name)
	}
}

func TestToUsageInfosSkipsNilEntries(t *testing.T) {
	got := toUsageInfos([]*azdext.AiModelUsage{nil, {Name: "quota", Limit: 10}})

	if len(got) != 1 || got[0].Name != "quota" {
		t.Fatalf("got %#v, want the nil skipped and the real entry kept", got)
	}
}

// aiToolNames are the tools this file contributes to the agent's toolset.
var aiToolNames = []string{
	"list_ai_models",
	"find_ai_model_locations_with_quota",
	"list_ai_quota_usage",
}

// TestAiToolsAreRegistered catches the failure mode where a tool is written but
// never wired into registerGRPCTools, which produces no compile error and no
// runtime error. The tool is simply absent from the agent's toolset.
func TestAiToolsAreRegistered(t *testing.T) {
	registered := toolNamesRegisteredIn(t, "mcp_ai_tools.go")

	for _, name := range aiToolNames {
		if !registered[name] {
			t.Errorf("tool %q is not registered by mcp_ai_tools.go", name)
		}
	}

	for name := range registered {
		var known bool
		for _, expected := range aiToolNames {
			if name == expected {
				known = true
			}
		}
		if !known {
			t.Errorf("tool %q was added to mcp_ai_tools.go but not to aiToolNames. Add it, and confirm "+
				"registerAiModelTools is still reached from registerGRPCTools", name)
		}
	}
}

// TestRegisterAiModelToolsIsReachable pins the wiring itself. Without this,
// every tool could be correctly declared and still never reach an agent.
func TestRegisterAiModelToolsIsReachable(t *testing.T) {
	source := readSourceFile(t, "mcp_grpc_tools.go")

	if !strings.Contains(source, "registerAiModelTools(builder)") {
		t.Fatal("registerGRPCTools no longer calls registerAiModelTools, so the AI grounding tools " +
			"are unreachable from the MCP server")
	}
}

// TestAiToolsAreReadOnly guards the safety posture. Every one of these tools
// only reads Azure metadata, so none may be marked destructive, which would
// make an agent prompt for approval it does not need.
func TestAiToolsAreReadOnly(t *testing.T) {
	source := readSourceFile(t, "mcp_ai_tools.go")

	if strings.Contains(source, "Destructive:") {
		t.Error("an AI model tool was marked Destructive. These tools only read Azure metadata; " +
			"if one now writes, that is a change worth reviewing on its own")
	}
	if got := len(readOnlyPattern.FindAllString(source, -1)); got != len(aiToolNames) {
		t.Errorf("got %d ReadOnly declarations, want %d, one per tool", got, len(aiToolNames))
	}
}

var readOnlyPattern = regexp.MustCompile(`ReadOnly:\s*true`)

// TestAiToolDescriptionsExplainWhenToUse keeps the descriptions useful. A tool
// an agent cannot tell when to reach for is a tool it will not call.
func TestAiToolDescriptionsExplainWhenToUse(t *testing.T) {
	source := readSourceFile(t, "mcp_ai_tools.go")

	if got := strings.Count(source, "Use this"); got < len(aiToolNames) {
		t.Errorf("got %d descriptions telling the agent when to use the tool, want at least %d",
			got, len(aiToolNames))
	}
}

var addToolPattern = regexp.MustCompile(`builder\.AddTool\("([a-z_]+)"`)

func toolNamesRegisteredIn(t *testing.T, filename string) map[string]bool {
	t.Helper()

	names := map[string]bool{}
	for _, match := range addToolPattern.FindAllStringSubmatch(readSourceFile(t, filename), -1) {
		names[match[1]] = true
	}

	if len(names) == 0 {
		t.Fatalf("found no builder.AddTool calls in %s; the scan pattern is stale", filename)
	}

	return names
}

func readSourceFile(t *testing.T, filename string) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Clean(filename))
	if err != nil {
		t.Fatalf("reading %s: %v", filename, err)
	}

	return string(content)
}

// toolArgs builds an azdext.ToolArgs for a test.
//
// azdext.ToolArgs holds an unexported map, so outside the SDK the only way to
// construct one carrying data is ParseToolArgs on an MCP request. That is
// worth knowing: extension tests cannot build tool arguments directly.
func toolArgs(raw map[string]any) azdext.ToolArgs {
	request := mcp.CallToolRequest{}
	request.Params.Arguments = raw
	return azdext.ParseToolArgs(request)
}
