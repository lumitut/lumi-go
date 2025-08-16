// Package middleware provides HTTP middleware components
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lumitut/lumi-go/internal/observability/logger"
	"github.com/lumitut/lumi-go/internal/observability/metrics"
	"go.uber.org/zap"
)

// RateLimiter interface for rate limiting implementations
type RateLimiter interface {
	Allow(key string) (bool, RateLimitInfo)
	Reset(key string)
}

// RateLimitInfo contains rate limit information
type RateLimitInfo struct {
	Limit     int
	Remaining int
	ResetTime time.Time
}

// TokenBucketLimiter implements token bucket algorithm
type TokenBucketLimiter struct {
	mu      sync.RWMutex
	buckets map[string]*bucket
	rate    int           // tokens per interval
	burst   int           // max tokens in bucket
	ttl     time.Duration // TTL for inactive buckets
	cleanup time.Duration // cleanup interval
}

type bucket struct {
	tokens    int
	lastFill  time.Time
	resetTime time.Time
}

// NewTokenBucketLimiter creates a new token bucket rate limiter
func NewTokenBucketLimiter(rate, burst int, interval, ttl time.Duration) *TokenBucketLimiter {
	limiter := &TokenBucketLimiter{
		buckets: make(map[string]*bucket),
		rate:    rate,
		burst:   burst,
		ttl:     ttl,
		cleanup: ttl / 2,
	}

	// Start cleanup goroutine
	go limiter.cleanupRoutine()

	return limiter
}

// Allow checks if request is allowed
func (l *TokenBucketLimiter) Allow(key string) (bool, RateLimitInfo) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, exists := l.buckets[key]

	if !exists {
		// Create new bucket
		b = &bucket{
			tokens:    l.burst - 1,
			lastFill:  now,
			resetTime: now.Add(time.Minute),
		}
		l.buckets[key] = b

		return true, RateLimitInfo{
			Limit:     l.rate,
			Remaining: b.tokens,
			ResetTime: b.resetTime,
		}
	}

	// Refill tokens based on time elapsed
	elapsed := now.Sub(b.lastFill)
	tokensToAdd := int(elapsed.Seconds()) * l.rate / 60 // tokens per second
	b.tokens = min(b.tokens+tokensToAdd, l.burst)
	b.lastFill = now

	// Check if we have tokens
	if b.tokens > 0 {
		b.tokens--
		return true, RateLimitInfo{
			Limit:     l.rate,
			Remaining: b.tokens,
			ResetTime: b.resetTime,
		}
	}

	// Update reset time if needed
	if now.After(b.resetTime) {
		b.resetTime = now.Add(time.Minute)
		b.tokens = l.burst - 1
		return true, RateLimitInfo{
			Limit:     l.rate,
			Remaining: b.tokens,
			ResetTime: b.resetTime,
		}
	}

	return false, RateLimitInfo{
		Limit:     l.rate,
		Remaining: 0,
		ResetTime: b.resetTime,
	}
}

// Reset resets the rate limit for a key
func (l *TokenBucketLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.buckets, key)
}

// cleanupRoutine periodically cleans up old buckets
func (l *TokenBucketLimiter) cleanupRoutine() {
	ticker := time.NewTicker(l.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for key, b := range l.buckets {
			if now.Sub(b.lastFill) > l.ttl {
				delete(l.buckets, key)
			}
		}
		l.mu.Unlock()
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RateLimitConfig provides configuration for rate limiting
type RateLimitConfig struct {
	// Enabled enables rate limiting
	Enabled bool
	// Rate is the number of requests per minute
	Rate int
	// Burst is the maximum burst size
	Burst int
	// KeyFunc generates the rate limit key from the request
	KeyFunc func(*gin.Context) string
	// ErrorHandler handles rate limit errors
	ErrorHandler func(*gin.Context, RateLimitInfo)
	// SkipPaths skips rate limiting for these paths
	SkipPaths []string
	// SkipFunc allows custom skip logic
	SkipFunc func(*gin.Context) bool
}

// DefaultRateLimitConfig returns default rate limit configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Enabled: true,
		Rate:    60, // 60 requests per minute
		Burst:   10, // Allow burst of 10
		KeyFunc: func(c *gin.Context) string {
			// Rate limit by IP by default
			return c.ClientIP()
		},
		ErrorHandler: defaultRateLimitErrorHandler,
		SkipPaths:    []string{"/health", "/ready", "/metrics"},
	}
}

