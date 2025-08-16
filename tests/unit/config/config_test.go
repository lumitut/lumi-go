package config_test

import (
	"os"
	"testing"

	"github.com/lumitut/lumi-go/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &config.Config{
				Service: config.ServiceConfig{
					Name:        "test-service",
					Environment: "development",
					LogLevel:    "info",
				},
				Server: config.ServerConfig{
					HTTPPort: "8080",
					RPCPort:  "8081",
				},
				Middleware: config.MiddlewareConfig{
					RateLimitEnabled: true,
					RateLimitType:    "ip",
				},
			},
			wantErr: false,
		},
		{
			name: "missing service name",
			config: &config.Config{
				Service: config.ServiceConfig{
					Environment: "development",
					LogLevel:    "info",
				},
				Server: config.ServerConfig{
					HTTPPort: "8080",
					RPCPort:  "8081",
				},
			},
			wantErr: true,
			errMsg:  "service name is required",
		},
		{
			name: "invalid environment",
			config: &config.Config{
				Service: config.ServiceConfig{
					Name:        "test-service",
					Environment: "invalid",
					LogLevel:    "info",
				},
				Server: config.ServerConfig{
					HTTPPort: "8080",
					RPCPort:  "8081",
				},
			},
			wantErr: true,
			errMsg:  "invalid environment",
		},
		{
			name: "invalid log level",
			config: &config.Config{
				Service: config.ServiceConfig{
					Name:        "test-service",
					Environment: "development",
					LogLevel:    "invalid",
				},
				Server: config.ServerConfig{
					HTTPPort: "8080",
					RPCPort:  "8081",
				},
			},
			wantErr: true,
			errMsg:  "invalid log level",
		},
		{
			name: "invalid HTTP port",
			config: &config.Config{
				Service: config.ServiceConfig{
					Name:        "test-service",
					Environment: "development",
					LogLevel:    "info",
				},
				Server: config.ServerConfig{
					HTTPPort: "99999",
					RPCPort:  "8081",
				},
			},
			wantErr: true,
			errMsg:  "invalid HTTP port",
		},
		{
			name: "invalid rate limit type",
			config: &config.Config{
				Service: config.ServiceConfig{
					Name:        "test-service",
					Environment: "development",
					LogLevel:    "info",
				},
				Server: config.ServerConfig{
					HTTPPort: "8080",
					RPCPort:  "8081",
				},
				Middleware: config.MiddlewareConfig{
					RateLimitEnabled: true,
					RateLimitType:    "invalid",
				},
			},
			wantErr: true,
			errMsg:  "invalid rate limit type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_GetDatabaseURL(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		wantURL string
		wantOK  bool
	}{
		{
			name: "database enabled with URL",
			config: &config.Config{
				Clients: config.ClientsConfig{
					Database: config.DatabaseClientConfig{
						Enabled: true,
						URL:     "postgres://localhost:5432/testdb",
					},
				},
			},
			wantURL: "postgres://localhost:5432/testdb",
			wantOK:  true,
		},
		{
			name: "database disabled",
			config: &config.Config{
				Clients: config.ClientsConfig{
					Database: config.DatabaseClientConfig{
						Enabled: false,
						URL:     "postgres://localhost:5432/testdb",
					},
				},
			},
			wantURL: "",
			wantOK:  false,
		},
		{
			name: "database enabled without URL",
			config: &config.Config{
				Clients: config.ClientsConfig{
					Database: config.DatabaseClientConfig{
						Enabled: true,
						URL:     "",
					},
				},
			},
			wantURL: "",
			wantOK:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, ok := tt.config.GetDatabaseURL()
			assert.Equal(t, tt.wantURL, url)
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}

func TestConfig_GetRedisURL(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		wantURL string
		wantOK  bool
	}{
		{
			name: "redis enabled with URL",
			config: &config.Config{
				Clients: config.ClientsConfig{
					Redis: config.RedisClientConfig{
						Enabled: true,
						URL:     "redis://localhost:6379/0",
					},
				},
			},
			wantURL: "redis://localhost:6379/0",
			wantOK:  true,
		},
		{
			name: "redis disabled",
			config: &config.Config{
				Clients: config.ClientsConfig{
					Redis: config.RedisClientConfig{
						Enabled: false,
						URL:     "redis://localhost:6379/0",
					},
				},
			},
			wantURL: "",
			wantOK:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, ok := tt.config.GetRedisURL()
			assert.Equal(t, tt.wantURL, url)
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}

func TestConfig_IsTracingEnabled(t *testing.T) {
	tests := []struct {
		name   string
		config *config.Config
		want   bool
	}{
		{
			name: "tracing enabled with endpoint",
			config: &config.Config{
				Clients: config.ClientsConfig{
					Tracing: config.TracingClientConfig{
						Enabled:  true,
						Endpoint: "localhost:4317",
					},
				},
			},
			want: true,
		},
		{
			name: "tracing disabled",
			config: &config.Config{
				Clients: config.ClientsConfig{
					Tracing: config.TracingClientConfig{
						Enabled:  false,
						Endpoint: "localhost:4317",
					},
				},
			},
			want: false,
		},
		{
			name: "tracing enabled without endpoint",
			config: &config.Config{
				Clients: config.ClientsConfig{
					Tracing: config.TracingClientConfig{
						Enabled:  true,
						Endpoint: "",
					},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.IsTracingEnabled()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLoadWithDefaults(t *testing.T) {
	cfg, err := config.LoadWithDefaults()
	require.NoError(t, err)

	// Check defaults are set
	assert.Equal(t, "lumi-go", cfg.Service.Name)
	assert.Equal(t, "development", cfg.Service.Environment)
	assert.Equal(t, "8080", cfg.Server.HTTPPort)
	assert.Equal(t, "8081", cfg.Server.RPCPort)
	assert.False(t, cfg.Clients.Database.Enabled)
	assert.False(t, cfg.Clients.Redis.Enabled)
	assert.False(t, cfg.Clients.Tracing.Enabled)
}

func TestConfigFromEnvironment(t *testing.T) {
	// Save current working directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	
	// Change to project root for config file access
	os.Chdir("../../../")
	
	// Set environment variables
	os.Setenv("LUMI_SERVICE_NAME", "test-service")
	os.Setenv("LUMI_SERVER_HTTPPORT", "9090")
	os.Setenv("LUMI_CLIENTS_DATABASE_ENABLED", "true")
	os.Setenv("LUMI_CLIENTS_DATABASE_URL", "postgres://testdb")
	defer func() {
		os.Unsetenv("LUMI_SERVICE_NAME")
		os.Unsetenv("LUMI_SERVER_HTTPPORT")
		os.Unsetenv("LUMI_CLIENTS_DATABASE_ENABLED")
		os.Unsetenv("LUMI_CLIENTS_DATABASE_URL")
	}()

	// Use Load with the config file path to test environment overrides
	cfg, err := config.Load("cmd/server/schema/lumi.json")
	require.NoError(t, err)

	// Check environment overrides
	assert.Equal(t, "test-service", cfg.Service.Name)
	assert.Equal(t, "9090", cfg.Server.HTTPPort)
	assert.True(t, cfg.Clients.Database.Enabled)
	assert.Equal(t, "postgres://testdb", cfg.Clients.Database.URL)
}
