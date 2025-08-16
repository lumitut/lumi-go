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

	// Database configuration
	Database DatabaseConfig `json:"database" mapstructure:"database"`

	// Redis configuration
	Redis RedisConfig `json:"redis" mapstructure:"redis"`

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

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string        `json:"host" mapstructure:"host"`
	Port            string        `json:"port" mapstructure:"port"`
	User            string        `json:"user" mapstructure:"user"`
	Password        string        `json:"password" mapstructure:"password"`
	Database        string        `json:"database" mapstructure:"database"`
	SSLMode         string        `json:"sslMode" mapstructure:"sslMode"`
	MaxOpenConns    int           `json:"maxOpenConns" mapstructure:"maxOpenConns"`
	MaxIdleConns    int           `json:"maxIdleConns" mapstructure:"maxIdleConns"`
	ConnMaxLifetime time.Duration `json:"connMaxLifetime" mapstructure:"connMaxLifetime"`
	ConnMaxIdleTime time.Duration `json:"connMaxIdleTime" mapstructure:"connMaxIdleTime"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host         string        `json:"host" mapstructure:"host"`
	Port         string        `json:"port" mapstructure:"port"`
	Password     string        `json:"password" mapstructure:"password"`
	DB           int           `json:"db" mapstructure:"db"`
	PoolSize     int           `json:"poolSize" mapstructure:"poolSize"`
	MinIdleConns int           `json:"minIdleConns" mapstructure:"minIdleConns"`
	DialTimeout  time.Duration `json:"dialTimeout" mapstructure:"dialTimeout"`
	ReadTimeout  time.Duration `json:"readTimeout" mapstructure:"readTimeout"`
	WriteTimeout time.Duration `json:"writeTimeout" mapstructure:"writeTimeout"`
	PoolTimeout  time.Duration `json:"poolTimeout" mapstructure:"poolTimeout"`
	IdleTimeout  time.Duration `json:"idleTimeout" mapstructure:"idleTimeout"`
	MaxRetries   int           `json:"maxRetries" mapstructure:"maxRetries"`
}

// ObservabilityConfig holds observability configuration
type ObservabilityConfig struct {
	// Logging
	LogLevel       string `json:"logLevel" mapstructure:"logLevel"`
	LogFormat      string `json:"logFormat" mapstructure:"logFormat"`
	LogOutput      string `json:"logOutput" mapstructure:"logOutput"`
	LogSampling    bool   `json:"logSampling" mapstructure:"logSampling"`
	LogDevelopment bool   `json:"logDevelopment" mapstructure:"logDevelopment"`

	// Metrics
	MetricsEnabled bool   `json:"metricsEnabled" mapstructure:"metricsEnabled"`
	MetricsPort    string `json:"metricsPort" mapstructure:"metricsPort"`
	MetricsPath    string `json:"metricsPath" mapstructure:"metricsPath"`

	// Tracing
	TracingEnabled  bool              `json:"tracingEnabled" mapstructure:"tracingEnabled"`
	TracingSampling float64           `json:"tracingSampling" mapstructure:"tracingSampling"`
	TracingEndpoint string            `json:"tracingEndpoint" mapstructure:"tracingEndpoint"`
	TracingInsecure bool              `json:"tracingInsecure" mapstructure:"tracingInsecure"`
	TracingHeaders  map[string]string `json:"tracingHeaders" mapstructure:"tracingHeaders"`
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
		zap.String("db_host", c.Database.Host),
		zap.String("db_port", c.Database.Port),
		zap.String("db_name", c.Database.Database),
		zap.String("redis_host", c.Redis.Host),
		zap.String("redis_port", c.Redis.Port),
		zap.Bool("metrics_enabled", c.Observability.MetricsEnabled),
		zap.Bool("tracing_enabled", c.Observability.TracingEnabled),
		zap.Bool("cors_enabled", c.Middleware.CORSEnabled),
		zap.Bool("rate_limit_enabled", c.Middleware.RateLimitEnabled),
		zap.Int("rate_limit_rate", c.Middleware.RateLimitRate),
		zap.Bool("maintenance_mode", c.Features.MaintenanceMode),
	)
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
