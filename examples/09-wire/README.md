# Example 09: Wire ProviderSets

Demonstrates how to use Google Wire with NCore ProviderSets to assemble core dependencies and optional subsystems.

## Features

- **Wire DI**: ProviderSets for config, logging, data, extensions
- **JWT Wiring**: TokenManager example using Wire providers
- **Worker Pools**: Worker pool initialization via Wire with lifecycle cleanup
- **Config Driven**: Uses `config.yaml` like other examples

## Project Structure

```text
09-wire/
├── app.go          # App container
├── main.go         # Entry point
├── providers.go    # Provider helpers
├── wire.go         # Wire injectors (build tag)
├── wire_gen.go     # Generated wiring
├── config.yaml     # Example configuration
├── go.mod
├── go.sum
└── README.md
```

## Prerequisites

- Go 1.21+
- [Google Wire](https://github.com/google/wire) CLI

## Generate

```bash
# Install Wire CLI
go install github.com/google/wire/cmd/wire@latest

# Generate wiring
go generate ./...
```

## Running

```bash
go run main.go -conf config.yaml
```

The program initializes core dependencies, a JWT token manager, and a worker pool, then exits after executing a sample
task.
