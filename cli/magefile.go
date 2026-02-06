//go:build mage

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

// killExtensionProcesses terminates any running azd copilot extension processes.
// It only targets the installed extension binaries (jongio-azd-copilot-*), NOT the
// generic "copilot" process, which would kill GitHub Copilot CLI sessions.
func killExtensionProcesses() error {
	extensionBinaryPrefix := strings.ReplaceAll(extensionID, ".", "-")

	if runtime.GOOS == "windows" {
		fmt.Println("Stopping any running extension processes...")
		for _, arch := range []string{"windows-amd64", "windows-arm64"} {
			procName := extensionBinaryPrefix + "-" + arch
			_ = exec.Command("powershell", "-NoProfile", "-Command",
				"Stop-Process -Name '"+procName+"' -Force -ErrorAction SilentlyContinue").Run()
		}
	} else {
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
	_ = killExtensionProcesses()

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
	commit := "unknown"
	if out, err := sh.Output("git", "rev-parse", "HEAD"); err == nil && out != "" {
		commit = out
	}
	ldflags := fmt.Sprintf("-X github.com/jongio/azd-copilot/cli/src/cmd/copilot/commands.Version=%s -X github.com/jongio/azd-copilot/cli/src/cmd/copilot/commands.BuildTime=%s -X github.com/jongio/azd-copilot/cli/src/cmd/copilot/commands.Commit=%s", version, buildTime, commit)

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

	var config map[string]interface{}
	if err != nil {
		// config.json doesn't exist yet — start with empty config
		config = make(map[string]interface{})
	} else if err := json.Unmarshal(configData, &config); err != nil {
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

const (
	skillsSourceRepo = "https://github.com/microsoft/GitHub-Copilot-for-Azure.git"
	skillsSourcePath = "plugin/skills"
	skillsTargetPath = "src/internal/assets/skills"
)

// SyncSkills syncs upstream skills from microsoft/GitHub-Copilot-for-Azure.
// Set SKILLS_SOURCE to a local clone path to skip cloning.
func SyncSkills() error {
	fmt.Println("🔄 Syncing upstream Azure skills...")

	sourceDir := os.Getenv("SKILLS_SOURCE")
	var tempDir string

	if sourceDir != "" {
		// Use local clone
		skillsDir := filepath.Join(sourceDir, skillsSourcePath)
		if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
			return fmt.Errorf("skills not found at %s", skillsDir)
		}
		fmt.Printf("📂 Using local source: %s\n", sourceDir)
	} else {
		// Sparse clone just the skills directory
		fmt.Println("📥 Cloning upstream repo (sparse)...")
		var err error
		tempDir, err = os.MkdirTemp("", "skills-sync-*")
		if err != nil {
			return fmt.Errorf("failed to create temp dir: %w", err)
		}
		defer os.RemoveAll(tempDir)

		sourceDir = tempDir
		cmds := [][]string{
			{"git", "clone", "--depth=1", "--filter=blob:none", "--sparse", skillsSourceRepo, tempDir},
			{"git", "-C", tempDir, "sparse-checkout", "set", skillsSourcePath},
		}
		for _, args := range cmds {
			if err := sh.RunV(args[0], args[1:]...); err != nil {
				return fmt.Errorf("git command failed: %w", err)
			}
		}
	}

	skillsSrc := filepath.Join(sourceDir, skillsSourcePath)

	// Read source skills
	entries, err := os.ReadDir(skillsSrc)
	if err != nil {
		return fmt.Errorf("failed to read source skills: %w", err)
	}

	var skillDirs []string
	for _, e := range entries {
		if e.IsDir() {
			skillDirs = append(skillDirs, e.Name())
		}
	}
	fmt.Printf("📦 Found %d upstream skills\n\n", len(skillDirs))

	// Wipe target (safe — only upstream content lives here)
	if err := os.RemoveAll(skillsTargetPath); err != nil {
		return fmt.Errorf("failed to clean target: %w", err)
	}
	if err := os.MkdirAll(skillsTargetPath, 0755); err != nil {
		return fmt.Errorf("failed to create target: %w", err)
	}

	// Copy each skill
	copied := 0
	for _, name := range skillDirs {
		src := filepath.Join(skillsSrc, name)
		dst := filepath.Join(skillsTargetPath, name)
		if err := copyDir(src, dst); err != nil {
			fmt.Printf("  ❌ %s: %v\n", name, err)
			continue
		}
		fmt.Printf("  ✅ %s\n", name)
		copied++
	}

	fmt.Printf("\n✨ Synced %d/%d upstream skills to %s\n", copied, len(skillDirs), skillsTargetPath)

	// Update counts in static files
	return UpdateCounts()
}

// copyDir recursively copies a directory tree.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0644)
	})
}

