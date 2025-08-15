# Tests Directory

This directory contains all tests for the lumi-go project, organized by component.

## 📂 Test Organization

Tests are organized to mirror the internal package structure:

```
tests/
├── unit/  
│   └── observability/     # Observability component tests
│       └── logger/        # Logger tests
├── middleware/            # Middleware tests (future)
├── services/              # Service layer tests (future)
├── repositories/          # Repository tests (future)
├── integration/           # Integration tests (future)
├── e2e/                   # End-to-end tests (future)
├── performance/           # Performance tests (future)
├── ...                    # 
├── fixtures/              # Fixtures
└──          # 

```

## 🧪 Running Tests

### All Tests
```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run with race detection
make test-race
```

### Specific Components
```bash
# Test logger
go test ./tests/observability/logger/...

# Test with verbose output
go test -v ./tests/observability/logger/...

# Run specific test
go test -run TestLoggerInitialization ./tests/observability/logger/
```

### Benchmarks
```bash
# Run all benchmarks
go test -bench=. ./tests/...

# Run specific benchmark
go test -bench=BenchmarkLogging ./tests/observability/logger/

# Run benchmarks with memory profiling
go test -bench=. -benchmem ./tests/observability/logger/
```

## 📝 Writing Tests

### Test File Naming
- Unit tests: `*_test.go`
- Integration tests: `*_integration_test.go`
- Benchmarks: Include `Benchmark*` functions in test files

### Test Structure
```go
func TestComponentFunction(t *testing.T) {
    // Arrange
    // ... setup test data
    
    // Act
    // ... execute function
    
    // Assert
    // ... verify results
}
```

### Table-Driven Tests
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:  "valid input",
            input: "test",
            want:  "TEST",
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

## 🎯 Test Coverage Goals

| Component | Target Coverage | Current |
|-----------|----------------|---------|
| Core Business Logic | 90% | - |
| HTTP Handlers | 80% | - |
| Middleware | 85% | - |
| Utilities | 95% | - |
| Database Repositories | 75% | - |

## 🔧 Test Utilities

### Mocking
```go
// Use interfaces for dependency injection
type UserService interface {
    GetUser(ctx context.Context, id string) (*User, error)
}

// Create mock implementations for testing
type mockUserService struct {
    getUserFunc func(ctx context.Context, id string) (*User, error)
}
```

### Test Fixtures
```go
// Store test data in testdata/ directories
data, err := os.ReadFile("testdata/user.json")
```

### Test Helpers
```go
// Create helper functions for common setup
func setupTestDB(t *testing.T) *sql.DB {
    // ... setup test database
}
```

## 🏃 CI/CD Integration

Tests are automatically run in CI/CD pipeline:
1. Unit tests on every push
2. Integration tests on PR
3. E2E tests before deployment
4. Performance tests weekly

## 📊 Test Reports

Test results are available in:
- Console output during development
- JUnit XML for CI/CD integration
- HTML coverage reports in `coverage/`
- Performance profiles in `profiles/`

## 🚀 Performance Testing

### Load Testing
```bash
# Run load tests
make test-load

# Custom load test
go test -run=XXX -bench=. -benchtime=10s ./tests/performance/
```

### Profiling
```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./tests/...

# Memory profiling
go test -memprofile=mem.prof -bench=. ./tests/...

# View profiles
go tool pprof cpu.prof
```

## 🔍 Debugging Tests

### Verbose Output
```bash
go test -v ./tests/...
```

### Debug Specific Test
```bash
# Use Delve debugger
dlv test ./tests/observability/logger/ -- -test.run TestLoggerInitialization
```

### Test Timeout
```bash
# Set custom timeout (default 10m)
go test -timeout 30s ./tests/...
```

## 📚 Best Practices

1. **Keep tests independent** - Each test should be able to run in isolation
2. **Use descriptive names** - Test names should clearly describe what they test
3. **Test one thing** - Each test should verify a single behavior
4. **Avoid test pollution** - Clean up resources after tests
5. **Mock external dependencies** - Don't rely on external services in unit tests
6. **Use t.Helper()** - Mark helper functions to improve error messages
7. **Prefer table-driven tests** - Easier to add test cases
8. **Test edge cases** - Include boundary conditions and error paths
9. **Keep tests fast** - Unit tests should run in milliseconds
10. **Document complex tests** - Add comments explaining non-obvious test logic

## 🔗 References

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Go Testing Best Practices](https://golang.org/doc/tutorial/add-a-test)
- [Table Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [Test Coverage](https://go.dev/blog/cover)
