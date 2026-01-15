# NCore Multi-Module Architecture

## Architecture Design

NCore adopts a multi-module architecture design where each sub-package is an independent Go module. The benefits of this design are:

1. **Reduced Dependencies**: Users only need to import the sub-modules they require, avoiding unnecessary dependencies
2. **Smaller Build Size**: Avoids bundling all NCore dependencies into the application
3. **Independent Version Management**: Each module can be upgraded independently without affecting others
4. **Clearer Module Boundaries**: Enforces clear dependency relationships between modules

## Module List

```text
github.com/ncobase/ncore/
├── concurrency    - Concurrency utilities
├── config         - Configuration management
├── consts         - Constants definitions
├── ctxutil        - Context utilities
├── data           - Data layer (databases, cache, search engines, etc.)
├── ecode          - Error codes
├── extension      - Extension and plugin system
├── logging        - Logging
├── messaging      - Message queues
├── net            - Network utilities
├── security       - Security features
├── types          - Common types
├── utils          - Utility functions
├── validation     - Data validation
└── version        - Version management
```

## Usage

### Using in Applications

In your application's `go.mod`, only import the modules you need:

```go
require (
    github.com/ncobase/ncore/config v0.0.0-20251022025300-781956ac0776
    github.com/ncobase/ncore/data v0.0.0-20251022025300-781956ac0776
    github.com/ncobase/ncore/logging v0.0.0-20251022025300-781956ac0776
    // Only import the modules you need
)
```

### Local Development

#### 1. Using go.work (Recommended)

The project root provides a `go.work` file for convenient local development and testing:

```bash
# In ncore root directory
go work sync  # Sync all module dependencies
bash scripts/test.sh # Test all modules
```

#### 2. Using local ncore in applications

Use the `replace` directive in your application's `go.mod`:

```go
replace (
    github.com/ncobase/ncore/data => /path/to/ncore/data
    github.com/ncobase/ncore/config => /path/to/ncore/config
    // Replace the modules you need to debug locally
)
```

## Release Process

### Release Single Module

```bash
cd data
git tag data/v0.1.0
git push origin data/v0.1.0
```

### Batch Release

```bash
# Tag all modules with the same version
./scripts/tag.sh v0.1.0
```

## Module Dependency Principles

1. **Minimal Dependencies**: Each module should only import necessary dependencies
2. **Avoid Circular Dependencies**: Modules cannot have circular dependencies
3. **Common Modules First**: `types`, `consts`, `ecode` and other common modules should have zero or minimal dependencies
4. **Large Dependency Isolation**: Modules like `data` containing large dependencies (databases, search engines) should be isolated

## FAQ

### Q: Why is there no go.mod in the root directory?

A: Because each sub-package is an independent module, the root directory doesn't need go.mod. The go.work file is sufficient for managing local development.

### Q: Should go.work be committed to git?

A: It can be committed. go.work facilitates local development for team members, but `go.work.sum` should not be committed (already in .gitignore).

### Q: How to add a new module?

A:

1. Create a new directory
2. Run `go mod init github.com/ncobase/ncore/new-module-name` in the directory
3. Add `./new-module-name` to the root directory's `go.work`

### Q: How do modules reference each other?

A: Use the full module path directly:

```go
import (
    "github.com/ncobase/ncore/config"
    "github.com/ncobase/ncore/logging"
)
```

During local development, go.work will automatically resolve these references.

## Dependency Injection Support (Google Wire)

NCore modules provide native [Google Wire](https://github.com/google/wire) support, with each module exposing a `ProviderSet` for dependency injection.

### Supported Modules

| Module              | ProviderSet               | Provides                         | Cleanup |
| ------------------- | ------------------------- | -------------------------------- | ------- |
| `config`            | `config.ProviderSet`      | `*Config` and sub-configurations | No      |
| `logging/logger`    | `logger.ProviderSet`      | `*Logger`                        | Yes     |
| `data`              | `data.ProviderSet`        | `*Data`                          | Yes     |
| `extension/manager` | `manager.ProviderSet`     | `*Manager`                       | Yes     |
| `security`          | `security.ProviderSet`    | JWT `*TokenManager`              | No      |
| `messaging`         | `messaging.ProviderSet`   | Email `Sender`                   | No      |
| `concurrency`       | `concurrency.ProviderSet` | Worker `*Pool`                   | Yes     |

### Wire Usage Example

```go
//go:build wireinject
// +build wireinject

package main

import (
    "github.com/google/wire"
    "github.com/ncobase/ncore/config"
    "github.com/ncobase/ncore/logging/logger"
    "github.com/ncobase/ncore/data"
    "github.com/ncobase/ncore/extension/manager"
)

// InitializeApp initializes the application
func InitializeApp() (*App, func(), error) {
    panic(wire.Build(
        // Configuration management
        config.ProviderSet,

        // Core components
        logger.ProviderSet,
        data.ProviderSet,
        manager.ProviderSet,

        // Application constructor
        NewApp,
    ))
}
```

### Features

1. **Cleanup Function Support**: Providers for `data`, `logger`, `manager`, and `concurrency` modules return cleanup functions
2. **Configuration Extraction**: `config.ProviderSet` automatically extracts sub-configurations required by other modules
3. **Interface Binding**: `security` module uses `wire.Bind` for interface binding
4. **Error Handling**: All providers properly handle and propagate errors

For detailed documentation, see:

- [README](README.md#dependency-injection-google-wire)
- [DEVELOPMENT.md](DEVELOPMENT.md#6-dependency-injection-google-wire)
- [Example Code](examples/09-wire)

## Related Resources

- [Go Modules Official Documentation](https://go.dev/doc/modules/managing-dependencies)
- [Go Workspaces Official Documentation](https://go.dev/doc/tutorial/workspaces)
- [Multi-module repositories Best Practices](https://github.com/golang/go/wiki/Modules#faqs--multi-module-repositories)
- [Google Wire Official Documentation](https://github.com/google/wire)