// defaultRateLimitErrorHandler is the default error handler
func defaultRateLimitErrorHandler(c *gin.Context, info RateLimitInfo) {
	c.Header("X-RateLimit-Limit", strconv.Itoa(info.Limit))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(info.ResetTime.Unix(), 10))
	c.Header("Retry-After", strconv.Itoa(int(time.Until(info.ResetTime).Seconds())))

	c.JSON(http.StatusTooManyRequests, gin.H{
		"error":       "rate_limit_exceeded",
		"message":     "Too many requests. Please try again later.",
		"retry_after": int(time.Until(info.ResetTime).Seconds()),
	})
	c.Abort()
}

// RateLimit creates a rate limiting middleware
func RateLimit(config RateLimitConfig) gin.HandlerFunc {
	if !config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// Create limiter
	limiter := NewTokenBucketLimiter(
		config.Rate,
		config.Burst,
		time.Minute,
		5*time.Minute,
	)

	// Build skip map
	skipMap := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipMap[path] = true
	}

	return func(c *gin.Context) {
		// Check if should skip
		if skipMap[c.Request.URL.Path] || (config.SkipFunc != nil && config.SkipFunc(c)) {
			c.Next()
			return
		}

		// Get rate limit key
		key := config.KeyFunc(c)
		if key == "" {
			c.Next()
			return
		}

		// Check rate limit
		allowed, info := limiter.Allow(key)

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(info.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(info.ResetTime.Unix(), 10))

		if !allowed {
			// Log rate limit exceeded
			logger.Warn(c.Request.Context(), "Rate limit exceeded",
				zap.String("key", key),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("ip", c.ClientIP()),
			)

			// Record metric
			if m := metrics.Get(); m != nil {
				m.HTTPRequestsTotal.WithLabelValues(
					c.Request.Method,
					c.FullPath(),
					"429",
				).Inc()
			}

			// Handle error
			config.ErrorHandler(c, info)
			return
		}

		c.Next()
	}
}

// IPRateLimit creates a simple IP-based rate limiter
func IPRateLimit(requestsPerMinute int) gin.HandlerFunc {
	config := DefaultRateLimitConfig()
	config.Rate = requestsPerMinute
	config.Burst = min(requestsPerMinute/6, 10) // Allow 10% burst or 10, whichever is smaller
	return RateLimit(config)
}

// UserRateLimit creates a user-based rate limiter
func UserRateLimit(requestsPerMinute int) gin.HandlerFunc {
	config := DefaultRateLimitConfig()
	config.Rate = requestsPerMinute
	config.Burst = min(requestsPerMinute/6, 20) // Allow 10% burst or 20, whichever is smaller
	config.KeyFunc = func(c *gin.Context) string {
		// Try to get user ID from context or header
		if userID := ExtractUserID(c); userID != "" {
			return fmt.Sprintf("user:%s", userID)
		}
		// Fall back to IP
		return c.ClientIP()
	}
	return RateLimit(config)
}

// APIKeyRateLimit creates an API key-based rate limiter
func APIKeyRateLimit(requestsPerMinute int) gin.HandlerFunc {
	config := DefaultRateLimitConfig()
	config.Rate = requestsPerMinute
	config.Burst = min(requestsPerMinute/6, 50) // Allow 10% burst or 50, whichever is smaller
	config.KeyFunc = func(c *gin.Context) string {
		// Try to get API key from header
		if apiKey := c.GetHeader("X-API-Key"); apiKey != "" {
			return fmt.Sprintf("api:%s", apiKey)
		}
		// Try Bearer token
		if auth := c.GetHeader("Authorization"); len(auth) > 7 && auth[:7] == "Bearer " {
			return fmt.Sprintf("bearer:%s", auth[7:])
		}
		// Fall back to IP
		return c.ClientIP()
	}
	return RateLimit(config)
}

