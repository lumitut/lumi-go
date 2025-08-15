# Quick Start Guide

Get lumi-go running in under 5 minutes.

## Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Make

Don't have these? See [engineering setup](./engineering.md).

## 1. Clone & Setup (30 seconds)

```bash
# Clone repository
git clone https://github.com/lumitut/lumi-go.git
cd lumi-go

# Install tools
make init
```

## 2. Start Services (2 minutes)

```bash
# Start everything with one command
make up
```

Wait for services to be healthy. You'll see:
```
✓ PostgreSQL is ready
✓ Redis is ready
✓ All services started successfully!
```

## 3. Run Application (30 seconds)

Open a new terminal:
```bash
# Run with hot-reload
make run
```

## 4. Verify It Works

```bash
# Check health
curl http://localhost:8080/healthz
# Should return: {"status":"healthy"}

# Check readiness
curl http://localhost:8080/readyz
# Should return: {"status":"ready"}

# Check metrics
curl http://localhost:9090/metrics
# Should return Prometheus metrics
```

## 5. Explore

| What | Where | Credentials |
|------|-------|-------------|
| **API Endpoints** | http://localhost:8080 | - |
| **Metrics** | http://localhost:9090/metrics | - |
| **Grafana Dashboards** | http://localhost:3000 | admin/admin |
| **Jaeger Tracing** | http://localhost:16686 | - |
| **Prometheus** | http://localhost:9091 | - |

## Next Steps

### Make Changes

1. Edit code in `internal/`
2. Save file
3. Watch automatic rebuild
4. Test your changes

### Run Tests

```bash
# Quick test
make test-short

# Full test suite
make test
```

### View Logs

```bash
# Application logs
docker-compose logs -f app

# All services
make logs
```

### Stop Everything

```bash
make down
```

## Common Commands

| Task | Command |
|------|---------|
| Start services | `make up` |
| Stop services | `make down` |
| Run app | `make run` |
| Run tests | `make test` |
| View logs | `make logs` |
| Database console | `./scripts/local.sh db` |
| Redis console | `./scripts/local.sh redis` |
| Clean everything | `make clean` |

## Quick Debugging

### Service Won't Start?

```bash
# Check Docker
docker ps

# Restart services
make restart

# Check logs
docker-compose logs postgres
```

### Port Already Used?

```bash
# Find what's using port 8080
lsof -i :8080

# Kill it
kill -9 <PID>
```

### Database Issues?

```bash
# Reset database
make migrate-reset

# Check connection
psql "postgres://lumigo:lumigo@localhost:5432/lumigo?sslmode=disable" -c "SELECT 1"
```

## Project Structure

```
lumi-go/
├── cmd/server/        # Application entrypoint
├── internal/          # Business logic
│   ├── service/       # Core services
│   ├── httpapi/       # HTTP handlers
│   └── repo/          # Data layer
├── migrations/        # Database migrations
├── deploy/           # Deployment configs
└── scripts/          # Utility scripts
```

## Development Workflow

```mermaid
graph LR
    A[Code] --> B[Save]
    B --> C[Auto Reload]
    C --> D[Test]
    D --> E{Pass?}
    E -->|Yes| F[Commit]
    E -->|No| A
    F --> G[Push]
```

## Getting Help

- 📖 [Full Documentation](./development.md)
- 🏗️ [Architecture Guide](./architecture.md)
- 🔧 [Engineering Setup](./engineering.md)
- 🐛 [Troubleshooting](./troubleshooting.md)
- 💬 [GitHub Issues](https://github.com/lumitut/lumi-go/issues)

## Tips

1. **Use Make targets** - They handle complexity for you
2. **Check service health** - Use `/healthz` and `/readyz`
3. **Watch the logs** - `make logs` is your friend
4. **Keep it simple** - Start with `make up` and `make run`

---

**Ready to build?** You're all set! 🚀
