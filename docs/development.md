# Development Guide

Comprehensive guide for developing with the Lumi-Go microservice template.

## Table of Contents
- [Prerequisites](#prerequisites)
- [Initial Setup](#initial-setup)
- [Development Workflow](#development-workflow)
- [Code Organization](#code-organization)
- [Testing](#testing)
- [Debugging](#debugging)
- [Performance](#performance)
- [Best Practices](#best-practices)

## Prerequisites

### Required Tools
- **Go**: Version 1.22 or higher
- **Git**: For version control
- **Make**: For running build commands

### Recommended Tools
- **Docker**: For containerized development
- **Air**: For hot reload during development
- **golangci-lint**: For code linting
- **Delve**: For debugging

### Installation

#### macOS
```bash
# Install Go
brew install go

# Install development tools
brew install make git docker

# Install Go tools
go install github.com/air-verse/air@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/go-delve/delve/cmd/dlv@latest
```

#### Linux
```bash
# Install Go (Ubuntu/Debian)
sudo apt update
sudo apt install golang-go make git

# Install Docker
curl -fsSL https://get.docker.com | sh

# Install Go tools
go install github.com/air-verse/air@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/go-delve/delve/cmd/dlv@latest
```

#### Windows
```powershell
# Install with Chocolatey
choco install golang make git docker-desktop

# Install Go tools
go install github.com/air-verse/air@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/go-delve/delve/cmd/dlv@latest
```

## Initial Setup

### 1. Clone Repository
```bash
git clone https://github.com/lumitut/lumi-go.git
cd lumi-go
```

### 2. Install Dependencies
```bash
# Download Go modules
go mod download

# Verify dependencies
go mod verify

# Tidy dependencies
go mod tidy
```

### 3. Setup Environment
```bash
# Copy example environment file
cp env.example .env

# Edit .env with your configuration
vim .env
```

### 4. Install Development Tools
```bash
make install-tools
```

## Development Workflow

### Starting Development

#### 1. Hot Reload Mode
Best for active development:
```bash
# Using Make
make run-dev

# Using Air directly
air

# With custom config
air -c .air.toml
```

#### 2. Standard Run
For testing without hot reload:
```bash
# Using Make
make run

# Using Go directly
go run cmd/server/main.go

# With specific config
go run cmd/server/main.go -config=config.json
```

#### 3. Docker Development
For containerized development:
```bash
# Start development container
docker-compose -f docker-compose.dev.yml up

# Rebuild on changes
docker-compose -f docker-compose.dev.yml up --build
```

### Code Formatting

Always format code before committing:
```bash
# Format all Go files
make fmt

# Or use gofmt directly
gofmt -w -s .

# Use goimports for import organization
goimports -w .
```

### Linting

Run linters to catch issues:
```bash
# Run golangci-lint
make lint

# Run go vet
make vet

# Run all checks
make ci
```

### Testing During Development

#### Run Tests Continuously
```bash
# Watch mode (requires entr)
find . -name '*.go' | entr -c go test ./...

# Or use a test watcher
go install github.com/onsi/ginkgo/v2/ginkgo@latest
ginkgo watch ./...
```

#### Run Specific Tests
```bash
# Test specific package
go test ./internal/config/...

# Run specific test
go test -run TestConfig ./internal/config/

# Verbose output
go test -v ./...

# With race detection
go test -race ./...
```

## Code Organization

### Directory Structure
```
lumi-go/
├── cmd/server/           # Application entrypoint
│   ├── main.go          # Main function
│   └── schema/          # Configuration files
├── internal/            # Private application code
│   ├── config/         # Configuration management
│   ├── httpapi/        # HTTP handlers and routes
│   ├── rpcapi/         # gRPC handlers
│   ├── service/        # Business logic
│   ├── domain/         # Domain models
│   ├── middleware/     # HTTP/gRPC middleware
│   ├── observability/  # Logging, metrics, tracing
│   └── clients/        # External service clients
├── api/                # API definitions
│   ├── openapi/       # OpenAPI specifications
│   └── proto/         # Protocol buffers
├── tests/             # Test suites
│   ├── unit/         # Unit tests
│   ├── integration/  # Integration tests
│   ├── e2e/         # End-to-end tests
│   └── fixtures/    # Test data
└── docs/             # Documentation
```

### Adding New Features

#### 1. Create Domain Model
```go
// internal/domain/user.go
package domain

type User struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}
```

#### 2. Implement Service Layer
```go
// internal/service/user_service.go
package service

type UserService struct {
    // Add dependencies
}

func NewUserService() *UserService {
    return &UserService{}
}

func (s *UserService) GetUser(ctx context.Context, id string) (*domain.User, error) {
    // Implementation
}
```

#### 3. Add HTTP Handler
```go
// internal/httpapi/user_handler.go
package httpapi

func (s *Server) getUserHandler(c *gin.Context) {
    userID := c.Param("id")
    
    user, err := s.userService.GetUser(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
        return
    }
    
    c.JSON(http.StatusOK, user)
}
```

#### 4. Register Route
```go
// internal/httpapi/routes.go
func registerAPIRoutes(router *gin.Engine, cfg *config.Config) {
    api := router.Group("/api/v1")
    {
        api.GET("/users/:id", s.getUserHandler)
    }
}
```

#### 5. Write Tests
```go
// tests/unit/service/user_service_test.go
func TestUserService_GetUser(t *testing.T) {
    service := NewUserService()
    
    user, err := service.GetUser(context.Background(), "123")
    
    assert.NoError(t, err)
    assert.Equal(t, "123", user.ID)
}
```

## Testing

### Test Organization
- **Unit Tests**: Test individual functions/methods
- **Integration Tests**: Test component interactions
- **E2E Tests**: Test complete workflows

### Writing Tests

#### Unit Test Example
```go
package config_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
    tests := []struct {
        name    string
        config  Config
        wantErr bool
    }{
        {
            name:    "valid config",
            config:  Config{Service: ServiceConfig{Name: "test"}},
            wantErr: false,
        },
        {
            name:    "invalid config",
            config:  Config{},
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Running Tests
```bash
# All tests
make test

# With coverage
make coverage

# Specific package
go test ./internal/config/...

# With race detection
go test -race ./...

# Benchmark tests
make bench
```

## Debugging

### Using Delve

#### 1. Debug the Application
```bash
# Start debugger
dlv debug cmd/server/main.go

# Set breakpoint
(dlv) break main.main

# Continue execution
(dlv) continue

# Print variable
(dlv) print cfg

# Step through code
(dlv) next
(dlv) step
```

#### 2. Debug Tests
```bash
# Debug specific test
dlv test ./internal/config -- -test.run TestConfig

# With breakpoint
(dlv) break config.Load
(dlv) continue
```

### Using VS Code

`.vscode/launch.json`:
```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Server",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/cmd/server",
            "env": {
                "LUMI_SERVICE_ENVIRONMENT": "development"
            }
        },
        {
            "name": "Debug Test",
            "type": "go",
            "request": "launch",
            "mode": "test",
            "program": "${workspaceFolder}/tests/unit/..."
        }
    ]
}
```

### Logging for Debugging
```go
import "github.com/lumitut/lumi-go/internal/observability/logger"

// Add debug logs
logger.Debug(ctx, "Processing request", 
    zap.String("user_id", userID),
    zap.Any("payload", payload))

// Conditional debug logging
if cfg.Service.Environment == "development" {
    logger.Debug(ctx, "Detailed debug info", zap.Any("data", data))
}
```

## Performance

### Profiling

#### CPU Profiling
```bash
# Enable pprof
export LUMI_SERVER_ENABLEPPROF=true

# Start service
make run

# Capture CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Analyze
(pprof) top
(pprof) list main.main
(pprof) web
```

#### Memory Profiling
```bash
# Capture memory profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Analyze allocations
(pprof) alloc_objects
(pprof) inuse_objects
```

### Benchmarking
```go
// Write benchmark
func BenchmarkUserService(b *testing.B) {
    service := NewUserService()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.GetUser(context.Background(), "123")
    }
}

// Run benchmark
go test -bench=. -benchmem -benchtime=10s ./...
```

### Load Testing
```bash
# Install hey
go install github.com/rakyll/hey@latest

# Run load test
hey -n 10000 -c 100 http://localhost:8080/api/users

# With custom headers
hey -n 1000 -c 10 -H "Authorization: Bearer token" http://localhost:8080/api/users
```

## Best Practices

### 1. Error Handling
```go
// Always wrap errors with context
if err != nil {
    return fmt.Errorf("failed to get user %s: %w", userID, err)
}

// Use custom error types
type NotFoundError struct {
    Resource string
    ID       string
}

func (e NotFoundError) Error() string {
    return fmt.Sprintf("%s with ID %s not found", e.Resource, e.ID)
}
```

### 2. Context Usage
```go
// Always accept context as first parameter
func GetUser(ctx context.Context, id string) (*User, error) {
    // Use context for cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        // Continue processing
    }
}
```

### 3. Dependency Injection
```go
// Use interfaces for dependencies
type UserRepository interface {
    GetUser(ctx context.Context, id string) (*User, error)
}

type UserService struct {
    repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
    return &UserService{repo: repo}
}
```

### 4. Configuration Management
```go
// Use structured configuration
type Config struct {
    Service ServiceConfig `json:"service"`
    Server  ServerConfig  `json:"server"`
}

// Validate configuration
func (c *Config) Validate() error {
    if c.Service.Name == "" {
        return errors.New("service name is required")
    }
    return nil
}
```

### 5. Logging
```go
// Use structured logging
logger.Info(ctx, "User created",
    zap.String("user_id", user.ID),
    zap.String("email", user.Email),
    zap.Time("created_at", user.CreatedAt))

// Add correlation IDs
ctx = context.WithValue(ctx, "correlation_id", uuid.New().String())
```

## Git Workflow

### Branch Strategy
```bash
# Create feature branch
git checkout -b feature/user-service

# Make changes and commit
git add .
git commit -m "feat: add user service"

# Push to remote
git push origin feature/user-service

# Create pull request
```

### Commit Messages
Follow conventional commits:
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation
- `style:` Code style
- `refactor:` Code refactoring
- `test:` Testing
- `chore:` Maintenance

### Pre-commit Hooks
```bash
# Install pre-commit
pip install pre-commit

# Install hooks
pre-commit install

# Run manually
pre-commit run --all-files
```

## Troubleshooting

### Common Issues

#### Module Download Failures
```bash
# Clear module cache
go clean -modcache

# Re-download
go mod download
```

#### Port Already in Use
```bash
# Find process using port
lsof -i :8080

# Kill process
kill -9 <PID>

# Or use different port
export LUMI_SERVER_HTTPPORT=8090
```

#### Build Failures
```bash
# Clean build cache
go clean -cache

# Rebuild
make clean build
```

## Resources

- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Go Proverbs](https://go-proverbs.github.io/)
