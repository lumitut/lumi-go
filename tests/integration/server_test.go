package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/lumitut/lumi-go/internal/httpapi"
	"github.com/lumitut/lumi-go/tests/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerLifecycle(t *testing.T) {
	// Setup
	cfg, cleanup := helpers.SetupTest(t)
	defer cleanup()

	// Use a random port to avoid conflicts
	cfg.Server.HTTPPort = "0" // Let the OS assign a port

	// Create server
	server := httpapi.NewServer(cfg)
	require.NotNil(t, server)

	// Start server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.Start(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Check if server is ready
	assert.True(t, server.IsReady())

	// Shutdown server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	err := server.Shutdown(shutdownCtx)
	assert.NoError(t, err)

	// Check if server is not ready after shutdown
	assert.False(t, server.IsReady())
}

func TestHealthEndpoints(t *testing.T) {
	// Setup test server
	ts, router, cleanup := helpers.SetupTestServer(t)
	defer cleanup()

	tests := []struct {
		name     string
		endpoint string
	}{
		{"health endpoint", "/health"},
		{"healthz endpoint", "/healthz"},
		{"ready endpoint", "/ready"},
		{"readyz endpoint", "/readyz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(ts.URL + tt.endpoint)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			if tt.endpoint == "/health" || tt.endpoint == "/healthz" {
				assert.Equal(t, "healthy", result["status"])
			} else {
				assert.Equal(t, "ready", result["status"])
			}
			assert.NotNil(t, result["time"])
		})
	}
}

func TestMetricsEndpoint(t *testing.T) {
	// Setup test server
	ts, _, cleanup := helpers.SetupTestServer(t)
	defer cleanup()

	resp, err := http.Get(ts.URL + "/metrics")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/plain")

	// Read a few bytes to verify it's Prometheus format
	buf := make([]byte, 100)
	n, _ := resp.Body.Read(buf)
	content := string(buf[:n])
	assert.Contains(t, content, "# HELP")
}

func TestVersionEndpoint(t *testing.T) {
	// Setup test server
	ts, _, cleanup := helpers.SetupTestServer(t)
	defer cleanup()

	resp, err := http.Get(ts.URL + "/version")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "test-service", result["service"])
	assert.Equal(t, "test", result["version"])
	assert.Equal(t, "test", result["environment"])
}

func TestAPIEndpoints(t *testing.T) {
	// Setup test server
	ts, _, cleanup := helpers.SetupTestServer(t)
	defer cleanup()

	t.Run("GET user", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/users/123")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var user map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&user)
		require.NoError(t, err)

		assert.Equal(t, "123", user["id"])
		assert.Equal(t, "john_doe", user["username"])
		assert.NotNil(t, user["email"])
		assert.NotNil(t, user["created_at"])
	})

	t.Run("LIST users", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/users")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		users, ok := result["users"].([]interface{})
		require.True(t, ok)
		assert.Len(t, users, 2)
		assert.Equal(t, float64(2), result["total"])
	})

	t.Run("CREATE user", func(t *testing.T) {
		client := &http.Client{}
		body := `{"username":"test_user","email":"test@example.com"}`
		req, err := http.NewRequest("POST", ts.URL+"/api/v1/users", bytes.NewBufferString(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var user map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&user)
		require.NoError(t, err)

		assert.NotEmpty(t, user["id"])
		assert.Equal(t, "test_user", user["username"])
		assert.Equal(t, "test@example.com", user["email"])
	})

	t.Run("CREATE user with invalid data", func(t *testing.T) {
		client := &http.Client{}
		body := `{"username":"test_user"}` // Missing email
		req, err := http.NewRequest("POST", ts.URL+"/api/v1/users", bytes.NewBufferString(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "invalid_request", result["error"])
	})

	t.Run("UPDATE user", func(t *testing.T) {
		client := &http.Client{}
		body := `{"username":"updated_user"}`
		req, err := http.NewRequest("PUT", ts.URL+"/api/v1/users/123", bytes.NewBufferString(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var user map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&user)
		require.NoError(t, err)

		assert.Equal(t, "123", user["id"])
		assert.NotNil(t, user["updated_at"])
	})

	t.Run("DELETE user", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("DELETE", ts.URL+"/api/v1/users/123", nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
}

func TestMiddlewareIntegration(t *testing.T) {
	// Setup test server
	ts, _, cleanup := helpers.SetupTestServer(t)
	defer cleanup()

	t.Run("correlation IDs", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", ts.URL+"/api/v1/users/123", nil)
		require.NoError(t, err)
		req.Header.Set("X-Request-ID", "test-request-123")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, "test-request-123", resp.Header.Get("X-Request-ID"))
		assert.NotEmpty(t, resp.Header.Get("X-Correlation-ID"))
	})

	t.Run("rate limiting", func(t *testing.T) {
		// Make multiple requests to trigger rate limit
		client := &http.Client{}

		var lastResp *http.Response
		for i := 0; i < 15; i++ {
			req, err := http.NewRequest("GET", ts.URL+"/api/v1/users/123", nil)
			require.NoError(t, err)

			resp, err := client.Do(req)
			require.NoError(t, err)

			if lastResp != nil {
				lastResp.Body.Close()
			}
			lastResp = resp
		}

		// Check rate limit headers
		assert.NotEmpty(t, lastResp.Header.Get("X-RateLimit-Limit"))
		assert.NotEmpty(t, lastResp.Header.Get("X-RateLimit-Remaining"))
		assert.NotEmpty(t, lastResp.Header.Get("X-RateLimit-Reset"))

		lastResp.Body.Close()
	})

	t.Run("404 handling", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/nonexistent")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestPprofEndpoints(t *testing.T) {
	// Setup test server
	ts, _, cleanup := helpers.SetupTestServer(t)
	defer cleanup()

	endpoints := []string{
		"/debug/pprof/",
		"/debug/pprof/heap",
		"/debug/pprof/goroutine",
		"/debug/pprof/allocs",
		"/debug/pprof/block",
		"/debug/pprof/mutex",
		"/debug/pprof/threadcreate",
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			resp, err := http.Get(ts.URL + endpoint)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			// Read a few bytes to ensure there's content
			buf := make([]byte, 100)
			n, _ := resp.Body.Read(buf)
			assert.Greater(t, n, 0)
		})
	}
}

func TestGracefulShutdown(t *testing.T) {
	// Setup
	cfg, cleanup := helpers.SetupTest(t)
	defer cleanup()

	// Create and start server
	server := httpapi.NewServer(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.Start(ctx)
	}()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)
	assert.True(t, server.IsReady())

	// Initiate shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.GracefulShutdownTimeout)
	defer shutdownCancel()

	err := server.Shutdown(shutdownCtx)
	assert.NoError(t, err)

	// Server should not be ready after shutdown
	assert.False(t, server.IsReady())

	// Wait for server goroutine to finish
	select {
	case <-serverErr:
		// Server stopped successfully
	case <-time.After(2 * time.Second):
		t.Fatal("Server did not stop within timeout")
	}
}
