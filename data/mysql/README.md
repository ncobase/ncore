# MySQL Driver for ncore/data

MySQL database driver implementation for ncore/data.

## Installation

```bash
go get github.com/ncobase/ncore/data/mysql
```

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
    
    _ "github.com/ncobase/ncore/data/mysql" // Import to auto-register
)

func main() {
    ctx := context.Background()

    // Get the MySQL driver
    driver, err := data.GetDatabaseDriver("mysql")
    if err != nil {
        log.Fatal(err)
    }

    // Configure connection
    cfg := &config.DBNode{
        Source:          "user:password@tcp(localhost:3306)/dbname?parseTime=true&charset=utf8mb4",
        MaxIdleConn:     10,
        MaxOpenConn:     100,
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
    
    log.Println("Successfully connected to MySQL")
}
```

## DSN Format

The MySQL driver uses the standard MySQL DSN format:

```
[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
```

### Common DSN Examples

```go
// TCP connection
"user:password@tcp(localhost:3306)/dbname"

// TCP with parameters
"user:password@tcp(localhost:3306)/dbname?parseTime=true&charset=utf8mb4"

// Unix socket
"user:password@unix(/var/run/mysqld/mysqld.sock)/dbname"

// With timeout
"user:password@tcp(localhost:3306)/dbname?timeout=30s&readTimeout=30s&writeTimeout=30s"
```

### Recommended Parameters

- `parseTime=true` - Parse DATE and DATETIME to time.Time
- `charset=utf8mb4` - Use UTF-8 encoding
- `loc=Local` - Use local timezone

## Configuration

The driver accepts `*config.DBNode` configuration:

```go
type DBNode struct {
    Driver          string        // "mysql"
    Source          string        // MySQL DSN
    Logging         bool          // Enable query logging
    MaxIdleConn     int           // Maximum idle connections
    MaxOpenConn     int           // Maximum open connections
    ConnMaxLifeTime time.Duration // Maximum connection lifetime
    Weight          int           // For load balancing (unused in driver)
}
```

## Features

- Automatic driver registration via `init()`
- Connection pooling support
- Connection health checks via Ping
- Standard `database/sql` compatibility
- Comprehensive error messages

## Dependencies

- [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql) v1.9.3
- [github.com/ncobase/ncore/data](https://github.com/ncobase/ncore/data) v0.2.0

## License

See main ncore LICENSE file.
