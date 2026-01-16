# MongoDB Driver for ncore/data

This package provides a MongoDB driver implementation for the ncore/data framework.

## Installation

```bash
go get github.com/ncobase/ncore/data/mongodb
```

## Usage

Import the driver in your application to register it automatically:

```go
import (
    "context"
    "github.com/ncobase/ncore/data"
    "github.com/ncobase/ncore/data/config"
    _ "github.com/ncobase/ncore/data/mongodb" // Register MongoDB driver
)

func main() {
    // Get the MongoDB driver
    driver, err := data.GetDatabaseDriver("mongodb")
    if err != nil {
        panic(err)
    }

    // Configure MongoDB connection
    cfg := &config.MongoDB{
        Master: &config.MongoNode{
            URI: "mongodb://localhost:27017/mydb",
        },
        Slaves: []*config.MongoNode{
            {
                URI: "mongodb://slave1:27017/mydb",
                Weight: 2,
            },
        },
        Strategy: "round_robin",
        MaxRetry: 3,
    }

    // Connect to MongoDB
    ctx := context.Background()
    conn, err := driver.Connect(ctx, cfg)
    if err != nil {
        panic(err)
    }
    defer driver.Close(conn)

    // Use the connection (conn is *mongo.Client)
    client := conn.(*mongo.Client)
    db := client.Database("mydb")
    // ... perform operations
}
```

## Configuration

The driver expects a `*config.MongoDB` configuration struct with the following fields:

### Master (required)

The primary MongoDB node configuration:

```go
Master: &config.MongoNode{
    URI:     "mongodb://localhost:27017/mydb",  // MongoDB connection URI
    Logging: true,                              // Enable logging (optional)
    Weight:  1,                                 // Weight for load balancing (optional)
}
```

### Slaves (optional)

Slave nodes for read operations:

```go
Slaves: []*config.MongoNode{
    {
        URI:    "mongodb://slave1:27017/mydb",
        Weight: 2,
    },
    {
        URI:    "mongodb://slave2:27017/mydb",
        Weight: 1,
    },
}
```

### Strategy (optional)

Load balancing strategy for slave nodes:

- `round_robin` (default): Distributes requests evenly across slaves
- `random`: Randomly selects a slave for each request
- `weight`: Distributes requests based on node weights

### MaxRetry (optional)

Maximum number of retry attempts for failed operations.

## MongoDB URI Format

The MongoDB URI follows the standard MongoDB connection string format:

```
mongodb://[username:password@]host1[:port1][,...hostN[:portN]][/[defaultauthdb][?options]]
mongodb+srv://[username:password@]host[/[defaultauthdb][?options]]
```

### Examples

**Local development:**

```
mongodb://localhost:27017/mydb
```

**With authentication:**

```
mongodb://user:password@localhost:27017/mydb?authSource=admin
```

**MongoDB Atlas (cloud):**

```
mongodb+srv://user:password@cluster.mongodb.net/mydb
```

**Replica set:**

```
mongodb://host1:27017,host2:27017,host3:27017/mydb?replicaSet=myrs
```

## Features

- ✅ Implements `data.DatabaseDriver` interface
- ✅ Returns `*mongo.Client` for full MongoDB functionality
- ✅ Automatic connection verification (ping on connect)
- ✅ Proper context handling for all operations
- ✅ Support for MongoDB connection options via URI
- ✅ Master/slave configuration support
- ✅ Multiple load balancing strategies
- ✅ Connection pooling (configured via URI options)

## Driver Methods

### Name() string

Returns the driver identifier: `"mongodb"`

### Connect(ctx context.Context, cfg interface{}) (interface{}, error)

Establishes a MongoDB connection using the provided `*config.MongoDB` configuration.
Returns a `*mongo.Client` on success.

### Close(conn interface{}) error

Closes the MongoDB connection and releases resources.

### Ping(ctx context.Context, conn interface{}) error

Verifies the MongoDB connection is alive and functional.

## Error Handling

The driver provides detailed error messages for common issues:

- Invalid configuration type
- Missing master configuration
- Empty URI
- Connection failures
- Ping failures
- Invalid connection type

All errors are wrapped with context using `fmt.Errorf` and `%w` for proper error chain support.

## Testing

Run the test suite:

```bash
cd data/mongodb
go test -v
```

## See Also

- [MongoDB Go Driver Documentation](https://www.mongodb.com/docs/drivers/go/current/)
- [ncore/data Package](../)
- [MongoDB Connection String Documentation](https://www.mongodb.com/docs/manual/reference/connection-string/)
