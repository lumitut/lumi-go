// Package middleware provides HTTP middleware components
package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// TracingConfig provides configuration for OpenTelemetry tracing middleware
type TracingConfig struct {
	// ServiceName is the name of the service
	ServiceName string
	// SpanNameFormatter formats the span name
	SpanNameFormatter func(*gin.Context) string
	// TracerProvider allows custom tracer provider
	TracerProvider trace.TracerProvider
	// Propagator allows custom propagator
	Propagator propagation.TextMapPropagator
	// SkipPaths skips tracing for these paths
	SkipPaths []string
	// RecordError records errors in spans
	RecordError bool
	// RecordRequestBody includes request body in span attributes
	RecordRequestBody bool
	// RecordResponseBody includes response body in span attributes
	RecordResponseBody bool
	// RecordHeaders includes headers in span attributes
	RecordHeaders bool
}

// DefaultTracingConfig returns default tracing configuration
func DefaultTracingConfig() TracingConfig {
	return TracingConfig{
		ServiceName: "lumi-go",
		SpanNameFormatter: func(c *gin.Context) string {
			return fmt.Sprintf("%s %s", c.Request.Method, c.FullPath())
		},
		TracerProvider:     otel.GetTracerProvider(),
		Propagator:         otel.GetTextMapPropagator(),
		SkipPaths:          []string{"/health", "/ready", "/metrics"},
		RecordError:        true,
		RecordRequestBody:  false,
		RecordResponseBody: false,
		RecordHeaders:      false,
	}
}

// Tracing creates OpenTelemetry tracing middleware
func Tracing() gin.HandlerFunc {
	return TracingWithConfig(DefaultTracingConfig())
}

// TracingWithConfig creates OpenTelemetry tracing middleware with custom configuration
func TracingWithConfig(config TracingConfig) gin.HandlerFunc {
	if config.ServiceName == "" {
		config.ServiceName = "lumi-go"
	}
	if config.SpanNameFormatter == nil {
		config.SpanNameFormatter = func(c *gin.Context) string {
			return fmt.Sprintf("%s %s", c.Request.Method, c.FullPath())
		}
	}
	if config.TracerProvider == nil {
		config.TracerProvider = otel.GetTracerProvider()
	}
	if config.Propagator == nil {
		config.Propagator = otel.GetTextMapPropagator()
	}

	tracer := config.TracerProvider.Tracer(
		config.ServiceName,
		trace.WithInstrumentationVersion("1.0.0"),
	)

	// Build skip map
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

		// Extract trace context from incoming request
		ctx := config.Propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// Start span
		spanName := config.SpanNameFormatter(c)
		if spanName == "" {
			spanName = fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)
		}

		opts := []trace.SpanStartOption{
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPMethod(c.Request.Method),
				semconv.HTTPTarget(c.Request.URL.String()),
				semconv.HTTPRoute(c.FullPath()),
				semconv.HTTPScheme(c.Request.URL.Scheme),
				semconv.NetHostName(c.Request.Host),
				attribute.String("http.user_agent", c.Request.UserAgent()),
				attribute.String("net.peer.ip", c.ClientIP()),
			),
		}

		ctx, span := tracer.Start(ctx, spanName, opts...)
		defer span.End()

		// Add request attributes
		span.SetAttributes(
			attribute.String("http.request_id", ExtractRequestID(c)),
			attribute.String("http.correlation_id", ExtractCorrelationID(c)),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.path", c.Request.URL.Path),
			attribute.String("http.query", c.Request.URL.RawQuery),
			attribute.Int64("http.request_content_length", c.Request.ContentLength),
		)

		// Add headers if configured
		if config.RecordHeaders {
			for key, values := range c.Request.Header {
				if len(values) > 0 {
					span.SetAttributes(attribute.StringSlice(fmt.Sprintf("http.request.header.%s", key), values))
				}
			}
		}

		// Store trace and span IDs in gin context
		if spanCtx := span.SpanContext(); spanCtx.IsValid() {
			c.Set("trace_id", spanCtx.TraceID().String())
			c.Set("span_id", spanCtx.SpanID().String())
		}

		// Update request context
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()

		// Set response attributes
		status := c.Writer.Status()
		span.SetAttributes(
			semconv.HTTPStatusCode(status),
			attribute.Int("http.response_size", c.Writer.Size()),
		)

		// Set span status based on HTTP status
		if status >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", status))

			// Record error if configured
			if config.RecordError && len(c.Errors) > 0 {
				for _, err := range c.Errors {
					span.RecordError(err.Err)
				}
			}
		} else {
			span.SetStatus(codes.Ok, "")
		}

		// Inject trace context into response headers
		config.Propagator.Inject(ctx, propagation.HeaderCarrier(c.Writer.Header()))
	}
}

// responseBodyWriter captures response body for tracing
type responseBodyWriter struct {
	gin.ResponseWriter
	body *[]byte
}

func (w *responseBodyWriter) Write(b []byte) (int, error) {
	*w.body = append(*w.body, b...)
	return w.ResponseWriter.Write(b)
}

// ExtractTraceContext extracts trace context from gin context
func ExtractTraceContext(c *gin.Context) trace.SpanContext {
	if span := trace.SpanFromContext(c.Request.Context()); span != nil {
		return span.SpanContext()
	}
	return trace.SpanContext{}
}

// InjectTraceContext injects trace context into outgoing request
func InjectTraceContext(c *gin.Context, req *http.Request) {
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(c.Request.Context(), propagation.HeaderCarrier(req.Header))
}

// StartSpan starts a new span for a gin context
func StartSpan(c *gin.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	tracer := otel.Tracer("lumi-go")
	return tracer.Start(c.Request.Context(), name, opts...)
}

// SpanFromContext returns the current span from gin context
func SpanFromContext(c *gin.Context) trace.Span {
	return trace.SpanFromContext(c.Request.Context())
}

// TracingResponseWriter wraps gin.ResponseWriter to capture status for tracing
type TracingResponseWriter struct {
	gin.ResponseWriter
	status int
	size   int
}

func (w *TracingResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *TracingResponseWriter) Write(data []byte) (int, error) {
	size, err := w.ResponseWriter.Write(data)
	w.size += size
	return size, err
}

func (w *TracingResponseWriter) Status() int {
	return w.status
}

func (w *TracingResponseWriter) Size() int {
	return w.size
}

// W3CTracePropagation creates tracing middleware with W3C Trace Context propagation
func W3CTracePropagation() gin.HandlerFunc {
	config := DefaultTracingConfig()
	config.Propagator = propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	return TracingWithConfig(config)
}

// AddSpanAttribute adds an attribute to the current span
func AddSpanAttribute(c *gin.Context, key string, value interface{}) {
	span := trace.SpanFromContext(c.Request.Context())
	if span == nil {
		return
	}

	switch v := value.(type) {
	case string:
		span.SetAttributes(attribute.String(key, v))
	case int:
		span.SetAttributes(attribute.Int(key, v))
	case int64:
		span.SetAttributes(attribute.Int64(key, v))
	case float64:
		span.SetAttributes(attribute.Float64(key, v))
	case bool:
		span.SetAttributes(attribute.Bool(key, v))
	case []string:
		span.SetAttributes(attribute.StringSlice(key, v))
	default:
		span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", v)))
	}
}

// AddSpanEvent adds an event to the current span
func AddSpanEvent(c *gin.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(c.Request.Context())
	if span != nil {
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

// RecordSpanError records an error in the current span
func RecordSpanError(c *gin.Context, err error) {
	span := trace.SpanFromContext(c.Request.Context())
	if span != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}
