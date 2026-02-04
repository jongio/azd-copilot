# Contributing to azd-copilot

Thank you for your interest in contributing to azd-copilot! This document provides guidelines for contributing to the project.

## Getting Started

### Prerequisites

Before contributing, ensure you have the following installed:

- **Go**: 1.25 or later
- **Node.js**: 20.0.0 or later
- **pnpm**: 9.0.0 or later
- **Azure Developer CLI (azd)**: Latest version

You can verify your versions:
```bash
go version          # Should be 1.25+
node --version      # Should be v20.0.0+
pnpm --version      # Should be 9.0.0+
azd version         # Should be latest
```

### Setup

1. **Fork the repository** and clone your fork

2. **Install Go dependencies**:
   ```bash
   go mod download
   ```

3. **Build the extension**:
   ```bash
   go build -o bin/copilot.exe ./src/cmd/copilot
   ```

4. **Install locally for testing**:
   ```bash
   # Add the local registry source (one-time)
   azd extension source add -n copilot -t file -l "<path-to-repo>/registry.json"
   
   # Install the extension
   azd extension install jongio.azd.copilot --source copilot --force
   
   # Verify installation
   azd copilot version
   ```

## Development Workflow

### 1. Create a Branch
```bash
git checkout -b feature/your-feature-name
```

### 2. Make Changes
- Follow Go code conventions
- Run `go fmt ./...` to format your code
- Add tests for new functionality
- Update documentation as needed

### 3. Test Your Changes
```bash
# Build
go build -o bin/copilot.exe ./src/cmd/copilot

# Run tests
go test ./...

# Run with coverage
go test -cover ./...
```

### 4. Commit Your Changes
```bash
git add .
git commit -m "feat: add support for X"
```

Follow [Conventional Commits](https://www.conventionalcommits.org/):
- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `test:` Adding or updating tests
- `refactor:` Code refactoring
- `chore:` Maintenance tasks

### 5. Push and Create Pull Request
```bash
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub.

## Code Guidelines

### Go Style
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` for formatting
- Run `golangci-lint run` before committing
- Keep functions small and focused
- Add comments for exported functions

### Testing
- Write tests for new functionality
- Aim for 80% code coverage minimum
- Use table-driven tests where appropriate
- Mock external dependencies

### Documentation
- Update README.md for user-facing changes
- Document non-obvious code with comments
- Update CHANGELOG.md for notable changes

## Project Structure

```
src/
├── cmd/copilot/       # Command implementations
│   └── commands/      # Individual commands
└── internal/          # Internal packages
    └── logging/       # Logging utilities

web/                   # Documentation website
scripts/               # PR install/uninstall scripts
```

## Quality Gates

Before submitting a PR, ensure:
- [ ] All tests pass: `go test ./...`
- [ ] Linter passes: `golangci-lint run`
- [ ] Code is formatted: `go fmt ./...`
- [ ] Documentation is updated
- [ ] Commit messages follow Conventional Commits

## Pull Request Process

1. Update documentation with details of changes
2. Update CHANGELOG.md with notable changes
3. Ensure all tests pass and coverage meets requirements
4. Request review from maintainers
5. Address review feedback
6. Once approved, maintainer will merge

## Getting Help

- Open an issue for bugs or feature requests
- Start a discussion for questions
- Check existing issues and documentation first

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
