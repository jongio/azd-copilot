// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package commands

import (
	"fmt"
	"strings"
)

// Guard for the run_workflow tool.
//
// run_workflow forwards each step's args straight to azd's root command. azd
// applies no allowlist of its own: the gRPC workflow service converts the
// steps and hands them to the cobra root, so any azd command is reachable,
// including hidden ones such as `auth token`.
//
// That is a consent problem more than an execution one. Nothing here shells
// out, and the workflow response is empty so no command output flows back into
// the model's context. But the tool is advertised to the agent, and to whoever
// approves the call, as running an azd workflow. Someone approving a call
// named run_workflow is agreeing to a project lifecycle sequence, not to
// arbitrary azd invocation, and the two are not the same consent.
//
// The fix is deliberately narrow. Workflow steps are lifecycle commands, so
// this denies only the command groups that mutate credentials or global
// machine state and have no place in a lifecycle sequence. Everything else
// stays permitted, including destructive lifecycle commands like `down`, which
// are legitimate workflow steps and are already covered by the tool's
// Destructive flag.
//
// This is not azdext.WithSecurityPolicy. That policy is a carrier the builder
// stores and never enforces, its URL half is inert here because no tool in
// this extension takes a URL, and its CheckPath half is inert because no tool
// here takes a path at all. It would add no protection to this surface. See
// task 3.5 in the azdext alignment plan for the full evaluation.

// deniedWorkflowCommands are azd command groups that may not appear as a
// workflow step. Each entry needs a reason that survives review.
var deniedWorkflowCommands = map[string]string{
	// Changes which identity every later step runs as, and `auth token` prints
	// an access token. Neither belongs in a lifecycle sequence.
	"auth": "changes or reveals credentials",
	// Writes user-global azd configuration, so its effect outlives the project
	// and the session that ran it.
	"config": "writes machine-global configuration",
}

// validateWorkflowStepArgs checks one step's command arguments.
func validateWorkflowStepArgs(index int, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("step %d has no command arguments", index)
	}

	command := strings.TrimSpace(args[0])
	if command == "" {
		return fmt.Errorf("step %d has an empty command", index)
	}

	// A leading dash means a flag arrived where the command name should be,
	// which is how an injected argument reaches a command the caller did not
	// name (CWE-88).
	if strings.HasPrefix(command, "-") {
		return fmt.Errorf(
			"step %d starts with %q, which is a flag rather than an azd command", index, command)
	}

	if reason, denied := deniedWorkflowCommands[strings.ToLower(command)]; denied {
		return fmt.Errorf(
			"step %d runs 'azd %s', which is not allowed in a workflow because it %s. "+
				"Workflow steps are project lifecycle commands", index, command, reason)
	}

	return nil
}
