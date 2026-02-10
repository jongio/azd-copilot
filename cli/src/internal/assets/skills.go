// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package assets

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/jongio/azd-core/fileutil"
	"gopkg.in/yaml.v3"
)

// Upstream skills synced from microsoft/GitHub-Copilot-for-Azure.
// Editable — changes survive `mage SyncSkills` (smart merge) and can be
// contributed back via `mage ContributeSkills`.
//
//go:embed ghcp4a-skills/*/SKILL.md ghcp4a-skills/*/*/*.md ghcp4a-skills/*/*/*/*.md ghcp4a-skills/*/*/*/*/*.md
var embeddedSkills embed.FS

// Skills maintained in this repo (azd-copilot custom skills).
//
//go:embed skills/*/SKILL.md skills/*/*.md
var embeddedCustomSkills embed.FS

// SkillInfo contains metadata about a skill
type SkillInfo struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Path        string
}

// allSkillSources returns all embedded skill filesystems with their root prefix.
func allSkillSources() []struct {
	fs     embed.FS
	prefix string
} {
	return []struct {
		fs     embed.FS
		prefix string
	}{
		{embeddedSkills, "ghcp4a-skills"},
		{embeddedCustomSkills, "skills"},
	}
}

// InstallSkills extracts embedded skills to ~/.azd/copilot/skills/
func InstallSkills() (string, int, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", 0, fmt.Errorf("failed to get home directory: %w", err)
	}

	destDir := filepath.Join(home, ".azd", "copilot", "skills")
	if err := fileutil.EnsureDir(destDir); err != nil {
		return "", 0, fmt.Errorf("failed to create skills directory: %w", err)
	}

	installed := 0
	for _, src := range allSkillSources() {
		err = fs.WalkDir(src.fs, src.prefix, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return nil
			}
			if d.IsDir() {
				return nil
			}

			data, readErr := src.fs.ReadFile(path)
			if readErr != nil {
				return nil
			}

			// Strip prefix for destination (both map to the same output dir)
			relPath := strings.TrimPrefix(path, src.prefix+"/")
			destPath := filepath.Join(destDir, relPath)

			// Create parent directories
			if err := fileutil.EnsureDir(filepath.Dir(destPath)); err != nil {
				return nil
			}

			// Write file atomically
			if err := fileutil.AtomicWriteFile(destPath, data, 0644); err != nil {
				return nil
			}

			// Count top-level skill directories
			parts := strings.Split(relPath, "/")
			if strings.HasSuffix(relPath, "SKILL.md") && len(parts) == 2 {
				installed++
			}

			return nil
		})
		if err != nil {
			return destDir, installed, err
		}
	}

	return destDir, installed, nil
}

// ListSkills returns information about all embedded skills
func ListSkills() ([]SkillInfo, error) {
	var skills []SkillInfo

	for _, src := range allSkillSources() {
		entries, err := fs.ReadDir(src.fs, src.prefix)
		if err != nil {
			return nil, fmt.Errorf("failed to read embedded skills from %s: %w", src.prefix, err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			skillPath := src.prefix + "/" + entry.Name() + "/SKILL.md"
			data, err := src.fs.ReadFile(skillPath)
			if err != nil {
				continue
			}

			skill := parseSkillInfo(entry.Name(), data)
			skills = append(skills, skill)
		}
	}

	return skills, nil
}

// GetSkill returns information about a specific skill
func GetSkill(name string) (*SkillInfo, error) {
	skills, err := ListSkills()
	if err != nil {
		return nil, err
	}

	for _, skill := range skills {
		if skill.Name == name {
			return &skill, nil
		}
	}

	return nil, fmt.Errorf("skill not found: %s", name)
}

// SkillCount returns the number of available skills.
func SkillCount() int {
	skills, err := ListSkills()
	if err != nil {
		return 0
	}
	return len(skills)
}

// parseSkillInfo extracts skill info from markdown with YAML frontmatter
func parseSkillInfo(dirName string, data []byte) SkillInfo {
	content := string(data)
	skill := SkillInfo{
		Name: dirName,
		Path: dirName,
	}

	// Parse YAML frontmatter
	if strings.HasPrefix(content, "---") {
		parts := strings.SplitN(content, "---", 3)
		if len(parts) >= 3 {
			var frontmatter struct {
				Name        string `yaml:"name"`
				Description string `yaml:"description"`
			}
			if err := yaml.Unmarshal([]byte(parts[1]), &frontmatter); err == nil {
				if frontmatter.Name != "" {
					skill.Name = frontmatter.Name
				}
				skill.Description = frontmatter.Description
			}
		}
	}

	// If no description from frontmatter, try to extract from content
	if skill.Description == "" {
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "---") && !strings.HasPrefix(line, ">") {
				skill.Description = line
				break
			}
		}
	}

	return skill
}
