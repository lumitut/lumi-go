# Release Notes

## v0.0.2 - Local Developer Experience (LDX) - 2025-08-15

### 🎯 Phase 1 Complete: Local Developer Experience

This release establishes a comprehensive local development environment with full observability stack, automated tooling, and extensive documentation.

### ✨ Features

#### Infrastructure & Services
- **Docker Compose Stack**: Complete local environment with all services
  - PostgreSQL 16 with health checks and migrations
  - Redis 7 with persistence
  - OpenTelemetry Collector with full configuration
  - Prometheus for metrics collection
  - Grafana for visualization (pre-configured datasources)
  - Jaeger for distributed tracing
- **Database**: Initial schema with users, sessions, audit logs, feature flags, and API keys
- **Migrations**: golang-migrate setup with up/down migrations

#### Developer Experience
- **Hot Reload**: Air configuration for automatic rebuilds
- **Single Command Setup**: `make up` starts everything
- **Local Management Script**: `./scripts/local.sh` for environment control
- **Comprehensive Makefile**: 40+ targets for common tasks
- **Validation Scripts**: 
  - `verify-setup.sh` - Check tool installations
  - `validate-fresh.sh` - Fresh machine validation

#### Documentation
- **Engineering Setup Guide**: Complete toolchain documentation
- **Development Guide**: Detailed workflow documentation
- **Quick Start Guide**: 5-minute setup instructions
- **Tools Reference**: Comprehensive guide for all 40+ tools
- **Architecture Decision Records (ADRs)**:
  - ADR-001: Go Web Framework (Gin)
  - ADR-002: RPC Framework (Connect)
  - ADR-004: Observability Stack (OpenTelemetry)
  - ADR-005: Logging Strategy (Zap)

#### Configuration Files
- **OpenTelemetry Collector**: Full pipeline configuration
- **Prometheus**: Scraping configuration for all services
- **Grafana**: Datasource provisioning (Prometheus, Jaeger, PostgreSQL, Redis)
- **Air**: Hot-reload development configuration
- **Docker**: Multi-stage Dockerfiles for production and development

### 📋 What's Included

```
lumi-go/
├── .air.toml                        # Hot-reload configuration
├── .dockerignore                    # Docker build exclusions
├── .env.example                     # Environment template
├── docker-compose.yml               # Local services orchestration
├── Makefile                         # Build automation (40+ targets)
├── cmd/server/                      # Application entry point
├── internal/                        # Business logic structure
│   ├── app/                        # Application wiring
│   ├── config/                     # Configuration management
│   ├── httpapi/                    # HTTP handlers
│   ├── rpcapi/                     # RPC handlers
│   ├── middleware/                 # Cross-cutting concerns
│   ├── domain/                     # Business interfaces
│   ├── service/                    # Service implementations
│   ├── repo/                       # Data repositories
│   ├── cache/                      # Caching layer
│   ├── clients/                    # External clients
│   ├── observability/              # Telemetry setup
│   └── version/                    # Version info
├── api/                            # API definitions
│   ├── openapi/                    # OpenAPI specs
│   └── proto/                      # Protocol buffers
├── migrations/                      # Database migrations
│   ├── 000001_init_schema.up.sql  # Initial schema
│   ├── 000001_init_schema.down.sql # Rollback
│   └── README.md                   # Migration guide
├── deploy/                         # Deployment configurations
│   ├── docker/                     # Docker setup
│   │   ├── Dockerfile              # Production image
│   │   ├── Dockerfile.dev          # Development image
│   │   ├── build.sh               # Build automation
│   │   ├── otel-collector-config.yaml
│   │   ├── prometheus.yml
│   │   └── grafana-datasource.yml
│   └── helm/                       # Kubernetes charts
│       ├── Chart.yaml             # Chart metadata
│       ├── values.yaml            # Default values
│       └── templates/             # K8s manifests
├── scripts/                        # Utility scripts
│   ├── local.sh                   # Environment management
│   ├── seed.sql                   # Database seeding
│   ├── verify-setup.sh            # Tool verification
│   └── validate-fresh.sh          # Fresh machine test
└── docs/                          # Documentation
    ├── quickstart.md              # Quick start guide
    ├── engineering.md             # Setup instructions
    ├── development.md             # Development workflow
    ├── tools.md                   # Tools reference
    └── adr/                       # Architecture decisions
```

### 🔧 Prerequisites

| Tool | Minimum Version | Purpose |
|------|-----------------|---------|
| Go | 1.22+ | Primary language |
| Docker | 24.0+ | Container runtime |
| Docker Compose | 2.23+ | Service orchestration |
| Make | 4.3+ | Build automation |

### 🚀 Quick Start

```bash
# Clone repository
git clone https://github.com/lumitut/lumi-go.git
cd lumi-go

# Verify setup
./scripts/verify-setup.sh

# Start services
make up

# Run application with hot-reload
make run

# Check health
curl http://localhost:8080/healthz
```

### 📊 Services & Ports

| Service | Port | URL |
|---------|------|-----|
| API | 8080 | http://localhost:8080 |
| gRPC | 8081 | http://localhost:8081 |
| Metrics | 9090 | http://localhost:9090/metrics |
| Prometheus | 9091 | http://localhost:9091 |
| Grafana | 3000 | http://localhost:3000 (admin/admin) |
| Jaeger | 16686 | http://localhost:16686 |
| PostgreSQL | 5432 | localhost:5432 (lumigo/lumigo) |
| Redis | 6379 | localhost:6379 |

### 🧪 Validation

Run the fresh machine validation to ensure everything works:

```bash
./scripts/validate-fresh.sh
```

Expected output:
```
✅ ALL TESTS PASSED!
Total: 50+ | Passed: 50+ | Failed: 0
```

### 📝 Changes from v0.0.1

- Added complete Docker Compose stack with 7 services
- Implemented hot-reload development with Air
- Created 40+ Make targets for automation
- Added database migrations and seeding
- Wrote comprehensive documentation (5 guides)
- Created 4 Architecture Decision Records
- Implemented validation and verification scripts
- Added GitHub issue and PR templates
- Configured security scanning (Dependabot, CodeQL, Trivy)

### 🐛 Known Issues

- None reported

### 🔮 Next Release (v0.0.3)

Phase 2 will focus on Observability:
- Structured logging with Zap
- Prometheus metrics registration
- OpenTelemetry integration
- Grafana dashboards
- Correlation IDs across telemetry

### 📦 Installation

```bash
# Using this template
git clone https://github.com/lumitut/lumi-go.git my-service
cd my-service
rm -rf .git
git init
make init
make up
```

### 🤝 Contributors

- Platform Team (@lumitut/platform-team)
- Security Team (@lumitut/security-team)

### 📄 License

MIT License - See [LICENSE](LICENSE) file for details.

---

## v0.0.1 - Foundations - 2025-08-14

### 🎯 Phase 0 Complete: Repository Foundations

Initial repository setup with security configurations, documentation templates, and basic structure.

### Features
- Repository structure with LICENSE, README, CONTRIBUTING
- GitHub security features configuration
- PR and issue templates
- Dependency scanning with Dependabot
- Container scanning workflows
- Secret scanning configuration
- Initial Helm and Docker skeletons

---

For detailed changes, see the [commit history](https://github.com/lumitut/lumi-go/commits/main).
