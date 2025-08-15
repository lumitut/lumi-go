# Development Environment Guide

This guide walks you through setting up and working with the lumi-go development environment.

## Prerequisites

Before starting, ensure you have completed the [engineering setup](./engineering.md).

## Initial Setup

### 1. Clone Repository

```bash
git clone https://github.com/lumitut/lumi-go.git
cd lumi-go
```

### 2. Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install development tools
make init
```

### 3. Configure Environment

```bash
# Copy environment template
cp .env.example .env

# Edit .env with your settings
# Key variables to configure:
# - DATABASE_URL
# - REDIS_URL
# - OTEL_EXPORTER_OTLP_ENDPOINT
```

## Local Development Stack

### Starting Services

```bash
# Start all services (recommended)
make up

# Or use the local script for more control
./scripts/local.sh start

# Start specific services
docker-compose up -d postgres redis
```

### Service URLs

| Service | URL | Credentials |
|---------|-----|-------------|
| API | http://localhost:8080 | - |
| gRPC | http://localhost:8081 | - |
| Metrics | http://localhost:9090/metrics | - |
| Prometheus | http://localhost:9091 | - |
| Grafana | http://localhost:3000 | admin/admin |
| Jaeger | http://localhost:16686 | - |
| PostgreSQL | localhost:5432 | lumigo/lumigo |
| Redis | localhost:6379 | - |

### Stopping Services

```bash
# Stop all services
make down

# Stop and clean volumes
./scripts/local.sh clean
```

## Running the Application

### Development Mode (Hot Reload)

```bash
# Run with hot reload
make run

# Or directly with air
air -c .air.toml
```

### Debug Mode

```bash
# Run with Delve debugger
dlv debug ./cmd/server -- serve

# Attach to running process
dlv attach $(pgrep lumi-go)
```

### Production Mode

```bash
# Build binary
make build

# Run binary
./bin/lumi-go serve
```

## Database Management

### Running Migrations

```bash
# Run all pending migrations
make migrate-up

# Rollback last migration
make migrate-down

# Reset database
make migrate-reset

# Check current version
make migrate-version
```

### Creating Migrations

```bash
# Create new migration
make migrate-create

# Or manually
migrate create -ext sql -dir migrations -seq your_migration_name
```

### Database Console

```bash
# Connect to PostgreSQL
./scripts/local.sh db

# Or directly
docker-compose exec postgres psql -U lumigo -d lumigo
```

### Seeding Data

```bash
# Run seed script
docker-compose exec -T postgres psql -U lumigo -d lumigo < scripts/seed.sql
```

## Testing

### Unit Tests

```bash
# Run all tests
make test

# Run with coverage
make coverage

# Run specific package
go test ./internal/service/...

# Run with race detection
go test -race ./...
```

### Integration Tests

```bash
# Run integration tests
make test-integration

# Run specific integration test
go test -tags=integration ./internal/repo -run TestUserRepo
```

### Benchmarks

```bash
# Run all benchmarks
make benchmark

# Run specific benchmark
go test -bench=BenchmarkUserService ./internal/service
```

## Code Quality

### Formatting

```bash
# Format all code
make fmt

# Check formatting
gofmt -l .
```

### Linting

```bash
# Run all linters
make lint

# Run specific linter
golangci-lint run --enable=gocyclo ./...
```

### Security Scanning

```bash
# Run security scans
make security-scan

# Individual tools
govulncheck ./...
gosec ./...
gitleaks detect
```

## Code Generation

### Protocol Buffers

```bash
# Generate from proto files
buf generate

# Lint proto files
buf lint
```

### SQL Code Generation

```bash
# Generate from SQL queries
sqlc generate

# Verify generation
sqlc compile
```

### Mocks

```bash
# Generate all mocks
go generate ./...

# Generate specific mock
mockery --name=UserService --dir=internal/service
```

## Debugging

### Application Logs

```bash
# View application logs
docker-compose logs -f app

# View all logs
make logs

# Filter logs
docker-compose logs app | grep ERROR
```

### Debugging with Delve

```bash
# Start with debugger
dlv debug ./cmd/server -- serve

# Set breakpoint
(dlv) break internal/service/user.go:45

# Continue execution
(dlv) continue

# Print variable
(dlv) print user

# Stack trace
(dlv) stack
```

### Performance Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof

# Live profiling
go tool pprof http://localhost:9090/debug/pprof/profile
```

