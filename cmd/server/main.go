package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/lumitut/lumi-go/internal/config"
	"github.com/lumitut/lumi-go/internal/httpapi"
	"github.com/lumitut/lumi-go/internal/observability/logger"
	"github.com/lumitut/lumi-go/internal/observability/metrics"
	"github.com/lumitut/lumi-go/internal/observability/tracing"
	"go.uber.org/zap"
)

func main() {
	// Create root context
	ctx := context.Background()

	// Load configuration
	// This supports:
	// 1. Command-line flags: -config=/path/to/config.json -env=production
	// 2. JSON config file: cmd/server/schema/lumi.json
	// 3. Environment variables: LUMI_SERVICE_NAME, LUMI_DATABASE_HOST, etc.
	// Priority: Env vars > Config file > Defaults
	cfg, err := config.LoadDefault(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logConfig := logger.Config{
		Level:             cfg.Observability.LogLevel,
		Format:            cfg.Observability.LogFormat,
		Development:       cfg.Observability.LogDevelopment,
		DisableCaller:     false,
		DisableStacktrace: false,
		SampleInitial:     100,
		SampleThereafter:  100,
	}

	if err := logger.Initialize(logConfig); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Log configuration (with sensitive values redacted)
	cfg.LogConfig(ctx)

	// Initialize metrics (replace hyphens with underscores for Prometheus compatibility)
	metricsNamespace := strings.ReplaceAll(cfg.Service.Name, "-", "_")
	metrics.Initialize(metricsNamespace, "api")
	metrics.StartUptimeCounter(ctx)

	// Initialize tracing if enabled
	if cfg.IsTracingEnabled() {
		tracingConfig := tracing.Config{
			ServiceName:      cfg.Service.Name,
			ServiceVersion:   cfg.Service.Version,
			Environment:      cfg.Service.Environment,
			Enabled:          true,
			SampleRate:       1.0, // Default sampling rate
			ExporterEndpoint: cfg.Clients.Tracing.Endpoint,
			ExporterProtocol: "grpc", // Default to gRPC
			Insecure:         true,   // Default for local dev
		}

		shutdown, err := tracing.Initialize(ctx, tracingConfig)
		if err != nil {
			logger.Error(ctx, "Failed to initialize tracing", err)
		} else {
			defer func() {
				shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.GracefulShutdownTimeout)
				defer cancel()
				if err := shutdown(shutdownCtx); err != nil {
					logger.Error(context.Background(), "Failed to shutdown tracing", err)
				}
			}()
		}
	}

	// Log startup
	logger.Info(ctx, "Starting lumi-go service",
		zap.String("version", cfg.Service.Version),
		zap.String("environment", cfg.Service.Environment),
		zap.String("service_name", cfg.Service.Name),
		zap.Bool("tracing_enabled", cfg.IsTracingEnabled()),
		zap.Bool("database_enabled", cfg.Clients.Database.Enabled),
		zap.Bool("redis_enabled", cfg.Clients.Redis.Enabled),
		zap.Bool("metrics_enabled", cfg.Observability.MetricsEnabled),
		zap.Bool("cors_enabled", cfg.Middleware.CORSEnabled),
		zap.Bool("rate_limit_enabled", cfg.Middleware.RateLimitEnabled),
	)

	// Check for maintenance mode
	if cfg.Features.MaintenanceMode {
		logger.Warn(ctx, "Service is in maintenance mode")
	}

	// Create HTTP server
	httpServer := httpapi.NewServer(cfg)

	// Start HTTP server in goroutine
	go func() {
		if err := httpServer.Start(ctx); err != nil {
			logger.Fatal(ctx, "Failed to start HTTP server",
				zap.Error(err),
			)
		}
	}()

	// TODO: Start RPC server when implemented
	// rpcServer := rpcapi.NewServer(cfg)
	// go func() {
	//     if err := rpcServer.Start(ctx); err != nil {
	//         logger.Fatal(ctx, "Failed to start RPC server",
	//             zap.Error(err),
	//         )
	//     }
	// }()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	logger.Info(ctx, "Received shutdown signal",
		zap.String("signal", sig.String()),
	)

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, cfg.Server.GracefulShutdownTimeout)
	defer cancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error(ctx, "Failed to gracefully shutdown HTTP server", err)
	}

	// TODO: Shutdown RPC server when implemented
	// if err := rpcServer.Shutdown(shutdownCtx); err != nil {
	//     logger.Error(ctx, "Failed to gracefully shutdown RPC server", err)
	// }

	// TODO: Close database connections when implemented
	// TODO: Close Redis connections when implemented

	logger.Info(ctx, "Service shutdown complete")
}
