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

// Build builds the CLI binary and installs it locally using azd x build.
func Build() error {
	_ = killExtensionProcesses()
	time.Sleep(500 * time.Millisecond)

	// Ensure azd extensions are set up (installs azd x if needed)
	if err := ensureAzdExtensions(); err != nil {
		return err
	}

	version, err := getVersion()
	if err != nil {
		return err
	}

	fmt.Println("Building and installing extension...")

	env := map[string]string{
		"EXTENSION_ID":      extensionID,
		"EXTENSION_VERSION": version,
	}

	// Build and install directly using azd x build
	if err := runWithEnvRetry(env, "azd", "x", "build"); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Printf("✅ Build complete! Version: %s\n", version)
	fmt.Println("   Run 'azd copilot version' to verify")
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

// Watch monitors files and rebuilds on changes (requires azd x watch).
func Watch() error {
	// Ensure azd extensions are set up
	if err := ensureAzdExtensions(); err != nil {
		return err
	}

	fmt.Println("Starting watch mode...")

	env := map[string]string{
		"EXTENSION_ID": extensionID,
	}

	return sh.RunWithV(env, "azd", "x", "watch")
}

// ensureAzdExtensions checks that azd is installed and the azd x extension is installed.
// This is a prerequisite for commands that use azd x (build, watch, etc.).
func ensureAzdExtensions() error {
	// Check if azd is available
	if _, err := sh.Output("azd", "version"); err != nil {
		return fmt.Errorf("azd is not installed or not in PATH. Install from https://aka.ms/azd")
	}

	// Check if azd x extension is available
	if _, err := sh.Output("azd", "x", "--help"); err != nil {
		fmt.Println("📦 Installing azd x extension (developer kit)...")
		if err := sh.RunV("azd", "extension", "install", "microsoft.azd.extensions", "--source", "azd", "--no-prompt"); err != nil {
			return fmt.Errorf("failed to install azd x extension: %w", err)
		}
		fmt.Println("✅ azd x extension installed!")
	}

	return nil
}

const (
	skillsSourceRepo = "https://github.com/microsoft/GitHub-Copilot-for-Azure.git"
	skillsSourcePath = "plugin/skills"
	skillsTargetPath = "src/internal/assets/ghcp4a-skills"
)

// SyncSkills syncs upstream skills from microsoft/GitHub-Copilot-for-Azure (main branch).
//
// For custom sources, use SyncSkillsFrom instead.
func SyncSkills() error {
	return SyncSkillsFrom("")
}

// SyncSkillsFrom syncs upstream skills from a custom source.
//
// When syncing from a local path, an exact sync is performed: the target is
// mirrored to match the source, including file deletions and content updates.
//
// When syncing from a remote repo, a smart merge is used: new files are added,
// deleted files are removed, but locally modified files are preserved.
//
// The source parameter controls where to sync from:
//
//	(empty)          — clone upstream main (default)
//	local path       — exact sync from a local clone (auto-detected)
//	repo URL         — smart merge from a custom repo (main branch)
//	repo URL@branch  — smart merge from a custom repo at a specific branch
//
// Examples:
//
//	mage SyncSkillsFrom C:\code\GitHub-Copilot-for-Azure                       # local folder
//	mage SyncSkillsFrom https://github.com/user/fork.git                       # custom repo
//	mage SyncSkillsFrom https://github.com/user/fork.git@my-branch             # custom repo + branch
func SyncSkillsFrom(source string) error {
	fmt.Println("🔄 Syncing upstream Azure skills...")

	var sourceDir string
	var tempDir string

	if source == "" {
		// Default: clone upstream main
		source = skillsSourceRepo
	}

	// Detect if source is a local path or a repo URL
	isLocal := isLocalPath(source)
	if isLocal {
		sourceDir = source
		skillsDir := filepath.Join(sourceDir, skillsSourcePath)
		if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
			return fmt.Errorf("skills not found at %s", skillsDir)
		}
		fmt.Printf("📂 Using local source: %s (exact sync)\n", sourceDir)
	} else {
		// Parse optional @branch suffix
		repo, branch := parseRepoSource(source)

		fmt.Printf("📥 Cloning %s (branch: %s, sparse)...\n", repo, branch)
		var err error
		tempDir, err = os.MkdirTemp("", "skills-sync-*")
		if err != nil {
			return fmt.Errorf("failed to create temp dir: %w", err)
		}
		defer os.RemoveAll(tempDir)

		sourceDir = tempDir
		cmds := [][]string{
			{"git", "clone", "--depth=1", "--branch", branch, "--filter=blob:none", "--sparse", repo, tempDir},
			{"git", "-C", tempDir, "sparse-checkout", "set", skillsSourcePath},
		}
		for _, args := range cmds {
			if err := runWithRetry(args[0], args[1:]...); err != nil {
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
			if !isLocal && isLocallyModified(dst) {
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
			fileAdded, fileUpdated, fileKept, fileRemoved, err := mergeSkillDir(src, dst, isLocal)
			if err != nil {
				fmt.Printf("  ❌ %s: %v\n", name, err)
				continue
			}
			if fileAdded+fileUpdated+fileRemoved > 0 {
				fmt.Printf("  ✅ %s (merged: %d new, %d updated, %d removed, %d kept)\n", name, fileAdded, fileUpdated, fileRemoved, fileKept)
			} else if fileKept > 0 {
				fmt.Printf("  🔒 %s (all %d files locally modified — kept)\n", name, fileKept)
			} else {
				fmt.Printf("  ✅ %s (unchanged)\n", name)
			}
			added += fileAdded
			updated += fileUpdated
			kept += fileKept
			removed += fileRemoved
		}
	}

	fmt.Printf("\n✨ Sync complete: %d added, %d updated, %d locally modified (kept), %d removed\n",
		added, updated, kept, removed)

	// Update counts in static files
	return UpdateCounts()
}

// runWithRetry runs a command with up to 3 retries on failure.
func runWithRetry(cmd string, args ...string) error {
	const maxRetries = 3
	var err error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			delay := time.Duration(i*5) * time.Second
			fmt.Printf("  ⚠️  Attempt %d/%d failed, retrying in %s...\n", i, maxRetries, delay)
			time.Sleep(delay)
		}
		if err = sh.RunV(cmd, args...); err == nil {
			return nil
		}
	}
	return err
}

