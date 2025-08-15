// Package tracing provides OpenTelemetry tracing setup
package tracing

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Config holds tracing configuration
type Config struct {
	// ServiceName is the name of the service
	ServiceName string
	// ServiceVersion is the version of the service
	ServiceVersion string
	// Environment is the deployment environment (dev, staging, prod)
	Environment string
	// ExporterEndpoint is the OTLP collector endpoint
	ExporterEndpoint string
	// ExporterProtocol is the protocol to use (grpc or http)
	ExporterProtocol string
	// Insecure disables TLS for the exporter
	Insecure bool
	// SampleRate is the sampling rate (0.0 to 1.0)
	SampleRate float64
	// Enabled enables or disables tracing
	Enabled bool
}

// DefaultConfig returns default tracing configuration
func DefaultConfig() Config {
	return Config{
		ServiceName:      getEnv("SERVICE_NAME", "lumi-go"),
		ServiceVersion:   getEnv("SERVICE_VERSION", "unknown"),
		Environment:      getEnv("ENVIRONMENT", "development"),
		ExporterEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		ExporterProtocol: getEnv("OTEL_EXPORTER_PROTOCOL", "grpc"),
		Insecure:         getEnv("OTEL_EXPORTER_INSECURE", "true") == "true",
		SampleRate:       getEnvFloat("OTEL_SAMPLE_RATE", 1.0),
		Enabled:          getEnv("OTEL_ENABLED", "true") == "true",
	}
}

var (
	// Global tracer provider
	globalTracerProvider *sdktrace.TracerProvider
)

// Initialize sets up OpenTelemetry tracing
func Initialize(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	if !cfg.Enabled {
		return func(context.Context) error { return nil }, nil
	}

	// Create resource
	res, err := createResource(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create exporter
	exporter, err := createExporter(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	// Create sampler
	sampler := sdktrace.TraceIDRatioBased(cfg.SampleRate)

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)
	globalTracerProvider = tp

	// Set global propagator
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	// Return shutdown function
	return func(ctx context.Context) error {
		return tp.Shutdown(ctx)
	}, nil
}

// createResource creates the OpenTelemetry resource
func createResource(cfg Config) (*resource.Resource, error) {
	// Get additional attributes from environment
	instanceID := getEnv("INSTANCE_ID", "unknown")
	hostname, _ := os.Hostname()

	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
			attribute.String("service.instance.id", instanceID),
			attribute.String("host.name", hostname),
			attribute.String("telemetry.sdk.name", "opentelemetry"),
			attribute.String("telemetry.sdk.language", "go"),
			attribute.String("telemetry.sdk.version", otel.Version()),
		),
	)
}

// createExporter creates the OTLP exporter based on protocol
func createExporter(ctx context.Context, cfg Config) (*otlptrace.Exporter, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	switch cfg.ExporterProtocol {
	case "grpc":
		return createGRPCExporter(ctx, cfg)
	case "http":
		return createHTTPExporter(ctx, cfg)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", cfg.ExporterProtocol)
	}
}

// createGRPCExporter creates a gRPC OTLP exporter
func createGRPCExporter(ctx context.Context, cfg Config) (*otlptrace.Exporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.ExporterEndpoint),
	}

	if cfg.Insecure {
		opts = append(opts, otlptracegrpc.WithTLSCredentials(insecure.NewCredentials()))
	}

	// Add retry configuration
	opts = append(opts,
		otlptracegrpc.WithRetry(otlptracegrpc.RetryConfig{
			Enabled:         true,
			InitialInterval: 1 * time.Second,
			MaxInterval:     30 * time.Second,
			MaxElapsedTime:  1 * time.Minute,
		}),
	)

	// Add compression
	opts = append(opts,
		otlptracegrpc.WithCompressor("gzip"),
	)

	// Create connection with additional options
	opts = append(opts,
		otlptracegrpc.WithDialOption(
			grpc.WithDefaultCallOptions(
				grpc.MaxCallSendMsgSize(10*1024*1024), // 10MB
				grpc.MaxCallRecvMsgSize(10*1024*1024), // 10MB
			),
		),
	)

	return otlptracegrpc.New(ctx, opts...)
}

// createHTTPExporter creates an HTTP OTLP exporter
func createHTTPExporter(ctx context.Context, cfg Config) (*otlptrace.Exporter, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(cfg.ExporterEndpoint),
	}

	if cfg.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	// Add retry configuration
	opts = append(opts,
		otlptracehttp.WithRetry(otlptracehttp.RetryConfig{
			Enabled:         true,
			InitialInterval: 1 * time.Second,
			MaxInterval:     30 * time.Second,
			MaxElapsedTime:  1 * time.Minute,
		}),
	)

	// Add compression
	opts = append(opts,
		otlptracehttp.WithCompression(otlptracehttp.GzipCompression),
	)

	return otlptracehttp.New(ctx, opts...)
}

// Tracer returns a tracer for the given name
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// StartSpan starts a new span with the given name
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer("").Start(ctx, name, opts...)
}

// SpanFromContext returns the span from the context
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// AddEvent adds an event to the current span
func AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SetStatus sets the status of the current span
func SetStatus(ctx context.Context, code codes.Code, description string) {
	span := trace.SpanFromContext(ctx)
	span.SetStatus(code, description)
}

// SetAttributes sets attributes on the current span
func SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// RecordError records an error on the current span
func RecordError(ctx context.Context, err error, opts ...trace.EventOption) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err, opts...)
}

// Shutdown gracefully shuts down the tracer provider
func Shutdown(ctx context.Context) error {
	if globalTracerProvider != nil {
		return globalTracerProvider.Shutdown(ctx)
	}
	return nil
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if f, err := parseFloat(value); err == nil {
			return f
		}
	}
	return defaultValue
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}
