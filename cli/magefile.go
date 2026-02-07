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
	skillsTargetPath = "src/internal/assets/ghcp4a-skills"
)

// SyncSkills syncs upstream skills from microsoft/GitHub-Copilot-for-Azure
// using a smart merge: new upstream files are added, deleted upstream files are
// removed, but locally modified files are preserved.
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

	// Ensure target directory exists
	if err := os.MkdirAll(skillsTargetPath, 0755); err != nil {
		return fmt.Errorf("failed to create target: %w", err)
	}

	// Smart merge: compare upstream vs local
	added, updated, kept, removed := 0, 0, 0, 0

	// Build set of upstream skill names
	upstreamSet := make(map[string]bool)
	for _, name := range skillDirs {
		upstreamSet[name] = true
	}

	// Remove local skills that no longer exist upstream
	localEntries, _ := os.ReadDir(skillsTargetPath)
	for _, e := range localEntries {
		if !e.IsDir() {
			continue
		}
		if !upstreamSet[e.Name()] {
			dst := filepath.Join(skillsTargetPath, e.Name())
			// Check if locally modified (has uncommitted changes via git)
			locallyModified := isLocallyModified(dst)
			if locallyModified {
				fmt.Printf("  ⚠️  %s: removed upstream but locally modified — keeping\n", e.Name())
				kept++
			} else {
				os.RemoveAll(dst)
				fmt.Printf("  🗑️  %s: removed (no longer upstream)\n", e.Name())
				removed++
			}
		}
	}

	// Sync each upstream skill
	for _, name := range skillDirs {
		src := filepath.Join(skillsSrc, name)
		dst := filepath.Join(skillsTargetPath, name)

		if _, err := os.Stat(dst); os.IsNotExist(err) {
			// New skill — copy it in
			if err := copyDir(src, dst); err != nil {
				fmt.Printf("  ❌ %s: %v\n", name, err)
				continue
			}
			fmt.Printf("  ✅ %s (new)\n", name)
			added++
		} else {
			// Existing skill — smart merge per file
			fileAdded, fileUpdated, fileKept, err := mergeSkillDir(src, dst)
			if err != nil {
				fmt.Printf("  ❌ %s: %v\n", name, err)
				continue
			}
			if fileAdded+fileUpdated > 0 {
				fmt.Printf("  ✅ %s (merged: %d new, %d updated, %d kept)\n", name, fileAdded, fileUpdated, fileKept)
			} else if fileKept > 0 {
				fmt.Printf("  🔒 %s (all %d files locally modified — kept)\n", name, fileKept)
			} else {
				fmt.Printf("  ✅ %s (unchanged)\n", name)
			}
			added += fileAdded
			updated += fileUpdated
			kept += fileKept
		}
	}

	fmt.Printf("\n✨ Sync complete: %d added, %d updated, %d locally modified (kept), %d removed\n",
		added, updated, kept, removed)

	// Update counts in static files
	return UpdateCounts()
}

// mergeSkillDir merges an upstream skill directory into a local one.
// Files modified locally are preserved; new/unchanged upstream files are copied.
func mergeSkillDir(src, dst string) (added, updated, kept int, err error) {
	return added, updated, kept, filepath.Walk(src, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			rel, _ := filepath.Rel(src, path)
			return os.MkdirAll(filepath.Join(dst, rel), 0755)
		}

		rel, _ := filepath.Rel(src, path)
		dstFile := filepath.Join(dst, rel)

		upstreamData, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		localData, localErr := os.ReadFile(dstFile)
		if localErr != nil {
			// File doesn't exist locally — add it
			if err := os.MkdirAll(filepath.Dir(dstFile), 0755); err != nil {
				return err
			}
			added++
			return os.WriteFile(dstFile, upstreamData, 0644)
		}

		// File exists locally — compare content
		if string(localData) == string(upstreamData) {
			// Identical — no action needed
			return nil
		}

		// Different — check if locally modified (git tracks this)
		if isLocallyModified(dstFile) {
			kept++
			return nil // Keep local version
		}

		// Not locally modified (upstream changed) — take upstream
		updated++
		return os.WriteFile(dstFile, upstreamData, 0644)
	})
}

