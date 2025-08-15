# lumi-go

A production-grade Go microservice template for Lumitut. It gives you a consistent, observable, and secure service skeleton with **Gin (HTTP/JSON)** and **Connect (RPC)** front doors, wired for **OpenTelemetry**, **Prometheus**, **Postgres/Redis**, and **Helm** deployments to AWS.

* Repo: [https://github.com/lumitut/lumi-go](https://github.com/lumitut/lumi-go)


---

## Features at a glance

* Fast HTTP via **Gin**, internal RPC via **Connect** (gRPC/gRPC-Web/JSON)
* One way to do **logging, metrics, tracing** (OTel → OTLP), **pprof**, and **health/readiness**
* **JWT** auth (JWKS), optional **SPIFFE/SPIRE** for mTLS
* Data layer: **pgx** with **sqlc** (or **ent**), **golang-migrate** migrations
* Caching: **Ristretto** (in-proc) + **Redis** (shared)
* Resilience: timeouts, retries/backoff, circuit breakers, rate limiting
* First-class DX: **docker-compose** local stack, **Makefile** targets, **golangci-lint**, **GitHub Actions**, **Helm** chart

---

## Tech stack

* Go ≥ 1.22
* HTTP: `gin-gonic/gin`
* RPC: `connectrpc.com/connect`
* OpenAPI: `deepmap/oapi-codegen` (optional)
* Protobuf: `buf.build` + `protoc-gen-connect-go` (optional)
* Logging: `uber-go/zap`
* Metrics: `prometheus/client_golang`
* Tracing: OpenTelemetry (`otel`, `otelgin`, `otelgrpc`) → OTLP
* Config: `envconfig` or `viper`
* Validation: `go-playground/validator/v10`
* Resilience: `cenkalti/backoff`, `sony/gobreaker`, `x/time/rate`
* Data: `jackc/pgx/v5` + **sqlc** or **ent**
* Migrations: `golang-migrate/migrate`
* Cache: `dgraph-io/ristretto`, `redis/go-redis/v9`
* Queue/Eventing (choose per service): Kafka (`segmentio/kafka-go`) or AWS SNS/SQS (`aws-sdk-go-v2`)

---

## Repository layout

```
lumi-go/
├─ cmd/server/                 # Entrypoint / DI wiring
├─ internal/
│  ├─ app/                     # Servers, wiring, lifecycle
│  ├─ config/                  # Config structs, load/validate
│  ├─ httpapi/                 # Gin routes & handlers
│  ├─ rpcapi/                  # Connect handlers & stubs
│  ├─ middleware/              # Auth, rate-limit, logging, recovery
│  ├─ domain/                  # Business interfaces & models
│  ├─ service/                 # Domain implementations
│  ├─ repo/                    # Postgres repositories (sqlc/ent)
│  ├─ cache/                   # Redis/Ristretto adapters
│  ├─ clients/                 # Outbound HTTP/RPC with resilience
│  ├─ observability/           # Zap, OTel, Prom, pprof
│  └─ version/                 # Build info (git SHA, build time)
├─ api/
│  ├─ openapi/                 # OpenAPI → oapi-codegen
│  └─ proto/                   # Protobuf → buf generate
├─ migrations/                 # SQL migrations
├─ deploy/
│  ├─ docker/                  # Dockerfile, scripts
│  └─ helm/                    # Chart, values
├─ docker-compose.yml          # Local dev stack
├─ Makefile                    # DX targets
├─ .github/workflows/          # CI pipelines
└─ TODO.md                     # Phase plan & checklist
```

---

## Quickstart

### Prerequisites

* Go 1.22+
* Docker & Docker Compose
* Make (or run the equivalent scripts)
* Helm & kubectl (for deploy)
* Optional generators: `buf`, `sqlc`, `oapi-codegen`
* Optional linting: `golangci-lint`

### 1) Clone and configure

* Copy `.env.example` to `.env` and set local values:

  * `LUMI_HTTP_ADDR=:8080`
  * `LUMI_RPC_ADDR=:8081`
  * `LUMI_PROM_ADDR=:9090`
  * `LUMI_PG_URL=postgres://...`
  * `LUMI_REDIS_URL=redis://...`
  * `OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318`

### 2) Bring up local infra

* Start `docker-compose` services (Postgres, Redis, OTel Collector, Prometheus, Grafana).

### 3) Run the service

* Start the HTTP and RPC servers.
* Visit:

  * `GET /healthz` → liveness
  * `GET /readyz` → readiness (checks DB/Redis)
  * `GET /metrics` → Prometheus
  * `GET /debug/pprof/*` → profiling (dev only)

### 4) Observe

* Confirm traces in your OTel backend and metrics in Prometheus/Grafana.

---

## Configuration

Environment-first, with the **LUMI\_** prefix where applicable.

Common variables:

* `ENV` (dev|staging|prod), `LUMI_HTTP_ADDR`, `LUMI_RPC_ADDR`, `LUMI_PROM_ADDR`
* `LUMI_READ_TIMEOUT`, `LUMI_WRITE_TIMEOUT`, `LUMI_IDLE_TIMEOUT`, `LUMI_GRACE_PERIOD`
* `LUMI_LOG_LEVEL`
* `OTEL_EXPORTER_OTLP_ENDPOINT`
* `LUMI_PG_URL`, `LUMI_REDIS_URL`
* `LUMI_JWT_ISSUER`, `LUMI_JWT_AUDIENCE`, `LUMI_JWKS_URL`
* `LUMI_RATE_RPS`
* Circuit breaker: `LUMI_CB_FAILURES`, `LUMI_CB_TIMEOUT`
* Flags: `LUMI_FLAGS_PROVIDER` (unleash|none), `LUMI_FLAGS_URL`, `LUMI_FLAGS_TOKEN`
* Gin: `GIN_MODE` (set to `release` in prod)

---

## Observability

* **Logging:** JSON via zap; includes `service.name`, `env`, `version`, `trace_id`, `span_id`, `request_id`. PII is never logged.
* **Metrics:** Default Go/process collectors + request duration histograms; Prometheus scrape at `/metrics`.
* **Tracing:** Auto-instrumented Gin, Connect, pgx, Redis; exported via OTLP to the collector.
* **Dashboards:** A starter Grafana dashboard is included under `deploy/helm/grafana/` (import or auto-provision).

---

## AuthN/Z

* **JWT** (RS256) with JWKS; validate `iss`, `aud`, `exp`; attach claims to context.
* Public endpoints: `/healthz`, `/readyz`, `/metrics` (restrict in prod), `/debug/pprof` (dev only).
* Optional **SPIFFE/SPIRE** for mTLS between services.

---

## Data layer

Choose one per service (documented in an ADR):

* **sqlc:** Raw SQL as the source of truth; fast, explicit.
* **ent:** Schema-first modeling; good for complex relations.

Migrations via `golang-migrate`. Pooling and telemetry via pgx + OTel.

---

## Resilience & outbound policy

* Bounded **timeouts**, **retries with exponential backoff**, **circuit breakers**, and **rate limits** on all outbound HTTP/RPC.
* Per-dependency metrics for latency, error rate, and breaker states.

---

## Running tests

* Unit tests (domain & services), transport tests (HTTP/RPC), integration tests (repo with ephemeral DB), and contract tests (OpenAPI/Proto).
* Race detector and coverage thresholds enforced in CI.

---

## CI/CD

* **GitHub Actions**: lint → codegen verify → tests (race, coverage) → build (multi-arch) → container scan → push → publish Helm chart → environment promotion (manual gates).
* **Artifacts**: container image (GHCR/ECR), Helm chart, SBOM/signatures (in later phases).
* See `TODO.md` for the phased release plan (`v0.0.1` → `v1.0.0`).

---

## Deploying to AWS (dev)

High-level flow:

1. Build and push the image to **ECR**.
2. Configure cluster access and a dev namespace.
3. Install/upgrade the **Helm** chart in `deploy/helm` with your `values-dev.yaml`.
4. Verify probes (`/healthz`, `/readyz`), metrics scrape, and traces.

Security defaults: non-root user, minimal capabilities, NetworkPolicy, and secrets via environment/External Secrets.

---

## Versioning

* **SemVer**: incremental template releases aligned with the checklist in `TODO.md`.
* Phase tags: `v0.0.1`, `v0.0.2`, …, final `v1.0.0`.

---

## Contributing

* Open an issue before large changes.
* Keep transport thin; put logic in `internal/service`.
* Update ADRs for significant decisions.
* Run linters and tests locally before pushing.

---

## Roadmap

* Template generator for new services (scaffold prompts)
* Example HTTP-only and RPC-only variants
* Sample Kafka/SNS/SQS adapters and contracts
* Golden dashboards & alert pack per SLO tier

---

## License

MIT © Lumitut. See `LICENSE`.
