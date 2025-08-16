// Package setup provides shared test setup utilities
package setup

import (
	"os"
	"testing"
)

// TestEnv sets up test environment variables
func TestEnv(t *testing.T) func() {
	t.Helper()

	// Save original environment
	originalEnv := make(map[string]string)
	for _, env := range os.Environ() {
		if len(env) > 0 {
			if i := indexOf(env, '='); i >= 0 {
				originalEnv[env[:i]] = env[i+1:]
			}
		}
	}

	// Set test environment variables
	testEnvVars := map[string]string{
		"ENVIRONMENT":        "test",
		"LOG_LEVEL":          "error",
		"SERVICE_NAME":       "test-service",
		"SERVICE_VERSION":    "test",
		"TRACING_ENABLED":    "false",
		"METRICS_ENABLED":    "true",
		"RATE_LIMIT_ENABLED": "true",
		"CORS_ENABLED":       "false",
		"DB_HOST":            "localhost",
		"DB_PORT":            "5432",
		"DB_NAME":            "test",
		"DB_USER":            "test",
		"DB_PASSWORD":        "test",
		"REDIS_HOST":         "localhost",
		"REDIS_PORT":         "6379",
	}

	for key, value := range testEnvVars {
		os.Setenv(key, value)
	}

	// Return cleanup function
	return func() {
		// Clear all env vars
		for key := range testEnvVars {
			os.Unsetenv(key)
		}

		// Restore original environment
		for key, value := range originalEnv {
			os.Setenv(key, value)
		}
	}
}

// SetTestDefaults sets default test configuration values
func SetTestDefaults() {
	defaults := map[string]string{
		"ENVIRONMENT":     "test",
		"LOG_LEVEL":       "error",
		"SERVICE_NAME":    "test-service",
		"SERVICE_VERSION": "test",
		"TRACING_ENABLED": "false",
	}

	for key, value := range defaults {
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}

func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}
