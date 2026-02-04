# azd-copilot

An Azure Developer CLI extension for GitHub Copilot integration.

## Prerequisites

- [Azure Developer CLI (azd)](https://learn.microsoft.com/azure/developer/azure-developer-cli/install-azd)
- [Go 1.25+](https://golang.org/dl/)
- [Mage](https://magefile.org/) (optional, for build automation)

## Installation

```bash
azd extension install jongio.azd.copilot
```

## Development

### Build and Install

```bash
# Using mage (recommended)
mage build

# Or using azd directly
azd x build
```

### Run Tests

```bash
mage test
```

### Lint

```bash
mage lint
```

## Usage

```bash
# Show version
azd copilot version

# Show help
azd copilot --help
```

## License

MIT
