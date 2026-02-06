// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package spec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestConstants(t *testing.T) {
	if DocsDir != "docs" {
		t.Errorf("DocsDir = %q, want %q", DocsDir, "docs")
	}
	if SpecFile != "spec.md" {
		t.Errorf("SpecFile = %q, want %q", SpecFile, "spec.md")
	}
	if MetadataFile != ".copilot.json" {
		t.Errorf("MetadataFile = %q, want %q", MetadataFile, ".copilot.json")
	}
}

func TestDefaultMetadata(t *testing.T) {
	m := DefaultMetadata()

	if m == nil {
		t.Fatal("DefaultMetadata() returned nil")
	}

	expectedSpecFile := filepath.Join(DocsDir, SpecFile)
	if m.SpecFile != expectedSpecFile {
		t.Errorf("DefaultMetadata().SpecFile = %q, want %q", m.SpecFile, expectedSpecFile)
	}

	expectedCheckpointDir := filepath.Join(DocsDir, "checkpoints")
	if m.CheckpointDir != expectedCheckpointDir {
		t.Errorf("DefaultMetadata().CheckpointDir = %q, want %q", m.CheckpointDir, expectedCheckpointDir)
	}
}

func TestLoadMetadata_NotFound(t *testing.T) {
	// When metadata file doesn't exist, should return defaults
	m, err := LoadMetadata()
	if err != nil {
		t.Errorf("LoadMetadata() error = %v, want nil", err)
	}

	// Should return defaults
	defaults := DefaultMetadata()
	if m.SpecFile != defaults.SpecFile {
		t.Errorf("LoadMetadata() SpecFile = %q, want default %q", m.SpecFile, defaults.SpecFile)
	}
}

func TestGetSpecPath(t *testing.T) {
	path := GetSpecPath()

	// Should contain docs/spec.md or the configured path
	if path == "" {
		t.Error("GetSpecPath() returned empty string")
	}

	// Should end with spec.md by default
	if !strings.HasSuffix(path, SpecFile) {
		t.Logf("GetSpecPath() = %q (may have custom config)", path)
	}
}

func TestPath(t *testing.T) {
	// Path() should return the same as GetSpecPath()
	if Path() != GetSpecPath() {
		t.Errorf("Path() = %q, GetSpecPath() = %q, want same", Path(), GetSpecPath())
	}
}

func TestExists_NotFound(t *testing.T) {
	// In most test environments, spec file won't exist
	if Exists() {
		t.Log("Spec file exists in test environment")
	} else {
		t.Log("Spec file does not exist (expected in clean test environment)")
	}
}

func TestSpec_Fields(t *testing.T) {
	s := Spec{
		Name:        "TestApp",
		Description: "A test application",
		Mode:        "prototype",
		CreatedAt:   time.Now(),
		Architecture: `graph TB
			A[Client] --> B[API]
			B --> C[Database]`,
		Services: []Service{
			{Name: "api", Type: "api", Language: "go", Framework: "gin", Description: "REST API"},
		},
		AzureResources: []AzureResource{
			{Name: "app", Type: "App Service", SKU: "B1", Purpose: "Hosting", FreeTier: false, MonthlyCost: 13.14},
		},
		CostEstimate: CostEstimate{
			Monthly:     50.0,
			Yearly:      600.0,
			FreeTierMax: 0.0,
			Confidence:  "medium",
		},
		Goals:    []string{"Fast deployment", "Low cost"},
		NonGoals: []string{"High availability"},
	}

	if s.Name != "TestApp" {
		t.Errorf("Spec.Name = %q, want %q", s.Name, "TestApp")
	}
	if s.Mode != "prototype" {
		t.Errorf("Spec.Mode = %q, want %q", s.Mode, "prototype")
	}
	if len(s.Services) != 1 {
		t.Errorf("Spec.Services length = %d, want 1", len(s.Services))
	}
	if len(s.Goals) != 2 {
		t.Errorf("Spec.Goals length = %d, want 2", len(s.Goals))
	}
}

func TestService_Fields(t *testing.T) {
	svc := Service{
		Name:        "api",
		Type:        "api",
		Language:    "go",
		Framework:   "gin",
		Description: "REST API service",
	}

	if svc.Name != "api" {
		t.Errorf("Service.Name = %q, want %q", svc.Name, "api")
	}
	if svc.Language != "go" {
		t.Errorf("Service.Language = %q, want %q", svc.Language, "go")
	}
}

func TestAzureResource_Fields(t *testing.T) {
	res := AzureResource{
		Name:        "storage",
		Type:        "Storage Account",
		SKU:         "Standard_LRS",
		Purpose:     "File storage",
		FreeTier:    true,
		MonthlyCost: 0.0,
	}

	if res.Name != "storage" {
		t.Errorf("AzureResource.Name = %q, want %q", res.Name, "storage")
	}
	if !res.FreeTier {
		t.Error("AzureResource.FreeTier should be true")
	}
	if res.MonthlyCost != 0.0 {
		t.Errorf("AzureResource.MonthlyCost = %f, want 0.0", res.MonthlyCost)
	}
}

