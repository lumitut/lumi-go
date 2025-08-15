# Tools Reference

Complete reference for all tools used in the lumi-go project.

## Core Tools

### Go (1.22+)
Primary programming language
- **Install:** `brew install go@1.22`
- **Verify:** `go version`
- **Docs:** https://go.dev/doc/

### Docker (24.0+)
Container runtime
- **Install:** `brew install docker`
- **Verify:** `docker --version`
- **Docs:** https://docs.docker.com/

### Docker Compose (2.23+)
Container orchestration
- **Install:** `brew install docker-compose`
- **Verify:** `docker-compose --version`
- **Docs:** https://docs.docker.com/compose/

### Make (4.3+)
Build automation
- **Install:** Built-in on Unix, `brew install make` on macOS
- **Verify:** `make --version`
- **Usage:** `make help`

## Kubernetes Tools

### kubectl (1.28+)
Kubernetes CLI
- **Install:** `brew install kubectl`
- **Verify:** `kubectl version --client`
- **Config:** `~/.kube/config`
- **Docs:** https://kubernetes.io/docs/reference/kubectl/

### Helm (3.13+)
Kubernetes package manager
- **Install:** `brew install helm`
- **Verify:** `helm version`
- **Charts:** `deploy/helm/`
- **Docs:** https://helm.sh/docs/

### kind (0.20+)
Local Kubernetes clusters
- **Install:** `brew install kind`
- **Create cluster:** `kind create cluster --name lumi-go`
- **Delete cluster:** `kind delete cluster --name lumi-go`
- **Docs:** https://kind.sigs.k8s.io/

## Code Generation

### buf (1.28.1)
Protocol buffer toolchain
- **Install:** `brew install buf`
- **Config:** `buf.yaml`
- **Generate:** `buf generate`
- **Lint:** `buf lint`
- **Docs:** https://buf.build/docs/

### protoc (3.21+)
Protocol buffer compiler
- **Install:** `brew install protobuf`
- **Verify:** `protoc --version`
- **Usage:** `protoc --go_out=. *.proto`

### sqlc (1.25.0)
SQL to Go code generator
- **Install:** `brew install sqlc`
- **Config:** `sqlc.yaml`
- **Generate:** `sqlc generate`
- **Docs:** https://docs.sqlc.dev/

### wire (0.5.0)
Dependency injection
- **Install:** `go install github.com/google/wire/cmd/wire@latest`
- **Generate:** `wire ./...`
- **Files:** `wire.go`, `wire_gen.go`
- **Docs:** https://github.com/google/wire

### mockery (2.38+)
Mock generation
- **Install:** `go install github.com/vektra/mockery/v2@latest`
- **Generate:** `mockery --all`
- **Config:** `.mockery.yaml`
- **Docs:** https://vektra.github.io/mockery/

## Database Tools

### golang-migrate (4.16+)
Database migrations
- **Install:** `brew install golang-migrate`
- **Create:** `migrate create -ext sql -dir migrations -seq name`
- **Up:** `migrate -path migrations -database $DATABASE_URL up`
- **Down:** `migrate -path migrations -database $DATABASE_URL down`
- **Docs:** https://github.com/golang-migrate/migrate

### psql
PostgreSQL client
- **Install:** `brew install postgresql`
- **Connect:** `psql $DATABASE_URL`
- **Docker:** `docker-compose exec postgres psql -U lumigo`

### redis-cli
Redis client
- **Install:** `brew install redis`
- **Connect:** `redis-cli -h localhost`
- **Docker:** `docker-compose exec redis redis-cli`

## Code Quality

### golangci-lint (1.55.2)
Linters aggregator
- **Install:** `go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2`
- **Config:** `.golangci.yml`
- **Run:** `golangci-lint run`
- **Fix:** `golangci-lint run --fix`
- **Docs:** https://golangci-lint.run/

### gofumpt (0.5.0)
Stricter gofmt
- **Install:** `go install mvdan.cc/gofumpt@latest`
- **Format:** `gofumpt -w .`
- **Check:** `gofumpt -l .`
- **Docs:** https://github.com/mvdan/gofumpt

