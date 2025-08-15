# Logging Contract

> **Document Status:** Production-ready logging standards for lumi-go services  
> **Last Updated:** December 2024  
> **Compliance:** GDPR, HIPAA, SOC2

## Overview

This document defines the structured logging contract for all lumi-go services. All services MUST adhere to these standards to ensure consistent observability, debugging capabilities, and compliance.

## Core Principles

1. **Structured Over Unstructured**: All logs MUST be structured (JSON format)
2. **Correlation Over Isolation**: All logs MUST include correlation identifiers
3. **Performance Over Verbosity**: Logging must not degrade service performance
4. **Privacy Over Transparency**: PII must be redacted or excluded
5. **Actionable Over Informational**: Each log level must drive specific actions

## Log Levels

### Level Definitions

| Level | Numeric | Use Case | Action Required | Example |
|-------|---------|----------|-----------------|----------|
| FATAL | 60 | System is unusable, immediate shutdown | Page on-call immediately | Database connection lost, out of memory |
| ERROR | 50 | Operation failed, system degraded | Alert on-call within 15 mins | Payment processing failed, API timeout |
| WARN | 40 | Unexpected behavior, system recovering | Review within 24 hours | Rate limit approaching, cache miss |
| INFO | 30 | Normal operations, business events | No action, for audit trail | User login, order placed, job completed |
| DEBUG | 20 | Detailed diagnostic information | Developer investigation | Function entry/exit, variable states |
| TRACE | 10 | Most detailed information | Deep debugging only | Every method call, all data |

### Environment Configuration

```bash
# Production
LOG_LEVEL=info
LOG_FORMAT=json

# Staging  
LOG_LEVEL=debug
LOG_FORMAT=json

# Development
LOG_LEVEL=debug
LOG_FORMAT=console  # Human-readable
```

## Required Fields

Every log entry MUST include these fields:

```json
{
  "timestamp": "2024-12-01T10:30:45.123Z",  // RFC3339 with milliseconds
  "level": "info",                           // Lowercase level
  "message": "User authentication successful", // Human-readable message
  "service": "auth-service",                 // Service name from SERVICE_NAME env
  "version": "v1.2.3",                        // Service version from SERVICE_VERSION env
  "env": "production",                        // Environment from ENVIRONMENT env
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000", // Request correlation ID
  "logger": "auth.handler"                   // Logger name/component
}
```

## Correlation Fields

Correlation fields link related log entries across services:

```json
{
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000", // Flows across services
  "request_id": "req_1234567890",              // Unique per HTTP request
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736", // OpenTelemetry trace ID
  "span_id": "00f067aa0ba902b7",                // OpenTelemetry span ID
  "parent_span_id": "00f067aa0ba902b6",         // Parent span for distributed tracing
  "user_id": "user_123",                        // User identifier (if authenticated)
  "tenant_id": "tenant_456",                    // Tenant in multi-tenant systems
  "session_id": "sess_abc123"                   // User session identifier
}
```

## HTTP Request Logging

All HTTP requests must be logged with these fields:

```json
{
  "method": "POST",
  "path": "/api/v1/users",
  "status": 201,
  "latency": 145.23,           // Milliseconds
  "latency_human": "145.23ms", // Human-readable
  "ip": "192.168.1.1",
  "user_agent": "Mozilla/5.0...",
  "request_size": 1024,        // Bytes
  "response_size": 2048,       // Bytes
  "query": "filter=active",    // Query string (redacted)
  "errors": ["validation error"] // If any
}
```

## Error Logging

Errors must include:

```json
{
  "error": "connection refused",              // Error message
  "error_code": "DB_CONN_REFUSED",            // Structured error code
  "error_type": "DatabaseError",              // Error classification
  "stacktrace": "...",                        // Full stack trace (ERROR level and above)
  "error_details": {                          // Additional context
    "host": "db.example.com",
    "port": 5432,
    "retry_count": 3
  }
}
```

## Performance Logging

Performance-critical operations:

```json
{
  "operation": "database_query",
  "duration": 45.67,              // Milliseconds
  "duration_human": "45.67ms",
  "query": "SELECT * FROM users", // Sanitized query
  "rows_affected": 150,
  "cache_hit": false,
  "slow_query": true              // If exceeds threshold
}
```

