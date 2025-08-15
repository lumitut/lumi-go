// Package middleware provides HTTP middleware components
package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lumitut/lumi-go/internal/observability/metrics"
)

// Metrics middleware records HTTP metrics
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Track in-flight requests
		metrics.Get().HTTPRequestsInFlight.Inc()
		defer metrics.Get().HTTPRequestsInFlight.Dec()

		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get request details
		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = "not_found"
		}
		status := strconv.Itoa(c.Writer.Status())
		size := c.Writer.Size()

		// Record metrics
		metrics.RecordHTTPRequest(method, path, status, duration, size)
	}
}

// MetricsWithConfig allows custom configuration for metrics middleware
type MetricsConfig struct {
	// SkipPaths specifies paths to skip from metrics
	SkipPaths []string
	// GroupedPaths groups similar paths together (e.g., /users/:id -> /users/{id})
	GroupedPaths map[string]string
	// IncludeQueryParams includes query parameters in path label
	IncludeQueryParams bool
}

// MetricsWithConfig creates a metrics middleware with custom configuration
func MetricsWithConfig(config MetricsConfig) gin.HandlerFunc {
	skipMap := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipMap[path] = true
	}

	return func(c *gin.Context) {
		// Skip if path is in skip list
		if skipMap[c.Request.URL.Path] {
			c.Next()
			return
		}

		// Track in-flight requests
		metrics.Get().HTTPRequestsInFlight.Inc()
		defer metrics.Get().HTTPRequestsInFlight.Dec()

		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get request details
		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = "not_found"
		}

		// Apply path grouping if configured
		if grouped, ok := config.GroupedPaths[path]; ok {
			path = grouped
		}

		// Include query params if configured
		if config.IncludeQueryParams && c.Request.URL.RawQuery != "" {
			path = path + "?" + c.Request.URL.RawQuery
		}

		status := strconv.Itoa(c.Writer.Status())
		size := c.Writer.Size()

		// Record metrics
		metrics.RecordHTTPRequest(method, path, status, duration, size)
	}
}
