// Package metrics provides Prometheus metrics collection
package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all application metrics
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal     *prometheus.CounterVec
	HTTPRequestDuration   *prometheus.HistogramVec
	HTTPRequestsInFlight  prometheus.Gauge
	HTTPResponseSizeBytes *prometheus.HistogramVec

	// gRPC metrics
	GRPCRequestsTotal   *prometheus.CounterVec
	GRPCRequestDuration *prometheus.HistogramVec
	GRPCStreamMsgs      *prometheus.CounterVec

	// Business metrics
	UserRegistrations  prometheus.Counter
	ActiveUsers        prometheus.Gauge
	BusinessOperations *prometheus.CounterVec
	OperationDuration  *prometheus.HistogramVec

	// Database metrics
	DBConnectionsOpen  prometheus.Gauge
	DBConnectionsInUse prometheus.Gauge
	DBQueryDuration    *prometheus.HistogramVec
	DBQueryTotal       *prometheus.CounterVec

	// Cache metrics
	CacheHits      *prometheus.CounterVec
	CacheMisses    *prometheus.CounterVec
	CacheEvictions prometheus.Counter

	// Custom application metrics
	AppInfo           *prometheus.GaugeVec
	ProcessUptime     prometheus.Counter
	HealthCheckStatus prometheus.Gauge
}

var (
	// Global metrics instance
	globalMetrics *Metrics

	// Default buckets for latency histograms (in seconds)
	defaultLatencyBuckets = []float64{
		0.001, // 1ms
		0.005, // 5ms
		0.01,  // 10ms
		0.025, // 25ms
		0.05,  // 50ms
		0.1,   // 100ms
		0.25,  // 250ms
		0.5,   // 500ms
		1.0,   // 1s
		2.5,   // 2.5s
		5.0,   // 5s
		10.0,  // 10s
	}

	// Default buckets for size histograms (in bytes)
	defaultSizeBuckets = []float64{
		100,
		1000,
		10000,
		100000,
		1000000,
		10000000,
		100000000,
	}
)

// Initialize creates and registers all metrics
func Initialize(namespace, subsystem string) *Metrics {
	if globalMetrics != nil {
		return globalMetrics
	}

	metrics := &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request latency in seconds",
				Buckets:   defaultLatencyBuckets,
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "http_requests_in_flight",
				Help:      "Number of HTTP requests currently being processed",
			},
		),
		HTTPResponseSizeBytes: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "http_response_size_bytes",
				Help:      "HTTP response size in bytes",
				Buckets:   defaultSizeBuckets,
			},
			[]string{"method", "path", "status"},
		),

		// gRPC metrics
		GRPCRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "grpc_requests_total",
				Help:      "Total number of gRPC requests",
			},
			[]string{"service", "method", "status"},
		),
		GRPCRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "grpc_request_duration_seconds",
				Help:      "gRPC request latency in seconds",
				Buckets:   defaultLatencyBuckets,
			},
			[]string{"service", "method", "status"},
		),
		GRPCStreamMsgs: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "grpc_stream_msgs_total",
				Help:      "Total number of gRPC stream messages",
			},
			[]string{"service", "method", "direction"},
		),

		// Business metrics
		UserRegistrations: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "user_registrations_total",
				Help:      "Total number of user registrations",
			},
		),
		ActiveUsers: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "active_users",
				Help:      "Number of active users",
			},
		),
		BusinessOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "business_operations_total",
				Help:      "Total number of business operations",
			},
			[]string{"operation", "status"},
		),
		OperationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "operation_duration_seconds",
				Help:      "Business operation duration in seconds",
				Buckets:   defaultLatencyBuckets,
			},
			[]string{"operation"},
		),

		// Database metrics
		DBConnectionsOpen: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "db_connections_open",
				Help:      "Number of open database connections",
			},
		),
		DBConnectionsInUse: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "db_connections_in_use",
				Help:      "Number of database connections in use",
			},
		),
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "db_query_duration_seconds",
				Help:      "Database query duration in seconds",
				Buckets:   defaultLatencyBuckets,
			},
			[]string{"query_type", "table"},
		),
		DBQueryTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "db_queries_total",
				Help:      "Total number of database queries",
			},
			[]string{"query_type", "table", "status"},
		),

		// Cache metrics
		CacheHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "cache_hits_total",
				Help:      "Total number of cache hits",
			},
			[]string{"cache_name"},
		),
		CacheMisses: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "cache_misses_total",
				Help:      "Total number of cache misses",
			},
			[]string{"cache_name"},
		),
		CacheEvictions: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "cache_evictions_total",
				Help:      "Total number of cache evictions",
			},
		),

		// Application info
		AppInfo: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "app_info",
				Help:      "Application information",
			},
			[]string{"version", "commit", "build_time", "go_version"},
		),
		ProcessUptime: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "process_uptime_seconds_total",
				Help:      "Process uptime in seconds",
			},
		),
		HealthCheckStatus: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "health_check_status",
				Help:      "Health check status (1 = healthy, 0 = unhealthy)",
			},
		),
	}

	// Set app info
	metrics.AppInfo.WithLabelValues(
		"unknown", // version - should be set from build
		"unknown", // commit - should be set from build
		"unknown", // build_time - should be set from build
		"unknown", // go_version - should be set from build
	).Set(1)

	// Initialize health check as healthy
	metrics.HealthCheckStatus.Set(1)

	globalMetrics = metrics
	return metrics
}

