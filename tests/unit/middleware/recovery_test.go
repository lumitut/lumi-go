package middleware_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lumitut/lumi-go/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecoveryMiddleware(t *testing.T) {
	t.Run("recovers from panic", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.Use(middleware.Recovery())

		router.GET("/panic", func(c *gin.Context) {
			panic("test panic")
		})

		req, _ := http.NewRequest("GET", "/panic", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "internal_server_error", response["error"])
		assert.Equal(t, "An internal server error occurred", response["message"])
		assert.NotEmpty(t, response["request_id"])
	})

	t.Run("normal requests pass through", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.Use(middleware.Recovery())

		router.GET("/normal", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req, _ := http.NewRequest("GET", "/normal", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "ok", response["status"])
	})

	t.Run("recovers from error panic", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.Use(middleware.Recovery())

		router.GET("/error-panic", func(c *gin.Context) {
			panic(errors.New("custom error"))
		})

		req, _ := http.NewRequest("GET", "/error-panic", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestRecoveryWithConfig(t *testing.T) {
	t.Run("custom config without stack trace", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		config := middleware.RecoveryConfig{
			EnableStackTrace: false,
			StackTraceSize:   0,
			PrintStack:       false,
			LogLevel:         "error",
			IncludeRequest:   false,
		}
		router.Use(middleware.RecoveryWithConfig(config))

		router.GET("/panic", func(c *gin.Context) {
			panic("test panic")
		})

		req, _ := http.NewRequest("GET", "/panic", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("custom error handler", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		customHandlerCalled := false
		config := middleware.RecoveryConfig{
			EnableStackTrace: true,
			StackTraceSize:   4096,
			PrintStack:       false,
			LogLevel:         "error",
			IncludeRequest:   true,
			CustomHandler: func(c *gin.Context, err interface{}) {
				customHandlerCalled = true
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error":   "custom_error",
					"message": "Custom error message",
				})
			},
		}
		router.Use(middleware.RecoveryWithConfig(config))

		router.GET("/panic", func(c *gin.Context) {
			panic("test panic")
		})

		req, _ := http.NewRequest("GET", "/panic", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.True(t, customHandlerCalled)
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "custom_error", response["error"])
		assert.Equal(t, "Custom error message", response["message"])
	})
}

func TestRecoveryJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RecoveryJSON())

	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	req, _ := http.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "internal_server_error", response["error"])
	assert.NotEmpty(t, response["request_id"])
	assert.NotEmpty(t, response["timestamp"])
}

func TestDevelopmentRecovery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.DevelopmentRecovery())

	router.GET("/panic", func(c *gin.Context) {
		panic("development panic")
	})

	req, _ := http.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "internal_server_error", response["error"])
	assert.Contains(t, response["message"], "development panic")
	assert.NotEmpty(t, response["stack_trace"]) // Development mode includes stack trace
	assert.NotEmpty(t, response["request_id"])
	assert.Equal(t, "GET", response["method"])
	assert.Equal(t, "/panic", response["path"])
}

func TestRecoveryWithWriter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Custom writer to capture panic logs
	var logBuffer bytes.Buffer
	router.Use(middleware.RecoveryWithWriter(&logBuffer))

	router.GET("/panic", func(c *gin.Context) {
		panic("logged panic")
	})

	req, _ := http.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, logBuffer.String(), "panic recovered")
	assert.Contains(t, logBuffer.String(), "logged panic")
}

func TestCustomRecoveryWithWriter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	var logBuffer bytes.Buffer
	customHandlerCalled := false

	router.Use(middleware.CustomRecoveryWithWriter(&logBuffer, func(c *gin.Context, err interface{}) {
		customHandlerCalled = true
		c.JSON(http.StatusBadGateway, gin.H{
			"error": "gateway_error",
			"info":  fmt.Sprintf("%v", err),
		})
	}))

	router.GET("/panic", func(c *gin.Context) {
		panic("custom logged panic")
	})

	req, _ := http.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, customHandlerCalled)
	assert.Equal(t, http.StatusBadGateway, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "gateway_error", response["error"])
	assert.Equal(t, "custom logged panic", response["info"])
}
