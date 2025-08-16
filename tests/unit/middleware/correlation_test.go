package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lumitut/lumi-go/internal/middleware"
	"github.com/lumitut/lumi-go/tests/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCorrelationMiddleware(t *testing.T) {
	tests := []struct {
		name               string
		requestHeaders     map[string]string
		expectedHeaders    []string
		checkContextValues []string
	}{
		{
			name:               "generates new request ID when not provided",
			requestHeaders:     map[string]string{},
			expectedHeaders:    []string{"X-Request-ID", "X-Correlation-ID"},
			checkContextValues: []string{"request_id", "correlation_id"},
		},
		{
			name: "uses provided request ID",
			requestHeaders: map[string]string{
				"X-Request-ID": "test-request-123",
			},
			expectedHeaders:    []string{"X-Request-ID", "X-Correlation-ID"},
			checkContextValues: []string{"request_id", "correlation_id"},
		},
		{
			name: "uses provided correlation ID",
			requestHeaders: map[string]string{
				"X-Correlation-ID": "test-correlation-456",
			},
			expectedHeaders:    []string{"X-Request-ID", "X-Correlation-ID"},
			checkContextValues: []string{"request_id", "correlation_id"},
		},
		{
			name: "extracts user and tenant IDs",
			requestHeaders: map[string]string{
				"X-User-ID":   "user-789",
				"X-Tenant-ID": "tenant-abc",
			},
			expectedHeaders:    []string{"X-Request-ID", "X-Correlation-ID"},
			checkContextValues: []string{"request_id", "correlation_id", "user_id", "tenant_id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(middleware.Correlation())

			// Test handler to capture context values
			var capturedValues map[string]interface{}
			router.GET("/test", func(c *gin.Context) {
				capturedValues = make(map[string]interface{})
				for _, key := range tt.checkContextValues {
					if val, exists := c.Get(key); exists {
						capturedValues[key] = val
					}
				}
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			// Create request
			req, err := http.NewRequest("GET", "/test", nil)
			require.NoError(t, err)

			// Add headers
			for key, value := range tt.requestHeaders {
				req.Header.Set(key, value)
			}

			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, http.StatusOK, w.Code)

			// Check response headers
			for _, header := range tt.expectedHeaders {
				assert.NotEmpty(t, w.Header().Get(header), "Expected header %s to be set", header)
			}

			// Check context values
			for _, key := range tt.checkContextValues {
				assert.Contains(t, capturedValues, key, "Expected context value %s to be set", key)
			}

			// Verify specific header values if provided
			if providedRequestID := tt.requestHeaders["X-Request-ID"]; providedRequestID != "" {
				assert.Equal(t, providedRequestID, w.Header().Get("X-Request-ID"))
				assert.Equal(t, providedRequestID, capturedValues["request_id"])
			}

			if providedCorrelationID := tt.requestHeaders["X-Correlation-ID"]; providedCorrelationID != "" {
				assert.Equal(t, providedCorrelationID, w.Header().Get("X-Correlation-ID"))
				assert.Equal(t, providedCorrelationID, capturedValues["correlation_id"])
			}

			if providedUserID := tt.requestHeaders["X-User-ID"]; providedUserID != "" {
				assert.Equal(t, providedUserID, capturedValues["user_id"])
			}

			if providedTenantID := tt.requestHeaders["X-Tenant-ID"]; providedTenantID != "" {
				assert.Equal(t, providedTenantID, capturedValues["tenant_id"])
			}
		})
	}
}

func TestCorrelationExtractors(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.Correlation())

	testRequestID := "test-request-789"
	testCorrelationID := "test-correlation-xyz"
	testUserID := "user-456"

	router.GET("/test", func(c *gin.Context) {
		// Test extractors
		extractedRequestID := middleware.ExtractRequestID(c)
		extractedCorrelationID := middleware.ExtractCorrelationID(c)
		extractedUserID := middleware.ExtractUserID(c)
		extractedTraceID := middleware.ExtractTraceID(c)

		c.JSON(http.StatusOK, gin.H{
			"request_id":     extractedRequestID,
			"correlation_id": extractedCorrelationID,
			"user_id":        extractedUserID,
			"trace_id":       extractedTraceID,
		})
	})

	// Create request
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	req.Header.Set("X-Request-ID", testRequestID)
	req.Header.Set("X-Correlation-ID", testCorrelationID)
	req.Header.Set("X-User-ID", testUserID)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Parse response
	var response map[string]string
	helpers.AssertJSONResponse(t, w, http.StatusOK, &response)

	// Assertions
	assert.Equal(t, testRequestID, response["request_id"])
	assert.Equal(t, testCorrelationID, response["correlation_id"])
	assert.Equal(t, testUserID, response["user_id"])
}

func TestContextFromGin(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.Correlation())

	router.GET("/test", func(c *gin.Context) {
		// Get context with correlation values
		ctx := middleware.ContextFromGin(c)

		// Context should have values
		assert.NotNil(t, ctx)

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Create request
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	req.Header.Set("X-Request-ID", "test-123")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
}