const (
	customSkillsPath = "src/internal/assets/custom-skills"
	agentsPath       = "src/internal/assets/agents"
	countsFile       = "counts.json"
)

type counts struct {
	Agents int `json:"agents"`
	Skills int `json:"skills"`
}

// countDirsWithSkillMD counts subdirectories containing a SKILL.md file.
func countDirsWithSkillMD(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, e.Name(), "SKILL.md")); err == nil {
			n++
		}
	}
	return n
}

// countAgentFiles counts .md files in the agents directory.
func countAgentFiles(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			n++
		}
	}
	return n
}

// UpdateCounts scans agents and skills directories, writes counts.json,
// and patches hardcoded counts in static files (extension.yaml, web, docs).
func UpdateCounts() error {
	fmt.Println("\n📊 Updating counts...")

	agentCount := countAgentFiles(agentsPath)
	skillCount := countDirsWithSkillMD(skillsTargetPath) + countDirsWithSkillMD(customSkillsPath)

	c := counts{Agents: agentCount, Skills: skillCount}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(countsFile, data, 0644); err != nil {
		return err
	}
	fmt.Printf("  ✅ %s: %d agents, %d skills\n", countsFile, c.Agents, c.Skills)

	// Patch static files with updated counts
	replacements := []struct {
		file    string
		patches []struct{ old, new string }
	}{
		{
			file: extensionFile,
			patches: []struct{ old, new string }{
				{`Includes \d+ specialized agents`, fmt.Sprintf("Includes %d specialized agents", c.Agents)},
				{`and \d+ Azure skills`, fmt.Sprintf("and %d Azure skills", c.Skills)},
			},
		},
		{
			file: "../web/src/pages/index.astro",
			patches: []struct{ old, new string }{
				{`"(\d+) Specialized Agents"`, fmt.Sprintf(`"%d Specialized Agents"`, c.Agents)},
				{`"(\d+) Azure Skills"`, fmt.Sprintf(`"%d Azure Skills"`, c.Skills)},
				{`value: "(\d+)", label: "Agents"`, fmt.Sprintf(`value: "%d", label: "Agents"`, c.Agents)},
				{`value: "(\d+)", label: "Skills"`, fmt.Sprintf(`value: "%d", label: "Skills"`, c.Skills)},
				{`(\d+) specialized agents and (\d+) curated skills`, fmt.Sprintf("%d specialized agents and %d curated skills", c.Agents, c.Skills)},
				{`<strong>(\d+) specialized agents</strong>`, fmt.Sprintf("<strong>%d specialized agents</strong>", c.Agents)},
				{`<strong>(\d+) curated skills</strong>`, fmt.Sprintf("<strong>%d curated skills</strong>", c.Skills)},
			},
		},
	}

	// Also patch custom skill docs that mention counts
	customSkillPatches := []string{
		"src/internal/assets/custom-skills/marketing/messaging.md",
		"src/internal/assets/custom-skills/marketing/azure-marketplace.md",
		"src/internal/assets/custom-skills/support/faq.md",
	}
	agentPattern := `(\d+) specialized (?:AI )?agents`
	agentReplace := fmt.Sprintf("%d specialized AI agents", c.Agents)
	for _, f := range customSkillPatches {
		if _, err := os.Stat(f); err == nil {
			replacements = append(replacements, struct {
				file    string
				patches []struct{ old, new string }
			}{
				file: f,
				patches: []struct{ old, new string }{
					{agentPattern, agentReplace},
				},
			})
		}
	}

	// Patch agents.ts "coordinates all N other agents"
	replacements = append(replacements, struct {
		file    string
		patches []struct{ old, new string }
	}{
		file: "../web/src/data/agents.ts",
		patches: []struct{ old, new string }{
			{`all \d+ other agents`, fmt.Sprintf("all %d other agents", c.Agents-1)},
		},
	})

	for _, r := range replacements {
		content, err := os.ReadFile(r.file)
		if err != nil {
			fmt.Printf("  ⚠️  %s: %v\n", r.file, err)
			continue
		}
		text := string(content)
		changed := false
		for _, p := range r.patches {
			updated := regexpReplace(text, p.old, p.new)
			if updated != text {
				text = updated
				changed = true
			}
		}
		if changed {
			if err := os.WriteFile(r.file, []byte(text), 0644); err != nil {
				fmt.Printf("  ❌ %s: %v\n", r.file, err)
				continue
			}
			fmt.Printf("  ✅ %s\n", r.file)
		}
	}

	fmt.Println("\n✨ Counts updated!")
	return nil
}

// regexpReplace replaces all matches of pattern with replacement in text.
func regexpReplace(text, pattern, replacement string) string {
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(text, replacement)
}
