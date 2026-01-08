# API Changelog Generator

Generate changelogs from OpenAPI specification changes.

## Features

- Compare two versions of OpenAPI specifications
- Automatically detect changes in endpoints, parameters, responses, and schemas
- Generate formatted changelog documents

## Installation

### From Source (Clone/Fork)

```bash
# Clone the repository
git clone https://github.com/yourusername/api-changelog-generator.git
cd api-changelog-generator

# Install to $GOPATH/bin (recommended)
make install
```

Or build without installing:

```bash
# Build the binary locally
make build

# The binary will be available at:
# - Windows: bin/api-changelog.exe
# - Unix: bin/api-changelog
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


## Requirements

- Go 1.25.2 or later
