# NCore

A comprehensive Go application components library for building modern, scalable applications.

## Features

- **Modular Architecture**: Import only the modules you need
- **Rich Integrations**: Database, search, messaging, and storage solutions
- **Security & Authentication**: JWT, OAuth, encryption utilities
- **Observability**: OpenTelemetry, logging, and monitoring

## Multi-Module Architecture

NCore uses a **multi-module architecture** where each sub-package is an independent Go module, providing minimal dependencies and independent versioning.

### Available Modules

```text
github.com/ncobase/ncore/concurrency    - Concurrency utilities
github.com/ncobase/ncore/config         - Configuration management
github.com/ncobase/ncore/consts         - Constants
github.com/ncobase/ncore/ctxutil        - Context utilities
github.com/ncobase/ncore/data           - Data layer (DB, cache, search)
github.com/ncobase/ncore/ecode          - Error codes
github.com/ncobase/ncore/extension      - Extension system
github.com/ncobase/ncore/logging        - Logging
github.com/ncobase/ncore/messaging      - Message queue
github.com/ncobase/ncore/net            - Network utilities
github.com/ncobase/ncore/security       - Security features
github.com/ncobase/ncore/types          - Common types
github.com/ncobase/ncore/utils          - Utility functions
github.com/ncobase/ncore/validation     - Validation
github.com/ncobase/ncore/version        - Version info
```

## Installation

Import only the modules you need:

```bash
go get github.com/ncobase/ncore/config
go get github.com/ncobase/ncore/data
go get github.com/ncobase/ncore/security
```

## Quick Start

```go
package main

import (
    "github.com/ncobase/ncore/config"
    "github.com/ncobase/ncore/logging"
)

func main() {
    // Load configuration
    cfg, err := config.LoadConfig("config.yaml")
    if err != nil {
        panic(err)
    }

    // Initialize logger
    logger := logging.NewLogger(cfg.Logging)
    logger.Info("Application started")
}
```

## Development

```bash
# Clone the repository
git clone https://github.com/ncobase/ncore.git
cd ncore

# Sync dependencies
go work sync

# Run tests
bash scripts/test.sh
```

## Documentation

- [DEVELOPMENT.md](DEVELOPMENT.md) - Development guide
- [MODULES.md](MODULES.md) - Multi-module architecture explanation

## Code Generation

For scaffolding new projects and components, use the CLI tool:

```bash
go install github.com/ncobase/cli@latest
nco create core auth-service
nco create business payment --use-mongo --with-test
```

## License

See [LICENSE](LICENSE) file for details.
