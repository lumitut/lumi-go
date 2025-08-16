// Package config provides configuration management for the application
package config

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/lumitut/lumi-go/internal/observability/logger"
	"go.uber.org/zap"
)

// Config holds all application configuration
type Config struct {
	// Service configuration
	Service ServiceConfig `json:"service" mapstructure:"service"`

	// Server configuration
	Server ServerConfig `json:"server" mapstructure:"server"`

	// External clients configuration (optional)
	Clients ClientsConfig `json:"clients" mapstructure:"clients"`

	// Observability configuration
	Observability ObservabilityConfig `json:"observability" mapstructure:"observability"`

	// Middleware configuration
	Middleware MiddlewareConfig `json:"middleware" mapstructure:"middleware"`

	// Feature flags
	Features FeaturesConfig `json:"features" mapstructure:"features"`
}

// ServiceConfig holds service-level configuration
type ServiceConfig struct {
	Name        string `json:"name" mapstructure:"name"`
	Version     string `json:"version" mapstructure:"version"`
	Environment string `json:"environment" mapstructure:"environment"`
	LogLevel    string `json:"logLevel" mapstructure:"logLevel"`
}

// ServerConfig holds HTTP/RPC server configuration
type ServerConfig struct {
	// HTTP server
	HTTPPort         string        `json:"httpPort" mapstructure:"httpPort"`
	HTTPReadTimeout  time.Duration `json:"httpReadTimeout" mapstructure:"httpReadTimeout"`
	HTTPWriteTimeout time.Duration `json:"httpWriteTimeout" mapstructure:"httpWriteTimeout"`
	HTTPIdleTimeout  time.Duration `json:"httpIdleTimeout" mapstructure:"httpIdleTimeout"`

	// RPC server
	RPCPort         string        `json:"rpcPort" mapstructure:"rpcPort"`
	RPCReadTimeout  time.Duration `json:"rpcReadTimeout" mapstructure:"rpcReadTimeout"`
	RPCWriteTimeout time.Duration `json:"rpcWriteTimeout" mapstructure:"rpcWriteTimeout"`

	// Common
	GracefulShutdownTimeout time.Duration `json:"gracefulShutdownTimeout" mapstructure:"gracefulShutdownTimeout"`
	EnablePProf             bool          `json:"enablePProf" mapstructure:"enablePProf"`
	PProfPort               string        `json:"pprofPort" mapstructure:"pprofPort"`
}

// ClientsConfig holds optional external client configurations
type ClientsConfig struct {
	// Database client configuration (simplified)
	Database DatabaseClientConfig `json:"database" mapstructure:"database"`

	// Redis client configuration (simplified)
	Redis RedisClientConfig `json:"redis" mapstructure:"redis"`

	// Tracing client configuration (simplified)
	Tracing TracingClientConfig `json:"tracing" mapstructure:"tracing"`
}

// DatabaseClientConfig holds simplified database client configuration
type DatabaseClientConfig struct {
	Enabled bool   `json:"enabled" mapstructure:"enabled"`
	URL     string `json:"url" mapstructure:"url"` // Connection string
}

// RedisClientConfig holds simplified Redis client configuration
type RedisClientConfig struct {
	Enabled bool   `json:"enabled" mapstructure:"enabled"`
	URL     string `json:"url" mapstructure:"url"` // Connection string
}

// TracingClientConfig holds simplified tracing client configuration
type TracingClientConfig struct {
	Enabled  bool   `json:"enabled" mapstructure:"enabled"`
	Endpoint string `json:"endpoint" mapstructure:"endpoint"` // OTLP endpoint
}

// ObservabilityConfig holds observability configuration
type ObservabilityConfig struct {
	// Logging
	LogLevel       string `json:"logLevel" mapstructure:"logLevel"`
	LogFormat      string `json:"logFormat" mapstructure:"logFormat"`
	LogOutput      string `json:"logOutput" mapstructure:"logOutput"`
	LogSampling    bool   `json:"logSampling" mapstructure:"logSampling"`
	LogDevelopment bool   `json:"logDevelopment" mapstructure:"logDevelopment"`

	// Metrics (service's own metrics)
	MetricsEnabled bool   `json:"metricsEnabled" mapstructure:"metricsEnabled"`
	MetricsPort    string `json:"metricsPort" mapstructure:"metricsPort"`
	MetricsPath    string `json:"metricsPath" mapstructure:"metricsPath"`
}

