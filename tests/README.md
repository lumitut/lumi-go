# Testing Guide for Lumi-Go

This directory contains all test suites for the lumi-go microservice. The tests are organized to focus on the core microservice functionality without requiring external dependencies.

## Test Organization

```
tests/
├── unit/          # Unit tests for individual components
├── integration/   # Integration tests for API endpoints
├── e2e/           # End-to-end tests for complete workflows
├── performance/   # Performance and load tests
├── smoke/         # Smoke tests
├── fixtures/      # Test data and mocks
├── helpers/       # Shared test utilities
└── setup/         # Test config
```

## Testing Philosophy

1. **No External Dependencies for Unit Tests**: Unit tests should run without databases, Redis, or external services
2. **Mocked External Services**: Use interfaces and mocks for external dependencies
3. **Fast Feedback**: Tests should run quickly to enable rapid development
4. **Clear Test Names**: Test names should describe what they test and expected behavior

## Running Tests

### All Tests
```bash
make test
```

### Unit Tests Only
```bash
make test-unit
# or
go test -v -race ./tests/unit/...
```

### Integration Tests
```bash
make test-integration
# or
go test -v -race -tags=integration ./tests/integration/...
```

### End-to-End Tests
```bash
make test-e2e
# or
go test -v -race -tags=e2e ./tests/e2e/...
```

### Coverage Report
```bash
make coverage
# View HTML report
open coverage/coverage.html
```

## Writing Tests

### Unit Test Example

```go
// tests/unit/config/config_test.go
package config_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/lumitut/lumi-go/internal/config"
)

func TestConfig_Validate(t *testing.T) {
    tests := []struct {
        name    string
        config  *config.Config
        wantErr bool
    }{
        {
            name: "valid config",
            config: &config.Config{
                Service: config.ServiceConfig{
                    Name:        "test-service",
                    Environment: "development",
                },
            },
            wantErr: false,
        },
        {
            name: "missing service name",
            config: &config.Config{
                Service: config.ServiceConfig{
                    Environment: "development",
                },
            },
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

### Integration Test Example

```go
// tests/integration/api_test.go
// +build integration

package integration_test

import (
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/lumitut/lumi-go/internal/httpapi"
)

func TestHealthEndpoint(t *testing.T) {
    // Setup
    server := httpapi.NewServer(testConfig())
    router := server.SetupRoutes()
    
    // Test
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/health", nil)
    router.ServeHTTP(w, req)
    
    // Assert
    assert.Equal(t, http.StatusOK, w.Code)
    assert.Contains(t, w.Body.String(), "healthy")
}
```

### Mocking External Services

```go
// tests/fixtures/mock_database.go
package fixtures

type MockDatabase struct {
    GetUserFunc func(id string) (*User, error)
}

func (m *MockDatabase) GetUser(id string) (*User, error) {
    if m.GetUserFunc != nil {
        return m.GetUserFunc(id)
    }
    return nil, nil
}

// Usage in tests
func TestServiceWithDatabase(t *testing.T) {
    mockDB := &MockDatabase{
        GetUserFunc: func(id string) (*User, error) {
            return &User{ID: id, Name: "Test User"}, nil
        },
    }
    
    service := NewService(mockDB)
    user, err := service.GetUser("123")
    
    assert.NoError(t, err)
    assert.Equal(t, "Test User", user.Name)
}
```

## Test Patterns

### Table-Driven Tests
Use table-driven tests for testing multiple scenarios:

```go
func TestCalculate(t *testing.T) {
    tests := []struct {
        name     string
        input    int
        expected int
    }{
        {"positive", 5, 10},
        {"negative", -5, 0},
        {"zero", 0, 0},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Calculate(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Parallel Tests
Run independent tests in parallel for faster execution:

```go
func TestParallel(t *testing.T) {
    t.Parallel()
    // test code
}
```

### Test Helpers
Create helper functions for common test setup:

```go
// tests/helpers/setup.go
func NewTestServer(t *testing.T) *httptest.Server {
    t.Helper()
    // setup code
    return server
}
```

## Benchmarking

Write benchmarks to measure performance:

```go
// tests/unit/service/service_bench_test.go
func BenchmarkProcessRequest(b *testing.B) {
    service := NewService()
    request := createTestRequest()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.ProcessRequest(request)
    }
}

// Run benchmarks
// make bench
```

## Testing with External Services

For tests that require external services, use build tags:

```go
// +build integration,postgres

package integration_test

func TestWithPostgres(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping test that requires PostgreSQL")
    }
    
    // Test with real PostgreSQL connection
    db := setupTestDatabase(t)
    defer cleanupDatabase(t, db)
    
    // test code
}
```

## Continuous Integration

The CI pipeline runs tests in this order:
1. Unit tests (no external dependencies)
2. Integration tests (may use test containers)
3. E2E tests (full system tests)

```yaml
# .github/workflows/test.yml
- name: Run tests
  run: |
    make test-unit
    make test-integration
    make coverage-check
```

## Best Practices

1. **Keep Tests Fast**: Aim for < 1 second per unit test
2. **Test Behavior, Not Implementation**: Focus on what the code does, not how
3. **Use Descriptive Names**: Test names should be self-documenting
4. **Isolate Tests**: Tests should not depend on each other
5. **Clean Up**: Always clean up resources in tests
6. **Use t.Helper()**: Mark helper functions with t.Helper()
7. **Check Error Messages**: Test both error occurrence and message content
8. **Use Subtests**: Group related tests using t.Run()

## Testing Commands Reference

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with race detection
go test -race ./...

# Run specific test
go test -run TestName ./...

# Generate coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. ./...

# Run with specific tags
go test -tags=integration ./...

# Skip long tests
go test -short ./...
```

## Troubleshooting

### Tests Hanging
- Check for deadlocks in concurrent code
- Add timeouts to tests: `ctx, cancel := context.WithTimeout(...)`

### Flaky Tests
- Remove time-dependent assertions
- Use deterministic test data
- Mock time.Now() when needed

### Slow Tests
- Run tests in parallel with `t.Parallel()`
- Use smaller datasets for unit tests
- Move slow tests to integration suite

### Coverage Issues
- Exclude generated code from coverage
- Focus on testing business logic
- Aim for 80%+ coverage of critical paths
