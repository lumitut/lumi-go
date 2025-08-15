# Architecture Decision Records (ADRs)

This directory contains Architecture Decision Records (ADRs) for the lumi-go microservice template. ADRs document important architectural decisions made during the project's development.

## What is an ADR?

An Architecture Decision Record captures a single architectural decision and its rationale. Each ADR describes:
- The context and problem statement
- The decision made
- The consequences of that decision
- Alternatives that were considered

## ADR Index

| ADR | Title | Status | Decision |
|-----|-------|--------|----------|
| [ADR-001](./ADR001-framework.md) | Go Web Framework | Accepted | **Gin** - High-performance HTTP framework with excellent developer experience |
| [ADR-002](./ADR002-rpc.md) | RPC Framework | Accepted | **Connect** - gRPC-compatible with browser support and debugging features |
| ADR-003 | Database Access Pattern | Proposed | TBD - sqlc vs ORM |
| [ADR-004](./ADR004-observability.md) | Observability Stack | Accepted | **OpenTelemetry + Prometheus + Jaeger + Grafana** - Vendor-neutral, comprehensive observability |
| [ADR-005](./ADR005-logging.md) | Logging Strategy | Accepted | **Zap** - High-performance structured logging with correlation |
| ADR-006 | SLI/SLO Definition | Proposed | TBD |
| ADR-007 | Security and Compliance | Proposed | TBD |
| ADR-008 | Testing Strategy | Proposed | TBD |
| ADR-009 | CI/CD Pipeline | Proposed | TBD |
| ADR-010 | Configuration Management | Proposed | TBD |

## ADR Template

Use [template.md](./template.md) when creating new ADRs. Copy it to a new file with the naming convention: `ADR{NUMBER}-{brief-description}.md`

## Quick Decisions Summary

### Technology Stack
- **Language**: Go 1.22+
- **HTTP Framework**: Gin
- **RPC Framework**: Connect (gRPC-compatible)
- **Logging**: Zap
- **Metrics**: Prometheus
- **Tracing**: Jaeger via OpenTelemetry
- **Visualization**: Grafana

### Key Principles
1. **Performance First**: Optimize for low latency and high throughput
2. **Security by Default**: Secure configurations out of the box
3. **Observability**: Comprehensive monitoring from day one
4. **Developer Experience**: Provide excellent tooling and documentation
5. **Vendor Independence**: Avoid lock-in, use open standards

## Creating a New ADR

1. Copy the template:
   ```bash
   cp template.md ADR{NUMBER}-{description}.md
   ```

2. Fill in the sections:
   - **Status**: Proposed / Accepted / Deprecated / Superseded
   - **Context**: Why was this decision needed?
   - **Decision**: What was decided and why?
   - **Consequences**: What are the trade-offs?
   - **Alternatives**: What else was considered?

3. Submit for review via pull request

4. Update this README index

## Resources

- [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
- [ADR Tools](https://github.com/npryce/adr-tools)
- [ADR GitHub Organization](https://adr.github.io/)
