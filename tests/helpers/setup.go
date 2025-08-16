package helpers

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lumitut/lumi-go/internal/config"
	"github.com/lumitut/lumi-go/internal/httpapi"
	"github.com/lumitut/lumi-go/internal/observability/logger"
	"github.com/stretchr/testify/require"
)

// SetupTest creates a test configuration and cleanup function
func SetupTest(t *testing.T) (*config.Config, func()) {
	cfg := TestConfig()

	// Initialize logger for tests
	loggerCfg := logger.Config{
		Level:       cfg.Observability.LogLevel,
		Format:      cfg.Observability.LogFormat,
		Development: cfg.Observability.LogDevelopment,
		OutputPaths: []string{cfg.Observability.LogOutput},
	}
	err := logger.Initialize(loggerCfg)
	require.NoError(t, err)

	cleanup := func() {
		// Any cleanup needed
	}

	return cfg, cleanup
}

// SetupTestServer creates a test HTTP server
func SetupTestServer(t *testing.T) (*httptest.Server, *gin.Engine, func()) {
	cfg := TestConfig()

	// Initialize logger
	loggerCfg := logger.Config{
		Level:       cfg.Observability.LogLevel,
		Format:      cfg.Observability.LogFormat,
		Development: cfg.Observability.LogDevelopment,
		OutputPaths: []string{cfg.Observability.LogOutput},
	}
	err := logger.Initialize(loggerCfg)
	require.NoError(t, err)

	// Create server
	server := httpapi.NewServer(cfg)
	require.NotNil(t, server)

	// Get the router
	router := server.Router()

	// Create test server
	ts := httptest.NewServer(router)

	cleanup := func() {
		ts.Close()
	}

	return ts, router, cleanup
}

// TestConfig returns a configuration suitable for testing
func TestConfig() *config.Config {
	return &config.Config{
		Service: config.ServiceConfig{
			Name:        "test-service",
			Version:     "test",
			Environment: "test",
			LogLevel:    "debug",
		},
		Server: config.ServerConfig{
			HTTPPort:                "0", // Random port
			RPCPort:                 "0",
			HTTPReadTimeout:         5 * time.Second,
			HTTPWriteTimeout:        5 * time.Second,
			HTTPIdleTimeout:         30 * time.Second,
			RPCReadTimeout:          10 * time.Second,
			RPCWriteTimeout:         10 * time.Second,
			GracefulShutdownTimeout: 5 * time.Second,
			EnablePProf:             true,
			PProfPort:               "0",
		},
		Clients: config.ClientsConfig{
			Database: config.DatabaseClientConfig{
				Enabled: false,
				URL:     "",
			},
			Redis: config.RedisClientConfig{
				Enabled: false,
				URL:     "",
			},
			Tracing: config.TracingClientConfig{
				Enabled:  false,
				Endpoint: "",
			},
		},
		Observability: config.ObservabilityConfig{
			LogLevel:       "debug",
			LogFormat:      "json",
			LogOutput:      "stdout",
			LogSampling:    false,
			LogDevelopment: true,
			MetricsEnabled: true,
			MetricsPort:    "0",
			MetricsPath:    "/metrics",
		},
		Middleware: config.MiddlewareConfig{
			CORSEnabled:          false,
			CORSAllowOrigins:     []string{"*"},
			CORSAllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			CORSAllowHeaders:     []string{"*"},
			CORSExposeHeaders:    []string{"X-Request-ID"},
			CORSAllowCredentials: false,
			CORSMaxAge:           12 * time.Hour,
			RateLimitEnabled:     false, // Disable rate limiting in tests
			RateLimitRate:        1000,
			RateLimitBurst:       2000,
			RateLimitType:        "ip",
			RecoveryStackTrace:   true,
			RecoveryStackSize:    4096,
			RecoveryPrintStack:   false,
			RequestIDHeader:      "X-Request-ID",
			TrustedProxies:       []string{},
			TrustAllProxies:      true,
			LogSkipPaths:         []string{"/health", "/ready", "/metrics"},
			LogRequestBody:       false,
			LogResponseBody:      false,
			LogSlowThreshold:     time.Second,
		},
		Features: config.FeaturesConfig{
			EnableNewAPI:       false,
			EnableBetaFeatures: false,
			MaintenanceMode:    false,
		},
	}
}

// WithContext creates a context with timeout for tests
func WithContext(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(func() {
		cancel()
	})
	return ctx, cancel
}

// AssertJSONResponse parses JSON response and checks status code
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, target interface{}) {
	t.Helper()

	require.Equal(t, expectedStatus, w.Code)

	if target != nil {
		err := json.NewDecoder(w.Body).Decode(target)
		require.NoError(t, err)
	}
}
