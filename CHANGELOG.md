## [0.1.5] - 2026-02-15

- Fix gosec G703: add #nosec annotations for known install paths (b1edc69)
- Update extension source to jongio.github.io/azd-extensions registry (f437511)
- Update magefile comment to remove 'enables extensions' reference (c186a80)
- Fix gosec G703: sanitize paths with filepath.Clean() (36e04c4)
- Remove extension preview/alpha requirement (99a0b73)
- docs: update installation steps for GitHub Copilot CLI and Azure Developer CLI in README and getting-started page (005d11f)
- chore: update registry for v0.1.4 (b4ff0de)

## [0.1.4] - 2026-02-11

- chore: update registry for v0.1.3 (ecf12da)

## [0.1.3] - 2026-02-11

- feat: update version in extension.yaml during release process and improve README formatting (bb5684f)
- chore: update registry for v0.1.4 (9bcce92)
- docs: update README with installation instructions for Azure Developer CLI and extension (7b0780f)
- Refactor code structure for improved readability and maintainability (d201bfb)
- fix: add GH_TOKEN to environment variables for GitHub actions (5b2fc51)
- feat: update version in extension.yaml and add debug steps for build artifacts (dacacee)
- feat: add release workflow for building and publishing azd extensions (1433a19)

## [0.1.2] - 2026-02-11

- fix: update working directory for build, package, and release steps in release workflow (0ca8c76)

## [0.1.1] - 2026-02-10

- feat: add notification step for azd-extensions on successful release (2ac86e9)
- fix: simplify command syntax for building a todo API in getting started guide (43e4927)
- fix: enhance hero description for Azure Copilot CLI experience (01e685d)
- fix: update references to Azure Copilot to Azure Copilot CLI for consistency (0026ea5)
- fix: specify directory for golangci-lint to improve linting accuracy (7f9e6ba)
- fix: update gosec command exclusions and adjust working directory for govulncheck (feadd00)
- fix: update Go version to 1.25.7 across CI workflows and module files (cc5cdfb)
- fix: update vulnerability check to use correct working directory and adjust gosec command exclusions (32da529)
- fix: update gosec command to exclude specific warnings for improved security scanning (f9b8be5)
- fix: update golangci-lint command to include specific directory and handle tty closure safely (f291959)
- fix: update consoleHandles struct to include conin and conout fields for better console management (457f3b0)
- fix: update golangci-lint installation path in CI and release workflows; remove local replacement in go.mod (ffe7af4)
- fix: update copyright notice in LICENSE and enhance validation steps in SKILL.md and README.md (1c81c42)
- Refactor Copilot CLI to install agents and skills in ~/.azd/copilot/ directory; update related documentation and tests (0b916cf)
- feat: Update Azure Storage documentation and SDK usage examples (395fecc)
- feat: enhance SyncSkills command with local and custom repo options in README and magefile (99ef131)
- feat: refactor Build and Install functions to utilize azd x commands and ensure extension setup (ffdfc79)
- Enhance README and web pages for azd copilot (b138bf9)
- feat: update CLI reference styles and improve command discovery handling (a47870f)
- feat: implement runWithRetry function for command execution with retry logic (d4f5a4e)
- feat: add comprehensive README for Scenario Runner with usage instructions and YAML format details (e6ac581)
- feat: update skill count to 28 in documentation and UI (df5142a)
- feat: enhance scenario verification and export functionality - Added verification steps to scenarios for improved testing - Implemented JSON export and import for scenario results - Updated database schema to support verification results - Refactored scenario execution to include verification process (1eef399)
- feat: add scenario automation tools for azd-copilot (f369d60)
- feat(azure-manager): enforce mandatory delegation for standard apps and clarify escalation rules feat(avm-bicep-rules): add common pitfalls for Container App + ACR authentication feat(container-app-acr-auth): introduce Bicep patterns for Container App ACR authentication (15170f6)
- feat(contribute-skills): update upstream repo handling and improve contribution instructions (94f424c)
- Add support skills documentation and error message templates (42eaf71)
- feat(azure-manager): improve Free SKU selection logic and error handling for SWA (8b43bb3)
- feat(azure-manager): enhance agent delegation process and clarify complexity classification (5b8bfea)
- feat(azure-manager): update deployment guidelines and add SKU selection preferences feat(avm-bicep-rules): enhance module discovery instructions and add reference file (5fabfc7)
- feat(azure-manager): enhance complexity classification and deployment guidelines (6a91464)
- feat(playwright): add Playwright E2E testing patterns and integration (1280bd7)
- feat(console): implement console handle management for Windows feat(agents): add package manager and project essentials to agent documentation feat(banner): update banner with additional build information (1ad8732)
- feat(logging): add logger package for structured logging (63d031c)
- Add cli and web (4108a8a)
- Init (d4320fa)

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-02-04

### Added
- Initial release
- Basic extension structure with version command
- Listen command for azd extension framework integration
