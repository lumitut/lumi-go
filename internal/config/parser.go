// Package config provides configuration parsing using viper
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Parser handles configuration parsing using viper
type Parser struct {
	viper *viper.Viper
}

// NewParser creates a new configuration parser
func NewParser() *Parser {
	v := viper.New()

	// Set configuration defaults
	v.SetDefault("service.name", "lumi-go")
	v.SetDefault("service.version", "unknown")
	v.SetDefault("service.environment", "development")
	v.SetDefault("service.logLevel", "info")

	v.SetDefault("server.httpPort", "8080")
	v.SetDefault("server.rpcPort", "8081")
	v.SetDefault("server.httpReadTimeout", "15s")
	v.SetDefault("server.httpWriteTimeout", "15s")
	v.SetDefault("server.httpIdleTimeout", "60s")
	v.SetDefault("server.rpcReadTimeout", "30s")
	v.SetDefault("server.rpcWriteTimeout", "30s")
	v.SetDefault("server.gracefulShutdownTimeout", "30s")
	v.SetDefault("server.enablePProf", false)
	v.SetDefault("server.pprofPort", "6060")

	// Client defaults (simplified - just connection strings)
	v.SetDefault("clients.database.enabled", false)
	v.SetDefault("clients.database.url", "")
	v.SetDefault("clients.redis.enabled", false)
	v.SetDefault("clients.redis.url", "")
	v.SetDefault("clients.tracing.enabled", false)
	v.SetDefault("clients.tracing.endpoint", "")

	v.SetDefault("observability.logLevel", "info")
	v.SetDefault("observability.logFormat", "json")
	v.SetDefault("observability.logOutput", "stdout")
	v.SetDefault("observability.logSampling", true)
	v.SetDefault("observability.metricsEnabled", true)
	v.SetDefault("observability.metricsPort", "9090")
	v.SetDefault("observability.metricsPath", "/metrics")

	v.SetDefault("middleware.corsEnabled", false)
	v.SetDefault("middleware.corsAllowMethods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("middleware.corsAllowHeaders", []string{"Origin", "Content-Type", "Accept", "Authorization"})
	v.SetDefault("middleware.corsExposeHeaders", []string{"X-Request-ID"})
	v.SetDefault("middleware.corsAllowCredentials", false)
	v.SetDefault("middleware.corsMaxAge", "12h")
	v.SetDefault("middleware.rateLimitEnabled", true)
	v.SetDefault("middleware.rateLimitRate", 60)
	v.SetDefault("middleware.rateLimitBurst", 10)
	v.SetDefault("middleware.rateLimitType", "ip")
	v.SetDefault("middleware.recoveryStackTrace", true)
	v.SetDefault("middleware.recoveryStackSize", 4096)
	v.SetDefault("middleware.recoveryPrintStack", false)
	v.SetDefault("middleware.requestIDHeader", "X-Request-ID")
	v.SetDefault("middleware.trustAllProxies", false)
	v.SetDefault("middleware.logSkipPaths", []string{"/health", "/ready", "/metrics"})
	v.SetDefault("middleware.logRequestBody", false)
	v.SetDefault("middleware.logResponseBody", false)
	v.SetDefault("middleware.logSlowThreshold", "1s")

	v.SetDefault("features.enableNewAPI", false)
	v.SetDefault("features.enableBetaFeatures", false)
	v.SetDefault("features.maintenanceMode", false)

	return &Parser{viper: v}
}

// LoadConfig loads configuration from file and environment variables
func (p *Parser) LoadConfig(configFile string) (*Config, error) {
	// Set config file
	p.viper.SetConfigFile(configFile)
	p.viper.SetConfigType("json")

	// Enable environment variable override
	// Environment variables will be in format: LUMI_SECTION_KEY
	// For example: LUMI_SERVICE_NAME, LUMI_DATABASE_HOST
	p.viper.SetEnvPrefix("LUMI")
	p.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	p.viper.AutomaticEnv()

	// Read config file
	if err := p.viper.ReadInConfig(); err != nil {
		// It's okay if config file doesn't exist, we'll use defaults and env vars
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal config
	var cfg Config
	if err := p.viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// GetViper returns the underlying viper instance for advanced usage
func (p *Parser) GetViper() *viper.Viper {
	return p.viper
}