// Get returns the global metrics instance
func Get() *Metrics {
	if globalMetrics == nil {
		Initialize("lumi", "go")
	}
	return globalMetrics
}

// Handler returns the Prometheus HTTP handler
func Handler() http.Handler {
	return promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	)
}

// RecordHTTPRequest records an HTTP request metric
func RecordHTTPRequest(method, path, status string, duration time.Duration, size int) {
	m := Get()
	m.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, path, status).Observe(duration.Seconds())
	m.HTTPResponseSizeBytes.WithLabelValues(method, path, status).Observe(float64(size))
}

// RecordGRPCRequest records a gRPC request metric
func RecordGRPCRequest(service, method, status string, duration time.Duration) {
	m := Get()
	m.GRPCRequestsTotal.WithLabelValues(service, method, status).Inc()
	m.GRPCRequestDuration.WithLabelValues(service, method, status).Observe(duration.Seconds())
}

// RecordBusinessOperation records a business operation metric
func RecordBusinessOperation(operation, status string, duration time.Duration) {
	m := Get()
	m.BusinessOperations.WithLabelValues(operation, status).Inc()
	m.OperationDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// RecordDBQuery records a database query metric
func RecordDBQuery(queryType, table, status string, duration time.Duration) {
	m := Get()
	m.DBQueryTotal.WithLabelValues(queryType, table, status).Inc()
	m.DBQueryDuration.WithLabelValues(queryType, table).Observe(duration.Seconds())
}

// RecordCacheHit records a cache hit
func RecordCacheHit(cacheName string) {
	Get().CacheHits.WithLabelValues(cacheName).Inc()
}

// RecordCacheMiss records a cache miss
func RecordCacheMiss(cacheName string) {
	Get().CacheMisses.WithLabelValues(cacheName).Inc()
}

// IncrementUserRegistrations increments the user registration counter
func IncrementUserRegistrations() {
	Get().UserRegistrations.Inc()
}

// SetActiveUsers sets the number of active users
func SetActiveUsers(count float64) {
	Get().ActiveUsers.Set(count)
}

// SetHealthStatus sets the health check status
func SetHealthStatus(healthy bool) {
	if healthy {
		Get().HealthCheckStatus.Set(1)
	} else {
		Get().HealthCheckStatus.Set(0)
	}
}

// UpdateDBConnectionMetrics updates database connection metrics
func UpdateDBConnectionMetrics(open, inUse int) {
	m := Get()
	m.DBConnectionsOpen.Set(float64(open))
	m.DBConnectionsInUse.Set(float64(inUse))
}

// StartUptimeCounter starts the uptime counter goroutine
func StartUptimeCounter(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				Get().ProcessUptime.Inc()
			}
		}
	}()
}