// runWithEnvRetry runs a command with environment variables, retrying up to 3 times on failure.
func runWithEnvRetry(env map[string]string, cmd string, args ...string) error {
	const maxRetries = 3
	var err error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			delay := time.Duration(i*5) * time.Second
			fmt.Printf("  ⚠️  Attempt %d/%d failed, retrying in %s...\n", i, maxRetries, delay)
			time.Sleep(delay)
		}
		if err = sh.RunWithV(env, cmd, args...); err == nil {
			return nil
		}
	}
	return err
}

// mergeSkillDir merges an upstream skill directory into a local one.
// When exactSync is true (local source), all upstream changes are applied and
// deleted files are removed without checking for local modifications.
// When exactSync is false (remote source), locally modified files are preserved.
func mergeSkillDir(src, dst string, exactSync bool) (added, updated, kept, removed int, err error) {
	// Build set of all files in upstream source (relative paths)
	upstreamFiles := make(map[string]bool)
	err = filepath.Walk(src, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return walkErr
		}
		rel, _ := filepath.Rel(src, path)
		upstreamFiles[rel] = true
		return nil
	})
	if err != nil {
		return
	}

	// Remove local files that no longer exist upstream
	err = filepath.Walk(dst, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return walkErr
		}
		rel, _ := filepath.Rel(dst, path)
		if upstreamFiles[rel] {
			return nil // File still exists upstream
		}
		if !exactSync && isLocallyModified(path) {
			kept++
			return nil
		}
		fmt.Printf("    🗑️  %s (deleted upstream)\n", rel)
		os.Remove(path)
		removed++
		return nil
	})
	if err != nil {
		return
	}

	// Sync upstream files into local
	err = filepath.Walk(src, func(path string, info os.FileInfo, walkErr error) error {
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
			return nil
		}

		// Different — check if locally modified (git tracks this)
		if !exactSync && isLocallyModified(dstFile) {
			kept++
			return nil
		}

		// Not locally modified (upstream changed) — take upstream
		updated++
		return os.WriteFile(dstFile, upstreamData, 0644)
	})

	// Clean up empty directories left after file removals
	if removed > 0 {
		removeEmptyDirs(dst)
	}

	return added, updated, kept, removed, err
}