## Docker Development

### Building Images

```bash
# Build development image
docker build -f deploy/docker/Dockerfile.dev -t lumi-go:dev .

# Build production image
make docker-build
```

### Running Containers

```bash
# Run development container
docker run -v $(pwd):/app -p 8080:8080 lumi-go:dev

# Run production container
make docker-run
```

## Git Workflow

### Branch Strategy

```bash
# Create feature branch
git checkout -b feature/your-feature

# Create bugfix branch
git checkout -b bugfix/your-fix

# Create hotfix branch
git checkout -b hotfix/critical-fix
```

### Commit Convention

```bash
# Format: <type>(<scope>): <subject>

# Examples:
git commit -m "feat(auth): add JWT refresh token support"
git commit -m "fix(db): resolve connection pool leak"
git commit -m "docs(api): update OpenAPI specification"
git commit -m "chore(deps): update Go modules"
```

### Pre-commit Hooks

```bash
# Install pre-commit hooks
make pre-commit

# Run manually
pre-commit run --all-files

# Skip hooks (emergency only)
git commit --no-verify
```

## Environment Variables

### Development Defaults

```bash
ENV=dev
LOG_LEVEL=debug
GIN_MODE=debug
HTTP_ADDR=:8080
RPC_ADDR=:8081
PROM_ADDR=:9090
PG_URL=postgres://lumigo:lumigo@localhost:5432/lumigo?sslmode=disable
REDIS_URL=redis://localhost:6379/0
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
```

### Testing Environment

```bash
ENV=test
LOG_LEVEL=error
GIN_MODE=test
PG_URL=postgres://test:test@localhost:5432/test?sslmode=disable
```

## IDE Configuration

### VS Code

Launch configuration (`.vscode/launch.json`):
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Server",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/server",
      "args": ["serve"],
      "env": {
        "ENV": "dev",
        "LOG_LEVEL": "debug"
      }
    }
  ]
}
```

### GoLand

Run configuration:
1. Go to Run → Edit Configurations
2. Add Go Build configuration
3. Set package path: `./cmd/server`
4. Set program arguments: `serve`
5. Set environment variables from `.env`

## Troubleshooting

### Port Already in Use

```bash
# Find process using port
lsof -i :8080

# Kill process
kill -9 <PID>
```

### Docker Issues

```bash
# Clean Docker system
docker system prune -a

# Reset Docker
docker-compose down -v
docker-compose up --force-recreate
```

### Database Connection

```bash
# Test connection
psql "postgres://lumigo:lumigo@localhost:5432/lumigo?sslmode=disable" -c "SELECT 1"

# Check Docker network
docker network ls
docker network inspect lumi-go_lumi-network
```

### Go Module Issues

```bash
# Clear module cache
go clean -modcache

# Update dependencies
go get -u ./...
go mod tidy
```

## Monitoring

### Metrics

View metrics at http://localhost:9090/metrics

Key metrics:
- `http_requests_total` - Request count
- `http_request_duration_seconds` - Request latency
- `go_goroutines` - Goroutine count
- `go_memstats_alloc_bytes` - Memory usage

### Tracing

View traces at http://localhost:16686

Finding traces:
1. Select service: `lumi-go`
2. Select operation or search by trace ID
3. View timeline and spans

### Dashboards

Access Grafana at http://localhost:3000

Default dashboards:
- Service Overview
- HTTP Metrics
- Database Performance
- Redis Cache Stats

## Best Practices

### Development Tips

1. **Always run tests before committing**
   ```bash
   make ci
   ```

2. **Keep dependencies updated**
   ```bash
   go get -u ./...
   go mod tidy
   ```

3. **Use conventional commits**
   - feat: New feature
   - fix: Bug fix
   - docs: Documentation
   - refactor: Code refactoring
   - test: Test updates
   - chore: Maintenance

4. **Profile before optimizing**
   ```bash
   go test -bench=. -cpuprofile=cpu.prof
   go tool pprof -http=:8080 cpu.prof
   ```

5. **Use structured logging**
   ```go
   logger.Info("user created",
       zap.String("user_id", userID),
       zap.String("email", email))
   ```

## Resources

- [Project README](../README.md)
- [API Documentation](./api.md)
- [Architecture Guide](./architecture.md)
- [Contributing Guide](../CONTRIBUTING.md)
- [Security Policy](../SECURITY.md)