func TestCostEstimate_Fields(t *testing.T) {
	cost := CostEstimate{
		Monthly:     100.0,
		Yearly:      1200.0,
		FreeTierMax: 10.0,
		Confidence:  "high",
	}

	if cost.Monthly != 100.0 {
		t.Errorf("CostEstimate.Monthly = %f, want 100.0", cost.Monthly)
	}
	if cost.Confidence != "high" {
		t.Errorf("CostEstimate.Confidence = %q, want %q", cost.Confidence, "high")
	}
}

func TestGenerateMarkdown(t *testing.T) {
	s := &Spec{
		Name:        "MyApp",
		Description: "A sample application",
		Mode:        "production",
		CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Architecture: `graph TB
			A --> B`,
		Services: []Service{
			{Name: "api", Type: "api", Language: "go", Framework: "chi", Description: "Main API"},
		},
		AzureResources: []AzureResource{
			{Name: "app", Type: "App Service", SKU: "P1v3", Purpose: "API hosting", FreeTier: false, MonthlyCost: 100.0},
		},
		CostEstimate: CostEstimate{
			Monthly:    100.0,
			Yearly:     1200.0,
			Confidence: "high",
		},
		Goals:    []string{"Scalability"},
		NonGoals: []string{"Mobile support"},
	}

	md := GenerateMarkdown(s)

	// Check for required sections
	sections := []string{
		"# Application Specification",
		"## Overview",
		"**Name:** MyApp",
		"**Mode:** production",
		"## Goals",
		"## Non-Goals",
		"## Architecture",
		"```mermaid",
		"## Services",
		"## Azure Resources",
		"## Cost Estimate",
		"## Next Steps",
	}

	for _, section := range sections {
		if !strings.Contains(md, section) {
			t.Errorf("GenerateMarkdown() should contain %q", section)
		}
	}

	// Check service table
	if !strings.Contains(md, "| api |") {
		t.Error("GenerateMarkdown() should contain service 'api' in table")
	}

	// Check resource table
	if !strings.Contains(md, "| app |") {
		t.Error("GenerateMarkdown() should contain resource 'app' in table")
	}
}

func TestGenerateMarkdown_FreeTier(t *testing.T) {
	s := &Spec{
		Name:        "FreeApp",
		Description: "Uses free tier",
		Mode:        "prototype",
		CreatedAt:   time.Now(),
		AzureResources: []AzureResource{
			{Name: "app", Type: "App Service", SKU: "F1", Purpose: "Hosting", FreeTier: true, MonthlyCost: 0.0},
		},
		CostEstimate: CostEstimate{
			Monthly:     0.0,
			Yearly:      0.0,
			FreeTierMax: 0.0,
			Confidence:  "high",
		},
	}

	md := GenerateMarkdown(s)

	// Should show ✅ for free tier and "Free" for cost
	if !strings.Contains(md, "✅") {
		t.Error("GenerateMarkdown() should contain ✅ for free tier resource")
	}
	if !strings.Contains(md, "Free") {
		t.Error("GenerateMarkdown() should contain 'Free' for free tier cost")
	}
}

func TestGeneratePrompt(t *testing.T) {
	tests := []struct {
		name        string
		description string
		mode        string
		contains    []string
	}{
		{
			name:        "prototype mode",
			description: "build a todo app",
			mode:        "prototype",
			contains: []string{
				"build a todo app",
				"prototype",
				"Prototype Mode Guidelines",
				"free tiers",
				"Minimize complexity",
			},
		},
		{
			name:        "production mode",
			description: "enterprise CRM system",
			mode:        "production",
			contains: []string{
				"enterprise CRM system",
				"production",
				"Production Mode Guidelines",
				"monitoring and alerting",
				"Multiple environments",
				"Security hardening",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := GeneratePrompt(tt.description, tt.mode)

			for _, want := range tt.contains {
				if !strings.Contains(prompt, want) {
					t.Errorf("GeneratePrompt() should contain %q", want)
				}
			}
		})
	}
}

func TestWriteAndRead(t *testing.T) {
	// Create a temp directory for testing
	tmpDir := t.TempDir()

	// Change to temp directory for this test
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create docs directory
	if err := os.MkdirAll(DocsDir, 0755); err != nil {
		t.Fatalf("Failed to create docs directory: %v", err)
	}

	// Test Write
	content := "# Test Spec\n\nThis is a test specification."
	err = Write(content)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Verify file exists
	if !Exists() {
		t.Error("Exists() should return true after Write()")
	}

	// Test Read
	readContent, err := Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if readContent != content {
		t.Errorf("Read() = %q, want %q", readContent, content)
	}
}

func TestMetadata_Fields(t *testing.T) {
	m := Metadata{
		SpecFile:       "custom/spec.md",
		CheckpointDir:  "custom/checkpoints",
		GeneratedFiles: []string{"file1.go", "file2.go"},
	}

	if m.SpecFile != "custom/spec.md" {
		t.Errorf("Metadata.SpecFile = %q, want %q", m.SpecFile, "custom/spec.md")
	}
	if len(m.GeneratedFiles) != 2 {
		t.Errorf("Metadata.GeneratedFiles length = %d, want 2", len(m.GeneratedFiles))
	}
}
