# API Changelog Generator

Generate changelogs from OpenAPI specification changes.

## Features

- Compare two versions of OpenAPI specifications
- Automatically detect changes in endpoints, parameters, responses, and schemas
- Generate formatted changelog documents

## Installation

```bash
go install steamedeo.dev/api-changelog-generator@latest
```

## Usage

```bash
# Show available commands
api-changelog --help

# Compare OpenAPI specs and generate a changelog
api-changelog compare --latest <new-spec> --previous <old-spec> --output <changelog-file>

# Show version info
api-changelog --version-info
```

### Example

```bash
api-changelog compare \
  --latest ./specs/v2.yaml \
  --previous ./specs/v1.yaml \
  --output ./CHANGELOG.md
```

## Building from Source

```bash
git clone <repository-url>
cd api-changelog-generator
make build
```

The binary will be available at `bin/api-changelog.exe`.

## Requirements

- Go 1.25.2 or later
