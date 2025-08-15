// Package logger provides structured logging with Zap
package logger

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ContextKey type for context values
type ContextKey string

const (
	// CorrelationIDKey is the context key for correlation ID
	CorrelationIDKey ContextKey = "correlation_id"
	// RequestIDKey is the context key for request ID
	RequestIDKey ContextKey = "request_id"
	// TraceIDKey is the context key for trace ID
	TraceIDKey ContextKey = "trace_id"
	// SpanIDKey is the context key for span ID
	SpanIDKey ContextKey = "span_id"
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
	// TenantIDKey is the context key for tenant ID
	TenantIDKey ContextKey = "tenant_id"
)

var (
	// global logger instance
	globalLogger *zap.Logger
	// global sugar logger for convenience
	globalSugar *zap.SugaredLogger
)

// Config holds logger configuration
type Config struct {
	// Level is the minimum enabled logging level
	Level string `env:"LOG_LEVEL" default:"info"`
	// Format is the output format (json or console)
	Format string `env:"LOG_FORMAT" default:"json"`
	// Development enables development mode (DPanic logs panic, more human-friendly output)
	Development bool `env:"LOG_DEVELOPMENT" default:"false"`
	// DisableCaller disables caller information
	DisableCaller bool `env:"LOG_DISABLE_CALLER" default:"false"`
	// DisableStacktrace disables stacktrace for errors
	DisableStacktrace bool `env:"LOG_DISABLE_STACKTRACE" default:"false"`
	// SampleInitial is the initial sampling rate (logs per second)
	SampleInitial int `env:"LOG_SAMPLE_INITIAL" default:"100"`
	// SampleThereafter is the sampling rate after initial
	SampleThereafter int `env:"LOG_SAMPLE_THEREAFTER" default:"100"`
	// OutputPaths is the list of output paths
	OutputPaths []string
	// ErrorOutputPaths is the list of error output paths
	ErrorOutputPaths []string
}

// Initialize sets up the global logger
func Initialize(cfg Config) error {
	// Parse log level
	level := zapcore.InfoLevel
	if cfg.Level != "" {
		if err := level.UnmarshalText([]byte(strings.ToLower(cfg.Level))); err != nil {
			return fmt.Errorf("invalid log level %q: %w", cfg.Level, err)
		}
	}

	// Set default format if not specified
	if cfg.Format == "" {
		cfg.Format = "json"
	}

	// Build encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Adjust encoder for console format
	if cfg.Format == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Setup output paths
	if len(cfg.OutputPaths) == 0 {
		cfg.OutputPaths = []string{"stdout"}
	}
	if len(cfg.ErrorOutputPaths) == 0 {
		cfg.ErrorOutputPaths = []string{"stderr"}
	}

	// Build zap config
	zapConfig := zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		Development:       cfg.Development,
		DisableCaller:     cfg.DisableCaller,
		DisableStacktrace: cfg.DisableStacktrace,
		Sampling: &zap.SamplingConfig{
			Initial:    cfg.SampleInitial,
			Thereafter: cfg.SampleThereafter,
		},
		Encoding:         cfg.Format,
		EncoderConfig:    encoderConfig,
		OutputPaths:      cfg.OutputPaths,
		ErrorOutputPaths: cfg.ErrorOutputPaths,
		InitialFields: map[string]interface{}{
			"service": os.Getenv("SERVICE_NAME"),
			"version": os.Getenv("SERVICE_VERSION"),
			"env":     os.Getenv("ENVIRONMENT"),
		},
	}

	// Build logger
	logger, err := zapConfig.Build(
		zap.AddCallerSkip(1), // Skip wrapper functions
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return fmt.Errorf("failed to build logger: %w", err)
	}

	// Set global logger
	globalLogger = logger
	globalSugar = logger.Sugar()

	// Replace zap global logger
	zap.ReplaceGlobals(logger)

	return nil
}

// Get returns the global logger instance
func Get() *zap.Logger {
	if globalLogger == nil {
		// Initialize with default config if not initialized
		_ = Initialize(Config{})
	}
	return globalLogger
}

// Sugar returns the global sugared logger instance
func Sugar() *zap.SugaredLogger {
	if globalSugar == nil {
		// Initialize with default config if not initialized
		_ = Initialize(Config{})
	}
	return globalSugar
}

// WithContext returns a logger with context fields
func WithContext(ctx context.Context) *zap.Logger {
	logger := Get()

	// Add correlation fields from context
	if correlationID := ctx.Value(CorrelationIDKey); correlationID != nil {
		logger = logger.With(zap.String("correlation_id", correlationID.(string)))
	}
	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		logger = logger.With(zap.String("request_id", requestID.(string)))
	}
	if traceID := ctx.Value(TraceIDKey); traceID != nil {
		logger = logger.With(zap.String("trace_id", traceID.(string)))
	}
	if spanID := ctx.Value(SpanIDKey); spanID != nil {
		logger = logger.With(zap.String("span_id", spanID.(string)))
	}
	if userID := ctx.Value(UserIDKey); userID != nil {
		logger = logger.With(zap.String("user_id", userID.(string)))
	}
	if tenantID := ctx.Value(TenantIDKey); tenantID != nil {
		logger = logger.With(zap.String("tenant_id", tenantID.(string)))
	}

	return logger
}

// WithFields returns a logger with additional fields
func WithFields(fields ...zap.Field) *zap.Logger {
	return Get().With(fields...)
}

// Sync flushes any buffered log entries
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// Helper functions for common logging patterns

// Error logs an error with additional context
func Error(ctx context.Context, msg string, err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	WithContext(ctx).Error(msg, fields...)
}

// Info logs an info message with context
func Info(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Info(msg, fields...)
}

// Warn logs a warning with context
func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Warn(msg, fields...)
}

// Debug logs a debug message with context
func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Debug(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Fatal(msg, fields...)
}

// Panic logs a panic message and panics
func Panic(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Panic(msg, fields...)
}

// Audit logs an audit event with mandatory fields
func Audit(ctx context.Context, action string, resource string, result string, fields ...zap.Field) {
	auditFields := []zap.Field{
		zap.String("audit", "true"),
		zap.String("action", action),
		zap.String("resource", resource),
		zap.String("result", result),
		zap.Time("audit_timestamp", time.Now().UTC()),
	}
	auditFields = append(auditFields, fields...)
	WithContext(ctx).Info("audit_event", auditFields...)
}

// Performance logs a performance metric
func Performance(ctx context.Context, operation string, duration time.Duration, fields ...zap.Field) {
	perfFields := []zap.Field{
		zap.String("operation", operation),
		zap.Duration("duration", duration),
		zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
	}
	perfFields = append(perfFields, fields...)
	WithContext(ctx).Info("performance_metric", perfFields...)
}
