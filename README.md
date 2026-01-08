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

### Example Output

The tool generates a formatted changelog that highlights all API changes:

```markdown
# API Changelog

## [2.0.0] - 2026-01-08

**Total Changes:** 23
**Breaking Changes:** 5

### ⚠️ Breaking Changes
- Endpoint `DELETE /api/v1/users/{id}` removed
- Parameter `userId` removed from `GET /api/v1/orders`
- Property `phoneNumber` removed from schema `User`
- Response code `200` removed from `POST /api/v1/auth/login`
- Required parameter `apiKey` added to `GET /api/v1/products`

### ✨ Added
- Endpoint `POST /api/v2/users` added
- Endpoint `GET /api/v2/users/{id}/profile` added
- Response code `201` added to `POST /api/v2/users`
- Response code `404` added to `GET /api/v2/users/{id}`
- Parameter `limit` added to `GET /api/v2/products`
- Parameter `offset` added to `GET /api/v2/products`
- Schema `UserProfile` added
- Schema `ProductCategory` added
- Property `email` added to schema `User`
- Property `createdAt` added to schema `User`
- Property `tags` added to schema `Product`

### 🔄 Modified
- API version updated from `1.5.3` to `2.0.0`
- Endpoint `GET /api/v1/products` path changed to `GET /api/v2/products`
- Parameter `search` type changed from `string` to `object` in `GET /api/v2/products`
- Property `price` type changed from `integer` to `number` in schema `Product`
- Response code `200` description updated in `GET /api/v2/users`

### 🗑️ Removed
- Endpoint `GET /api/v1/legacy/stats` removed
- Parameter `deprecated_filter` removed from `GET /api/v2/products`
- Property `oldStatus` removed from schema `Order`

### 🔧 Deprecated
- Endpoint `GET /api/v1/users` marked as deprecated
- Parameter `old_format` marked as deprecated in `GET /api/v2/reports`
```


## Requirements

- Go 1.25.2 or later