### goimports
Import management
- **Install:** `go install golang.org/x/tools/cmd/goimports@latest`
- **Format:** `goimports -w .`
- **Check:** `goimports -l .`

## Security Tools

### gosec (2.18+)
Security analyzer
- **Install:** `go install github.com/securego/gosec/v2/cmd/gosec@latest`
- **Scan:** `gosec ./...`
- **Report:** `gosec -fmt sarif -out results.sarif ./...`
- **Docs:** https://github.com/securego/gosec

### govulncheck
Vulnerability checker
- **Install:** `go install golang.org/x/vuln/cmd/govulncheck@latest`
- **Check:** `govulncheck ./...`
- **Database:** Updated automatically
- **Docs:** https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck

### gitleaks (8.18+)
Secret scanner
- **Install:** `brew install gitleaks`
- **Config:** `.gitleaks.toml`
- **Scan:** `gitleaks detect`
- **Pre-commit:** `gitleaks protect`
- **Docs:** https://github.com/gitleaks/gitleaks

### trivy (0.48+)
Container scanner
- **Install:** `brew install trivy`
- **Scan image:** `trivy image lumi-go:latest`
- **Scan filesystem:** `trivy fs .`
- **Scan config:** `trivy config .`
- **Docs:** https://aquasecurity.github.io/trivy/

## Development Tools

### air (1.49+)
Hot reload
- **Install:** `go install github.com/cosmtrek/air@latest`
- **Config:** `.air.toml`
- **Run:** `air`
- **Debug:** `air -d`
- **Docs:** https://github.com/cosmtrek/air

### delve (1.21+)
Go debugger
- **Install:** `go install github.com/go-delve/delve/cmd/dlv@latest`
- **Debug:** `dlv debug ./cmd/server`
- **Attach:** `dlv attach <PID>`
- **Test:** `dlv test`
- **Docs:** https://github.com/go-delve/delve

### evans (0.10+)
gRPC client
- **Install:** `brew install evans`
- **REPL:** `evans --host localhost --port 8081 repl`
- **CLI:** `evans --host localhost --port 8081 cli call service.Method`
- **Docs:** https://github.com/ktr0731/evans

### grpcurl (1.8+)
gRPC curl
- **Install:** `brew install grpcurl`
- **List:** `grpcurl -plaintext localhost:8081 list`
- **Call:** `grpcurl -plaintext -d '{}' localhost:8081 service/Method`
- **Docs:** https://github.com/fullstorydev/grpcurl

## Observability Tools

### OTEL Collector (0.91+)
Telemetry collector
- **Config:** `deploy/docker/otel-collector-config.yaml`
- **Receivers:** OTLP, Prometheus
- **Exporters:** Jaeger, Prometheus
- **Docs:** https://opentelemetry.io/docs/collector/

### Prometheus (2.48+)
Metrics storage
- **Config:** `deploy/docker/prometheus.yml`
- **UI:** http://localhost:9091
- **Query:** PromQL
- **Docs:** https://prometheus.io/docs/

### Grafana (10.2+)
Visualization
- **Config:** `deploy/docker/grafana-datasource.yml`
- **UI:** http://localhost:3000
- **Default:** admin/admin
- **Docs:** https://grafana.com/docs/

### Jaeger (1.52+)
Distributed tracing
- **UI:** http://localhost:16686
- **Ports:** 14268 (HTTP), 14250 (gRPC)
- **Storage:** In-memory (dev)
- **Docs:** https://www.jaegertracing.io/docs/

## Version Control

### Git (2.40+)
Version control
- **Config:** `.gitconfig`
- **Ignore:** `.gitignore`
- **Hooks:** `.git/hooks/`
- **Docs:** https://git-scm.com/doc

### pre-commit (3.6+)
Git hooks framework
- **Install:** `pip install pre-commit`
- **Config:** `.pre-commit-config.yaml`
- **Install hooks:** `pre-commit install`
- **Run:** `pre-commit run --all-files`
- **Docs:** https://pre-commit.com/

### GitHub CLI (2.40+)
GitHub operations
- **Install:** `brew install gh`
- **Auth:** `gh auth login`
- **PR:** `gh pr create`
- **Issues:** `gh issue list`
- **Docs:** https://cli.github.com/

