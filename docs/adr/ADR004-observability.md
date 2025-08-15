# ADR 004: Observability Stack

## Status
**Accepted** - August 2025

## Context

Modern microservices require comprehensive observability to maintain reliability, performance, and security in production. The lumi-go template needs a robust observability stack that provides:

- **Metrics**: System and application performance indicators
- **Logging**: Structured event recording for debugging and audit
- **Tracing**: Distributed request flow visualization
- **Unified Telemetry**: Correlated data across all three pillars

Key requirements for our observability stack:
- Low performance overhead (<5% CPU, <50MB memory)
- Vendor-agnostic instrumentation
- Cloud-native and Kubernetes-ready
- Support for high cardinality metrics
- Distributed tracing across service boundaries
- Structured logging with correlation IDs
- Real-time alerting capabilities
- Cost-effective at scale
- Developer-friendly local setup
- GDPR/privacy compliance (PII handling)

The observability stack must integrate seamlessly with our chosen frameworks (Gin for HTTP, Connect for RPC) and support our polyglot future (potential Node.js, Python services).

## Decision

We have adopted **OpenTelemetry** as our unified observability framework with the following stack:

### Core Components:

1. **OpenTelemetry (OTEL)** - Instrumentation and collection
   - Vendor-neutral telemetry standard
   - Auto-instrumentation for frameworks
   - Unified SDK for metrics, logs, and traces

2. **Prometheus** - Metrics storage and querying
   - Time-series database for metrics
   - PromQL for powerful queries
   - Pull-based model with service discovery

3. **Jaeger** - Distributed tracing
   - Trace storage and visualization
   - OpenTelemetry native support
   - Sampling strategies for scale

4. **Grafana** - Visualization and dashboards
   - Unified view across data sources
   - Alerting and anomaly detection
   - Pre-built dashboards for common patterns

5. **Zap** - Structured logging
   - High-performance JSON logging
   - Zero-allocation in hot paths
   - Correlation with traces and metrics

### Architecture:

```
Application → OTEL SDK → OTEL Collector → Storage Backends
                ↓              ↓                ↓
            (metrics)      (traces)         (logs)
                ↓              ↓                ↓
           Prometheus       Jaeger      Loki (future)
                ↓              ↓                ↓
                    ← Grafana Dashboards →
```

### Key Design Decisions:

1. **OTEL Collector as Central Hub**: All telemetry flows through the collector for processing, filtering, and routing
2. **Pull + Push Hybrid**: Prometheus pulls metrics, while traces/logs are pushed
3. **Sampling Strategy**: Head-based sampling in development, tail-based in production
4. **Correlation**: Request ID, Trace ID, and Span ID propagation across all telemetry
5. **Local Development**: Full stack runs in Docker Compose with minimal resources

## Consequences

### Positive Consequences

- **Vendor Independence**: Can switch backends without changing instrumentation
- **Industry Standard**: OTEL is becoming the de facto standard
- **Comprehensive Coverage**: All three pillars from one SDK
- **Cost Optimization**: Sampling and filtering reduce storage costs
- **Developer Experience**: Local stack mirrors production
- **Performance**: Minimal overhead with batching and async collection
- **Future-Proof**: Supports emerging standards and protocols

### Negative Consequences

- **Complexity**: Multiple components to manage and upgrade
- **Resource Usage**: Collector adds another hop and resource consumption
- **Learning Curve**: Teams need to understand OTEL concepts
- **Configuration Overhead**: Extensive configuration options
- **Data Volume**: Can generate massive amounts of telemetry

### Mitigations

- Pre-configured collector pipelines for common use cases
- Aggressive sampling in non-production environments
- Automated dashboard provisioning
- Comprehensive documentation and runbooks
- Resource limits and quotas on all components
- PII scrubbing processors in collector pipeline

## Alternatives Considered

### Option 1: Datadog APM (Full SaaS)

**Pros:**
- Fully managed, no infrastructure
- Excellent UI/UX
- Advanced ML-based anomaly detection
- Integrated logs, metrics, traces

**Cons:**
- Expensive at scale ($31/host/month + data costs)
- Vendor lock-in with proprietary agents
- Data sovereignty concerns
- Limited customization

**Reason not chosen:** Cost prohibitive for our scale and vendor lock-in concerns.

### Option 2: Elastic Stack (ELK)

**Pros:**
- Mature ecosystem
- Powerful search capabilities
- Single stack for logs, metrics, APM
- Self-hosted option available

**Cons:**
- Resource intensive (especially Elasticsearch)
- Complex cluster management
- Licensing concerns (not fully open source)
- Weaker tracing capabilities

**Reason not chosen:** Resource requirements too high and tracing support not as mature as Jaeger.

### Option 3: New Relic One

**Pros:**
- Comprehensive platform
- Good developer experience
- Strong APM capabilities
- Programmable platform

**Cons:**
- Expensive ($99/user/month minimum)
- Proprietary instrumentation
- Cloud-only (no self-hosted)
- Complex pricing model

**Reason not chosen:** Cost and lack of self-hosted option.

### Option 4: Native Cloud Provider (AWS CloudWatch/X-Ray)

**Pros:**
- Deep AWS integration
- Pay-per-use pricing
- No infrastructure management
- Native service mesh support

**Cons:**
- AWS lock-in
- Limited cross-cloud support
- Weaker visualization (CloudWatch)
- Higher latency for metrics

**Reason not chosen:** Cloud vendor lock-in and limited functionality compared to Grafana.

### Option 5: Custom Stack (StatsD + ELK + Zipkin)

**Pros:**
- Full control
- Can optimize for specific needs
- Mix and match best tools

**Cons:**
- High maintenance burden
- No unified instrumentation
- Integration complexity
- Lack of industry standard

**Reason not chosen:** Maintenance overhead and lack of unified standards.

## Implementation Guidelines

### Instrumentation Standards

1. **Metrics Naming**: Follow Prometheus conventions
   - `service_component_unit_suffix`
   - Example: `http_requests_duration_seconds`

2. **Logging Levels**:
   - ERROR: Actionable errors requiring intervention
   - WARN: Degraded behavior but recoverable
   - INFO: Important business events
   - DEBUG: Detailed diagnostic information

3. **Tracing Sampling**:
   - Development: 100% sampling
   - Staging: 10% sampling
   - Production: 1% baseline, 100% for errors

4. **Resource Attributes**:
   - service.name
   - service.version
   - deployment.environment
   - k8s.pod.name
   - cloud.region

### Privacy and Compliance

- No PII in metrics labels or span attributes
- Structured logging with redaction rules
- Data retention: 15 days traces, 30 days metrics, 7 days logs
- GDPR-compliant data handling with right to deletion

### Performance Targets

- Instrumentation overhead: <2% CPU, <20MB memory
- Collector processing: <100ms p99 latency
- Metric scrape interval: 15 seconds
- Trace sampling decision: <1ms

## Migration Path

1. **Phase 1**: Instrument with OTEL SDK (metrics + traces)
2. **Phase 2**: Migrate logging to structured format
3. **Phase 3**: Deploy collectors and backends
4. **Phase 4**: Create dashboards and alerts
5. **Phase 5**: Implement SLOs and error budgets

## Related

- ADR 001: Go Web Framework
- ADR 002: RPC Framework
- ADR 005: Logging Strategy
- ADR 006: SLI/SLO Definition
- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/)
- [Jaeger Architecture](https://www.jaegertracing.io/docs/architecture/)
- [Grafana Dashboard Guide](https://grafana.com/docs/grafana/latest/dashboards/)