// isLocallyModified checks if a path has uncommitted local modifications via git.
func isLocallyModified(path string) bool {
	out, err := sh.Output("git", "diff", "--name-only", "HEAD", "--", path)
	if err != nil {
		return false // Can't determine — assume not modified
	}
	return strings.TrimSpace(out) != ""
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
	customSkillsPath = "src/internal/assets/skills"
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
		"src/internal/assets/skills/marketing/messaging.md",
		"src/internal/assets/skills/marketing/azure-marketplace.md",
		"src/internal/assets/skills/support/faq.md",
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

// ContributeSkills creates a branch in a local clone of the upstream repo
// with your local changes to ghcp4a-skills, ready for a PR.
// Set UPSTREAM_FORK to your fork URL (default: origin upstream).
func ContributeSkills() error {
	fmt.Println("🚀 Preparing upstream contribution...")

	// Find locally modified files in ghcp4a-skills
	out, err := sh.Output("git", "diff", "--name-only", "HEAD", "--", skillsTargetPath)
	if err != nil {
		return fmt.Errorf("failed to check local changes: %w", err)
	}

	// Also check staged changes
	stagedOut, err := sh.Output("git", "diff", "--name-only", "--cached", "--", skillsTargetPath)
	if err == nil && stagedOut != "" {
		if out != "" {
			out += "\n" + stagedOut
		} else {
			out = stagedOut
		}
	}

	if strings.TrimSpace(out) == "" {
		fmt.Println("No local changes found in ghcp4a-skills/. Nothing to contribute.")
		fmt.Println("Tip: Make changes to files in src/internal/assets/ghcp4a-skills/ first.")
		return nil
	}

	changedFiles := strings.Split(strings.TrimSpace(out), "\n")
	fmt.Printf("📝 Found %d changed file(s):\n", len(changedFiles))
	for _, f := range changedFiles {
		fmt.Printf("  • %s\n", f)
	}

	// Clone upstream repo
	fmt.Println("\n📥 Cloning upstream repo...")
	tempDir, err := os.MkdirTemp("", "contribute-skills-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	// Don't defer RemoveAll — user needs the clone to push from

	forkURL := os.Getenv("UPSTREAM_FORK")
	if forkURL == "" {
		forkURL = skillsSourceRepo
	}

	if err := sh.RunV("git", "clone", "--depth=1", forkURL, tempDir); err != nil {
		os.RemoveAll(tempDir)
		return fmt.Errorf("failed to clone upstream: %w", err)
	}

	// Create a branch
	branchName := fmt.Sprintf("contribute-skills-%s", time.Now().Format("20060102-150405"))
	if err := sh.RunV("git", "-C", tempDir, "checkout", "-b", branchName); err != nil {
		os.RemoveAll(tempDir)
		return fmt.Errorf("failed to create branch: %w", err)
	}

	// Copy changed files into the upstream clone
	copied := 0
	for _, localPath := range changedFiles {
		// localPath is like: src/internal/assets/ghcp4a-skills/azure-deploy/SKILL.md
		// upstream path is:  plugin/skills/azure-deploy/SKILL.md
		rel := strings.TrimPrefix(localPath, skillsTargetPath+"/")
		if rel == localPath {
			rel = strings.TrimPrefix(localPath, strings.ReplaceAll(skillsTargetPath, "\\", "/")+"/")
		}
		upstreamPath := filepath.Join(tempDir, skillsSourcePath, rel)

		// Read local file
		data, readErr := os.ReadFile(localPath)
		if readErr != nil {
			fmt.Printf("  ⚠️  %s: %v\n", localPath, readErr)
			continue
		}

		// Ensure parent dir exists
		if err := os.MkdirAll(filepath.Dir(upstreamPath), 0755); err != nil {
			fmt.Printf("  ⚠️  %s: %v\n", upstreamPath, err)
			continue
		}

		if err := os.WriteFile(upstreamPath, data, 0644); err != nil {
			fmt.Printf("  ❌ %s: %v\n", rel, err)
			continue
		}
		fmt.Printf("  ✅ %s\n", rel)
		copied++
	}

	if copied == 0 {
		os.RemoveAll(tempDir)
		return fmt.Errorf("no files were copied — nothing to contribute")
	}

	// Stage and commit
	if err := sh.RunV("git", "-C", tempDir, "add", "."); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}
	commitMsg := fmt.Sprintf("feat: contribute skill changes from azd-copilot (%d files)", copied)
	if err := sh.RunV("git", "-C", tempDir, "commit", "-m", commitMsg); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	fmt.Printf("\n✨ Branch '%s' ready at: %s\n", branchName, tempDir)
	fmt.Println("\nNext steps:")
	fmt.Printf("  1. cd %s\n", tempDir)
	fmt.Println("  2. git remote set-url origin <your-fork-url>  (if using upstream directly)")
	fmt.Println("  3. git push -u origin " + branchName)
	fmt.Println("  4. Open a PR at https://github.com/microsoft/GitHub-Copilot-for-Azure/pulls")
	fmt.Println("\nOr set UPSTREAM_FORK=<your-fork-url> to clone your fork directly.")

	return nil
}