## Package Management

### Go Modules
Go dependency management
- **Init:** `go mod init`
- **Download:** `go mod download`
- **Tidy:** `go mod tidy`
- **Vendor:** `go mod vendor`
- **Docs:** https://go.dev/ref/mod

### npm/yarn (optional)
For web assets
- **Install npm:** `brew install node`
- **Install yarn:** `npm install -g yarn`
- **Install deps:** `npm install` or `yarn install`

## Testing Tools

### go test
Built-in testing
- **Run all:** `go test ./...`
- **Verbose:** `go test -v ./...`
- **Coverage:** `go test -cover ./...`
- **Race:** `go test -race ./...`
- **Docs:** https://pkg.go.dev/testing

### testify
Test assertions
- **Install:** `go get github.com/stretchr/testify`
- **Assert:** `assert.Equal(t, expected, actual)`
- **Require:** `require.NoError(t, err)`
- **Docs:** https://github.com/stretchr/testify

### httpexpect
HTTP testing
- **Install:** `go get github.com/gavv/httpexpect/v2`
- **Usage:** Integration tests
- **Docs:** https://github.com/gavv/httpexpect

## Utility Tools

### jq
JSON processor
- **Install:** `brew install jq`
- **Pretty:** `curl localhost:8080/api | jq .`
- **Filter:** `jq '.data.items[]'`
- **Docs:** https://jqlang.github.io/jq/

### yq
YAML processor
- **Install:** `brew install yq`
- **Read:** `yq '.spec.containers[0].image' deployment.yaml`
- **Write:** `yq '.version = "1.2.3"' -i Chart.yaml`
- **Docs:** https://mikefarah.gitbook.io/yq/

### curl
HTTP client
- **GET:** `curl http://localhost:8080/healthz`
- **POST:** `curl -X POST -d '{}' http://localhost:8080/api`
- **Headers:** `curl -H "Authorization: Bearer token"`
- **Docs:** https://curl.se/docs/

### httpie
Modern HTTP client
- **Install:** `brew install httpie`
- **GET:** `http localhost:8080/healthz`
- **POST:** `http POST localhost:8080/api key=value`
- **Docs:** https://httpie.io/docs

## Tool Configuration Files

| Tool | Config File | Purpose |
|------|------------|---------|
| Go | `go.mod`, `go.sum` | Dependencies |
| Docker | `Dockerfile` | Container build |
| Docker Compose | `docker-compose.yml` | Local orchestration |
| Air | `.air.toml` | Hot reload |
| Buf | `buf.yaml` | Protocol buffers |
| SQLC | `sqlc.yaml` | SQL generation |
| Golangci-lint | `.golangci.yml` | Linting rules |
| Gitleaks | `.gitleaks.toml` | Secret scanning |
| Pre-commit | `.pre-commit-config.yaml` | Git hooks |
| Helm | `Chart.yaml`, `values.yaml` | Kubernetes deployment |
| Prometheus | `prometheus.yml` | Metrics collection |
| Grafana | `grafana-datasource.yml` | Data sources |

## Environment Variables

Common environment variables used by tools:

```bash
# Go
GOPATH=$HOME/go
GOROOT=/usr/local/go
GO111MODULE=on
GOPROXY=https://proxy.golang.org

# Docker
DOCKER_BUILDKIT=1
COMPOSE_DOCKER_CLI_BUILD=1

# Database
DATABASE_URL=postgres://user:pass@localhost:5432/db

# Application
ENV=dev
LOG_LEVEL=debug
```

## Troubleshooting Tools

### System Tools

```bash
# Check ports
lsof -i :8080
netstat -an | grep 8080

# Check processes
ps aux | grep lumi-go
pgrep -f lumi-go

# Check resources
top
htop
docker stats
```

### Debug Commands

```bash
# Go debugging
go env
go version
go list -m all

# Docker debugging
docker version
docker info
docker inspect <container>
docker logs <container>

# Network debugging
ping localhost
telnet localhost 8080
nc -zv localhost 8080
```

## Getting Help

- Run `<tool> --help` for command options
- Check `make help` for project commands
- See tool documentation links above
- Ask in project discussions or issues
