# Release Notes

## [2.0.0] - 2024 - MAJOR SIMPLIFICATION RELEASE 🎯

### ⚡ Breaking Changes
- **Simplified Configuration**: Moved from detailed service configs to simple client URLs
- **Docker Compose**: Removed embedded infrastructure services (PostgreSQL, Redis, Grafana, etc.)
- **Configuration Structure**: Changed from nested service configs to `clients.*` pattern
- **Environment Variables**: Now use `LUMI_` prefix instead of direct names

### ✨ New Features
- **Lean Architecture**: Pure Go microservice without embedded dependencies
- **Optional Clients**: All external services are now optional and configured via simple URLs
- **Improved Test Structure**: Tests run without any external dependencies
- **Hot Reload Support**: Added Air configuration for development
- **Comprehensive Documentation**: New guides for development, quickstart, and external services

### 🔧 Improvements
- **Faster Startup**: Service starts immediately without waiting for databases
- **Smaller Docker Image**: Reduced image size by removing unnecessary components
- **Better Separation of Concerns**: External services managed independently
- **Simplified Configuration**: Single JSON file + environment variables
- **Improved CI/CD**: Added GitHub Actions workflows for CI, security, and releases

### 📚 Documentation
- Added comprehensive development guide
- Created quickstart guide for getting started in 5 minutes
- Added external services integration guide
- Updated API documentation (OpenAPI and Proto)
- Improved README with clear architecture overview

### 🐛 Bug Fixes
- Fixed test compilation issues with new config structure
- Resolved linting errors in test files
- Fixed configuration validation

### 🔄 Migration Guide

#### Configuration Changes
```bash
# Old format
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres

# New format
LUMI_CLIENTS_DATABASE_ENABLED=true
LUMI_CLIENTS_DATABASE_URL=postgres://postgres@localhost:5432/db
```

#### Docker Compose
```bash
# Old: Everything in one compose file
docker-compose up

# New: Lean service only
docker-compose up
# Optional: Run external services separately
cd ../lumi-postgres && docker-compose up -d
```

#### Code Changes
```go
// Old
cfg.Database.Host
cfg.Redis.Host

// New
if dbURL, enabled := cfg.GetDatabaseURL(); enabled {
    // Use database
}
```

---

## [1.5.0] - 2024-01-15

### Added
- gRPC service support with Protocol Buffers
- OpenAPI 3.0 specification
- Prometheus metrics endpoint
- Health and readiness checks
- Distributed tracing with OpenTelemetry

### Changed
- Improved middleware chain
- Enhanced error handling
- Updated dependencies

### Fixed
- Race condition in concurrent requests
- Memory leak in logger
- Configuration validation issues

---

## [1.4.0] - 2023-12-01

### Added
- Rate limiting middleware
- CORS configuration
- Request ID tracking
- Structured logging with zap

### Changed
- Migrated from logrus to zap for better performance
- Improved configuration management with Viper
- Enhanced test coverage

### Fixed
- Panic recovery in middleware
- JSON parsing errors
- Database connection pooling issues

---

## [1.3.0] - 2023-10-15

### Added
- Docker multi-stage build
- Kubernetes deployment manifests
- Helm chart support
- GitHub Actions CI/CD

### Changed
- Optimized Docker image size
- Improved build process
- Updated Go version to 1.21

### Fixed
- Docker build cache issues
- Helm template errors
- CI pipeline failures

---

## [1.2.0] - 2023-08-20

### Added
- Database migration support
- Redis caching layer
- Session management
- Feature flags

### Changed
- Refactored repository pattern
- Improved error handling
- Enhanced validation

### Fixed
- SQL injection vulnerabilities
- Cache invalidation bugs
- Session timeout issues

---

## [1.1.0] - 2023-06-10

### Added
- Integration tests
- Benchmark tests
- Load testing scripts
- Performance profiling

### Changed
- Optimized database queries
- Improved response times
- Reduced memory usage

### Fixed
- Goroutine leaks
- Deadlock conditions
- Resource exhaustion

---

## [1.0.0] - 2023-04-01 - Initial Release

### Added
- Basic HTTP server with Gin framework
- Configuration management
- PostgreSQL integration
- Basic middleware (logging, recovery)
- Unit tests
- Docker support
- Makefile for common tasks
- Basic documentation

### Features
- RESTful API structure
- Environment-based configuration
- Database connection pooling
- Graceful shutdown
- Health checks
- Structured project layout

---

## Version History

| Version | Date | Type | Description |
|---------|------|------|-------------|
| 2.0.0 | 2024 | Major | Architecture simplification |
| 1.5.0 | 2024-01-15 | Minor | gRPC and observability |
| 1.4.0 | 2023-12-01 | Minor | Enhanced middleware |
| 1.3.0 | 2023-10-15 | Minor | Kubernetes support |
| 1.2.0 | 2023-08-20 | Minor | Database and caching |
| 1.1.0 | 2023-06-10 | Minor | Testing improvements |
| 1.0.0 | 2023-04-01 | Major | Initial release |

---

## Upgrade Notes

### From 1.x to 2.0

**⚠️ Breaking Changes - Action Required**

1. **Update Configuration Files**
   - Migrate from old config format to new simplified format
   - Update environment variables to use `LUMI_` prefix

2. **Update Docker Compose**
   - External services are no longer included
   - Set up external services separately if needed

3. **Update Code**
   - Change database/Redis access to use new client pattern
   - Update configuration structs in your code

4. **Update CI/CD**
   - Review and update deployment scripts
   - Update environment variable references

### Deprecations

- Direct database configuration fields (use connection URLs instead)
- Embedded infrastructure services in docker-compose
- Old environment variable format without `LUMI_` prefix

### Removed

- `deploy/grafana/` directory (use lumi-grafana template)
- Complex database/Redis configuration options
- Built-in infrastructure service management

---

## Support

For issues and questions:
- GitHub Issues: [github.com/lumitut/lumi-go/issues](https://github.com/lumitut/lumi-go/issues)
- Discussions: [github.com/lumitut/lumi-go/discussions](https://github.com/lumitut/lumi-go/discussions)

## Contributors

Thanks to all contributors who have helped shape this project!

## License

MIT License - see LICENSE file for details
