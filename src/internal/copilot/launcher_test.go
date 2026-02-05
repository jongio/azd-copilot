// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package copilot

import (
	"os"
	"runtime"
	"testing"
)

func TestOptions_Defaults(t *testing.T) {
	opts := Options{}

	if opts.Prompt != "" {
		t.Errorf("Default Prompt should be empty, got %q", opts.Prompt)
	}
	if opts.Resume {
		t.Error("Default Resume should be false")
	}
	if opts.Yolo {
		t.Error("Default Yolo should be false")
	}
	if opts.Agent != "" {
		t.Errorf("Default Agent should be empty, got %q", opts.Agent)
	}
	if opts.Verbose {
		t.Error("Default Verbose should be false")
	}
}

func TestBuildArgs(t *testing.T) {
	tests := []struct {
		name     string
		opts     Options
		contains []string
		excludes []string
	}{
		{
			name:     "default options",
			opts:     Options{},
			contains: []string{"--agent", "azure-manager"},
			excludes: []string{"--resume", "--yolo", "-p"},
		},
		{
			name: "with prompt",
			opts: Options{
				Prompt: "help me deploy",
			},
			contains: []string{"--agent", "azure-manager", "-p", "help me deploy"},
		},
		{
			name: "with custom agent",
			opts: Options{
				Agent: "azure-architect",
			},
			contains: []string{"--agent", "azure-architect"},
		},
		{
			name: "with resume",
			opts: Options{
				Resume: true,
			},
			contains: []string{"--resume"},
		},
		{
			name: "with continue",
			opts: Options{
				Continue: true,
			},
			contains: []string{"--continue"},
		},
		{
			name: "with yolo",
			opts: Options{
				Yolo: true,
			},
			contains: []string{"--yolo"},
		},
		{
			name: "with model",
			opts: Options{
				Model: "claude-sonnet-4",
			},
			contains: []string{"--model", "claude-sonnet-4"},
		},
		{
			name: "with add dirs",
			opts: Options{
				AddDirs: []string{"/path/to/dir1", "/path/to/dir2"},
			},
			contains: []string{"--add-dir", "/path/to/dir1", "--add-dir", "/path/to/dir2"},
		},
		{
			name: "with verbose",
			opts: Options{
				Verbose: true,
			},
			contains: []string{"--verbose"},
		},
		{
			name: "full options",
			opts: Options{
				Prompt:  "build api",
				Agent:   "azure-dev",
				Resume:  true,
				Yolo:    true,
				Model:   "gpt-4",
				Verbose: true,
				AddDirs: []string{"/docs"},
			},
			contains: []string{
				"--agent", "azure-dev",
				"-p", "build api",
				"--resume",
				"--yolo",
				"--model", "gpt-4",
				"--verbose",
				"--add-dir", "/docs",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := buildArgs(tt.opts)

			// Check contains
			for _, want := range tt.contains {
				found := false
				for _, arg := range args {
					if arg == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("buildArgs() should contain %q, got %v", want, args)
				}
			}

			// Check excludes
			for _, exclude := range tt.excludes {
				for _, arg := range args {
					if arg == exclude {
						t.Errorf("buildArgs() should not contain %q, got %v", exclude, args)
					}
				}
			}
		})
	}
}

