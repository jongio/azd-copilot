// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package copilot

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/jongio/azd-core/fileutil"
)

// Version is set by the main package at startup.
var Version = "0.1.0"

// Options configures the Copilot CLI launch
type Options struct {
	Prompt         string
	Resume         bool
	Continue       bool
	Yolo           bool
	Agent          string
	Model          string
	AddDirs        []string
	Verbose        bool
	Debug          bool
	ProjectContext *ProjectContext
}

// ProjectContext contains azd project information
type ProjectContext struct {
	Name           string
	Path           string
	Services       []ServiceInfo
	Environment    map[string]string
	AzureAccount   *AzureAccountInfo
	Infrastructure *InfrastructureInfo
}

// ServiceInfo contains service details
type ServiceInfo struct {
	Name     string
	Language string
	Host     string
	Path     string
}

// AzureAccountInfo contains Azure account details
type AzureAccountInfo struct {
	SubscriptionID   string
	SubscriptionName string
	TenantID         string
	UserName         string
}

// InfrastructureInfo contains infrastructure details
type InfrastructureInfo struct {
	Path     string
	Module   string
	HasBicep bool
}

// Launch starts the GitHub Copilot CLI with configured options
func Launch(ctx context.Context, opts Options) error {
	copilotPath, err := FindCopilotCLI()
	if err != nil {
		return err
	}

	args := buildArgs(opts)

	if opts.Debug {
		fmt.Printf("DEBUG: Copilot path: %s (IsNode: %v)\n", copilotPath.Path, copilotPath.IsNode)
		fmt.Printf("DEBUG: Args: %v\n", args)
	}

	// Check if we're running via azd (which captures stdio)
	isRunningViaAzd := os.Getenv("AZD_SERVER") != ""
	isInteractive := opts.Prompt == ""

	if opts.Debug {
		fmt.Printf("DEBUG: Running via azd: %v, Interactive: %v\n", isRunningViaAzd, isInteractive)
	}

	// When running interactively via azd, use direct console/tty access
	// which bypasses azd's stdio capture for TUI apps
	if isInteractive && isRunningViaAzd {
		if opts.Debug {
			fmt.Printf("DEBUG: Using direct console/tty for azd interactive mode\n")
		}
		return launchViaConsole(ctx, copilotPath, args, opts)
	}

	var cmd *exec.Cmd

	if copilotPath.IsNode {
		// Run via node
		nodeArgs := append([]string{copilotPath.Path}, args...)
		cmd = exec.CommandContext(ctx, "node", nodeArgs...)
		if opts.Debug || os.Getenv("AZD_COPILOT_DEBUG") == "true" {
			fmt.Printf("DEBUG: Running via node\n")
		}
	} else if runtime.GOOS == "windows" && (strings.HasSuffix(copilotPath.Path, ".bat") || strings.HasSuffix(copilotPath.Path, ".cmd")) {
		// On Windows, .bat/.cmd files need to be run via cmd.exe
		cmdArgs := append([]string{"/c", copilotPath.Path}, args...)
		cmd = exec.CommandContext(ctx, "cmd.exe", cmdArgs...)
		if opts.Debug || os.Getenv("AZD_COPILOT_DEBUG") == "true" {
			fmt.Printf("DEBUG: Running via cmd.exe /c\n")
		}
	} else {
		cmd = exec.CommandContext(ctx, copilotPath.Path, args...)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set environment
	env := append(os.Environ(), buildEnv(opts)...)

	// Force terminal settings for interactive mode
	if isInteractive {
		env = append(env, "FORCE_COLOR=1")
		env = append(env, "CI=false")
	}

	cmd.Env = env

	return cmd.Run()
}

// launchViaConsole bypasses azd's stdio capture for interactive TUI apps.
// On macOS/Linux it opens /dev/tty directly.
// On Windows it uses SetStdHandle to redirect the process's standard handles
// to CONIN$/CONOUT$ before spawning the child, so Node.js detects a real TTY
// (isTTY=true, columns/rows populated). Simply passing CONOUT$ file handles
// as cmd.Stdout does NOT work — Node.js only recognises a TTY when the
// underlying Windows handle was the process's standard output at spawn time.
func launchViaConsole(ctx context.Context, copilotPath *CopilotPath, args []string, opts Options) error {
	if opts.Debug {
		fmt.Printf("DEBUG: Opening console/tty directly for interactive mode\n")
	}

	var cmd *exec.Cmd

	if copilotPath.IsNode {
		nodeArgs := append([]string{copilotPath.Path}, args...)
		cmd = exec.CommandContext(ctx, "node", nodeArgs...)
	} else if runtime.GOOS == "windows" && (strings.HasSuffix(copilotPath.Path, ".bat") || strings.HasSuffix(copilotPath.Path, ".cmd")) {
		cmdArgs := append([]string{"/c", copilotPath.Path}, args...)
		cmd = exec.CommandContext(ctx, "cmd.exe", cmdArgs...)
	} else {
		cmd = exec.CommandContext(ctx, copilotPath.Path, args...)
	}

	if runtime.GOOS == "windows" {
		// On Windows, use SetStdHandle to point our process's standard handles
		// at the console (CONIN$/CONOUT$). The child process then inherits real
		// console handles that Node.js recognises as a TTY.
		consoleH, err := attachConsole()
		if err != nil {
			if opts.Debug {
				fmt.Printf("DEBUG: Failed to attach console: %v, falling back\n", err)
			}
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		} else {
			defer consoleH.restore()
			cmd.Stdin = consoleH.conin
			cmd.Stdout = consoleH.conout
			cmd.Stderr = consoleH.conout
		}
	} else {
		// On macOS/Linux, open /dev/tty to bypass pipe redirection
		tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if err != nil {
			if opts.Debug {
				fmt.Printf("DEBUG: Failed to open /dev/tty: %v, falling back\n", err)
			}
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		} else {
			cmd.Stdin = tty
			cmd.Stdout = tty
			cmd.Stderr = tty
			defer tty.Close()
		}
	}

	// Build environment
	env := append(os.Environ(), buildEnv(opts)...)
	env = append(env, "FORCE_COLOR=1")
	env = append(env, "CI=false")
	cmd.Env = env

	if opts.Debug {
		fmt.Printf("DEBUG: execPath=%s, args=%v\n", cmd.Path, cmd.Args)
	}

	return cmd.Run()
}

// CopilotPath contains information about how to run copilot
type CopilotPath struct {
	Path   string // Path to the executable or script
	IsNode bool   // If true, run via "node <path>"
}

// FindCopilotCLI locates the GitHub Copilot CLI executable
func FindCopilotCLI() (*CopilotPath, error) {
	// Platform-specific locations - check these FIRST before PATH
	// to avoid finding npm shims or .ps1 launchers (which cause file lock issues)
	if runtime.GOOS == "windows" {
		home := os.Getenv("USERPROFILE")
		appData := os.Getenv("APPDATA")

		// Check for npm-loader.js directly (most reliable on Windows)
		// This avoids issues with .bat/.cmd/.ps1 bootstrap scripts
		npmGlobalPath := os.Getenv("npm_config_prefix")
		if npmGlobalPath == "" {
			// Query npm for its global prefix (handles custom configurations)
			if out, err := exec.Command("npm", "config", "get", "prefix").Output(); err == nil {
				npmGlobalPath = strings.TrimSpace(string(out))
			}
		}
		if npmGlobalPath != "" {
			npmLoader := filepath.Join(npmGlobalPath, "node_modules", "@github", "copilot", "npm-loader.js")
			if _, err := os.Stat(npmLoader); err == nil {
				return &CopilotPath{Path: npmLoader, IsNode: true}, nil
			}
		}

		// NVM for Windows
		nvmRoot := os.Getenv("NVM_HOME")
		if nvmRoot != "" {
			currentVersion := os.Getenv("NVM_SYMLINK")
			if currentVersion != "" {
				npmLoader := filepath.Join(currentVersion, "node_modules", "@github", "copilot", "npm-loader.js")
				if _, err := os.Stat(npmLoader); err == nil {
					return &CopilotPath{Path: npmLoader, IsNode: true}, nil
				}
			}
		}
		// Check appdata nvm location
		nvmPatterns := []string{
			filepath.Join(appData, "nvm", "*", "node_modules", "@github", "copilot", "npm-loader.js"),
		}
		for _, pattern := range nvmPatterns {
			matches, _ := filepath.Glob(pattern)
			if len(matches) > 0 {
				// Use the newest version (last in sorted order)
				return &CopilotPath{Path: matches[len(matches)-1], IsNode: true}, nil
			}
		}

		// Check Program Files nodejs location
		npmLoader := filepath.Join("C:", "Program Files", "nodejs", "node_modules", "@github", "copilot", "npm-loader.js")
		if _, err := os.Stat(npmLoader); err == nil {
			return &CopilotPath{Path: npmLoader, IsNode: true}, nil
		}

		// Priority list of known installation locations (.exe files)
		candidates := []string{
			// Direct installation
			filepath.Join(home, "AppData", "Local", "Programs", "copilot-cli", "copilot.exe"),
		}

		// Check WinGet installation path with glob
		wingetPattern := filepath.Join(home, "AppData", "Local", "Microsoft", "WinGet", "Packages", "GitHub.Copilot_*", "copilot.exe")
		matches, _ := filepath.Glob(wingetPattern)
		candidates = append(candidates, matches...)

		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				return &CopilotPath{Path: candidate, IsNode: false}, nil
			}
		}
	} else {
		// macOS/Linux
		home, _ := os.UserHomeDir()

		// Check npm global node_modules first (most reliable)
		npmGlobalPaths := []string{
			// Standard npm global location
			"/usr/local/lib/node_modules/@github/copilot/npm-loader.js",
			"/usr/lib/node_modules/@github/copilot/npm-loader.js",
			// User-local npm global
			filepath.Join(home, ".npm-global", "lib", "node_modules", "@github", "copilot", "npm-loader.js"),
		}

		// nvm locations - check current node version via NVM_BIN
		nvmBin := os.Getenv("NVM_BIN")
		if nvmBin != "" {
			// NVM_BIN is like ~/.nvm/versions/node/v20.x.x/bin, node_modules is sibling to bin
			nodeVersionDir := filepath.Dir(nvmBin)
			npmLoader := filepath.Join(nodeVersionDir, "lib", "node_modules", "@github", "copilot", "npm-loader.js")
			npmGlobalPaths = append([]string{npmLoader}, npmGlobalPaths...)
		}

		// fnm (Fast Node Manager) locations
		fnmDir := os.Getenv("FNM_MULTISHELL_PATH")
		if fnmDir != "" {
			npmLoader := filepath.Join(fnmDir, "lib", "node_modules", "@github", "copilot", "npm-loader.js")
			npmGlobalPaths = append([]string{npmLoader}, npmGlobalPaths...)
		}

		// Homebrew on macOS
		if runtime.GOOS == "darwin" {
			brewPaths := []string{
				"/opt/homebrew/lib/node_modules/@github/copilot/npm-loader.js",
				"/usr/local/lib/node_modules/@github/copilot/npm-loader.js",
			}
			npmGlobalPaths = append(npmGlobalPaths, brewPaths...)
		}

		for _, npmLoader := range npmGlobalPaths {
			if _, err := os.Stat(npmLoader); err == nil {
				return &CopilotPath{Path: npmLoader, IsNode: true}, nil
			}
		}

		// Binary candidates
		candidates := []string{
			"/usr/local/bin/copilot",
			filepath.Join(home, ".local", "bin", "copilot"),
		}

		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				return &CopilotPath{Path: candidate, IsNode: false}, nil
			}
		}
	}

	// Fall back to PATH lookup (may find .cmd/.bat but that's last resort)
	// Skip .ps1 files - they cause file lock issues on Windows and aren't cross-platform
	if path, err := exec.LookPath("copilot"); err == nil {
		if !strings.HasSuffix(strings.ToLower(path), ".ps1") {
			return &CopilotPath{Path: path, IsNode: false}, nil
		}
	}

	installHint := "npm install -g @github/copilot"
	if runtime.GOOS == "windows" {
		installHint = "winget install GitHub.Copilot"
	}
	return nil, fmt.Errorf("GitHub Copilot CLI not found. Install with: %s", installHint)
}

