// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package copilot

import (
	"reflect"
	"testing"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
)

// This file records why azd copilot keeps launching the GitHub Copilot CLI as
// a subprocess instead of routing non-interactive work through
// azdext.AzdClient.Copilot(), and pins the specific facts that decision rests
// on so a future SDK release that closes the gap fails the build here.
//
// The evaluation did not get as far as measuring latency, because latency only
// matters between two paths that can do the same job. The gRPC surface cannot
// express the inputs this extension is built around.
//
// SendCopilotMessageRequest carries exactly: prompt, session_id, model,
// reasoning_effort, system_message, mode, debug, headless. The complete field
// set across the whole copilot proto adds only working_directory (on session
// listing), plus response-side usage and file-change reporting. That leaves
// four hard blockers:
//
//  1. No agent selection. Every launch this extension performs names an agent
//     (--agent, defaulting to azure-manager) and the repository ships 16 of
//     them. There is no proto field to name one.
//
//  2. No custom instructions directory. The extension ships over 150 skill
//     directories and points the CLI at them through
//     COPILOT_CUSTOM_INSTRUCTIONS_DIRS and --add-dir. Skills are loaded on
//     demand by name, which is the entire reason the catalog can be that
//     large. There is no proto field for either, and folding the catalog into
//     system_message would send everything on every turn and still lose the
//     skill invocation tool.
//
//  3. No environment injection. The extension passes project name, service
//     inventory, subscription, tenant, user, infra layout and every AZURE_*
//     value from the azd environment as process environment. The agents read
//     those. gRPC has no equivalent, so the agent would run blind.
//
//  4. No response text. SendMessage returns session_id, usage and file
//     changes, and nothing else. Reading what the agent actually said needs a
//     second GetMessages call, and the unary call produces no incremental
//     output at all, so a long autopilot run would be silent until it
//     finished.
//
// Task 3.9, which proposed routing --prompt, autopilot and MCP tool paths
// through gRPC if this evaluation was favorable, does not apply.

// launcherOptionsRequiringSubprocess lists the Options fields that have no
// representation in the gRPC surface. Each one on its own is enough to rule
// out a swap.
var launcherOptionsRequiringSubprocess = []string{
	"Agent",
	"AddDirs",
	"Yolo",
	"Resume",
	"Continue",
	"ProjectContext",
}

// TestLauncherOptionsThatBlockGrpcStillExist keeps the list above honest. If a
// field is renamed or removed, the rationale referring to it is stale and must
// be rewritten rather than left to describe code that no longer exists.
func TestLauncherOptionsThatBlockGrpcStillExist(t *testing.T) {
	optionsType := reflect.TypeOf(Options{})

	for _, name := range launcherOptionsRequiringSubprocess {
		if _, ok := optionsType.FieldByName(name); !ok {
			t.Errorf("Options.%s no longer exists. It is cited in the rationale for keeping the "+
				"subprocess launcher; update that rationale in the same change", name)
		}
	}
}

// TestCopilotRequestCannotExpressLauncherOptions is the trigger to
// re-evaluate. It asserts the SDK request type still lacks the fields the
// launcher needs. The day azd adds an agent or directories field this fails,
// which is the signal that routing non-interactive paths through gRPC became
// possible.
func TestCopilotRequestCannotExpressLauncherOptions(t *testing.T) {
	requestType := reflect.TypeOf(azdext.SendCopilotMessageRequest{})

	// Names azd would plausibly choose for each blocking capability. Matching
	// on several spellings keeps the check from being defeated by a rename.
	missing := map[string][]string{
		"agent selection":               {"Agent", "AgentName", "AgentId"},
		"custom instructions directory": {"AddDirs", "Directories", "CustomInstructionsDirs", "InstructionDirs"},
		"environment injection":         {"Env", "Environment", "EnvVars"},
		"working directory":             {"WorkingDirectory", "Cwd"},
		"response text":                 {"Response", "Content", "Text", "Message"},
	}

	for capability, candidates := range missing {
		for _, name := range candidates {
			if _, ok := requestType.FieldByName(name); ok {
				t.Errorf("azdext.SendCopilotMessageRequest now has %q, which covers %s. "+
					"Re-evaluate routing non-interactive paths through client.Copilot() and "+
					"delete this test as part of that change", name, capability)
			}
		}
	}
}

// TestCopilotResponseStillOmitsAssistantText pins the response-side half. The
// unary call returning no assistant text is a separate blocker from the
// request-side gaps and would need to be fixed independently.
func TestCopilotResponseStillOmitsAssistantText(t *testing.T) {
	responseType := reflect.TypeOf(azdext.SendCopilotMessageResponse{})

	for _, name := range []string{"Response", "Content", "Text", "Message", "Output"} {
		if _, ok := responseType.FieldByName(name); ok {
			t.Errorf("azdext.SendCopilotMessageResponse now has %q. Reading the agent's reply no "+
				"longer needs a second GetMessages call, so re-evaluate the gRPC path", name)
		}
	}
}

// TestDefaultAgentIsNamed pins the concrete fact behind blocker 1: every
// launch names an agent, so an API with no agent field cannot serve any of
// them.
func TestDefaultAgentIsNamed(t *testing.T) {
	args := buildArgs(Options{})

	var found bool
	for i, a := range args {
		if a == "--agent" && i+1 < len(args) {
			found = true
			if args[i+1] == "" {
				t.Error("--agent was passed with an empty value")
			}
		}
	}

	if !found {
		t.Error("buildArgs no longer passes --agent. The gRPC evaluation rests on every launch " +
			"naming an agent; if that is no longer true, redo the evaluation")
	}
}