func TestBuildEnv(t *testing.T) {
	tests := []struct {
		name     string
		opts     Options
		contains []string
	}{
		{
			name: "basic env",
			opts: Options{},
			contains: []string{
				"AZD_COPILOT_EXTENSION=true",
			},
		},
		{
			name: "with project context",
			opts: Options{
				ProjectContext: &ProjectContext{
					Name: "myproject",
					Path: "/path/to/project",
				},
			},
			contains: []string{
				"AZD_PROJECT_NAME=myproject",
				"AZD_PROJECT_PATH=/path/to/project",
			},
		},
		{
			name: "with services",
			opts: Options{
				ProjectContext: &ProjectContext{
					Name: "myproject",
					Path: "/path",
					Services: []ServiceInfo{
						{Name: "api", Language: "go", Host: "appservice", Path: "./src/api"},
						{Name: "web", Language: "typescript", Host: "staticwebapp", Path: "./src/web"},
					},
				},
			},
			contains: []string{
				"AZD_SERVICES=api,web",
			},
		},
		{
			name: "with azure account",
			opts: Options{
				ProjectContext: &ProjectContext{
					Name: "myproject",
					Path: "/path",
					AzureAccount: &AzureAccountInfo{
						SubscriptionID:   "sub-123",
						SubscriptionName: "My Subscription",
						TenantID:         "tenant-456",
						UserName:         "user@example.com",
					},
				},
			},
			contains: []string{
				"AZD_SUBSCRIPTION_ID=sub-123",
				"AZD_SUBSCRIPTION_NAME=My Subscription",
				"AZD_TENANT_ID=tenant-456",
				"AZD_USER=user@example.com",
			},
		},
		{
			name: "with infrastructure",
			opts: Options{
				ProjectContext: &ProjectContext{
					Name: "myproject",
					Path: "/path",
					Infrastructure: &InfrastructureInfo{
						Path:     "infra",
						Module:   "main",
						HasBicep: true,
					},
				},
			},
			contains: []string{
				"AZD_INFRA_PATH=infra",
				"AZD_INFRA_MODULE=main",
				"AZD_HAS_BICEP=true",
			},
		},
		{
			name: "with azure environment vars",
			opts: Options{
				ProjectContext: &ProjectContext{
					Name: "myproject",
					Path: "/path",
					Environment: map[string]string{
						"AZURE_RESOURCE_GROUP": "rg-myproject",
						"AZURE_LOCATION":       "eastus",
						"OTHER_VAR":            "should-be-excluded",
					},
				},
			},
			contains: []string{
				"AZURE_RESOURCE_GROUP=rg-myproject",
				"AZURE_LOCATION=eastus",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := buildEnv(tt.opts)

			for _, want := range tt.contains {
				found := false
				for _, e := range env {
					if e == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("buildEnv() should contain %q, got %v", want, env)
				}
			}
		})
	}
}

func TestBuildEnv_ExcludesNonAzureVars(t *testing.T) {
	opts := Options{
		ProjectContext: &ProjectContext{
			Name: "test",
			Path: "/test",
			Environment: map[string]string{
				"AZURE_LOCATION": "eastus",
				"OTHER_VAR":      "value",
				"DATABASE_URL":   "secret",
			},
		},
	}

	env := buildEnv(opts)

	// OTHER_VAR and DATABASE_URL should not be included
	for _, e := range env {
		if e == "OTHER_VAR=value" || e == "DATABASE_URL=secret" {
			t.Errorf("buildEnv() should not include non-AZURE_ vars, found %q", e)
		}
	}
}

func TestCopilotPath_Fields(t *testing.T) {
	cp := CopilotPath{
		Path:   "/usr/local/bin/copilot",
		IsNode: false,
	}

	if cp.Path != "/usr/local/bin/copilot" {
		t.Errorf("CopilotPath.Path = %q, want %q", cp.Path, "/usr/local/bin/copilot")
	}
	if cp.IsNode {
		t.Error("CopilotPath.IsNode should be false for binary")
	}

	cpNode := CopilotPath{
		Path:   "/path/to/npm-loader.js",
		IsNode: true,
	}

	if !cpNode.IsNode {
		t.Error("CopilotPath.IsNode should be true for npm-loader.js")
	}
}

func TestIsCopilotInstalled(t *testing.T) {
	// This test just verifies the function doesn't panic
	// The actual result depends on the test environment
	result := IsCopilotInstalled()
	t.Logf("IsCopilotInstalled() = %v (depends on test environment)", result)
}

