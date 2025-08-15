// Package middleware provides HTTP middleware components
package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lumitut/lumi-go/internal/observability/logger"
	"go.uber.org/zap"
)

// bodyLogWriter captures response body for logging
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Logging middleware logs HTTP requests and responses
func Logging(skipPaths ...string) gin.HandlerFunc {
	skipMap := make(map[string]bool)
	for _, path := range skipPaths {
		skipMap[path] = true
	}

	return func(c *gin.Context) {
		// Skip logging for certain paths (e.g., health checks)
		if skipMap[c.Request.URL.Path] {
			c.Next()
			return
		}

		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Capture request body if needed (be careful with large bodies)
		var requestBody []byte
		if c.Request.Body != nil && c.Request.ContentLength > 0 && c.Request.ContentLength < 10*1024 { // Only log bodies < 10KB
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get request/correlation IDs
		requestID := ExtractRequestID(c)
		correlationID := ExtractCorrelationID(c)

		// Build log fields
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("ip", c.ClientIP()),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.Float64("latency_ms", float64(latency.Nanoseconds())/1e6),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("request_id", requestID),
			zap.String("correlation_id", correlationID),
		}

		// Add query string if present
		if raw != "" {
			fields = append(fields, zap.String("query", raw))
		}

		// Add error if present
		if len(c.Errors) > 0 {
			errorMessages := make([]string, len(c.Errors))
			for i, err := range c.Errors {
				errorMessages[i] = err.Error()
			}
			fields = append(fields, zap.Strings("errors", errorMessages))
		}

		// Add request body if captured (redact sensitive data)
		if len(requestBody) > 0 {
			opts := logger.DefaultRedactOptions()
			redactedBody := logger.RedactJSON(string(requestBody), opts)
			fields = append(fields, zap.String("request_body", redactedBody))
		}

		// Add response size
		fields = append(fields, zap.Int("response_size", c.Writer.Size()))

		// Get logger with context
		log := logger.WithContext(c.Request.Context())

		// Log based on status code
		switch {
		case c.Writer.Status() >= 500:
			log.Error("HTTP request failed", fields...)
		case c.Writer.Status() >= 400:
			log.Warn("HTTP request client error", fields...)
		case c.Writer.Status() >= 300:
			log.Info("HTTP request redirected", fields...)
		default:
			log.Info("HTTP request completed", fields...)
		}

		// Log slow requests
		if latency > 1*time.Second {
			log.Warn("Slow HTTP request detected",
				zap.String("method", c.Request.Method),
				zap.String("path", path),
				zap.Duration("latency", latency),
			)
		}

		// Audit log for state-changing operations
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" && c.Request.Method != "OPTIONS" {
			result := "success"
			if c.Writer.Status() >= 400 {
				result = "failure"
			}
			logger.Audit(
				c.Request.Context(),
				c.Request.Method,
				path,
				result,
				zap.Int("status_code", c.Writer.Status()),
				zap.String("client_ip", c.ClientIP()),
			)
		}
	}
}

// LoggingConfig provides configuration for the logging middleware
type LoggingConfig struct {
	SkipPaths      []string
	LogRequestBody bool
	LogResponseBody bool
	MaxBodySize    int64
	SlowThreshold  time.Duration
}

// LoggingWithConfig creates a logging middleware with custom configuration
func LoggingWithConfig(config LoggingConfig) gin.HandlerFunc {
	skipMap := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipMap[path] = true
	}

	if config.MaxBodySize == 0 {
		config.MaxBodySize = 10 * 1024 // 10KB default
	}
	if config.SlowThreshold == 0 {
		config.SlowThreshold = 1 * time.Second
	}

	return func(c *gin.Context) {
		// Skip logging for certain paths
		if skipMap[c.Request.URL.Path] {
			c.Next()
			return
		}

		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Capture request body if configured
		var requestBody []byte
		if config.LogRequestBody && c.Request.Body != nil && 
		   c.Request.ContentLength > 0 && c.Request.ContentLength < config.MaxBodySize {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Capture response body if configured
		var blw *bodyLogWriter
		if config.LogResponseBody {
			blw = &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
			c.Writer = blw
		}

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Build log fields
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("ip", c.ClientIP()),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.Float64("latency_ms", float64(latency.Nanoseconds())/1e6),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("request_id", ExtractRequestID(c)),
			zap.String("correlation_id", ExtractCorrelationID(c)),
		}

		if raw != "" {
			fields = append(fields, zap.String("query", raw))
		}

		// Add request body if captured
		if len(requestBody) > 0 {
			opts := logger.DefaultRedactOptions()
			redactedBody := logger.RedactJSON(string(requestBody), opts)
			fields = append(fields, zap.String("request_body", redactedBody))
		}

		// Add response body if captured
		if blw != nil && blw.body.Len() > 0 && int64(blw.body.Len()) < config.MaxBodySize {
			opts := logger.DefaultRedactOptions()
			redactedBody := logger.RedactJSON(blw.body.String(), opts)
			fields = append(fields, zap.String("response_body", redactedBody))
		}

		// Get logger with context
		log := logger.WithContext(c.Request.Context())

		// Log based on status code
		switch {
		case c.Writer.Status() >= 500:
			log.Error("HTTP request failed", fields...)
		case c.Writer.Status() >= 400:
			log.Warn("HTTP request client error", fields...)
		default:
			log.Info("HTTP request completed", fields...)
		}

		// Log slow requests
		if latency > config.SlowThreshold {
			log.Warn("Slow HTTP request detected",
				zap.String("method", c.Request.Method),
				zap.String("path", path),
				zap.Duration("latency", latency),
				zap.Duration("threshold", config.SlowThreshold),
			)
		}
	}
}
