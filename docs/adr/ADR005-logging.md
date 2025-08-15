# ADR 005: Logging Strategy

## Status
**Accepted** - August 2025

## Context

Effective logging is critical for debugging, monitoring, auditing, and compliance in distributed systems. The lumi-go template requires a comprehensive logging strategy that balances:

- Performance impact on application throughput
- Debugging capabilities for development and production
- Cost of log storage and processing
- Security and privacy requirements (GDPR, PII handling)
- Integration with observability stack
- Developer experience and ease of use

Key requirements for our logging strategy:
- Structured logging for machine parsing
- High performance with minimal allocations
- Correlation with traces and metrics
- Contextual information propagation
- Log sampling and filtering capabilities
- Sensitive data redaction
- Multi-destination output (stdout, files, collectors)
- Different verbosity levels per environment
- Standardized format across all services

The logging strategy must align with our observability stack (OpenTelemetry, Prometheus, Jaeger) and support our compliance requirements.

## Decision

We have adopted **Uber's Zap** (go.uber.org/zap) as our structured logging library with a comprehensive logging strategy.

### Core Components:

1. **Zap Logger** - High-performance structured logging
   - Zero-allocation JSON encoder for production
   - Configurable encoders (JSON, console)
   - Field-based structured logging
   - Leveled logging with sampling

2. **Log Levels** - Semantic meaning for each level
   - **DEBUG**: Detailed diagnostic information
   - **INFO**: Important business events
   - **WARN**: Potentially harmful situations
   - **ERROR**: Error events that allow continued operation
   - **FATAL**: Severe errors requiring termination

3. **Correlation Strategy** - Unified context propagation
   - Request ID (UUID per request)
   - Trace ID (OpenTelemetry trace)
   - Span ID (OpenTelemetry span)
   - User ID (authenticated user)
   - Service metadata (name, version, environment)

4. **Output Destinations** - Environment-specific routing
   - Development: Console encoder to stdout
   - Production: JSON encoder to stdout → OTEL Collector
   - Audit logs: Separate file/stream with retention

### Logging Architecture:

```
┌─────────────┐
│ Application │
└──────┬──────┘
       │
   Zap Logger
       │
┌──────▼──────┐
│  Structured │
│    Fields   │
└──────┬──────┘
       │
┌──────▼──────┐
│   Encoder   │ ← JSON (prod) / Console (dev)
└──────┬──────┘
       │
┌──────▼──────┐
│   Output    │ → stdout / file / multi
└──────┬──────┘
       │
┌──────▼──────┐
│ OTEL Collector │ → Loki / CloudWatch / etc.
└─────────────┘
```

### Standard Log Fields:

```go
type LogContext struct {
    // Request context
    RequestID  string `json:"request_id"`
    TraceID    string `json:"trace_id"`
    SpanID     string `json:"span_id"`
    
    // User context
    UserID     string `json:"user_id,omitempty"`
    TenantID   string `json:"tenant_id,omitempty"`
    
    // Service context
    Service    string `json:"service"`
    Version    string `json:"version"`
    Env        string `json:"environment"`
    
    // HTTP context
    Method     string `json:"method,omitempty"`
    Path       string `json:"path,omitempty"`
    StatusCode int    `json:"status_code,omitempty"`
    Duration   int64  `json:"duration_ms,omitempty"`
    
    // Error context
    Error      string `json:"error,omitempty"`
    ErrorCode  string `json:"error_code,omitempty"`
    StackTrace string `json:"stack_trace,omitempty"`
}
```

## Consequences

### Positive Consequences

- **Performance**: Zero-allocation in hot paths
- **Structured Data**: Easy parsing and querying
- **Correlation**: Unified view across logs, traces, metrics
- **Flexibility**: Multiple encoders and outputs
- **Type Safety**: Compile-time field validation
- **Sampling**: Reduced volume in production
- **Integration**: Native OpenTelemetry support

### Negative Consequences

- **Learning Curve**: Different API from fmt/log packages
- **Verbosity**: More code for structured fields
- **Configuration Complexity**: Many options to understand
- **Migration Effort**: Existing fmt.Print statements need updating
- **Field Standardization**: Requires discipline across teams

### Mitigations

- Provide logging helpers and middleware
- Create standard field constants
- Generate logging code from OpenAPI/protobuf
- Establish logging guidelines and review process
- Implement automatic PII detection and redaction

## Alternatives Considered

### Option 1: Standard Library (log/slog)

**Pros:**
- Part of standard library (Go 1.21+)
- No external dependencies
- Structured logging support
- Familiar API
- Good performance

**Cons:**
- Newer, less mature (added in Go 1.21)
- Fewer features than Zap
- Less ecosystem support
- No sampling capabilities
- Weaker performance than Zap

**Reason not chosen:** Less mature with fewer features, and performance not as optimized as Zap.

### Option 2: Logrus

**Pros:**
- Popular and mature
- Easy to use
- Extensive formatter options
- Hook system for extensibility
- Good documentation

