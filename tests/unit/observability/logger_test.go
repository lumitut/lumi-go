package observability_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/lumitut/lumi-go/internal/observability/logger"
	"github.com/lumitut/lumi-go/tests/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestMain(m *testing.M) {
	setup.VerifyNoLeaks(m)
}

func TestLoggerInitialize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  logger.Config
		wantErr bool
	}{
		{
			name: "default config",
			config: logger.Config{
				Level:             "info",
				Format:            "json",
				Development:       false,
				DisableCaller:     false,
				DisableStacktrace: false,
				SampleInitial:     100,
				SampleThereafter:  100,
			},
			wantErr: false,
		},
		{
			name: "development config",
			config: logger.Config{
				Level:             "debug",
				Format:            "console",
				Development:       true,
				DisableCaller:     false,
				DisableStacktrace: false,
				SampleInitial:     100,
				SampleThereafter:  100,
			},
			wantErr: false,
		},
		{
			name: "production config",
			config: logger.Config{
				Level:             "error",
				Format:            "json",
				Development:       false,
				DisableCaller:     true,
				DisableStacktrace: true,
				SampleInitial:     1000,
				SampleThereafter:  1000,
			},
			wantErr: false,
		},
		{
			name: "invalid level",
			config: logger.Config{
				Level:  "invalid",
				Format: "json",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := logger.Initialize(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Test that we can log after initialization
				logger.Info(context.Background(), "test message")
				logger.Sync()
			}
		})
	}
}

func TestLoggerLevels(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer

	// Create custom logger config that writes to buffer
	config := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	encoder := zapcore.NewJSONEncoder(config)
	writer := zapcore.AddSync(&buf)
	core := zapcore.NewCore(encoder, writer, zapcore.DebugLevel)

	// Initialize logger with custom core
	customLogger := zap.New(core)
	logger.SetGlobal(customLogger)

	ctx := context.Background()

	// Test different log levels
	logger.Debug(ctx, "debug message", zap.String("key", "value"))
	logger.Info(ctx, "info message", zap.Int("count", 42))
	logger.Warn(ctx, "warn message", zap.Bool("flag", true))
	logger.Error(ctx, "error message", nil, zap.Float64("rate", 0.95))

	// Parse and verify log output
	lines := bytes.Split(buf.Bytes(), []byte("\n"))

	// Should have 4 log lines (one for each level)
	assert.GreaterOrEqual(t, len(lines), 4)

	// Verify each log line
	for i, expectedLevel := range []string{"debug", "info", "warn", "error"} {
		if i < len(lines) && len(lines[i]) > 0 {
			var logEntry map[string]interface{}
			err := json.Unmarshal(lines[i], &logEntry)
			require.NoError(t, err)
			assert.Equal(t, expectedLevel, logEntry["level"])
		}
	}
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
		opts     logger.RedactOptions
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
			opts: logger.RedactOptions{
				Patterns:    []string{`"api_key":\s*"[^"]+"`},
				Replacement: `"api_key":"[REDACTED]"`,
			},
			expected: `{"api_key":"[REDACTED]","public_key":"pk_test_xyz789"}`,
		},
		{
			name:     "invalid JSON",
			input:    `not json`,
			opts:     logger.DefaultRedactOptions(),
			expected: `not json`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := logger.RedactJSON(tt.input, tt.opts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRedactString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "email in text",
			input:    "Contact us at support@example.com for help",
			expected: "Contact us at [REDACTED] for help",
		},
		{
			name:     "credit card in text",
			input:    "Card ending in 4111111111111111",
			expected: "Card ending in [REDACTED]",
		},
		{
			name:     "SSN in text",
			input:    "SSN: 123-45-6789",
			expected: "SSN: [REDACTED]",
		},
		{
			name:     "no sensitive data",
			input:    "This is a normal message",
			expected: "This is a normal message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := logger.RedactString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRedactMap(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "redact sensitive fields",
			input: map[string]interface{}{
				"username": "john",
				"password": "secret123",
				"email":    "john@example.com",
			},
			expected: map[string]interface{}{
				"username": "john",
				"password": "[REDACTED]",
				"email":    "[REDACTED]",
			},
		},
		{
			name: "nested map",
			input: map[string]interface{}{
				"user": map[string]interface{}{
					"id":       123,
					"password": "secret",
				},
			},
			expected: map[string]interface{}{
				"user": map[string]interface{}{
					"id":       123,
					"password": "[REDACTED]",
				},
			},
		},
		{
			name: "array in map",
			input: map[string]interface{}{
				"tokens": []string{"token1", "token2"},
			},
			expected: map[string]interface{}{
				"tokens": "[REDACTED]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := logger.RedactMap(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark logger performance
func BenchmarkLogger(b *testing.B) {
	err := logger.Initialize(logger.Config{
		Level:             "info",
		Format:            "json",
		DisableCaller:     true,
		DisableStacktrace: true,
	})
	require.NoError(b, err)
	defer logger.Sync()

	ctx := context.Background()

	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(ctx, "benchmark message",
				zap.Int("iteration", i),
				zap.String("key", "value"),
			)
		}
	})

	b.Run("InfoWithContext", func(b *testing.B) {
		ctx := context.WithValue(ctx, logger.RequestIDKey, "bench-123")
		ctx = context.WithValue(ctx, logger.UserIDKey, "user-456")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(ctx, "benchmark message with context",
				zap.Int("iteration", i),
			)
		}
	})

	b.Run("RedactJSON", func(b *testing.B) {
		json := `{"email":"user@example.com","password":"secret","data":"value"}`
		opts := logger.DefaultRedactOptions()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = logger.RedactJSON(json, opts)
		}
	})
}