// IsCopilotInstalled checks if Copilot CLI is available
func IsCopilotInstalled() bool {
	_, err := FindCopilotCLI()
	return err == nil
}

func buildArgs(opts Options) []string {
	var args []string

	// Agent (default to azure-manager)
	agent := opts.Agent
	if agent == "" {
		agent = "azure-manager"
	}
	args = append(args, "--agent", agent)

	// Prompt
	if opts.Prompt != "" {
		args = append(args, "-p", opts.Prompt)
	}

	// Session management
	if opts.Resume {
		args = append(args, "--resume")
	}
	if opts.Continue {
		args = append(args, "--continue")
	}

	// Auto-approve
	if opts.Yolo {
		args = append(args, "--yolo")
	}

	// Model
	if opts.Model != "" {
		args = append(args, "--model", opts.Model)
	}

	// Additional directories
	for _, dir := range opts.AddDirs {
		args = append(args, "--add-dir", dir)
	}

	// Verbose
	if opts.Verbose {
		args = append(args, "--verbose")
	}

	return args
}

func buildEnv(opts Options) []string {
	env := []string{
		"AZD_COPILOT_EXTENSION=true",
		fmt.Sprintf("AZD_COPILOT_VERSION=%s", Version),
	}

	// Inject project context as environment variables
	if opts.ProjectContext != nil {
		env = append(env,
			fmt.Sprintf("AZD_PROJECT_NAME=%s", opts.ProjectContext.Name),
			fmt.Sprintf("AZD_PROJECT_PATH=%s", opts.ProjectContext.Path),
		)

		if len(opts.ProjectContext.Services) > 0 {
			var names []string
			var serviceDetails []string
			for _, svc := range opts.ProjectContext.Services {
				names = append(names, svc.Name)
				// Format: name:language:host:path
				serviceDetails = append(serviceDetails, fmt.Sprintf("%s:%s:%s:%s", svc.Name, svc.Language, svc.Host, svc.Path))
			}
			env = append(env, fmt.Sprintf("AZD_SERVICES=%s", strings.Join(names, ",")))
			env = append(env, fmt.Sprintf("AZD_SERVICE_DETAILS=%s", strings.Join(serviceDetails, ";")))
		}

		// Include Azure account info
		if opts.ProjectContext.AzureAccount != nil {
			acct := opts.ProjectContext.AzureAccount
			if acct.SubscriptionID != "" {
				env = append(env, fmt.Sprintf("AZD_SUBSCRIPTION_ID=%s", acct.SubscriptionID))
			}
			if acct.SubscriptionName != "" {
				env = append(env, fmt.Sprintf("AZD_SUBSCRIPTION_NAME=%s", acct.SubscriptionName))
			}
			if acct.TenantID != "" {
				env = append(env, fmt.Sprintf("AZD_TENANT_ID=%s", acct.TenantID))
			}
			if acct.UserName != "" {
				env = append(env, fmt.Sprintf("AZD_USER=%s", acct.UserName))
			}
		}

		// Include infrastructure info
		if opts.ProjectContext.Infrastructure != nil {
			infra := opts.ProjectContext.Infrastructure
			if infra.Path != "" {
				env = append(env, fmt.Sprintf("AZD_INFRA_PATH=%s", infra.Path))
			}
			if infra.Module != "" {
				env = append(env, fmt.Sprintf("AZD_INFRA_MODULE=%s", infra.Module))
			}
			if infra.HasBicep {
				env = append(env, "AZD_HAS_BICEP=true")
			}
		}

		// Include environment variables (AZURE_* from azd env)
		for k, v := range opts.ProjectContext.Environment {
			if strings.HasPrefix(k, "AZURE_") {
				env = append(env, fmt.Sprintf("%s=%s", k, v))
			}
		}
	}

	return env
}

