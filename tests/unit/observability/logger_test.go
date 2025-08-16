package observability_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/lumitut/lumi-go/internal/observability/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestLoggerInitialization(t *testing.T) {
	tests := []struct {
		name      string
		config    logger.Config
		wantError bool
	}{
		{
			name: "valid JSON config",
			config: logger.Config{
				Level:       "info",
				Format:      "json",
				Development: false,
			},
			wantError: false,
		},
		{
			name: "valid console config",
			config: logger.Config{
				Level:       "debug",
				Format:      "console",
				Development: true,
			},
			wantError: false,
		},
		{
			name: "invalid log level",
			config: logger.Config{
				Level:  "invalid",
				Format: "json",
			},
			wantError: true,
		},
		{
			name: "invalid format",
			config: logger.Config{
				Level:  "info",
				Format: "invalid",
			},
			wantError: true,
		},
		{
			name: "development mode",
			config: logger.Config{
				Level:       "debug",
				Format:      "console",
				Development: true,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := logger.Initialize(tt.config)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify logger is initialized
				assert.NotNil(t, logger.Get())
				assert.NotNil(t, logger.Sugar())
			}
		})
	}
}

func TestLoggerLevels(t *testing.T) {
	// Initialize logger with debug level to capture all logs
	err := logger.Initialize(logger.Config{
		Level:       "debug",
		Format:      "json",
		Development: false,
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Test that different log levels work without panicking
	logger.Debug(ctx, "debug message", zap.String("key", "value"))
	logger.Info(ctx, "info message", zap.Int("count", 42))
	logger.Warn(ctx, "warn message", zap.Bool("flag", true))
	logger.Error(ctx, "error message", nil, zap.Float64("rate", 0.95))

	// Verify logger is functioning
	assert.NotNil(t, logger.Get())
}

func TestLoggerContext(t *testing.T) {
	// Initialize logger
	err := logger.Initialize(logger.Config{
		Level:  "debug",
		Format: "json",
	})
	require.NoError(t, err)
	defer logger.Sync()

	// Create context with values
	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.RequestIDKey, "test-request-123")
	ctx = context.WithValue(ctx, logger.CorrelationIDKey, "test-correlation-456")
	ctx = context.WithValue(ctx, logger.UserIDKey, "user-789")
	ctx = context.WithValue(ctx, logger.TenantIDKey, "tenant-abc")
	ctx = context.WithValue(ctx, logger.TraceIDKey, "trace-xyz")
	ctx = context.WithValue(ctx, logger.SpanIDKey, "span-123")

	// Log with context
	logger.Info(ctx, "message with context")

	// Create logger with context
	log := logger.WithContext(ctx)
	assert.NotNil(t, log)
}

func TestLoggerAudit(t *testing.T) {
	// Initialize logger
	err := logger.Initialize(logger.Config{
		Level:  "info",
		Format: "json",
	})
	require.NoError(t, err)
	defer logger.Sync()

	ctx := context.Background()

	// Test audit logging
	logger.Audit(ctx, "USER_LOGIN", "user:123", "success",
		zap.String("ip", "192.168.1.1"),
		zap.String("user_agent", "Mozilla/5.0"),
	)

	logger.Audit(ctx, "USER_DELETE", "user:456", "failure",
		zap.String("reason", "permission_denied"),
	)
}

func TestLoggerPerformance(t *testing.T) {
	// Initialize logger
	err := logger.Initialize(logger.Config{
		Level:  "info",
		Format: "json",
	})
	require.NoError(t, err)
	defer logger.Sync()

	ctx := context.Background()

	// Test performance logging
	logger.Performance(ctx, "database_query", 150*time.Millisecond,
		zap.String("query", "SELECT * FROM users"),
		zap.Int("rows", 100),
	)

	logger.Performance(ctx, "api_call", 500*time.Millisecond,
		zap.String("endpoint", "https://api.example.com/data"),
		zap.Int("status", 200),
	)
}

func TestLoggerPanic(t *testing.T) {
	// Initialize logger
	err := logger.Initialize(logger.Config{
		Level:  "info",
		Format: "json",
	})
	require.NoError(t, err)
	defer logger.Sync()

	// Test that Fatal panics
	assert.Panics(t, func() {
		logger.Fatal(context.Background(), "fatal message")
	})
}

func TestRedactJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		opts     logger.RedactOption
		expected string
	}{
		{
			name:     "redact password",
			input:    `{"username":"john","password":"secret123"}`,
			opts:     logger.DefaultRedactOptions(),
			expected: `{"username":"john","password":"[REDACTED]"}`,
		},
		{
			name:     "redact email",
			input:    `{"email":"john@example.com","name":"John"}`,
			opts:     logger.DefaultRedactOptions(),
			expected: `{"email":"[REDACTED]","name":"John"}`,
		},
		{
			name:     "redact token",
			input:    `{"auth_token":"Bearer abc123xyz","data":"value"}`,
			opts:     logger.DefaultRedactOptions(),
			expected: `{"auth_token":"[REDACTED]","data":"value"}`,
		},
		{
			name:     "redact credit card",
			input:    `{"card_number":"4111111111111111","amount":100}`,
			opts:     logger.DefaultRedactOptions(),
			expected: `{"card_number":"[REDACTED]","amount":100}`,
		},
		{
			name:     "redact SSN",
			input:    `{"ssn":"123-45-6789","name":"John"}`,
			opts:     logger.DefaultRedactOptions(),
			expected: `{"ssn":"[REDACTED]","name":"John"}`,
		},
		{
			name:     "nested redaction",
			input:    `{"user":{"email":"test@example.com","password":"secret"}}`,
			opts:     logger.DefaultRedactOptions(),
			expected: `{"user":{"email":"[REDACTED]","password":"[REDACTED]"}}`,
		},
		{
			name:     "array redaction",
			input:    `{"users":[{"email":"user1@example.com"},{"email":"user2@example.com"}]}`,
			opts:     logger.DefaultRedactOptions(),
			expected: `{"users":[{"email":"[REDACTED]"},{"email":"[REDACTED]"}]}`,
		},
		{
			name:  "custom pattern",
			input: `{"api_key":"sk_live_abc123","public_key":"pk_test_xyz789"}`,
			opts: logger.RedactOption{
				APIKeys: true,
			},
			expected: `{"api_key":"[REDACTED]","public_key":"pk_test_xyz789"}`,
		},
		{
			name:     "invalid JSON",
			input:    `{not valid json}`,
			opts:     logger.DefaultRedactOptions(),
			expected: `{not valid json}`, // Should return original on error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := logger.RedactJSON(tt.input, tt.opts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRedactHeaders(t *testing.T) {
	headers := map[string][]string{
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer secret-token-123"},
		"X-API-Key":     {"api-key-secret"},
		"User-Agent":    {"Mozilla/5.0"},
	}

	redacted := logger.RedactHeaders(headers)

	// Check that sensitive headers are redacted
	assert.Equal(t, []string{"application/json"}, redacted["Content-Type"])
	assert.Equal(t, []string{"[REDACTED]"}, redacted["Authorization"])
	assert.Equal(t, []string{"[REDACTED]"}, redacted["X-API-Key"])
	assert.Equal(t, []string{"Mozilla/5.0"}, redacted["User-Agent"])
}

func TestLoggerWithFields(t *testing.T) {
	// Initialize logger
	err := logger.Initialize(logger.Config{
		Level:  "debug",
		Format: "json",
	})
	require.NoError(t, err)

	// Create logger with fields
	log := logger.WithFields(
		zap.String("service", "test-service"),
		zap.Int("version", 1),
	)

	assert.NotNil(t, log)
}

func TestLoggerError(t *testing.T) {
	// Initialize logger
	err := logger.Initialize(logger.Config{
		Level:  "debug",
		Format: "json",
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Test logging with error
	testErr := assert.AnError
	logger.Error(ctx, "test error", testErr, zap.String("detail", "additional info"))

	// Test logging without error
	logger.Error(ctx, "error message without error object", nil, zap.Int("code", 500))
}

func TestLoggerOutputCapture(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create a custom zap core that writes to buffer
	config := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		MessageKey:     "message",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}

	encoder := zapcore.NewJSONEncoder(config)
	writer := zapcore.AddSync(&buf)
	core := zapcore.NewCore(encoder, writer, zapcore.DebugLevel)

	// Note: In a real implementation, you would need to modify the logger
	// package to support custom cores for testing. For now, we just verify
	// the core can be created without errors.
	assert.NotNil(t, core)
	assert.NotNil(t, encoder)
	assert.NotNil(t, writer)
}
