# ncore/logger

A powerful logging system built on [logrus](https://github.com/sirupsen/logrus) with multi-output support, search engine
integrations, and data desensitization.

## Features

- Multiple log levels with structured JSON logging
- Context-aware tracing with automatic trace ID propagation
- **Data desensitization with deep structure support**
- Multiple outputs: console, file (auto-rotation), Elasticsearch, OpenSearch, Meilisearch
- Fixed-length masking to prevent sensitive data length disclosure

## Quick Start

```go
import (
    "context"
    "github.com/ncobase/ncore/logging/logger"
    "github.com/ncobase/ncore/logging/logger/config"
)

// Basic setup
cleanup, err := logger.New(&config.Config{
    Level:  4, // Info level
    Format: "json",
    Output: "stdout",
})
if err != nil {
    panic(err)
}
defer cleanup()

// Logging
ctx := context.Background()
logger.Info(ctx, "Application started")
logger.WithFields(ctx, logrus.Fields{
    "user_id": "123",
    "action":  "login",
}).Info("User login")
```

## Data Desensitization

Automatically protects sensitive data in logs with fixed-length masking:

```go
// Secure configuration (recommended)
&config.Config{
    Level:  4,
    Format: "json",
    Output: "file",
    OutputFile: "./logs/app.log",
    Desensitization: &config.Desensitization{
        Enabled:         true,
        UseFixedLength:  true,  // All sensitive data → "********"
        FixedMaskLength: 8,
        MaskChar:        "*",
    },
}

// Usage - sensitive fields automatically masked
logger.WithFields(ctx, logrus.Fields{
    "username": "john",
    "password": "secret123",     // → "********"
    "email":    "john@test.com", // → "********"
    "token":    "eyJhbGci...",   // → "********"
}).Info("User authenticated")
```

### Deep Structure Support

Automatically processes nested objects, arrays, and maps:

```go
type User struct {
    Username string            `json:"username"`
    Password string            `json:"password"`
    Profile  map[string]string `json:"profile"`
    APIKeys  []string          `json:"api_keys"`
}

user := User{
    Username: "john",
    Password: "secret",
    Profile:  map[string]string{"email": "john@test.com"},
    APIKeys:  []string{"sk_test_123", "pk_live_456"},
}

// All nested sensitive data automatically masked
logger.WithFields(ctx, logrus.Fields{
    "user": user, // Deep structure processed
}).Info("User created")
```

## Configuration

### File Output

```go
&config.Config{
    Output:     "file",
    OutputFile: "./logs/app.log", // Daily rotation
}
```

### Search Engines

```go
// Elasticsearch
Elasticsearch: &config.Elasticsearch{
    Addresses: []string{"http://localhost:9200"},
    Username:  "elastic",
    Password:  "password",
}

// OpenSearch  
OpenSearch: &config.OpenSearch{
    Addresses: []string{"https://localhost:9200"},
    Username:  "admin",
    Password:  "admin",
}

// Meilisearch
Meilisearch: &config.Meilisearch{
    Host:   "http://localhost:7700",
    APIKey: "masterKey",
}
```

### Custom Desensitization

```go
Desensitization: &config.Desensitization{
    Enabled:               true,
    UseFixedLength:        true,
    FixedMaskLength:       8,
    SensitiveFields:       []string{"password", "token", "secret", "api_key"},
    CustomPatterns:        []string{`\b\d{4}-\d{4}-\d{4}-\d{4}\b`}, // Credit cards
    EnableDefaultPatterns: true,  // Enable built-in patterns (credit cards, emails, etc.)
    ExactFieldMatch:       false, // false: fuzzy match, true: exact match
}
```

**Field Matching Modes:**

- `ExactFieldMatch: false` (default): Fuzzy match - `"password"` matches `"user_password"`, `"password_hash"`
- `ExactFieldMatch: true`: Exact match - `"password"` only matches `"password"`

## Request Tracing

```go
func HandleRequest(w http.ResponseWriter, r *http.Request) {
    ctx, traceID := logger.EnsureTraceID(r.Context())
    w.Header().Set("X-Trace-ID", traceID)
    
    logger.Info(ctx, "Request started")
    // All logs in this context include the same trace ID
    processRequest(ctx)
    logger.Info(ctx, "Request completed")
}
```

## Production Configuration

```yaml
logger:
  level: 4
  format: json
  output: file
  output_file: /var/log/app.log
  
  desensitization:
    enabled: true
    use_fixed_length: true
    fixed_mask_length: 8
    enable_default_patterns: true # Built-in patterns for credit cards, emails, etc.
    
  elasticsearch:
    addresses: ["http://es:9200"]
    username: elastic
    password: ${ES_PASSWORD}
```

## API Reference

```go
// Initialization
func New(c *config.Config) (func(), error)

// Logging
func Debug/Info/Warn/Error/Fatal/Panic(ctx context.Context, args ...any)
func Debugf/Infof/Warnf/Errorf/Fatalf/Panicf(ctx context.Context, format string, args ...any)
func WithFields(ctx context.Context, fields logrus.Fields) *logrus.Entry

// Tracing
func EnsureTraceID(ctx context.Context) (context.Context, string)
```

## Log Levels

| Level | Value | Usage               |
|-------|-------|---------------------|
| Trace | 6     | Detailed debugging  |
| Debug | 5     | Debug information   |
| Info  | 4     | General information |
| Warn  | 3     | Warnings            |
| Error | 2     | Errors              |
| Fatal | 1     | Critical errors     |
| Panic | 0     | System panic        |
