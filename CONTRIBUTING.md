# Contributing to lumi-go

Thanks for helping improve **lumi-go** — Lumitut’s production-grade Go service template.

This guide covers how to propose changes, our development workflow, coding standards, and release practices.

---

## Ground rules

- Discuss larger changes in an issue first.
- Keep handlers thin; put business logic in `internal/service`.
- No secrets in code or config files committed to the repo.
- Add or update ADRs (`/docs/adr/`) for significant decisions.
- Write tests for new behavior and keep CI green.

---

## Development workflow

1. **Fork/branch**
   - Branch from `main` using a descriptive name: `feat/<topic>`, `fix/<bug>`, `docs/<area>`.

2. **Local setup**
   - Ensure toolchain versions match the README prerequisites.
   - Copy `.env.example` → `.env` and adjust for local dev.
   - Bring up local infra via Docker Compose (Postgres, Redis, OTel, Prometheus, Grafana).

3. **Make small, focused changes**
   - Follow the repository layout conventions.
   - Update or add ADRs if you’re changing architecture or dependencies.

4. **Quality gates (run locally)**
   - Format and tidy modules.
   - Lint and static analysis.
   - Run tests with race detector and measure coverage.
   - Generate code if you touched APIs or schemas and ensure a clean working tree.

5. **Open a Pull Request**
   - Fill out the PR template.
   - Link to the issue.
   - Describe the change, testing performed, and any rollout or migration notes.

6. **Review & merge**
   - One approval from CODEOWNERS is required.
   - Squash merge using Conventional Commits (see below).

---

## Conventional Commits

Use Conventional Commit prefixes to keep history and changelogs clean:

- `feat:` new functionality
- `fix:` bug fix
- `docs:` documentation only
- `refactor:` code change that neither fixes a bug nor adds a feature
- `perf:` performance improvement
- `test:` adding or updating tests
- `build:` build system or dependencies
- `ci:` CI configuration or scripts
- `chore:` other changes that don’t modify src or tests
- `revert:` revert a previous commit

Examples:
- `feat(http): add request-id middleware`
- `fix(repo): correct transaction rollback on error`

---

## Coding standards

- **Formatting:** gofmt/goimports; use gofumpt for stricter formatting.
- **Linting:** golangci-lint with the project’s config.
- **EditorConfig:** editors should respect `.editorconfig`.
- **Errors:** prefer wrapped errors with context; map to stable error codes at the transport edge.
- **Logging:** structured JSON via zap; include correlation fields (request_id, trace_id).
- **Telemetry:** instrument new components (HTTP/RPC handlers, DB calls) with OTel.
- **Validation:** use typed models and validator constraints; fail fast on invalid input.
- **Security:** no PII in logs; validate JWT `iss`, `aud`, `exp`; use least privilege defaults.

---

## Tests

- **Unit tests:** domain logic and services.
- **Transport tests:** HTTP (Gin) and RPC (Connect) using in-memory servers.
- **Integration tests:** repositories using ephemeral Postgres; apply migrations in setup.
- **Contract tests:** when OpenAPI/Proto change, regenerate clients/servers and run conformance tests.
- **Race & coverage:** run with `-race` and enforce project coverage thresholds in CI.

---

## Documentation

- Update `README.md` when user-facing behavior or setup changes.
- Add or update ADRs in `docs/adr/` for notable decisions (template available).
- Keep `TODO.md` in sync if you add new phases or tasks.

---

## Dependencies

- Prefer standard library first.
- Justify new dependencies in an ADR (scope, maintenance, alternatives).
- Pin versions; rely on Dependabot or similar for updates.
- Remove unused dependencies promptly.

---

## Security

- Never commit secrets; use environment variables or secret managers.
- Report vulnerabilities via the instructions in `SECURITY.md`.
- Keep third-party libraries up to date; address high/critical CVEs promptly.

---

## Releases

- Template uses SemVer.
- Phased releases (`v0.0.x`) culminate in `v1.0.0` per `TODO.md`.
- CI creates images and Helm charts; release notes summarize changes and any migrations.

---

## Code of Conduct

Be respectful and constructive. Disagreements are resolved through evidence and empathy. Escalate via CODEOWNERS if needed.

---

## Getting help

- Open an issue with repro steps and environment details.
- Tag the area in the title: `[http]`, `[rpc]`, `[repo]`, `[obs]`, `[helm]`, `[ci]`, etc.