// MiddlewareConfig holds middleware configuration
type MiddlewareConfig struct {
	// CORS
	CORSEnabled          bool          `json:"corsEnabled" mapstructure:"corsEnabled"`
	CORSAllowOrigins     []string      `json:"corsAllowOrigins" mapstructure:"corsAllowOrigins"`
	CORSAllowMethods     []string      `json:"corsAllowMethods" mapstructure:"corsAllowMethods"`
	CORSAllowHeaders     []string      `json:"corsAllowHeaders" mapstructure:"corsAllowHeaders"`
	CORSExposeHeaders    []string      `json:"corsExposeHeaders" mapstructure:"corsExposeHeaders"`
	CORSAllowCredentials bool          `json:"corsAllowCredentials" mapstructure:"corsAllowCredentials"`
	CORSMaxAge           time.Duration `json:"corsMaxAge" mapstructure:"corsMaxAge"`

	// Rate Limiting
	RateLimitEnabled bool   `json:"rateLimitEnabled" mapstructure:"rateLimitEnabled"`
	RateLimitRate    int    `json:"rateLimitRate" mapstructure:"rateLimitRate"` // requests per minute
	RateLimitBurst   int    `json:"rateLimitBurst" mapstructure:"rateLimitBurst"`
	RateLimitType    string `json:"rateLimitType" mapstructure:"rateLimitType"` // "ip", "user", "api_key"

	// Recovery
	RecoveryStackTrace bool `json:"recoveryStackTrace" mapstructure:"recoveryStackTrace"`
	RecoveryStackSize  int  `json:"recoveryStackSize" mapstructure:"recoveryStackSize"`
	RecoveryPrintStack bool `json:"recoveryPrintStack" mapstructure:"recoveryPrintStack"`

	// Request ID
	RequestIDHeader string `json:"requestIDHeader" mapstructure:"requestIDHeader"`

	// Real IP
	TrustedProxies  []string `json:"trustedProxies" mapstructure:"trustedProxies"`
	TrustAllProxies bool     `json:"trustAllProxies" mapstructure:"trustAllProxies"`

	// Logging
	LogSkipPaths     []string      `json:"logSkipPaths" mapstructure:"logSkipPaths"`
	LogRequestBody   bool          `json:"logRequestBody" mapstructure:"logRequestBody"`
	LogResponseBody  bool          `json:"logResponseBody" mapstructure:"logResponseBody"`
	LogSlowThreshold time.Duration `json:"logSlowThreshold" mapstructure:"logSlowThreshold"`
}

// FeaturesConfig holds feature flags
type FeaturesConfig struct {
	EnableNewAPI       bool `json:"enableNewAPI" mapstructure:"enableNewAPI"`
	EnableBetaFeatures bool `json:"enableBetaFeatures" mapstructure:"enableBetaFeatures"`
	MaintenanceMode    bool `json:"maintenanceMode" mapstructure:"maintenanceMode"`
}

// Load loads configuration using viper from JSON file and environment variables
func Load(configPath string) (*Config, error) {
	// If no config path provided, use default
	if configPath == "" {
		configPath = "cmd/server/schema/lumi.json"
	}

	// Make path absolute if it's relative
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(".", configPath)
	}

	// Create parser
	parser := NewParser()

	// Load configuration
	cfg, err := parser.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Set logDevelopment based on environment if not explicitly set
	if cfg.Service.Environment == "development" {
		cfg.Observability.LogDevelopment = true
	}

	return cfg, nil
}

// LoadWithDefaults loads configuration with defaults (no config file required)
func LoadWithDefaults() (*Config, error) {
	// Create parser
	parser := NewParser()

	// Create empty config and let viper populate with defaults
	var cfg Config
	if err := parser.GetViper().Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal defaults: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate service name
	if c.Service.Name == "" {
		return fmt.Errorf("service name is required")
	}

	// Validate environment
	validEnvs := map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
	}
	if !validEnvs[c.Service.Environment] {
		return fmt.Errorf("invalid environment: %s", c.Service.Environment)
	}

	// Validate log level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}
	if !validLevels[c.Service.LogLevel] {
		return fmt.Errorf("invalid log level: %s", c.Service.LogLevel)
	}

	// Validate ports
	if err := validatePort(c.Server.HTTPPort); err != nil {
		return fmt.Errorf("invalid HTTP port: %w", err)
	}
	if err := validatePort(c.Server.RPCPort); err != nil {
		return fmt.Errorf("invalid RPC port: %w", err)
	}

	// Validate rate limit type
	validRateLimitTypes := map[string]bool{
		"ip":      true,
		"user":    true,
		"api_key": true,
	}
	if c.Middleware.RateLimitEnabled && !validRateLimitTypes[c.Middleware.RateLimitType] {
		return fmt.Errorf("invalid rate limit type: %s", c.Middleware.RateLimitType)
	}

	return nil
}

// LogConfig logs the configuration (with sensitive values redacted)
func (c *Config) LogConfig(ctx context.Context) {
	logger.Info(ctx, "Configuration loaded",
		zap.String("service_name", c.Service.Name),
		zap.String("service_version", c.Service.Version),
		zap.String("environment", c.Service.Environment),
		zap.String("log_level", c.Service.LogLevel),
		zap.String("http_port", c.Server.HTTPPort),
		zap.String("rpc_port", c.Server.RPCPort),
		zap.Bool("database_enabled", c.Clients.Database.Enabled),
		zap.Bool("redis_enabled", c.Clients.Redis.Enabled),
		zap.Bool("tracing_enabled", c.Clients.Tracing.Enabled),
		zap.Bool("metrics_enabled", c.Observability.MetricsEnabled),
		zap.Bool("cors_enabled", c.Middleware.CORSEnabled),
		zap.Bool("rate_limit_enabled", c.Middleware.RateLimitEnabled),
		zap.Int("rate_limit_rate", c.Middleware.RateLimitRate),
		zap.Bool("maintenance_mode", c.Features.MaintenanceMode),
	)
}

// GetDatabaseURL returns the database connection URL if database is enabled
func (c *Config) GetDatabaseURL() (string, bool) {
	if c.Clients.Database.Enabled && c.Clients.Database.URL != "" {
		return c.Clients.Database.URL, true
	}
	return "", false
}

// GetRedisURL returns the Redis connection URL if Redis is enabled
func (c *Config) GetRedisURL() (string, bool) {
	if c.Clients.Redis.Enabled && c.Clients.Redis.URL != "" {
		return c.Clients.Redis.URL, true
	}
	return "", false
}

// IsTracingEnabled returns whether tracing is enabled
func (c *Config) IsTracingEnabled() bool {
	return c.Clients.Tracing.Enabled && c.Clients.Tracing.Endpoint != ""
}

// Helper functions

func validatePort(port string) error {
	p, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	if p < 1 || p > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	return nil
}
