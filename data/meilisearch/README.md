# Meilisearch Driver for ncore/data

This package provides a Meilisearch driver implementation for the ncore/data abstraction layer.

## Installation

```bash
go get github.com/ncobase/ncore/data/meilisearch
```

## Usage

Import the driver with a blank import to register it:

```go
import (
    "context"
    "github.com/ncobase/ncore/data"
    "github.com/ncobase/ncore/data/config"
    _ "github.com/ncobase/ncore/data/meilisearch"  // Register driver
)
```

### Basic Example

```go
// Get the registered driver
driver, err := data.GetSearchDriver("meilisearch")
if err != nil {
    log.Fatal(err)
}

// Configure connection
cfg := &config.Meilisearch{
    Host:   "http://localhost:7700",
    APIKey: "masterKey",  // Optional for development
}

// Connect
ctx := context.Background()
conn, err := driver.Connect(ctx, cfg)
if err != nil {
    log.Fatal(err)
}
defer driver.Close(conn)

// Use the connection (type assertion to *meili.Client)
client := conn.(*meili.Client)

// Perform operations
health, err := client.Health()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Meilisearch status: %s\n", health.Status)
```

## Configuration

The driver expects a `*config.Meilisearch` configuration:

```go
type Meilisearch struct {
    Host   string  // Meilisearch server URL (e.g., "http://localhost:7700")
    APIKey string  // API key for authentication (optional for development)
}
```

### YAML Configuration Example

```yaml
data:
  search:
    default_engine: meilisearch
    meilisearch:
      host: http://localhost:7700
      api_key: your_master_key
```

## Features

- **Automatic Registration**: Driver registers itself on import via `init()`
- **Health Verification**: Connection is verified with a health check before being returned
- **Client Wrapper**: Returns a `*meili.Client` with a comprehensive API for Meilisearch operations
- **Error Handling**: Provides detailed error messages for debugging

## Architecture

This driver implements the `data.SearchDriver` interface:

```go
type SearchDriver interface {
    Name() string
    Connect(ctx context.Context, cfg interface{}) (interface{}, error)
    Close(conn interface{}) error
}
```

The driver wraps the existing `data/search/meili` client, which provides:

- Document management (add, update, delete)
- Search operations
- Index management
- Task monitoring
- Settings configuration

## Development

Run tests:

```bash
cd data/meilisearch
go test -v
```

Note: Some tests require a running Meilisearch instance. If Meilisearch is not available, connection tests will
gracefully skip.

## Dependencies

- `github.com/meilisearch/meilisearch-go` - Official Meilisearch Go SDK
- `github.com/ncobase/ncore/data` - Core data layer abstraction
- `github.com/ncobase/ncore/data/search/meili` - Meilisearch client wrapper

## License

See the main ncore repository for license information.