// isLocallyModified checks if a path has uncommitted local modifications via git.
func isLocallyModified(path string) bool {
	out, err := sh.Output("git", "diff", "--name-only", "HEAD", "--", path)
	if err != nil {
		return false // Can't determine — assume not modified
	}
	return strings.TrimSpace(out) != ""
}

// isLocalPath returns true if source looks like a local filesystem path
// rather than a git repo URL.
func isLocalPath(source string) bool {
	// URLs start with a scheme or git@ notation
	if strings.HasPrefix(source, "https://") ||
		strings.HasPrefix(source, "http://") ||
		strings.HasPrefix(source, "git@") ||
		strings.HasPrefix(source, "ssh://") {
		return false
	}
	// Check if the path actually exists on disk
	_, err := os.Stat(source)
	return err == nil
}

// parseRepoSource splits a source string into repo URL and branch.
// Supports "repo@branch" syntax; defaults to "main" if no branch specified.
func parseRepoSource(source string) (repo, branch string) {
	// Split on last @ that comes after the scheme (to avoid splitting user@host)
	// Look for @ after ".git" or after the path portion
	repo = source
	branch = "main"

	// Find @ that's not part of the scheme (git@...)
	// We look for @ after "github.com" or similar host portion
	idx := strings.LastIndex(source, "@")
	if idx > 0 {
		// Make sure the @ isn't part of git@github.com style prefix
		beforeAt := source[:idx]
		if strings.Contains(beforeAt, "/") {
			repo = beforeAt
			branch = source[idx+1:]
		}
	}
	return repo, branch
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

// removeEmptyDirs removes empty directories within root (bottom-up).
func removeEmptyDirs(root string) {
	// Walk bottom-up by collecting dirs first, then checking in reverse
	var dirs []string
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() && path != root {
			dirs = append(dirs, path)
		}
		return nil
	})
	// Remove in reverse order (deepest first)
	for i := len(dirs) - 1; i >= 0; i-- {
		entries, err := os.ReadDir(dirs[i])
		if err == nil && len(entries) == 0 {
			os.Remove(dirs[i])
		}
	}
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

// ContributeSkills creates a branch in the upstream repo clone at
// c:\code\github-copilot-for-azure (or GHCP4A_REPO env var) with your local
// changes to ghcp4a-skills, ready for a PR.
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

	// Use local upstream repo clone
	upstreamRepo := os.Getenv("GHCP4A_REPO")
	if upstreamRepo == "" {
		upstreamRepo = `c:\code\github-copilot-for-azure`
	}

	if _, err := os.Stat(filepath.Join(upstreamRepo, ".git")); os.IsNotExist(err) {
		return fmt.Errorf("upstream repo not found at %s — set GHCP4A_REPO env var to the correct path", upstreamRepo)
	}
	fmt.Printf("📂 Using upstream repo: %s\n", upstreamRepo)

	// Create a branch
	branchName := fmt.Sprintf("contribute-skills-%s", time.Now().Format("20060102-150405"))
	if err := sh.RunV("git", "-C", upstreamRepo, "checkout", "-b", branchName); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	// Copy changed files into the upstream repo
	copied := 0
	for _, localPath := range changedFiles {
		// localPath is like: src/internal/assets/ghcp4a-skills/azure-deploy/SKILL.md
		// upstream path is:  plugin/skills/azure-deploy/SKILL.md
		rel := strings.TrimPrefix(localPath, skillsTargetPath+"/")
		if rel == localPath {
			rel = strings.TrimPrefix(localPath, strings.ReplaceAll(skillsTargetPath, "\\", "/")+"/")
		}
		upstreamPath := filepath.Join(upstreamRepo, skillsSourcePath, rel)

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
		return fmt.Errorf("no files were copied — nothing to contribute")
	}

	// Stage and commit
	if err := sh.RunV("git", "-C", upstreamRepo, "add", "."); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}
	commitMsg := fmt.Sprintf("feat: contribute skill changes from azd-copilot (%d files)", copied)
	if err := sh.RunV("git", "-C", upstreamRepo, "commit", "-m", commitMsg); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	fmt.Printf("\n✨ Branch '%s' ready in: %s\n", branchName, upstreamRepo)
	fmt.Println("\nNext steps:")
	fmt.Printf("  cd %s\n", upstreamRepo)
	fmt.Println("  git push -u origin " + branchName)
	fmt.Println("  # Then open a PR at https://github.com/microsoft/GitHub-Copilot-for-Azure/pulls")

	return nil
}
