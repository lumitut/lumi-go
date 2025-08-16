package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lumitut/lumi-go/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware(t *testing.T) {
	t.Run("CORS disabled by default", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		config := middleware.DefaultCORSConfig()
		router.Use(middleware.CORS(config))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://example.com")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("CORS enabled with allowed origins", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		config := middleware.CORSConfig{
			Enabled:      true,
			AllowOrigins: []string{"http://example.com", "https://app.example.com"},
			AllowMethods: []string{"GET", "POST"},
			AllowHeaders: []string{"Content-Type", "Authorization"},
		}
		router.Use(middleware.CORS(config))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Test allowed origin
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://example.com")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "http://example.com", w.Header().Get("Access-Control-Allow-Origin"))

		// Test disallowed origin
		req2, _ := http.NewRequest("GET", "/test", nil)
		req2.Header.Set("Origin", "http://evil.com")
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusOK, w2.Code)
		assert.Empty(t, w2.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("preflight request handling", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		config := middleware.CORSConfig{
			Enabled:          true,
			AllowOrigins:     []string{"http://example.com"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
			AllowHeaders:     []string{"Content-Type", "Authorization", "X-Custom-Header"},
			ExposeHeaders:    []string{"X-Request-ID"},
			MaxAge:           12 * time.Hour,
			AllowCredentials: true,
		}
		router.Use(middleware.CORS(config))

		router.POST("/api/resource", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Preflight request
		req, _ := http.NewRequest("OPTIONS", "/api/resource", nil)
		req.Header.Set("Origin", "http://example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "http://example.com", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization, X-Custom-Header", w.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "43200", w.Header().Get("Access-Control-Max-Age"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("wildcard origin matching", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		config := middleware.CORSConfig{
			Enabled:       true,
			AllowOrigins:  []string{"http://*.example.com"},
			AllowMethods:  []string{"GET"},
			AllowWildcard: true,
		}
		router.Use(middleware.CORS(config))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Test subdomain match
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://api.example.com")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "http://api.example.com", w.Header().Get("Access-Control-Allow-Origin"))

		// Test another subdomain
		req2, _ := http.NewRequest("GET", "/test", nil)
		req2.Header.Set("Origin", "http://app.example.com")
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusOK, w2.Code)
		assert.Equal(t, "http://app.example.com", w2.Header().Get("Access-Control-Allow-Origin"))

		// Test non-matching domain
		req3, _ := http.NewRequest("GET", "/test", nil)
		req3.Header.Set("Origin", "http://example.org")
		w3 := httptest.NewRecorder()
		router.ServeHTTP(w3, req3)

		assert.Equal(t, http.StatusOK, w3.Code)
		assert.Empty(t, w3.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("development CORS config", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		config := middleware.DevelopmentCORSConfig()
		router.Use(middleware.CORS(config))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Test localhost origin
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("custom origin function", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		allowedOrigins := map[string]bool{
			"http://app1.com": true,
			"http://app2.com": true,
		}

		config := middleware.CORSConfig{
			Enabled: true,
			AllowOriginFunc: func(origin string) bool {
				return allowedOrigins[origin]
			},
			AllowMethods: []string{"GET", "POST"},
		}
		router.Use(middleware.CORS(config))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Test allowed origin
		req1, _ := http.NewRequest("GET", "/test", nil)
		req1.Header.Set("Origin", "http://app1.com")
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)

		assert.Equal(t, "http://app1.com", w1.Header().Get("Access-Control-Allow-Origin"))

		// Test another allowed origin
		req2, _ := http.NewRequest("GET", "/test", nil)
		req2.Header.Set("Origin", "http://app2.com")
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		assert.Equal(t, "http://app2.com", w2.Header().Get("Access-Control-Allow-Origin"))

		// Test disallowed origin
		req3, _ := http.NewRequest("GET", "/test", nil)
		req3.Header.Set("Origin", "http://app3.com")
		w3 := httptest.NewRecorder()
		router.ServeHTTP(w3, req3)

		assert.Empty(t, w3.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("browser extensions support", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		config := middleware.CORSConfig{
			Enabled:                true,
			AllowBrowserExtensions: true,
			AllowMethods:           []string{"GET"},
		}
		router.Use(middleware.CORS(config))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Test Chrome extension
		req1, _ := http.NewRequest("GET", "/test", nil)
		req1.Header.Set("Origin", "chrome-extension://")
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)

		assert.Equal(t, "chrome-extension://", w1.Header().Get("Access-Control-Allow-Origin"))

		// Test Firefox extension
		req2, _ := http.NewRequest("GET", "/test", nil)
		req2.Header.Set("Origin", "moz-extension://uuid-here")
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		assert.Equal(t, "moz-extension://uuid-here", w2.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("port wildcard matching", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		config := middleware.CORSConfig{
			Enabled:       true,
			AllowOrigins:  []string{"http://localhost:*"},
			AllowMethods:  []string{"GET"},
			AllowWildcard: true,
		}
		router.Use(middleware.CORS(config))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Test different ports
		ports := []string{"3000", "8080", "5173"}
		for _, port := range ports {
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", "http://localhost:"+port)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, "http://localhost:"+port, w.Header().Get("Access-Control-Allow-Origin"))
		}
	})
}

func TestCORSHelperFunctions(t *testing.T) {
	t.Run("ValidateOrigin", func(t *testing.T) {
		validOrigins := []string{"http://app1.com", "http://app2.com"}
		validator := middleware.ValidateOrigin(validOrigins)

		assert.True(t, validator("http://app1.com"))
		assert.True(t, validator("http://app2.com"))
		assert.False(t, validator("http://app3.com"))
	})

	t.Run("AllowLocalhost", func(t *testing.T) {
		validator := middleware.AllowLocalhost()

		// Should allow various localhost formats
		assert.True(t, validator("http://localhost"))
		assert.True(t, validator("https://localhost"))
		assert.True(t, validator("http://localhost:3000"))
		assert.True(t, validator("https://localhost:8080"))
		assert.True(t, validator("http://127.0.0.1"))
		assert.True(t, validator("https://127.0.0.1"))
		assert.True(t, validator("http://127.0.0.1:3000"))
		assert.True(t, validator("https://127.0.0.1:8080"))

		// Should not allow other origins
		assert.False(t, validator("http://example.com"))
		assert.False(t, validator("http://192.168.1.1"))
	})
}

func TestCORSWithDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.CORSWithDefaults())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Should not allow any origin by default (must be explicitly configured)
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}