**Cons:**
- Performance overhead (3-10x slower than Zap)
- Not optimized for zero allocation
- In maintenance mode (not actively developed)
- Synchronous by default
- Global logger anti-pattern

**Reason not chosen:** Performance overhead unacceptable for high-throughput services.

### Option 3: Zerolog

**Pros:**
- Zero allocation design
- Excellent performance
- CBOR encoding option
- Clean API
- Context chaining

**Cons:**
- Less popular than Zap
- Smaller ecosystem
- Different API paradigm
- Less documentation
- Fewer integration options

**Reason not chosen:** Smaller community and ecosystem compared to Zap.

### Option 4: apex/log

**Pros:**
- Clean, simple API
- Handler interface for flexibility
- Built-in handlers for various services
- Good for Lambda/serverless

**Cons:**
- Less performant than Zap
- Smaller community
- Less active development
- Limited structured logging
- No sampling support

**Reason not chosen:** Performance and feature limitations for our scale.

## Implementation Guidelines

### Logging Patterns

#### 1. Logger Initialization

```go
func NewLogger(env string) *zap.Logger {
    var config zap.Config
    
    if env == "production" {
        config = zap.NewProductionConfig()
        config.Sampling = &zap.SamplingConfig{
            Initial:    100,
            Thereafter: 100,
        }
    } else {
        config = zap.NewDevelopmentConfig()
        config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
    }
    
    config.OutputPaths = []string{"stdout"}
    config.InitialFields = map[string]interface{}{
        "service": "lumi-go",
        "version": version.Version,
        "env":     env,
    }
    
    logger, _ := config.Build()
    return logger
}
```

#### 2. Context Propagation

```go
func LoggerFromContext(ctx context.Context) *zap.Logger {
    logger := zap.L()
    
    if requestID := ctx.Value("request_id"); requestID != nil {
        logger = logger.With(zap.String("request_id", requestID.(string)))
    }
    
    if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
        logger = logger.With(
            zap.String("trace_id", span.SpanContext().TraceID().String()),
            zap.String("span_id", span.SpanContext().SpanID().String()),
        )
    }
    
    return logger
}
```

#### 3. Error Logging

```go
func LogError(ctx context.Context, err error, msg string, fields ...zap.Field) {
    logger := LoggerFromContext(ctx)
    
    if stackErr, ok := err.(stackTracer); ok {
        fields = append(fields, zap.String("stack_trace", fmt.Sprintf("%+v", stackErr.StackTrace())))
    }
    
    logger.Error(msg, append(fields, zap.Error(err))...)
}
```

### PII Redaction

```go
type RedactingEncoder struct {
    zapcore.Encoder
    patterns []*regexp.Regexp
}

func (r *RedactingEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
    // Redact sensitive fields
    for i, field := range fields {
        if isSensitiveField(field.Key) {
            fields[i].String = "[REDACTED]"
        }
    }
    return r.Encoder.EncodeEntry(entry, fields)
}
```

### Log Levels by Environment

| Environment | Console | Level | Sampling | Destination |
|-------------|---------|-------|----------|-------------|
| Development | Yes | DEBUG | None | stdout |
| Testing | No | WARN | None | buffer |
| Staging | No | INFO | 10% | stdout → OTEL |
| Production | No | INFO | 1% | stdout → OTEL |

### Audit Logging

Separate audit logs for compliance:

```go
type AuditLogger interface {
    LogAccess(ctx context.Context, resource, action string, allowed bool)
    LogDataChange(ctx context.Context, entity, operation string, before, after interface{})
    LogSecurityEvent(ctx context.Context, event string, details map[string]interface{})
}
```

### Performance Guidelines

1. **Use fields, not string concatenation**
   ```go
   // Good
   logger.Info("user login", zap.String("user_id", userID))
   
   // Bad
   logger.Info("user login for user: " + userID)
   ```

2. **Avoid logging in hot paths**
   - Use sampling for high-frequency operations
   - Aggregate metrics instead of logging each event

3. **Lazy evaluation for expensive operations**
   ```go
   if ce := logger.Check(zap.DebugLevel, "expensive log"); ce != nil {
       ce.Write(zap.String("data", expensiveOperation()))
   }
   ```

## Migration Path

1. **Phase 1**: Set up Zap logger with standard configuration
2. **Phase 2**: Add context propagation middleware
3. **Phase 3**: Migrate fmt.Print statements to structured logging
4. **Phase 4**: Implement PII redaction
5. **Phase 5**: Set up log aggregation with OTEL Collector

## Related

- ADR 004: Observability Stack
- ADR 006: Error Handling
- ADR 007: Security and Compliance
- [Zap Documentation](https://pkg.go.dev/go.uber.org/zap)
- [OpenTelemetry Logging](https://opentelemetry.io/docs/specs/otel/logs/)
- [Structured Logging Best Practices](https://www.sumologic.com/blog/structured-logging-best-practices/)
