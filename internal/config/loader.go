// Package config provides centralized configuration management
package config

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// ConfigLoader provides methods for loading application configuration
type ConfigLoader struct {
	configPath string
	parser     *Parser
}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{
		parser: NewParser(),
	}
}

// LoadFromFlags loads configuration using command-line flags
func (cl *ConfigLoader) LoadFromFlags() (*Config, error) {
	// Define command-line flags
	configFile := flag.String("config", "", "Path to configuration file (default: cmd/server/schema/lumi.json)")
	envOverride := flag.String("env", "", "Override environment (development/staging/production)")
	flag.Parse()

	// Set config path
	cl.configPath = *configFile
	if cl.configPath == "" {
		cl.configPath = "cmd/server/schema/lumi.json"
	}

	// Load configuration
	cfg, err := cl.parser.LoadConfig(cl.configPath)
	if err != nil {
		return nil, err
	}

	// Override environment if specified
	if *envOverride != "" {
		cfg.Service.Environment = *envOverride
	}

	return cfg, nil
}

// LoadFromFile loads configuration from a specific file
func (cl *ConfigLoader) LoadFromFile(configPath string) (*Config, error) {
	cl.configPath = configPath
	return cl.parser.LoadConfig(configPath)
}

// LoadFromEnvironment loads configuration from environment variables only
func (cl *ConfigLoader) LoadFromEnvironment() (*Config, error) {
	// Create a parser that doesn't require a config file
	parser := NewParser()

	// Don't set a config file, just use defaults and env vars
	var cfg Config
	if err := parser.GetViper().Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to load from environment: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// LoadDefault loads the default configuration
// Priority order: Environment variables > Config file > Defaults
func LoadDefault(ctx context.Context) (*Config, error) {
	loader := NewConfigLoader()

	// Check if running with flags
	if len(os.Args) > 1 {
		return loader.LoadFromFlags()
	}

	// Try to load from default config file
	defaultPath := "cmd/server/schema/lumi.json"

	// Check if we're in the root directory or need to find it
	if _, err := os.Stat(defaultPath); err == nil {
		return loader.LoadFromFile(defaultPath)
	}

	// Try to find config file relative to executable
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		configPath := filepath.Join(exeDir, "schema", "lumi.json")
		if _, err := os.Stat(configPath); err == nil {
			return loader.LoadFromFile(configPath)
		}
	}

	// Fall back to environment variables and defaults only
	return loader.LoadFromEnvironment()
}

// MustLoad loads configuration and panics on error
func MustLoad(ctx context.Context) *Config {
	cfg, err := LoadDefault(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to load configuration: %v", err))
	}
	return cfg
}

// GetConfigPath returns the path to the configuration file used
func (cl *ConfigLoader) GetConfigPath() string {
	return cl.configPath
}
