# Metrics Guide

## Overview

The lumi-go template provides comprehensive metrics collection using Prometheus. All metrics are exposed on the `/metrics` endpoint and collected by Prometheus for visualization in Grafana.

## Available Metrics

### HTTP Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `lumi_go_api_http_requests_total` | Counter | method, path, status | Total HTTP requests |
| `lumi_go_api_http_request_duration_seconds` | Histogram | method, path, status | Request latency |
| `lumi_go_api_http_requests_in_flight` | Gauge | - | Currently active requests |
| `lumi_go_api_http_response_size_bytes` | Histogram | method, path, status | Response size |

### gRPC Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `lumi_go_api_grpc_requests_total` | Counter | service, method, status | Total gRPC requests |
| `lumi_go_api_grpc_request_duration_seconds` | Histogram | service, method, status | gRPC latency |
| `lumi_go_api_grpc_stream_msgs_total` | Counter | service, method, direction | Stream messages |

### Business Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `lumi_go_api_user_registrations_total` | Counter | - | User registrations |
| `lumi_go_api_active_users` | Gauge | - | Active user count |
| `lumi_go_api_business_operations_total` | Counter | operation, status | Business operations |
| `lumi_go_api_operation_duration_seconds` | Histogram | operation | Operation duration |

### Database Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `lumi_go_api_db_connections_open` | Gauge | - | Open connections |
| `lumi_go_api_db_connections_in_use` | Gauge | - | Active connections |
| `lumi_go_api_db_query_duration_seconds` | Histogram | query_type, table | Query latency |
| `lumi_go_api_db_queries_total` | Counter | query_type, table, status | Total queries |

### Cache Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `lumi_go_api_cache_hits_total` | Counter | cache_name | Cache hits |
| `lumi_go_api_cache_misses_total` | Counter | cache_name | Cache misses |
| `lumi_go_api_cache_evictions_total` | Counter | - | Cache evictions |

### Application Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `lumi_go_api_app_info` | Gauge | version, commit, build_time, go_version | App metadata |
| `lumi_go_api_process_uptime_seconds_total` | Counter | - | Process uptime |
| `lumi_go_api_health_check_status` | Gauge | - | Health status (1=healthy, 0=unhealthy) |

## Usage

### Recording Metrics

```go
import "github.com/lumitut/lumi-go/internal/observability/metrics"

// Record HTTP request
metrics.RecordHTTPRequest(method, path, status, duration, size)

// Record business operation
start := time.Now()
// ... perform operation ...
metrics.RecordBusinessOperation("create_order", "success", time.Since(start))

// Update gauge
metrics.SetActiveUsers(125)

// Increment counter
metrics.IncrementUserRegistrations()

// Record cache hit/miss
if cached {
    metrics.RecordCacheHit("users")
} else {
    metrics.RecordCacheMiss("users")
}
```

### Custom Metrics

```go
// Get metrics instance
m := metrics.Get()

// Use Prometheus metrics directly
m.BusinessOperations.WithLabelValues("custom_op", "success").Inc()
m.ActiveUsers.Set(float64(userCount))
```

## Viewing Metrics

### Local Development

1. **Raw metrics endpoint:**
   ```bash
   curl http://localhost:8080/metrics
   ```

2. **Prometheus UI:**
   - URL: http://localhost:9090
   - Query examples:
     ```promql
     # Request rate
     rate(lumi_go_api_http_requests_total[5m])
     
     # Error rate
     rate(lumi_go_api_http_requests_total{status=~"5.."}[5m])
     
     # P95 latency
     histogram_quantile(0.95, rate(lumi_go_api_http_request_duration_seconds_bucket[5m]))
     ```

3. **Grafana Dashboard:**
   - URL: http://localhost:3000
   - Login: admin/admin
   - Dashboard: lumi-go Application Dashboard

## Grafana Dashboard

The template includes a pre-configured Grafana dashboard with:

### Key Panels
- **Request Rate (RPS)** - Requests per second by status
- **Error Rate** - Percentage of 5xx responses
- **Request Latency** - P50, P95, P99 latencies
- **Active Requests** - Currently processing requests
- **Health Status** - Application health indicator
- **Process Uptime** - Time since startup

### Detailed Views
- Latency by endpoint
- Request rate by HTTP method
- Database connection pool status
- Cache hit rate
- Business operations tracking

## Best Practices

### 1. Label Cardinality

❌ **Bad:** High cardinality labels
```go
// Don't use user IDs as labels
metrics.RecordHTTPRequest(method, path, status, userID, ...)
```

✅ **Good:** Bounded label values
```go
// Use fixed set of labels
metrics.RecordHTTPRequest(method, path, status, ...)
```

### 2. Histogram Buckets

Choose appropriate buckets for your use case:

```go
// Fast API endpoints (ms)
fastBuckets := []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1}

// Slow operations (seconds)
slowBuckets := []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0}
```

### 3. Metric Naming

Follow Prometheus naming conventions:
- Use `_total` suffix for counters
- Use `_seconds` suffix for durations
- Use `_bytes` suffix for sizes
- Use lowercase with underscores

### 4. Resource Tracking

```go
// Track database connections
func updateDBMetrics(db *sql.DB) {
    stats := db.Stats()
    metrics.UpdateDBConnectionMetrics(stats.OpenConnections, stats.InUse)
}
```

## Alerting Rules

Example Prometheus alerting rules:

```yaml
groups:
  - name: lumi-go
    rules:
      - alert: HighErrorRate
        expr: rate(lumi_go_api_http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: High error rate detected
          
      - alert: HighLatency
        expr: histogram_quantile(0.95, rate(lumi_go_api_http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: High P95 latency
          
      - alert: ServiceDown
        expr: up{job="lumi-go"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: Service is down
```

## Performance Considerations

### Metric Collection Overhead

- **CPU**: <1% for typical workloads
- **Memory**: ~10MB for default cardinality
- **Network**: ~100KB/min to Prometheus

### Optimization Tips

1. **Use middleware selectively:**
   ```go
   // Skip metrics for health checks
   router.Use(middleware.MetricsWithConfig(middleware.MetricsConfig{
       SkipPaths: []string{"/health", "/ready"},
   }))
   ```

2. **Group similar paths:**
   ```go
   // Avoid cardinality explosion
   GroupedPaths: map[string]string{
       "/users/:id": "/users/{id}",
       "/orders/:id": "/orders/{id}",
   }
   ```

3. **Sample high-volume metrics:**
   ```go
   // Sample 10% of requests
   if rand.Float64() < 0.1 {
       metrics.RecordHTTPRequest(...)
   }
   ```

## Troubleshooting

### Missing Metrics

1. Check service is running: `curl http://localhost:8080/health`
2. Verify metrics endpoint: `curl http://localhost:8080/metrics`
3. Check Prometheus targets: http://localhost:9090/targets
4. Verify scrape configuration in `prometheus.yml`

### High Memory Usage

1. Check label cardinality:
   ```promql
   count by (__name__)({__name__=~"lumi_go_api.*"})
   ```

2. Review unique label combinations:
   ```promql
   count(count by (path) (lumi_go_api_http_requests_total))
   ```

### Grafana Dashboard Issues

1. Check datasource configuration
2. Verify Prometheus is accessible
3. Check metric names match dashboard queries
4. Review time range settings

## References

- [Prometheus Best Practices](https://prometheus.io/docs/practices/)
- [Grafana Documentation](https://grafana.com/docs/)
- [OpenTelemetry Metrics](https://opentelemetry.io/docs/concepts/signals/metrics/)
- [RED Method](https://www.weave.works/blog/the-red-method-key-metrics-for-microservices-architecture/)
- [USE Method](https://www.brendangregg.com/usemethod.html)
