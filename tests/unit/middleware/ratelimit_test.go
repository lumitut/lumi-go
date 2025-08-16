package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lumitut/lumi-go/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenBucketLimiter(t *testing.T) {
	t.Run("allows requests within rate limit", func(t *testing.T) {
		limiter := middleware.NewTokenBucketLimiter(10, 5, time.Minute, 5*time.Minute)

		// First 5 requests should be allowed (burst)
		for i := 0; i < 5; i++ {
			allowed, info := limiter.Allow("test-key")
			assert.True(t, allowed, "Request %d should be allowed", i+1)
			assert.Equal(t, 4-i, info.Remaining)
		}

		// 6th request should be denied
		allowed, info := limiter.Allow("test-key")
		assert.False(t, allowed, "6th request should be denied")
		assert.Equal(t, 0, info.Remaining)
	})

	t.Run("different keys have separate limits", func(t *testing.T) {
		limiter := middleware.NewTokenBucketLimiter(10, 2, time.Minute, 5*time.Minute)

		// Use limit for key1
		allowed1, _ := limiter.Allow("key1")
		allowed2, _ := limiter.Allow("key1")
		allowed3, _ := limiter.Allow("key1")

		assert.True(t, allowed1)
		assert.True(t, allowed2)
		assert.False(t, allowed3)

		// key2 should still have its limit
		allowed4, _ := limiter.Allow("key2")
		allowed5, _ := limiter.Allow("key2")
		allowed6, _ := limiter.Allow("key2")

		assert.True(t, allowed4)
		assert.True(t, allowed5)
		assert.False(t, allowed6)
	})

	t.Run("reset clears limit for key", func(t *testing.T) {
		limiter := middleware.NewTokenBucketLimiter(10, 2, time.Minute, 5*time.Minute)

		// Use up limit
		limiter.Allow("test-key")
		limiter.Allow("test-key")
		allowed, _ := limiter.Allow("test-key")
		assert.False(t, allowed)

		// Reset
		limiter.Reset("test-key")

		// Should be allowed again
		allowed, _ = limiter.Allow("test-key")
		assert.True(t, allowed)
	})
}

func TestRateLimitMiddleware(t *testing.T) {
	t.Run("allows requests within limit", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		config := middleware.DefaultRateLimitConfig()
		config.Rate = 10
		config.Burst = 3
		router.Use(middleware.RateLimit(config))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Make 3 requests (within burst)
		for i := 0; i < 3; i++ {
			req, _ := http.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.NotEmpty(t, w.Header().Get("X-RateLimit-Limit"))
			assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"))
		}
	})

	t.Run("blocks requests over limit", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		config := middleware.DefaultRateLimitConfig()
		config.Rate = 10
		config.Burst = 2
		router.Use(middleware.RateLimit(config))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Make requests up to burst limit
		for i := 0; i < 2; i++ {
			req, _ := http.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}

		// Next request should be rate limited
		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.NotEmpty(t, w.Header().Get("Retry-After"))

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "rate_limit_exceeded", response["error"])
	})

	t.Run("skips configured paths", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		config := middleware.DefaultRateLimitConfig()
		config.Rate = 1
		config.Burst = 1
		config.SkipPaths = []string{"/health"}
		router.Use(middleware.RateLimit(config))

		router.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "healthy"})
		})
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Health endpoint should not be rate limited
		for i := 0; i < 5; i++ {
			req, _ := http.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}

		// Test endpoint should be rate limited
		req1, _ := http.NewRequest("GET", "/test", nil)
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		req2, _ := http.NewRequest("GET", "/test", nil)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	})

	t.Run("custom key function", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		config := middleware.DefaultRateLimitConfig()
		config.Rate = 10
		config.Burst = 2
		config.KeyFunc = func(c *gin.Context) string {
			// Rate limit by API key
			return c.GetHeader("X-API-Key")
		}
		router.Use(middleware.RateLimit(config))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Requests with different API keys should have separate limits
		for i := 0; i < 2; i++ {
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("X-API-Key", "key1")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}

		// key1 should be rate limited now
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", "key1")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)

		// key2 should still work
		req2, _ := http.NewRequest("GET", "/test", nil)
		req2.Header.Set("X-API-Key", "key2")
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)
	})
}

func TestIPRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.IPRateLimit(5))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Should allow up to burst limit
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "5", w.Header().Get("X-RateLimit-Limit"))
}

func TestUserRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add correlation middleware first to set user ID
	router.Use(middleware.Correlation())
	router.Use(middleware.UserRateLimit(10))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Request with user ID
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-User-ID", "user-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "10", w.Header().Get("X-RateLimit-Limit"))
}

func TestAPIKeyRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.APIKeyRateLimit(20))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Request with API key
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "test-api-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "20", w.Header().Get("X-RateLimit-Limit"))
}

func TestSlidingWindowLimiter(t *testing.T) {
	t.Run("allows requests within window", func(t *testing.T) {
		limiter := middleware.NewSlidingWindowLimiter(3, time.Second, 5*time.Second)

		// Should allow 3 requests
		for i := 0; i < 3; i++ {
			allowed, info := limiter.Allow("test-key")
			assert.True(t, allowed)
			assert.Equal(t, 2-i, info.Remaining)
		}

		// 4th request should be denied
		allowed, _ := limiter.Allow("test-key")
		assert.False(t, allowed)

		// Wait for window to pass
		time.Sleep(1100 * time.Millisecond)

		// Should allow new request
		allowed, _ = limiter.Allow("test-key")
		assert.True(t, allowed)
	})

	t.Run("reset clears window", func(t *testing.T) {
		limiter := middleware.NewSlidingWindowLimiter(2, time.Second, 5*time.Second)

		// Use up limit
		limiter.Allow("test-key")
		limiter.Allow("test-key")
		allowed, _ := limiter.Allow("test-key")
		assert.False(t, allowed)

		// Reset
		limiter.Reset("test-key")

		// Should be allowed again
		allowed, _ = limiter.Allow("test-key")
		assert.True(t, allowed)
	})
}

func TestRateLimitByEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	limits := map[string]int{
		"/api/v1/users": 10,
		"/api/v1/posts": 20,
	}
	router.Use(middleware.RateLimitByEndpoint(limits))

	router.GET("/api/v1/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/api/v1/posts", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/api/v1/other", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Test users endpoint limit
	req1, _ := http.NewRequest("GET", "/api/v1/users", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, "10", w1.Header().Get("X-RateLimit-Limit"))

	// Test posts endpoint limit
	req2, _ := http.NewRequest("GET", "/api/v1/posts", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, "20", w2.Header().Get("X-RateLimit-Limit"))

	// Test default limit for other endpoints
	req3, _ := http.NewRequest("GET", "/api/v1/other", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, "60", w3.Header().Get("X-RateLimit-Limit")) // Default
}

func TestConcurrentRateLimiting(t *testing.T) {
	limiter := middleware.NewTokenBucketLimiter(100, 10, time.Minute, 5*time.Minute)

	var wg sync.WaitGroup
	successCount := 0
	mu := sync.Mutex{}

	// Run 20 concurrent requests
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed, _ := limiter.Allow("concurrent-key")
			if allowed {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Should allow up to burst limit
	assert.LessOrEqual(t, successCount, 10)
	assert.GreaterOrEqual(t, successCount, 8) // Allow some variance for timing
}
