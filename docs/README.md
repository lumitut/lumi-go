# Lumi-Go Documentation

Welcome to the comprehensive documentation for the Lumi-Go microservice template.

## Documentation Structure

### Getting Started
- [**Quickstart Guide**](quickstart.md) - Get up and running in 5 minutes
- [**Development Guide**](development.md) - Local development setup and workflows
- [**Docker Guide**](docker.md) - Container-based development and deployment

### Architecture
- [**Architecture Decision Records**](adr/) - Key architectural decisions
- [**Engineering Principles**](engineering.md) - Development best practices
- [**External Services**](external-services.md) - Integration with databases, caches, etc.

### Operations
- [**Observability**](observability.md) - Logging, metrics, and tracing
- [**Logging Guide**](logging.md) - Structured logging practices
- [**Metrics Guide**](metrics.md) - Prometheus metrics and monitoring
- [**Deployment**](helm.md) - Kubernetes and Helm deployment

### Development Tools
- [**Tools Guide**](tools.md) - Required and recommended development tools
- [**Migration Guide**](migrations.md) - Database migration management

### Reference
- [**Simplification Summary**](simplification-summary.md) - Recent architecture simplification

## Quick Links

| Resource | Description |
|----------|-------------|
| [API Documentation](../api/) | OpenAPI and Protocol Buffer definitions |
| [Configuration](../cmd/server/schema/lumi.json) | Service configuration schema |
| [Examples](../examples/) | Example implementations |
| [Tests](../tests/) | Testing guidelines and examples |

## Architecture Overview

```
┌─────────────────────────────────────────────┐
│              Load Balancer                  │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│           Lumi-Go Service                   │
│  ┌────────────────────────────────────┐     │
│  │         HTTP/gRPC Server           │     │
│  └────────────────┬───────────────────┘     │
│                   │                         │
│  ┌────────────────▼───────────────────┐     │
│  │          Middleware Layer          │     │
│  │  • Rate Limiting                   │     │
│  │  • CORS                            │     │
│  │  • Authentication                  │     │
│  │  • Request Logging                 │     │
│  └────────────────┬───────────────────┘     │
│                   │                         │
│  ┌────────────────▼───────────────────┐     │
│  │         Business Logic             │     │
│  └────────────────┬───────────────────┘     │
│                   │                         │
│  ┌────────────────▼───────────────────┐     │
│  │     Optional External Clients      │     │
│  │  • Database (PostgreSQL)           │     │
│  │  • Cache (Redis)                   │     │
│  │  • Message Queue                   │     │
│  └────────────────────────────────────┘     │
└─────────────────────────────────────────────┘
```

## Key Features

- **Lean Architecture**: Minimal dependencies, maximum performance
- **Cloud-Native**: Designed for Kubernetes and containerized environments
- **Observable**: Built-in metrics, structured logging, distributed tracing
- **Configurable**: JSON + environment variable configuration
- **Testable**: Comprehensive testing support with mocks and fixtures
- **Secure**: Security scanning, rate limiting, CORS support

## Configuration Priority

1. **Environment Variables** (highest priority)
2. **Configuration File** (`lumi.json`)
3. **Built-in Defaults** (lowest priority)

Example:
```bash
# Override configuration with environment variables
export LUMI_SERVICE_NAME=my-service
export LUMI_SERVER_HTTPPORT=8080
export LUMI_CLIENTS_DATABASE_URL=postgres://localhost:5432/db
```

## External Services

External services are **optional** and configured as clients:

```json
{
  "clients": {
    "database": {
      "enabled": true,
      "url": "postgres://localhost:5432/mydb"
    },
    "redis": {
      "enabled": false,
      "url": ""
    }
  }
}
```

## Observability

### Metrics
- Exposed at `/metrics` endpoint
- Prometheus-compatible format
- HTTP request metrics
- Custom business metrics

### Logging
- Structured JSON logging
- Correlation IDs for request tracing
- Log levels: debug, info, warn, error
- PII redaction support

### Tracing
- OpenTelemetry support
- Optional integration with Jaeger/Zipkin
- Automatic span creation for HTTP/gRPC

## Testing

```bash
# Run all tests
make test

# Unit tests only
make test-unit

# Integration tests
make test-integration

# Generate coverage report
make coverage
```

## Deployment

### Docker
```bash
# Build image
make docker-build

# Run container
docker run -p 8080:8080 lumitut/lumi-go:latest
```

### Kubernetes
```bash
# Using Helm
helm install lumi-go ./deploy/helm

# Using kubectl
kubectl apply -f ./deploy/k8s/
```

## Contributing

1. Read the [Engineering Guide](engineering.md)
2. Follow the [Development Setup](development.md)
3. Check [Architecture Decision Records](adr/)
4. Submit PR following the template

## Support

- **Issues**: [GitHub Issues](https://github.com/lumitut/lumi-go/issues)
- **Discussions**: [GitHub Discussions](https://github.com/lumitut/lumi-go/discussions)
- **Security**: See [SECURITY.md](../SECURITY.md)

## License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.
