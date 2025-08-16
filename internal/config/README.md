# Configuration Management

This package provides a robust, centralized configuration management system for the lumi-go service using [Viper](https://github.com/spf13/viper).

## Features

- **Multiple configuration sources** with priority ordering
- **JSON configuration file** with schema validation
- **Environment variable overrides** for deployment flexibility
- **Command-line flags** for runtime configuration
- **Type-safe configuration structures** with validation
- **Sensible defaults** for all configuration values

## Configuration Priority

The configuration system loads values in the following priority order (highest to lowest):

1. **Environment Variables** - Override any other configuration
2. **Configuration File** - JSON file with structured configuration
3. **Defaults** - Built-in sensible defaults

## Usage

### Basic Usage

```go
import "github.com/lumitut/lumi-go/internal/config"

// Load configuration with all available sources
cfg, err := config.LoadDefault(ctx)
if err != nil {
    log.Fatal("Failed to load configuration:", err)
}

// Use configuration
fmt.Println("Service:", cfg.Service.Name)
fmt.Println("Port:", cfg.Server.HTTPPort)
```

### Command-Line Flags

The service supports the following command-line flags:

```bash
# Use a custom configuration file
./lumi-go -config=/path/to/config.json

# Override environment
./lumi-go -env=production
```

### Configuration File

The default configuration file is located at `cmd/server/schema/lumi.json`. Here's the structure:

```json
{
  "service": {
    "name": "lumi-go",
    "version": "1.0.0",
    "environment": "development",
    "logLevel": "info"
  },
  "server": {
    "httpPort": "8080",
    "rpcPort": "8081",
    "httpReadTimeout": "15s",
    "httpWriteTimeout": "15s"
  },
  "database": {
    "host": "localhost",
    "port": "5432",
    "user": "postgres",
    "database": "lumi"
  }
  // ... more configuration sections
}
```

### Environment Variables

All configuration values can be overridden using environment variables. The format is:

```
LUMI_<SECTION>_<KEY>
```

Examples:

```bash
# Service configuration
export LUMI_SERVICE_NAME=my-service
export LUMI_SERVICE_ENVIRONMENT=production
export LUMI_SERVICE_LOGLEVEL=debug

# Database configuration
export LUMI_DATABASE_HOST=db.example.com
export LUMI_DATABASE_PORT=5432
export LUMI_DATABASE_PASSWORD=secret

# Server configuration
export LUMI_SERVER_HTTPPORT=8080
export LUMI_SERVER_RPCPORT=8081

# Redis configuration
export LUMI_REDIS_HOST=redis.example.com
export LUMI_REDIS_PASSWORD=redis-secret

# Feature flags
export LUMI_FEATURES_MAINTENANCEMODE=true
export LUMI_FEATURES_ENABLENEWAPI=false
```

## Configuration Sections

### Service Configuration
Basic service metadata and environment settings.

### Server Configuration
HTTP and RPC server settings including ports, timeouts, and profiling options.

### Database Configuration
PostgreSQL connection settings and connection pool configuration.

### Redis Configuration
Redis connection settings and connection pool configuration.

### Observability Configuration
Logging, metrics, and distributed tracing settings.

### Middleware Configuration
Settings for CORS, rate limiting, recovery, and request logging.

### Features Configuration
Feature flags for controlling application behavior.

## Advanced Usage

### Using the ConfigLoader

```go
// Create a custom loader
loader := config.NewConfigLoader()

// Load from a specific file
cfg, err := loader.LoadFromFile("/custom/path/config.json")

// Load from environment only (no config file)
cfg, err := loader.LoadFromEnvironment()

// Load with command-line flags
cfg, err := loader.LoadFromFlags()
```

### Direct Parser Access

```go
// Create a parser with defaults
parser := config.NewParser()

// Load configuration
cfg, err := parser.LoadConfig("path/to/config.json")

// Access underlying viper instance for advanced usage
v := parser.GetViper()
v.Set("custom.key", "value")
```

### MustLoad Pattern

For applications where configuration is critical:

```go
// Panics if configuration cannot be loaded
cfg := config.MustLoad(ctx)
```

## Validation

Configuration is automatically validated when loaded. The validation includes:

- Port ranges (1-65535)
- Environment values (development, staging, production)
- Log levels (debug, info, warn, error, fatal)
- Rate limit types (ip, user, api_key)

## Best Practices

1. **Use environment variables for secrets** - Never commit passwords or API keys to the config file
2. **Keep defaults sensible** - Defaults should work for local development
3. **Document all configuration** - Add comments in the JSON schema
4. **Validate early** - Configuration validation happens at load time
5. **Use appropriate timeouts** - Set realistic timeouts for your use case

## Testing

When testing, you can create a test configuration:

```go
func TestMyFunction(t *testing.T) {
    // Create test config
    cfg := &config.Config{
        Service: config.ServiceConfig{
            Name: "test-service",
            Environment: "test",
        },
        // ... other test values
    }
    
    // Use in tests
    server := NewServer(cfg)
    // ...
}
```

## Migration from Environment-Only Configuration

If migrating from the previous environment-only configuration:

1. Existing environment variables still work (with LUMI_ prefix instead of direct names)
2. Create a `lumi.json` file with your standard configuration
3. Use environment variables only for deployment-specific overrides

Example migration:
- Old: `SERVICE_NAME=myapp`
- New: `LUMI_SERVICE_NAME=myapp`

## Troubleshooting

### Configuration Not Loading

1. Check file path - ensure `cmd/server/schema/lumi.json` exists
2. Verify JSON syntax - use a JSON validator
3. Check environment variable format - must use LUMI_ prefix
4. Review logs - configuration loading errors are logged

### Environment Variables Not Working

1. Ensure correct prefix: `LUMI_`
2. Use underscores for nested values: `LUMI_DATABASE_HOST`
3. Check case sensitivity - use uppercase for env vars

### Validation Errors

1. Check port ranges (1-65535)
2. Verify environment is one of: development, staging, production
3. Ensure required fields are provided
