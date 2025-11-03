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
# 1. Create module directory
mkdir newmodule

# 2. Initialize module
cd newmodule
go mod init github.com/ncobase/ncore/newmodule

# 3. Add to workspace
cd ..
echo " ./newmodule" >> go.work

# 4. Sync dependencies
go work sync
```

## Module Development Guidelines

### 1. Module Naming

- Use lowercase letters
- Multiple words separated by underscores or directly concatenated
- Examples: `ctxutil`, `data`, `messaging`

### 2. Version Management

Each module has independent versioning:

```bash
# Release single module
cd data
git tag data/v0.1.0
git push origin data/v0.1.0

# Batch release all modules
./scripts/tag.sh v0.1.0
git push origin --tags
```

### 3. Dependency Management

#### Add Dependencies

```bash
cd <module-name>
go get <dependency>
go mod tidy
```

#### Update Dependencies

```bash
# Method 1: Update all dependencies for all modules (recommended)
make update            # Upgrade all module dependencies
make sync              # Sync workspace

# Method 2: Use scripts
./scripts/update-deps.sh           # Update all modules
./scripts/update-deps.sh data      # Update only data module

# Method 3: Manual update for specific module
cd <module-name>
go get -u ./...        # Upgrade all dependencies to latest minor/patch
go get -u <dependency> # Upgrade specific dependency
go mod tidy

# Check outdated dependencies
make check-outdated
# or
./scripts/check-outdated.sh
```

**Important Notes**:

- ⚠️ Since the root directory has no go.mod, **cannot** run `go get -u ./...` directly in the root
- ✅ Must use `make update` or scripts to update all modules
- ✅ Or manually update individual modules in their directories

#### Inter-module Dependencies

```go
// In go.mod
require (
    github.com/ncobase/ncore/types v0.0.0-20251022025300-781956ac0776
)

// In code
import "github.com/ncobase/ncore/types"
```

### 4. Testing

Every module should have comprehensive tests:

```bash
# Run tests
go test ./...

# With coverage
go test -cover ./...

# Verbose output
go test -v ./...
```

### 5. Code Formatting

```bash
# Format code
go fmt ./...

# Run linter (if configured)
golangci-lint run
```

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

### Development Environment Setup

```bash
# Clone project
git clone https://github.com/ncobase/ncore.git
cd ncore

# Sync dependencies
make sync
```

### Dependency Management

#### ❌ Wrong Approach

```bash
# This won't work! Root directory has no go.mod
go get -u ./...
```

#### ✅ Correct Approach

```bash
# Upgrade all dependencies for all modules
make update
make sync

# Upgrade specific module
./scripts/update-deps.sh data

# Check which dependencies are outdated
make check-outdated

# Manually upgrade single module
cd data
go get -u ./...
go mod tidy
cd ..
```

### Testing

```bash
# Run all tests
make test

# Verbose output
make test-v

# With coverage
make test-cover

# Test single module
cd data
go test -v ./...
```

### Code Quality

```bash
# Format code
make fmt

# Run linter (need to install golangci-lint first)
make lint

# Clean build artifacts
make clean
```

### Version Release

```bash
# Tag all modules
make tag VERSION=v0.1.0

# Push tags
git push origin --tags

# Tag only single module
cd data
git tag data/v0.1.0
git push origin data/v0.1.0
```

## Available Scripts

### `scripts/update-deps.sh`

Script to upgrade dependencies

```bash
# Upgrade all modules
./scripts/update-deps.sh

# Upgrade only specific module
./scripts/update-deps.sh data
```

### `scripts/check-outdated.sh`

Check outdated dependencies

```bash
./scripts/check-outdated.sh
```

### `scripts/test.sh`

Run all tests

```bash
# Basic test
./scripts/test.sh

# Verbose output
./scripts/test.sh -v

# With coverage
./scripts/test.sh -cover
```

### `scripts/tag.sh`

Batch tagging

```bash
./scripts/tag.sh v0.1.0
```

## Makefile Targets

| Command | Description |
|---------|-------------|
| `make help` | Show help information |
| `make sync` | Sync workspace dependencies |
| `make test` | Run all tests |
| `make test-v` | Run all tests (verbose) |
| `make test-cover` | Run tests with coverage |
| `make update` | Update all dependencies |
| `make check-outdated` | Check outdated dependencies |
| `make tag VERSION=v0.1.0` | Create tags |
| `make fmt` | Format code |
| `make lint` | Run linter |
| `make clean` | Clean build artifacts |

## Common Issues

### Q: `go work sync` reports errors, what to do?

A: Try these steps:

```bash
# Clean module cache
go clean -modcache

# Resync
go work sync

# If still issues, update modules individually
cd <module-name>
go mod tidy
```

### Q: How to view module dependency relationships?

```bash
cd <module-name>
go mod graph
```

### Q: How to upgrade dependencies for all modules?

```bash
# Create script or execute manually
for dir in */; do
    if [ -f "$dir/go.mod" ]; then
        echo "Updating $dir"
        cd "$dir"
        go get -u ./...
        go mod tidy
        cd ..
    fi
done
```

### Q: What to do about circular dependencies between modules?

A: Redesign module structure, possible solutions:

1. Extract shared code to new common module (like `types`)
2. Use interfaces instead of concrete implementations
3. Adjust module responsibility division

## CI/CD Recommendations

### GitHub Actions Example

```yaml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Sync workspace
        run: go work sync

      - name: Run tests
        run: bash scripts/test.sh

      - name: Test each module
        run: |
          for dir in */; do
            if [ -f "$dir/go.mod" ]; then
              echo "Testing $dir"
              cd "$dir"
              go test -v ./...
              cd ..
            fi
          done
```

## Performance Optimization Recommendations

1. **Minimize Dependencies**: Each module should only introduce necessary dependencies
2. **Lazy Loading**: Large dependencies (like database drivers) in separate modules
3. **Interface First**: Modules interact through interfaces to reduce coupling
4. **Complete Documentation**: Clear module responsibilities and API documentation

## Contributing Guidelines

1. Fork the project
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Create Pull Request

## Recommended Tools

- **golangci-lint**: Code checking
- **go-mod-outdated**: Check outdated dependencies
- **go-mod-upgrade**: Batch upgrade dependencies
- **air**: Hot reload (during development)

## Workflow Examples

### Add New Feature

```bash
# 1. Sync dependencies
make sync

# 2. Develop feature
cd data
# ... write code ...

# 3. If new dependencies needed
go get github.com/some/package
go mod tidy

# 4. Run tests
go test ./...

# 5. Return to root directory, test all modules
cd ..
make test

# 6. Format code
make fmt

# 7. Commit
git add .
git commit -m "Add new feature"
```

### Fix Bug

```bash
# 1. Locate bug module
cd <module>

# 2. Fix code

# 3. Run tests
go test ./...

# 4. Return to root directory
cd ..

# 5. Run all tests
make test

# 6. If important fix, release patch version
make tag VERSION=v0.1.1
git push origin --tags
```

### Upgrade Dependencies

```bash
# 1. Check outdated dependencies
make check-outdated

# 2. Upgrade dependencies
make update

# 3. Sync workspace
make sync

# 4. Run tests to ensure everything works
make test

# 5. Commit changes
git add .
git commit -m "Update dependencies"
```

## Tips

### Test Only Specific Packages

```bash
cd data/databases
go test -v .
```

### Testing with Race Detection

```bash
cd <module>
go test -race ./...
```

### View Test Coverage Report

```bash
cd <module>
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Clean Module Cache

```bash
go clean -modcache
go work sync
```
