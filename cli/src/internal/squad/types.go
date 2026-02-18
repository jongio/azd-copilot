// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package squad

// InitOptions configures team initialization.
type InitOptions struct {
	ProjectName string
	TechStack   string
	UserName    string
	UserEmail   string
}

// Member represents a squad team member.
type Member struct {
	Name        string
	Role        string
	Emoji       string
	CharterPath string
	Status      string // "active", "silent", "monitor"
}

// Decision represents a team decision.
type Decision struct {
	Author  string
	Summary string
	Detail  string
	Slug    string
}

// AzureRole defines a pre-configured Azure squad role.
type AzureRole struct {
	Role      string
	Emoji     string
	Expertise string
	Owns      []string
	Skills    []string // Azure skills this role references
}

// DefaultAzureRoles returns the pre-configured Azure squad roles.
func DefaultAzureRoles() []AzureRole {
	return []AzureRole{
		{
			Role:      "Azure Architect",
			Emoji:     "🏗️",
			Expertise: "Bicep/AVM modules, Container Apps, App Service, networking, identity",
			Owns:      []string{"Bicep/IaC", "azure.yaml", "networking", "identity", "resource naming"},
			Skills:    []string{"avm-bicep-rules", "azure-prepare", "azure-validate", "container-app-acr-auth"},
		},
		{
			Role:      "Azure Developer",
			Emoji:     "💻",
			Expertise: "Backend APIs, frontend frameworks, Azure SDKs, containerization",
			Owns:      []string{"application code", "API endpoints", "Dockerfiles", "SDK integration"},
			Skills:    []string{"azure-functions", "azure-ai"},
		},
		{
			Role:      "Azure Data Engineer",
			Emoji:     "🗄️",
			Expertise: "PostgreSQL, Cosmos DB, schema design, migrations, query optimization",
			Owns:      []string{"database schemas", "migrations", "queries", "data modeling"},
			Skills:    []string{"azure-postgres", "azure-kusto"},
		},
		{
			Role:      "Azure Security",
			Emoji:     "🔒",
			Expertise: "RBAC, managed identity, vulnerability scanning, compliance, Key Vault",
			Owns:      []string{"RBAC", "identity", "Key Vault", "network security", "compliance"},
			Skills:    []string{"azure-compliance", "azure-role-selector"},
		},
		{
			Role:      "Azure DevOps",
			Emoji:     "🚀",
			Expertise: "GitHub Actions, App Insights, Azure Monitor, deployment strategies",
			Owns:      []string{"CI/CD pipelines", "monitoring", "alerts", "container registry"},
			Skills:    []string{"appinsights-instrumentation", "azure-observability"},
		},
		{
			Role:      "Azure Quality",
			Emoji:     "✅",
			Expertise: "Unit testing, integration testing, code review, test automation",
			Owns:      []string{"test suites", "code review", "quality gates", "coverage analysis"},
			Skills:    []string{},
		},
		{
			Role:      "Azure AI/ML Engineer",
			Emoji:     "🤖",
			Expertise: "Azure OpenAI, AI Search, Foundry, RAG, agent frameworks, model deployment",
			Owns:      []string{"AI services", "search indexes", "prompt flows", "embeddings", "model selection"},
			Skills:    []string{"azure-ai"},
		},
		{
			Role:      "Azure Analytics",
			Emoji:     "📊",
			Expertise: "Usage analytics, dashboards, metrics design, KQL, reporting",
			Owns:      []string{"dashboards", "metrics", "KQL queries", "workbooks", "reporting"},
			Skills:    []string{"analytics", "azure-observability"},
		},
		{
			Role:      "Azure Compliance",
			Emoji:     "📋",
			Expertise: "GDPR, SOC2, HIPAA assessment, gap analysis, remediation guidance",
			Owns:      []string{"compliance frameworks", "policy assessment", "audit reports", "remediation"},
			Skills:    []string{"azure-compliance"},
		},
		{
			Role:      "Azure UX/Accessibility",
			Emoji:     "🎨",
			Expertise: "WCAG compliance, accessibility audits, UI review",
			Owns:      []string{"accessibility", "UI review", "WCAG compliance", "design standards"},
			Skills:    []string{},
		},
		{
			Role:      "Azure Technical Writer",
			Emoji:     "📝",
			Expertise: "README, API documentation, ADRs, runbooks, code documentation",
			Owns:      []string{"README", "API docs", "ADRs", "runbooks", "onboarding guides"},
			Skills:    []string{},
		},
		{
			Role:      "Azure FinOps",
			Emoji:     "💰",
			Expertise: "Cost estimation, optimization, waste identification, TCO analysis",
			Owns:      []string{"cost estimates", "pricing analysis", "optimization", "TCO reports"},
			Skills:    []string{"azure-cost-optimization"},
		},
		{
			Role:      "Azure Marketing",
			Emoji:     "📣",
			Expertise: "Positioning, landing pages, feature communication, competitive analysis",
			Owns:      []string{"landing pages", "feature announcements", "positioning", "competitive analysis"},
			Skills:    []string{},
		},
		{
			Role:      "Azure Product Manager",
			Emoji:     "🎯",
			Expertise: "User needs, specs, requirements, acceptance criteria",
			Owns:      []string{"requirements", "acceptance criteria", "user stories", "feature specs"},
			Skills:    []string{},
		},
		{
			Role:      "Azure Customer Success",
			Emoji:     "🛟",
			Expertise: "Troubleshooting, FAQ generation, error messages, onboarding",
			Owns:      []string{"troubleshooting guides", "FAQs", "error messages", "onboarding"},
			Skills:    []string{"azure-diagnostics"},
		},
		{
			Role:      "Session Logger",
			Emoji:     "📋",
			Expertise: "Session logging, decision merging, context propagation",
			Owns:      []string{"session logs", "decisions.md", "decision inbox"},
			Skills:    []string{},
		},
	}
}
