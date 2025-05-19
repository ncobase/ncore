# ncore/loger

A powerful logging system built on [logrus](https://github.com/sirupsen/logrus) with support for multiple output targets and search engine integrations (Elasticsearch, OpenSearch, and Meilisearch).

## Features

- Multiple log levels (Trace, Debug, Info, Warn, Error, Fatal, Panic)
- Structured JSON logging
- Context-aware tracing
- File rotation
- Multiple outputs (console, file, search engines)
- Search engine integrations:
  - Elasticsearch
  - OpenSearch
  - Meilisearch

## Usage

### Initialization

```go
import (
    "github.com/ncobase/ncore/config"
    "github.com/ncobase/ncore/logging/logger"
)

// Create configuration
loggerConfig := &config.Logger{
    Level:    4, // Info level
    Format:   "json",
    Output:   "stdout",
    IndexName: "application-logs",
}

// Initialize logger
cleanup, err := logger.New(loggerConfig)
if err != nil {
    panic(err)
}
defer cleanup()

// Set application version (optional)
logger.SetVersion("1.0.0")
```

### Basic Logging

```go
import (
    "context"
    "github.com/ncobase/ncore/logging/logger"
    "github.com/sirupsen/logrus"
)

ctx := context.Background()

// Basic logs
logger.Debug(ctx, "Debug message")
logger.Info(ctx, "Info message")
logger.Warn(ctx, "Warning message")
logger.Error(ctx, "Error message")

// Formatted logs
logger.Infof(ctx, "User %s logged in with role: %s", "john", "admin")

// With fields
logger.WithFields(ctx, logrus.Fields{
    "user_id": "12345",
    "action":  "login",
    "ip":      "192.168.1.1",
}).Info("User login successful")
```

### Configure File Output

```go
loggerConfig := &config.Logger{
    Level:      4,
    Format:     "json",
    Output:     "file",
    OutputFile: "./logs/app.log",
}
```

### Configure Elasticsearch

```go
loggerConfig := &config.Logger{
    Level:      4,
    Format:     "json",
    IndexName:  "application-logs",
    Elasticsearch: struct {
        Addresses []string
        Username  string
        Password  string
    }{
        Addresses: []string{"http://elasticsearch:9200"},
        Username:  "elastic",
        Password:  "password",
    },
}
```

### Configure OpenSearch

```go
loggerConfig := &config.Logger{
    Level:      4,
    Format:     "json",
    IndexName:  "application-logs",
    OpenSearch: struct {
        Addresses      []string
        Username       string
        Password       string
        InsecureSkipTLS bool
    }{
        Addresses:      []string{"https://opensearch:9200"},
        Username:       "admin",
        Password:       "admin",
        InsecureSkipTLS: true,
    },
}
```

### Configure Meilisearch

```go
loggerConfig := &config.Logger{
    Level:      4,
    Format:     "json",
    IndexName:  "application-logs",
    Meilisearch: struct {
        Host   string
        APIKey string
    }{
        Host:   "http://meilisearch:7700",
        APIKey: "masterKey",
    },
}
```

### Request Tracing

```go
func HandleRequest(w http.ResponseWriter, r *http.Request) {
    // Create context with trace ID
    ctx, traceID := logger.EnsureTraceID(r.Context())
    
    // Add trace ID to response header
    w.Header().Set("X-Trace-ID", traceID)
    
    logger.Infof(ctx, "Processing request: %s %s", r.Method, r.URL.Path)
    
    // Request handling logic
    
    logger.Infof(ctx, "Request completed: %s %s", r.Method, r.URL.Path)
}
```

### Error Handling

```go
func ProcessData(ctx context.Context, data []byte) error {
    if len(data) == 0 {
        logger.Warn(ctx, "Received empty data")
        return nil
    }
    
    result, err := parseData(data)
    if err != nil {
        logger.WithFields(ctx, logrus.Fields{
            "error": err.Error(),
            "data_length": len(data),
        }).Error("Data parsing failed")
        return err
    }
    
    logger.WithFields(ctx, logrus.Fields{
        "result_count": len(result),
    }).Info("Data processing successful")
    
    return nil
}
```

## Log Levels

- **Trace** (6): Extremely detailed information
- **Debug** (5): Detailed debugging information
- **Info** (4): General operational information
- **Warn** (3): Warnings, potentially problematic situations
- **Error** (2): Error conditions, operational failures
- **Fatal** (1): Severe errors causing application termination
- **Panic** (0): Critical errors causing application panic