// ConfigureMCPServer ensures Azure and azd MCP servers are configured in ~/.copilot/mcp-config.json
// This is for REGISTERING external MCP servers that GitHub Copilot CLI will use
func ConfigureMCPServer() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	copilotDir := filepath.Join(home, ".copilot")
	configPath := filepath.Join(copilotDir, "mcp-config.json")

	// Create .copilot directory if needed
	if err := fileutil.EnsureDir(copilotDir); err != nil {
		return fmt.Errorf("failed to create .copilot directory: %w", err)
	}

	// Required MCP servers
	requiredServers := map[string]string{
		"azure": `{
      "type": "local",
      "command": "npx",
      "args": ["-y", "@azure/mcp@latest", "server", "start"],
      "tools": ["*"]
    }`,
		"azd": `{
      "type": "local",
      "command": "azd",
      "args": ["mcp", "server"],
      "tools": ["*"]
    }`,
		"microsoft-learn": `{
      "type": "local",
      "command": "npx",
      "args": ["-y", "@anthropic/mcp-microsoft-learn@latest"],
      "tools": ["*"]
    }`,
		"context7": `{
      "type": "local",
      "command": "npx",
      "args": ["-y", "@upstash/context7-mcp@latest"],
      "tools": ["*"]
    }`,
		"azd-app": `{
      "type": "local",
      "command": "azd",
      "args": ["copilot", "mcp", "serve"],
      "tools": ["*"]
    }`,
	}

	// Read existing config
	existingConfig, err := os.ReadFile(configPath)

	// Check which servers are missing
	var missingServers []string
	for name := range requiredServers {
		if err != nil || !strings.Contains(string(existingConfig), `"`+name+`"`) {
			missingServers = append(missingServers, name)
		}
	}

	// All servers present
	if len(missingServers) == 0 {
		return nil
	}

	// If no existing config, create new one
	if err != nil || len(existingConfig) < 10 {
		mcpConfig := `{
  "mcpServers": {
    "azure": {
      "type": "local",
      "command": "npx",
      "args": ["-y", "@azure/mcp@latest", "server", "start"],
      "tools": ["*"]
    },
    "azd": {
      "type": "local",
      "command": "azd",
      "args": ["mcp", "server"],
      "tools": ["*"]
    },
    "microsoft-learn": {
      "type": "local",
      "command": "npx",
      "args": ["-y", "@anthropic/mcp-microsoft-learn@latest"],
      "tools": ["*"]
    },
    "context7": {
      "type": "local",
      "command": "npx",
      "args": ["-y", "@upstash/context7-mcp@latest"],
      "tools": ["*"]
    },
    "azd-app": {
      "type": "local",
      "command": "azd",
      "args": ["copilot", "mcp", "serve"],
      "tools": ["*"]
    }
  }
}
`
		if err := fileutil.AtomicWriteFile(configPath, []byte(mcpConfig), 0644); err != nil {
			return fmt.Errorf("failed to write mcp-config.json: %w", err)
		}
		return nil
	}

	// Existing config exists - merge in missing servers
	var config map[string]interface{}
	if err := json.Unmarshal(existingConfig, &config); err != nil {
		fmt.Printf("   Note: Add these MCP servers to ~/.copilot/mcp-config.json: %s\n", strings.Join(missingServers, ", "))
		return nil
	}

	servers, ok := config["mcpServers"].(map[string]interface{})
	if !ok {
		servers = make(map[string]interface{})
		config["mcpServers"] = servers
	}

	for _, name := range missingServers {
		var serverConfig interface{}
		if err := json.Unmarshal([]byte(requiredServers[name]), &serverConfig); err != nil {
			continue
		}
		servers[name] = serverConfig
	}

	updated, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal mcp-config.json: %w", err)
	}

	if err := fileutil.AtomicWriteFile(configPath, append(updated, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write mcp-config.json: %w", err)
	}

	return nil
}

// EnsureExtensionsInstalled checks and installs required azd extensions
func EnsureExtensionsInstalled() error {
	extensions := []struct {
		id     string
		name   string
		source string
	}{
		{"jongio.azd.app", "App Extension", "app"},
		{"jongio.azd.exec", "Exec Extension", "azd-exec"},
	}

	for _, ext := range extensions {
		// Check if extension is installed using a quick timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		cmd := exec.CommandContext(ctx, "azd", "extension", "show", ext.id)
		cmd.Stdout = nil
		cmd.Stderr = nil
		err := cmd.Run()
		cancel()

		if err != nil {
			// Extension not installed or check timed out, try to install
			fmt.Printf("📦 Installing %s...\n", ext.name)
			installCtx, installCancel := context.WithTimeout(context.Background(), 60*time.Second)
			installCmd := exec.CommandContext(installCtx, "azd", "extension", "install", ext.id, "--source", ext.source, "--no-prompt")
			installCmd.Stdout = nil
			installCmd.Stderr = nil
			// Silently skip errors - extension might already be installed or source not configured
			_ = installCmd.Run()
			installCancel()
		}
	}

	return nil
}
