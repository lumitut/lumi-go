// Package app provides application-level initialization and coordination
package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lumitut/lumi-go/internal/config"
	"github.com/lumitut/lumi-go/internal/httpapi"
	"github.com/lumitut/lumi-go/internal/observability/logger"
	"go.uber.org/zap"
)

// Application represents the main application
type Application struct {
	cfg        *config.Config
	httpServer *httpapi.Server
	// Add other servers/services as needed
	// rpcServer  *rpcapi.Server
	// workers    []Worker
}

// New creates a new application instance
func New(cfg *config.Config) (*Application, error) {
	app := &Application{
		cfg: cfg,
	}

	// Initialize HTTP server
	app.httpServer = httpapi.NewServer(cfg)

	// Initialize other components as needed
	// app.rpcServer = rpcapi.NewServer(cfg)

	return app, nil
}

// Run starts the application and blocks until shutdown
func (a *Application) Run(ctx context.Context) error {
	// Start HTTP server
	httpErrCh := make(chan error, 1)
	go func() {
		logger.Info(ctx, "Starting HTTP server",
			zap.String("port", a.cfg.Server.HTTPPort))
		if err := a.httpServer.Start(ctx); err != nil && err != http.ErrServerClosed {
			httpErrCh <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	// Start other servers as needed
	// rpcErrCh := make(chan error, 1)
	// go func() {
	//     if err := a.rpcServer.Start(ctx); err != nil {
	//         rpcErrCh <- fmt.Errorf("RPC server error: %w", err)
	//     }
	// }()

	// Wait for interrupt signal or error
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		logger.Info(ctx, "Received shutdown signal", zap.String("signal", sig.String()))
		return a.Shutdown(ctx)
	case err := <-httpErrCh:
		return err
	case <-ctx.Done():
		return a.Shutdown(ctx)
	}
}

// Shutdown gracefully shuts down the application
func (a *Application) Shutdown(ctx context.Context) error {
	logger.Info(ctx, "Shutting down application")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, a.cfg.Server.GracefulShutdownTimeout)
	defer cancel()

	// Shutdown HTTP server
	if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error(ctx, "Failed to shutdown HTTP server", err)
	}

	// Shutdown other components
	// if err := a.rpcServer.Shutdown(shutdownCtx); err != nil {
	//     logger.Error(ctx, "Failed to shutdown RPC server", err)
	// }

	// Wait for graceful shutdown or timeout
	select {
	case <-shutdownCtx.Done():
		if shutdownCtx.Err() == context.DeadlineExceeded {
			logger.Warn(ctx, "Shutdown timeout exceeded, forcing shutdown")
		}
	default:
		logger.Info(ctx, "Graceful shutdown completed")
	}

	return nil
}

// Health returns the health status of the application
func (a *Application) Health() map[string]interface{} {
	return map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"service":   a.cfg.Service.Name,
		"version":   a.cfg.Service.Version,
	}
}

// Ready returns the readiness status of the application
func (a *Application) Ready() map[string]interface{} {
	ready := true
	checks := make(map[string]bool)

	// Check HTTP server
	if a.httpServer != nil && a.httpServer.IsReady() {
		checks["http"] = true
	} else {
		checks["http"] = false
		ready = false
	}

	// Add other readiness checks
	// checks["database"] = a.checkDatabase()
	// checks["redis"] = a.checkRedis()

	return map[string]interface{}{
		"ready":   ready,
		"checks":  checks,
		"time":    time.Now().Unix(),
		"service": a.cfg.Service.Name,
	}
}
