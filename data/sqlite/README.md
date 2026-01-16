# SQLite Driver for ncore/data

SQLite database driver implementation for ncore/data using mattn/go-sqlite3.

## Installation

```bash
go get github.com/ncobase/ncore/data/sqlite
```

**Note**: This driver requires CGO to be enabled because it uses the C-based sqlite3 library.

## Usage

Import the driver in your application:

```go
package main

import (
    "context"
    "database/sql"
    "log"

    "github.com/ncobase/ncore/data"
    "github.com/ncobase/ncore/data/config"
    
    _ "github.com/ncobase/ncore/data/sqlite" // Import to auto-register
)

func main() {
    ctx := context.Background()

    // Get the SQLite driver
    driver, err := data.GetDatabaseDriver("sqlite")
    if err != nil {
        log.Fatal(err)
    }

    // Configure connection
    cfg := &config.DBNode{
        Source:          "file:test.db?cache=shared&mode=rwc",
        MaxIdleConn:     2,
        MaxOpenConn:     1, // SQLite typically uses 1 for write safety
        ConnMaxLifeTime: 0,
    }

    // Connect to database
    conn, err := driver.Connect(ctx, cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer driver.Close(conn)

    // Use the connection
    db := conn.(*sql.DB)
    
    // Ping to verify connection
    if err := driver.Ping(ctx, conn); err != nil {
        log.Fatal(err)
    }
    
    log.Println("Successfully connected to SQLite")
}
```

## Connection String Format

SQLite supports multiple connection string formats:

### Simple File Path

```go
"test.db"               // Creates/opens test.db in current directory
"/path/to/database.db"  // Absolute path
```

### URI Format

```
file:<path>?<parameters>
```

### Examples

```go
// URI with shared cache
"file:test.db?cache=shared&mode=rwc"

// In-memory database (private to connection)
":memory:"

// Shared in-memory database (shared across connections)
"file::memory:?cache=shared"

// Read-only mode
"file:test.db?mode=ro"

// WAL mode with timeout
"file:test.db?_journal_mode=WAL&_timeout=5000"

// Busy timeout
"file:test.db?_busy_timeout=10000"
```

## URI Parameters

### Cache Mode (`cache`)

- `shared`: Database cache is shared across connections
- `private` (default): Each connection has its own cache

### Open Mode (`mode`)

- `ro`: Read-only mode
- `rw`: Read-write mode (database must exist)
- `rwc` (default): Read-write-create mode
- `memory`: In-memory database

### SQLite Pragmas

You can set SQLite pragmas using underscore-prefixed parameters:

```go
// Journal mode
"_journal_mode=WAL"    // DELETE, TRUNCATE, PERSIST, MEMORY, WAL, OFF

// Synchronous mode
"_synchronous=NORMAL"  // OFF, NORMAL, FULL, EXTRA

// Busy timeout (milliseconds)
"_busy_timeout=5000"

// Foreign keys
"_foreign_keys=true"   // true, false

// Cache size (negative = KB, positive = pages)
"_cache_size=-10000"   // 10MB cache
```

## Configuration

The driver accepts `*config.DBNode` configuration:

```go
type DBNode struct {
    Driver          string        // "sqlite"
    Source          string        // SQLite connection string
    Logging         bool          // Enable query logging
    MaxIdleConn     int           // Maximum idle connections (default: 2)
    MaxOpenConn     int           // Maximum open connections (recommended: 1)
    ConnMaxLifeTime time.Duration // Maximum connection lifetime
    Weight          int           // For load balancing (unused in driver)
}
```

### Connection Pool Recommendations

SQLite has specific limitations regarding concurrency:

```go
// Recommended for most use cases (prevents database locking issues)
MaxIdleConn: 2
MaxOpenConn: 1  // IMPORTANT: SQLite works best with 1 writer

// For read-heavy workloads with WAL mode
MaxIdleConn: 5
MaxOpenConn: 5  // Multiple readers allowed with WAL
```

**Note**: Setting `MaxOpenConn > 1` can lead to "database is locked" errors unless you:

- Enable WAL (Write-Ahead Logging) mode
- Use `cache=shared` in connection string
- Handle busy timeouts properly

## Features

- Automatic driver registration via `init()`
- Connection pooling support (with SQLite-specific recommendations)
- Connection health checks via Ping
- Standard `database/sql` compatibility
- Support for all SQLite connection modes and pragmas
- Comprehensive error messages

## Performance Tips

1. **Use WAL Mode for Better Concurrency**:
   ```go
   Source: "file:test.db?_journal_mode=WAL"
   ```

2. **Enable Shared Cache for Multiple Connections**:
   ```go
   Source: "file:test.db?cache=shared"
   ```

3. **Optimize Synchronous Mode**:
   ```go
   Source: "file:test.db?_synchronous=NORMAL"  // Balance between speed and safety
   ```

4. **Set Busy Timeout**:
   ```go
   Source: "file:test.db?_busy_timeout=5000"  // Wait 5s when database is locked
   ```

5. **Use Transactions for Batch Operations**:
   ```go
   tx, _ := db.Begin()
   // ... multiple operations
   tx.Commit()
   ```

## Common Issues

### "database is locked"

**Cause**: Multiple connections trying to write simultaneously.

**Solutions**:

- Set `MaxOpenConn: 1`
- Enable WAL mode: `_journal_mode=WAL`
- Increase busy timeout: `_busy_timeout=10000`
- Use shared cache: `cache=shared`

### "unable to open database file"

**Cause**: Permission issues or invalid path.

**Solutions**:

- Check file/directory permissions
- Ensure parent directory exists
- Use absolute path or verify working directory

### CGO Build Errors

**Cause**: CGO is disabled or C compiler not available.

**Solutions**:

```bash
# Enable CGO
export CGO_ENABLED=1

# Install C compiler (macOS)
xcode-select --install

# Install C compiler (Linux/Debian)
apt-get install build-essential

# Install C compiler (Linux/RHEL)
yum groupinstall "Development Tools"
```

## Dependencies

- [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) v1.14.33 (requires CGO)
- [github.com/ncobase/ncore/data](https://github.com/ncobase/ncore/data) v0.2.0

## Alternatives

If you cannot use CGO, consider:

- [modernc.org/sqlite](https://gitlab.com/cznic/sqlite) - Pure Go SQLite (slower but no CGO)
- PostgreSQL/MySQL for production workloads

## License

See main ncore LICENSE file.
