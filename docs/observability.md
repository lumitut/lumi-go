# Observability Package

## Overview

The observability package provides comprehensive instrumentation for the lumi-go service, implementing the three pillars of observability: **Logging**, **Metrics**, and **Tracing**.

## Components

### 1. Logger (`/logger`)

Structured logging with Zap providing:
- JSON-formatted logs with correlation IDs
- PII redaction and compliance
- Performance tracking
- Audit logging
- Context propagation

### 2. Metrics (Coming in Phase 2)

Prometheus metrics including:
- HTTP request metrics (rate, latency, errors)
- Business metrics
- Custom application metrics
- Resource utilization

### 3. Tracing (Coming in Phase 2)

OpenTelemetry distributed tracing:
- Request flow visualization
- Latency analysis
- Dependency mapping
- Error tracking

## Quick Start

### Initialize Logging

```go
import "github.com/lumitut/lumi-go/internal/observability/logger"

func main() {
    // Configure logger
    cfg := logger.Config{
        Level:       "info",
        Format:      "json",
        Development: false,
    }
    
    // Initialize
    if err := logger.Initialize(cfg); err != nil {
        log.Fatal(err)
    }
    defer logger.Sync()
    
    // Use logger
    logger.Info(ctx, "Service started")
}
```

### Use with Middleware

```go
import "github.com/lumitut/lumi-go/internal/middleware"

func setupRouter() *gin.Engine {
    r := gin.New()
    
    // Add correlation tracking
    r.Use(middleware.Correlation())
    
    // Add request logging
    r.Use(middleware.Logging("/health", "/metrics"))
    
    return r
}
```

### Context Propagation

```go
func handleRequest(c *gin.Context) {
    // Context has correlation IDs from middleware
    ctx := c.Request.Context()
    
    // Logs include correlation fields automatically
    logger.Info(ctx, "Processing request")
    
    // Pass context through layers
    result, err := service.Process(ctx, data)
    if err != nil {
        logger.Error(ctx, "Processing failed", err)
        return
    }
    
    // Audit important operations
    logger.Audit(ctx, "DATA_PROCESSED", "resource:123", "success")
}
```

## Configuration

### Environment Variables

```bash
# Logging
LOG_LEVEL=info              # debug, info, warn, error, fatal
LOG_FORMAT=json             # json or console
LOG_DEVELOPMENT=false       # Development mode
LOG_DISABLE_CALLER=false    # Include caller info
LOG_DISABLE_STACKTRACE=false # Include stack traces
LOG_SAMPLE_INITIAL=100      # Initial sampling rate
LOG_SAMPLE_THEREAFTER=100   # Ongoing sampling rate

# Service metadata (included in all logs)
SERVICE_NAME=my-service
SERVICE_VERSION=v1.0.0
ENVIRONMENT=production
```

## Features

### PII Redaction

Automatic redaction of sensitive data:

```go
// Input
logData := "User email: john@example.com, SSN: 123-45-6789"

// Automatically redacted in logs
// Output: "User email: [REDACTED_EMAIL], SSN: [REDACTED_SSN]"
```

Supported patterns:
- Email addresses
- Phone numbers
- SSNs
- Credit card numbers
- API keys and tokens
- Passwords
- JWTs

### Correlation Tracking

Automatic correlation across service boundaries:

```go
// Headers automatically propagated
X-Request-ID: abc-123
X-Correlation-ID: xyz-789
X-Trace-ID: trace-456

// Available in context
ctx.Value(logger.RequestIDKey)     // "abc-123"
ctx.Value(logger.CorrelationIDKey) // "xyz-789"
ctx.Value(logger.TraceIDKey)       // "trace-456"
```

### Performance Tracking

```go
start := time.Now()
result := expensiveOperation()

logger.Performance(ctx, "expensive_op", time.Since(start),
    zap.Int("items", len(result)),
)
// Logs: {"operation":"expensive_op","duration_ms":123.45,"items":100}
```

### Audit Logging

```go
logger.Audit(ctx, "USER_DELETED", "user:123", "success",
    zap.String("admin", adminID),
    zap.String("reason", "GDPR request"),
)
// Logs: {"audit":"true","action":"USER_DELETED","resource":"user:123","result":"success",...}
```

## Best Practices

### DO

✅ Use structured fields instead of string formatting
```go
// Good
logger.Info(ctx, "User created", 
    zap.String("user_id", userID),
    zap.String("role", role),
)

// Bad
logger.Info(ctx, fmt.Sprintf("User %s created with role %s", userID, role))
```

✅ Include context in all service/repository methods
```go
func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    logger.Debug(ctx, "Fetching user", zap.String("id", id))
    // ...
}
```

✅ Use appropriate log levels
```go
logger.Debug(ctx, "Cache miss")           // Diagnostic info
logger.Info(ctx, "Order placed")          // Business events
logger.Warn(ctx, "Rate limit near")       // Warnings
logger.Error(ctx, "Payment failed", err)  // Errors
```

✅ Log once at the appropriate layer
```go
// Log at the boundary where the error is handled
if err := db.Query(ctx, query); err != nil {
    logger.Error(ctx, "Database query failed", err)
    return nil, fmt.Errorf("query failed: %w", err)
}
```

### DON'T

❌ Log sensitive data
```go
// Bad
logger.Info(ctx, "Login", zap.String("password", password))

// Good
logger.Info(ctx, "Login attempt", zap.String("username", username))
```

❌ Use string concatenation for messages
```go
// Bad
logger.Info(ctx, "User " + userID + " logged in")

// Good  
logger.Info(ctx, "User logged in", zap.String("user_id", userID))
```

❌ Log in tight loops
```go
// Bad
for _, item := range items {
    logger.Debug(ctx, "Processing item", zap.Any("item", item))
    process(item)
}

// Good
logger.Debug(ctx, "Processing items", zap.Int("count", len(items)))
for _, item := range items {
    process(item)
}
logger.Debug(ctx, "Items processed", zap.Int("count", len(items)))
```

## Testing

### Unit Tests

```bash
go test ./internal/observability/...
```

### Benchmarks

```bash
go test -bench=. ./internal/observability/logger
```

Expected performance:
- Simple log: ~150ns
- With 5 fields: ~500ns
- PII redaction: ~5μs

## Troubleshooting

### No Logs Appearing

1. Check log level: `LOG_LEVEL=debug`
2. Verify initialization: Ensure `logger.Initialize()` is called
3. Check output: Logs go to stdout/stderr by default

### Missing Correlation IDs

1. Ensure correlation middleware runs first
2. Verify context is passed through call chain
3. Check header names match configuration

### High Memory Usage

1. Enable sampling in production
2. Reduce log verbosity
3. Check for logging in loops

### Performance Impact

1. Disable caller info: `LOG_DISABLE_CALLER=true`
2. Disable stack traces: `LOG_DISABLE_STACKTRACE=true`  
3. Increase sampling rates

## References

- [Logging Contract](./logging.md)
- [Uber Zap Documentation](https://github.com/uber-go/zap)
- [OpenTelemetry Specification](https://opentelemetry.io/docs/)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/naming/)