// SlidingWindowLimiter implements sliding window algorithm
type SlidingWindowLimiter struct {
	mu      sync.RWMutex
	windows map[string]*slidingWindow
	limit   int
	window  time.Duration
	ttl     time.Duration
	cleanup time.Duration
}

type slidingWindow struct {
	requests []time.Time
	lastSeen time.Time
}

// NewSlidingWindowLimiter creates a new sliding window rate limiter
func NewSlidingWindowLimiter(limit int, window, ttl time.Duration) *SlidingWindowLimiter {
	limiter := &SlidingWindowLimiter{
		windows: make(map[string]*slidingWindow),
		limit:   limit,
		window:  window,
		ttl:     ttl,
		cleanup: ttl / 2,
	}

	// Start cleanup goroutine
	go limiter.cleanupRoutine()

	return limiter
}

// Allow checks if request is allowed
func (l *SlidingWindowLimiter) Allow(key string) (bool, RateLimitInfo) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	w, exists := l.windows[key]

	if !exists {
		// Create new window
		w = &slidingWindow{
			requests: []time.Time{now},
			lastSeen: now,
		}
		l.windows[key] = w

		return true, RateLimitInfo{
			Limit:     l.limit,
			Remaining: l.limit - 1,
			ResetTime: now.Add(l.window),
		}
	}

	// Remove old requests outside the window
	cutoff := now.Add(-l.window)
	validRequests := []time.Time{}
	for _, req := range w.requests {
		if req.After(cutoff) {
			validRequests = append(validRequests, req)
		}
	}
	w.requests = validRequests
	w.lastSeen = now

	// Check if we're under the limit
	if len(w.requests) < l.limit {
		w.requests = append(w.requests, now)
		return true, RateLimitInfo{
			Limit:     l.limit,
			Remaining: l.limit - len(w.requests),
			ResetTime: w.requests[0].Add(l.window),
		}
	}

	// Calculate when the oldest request will expire
	resetTime := w.requests[0].Add(l.window)

	return false, RateLimitInfo{
		Limit:     l.limit,
		Remaining: 0,
		ResetTime: resetTime,
	}
}

// Reset resets the rate limit for a key
func (l *SlidingWindowLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.windows, key)
}

// cleanupRoutine periodically cleans up old windows
func (l *SlidingWindowLimiter) cleanupRoutine() {
	ticker := time.NewTicker(l.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for key, w := range l.windows {
			if now.Sub(w.lastSeen) > l.ttl {
				delete(l.windows, key)
			}
		}
		l.mu.Unlock()
	}
}

// DistributedRateLimiter implements distributed rate limiting using Redis
type DistributedRateLimiter struct {
	// This would use Redis or another distributed store
	// Implementation depends on your cache/redis package
	// Placeholder for now
}

// RateLimitByEndpoint creates per-endpoint rate limits
func RateLimitByEndpoint(limits map[string]int) gin.HandlerFunc {
	limiters := make(map[string]*TokenBucketLimiter)

	for endpoint, limit := range limits {
		limiters[endpoint] = NewTokenBucketLimiter(
			limit,
			min(limit/6, 10),
			time.Minute,
			5*time.Minute,
		)
	}

	defaultLimiter := NewTokenBucketLimiter(60, 10, time.Minute, 5*time.Minute)

	return func(c *gin.Context) {
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Get limiter for this endpoint
		limiter, exists := limiters[path]
		if !exists {
			limiter = defaultLimiter
		}

		// Get rate limit key (by IP)
		key := c.ClientIP()

		// Check rate limit
		allowed, info := limiter.Allow(key)

		// Set headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(info.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(info.ResetTime.Unix(), 10))

		if !allowed {
			logger.Warn(context.Background(), "Rate limit exceeded",
				zap.String("endpoint", path),
				zap.String("ip", key),
			)

			defaultRateLimitErrorHandler(c, info)
			return
		}

		c.Next()
	}
}
