// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package commands

import (
	"strings"
	"testing"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
)

func TestValidateWorkflowStepArgsAcceptsLifecycleCommands(t *testing.T) {
	// Every one of these is a legitimate workflow step. The guard exists to
	// narrow the surface, not to make the tool unusable, so a regression that
	// blocks these is as bad as one that blocks nothing.
	for _, args := range [][]string{
		{"up"},
		{"provision"},
		{"deploy", "--service", "api"},
		{"package"},
		{"restore"},
		{"down", "--force", "--purge"},
		{"env", "set", "KEY", "value"},
		{"hooks", "run", "predeploy"},
		{"init", "--template", "todo-nodejs-mongo"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			if err := validateWorkflowStepArgs(0, args); err != nil {
				t.Errorf("rejected a legitimate workflow step: %v", err)
			}
		})
	}
}

func TestValidateWorkflowStepArgsRejectsCredentialAndConfigCommands(t *testing.T) {
	cases := map[string][]string{
		// azd auth token is hidden but reachable, and prints an access token.
		"auth token":  {"auth", "token"},
		"auth logout": {"auth", "logout"},
		"auth login":  {"auth", "login"},
		// Case is not a defense.
		"mixed case":   {"AuTh", "token"},
		"config set":   {"config", "set", "defaults.subscription", "..."},
		"config unset": {"config", "unset", "defaults.location"},
	}

	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			err := validateWorkflowStepArgs(0, args)
			if err == nil {
				t.Fatalf("accepted %q, which is denied for workflow steps", strings.Join(args, " "))
			}
			// The message has to say which step and why, because it is read by
			// an agent deciding what to do next.
			if !strings.Contains(err.Error(), "step 0") {
				t.Errorf("error should name the step: %v", err)
			}
			if !strings.Contains(err.Error(), "lifecycle") {
				t.Errorf("error should say what a workflow step is for: %v", err)
			}
		})
	}
}

func TestValidateWorkflowStepArgsRejectsLeadingFlag(t *testing.T) {
	// A flag where the command name belongs is how an injected argument
	// reaches a command the caller never named (CWE-88).
	for _, args := range [][]string{
		{"--cwd", "/etc"},
		{"-C", "/etc"},
		{"--help"},
	} {
		t.Run(args[0], func(t *testing.T) {
			err := validateWorkflowStepArgs(2, args)
			if err == nil {
				t.Fatalf("accepted %q as a step command", args[0])
			}
			if !strings.Contains(err.Error(), "flag") {
				t.Errorf("error should explain that a flag is not a command: %v", err)
			}
		})
	}
}

func TestValidateWorkflowStepArgsRejectsEmptyCommands(t *testing.T) {
	cases := map[string][]string{
		"no args":    {},
		"empty":      {""},
		"whitespace": {"   "},
		"tab":        {"\t"},
	}

	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			if err := validateWorkflowStepArgs(1, args); err == nil {
				t.Fatalf("accepted %#v as a step command", args)
			}
		})
	}
}

func TestDeniedWorkflowCommandsEachHaveAReason(t *testing.T) {
	if len(deniedWorkflowCommands) == 0 {
		t.Fatal("the denylist is empty, so the guard permits credential and config commands")
	}

	for command, reason := range deniedWorkflowCommands {
		if command != strings.ToLower(command) {
			t.Errorf("denylist key %q must be lowercase; lookups lowercase the input", command)
		}
		if strings.TrimSpace(reason) == "" {
			t.Errorf("denied command %q has no reason, so its error message will not explain itself", command)
		}
	}
}

// TestWorkflowGuardIsWired catches the case where the guard exists and is
// tested but never called, which produces no compile error and leaves the tool
// exactly as unprotected as before.
func TestWorkflowGuardIsWired(t *testing.T) {
	source := readSourceFile(t, "mcp_grpc_tools.go")

	if !strings.Contains(source, "validateWorkflowStepArgs(i, cmdArgs)") {
		t.Fatal("run_workflow no longer calls validateWorkflowStepArgs, so any azd command is " +
			"reachable through a tool advertised as running a project workflow")
	}
}

// TestRunWorkflowDescriptionMatchesBehavior keeps the advertised contract
// honest. The original description said the tool runs "an azd workflow", which
// reads as a project-defined sequence and understates what an approver is
// agreeing to.
func TestRunWorkflowDescriptionMatchesBehavior(t *testing.T) {
	source := readSourceFile(t, "mcp_grpc_tools.go")

	if !strings.Contains(source, "azd project lifecycle commands") {
		t.Error("the run_workflow description should say it runs azd lifecycle commands, so whoever " +
			"approves the call knows what they are approving")
	}
	if !strings.Contains(source, "Credential and global configuration commands are rejected") {
		t.Error("the run_workflow description should state the guard, so an agent does not retry a " +
			"denied command")
	}
}

// TestNoMcpToolTakesAPathOrUrl records why azdext.WithSecurityPolicy was not
// adopted. Its path check and its SSRF guard both need a parameter to act on,
// and this extension exposes neither. If a tool ever takes one, this fails and
// the policy is worth re-evaluating.
func TestNoMcpToolTakesAPathOrUrl(t *testing.T) {
	// Parameter names that would mean a filesystem path or an outbound
	// request, which are the two things the SDK policy guards.
	suspicious := []string{
		"\"path\"", "\"file\"", "\"file_path\"", "\"directory\"", "\"dir\"",
		"\"url\"", "\"uri\"", "\"endpoint\"", "\"webhook\"",
	}

	for _, filename := range []string{"mcp_ai_tools.go", "mcp_grpc_tools.go", "mcp.go"} {
		source := readSourceFile(t, filename)
		for _, name := range suspicious {
			needle := "mcp.WithString(" + name
			if strings.Contains(source, needle) {
				t.Errorf("%s declares a tool parameter %s. azdext.WithSecurityPolicy was rejected "+
					"because no tool took a path or a URL; re-evaluate it now that one does",
					filename, name)
			}
		}
	}
}

// TestSecurityPolicyIsNotAdopted pins the decision itself. Delete this test
// rather than editing it, and only after reading the rationale at the top of
// mcp_workflow_guard.go.
func TestSecurityPolicyIsNotAdopted(t *testing.T) {
	for _, filename := range []string{"mcp.go", "mcp_grpc_tools.go", "mcp_ai_tools.go"} {
		if strings.Contains(readSourceFile(t, filename), "WithSecurityPolicy") {
			t.Errorf("%s now calls azdext.WithSecurityPolicy. The builder stores that policy and "+
				"enforces nothing, and both of its checks are inert on this tool surface. If it now "+
				"enforces, delete this test as part of that change", filename)
		}
	}
}

// TestWorkflowGuardDoesNotDependOnAzdValidation records that azd applies no
// allowlist of its own, which is why this guard has to exist at all.
func TestWorkflowGuardDoesNotDependOnAzdValidation(t *testing.T) {
	// The workflow request type carries raw args and nothing else. There is no
	// field through which azd could be told to restrict them, which is the
	// concrete reason the check belongs on this side of the wire.
	step := &azdext.WorkflowStep{Command: &azdext.WorkflowCommand{Args: []string{"auth", "token"}}}

	if step.Command == nil || len(step.Command.Args) != 2 {
		t.Fatal("azdext.WorkflowCommand no longer carries plain args; re-check where validation belongs")
	}
}