## Audit Logging

All state-changing operations must generate audit logs:

```json
{
  "audit": "true",                            // Marker for audit logs
  "action": "USER_DELETED",                   // Action performed
  "resource": "user:123",                     // Resource affected
  "result": "success",                        // success/failure
  "actor": "admin@example.com",               // Who performed the action
  "actor_ip": "192.168.1.100",                // Actor's IP address
  "audit_timestamp": "2024-12-01T10:30:45Z",  // When it happened
  "changes": {                                // What changed (no PII)
    "status": {"old": "active", "new": "deleted"}
  }
}
```

## PII Handling

### Never Log

- Passwords, tokens, API keys
- Credit card numbers
- Social Security Numbers
- Full names (use user IDs)
- Email addresses (hash if needed)
- Phone numbers
- Addresses
- Medical information
- Financial data

### Redaction Examples

```json
// Bad
{"email": "john.doe@example.com", "password": "secret123"}

// Good
{"email": "[REDACTED_EMAIL]", "password": "[REDACTED]"}

// Better
{"user_id": "user_123", "auth_method": "password"}
```

### Redaction Patterns

The logging system automatically redacts:

- Email addresses: `[REDACTED_EMAIL]`
- Credit cards: `[REDACTED_CC]`
- SSNs: `[REDACTED_SSN]`
- Phone numbers: `[REDACTED_PHONE]`
- JWTs: `[REDACTED_JWT]`
- API keys: `[REDACTED_API_KEY]`
- Passwords: `[REDACTED_PASSWORD]`
- HTTP Authorization headers: `[REDACTED]`

## Sampling Strategy

### Configuration

```go
type SamplingConfig struct {
    Initial    int  // Logs per second before sampling
    Thereafter int  // Sample rate after initial
}
```

### Environment Settings

| Environment | Initial | Thereafter | Rationale |
|-------------|---------|------------|------------|
| Production | 100 | 100 | Balance between visibility and cost |
| Staging | 1000 | 500 | Higher visibility for testing |
| Development | -1 | -1 | No sampling (all logs) |

## Usage Examples

### Basic Logging

```go
// Import
import "github.com/lumitut/lumi-go/internal/observability/logger"

// Initialize at startup
func main() {
    cfg := logger.Config{
        Level:       os.Getenv("LOG_LEVEL"),
        Format:      os.Getenv("LOG_FORMAT"),
        Development: os.Getenv("ENVIRONMENT") == "development",
    }
    if err := logger.Initialize(cfg); err != nil {
        log.Fatal("Failed to initialize logger", err)
    }
    defer logger.Sync()
}

// Simple logging
logger.Info(ctx, "User logged in",
    zap.String("user_id", userID),
    zap.String("method", "oauth"),
)

// Error logging
if err := processPayment(order); err != nil {
    logger.Error(ctx, "Payment processing failed", err,
        zap.String("order_id", order.ID),
        zap.Float64("amount", order.Amount),
    )
    return err
}

// Performance logging
start := time.Now()
result := expensiveOperation()
logger.Performance(ctx, "expensive_operation", time.Since(start),
    zap.Int("items_processed", len(result)),
)

// Audit logging
logger.Audit(ctx, "USER_DELETED", fmt.Sprintf("user:%s", userID), "success",
    zap.String("actor", adminID),
    zap.String("reason", "account_closure"),
)
```

### With Gin Middleware

```go
import (
    "github.com/lumitut/lumi-go/internal/middleware"
)

func setupRouter() *gin.Engine {
    r := gin.New()
    
    // Add correlation middleware first
    r.Use(middleware.Correlation())
    
    // Add logging middleware
    r.Use(middleware.Logging(
        "/health",  // Skip health checks
        "/metrics", // Skip metrics endpoint
    ))
    
    return r
}
```

### Context Propagation

```go
// In HTTP handler
func handleRequest(c *gin.Context) {
    // Context already has correlation IDs from middleware
    ctx := c.Request.Context()
    
    // Use context for logging
    logger.Info(ctx, "Processing request")
    
    // Pass context to service layer
    result, err := service.Process(ctx, data)
    if err != nil {
        logger.Error(ctx, "Service processing failed", err)
        c.JSON(500, gin.H{"error": "internal_error"})
        return
    }
    
    logger.Info(ctx, "Request completed successfully")
    c.JSON(200, result)
}

// In service layer
func (s *Service) Process(ctx context.Context, data Data) (Result, error) {
    // Context flows through, maintaining correlation
    logger.Debug(ctx, "Starting processing",
        zap.Any("data", data),
    )
    
    // ... processing logic ...
    
    return result, nil
}
```

