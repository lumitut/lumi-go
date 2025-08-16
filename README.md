# Lumi-Go: Lean Go Microservice Template

A production-ready, lean Go microservice template focused on simplicity, performance, and maintainability. This template provides a solid foundation for building microservices without unnecessary bloat.

## Features

### Core Service
- **HTTP & gRPC Support**: Dual protocol support for flexible API design
- **Structured Configuration**: JSON + environment variables with Viper
- **Graceful Shutdown**: Proper cleanup and connection draining
- **Health Checks**: Built-in `/health` and `/ready` endpoints
- **Metrics**: Prometheus-compatible metrics endpoint

### Observability
- **Structured Logging**: JSON logging with zap
- **Metrics Collection**: Built-in Prometheus metrics
- **Distributed Tracing**: Optional OpenTelemetry support
- **Request Tracking**: Correlation IDs for request tracing

### Development Experience
- **Hot Reload**: Development mode with Air
- **Docker Support**: Multi-stage Dockerfile for minimal images
- **Testing Framework**: Unit, integration, and E2E test structure
- **Code Generation**: Mock generation support
- **CI/CD Ready**: GitHub Actions workflows included

### Security & Reliability
- **Rate Limiting**: Configurable per-IP rate limiting
- **CORS Support**: Configurable cross-origin resource sharing
- **Panic Recovery**: Graceful error handling
- **Input Validation**: Request validation middleware

## Quick Start

### Prerequisites
- Go 1.22+
- Docker & Docker Compose (optional)
- Make (optional but recommended)

### Installation

1. Clone the template:
```bash
git clone https://github.com/lumitut/lumi-go.git my-service
cd my-service
```

2. Install dependencies:
```bash
make deps
```

3. Run locally:
```bash
make run
```

Or with hot reload:
```bash
make run-dev
```

### Docker Development

Build and run with Docker:
```bash
make docker-dev
```

This starts the service with hot reload enabled.

## Configuration

The service uses a layered configuration approach:

1. **Defaults**: Built-in sensible defaults
2. **JSON Config**: `cmd/server/schema/lumi.json`
3. **Environment Variables**: `LUMI_*` prefixed variables

### Configuration Structure

```json
{
  "service": {
    "name": "lumi-go",
    "version": "1.0.0",
    "environment": "development"
  },
  "server": {
    "httpPort": "8080",
    "rpcPort": "8081"
  },
  "clients": {
    "database": {
      "enabled": false,
      "url": ""
    },
    "redis": {
      "enabled": false,
      "url": ""
    }
  }
}
```

### Environment Variables

Override any configuration via environment variables:

```bash
export LUMI_SERVICE_NAME=my-service
export LUMI_SERVER_HTTPPORT=8080
export LUMI_CLIENTS_DATABASE_ENABLED=true
export LUMI_CLIENTS_DATABASE_URL=postgres://localhost:5432/mydb
```

## External Services

This template is designed to be lean. External services (databases, caches, message queues) are configured as optional clients. 

### Connecting to External Services

For local development with external services, use the dedicated templates:

- **PostgreSQL**: `templates/lumi-postgres`
- **Redis**: `templates/lumi-redis`
- **Grafana**: `templates/lumi-grafana`

Example:
```bash
# Start PostgreSQL
cd ../lumi-postgres && docker-compose up -d

# Configure connection in lumi-go
export LUMI_CLIENTS_DATABASE_ENABLED=true
export LUMI_CLIENTS_DATABASE_URL=postgres://user:pass@localhost:5432/db

# Run the service
make run
```

For production, use managed cloud services (AWS RDS, Azure Database, Google Cloud SQL, etc.).

## Project Structure

```
lumi-go/
├── cmd/
│   └── server/           # Application entrypoint
│       ├── main.go
│       └── schema/       # Configuration schema
│           └── lumi.json
├── internal/            # Private application code
│   ├── config/          # Configuration management
│   ├── httpapi/         # HTTP handlers
│   ├── rpcapi/          # gRPC/Connect handlers
│   ├── middleware/      # HTTP/gRPC middleware
│   ├── service/         # Business logic
│   ├── domain/          # Domain models
│   └── observability/   # Logging, metrics, tracing
├── api/                 # API definitions
│   ├── openapi/         # OpenAPI specs
│   └── proto/           # Protocol buffer definitions
├── tests/               # Test suites
│   ├── unit/
│   ├── integration/
│   └── e2e/
├── deploy/              # Deployment configurations
│   └── docker/
├── docs/                # Documentation
└── scripts/             # Utility scripts
```

## Development

### Running Tests

```bash
# All tests
make test

# Unit tests only
make test-unit

# Integration tests
make test-integration

# Coverage report
make coverage
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Run vet
make vet
```

### Building

```bash
# Build binary
make build

# Build Docker image
make docker-build
```

## API Endpoints

### Health Checks

- `GET /health` - Liveness probe
- `GET /ready` - Readiness probe
- `GET /metrics` - Prometheus metrics

### Application APIs

Define your APIs in:
- `api/openapi/` for REST APIs
- `api/proto/` for gRPC APIs

## Deployment

### Docker

The service builds into a minimal distroless image (~15MB):

```bash
docker build -t my-service:latest -f deploy/docker/Dockerfile .
docker run -p 8080:8080 my-service:latest
```

### Kubernetes

Example deployment manifests are provided in `deploy/k8s/`.

### Cloud Platforms

The service is designed to run on any container platform:
- AWS ECS/Fargate
- Google Cloud Run
- Azure Container Instances
- Kubernetes (EKS, GKE, AKS)

## Performance

The template is optimized for:
- **Fast startup**: < 100ms cold start
- **Low memory**: ~10MB baseline memory usage
- **High throughput**: 10k+ requests/second
- **Small image**: ~15MB Docker image

## Security

- Non-root container execution
- Distroless base image
- Secret management via environment variables
- Rate limiting and CORS protection
- Structured audit logging

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Open a Pull Request

## License

MIT License - see LICENSE file for details

## Support

For issues and questions:
- GitHub Issues: [github.com/lumitut/lumi-go/issues](https://github.com/lumitut/lumi-go/issues)
- Documentation: [docs/](docs/)

## Acknowledgments

Built with:
- [Gin](https://github.com/gin-gonic/gin) - HTTP framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [Zap](https://github.com/uber-go/zap) - Structured logging
- [OpenTelemetry](https://opentelemetry.io/) - Observability
