package config_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/lumitut/lumi-go/internal/config"
	"github.com/lumitut/lumi-go/tests/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Check for goroutine leaks
	setup.VerifyNoLeaks(m)
}

func TestConfigLoad(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		envVars     map[string]string
		wantErr     bool
		errContains string
		validate    func(t *testing.T, cfg *config.Config)
	}{
		{
			name: "default configuration",
			envVars: map[string]string{
				"SERVICE_NAME": "test-service",
				"ENVIRONMENT":  "development",
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config) {
				assert.Equal(t, "test-service", cfg.Service.Name)
				assert.Equal(t, "development", cfg.Service.Environment)
				assert.Equal(t, "info", cfg.Service.LogLevel)
				assert.Equal(t, "8080", cfg.Server.HTTPPort)
				assert.Equal(t, 15*time.Second, cfg.Server.HTTPReadTimeout)
				assert.True(t, cfg.Observability.MetricsEnabled)
			},
		},
		{
			name: "production configuration",
			envVars: map[string]string{
				"SERVICE_NAME":       "prod-service",
				"ENVIRONMENT":        "production",
				"LOG_LEVEL":          "warn",
				"HTTP_PORT":          "3000",
				"ENABLE_PPROF":       "false",
				"CORS_ENABLED":       "true",
				"CORS_ALLOW_ORIGINS": "https://app.example.com,https://api.example.com",
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config) {
				assert.Equal(t, "prod-service", cfg.Service.Name)
				assert.Equal(t, "production", cfg.Service.Environment)
				assert.Equal(t, "warn", cfg.Service.LogLevel)
				assert.Equal(t, "3000", cfg.Server.HTTPPort)
				assert.False(t, cfg.Server.EnablePProf)
				assert.True(t, cfg.Middleware.CORSEnabled)
				assert.Contains(t, cfg.Middleware.CORSAllowOrigins, "https://app.example.com")
				assert.Contains(t, cfg.Middleware.CORSAllowOrigins, "https://api.example.com")
			},
		},
		{
			name: "database configuration",
			envVars: map[string]string{
				"DB_HOST":           "db.example.com",
				"DB_PORT":           "5433",
				"DB_USER":           "dbuser",
				"DB_PASSWORD":       "dbpass",
				"DB_NAME":           "mydb",
				"DB_SSL_MODE":       "require",
				"DB_MAX_OPEN_CONNS": "50",
				"DB_MAX_IDLE_CONNS": "10",
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config) {
				assert.Equal(t, "db.example.com", cfg.Database.Host)
				assert.Equal(t, "5433", cfg.Database.Port)
				assert.Equal(t, "dbuser", cfg.Database.User)
				assert.Equal(t, "dbpass", cfg.Database.Password)
				assert.Equal(t, "mydb", cfg.Database.Database)
				assert.Equal(t, "require", cfg.Database.SSLMode)
				assert.Equal(t, 50, cfg.Database.MaxOpenConns)
				assert.Equal(t, 10, cfg.Database.MaxIdleConns)
			},
		},
		{
			name: "redis configuration",
			envVars: map[string]string{
				"REDIS_HOST":      "redis.example.com",
				"REDIS_PORT":      "6380",
				"REDIS_PASSWORD":  "redispass",
				"REDIS_DB":        "2",
				"REDIS_POOL_SIZE": "20",
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config) {
				assert.Equal(t, "redis.example.com", cfg.Redis.Host)
				assert.Equal(t, "6380", cfg.Redis.Port)
				assert.Equal(t, "redispass", cfg.Redis.Password)
				assert.Equal(t, 2, cfg.Redis.DB)
				assert.Equal(t, 20, cfg.Redis.PoolSize)
			},
		},
		{
			name: "observability configuration",
			envVars: map[string]string{
				"LOG_LEVEL":                   "debug",
				"LOG_FORMAT":                  "text",
				"METRICS_ENABLED":             "false",
				"TRACING_ENABLED":             "true",
				"TRACING_SAMPLING":            "0.5",
				"OTEL_EXPORTER_OTLP_ENDPOINT": "otel.example.com:4317",
				"OTEL_EXPORTER_OTLP_INSECURE": "false",
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config) {
				assert.Equal(t, "debug", cfg.Observability.LogLevel)
				assert.Equal(t, "text", cfg.Observability.LogFormat)
				assert.False(t, cfg.Observability.MetricsEnabled)
				assert.True(t, cfg.Observability.TracingEnabled)
				assert.Equal(t, 0.5, cfg.Observability.TracingSampling)
				assert.Equal(t, "otel.example.com:4317", cfg.Observability.TracingEndpoint)
				assert.False(t, cfg.Observability.TracingInsecure)
			},
		},
		{
			name: "middleware configuration",
			envVars: map[string]string{
				"RATE_LIMIT_ENABLED":   "true",
				"RATE_LIMIT_RATE":      "100",
				"RATE_LIMIT_BURST":     "20",
				"RATE_LIMIT_TYPE":      "user",
				"RECOVERY_STACK_TRACE": "false",
				"TRUSTED_PROXIES":      "10.0.0.0/8,192.168.0.0/16",
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config) {
				assert.True(t, cfg.Middleware.RateLimitEnabled)
				assert.Equal(t, 100, cfg.Middleware.RateLimitRate)
				assert.Equal(t, 20, cfg.Middleware.RateLimitBurst)
				assert.Equal(t, "user", cfg.Middleware.RateLimitType)
				assert.False(t, cfg.Middleware.RecoveryStackTrace)
				assert.Contains(t, cfg.Middleware.TrustedProxies, "10.0.0.0/8")
				assert.Contains(t, cfg.Middleware.TrustedProxies, "192.168.0.0/16")
			},
		},
		{
			name: "feature flags",
			envVars: map[string]string{
				"FEATURE_NEW_API":  "true",
				"FEATURE_BETA":     "true",
				"MAINTENANCE_MODE": "true",
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config) {
				assert.True(t, cfg.Features.EnableNewAPI)
				assert.True(t, cfg.Features.EnableBetaFeatures)
				assert.True(t, cfg.Features.MaintenanceMode)
			},
		},
		{
			name: "duration parsing",
			envVars: map[string]string{
				"HTTP_READ_TIMEOUT":         "30s",
				"HTTP_WRITE_TIMEOUT":        "1m",
				"GRACEFUL_SHUTDOWN_TIMEOUT": "45s",
				"DB_CONN_MAX_LIFETIME":      "10m",
				"REDIS_DIAL_TIMEOUT":        "10s",
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config) {
				assert.Equal(t, 30*time.Second, cfg.Server.HTTPReadTimeout)
				assert.Equal(t, 1*time.Minute, cfg.Server.HTTPWriteTimeout)
				assert.Equal(t, 45*time.Second, cfg.Server.GracefulShutdownTimeout)
				assert.Equal(t, 10*time.Minute, cfg.Database.ConnMaxLifetime)
				assert.Equal(t, 10*time.Second, cfg.Redis.DialTimeout)
			},
		},
		{
			name: "invalid environment",
			envVars: map[string]string{
				"ENVIRONMENT": "invalid",
			},
			wantErr:     true,
			errContains: "invalid environment",
		},
		{
			name: "invalid log level",
			envVars: map[string]string{
				"LOG_LEVEL": "invalid",
			},
			wantErr:     true,
			errContains: "invalid log level",
		},
		{
			name: "invalid port",
			envVars: map[string]string{
				"HTTP_PORT": "99999",
			},
			wantErr:     true,
			errContains: "port must be between",
		},
		{
			name: "invalid rate limit type",
			envVars: map[string]string{
				"RATE_LIMIT_ENABLED": "true",
				"RATE_LIMIT_TYPE":    "invalid",
			},
			wantErr:     true,
			errContains: "invalid rate limit type",
		},
		{
			name: "empty service name",
			envVars: map[string]string{
				"SERVICE_NAME": "",
			},
			wantErr:     true,
			errContains: "service name is required",
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange: Set up environment
			cleanup := setup.TestEnv(t)
			defer cleanup()

			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Act: Load configuration
			cfg, err := config.Load()

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, cfg)
				if tt.validate != nil {
					tt.validate(t, cfg)
				}
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		modify      func(*config.Config)
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid configuration",
			modify:  func(c *config.Config) {},
			wantErr: false,
		},
		{
			name: "empty service name",
			modify: func(c *config.Config) {
				c.Service.Name = ""
			},
			wantErr:     true,
			errContains: "service name is required",
		},
		{
			name: "invalid environment",
			modify: func(c *config.Config) {
				c.Service.Environment = "testing"
			},
			wantErr:     true,
			errContains: "invalid environment",
		},
		{
			name: "invalid log level",
			modify: func(c *config.Config) {
				c.Service.LogLevel = "verbose"
			},
			wantErr:     true,
			errContains: "invalid log level",
		},
		{
			name: "invalid HTTP port - too low",
			modify: func(c *config.Config) {
				c.Server.HTTPPort = "0"
			},
			wantErr:     true,
			errContains: "port must be between",
		},
		{
			name: "invalid HTTP port - too high",
			modify: func(c *config.Config) {
				c.Server.HTTPPort = "70000"
			},
			wantErr:     true,
			errContains: "port must be between",
		},
		{
			name: "invalid HTTP port - not a number",
			modify: func(c *config.Config) {
				c.Server.HTTPPort = "abc"
			},
			wantErr: true,
		},
		{
			name: "invalid RPC port",
			modify: func(c *config.Config) {
				c.Server.RPCPort = "invalid"
			},
			wantErr: true,
		},
		{
			name: "invalid rate limit type",
			modify: func(c *config.Config) {
				c.Middleware.RateLimitEnabled = true
				c.Middleware.RateLimitType = "unknown"
			},
			wantErr:     true,
			errContains: "invalid rate limit type",
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange: Create base config
			cfg := &config.Config{
				Service: config.ServiceConfig{
					Name:        "test-service",
					Version:     "1.0.0",
					Environment: "development",
					LogLevel:    "info",
				},
				Server: config.ServerConfig{
					HTTPPort: "8080",
					RPCPort:  "8081",
				},
				Middleware: config.MiddlewareConfig{
					RateLimitType: "ip",
				},
			}

			// Act: Modify config
			tt.modify(cfg)
			err := cfg.Validate()

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConfigLogConfig(t *testing.T) {
	// Arrange
	cleanup := setup.TestEnv(t)
	defer cleanup()

	cfg := &config.Config{
		Service: config.ServiceConfig{
			Name:        "test-service",
			Version:     "1.0.0",
			Environment: "test",
			LogLevel:    "info",
		},
		Server: config.ServerConfig{
			HTTPPort: "8080",
			RPCPort:  "8081",
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			Database: "testdb",
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: "6379",
		},
		Observability: config.ObservabilityConfig{
			MetricsEnabled: true,
			TracingEnabled: false,
		},
		Middleware: config.MiddlewareConfig{
			CORSEnabled:      false,
			RateLimitEnabled: true,
			RateLimitRate:    60,
		},
		Features: config.FeaturesConfig{
			MaintenanceMode: false,
		},
	}

	// Act & Assert - should not panic
	require.NotPanics(t, func() {
		cfg.LogConfig(context.Background())
	})
}

func TestParseHelpers(t *testing.T) {
	t.Parallel()

	t.Run("parseList", func(t *testing.T) {
		tests := []struct {
			input    string
			expected []string
		}{
			{"", []string{}},
			{"single", []string{"single"}},
			{"one,two,three", []string{"one", "two", "three"}},
			{"  one , two , three  ", []string{"one", "two", "three"}},
			{"one,,three", []string{"one", "three"}},
			{",,,", []string{}},
		}

		for _, tt := range tests {
			os.Setenv("TEST_LIST", tt.input)

			// Use one of the list fields to verify parsing
			os.Setenv("CORS_ALLOW_ORIGINS", tt.input)
			cfg2, err := config.Load()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, cfg2.Middleware.CORSAllowOrigins)
		}
	})

	t.Run("parseHeaders", func(t *testing.T) {
		tests := []struct {
			input    string
			expected map[string]string
		}{
			{"", map[string]string{}},
			{"key=value", map[string]string{"key": "value"}},
			{"key1=value1,key2=value2", map[string]string{"key1": "value1", "key2": "value2"}},
			{"  key1 = value1 , key2 = value2  ", map[string]string{"key1": "value1", "key2": "value2"}},
			{"invalid", map[string]string{}},
			{"key=", map[string]string{}},
		}

		for _, tt := range tests {
			os.Setenv("OTEL_EXPORTER_OTLP_HEADERS", tt.input)
			cfg, err := config.Load()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, cfg.Observability.TracingHeaders)
		}
	})
}

// Benchmark configuration loading
func BenchmarkConfigLoad(b *testing.B) {
	cleanup := setup.TestEnv(&testing.T{})
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = config.Load()
	}
}

func BenchmarkConfigValidate(b *testing.B) {
	cfg := &config.Config{
		Service: config.ServiceConfig{
			Name:        "bench-service",
			Version:     "1.0.0",
			Environment: "production",
			LogLevel:    "info",
		},
		Server: config.ServerConfig{
			HTTPPort: "8080",
			RPCPort:  "8081",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg.Validate()
	}
}
