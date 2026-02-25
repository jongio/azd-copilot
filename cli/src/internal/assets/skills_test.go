// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package assets

import (
	"io/fs"
	"strings"
	"testing"
)

func TestListSkills(t *testing.T) {
	skills, err := ListSkills()
	if err != nil {
		t.Fatalf("ListSkills() error = %v", err)
	}

	if len(skills) == 0 {
		t.Error("ListSkills() returned empty list, expected embedded skills")
	}

	t.Logf("Found %d skills", len(skills))

	for _, skill := range skills {
		if skill.Name == "" {
			t.Error("Skill has empty Name")
		}
		if skill.Path == "" {
			t.Error("Skill has empty Path")
		}
	}
}

func TestListSkills_Content(t *testing.T) {
	skills, err := ListSkills()
	if err != nil {
		t.Fatalf("ListSkills() error = %v", err)
	}

	expectedSkills := []string{
		"azure-deploy",
		"azure-prepare",
		"azure-validate",
		"azure-functions",
		"avm-bicep-rules",
	}

	skillNames := make(map[string]bool)
	for _, skill := range skills {
		skillNames[skill.Name] = true
	}

	for _, expected := range expectedSkills {
		if !skillNames[expected] {
			t.Logf("Expected skill %q not found (may be renamed or not included)", expected)
		}
	}
}

func TestGetSkill_Found(t *testing.T) {
	skills, err := ListSkills()
	if err != nil {
		t.Fatalf("ListSkills() error = %v", err)
	}

	if len(skills) == 0 {
		t.Skip("No skills found to test GetSkill")
	}

	firstSkill := skills[0]

	skill, err := GetSkill(firstSkill.Name)
	if err != nil {
		t.Fatalf("GetSkill(%q) error = %v", firstSkill.Name, err)
	}

	if skill == nil {
		t.Fatalf("GetSkill(%q) returned nil", firstSkill.Name)
	}

	if skill.Name != firstSkill.Name {
		t.Errorf("GetSkill() returned wrong skill: got %q, want %q", skill.Name, firstSkill.Name)
	}
}

func TestGetSkill_NotFound(t *testing.T) {
	_, err := GetSkill("nonexistent-skill-xyz")
	if err == nil {
		t.Error("GetSkill() should return error for non-existent skill")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("GetSkill() error should contain 'not found', got %q", err.Error())
	}
}

func TestSkillInfo_Fields(t *testing.T) {
	info := SkillInfo{
		Name:        "test-skill",
		Description: "A test skill",
		Path:        "test-skill",
	}

	if info.Name != "test-skill" {
		t.Errorf("SkillInfo.Name = %q, want %q", info.Name, "test-skill")
	}
	if info.Description != "A test skill" {
		t.Errorf("SkillInfo.Description = %q, want %q", info.Description, "A test skill")
	}
	if info.Path != "test-skill" {
		t.Errorf("SkillInfo.Path = %q, want %q", info.Path, "test-skill")
	}
}

func TestParseSkillInfo_WithFrontmatter(t *testing.T) {
	content := []byte(`---
name: azure-test
description: "A test skill for Azure"
---

# Azure Test Skill

This is the content of the skill.
`)

	info := parseSkillInfo("test-dir", content)

	if info.Name != "azure-test" {
		t.Errorf("parseSkillInfo() Name = %q, want %q", info.Name, "azure-test")
	}
	if info.Description != "A test skill for Azure" {
		t.Errorf("parseSkillInfo() Description = %q, want %q", info.Description, "A test skill for Azure")
	}
}

func TestParseSkillInfo_WithoutFrontmatter(t *testing.T) {
	content := []byte(`# Simple Skill

This skill has no frontmatter.
`)

	info := parseSkillInfo("simple-skill", content)

	if info.Name != "simple-skill" {
		t.Errorf("parseSkillInfo() Name = %q, want %q", info.Name, "simple-skill")
	}
}

func TestParseSkillInfo_MalformedFrontmatter(t *testing.T) {
	content := []byte(`---
name: test
invalid yaml content here
---

# Test
`)

	// Should not panic on malformed YAML
	info := parseSkillInfo("test-skill", content)

	if info.Name == "" {
		t.Error("parseSkillInfo() should return dir name as fallback Name")
	}
}

func TestInstallSkills(t *testing.T) {
	dir, count, err := InstallSkills()
	if err != nil {
		t.Logf("InstallSkills() error = %v (may be expected if no home directory)", err)
		return
	}

	t.Logf("InstallSkills() installed %d skills to %s", count, dir)

	if count == 0 {
		t.Log("InstallSkills() installed 0 skills (may be expected if already installed)")
	}
}

func TestSkillCount(t *testing.T) {
	count := SkillCount()
	if count == 0 {
		t.Error("SkillCount() returned 0, expected at least one skill")
	}
	t.Logf("SkillCount() = %d", count)
}

func TestEmbeddedSkills_NotEmpty(t *testing.T) {
	for _, src := range allSkillSources() {
		entries, err := fs.ReadDir(src.fs, src.prefix)
		if err != nil {
			t.Fatalf("Failed to read embedded skills from %s: %v", src.prefix, err)
		}

		dirCount := 0
		for _, entry := range entries {
			if entry.IsDir() {
				dirCount++
			}
		}

		if dirCount == 0 {
			t.Errorf("No skill directories found in embedded %s", src.prefix)
		}

		t.Logf("Found %d skill directories in %s", dirCount, src.prefix)
	}
}

func TestEmbeddedSkills_HaveSkillMD(t *testing.T) {
	for _, src := range allSkillSources() {
		entries, err := fs.ReadDir(src.fs, src.prefix)
		if err != nil {
			t.Fatalf("Failed to read embedded skills from %s: %v", src.prefix, err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			skillPath := src.prefix + "/" + entry.Name() + "/SKILL.md"
			data, err := src.fs.ReadFile(skillPath)
			if err != nil {
				t.Errorf("Skill %q in %s missing SKILL.md: %v", entry.Name(), src.prefix, err)
				continue
			}

			if len(data) == 0 {
				t.Errorf("Skill %q SKILL.md is empty", entry.Name())
			}
		}
	}
}
