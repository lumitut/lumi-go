# Quick Start Guide

Get the Lumi-Go microservice up and running in 5 minutes!

## Prerequisites

- Go 1.22+ installed
- Docker (optional, for containerized deployment)
- Make (optional, for simplified commands)

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/lumitut/lumi-go.git
cd lumi-go
```

### 2. Setup Development Environment

#### Option A: Automated Setup (Recommended)
```bash
# One command to setup everything
./scripts/local.sh setup
```

#### Option B: Manual Setup
```bash
go mod download
# or
make deps
```

### 3. Run the Service

#### Option A: Using Development Script (Recommended)
```bash
./scripts/local.sh start
```

#### Option B: Using Make
```bash
make run
```

#### Option C: With Hot Reload (Development)
```bash
make run-dev
# or
air
```

#### Option D: Using Docker
```bash
# Build and run with Docker
make docker-build
make docker-run

# Or using docker-compose
docker-compose up
```

## Verify Installation

### Health Check
```bash
curl http://localhost:8080/health
# Expected: {"status":"healthy","timestamp":...}
```

### Readiness Check
```bash
curl http://localhost:8080/ready
# Expected: {"status":"ready","time":...}
```

### Metrics
```bash
curl http://localhost:9090/metrics
# Expected: Prometheus metrics output
```

## Basic Configuration

### Using Environment Variables
```bash
# Set service configuration
export LUMI_SERVICE_NAME=my-service
export LUMI_SERVICE_ENVIRONMENT=development
export LUMI_SERVER_HTTPPORT=8080

# Run the service
make run
```

### Using Configuration File
Edit `cmd/server/schema/lumi.json`:
```json
{
  "service": {
    "name": "my-service",
    "environment": "development"
  },
  "server": {
    "httpPort": "8080"
  }
}
```

## Adding External Services (Optional)

### PostgreSQL Database
```bash
# Enable database client
export LUMI_CLIENTS_DATABASE_ENABLED=true
export LUMI_CLIENTS_DATABASE_URL=postgres://user:pass@localhost:5432/mydb

# Run service
make run
```

### Redis Cache
```bash
# Enable Redis client
export LUMI_CLIENTS_REDIS_ENABLED=true
export LUMI_CLIENTS_REDIS_URL=redis://localhost:6379/0

# Run service
make run
```

## Development Workflow

### 1. Start Development Environment
```bash
# Start with hot reload
make run-dev
```

### 2. Run Tests
```bash
# All tests
make test

# Unit tests only
make test-unit

# With coverage
make coverage
```

### 3. Format and Lint
```bash
# Format code
make fmt

# Run linter
make lint

# Run go vet
make vet
```

### 4. Build Binary
```bash
# Build for current platform
make build

# Build for specific platform
GOOS=linux GOARCH=amd64 make build
```

## Project Structure

```
lumi-go/
├── cmd/server/          # Application entrypoint
│   ├── main.go         # Main function
│   └── schema/         # Configuration schema
│       └── lumi.json   # Default configuration
├── internal/           # Private application code
│   ├── config/        # Configuration management
│   ├── httpapi/       # HTTP handlers
│   ├── middleware/    # HTTP middleware
│   └── observability/ # Logging, metrics, tracing
├── api/               # API definitions
├── tests/             # Test suites
└── deploy/            # Deployment configurations
```

## Common Commands

| Command | Description |
|---------|-------------|
| `make run` | Run the service |
| `make run-dev` | Run with hot reload |
| `make test` | Run all tests |
| `make build` | Build binary |
| `make docker-build` | Build Docker image |
| `make docker-run` | Run Docker container |
| `make clean` | Clean build artifacts |
| `make help` | Show all available commands |

## API Endpoints

### System Endpoints
- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /metrics` - Prometheus metrics

### Application Endpoints
Add your custom endpoints in `internal/httpapi/routes.go`:
```go
func registerAPIRoutes(router *gin.Engine, cfg *config.Config) {
    api := router.Group("/api/v1")
    {
        api.GET("/users", getUsersHandler)
        api.POST("/users", createUserHandler)
        // Add more endpoints here
    }
}
```

## Environment Variables

Key environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `LUMI_SERVICE_NAME` | Service name | lumi-go |
| `LUMI_SERVICE_ENVIRONMENT` | Environment (development/staging/production) | development |
| `LUMI_SERVER_HTTPPORT` | HTTP server port | 8080 |
| `LUMI_SERVER_RPCPORT` | gRPC server port | 8081 |
| `LUMI_OBSERVABILITY_LOGLEVEL` | Log level | info |
| `LUMI_OBSERVABILITY_METRICSENABLED` | Enable metrics | true |

## Troubleshooting

### Port Already in Use
```bash
# Change the port
export LUMI_SERVER_HTTPPORT=8090
make run
```

### Configuration Not Loading
```bash
# Check config file exists
ls cmd/server/schema/lumi.json

# Verify environment variables
env | grep LUMI_
```

### Build Failures
```bash
# Clean and rebuild
make clean
go mod tidy
make build
```

## Next Steps

1. Read the [Development Guide](development.md) for detailed setup
2. Check [External Services](external-services.md) for database/cache integration
3. Review [Observability](observability.md) for monitoring setup
4. See [Engineering Guide](engineering.md) for best practices

## Getting Help

- Check the [FAQ](faq.md)
- Browse [GitHub Issues](https://github.com/lumitut/lumi-go/issues)
- Ask in [Discussions](https://github.com/lumitut/lumi-go/discussions)
