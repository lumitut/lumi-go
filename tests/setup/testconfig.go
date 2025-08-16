// Package helpers provides test utilities and helpers
package setup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lumitut/lumi-go/internal/config"
	"github.com/lumitut/lumi-go/internal/httpapi"
	"github.com/lumitut/lumi-go/internal/observability/logger"
	"github.com/lumitut/lumi-go/internal/observability/metrics"
	"github.com/stretchr/testify/require"
)

// TestConfig returns a test configuration
func TestConfig() *config.Config {
	return &config.Config{
		Service: config.ServiceConfig{
			Name:        "test-service",
			Version:     "test",
			Environment: "test",
			LogLevel:    "error", // Reduce noise in tests
		},
		Server: config.ServerConfig{
			HTTPPort:                "8080",
			HTTPReadTimeout:         15 * time.Second,
			HTTPWriteTimeout:        15 * time.Second,
			HTTPIdleTimeout:         60 * time.Second,
			RPCPort:                 "8081",
			RPCReadTimeout:          30 * time.Second,
			RPCWriteTimeout:         30 * time.Second,
			GracefulShutdownTimeout: 5 * time.Second,
			EnablePProf:             true,
			PProfPort:               "6060",
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
			LogLevel:       "error",
			LogFormat:      "json",
			LogOutput:      "stdout",
			LogSampling:    false,
			LogDevelopment: false,
			MetricsEnabled: true,
			MetricsPort:    "9090",
			MetricsPath:    "/metrics",
		},
		Middleware: config.MiddlewareConfig{
			CORSEnabled:          false,
			CORSAllowOrigins:     []string{},
			CORSAllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
			CORSAllowHeaders:     []string{"Content-Type", "Authorization"},
			CORSExposeHeaders:    []string{"X-Request-ID"},
			CORSAllowCredentials: false,
			CORSMaxAge:           12 * time.Hour,
			RateLimitEnabled:     true,
			RateLimitRate:        100,
			RateLimitBurst:       10,
			RateLimitType:        "ip",
			RecoveryStackTrace:   true,
			RecoveryStackSize:    4096,
			RecoveryPrintStack:   false,
			RequestIDHeader:      "X-Request-ID",
			TrustedProxies:       []string{},
			TrustAllProxies:      false,
			LogSkipPaths:         []string{"/health", "/ready", "/metrics"},
			LogRequestBody:       false,
			LogResponseBody:      false,
			LogSlowThreshold:     1 * time.Second,
		},
		Features: config.FeaturesConfig{
			EnableNewAPI:       false,
			EnableBetaFeatures: false,
			MaintenanceMode:    false,
		},
	}
}

// SetupTest sets up the test environment
func SetupTest(t *testing.T) (*config.Config, func()) {
	// Save original env vars
	originalEnv := os.Environ()

	// Set test environment
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("LOG_LEVEL", "error")
	os.Setenv("TRACING_ENABLED", "false")

	// Initialize logger
	logConfig := logger.Config{
		Level:             "error",
		Format:            "json",
		Development:       false,
		DisableCaller:     true,
		DisableStacktrace: true,
		SampleInitial:     100,
		SampleThereafter:  100,
	}
	err := logger.Initialize(logConfig)
	require.NoError(t, err)

	// Initialize metrics
	metrics.Initialize("test_service", "api")

	// Get test config
	cfg := TestConfig()

	// Return cleanup function
	cleanup := func() {
		logger.Sync()

		// Restore original env vars
		os.Clearenv()
		for _, env := range originalEnv {
			if i := bytes.IndexByte([]byte(env), '='); i >= 0 {
				os.Setenv(env[:i], env[i+1:])
			}
		}
	}

	return cfg, cleanup
}

// SetupTestServer creates a test HTTP server
func SetupTestServer(t *testing.T) (*httptest.Server, *gin.Engine, func()) {
	cfg, cleanup := SetupTest(t)

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create server
	server := httpapi.NewServer(cfg)

	// Create test server
	ts := httptest.NewServer(server.Router())

	// Return cleanup function
	serverCleanup := func() {
		ts.Close()
		cleanup()
	}

	return ts, server.Router(), serverCleanup
}

// CreateTestContext creates a test gin context
func CreateTestContext(w http.ResponseWriter) (*gin.Context, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	ctx, router := gin.CreateTestContext(w)
	return ctx, router
}

// MakeRequest makes a test HTTP request
func MakeRequest(t *testing.T, router *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, path, bodyReader)
	require.NoError(t, err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}

// MakeRequestWithHeaders makes a test HTTP request with custom headers
func MakeRequestWithHeaders(t *testing.T, router *gin.Engine, method, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, path, bodyReader)
	require.NoError(t, err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}

// AssertJSONResponse asserts that the response is valid JSON and unmarshals it
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, target interface{}) {
	require.Equal(t, expectedStatus, w.Code)
	require.Contains(t, w.Header().Get("Content-Type"), "application/json")

	if target != nil && w.Body.Len() > 0 {
		err := json.Unmarshal(w.Body.Bytes(), target)
		require.NoError(t, err)
	}
}

// TestContext returns a test context
func TestContext() context.Context {
	return context.Background()
}

// WaitForServer waits for the server to be ready
func WaitForServer(url string, timeout time.Duration) error {
	client := &http.Client{Timeout: time.Second}
	start := time.Now()

	for time.Since(start) < timeout {
		resp, err := client.Get(url + "/healthz")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("server did not become ready within %v", timeout)
}

// CleanupTest performs common test cleanup
func CleanupTest() {
	// Reset gin mode
	gin.SetMode(gin.DebugMode)

	// Clear any test data
	os.Clearenv()
}