## Log Aggregation

### Local Development

Logs are written to stdout/stderr and can be viewed:

```bash
# View all logs
docker-compose logs -f app

# Filter by level
docker-compose logs app | jq 'select(.level == "error")'

# Filter by correlation ID
docker-compose logs app | jq 'select(.correlation_id == "xxx")'
```

### Production

Logs flow through:

1. Application → stdout/stderr
2. Container runtime → Log driver
3. Fluentd/Fluent Bit → Collection
4. OpenSearch/Loki → Storage
5. Grafana → Visualization

## Retention Policy

| Environment | Retention | Rationale |
|-------------|-----------|------------|
| Production | 30 days | Compliance and debugging |
| Staging | 7 days | Testing validation |
| Development | 1 day | Local debugging only |

## Compliance Considerations

### GDPR

- No PII in logs without legal basis
- Right to erasure must be implementable
- Audit logs may be retained for legal obligations

### HIPAA

- No PHI (Protected Health Information) in logs
- Audit logs required for access to PHI
- Minimum necessary standard applies

### SOC2

- Audit trails for all access and changes
- Log integrity and tamper protection
- Defined retention and disposal procedures

## Performance Impact

### Benchmarks

| Operation | Latency | Allocations |
|-----------|---------|-------------|
| Info log (no fields) | ~150ns | 0 |
| Info log (5 fields) | ~500ns | 0 |
| Error log with stack | ~1μs | 2 |
| Context extraction | ~50ns | 0 |
| PII redaction | ~5μs | 1-3 |

### Best Practices

1. **Pre-allocate fields**: Use `With()` for repeated fields
2. **Avoid string formatting**: Use structured fields instead
3. **Sample in production**: Configure appropriate sampling rates
4. **Async where possible**: Use buffered writers for high throughput
5. **Lazy evaluation**: Use `zap.Stringer` for expensive operations

## Monitoring & Alerts

### Key Metrics

```prometheus
# Log volume by level
log_messages_total{level="error",service="auth-service"}

# Log sampling rate
log_messages_sampled_total / log_messages_total

# Logging latency
log_write_duration_seconds{quantile="0.99"}
```

### Alert Rules

```yaml
# High error rate
alert: HighErrorLogRate
expr: rate(log_messages_total{level="error"}[5m]) > 10
for: 5m

# Logging failure
alert: LoggingSystemDown  
expr: up{job="logger"} == 0
for: 1m
```

## Migration Guide

### From fmt/log to Structured

```go
// Before
log.Printf("User %s logged in from %s", userID, ip)

// After
logger.Info(ctx, "User logged in",
    zap.String("user_id", userID),
    zap.String("ip", ip),
)
```

### Adding to Existing Service

1. Add logger dependency:
   ```bash
   go get github.com/lumitut/lumi-go/internal/observability/logger
   ```

2. Initialize at startup:
   ```go
   logger.Initialize(logger.Config{
       Level: "info",
       Format: "json",
   })
   ```

3. Add middleware:
   ```go
   router.Use(middleware.Correlation())
   router.Use(middleware.Logging())
   ```

4. Replace existing logs gradually

## Troubleshooting

### Common Issues

1. **Missing correlation IDs**: Ensure correlation middleware runs first
2. **PII in logs**: Enable redaction, review log statements
3. **High log volume**: Adjust sampling rates, review log levels
4. **Performance degradation**: Check for string formatting, reduce DEBUG logs
5. **Lost logs**: Verify buffer sizes, check back-pressure handling

## References

- [OpenTelemetry Logging](https://opentelemetry.io/docs/specs/otel/logs/)
- [Zap Performance](https://github.com/uber-go/zap#performance)
- [GDPR Logging Guidelines](https://gdpr.eu/eu-gdpr-personal-data/)
- [NIST Logging Standards](https://nvlpubs.nist.gov/nistpubs/SpecialPublications/NIST.SP.800-92.pdf)
