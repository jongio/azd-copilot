//go:build mage

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	binaryName    = "copilot"
	srcDir        = "src/cmd/copilot"
	binDir        = "bin"
	extensionFile = "extension.yaml"
	extensionID   = "jongio.azd.copilot"
	goSrcPattern  = "./src/..."
)

// Default target runs all checks and builds.
var Default = All

// killCopilotProcesses terminates any running azd copilot processes.
func killCopilotProcesses() error {
	if runtime.GOOS == "windows" {
		fmt.Println("Stopping any running copilot processes...")
		_ = exec.Command("powershell", "-NoProfile", "-Command",
			"Stop-Process -Name '"+binaryName+"' -Force -ErrorAction SilentlyContinue").Run()

		extensionBinaryPrefix := strings.ReplaceAll(extensionID, ".", "-")
		for _, arch := range []string{"windows-amd64", "windows-arm64"} {
			procName := extensionBinaryPrefix + "-" + arch
			_ = exec.Command("powershell", "-NoProfile", "-Command",
				"Stop-Process -Name '"+procName+"' -Force -ErrorAction SilentlyContinue").Run()
		}
	} else {
		_ = exec.Command("pkill", "-f", binaryName).Run()
		extensionBinaryPrefix := strings.ReplaceAll(extensionID, ".", "-")
		_ = exec.Command("pkill", "-f", extensionBinaryPrefix).Run()
	}
	return nil
}

// getVersion reads the current version from extension.yaml.
func getVersion() (string, error) {
	data, err := os.ReadFile(extensionFile)
	if err != nil {
		return "", fmt.Errorf("failed to read extension.yaml: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "version:") {
			version := strings.TrimSpace(strings.TrimPrefix(line, "version:"))
			return version, nil
		}
	}
	return "", fmt.Errorf("version not found in extension.yaml")
}

// All runs lint, test, and build in dependency order.
func All() error {
	mg.Deps(Fmt, Lint, Test)
	return Build()
}

// Build builds the CLI binary and installs it locally.
func Build() error {
	_ = killCopilotProcesses()

	version, err := getVersion()
	if err != nil {
		return err
	}

	fmt.Println("Building extension...")

	// Determine platform-specific binary name
	binaryExt := ""
	if runtime.GOOS == "windows" {
		binaryExt = ".exe"
	}

	// Create bin directory
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Build the binary with version info
	buildTime := time.Now().UTC().Format(time.RFC3339)
	ldflags := fmt.Sprintf("-X github.com/jongio/azd-copilot/src/cmd/copilot/commands.Version=%s -X github.com/jongio/azd-copilot/src/cmd/copilot/commands.BuildTime=%s", version, buildTime)

	binaryPath := filepath.Join(binDir, binaryName+binaryExt)
	if err := sh.RunV("go", "build", "-ldflags", ldflags, "-o", binaryPath, "./"+srcDir); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Printf("✅ Build complete! Version: %s\n", version)

	// Install the extension
	if err := Install(); err != nil {
		return err
	}

	fmt.Println("   Run 'azd copilot version' to verify")
	return nil
}

// Install installs the extension to the azd extensions directory.
func Install() error {
	version, err := getVersion()
	if err != nil {
		return err
	}

	fmt.Println("Installing extension...")

	// Determine paths
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	goos := runtime.GOOS
	goarch := runtime.GOARCH
	binaryExt := ""
	if goos == "windows" {
		binaryExt = ".exe"
	}

	extensionDir := filepath.Join(homeDir, ".azd", "extensions", extensionID)
	installedBinaryName := strings.ReplaceAll(extensionID, ".", "-") + "-" + goos + "-" + goarch + binaryExt
	installedBinaryPath := filepath.Join(extensionDir, installedBinaryName)

	// Create extension directory
	if err := os.MkdirAll(extensionDir, 0755); err != nil {
		return fmt.Errorf("failed to create extension directory: %w", err)
	}

	// Copy binary
	srcBinary := filepath.Join(binDir, binaryName+binaryExt)
	srcData, err := os.ReadFile(srcBinary)
	if err != nil {
		return fmt.Errorf("failed to read built binary: %w", err)
	}
	if err := os.WriteFile(installedBinaryPath, srcData, 0755); err != nil {
		return fmt.Errorf("failed to write installed binary: %w", err)
	}

	// Copy extension.yaml
	extYamlData, err := os.ReadFile(extensionFile)
	if err != nil {
		return fmt.Errorf("failed to read extension.yaml: %w", err)
	}
	if err := os.WriteFile(filepath.Join(extensionDir, "extension.yaml"), extYamlData, 0644); err != nil {
		return fmt.Errorf("failed to write extension.yaml: %w", err)
	}

	// Update azd config.json to register the extension
	configPath := filepath.Join(homeDir, ".azd", "config.json")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read azd config: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("failed to parse azd config: %w", err)
	}

	// Ensure extension.installed exists
	extension, ok := config["extension"].(map[string]interface{})
	if !ok {
		extension = make(map[string]interface{})
		config["extension"] = extension
	}
	installed, ok := extension["installed"].(map[string]interface{})
	if !ok {
		installed = make(map[string]interface{})
		extension["installed"] = installed
	}

	// Register the extension
	relativePath := filepath.Join("extensions", extensionID, installedBinaryName)
	installed[extensionID] = map[string]interface{}{
		"id":          extensionID,
		"namespace":   "copilot",
		"displayName": "Copilot Extension",
		"description": "GitHub Copilot integration for Azure Developer CLI",
		"usage":       "azd copilot <command> [options]",
		"version":     version,
		"capabilities": []string{"custom-commands"},
		"path":        relativePath,
		"source":      "local",
	}

	// Write updated config
	updatedConfig, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(configPath, updatedConfig, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Println("✅ Extension installed!")
	return nil
}

// Fmt formats all Go source files.
func Fmt() error {
	fmt.Println("Formatting Go code...")
	return sh.RunV("go", "fmt", goSrcPattern)
}

// Lint runs golangci-lint on all Go source files.
func Lint() error {
	fmt.Println("Running linter...")
	return sh.RunV("golangci-lint", "run", goSrcPattern)
}

// Test runs all unit tests.
func Test() error {
	fmt.Println("Running tests...")
	return sh.RunV("go", "test", "-v", goSrcPattern)
}

// Clean removes build artifacts.
func Clean() error {
	fmt.Println("Cleaning build artifacts...")
	return os.RemoveAll(binDir)
}
