# TODO.md — lumi-go Template Build Checklist

> Use this checklist to deliver the lumi-go in small, deployable phases.
> Each phase ends with a tagged release and an AWS deployable artefact for testing.
> Versioning plan: Phase 0 → `v0.0.1`, Phase 1 → `v0.0.2`, … Phase 10 → `v0.0.11`, Phase 11 → `v1.0.0`.

---

## Phase 0 — Foundations → Release `v0.0.1`

- [x] Create repository with LICENSE, README, CONTRIBUTING, CODEOWNERS, SECURITY, ADR index.
- [x] Configure branch protections (required reviews, status checks).
- [x] Define PR templates and labels (obs, security, docs, breaking-change).
- [ ] Document engineering setup (toolchain versions for Go, Docker, Helm, kubectl, buf, sqlc, golangci-lint, OTEL Collector).
- [x] Enable dependency, container, and secret scanning in the repo.
- [x] Create initial `deploy/helm` skeleton and `deploy/docker` skeleton (placeholders only; no code).
- [ ] Tag and publish release `v0.0.1`.


---

## Phase 1 — Local Developer Experience (LDX) → Release `v0.0.2`

* [ ] Add docker-compose services for Postgres, Redis, OTEL Collector, Prometheus, Grafana.
* [ ] Provide a single command to start/stop local infra and a DB seed script (placeholder).
* [x] Establish repo directory layout (cmd, internal/*, api/*, deploy/\*, migrations) with README stubs.
* [ ] Validate clean local bring-up on a fresh machine.
- [ ] AWS prep: create ECR repository, dev EKS namespace, and CI OIDC/IAM role for future pushes.
* [ ] Tag and publish release `v0.0.2`.
* [ ] AWS smoke: publish image to ECR; install/update Helm chart to dev with ops placeholders; verify pod runs.

---

## Phase 2 — Observability First → Release `v0.0.3`

* [ ] Wire structured logging standard (JSON, correlation fields) — documented contract.
* [ ] Register default metrics and Prometheus scrape endpoint.
* [ ] Configure OTLP export to local collector; define service resource attrs (name, env, version).
* [ ] Provide a starter Grafana dashboard JSON (latency, errors, RPS placeholders).
* [ ] Tag and publish release `v0.0.3`.
* [ ] AWS smoke: deploy to dev; confirm `/metrics` scraped by Prometheus and traces reach collector.

---

## Phase 3 — Transport Surfaces → Release `v0.0.4`

* [ ] HTTP front door with Gin (release mode in prod), standard middleware list (request-id, real-IP, recovery, access log, rate limit, CORS off-by-default, OTEL).
* [ ] Mount ops routes: `/healthz`, `/readyz`, `/metrics`, `/debug/pprof` (restricted in non-dev).
* [ ] RPC front door with Connect on separate listener (interceptors: tracing, logging, auth).
* [ ] Graceful lifecycle and shutdown with readiness flip on drain.
* [ ] Tag and publish release `v0.0.4`.
* [ ] AWS smoke: deploy to dev; verify `/healthz` 200, `/readyz` flips ready, and RPC stub responds.

---

## Phase 4 — Cross-Cutting Concerns → Release `v0.0.5`

* [ ] Configuration system: env-first, defaults, validation, redacted boot log; publish `.env.example`.
* [ ] AuthN/Z pattern: JWT verification via JWKS; route group protections; claims in context.
* [ ] Validation and error model: stable error codes, HTTP mappings, response envelope.
* [ ] Outbound policy: timeouts, retry/backoff, circuit breaker, rate limit; metrics per dependency.
* [ ] Tag and publish release `v0.0.5`.
* [ ] AWS smoke: deploy to dev; verify protected route denies without token and allows with valid token.

---

## Phase 5 — Data & Caching → Release `v0.0.6`

* [ ] Decide and document data layer choice (sqlc or ent) and when to use each.
* [ ] Add baseline schema and empty migrations; wrap pgx pool with metrics/tracing.
* [ ] Define repository interfaces separated from business logic.
* [ ] Caching strategy: in-proc (Ristretto) + shared (Redis) adapters; TTL and invalidation guidance.
* [ ] Tag and publish release `v0.0.6`.
* [ ] AWS smoke: deploy to dev; run a repo read/write against dev Postgres; confirm Redis connectivity.

---

## Phase 6 — Testing & Quality Gates → Release `v0.0.7`

* [ ] Establish testing pyramid and minimum coverage policy.
* [ ] Add transport tests (HTTP/RPC), integration tests with ephemeral DB, and contract test harness (OpenAPI/Proto).
* [ ] Enable race detector and benchmarks for hot paths (policy only; runnable placeholders).
* [ ] Configure golangci-lint and static analysis rules; pre-commit guidance.
* [ ] Tag and publish release `v0.0.7`.
* [ ] AWS smoke: run a minimal contract test against the dev deployment endpoint.

---

## Phase 7 — CI/CD & Environments → Release `v0.0.8`

* [ ] CI pipeline v1: lint → codegen verify → tests (race, coverage) → build (multi-arch) → container scan → push image → publish Helm chart.
* [ ] Configure caches for modules and codegen artefacts; fail on dirty tree after codegen.
* [ ] Define environments (dev, staging, prod) and promotion gates with manual approvals.
* [ ] Document rollback process and image pinning in Helm values.
* [ ] Tag and publish release `v0.0.8`.
* [ ] AWS smoke: CI builds and pushes to ECR; pipeline deploys to dev automatically on main.

---

## Phase 8 — Kubernetes, Helm, and Ops → Release `v0.0.9`

* [ ] Standard GMT Helm chart: ports, probes, resources, HPA, pod security, network policy.
* [ ] Secrets management: env/secret refs, rotation guidance, external secrets compatibility.
* [ ] Enforce non-root user, read-only root filesystem (where feasible), seccomp/profile, minimal caps.
* [ ] Provide values files per environment (dev, staging, prod).
* [ ] Tag and publish release `v0.0.9`.
* [ ] AWS smoke: helm upgrade in dev namespace with chart defaults; probes pass and HPA scales under load.

---

## Phase 9 — Flags, Telemetry Hygiene, and DX Polish → Release `v0.0.10`

* [ ] Integrate feature flag provider (Unleash or go-feature-flag); document kill-switch pattern.
* [ ] Telemetry hygiene policy: PII redaction rules, sampling, field naming conventions.
* [ ] DX enhancements: New Service Wizard checklist, ADR templates, onboarding guide, short screencast.
* [ ] Tag and publish release `v0.0.10`.
* [ ] AWS smoke: toggle a feature flag in dev and observe behavior change without redeploy.

---

## Phase 10 — Hardening & Performance → Release `v0.0.11`

* [ ] Performance baseline: record p50/p95/p99, error rate, CPU/mem at target RPS; set SLOs.
* [ ] Chaos drills: inject latency/failures in dependencies; verify breakers, timeouts, readiness flips, autoscaling.
* [ ] Supply chain: generate SBOMs, sign images and charts, verify attestations in CI.
* [ ] Tag and publish release `v0.0.11`.
* [ ] AWS smoke: run baseline load test in dev; confirm alerts fire when SLOs are breached.

---

## Phase 11 — Template Release & Scale-Out → Release `v1.0.0`

* [ ] Versioned docs: publish a minimal docs site (or Confluence space) with how-tos, playbooks, ADR examples.
* [ ] Create two pilot services from the template (HTTP-only; HTTP+RPC with DB/cache) and run full lifecycle to staging.
* [ ] Governance: establish owners group, monthly template upgrades, quarterly dependency refresh, breaking-change policy.
* [ ] Compile migration notes since `v0.0.1` and finalize release notes.
* [ ] Tag and publish release `v1.0.0`.
* [ ] AWS validation: both pilot services healthy in staging with dashboards and alerts passing; sign-off recorded.

---

## Continuous Items (track throughout)

* [ ] Keep ADRs updated for significant decisions.
* [ ] Review and prune dependencies quarterly.
* [ ] Rotate secrets regularly and on personnel changes.
* [ ] Audit dashboards/alerts after each phase’s deploy.
* [ ] Run security scans and address findings within SLA.

---