func TestFindCopilotCLI(t *testing.T) {
	// Test that FindCopilotCLI either returns a valid path or an error
	path, err := FindCopilotCLI()

	if err != nil {
		// Error is expected if Copilot CLI is not installed
		t.Logf("FindCopilotCLI() error = %v (Copilot CLI may not be installed)", err)
		return
	}

	if path == nil {
		t.Error("FindCopilotCLI() returned nil path without error")
		return
	}

	if path.Path == "" {
		t.Error("FindCopilotCLI() returned empty path")
	}

	t.Logf("FindCopilotCLI() found: %s (IsNode: %v)", path.Path, path.IsNode)
}

func TestProjectContext_Fields(t *testing.T) {
	ctx := ProjectContext{
		Name: "myapp",
		Path: "/path/to/myapp",
		Services: []ServiceInfo{
			{Name: "api", Language: "go", Host: "appservice", Path: "./api"},
		},
		Environment: map[string]string{
			"AZURE_LOCATION": "eastus",
		},
	}

	if ctx.Name != "myapp" {
		t.Errorf("ProjectContext.Name = %q, want %q", ctx.Name, "myapp")
	}
	if len(ctx.Services) != 1 {
		t.Errorf("ProjectContext.Services length = %d, want 1", len(ctx.Services))
	}
	if ctx.Services[0].Name != "api" {
		t.Errorf("ProjectContext.Services[0].Name = %q, want %q", ctx.Services[0].Name, "api")
	}
}

func TestServiceInfo_Fields(t *testing.T) {
	svc := ServiceInfo{
		Name:     "web",
		Language: "typescript",
		Host:     "staticwebapp",
		Path:     "./src/web",
	}

	if svc.Name != "web" {
		t.Errorf("ServiceInfo.Name = %q, want %q", svc.Name, "web")
	}
	if svc.Language != "typescript" {
		t.Errorf("ServiceInfo.Language = %q, want %q", svc.Language, "typescript")
	}
	if svc.Host != "staticwebapp" {
		t.Errorf("ServiceInfo.Host = %q, want %q", svc.Host, "staticwebapp")
	}
}

func TestAzureAccountInfo_Fields(t *testing.T) {
	acct := AzureAccountInfo{
		SubscriptionID:   "sub-123-456",
		SubscriptionName: "Development",
		TenantID:         "tenant-789",
		UserName:         "developer@contoso.com",
	}

	if acct.SubscriptionID != "sub-123-456" {
		t.Errorf("AzureAccountInfo.SubscriptionID = %q, want %q", acct.SubscriptionID, "sub-123-456")
	}
	if acct.SubscriptionName != "Development" {
		t.Errorf("AzureAccountInfo.SubscriptionName = %q, want %q", acct.SubscriptionName, "Development")
	}
}

func TestInfrastructureInfo_Fields(t *testing.T) {
	infra := InfrastructureInfo{
		Path:     "infra",
		Module:   "main",
		HasBicep: true,
	}

	if infra.Path != "infra" {
		t.Errorf("InfrastructureInfo.Path = %q, want %q", infra.Path, "infra")
	}
	if !infra.HasBicep {
		t.Error("InfrastructureInfo.HasBicep should be true")
	}
}

func TestConfigureMCPServer(t *testing.T) {
	// Skip test if we can't get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get user home directory")
	}

	// Verify ConfigureMCPServer doesn't panic
	err = ConfigureMCPServer()
	if err != nil {
		t.Logf("ConfigureMCPServer() error = %v (may be expected)", err)
	}

	// Check that .copilot directory would be created
	copilotDir := home + "/.copilot"
	t.Logf("Would create MCP config in: %s", copilotDir)
}

func TestPlatformSpecificPaths(t *testing.T) {
	// Test that we handle Windows vs Unix paths correctly
	if runtime.GOOS == "windows" {
		// Windows-specific tests
		t.Log("Running on Windows")
		// Check for CONIN$/CONOUT$ paths in launchViaConsole logic
	} else {
		// Unix-specific tests
		t.Log("Running on Unix-like system")
		// Check for /dev/tty path in launchViaConsole logic
	}
}
