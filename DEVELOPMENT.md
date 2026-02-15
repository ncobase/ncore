# NCore Development Guide

## Quick Start

### Clone Project

```bash
git clone https://github.com/ncobase/ncore.git
cd ncore
```

### Sync Dependencies

```bash
# Method 1: Using Makefile (recommended)
make sync

# Method 2: Direct go command
go work sync
```

### Run Tests

```bash
# Method 1: Using Makefile (recommended)
make test              # Run all tests
make test-v            # Verbose output
make test-cover        # With coverage

# Method 2: Using scripts
./scripts/test.sh
./scripts/test.sh -v
./scripts/test.sh -cover

# Test specific module
cd data
go test ./...
```

### Add New Module

```bash
mkdir newmodule && cd newmodule
go mod init github.com/ncobase/ncore/newmodule
cd .. && echo " ./newmodule" >> go.work
go work sync
```

## Module Development Guidelines

### Module Naming & Versioning

- **Naming**: Lowercase, words concatenated or underscore-separated (e.g., `ctxutil`, `data`)
- **Versioning**: Independent per module using git tags (e.g., `data/v0.1.0`)

```bash
./scripts/tag.sh v0.1.0  # Batch release all modules
```

### Dependency Management

```bash
# Update all modules
make update && make sync

# Update specific module
./scripts/update-deps.sh data

# Check outdated
make check-outdated
```

**Note**: Root has no go.mod - use `make update` or scripts, not `go get -u ./...`

### Testing & Formatting

```bash
go test ./...              # Run tests
go test -cover ./...       # With coverage
go fmt ./...               # Format code
golangci-lint run          # Lint (if installed)
```

### Dependency Injection (Google Wire)

Modules provide `ProviderSet` for Wire integration:

```go
//go:build wireinject

func InitializeApp() (*App, func(), error) {
    panic(wire.Build(
        config.ProviderSet,
        logger.ProviderSet,
        data.ProviderSet,
        NewApp,
    ))
}
```

Generate wire code: `wire ./...`

## Integration with Applications

### Method 1: Using Released Versions

In your application's `go.mod`:

```go
require (
    github.com/ncobase/ncore/data v0.1.0
    github.com/ncobase/ncore/config v0.1.0
)
```

### Method 2: Using replace for Local Development

Add to your application's `go.mod`:

```go
replace (
    github.com/ncobase/ncore/data => /path/to/ncore/data
    github.com/ncobase/ncore/config => /path/to/ncore/config
)
```

Then:

```bash
cd <your-app>
go mod tidy
```

### Method 3: Using workspace (recommended for development)

Create `go.work` in your application directory:

```text
go 1.24

use (
    .
    /path/to/ncore/data
    /path/to/ncore/config
    // Add needed modules
)
```

## Common Commands

```bash
# Setup
git clone https://github.com/ncobase/ncore.git && cd ncore && make sync

# Testing
make test           # All tests
make test-v         # Verbose
make test-cover     # With coverage

# Dependencies
make update && make sync        # Update all
./scripts/update-deps.sh data   # Update specific module
make check-outdated             # Check outdated

# Code Quality
make fmt            # Format
make lint           # Lint (requires golangci-lint)
make clean          # Clean artifacts

# Versioning
make tag VERSION=v0.1.0         # Tag all modules
git push origin --tags          # Push tags
```

## Available Scripts

```bash
./scripts/update-deps.sh [module]   # Update dependencies
./scripts/check-outdated.sh         # Check outdated deps
./scripts/test.sh [-v] [--cover]    # Run tests
./scripts/tag.sh v0.1.0             # Batch tag modules
```

## Makefile Targets

Run `make help` to see all available targets.

## Common Issues

**`go work sync` errors:** Run `go clean -modcache && go work sync`

**View dependencies:** `cd <module> && go mod graph`

**Circular dependencies:** Extract shared code to common module or use interfaces

## CI/CD Example

```yaml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with: { go-version: "1.24" }
      - run: go work sync && bash scripts/test.sh
```

## Contributing

1. Fork → 2. Feature branch → 3. Commit → 4. Push → 5. Pull Request

## Tips

```bash
# Race detection
go test -race ./...

# Coverage report
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

# Clean cache
go clean -modcache && go work sync
```
