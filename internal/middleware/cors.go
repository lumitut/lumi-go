// Package middleware provides HTTP middleware components
package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CORSConfig provides configuration for CORS middleware
type CORSConfig struct {
	// Enabled determines if CORS is enabled (default: false for security)
	Enabled bool

	// AllowOrigins is a list of origins that are allowed.
	// Default value is []
	AllowOrigins []string

	// AllowOriginFunc is a function to determine if origin is allowed.
	// This allows for dynamic origin validation
	AllowOriginFunc func(origin string) bool

	// AllowMethods is a list of methods the client is allowed to use.
	// Default value is ["GET", "POST"]
	AllowMethods []string

	// AllowHeaders is list of headers the client is allowed to use.
	// Default value is ["Origin", "Content-Type", "Accept"]
	AllowHeaders []string

	// ExposeHeaders indicates which headers are safe to expose.
	// Default value is []
	ExposeHeaders []string

	// MaxAge indicates how long the results of a preflight request can be cached.
	// Default value is 12 hours
	MaxAge time.Duration

	// AllowCredentials indicates whether the request can include user credentials.
	// Default value is false
	AllowCredentials bool

	// AllowWildcard allows to add origins with wildcards like "http://*.example.com"
	AllowWildcard bool

	// AllowBrowserExtensions allows browser extensions to make requests.
	// Default value is false
	AllowBrowserExtensions bool

	// AllowWebSockets allows websocket upgrades
	// Default value is false
	AllowWebSockets bool

	// AllowFiles allows file:// schema (dangerous, use only in development)
	// Default value is false
	AllowFiles bool
}

// DefaultCORSConfig returns a secure default CORS configuration (CORS disabled)
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		Enabled:          false, // CORS off by default for security
		AllowOrigins:     []string{},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{},
		MaxAge:           12 * time.Hour,
		AllowCredentials: false,
		AllowWildcard:    false,
	}
}

// DevelopmentCORSConfig returns a permissive CORS configuration for development
func DevelopmentCORSConfig() CORSConfig {
	return CORSConfig{
		Enabled:          true,
		AllowOrigins:     []string{"http://localhost:*", "http://127.0.0.1:*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		MaxAge:           12 * time.Hour,
		AllowCredentials: true,
		AllowWildcard:    true,
		AllowWebSockets:  true,
	}
}

// CORS creates a CORS middleware with the given configuration
func CORS(config CORSConfig) gin.HandlerFunc {
	// If CORS is disabled, return a no-op middleware
	if !config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// Normalize configuration
	if len(config.AllowMethods) == 0 {
		config.AllowMethods = []string{"GET", "POST"}
	}
	if len(config.AllowHeaders) == 0 {
		config.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
	}
	if config.MaxAge == 0 {
		config.MaxAge = 12 * time.Hour
	}

	// Pre-compute header strings
	allowMethodsStr := strings.Join(config.AllowMethods, ", ")
	allowHeadersStr := strings.Join(config.AllowHeaders, ", ")
	exposeHeadersStr := strings.Join(config.ExposeHeaders, ", ")
	maxAgeStr := strconv.Itoa(int(config.MaxAge.Seconds()))

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		if origin != "" {
			allowed := isOriginAllowed(origin, config)

			if allowed {
				// Set CORS headers
				c.Header("Access-Control-Allow-Origin", origin)

				if config.AllowCredentials {
					c.Header("Access-Control-Allow-Credentials", "true")
				}

				if len(config.ExposeHeaders) > 0 {
					c.Header("Access-Control-Expose-Headers", exposeHeadersStr)
				}
			}

			// Handle preflight request
			if c.Request.Method == "OPTIONS" {
				if allowed {
					c.Header("Access-Control-Allow-Methods", allowMethodsStr)
					c.Header("Access-Control-Allow-Headers", allowHeadersStr)
					c.Header("Access-Control-Max-Age", maxAgeStr)
				}

				c.AbortWithStatus(http.StatusNoContent)
				return
			}
		}

		c.Next()
	}
}

// isOriginAllowed checks if the origin is allowed based on configuration
func isOriginAllowed(origin string, config CORSConfig) bool {
	// Check AllowOriginFunc first if provided
	if config.AllowOriginFunc != nil {
		return config.AllowOriginFunc(origin)
	}

	// Allow browser extensions if configured
	if config.AllowBrowserExtensions {
		if origin == "chrome-extension://" || strings.HasPrefix(origin, "moz-extension://") {
			return true
		}
	}

	// Allow file schema if configured (dangerous!)
	if config.AllowFiles && strings.HasPrefix(origin, "file://") {
		return true
	}

	// Allow WebSocket origins if configured
	if config.AllowWebSockets {
		if strings.HasPrefix(origin, "ws://") || strings.HasPrefix(origin, "wss://") {
			return true
		}
	}

	// Check if origin is in allowed list
	for _, allowedOrigin := range config.AllowOrigins {
		if config.AllowWildcard && strings.Contains(allowedOrigin, "*") {
			if matchWildcard(origin, allowedOrigin) {
				return true
			}
		} else if origin == allowedOrigin {
			return true
		}
	}

	return false
}

// matchWildcard matches origin against pattern with wildcards
func matchWildcard(origin, pattern string) bool {
	// Simple wildcard matching (e.g., "http://*.example.com")
	if !strings.Contains(pattern, "*") {
		return origin == pattern
	}

	// Handle subdomain wildcards
	if strings.HasPrefix(pattern, "http://*.") || strings.HasPrefix(pattern, "https://*.") {
		// Extract the domain part after *.
		parts := strings.SplitN(pattern, "*.", 2)
		if len(parts) == 2 {
			scheme := parts[0]
			domain := parts[1]

			// Check if origin matches the pattern
			if strings.HasPrefix(origin, scheme) && strings.HasSuffix(origin, domain) {
				// Ensure there's actually a subdomain
				remainingOrigin := strings.TrimPrefix(origin, scheme)
				remainingOrigin = strings.TrimSuffix(remainingOrigin, domain)
				// The remaining part should be a valid subdomain (not empty, not containing /)
				if remainingOrigin != "" && !strings.Contains(remainingOrigin, "/") {
					return true
				}
			}
		}
	}

	// Handle port wildcards (e.g., "http://localhost:*")
	if strings.HasSuffix(pattern, ":*") {
		base := strings.TrimSuffix(pattern, ":*")
		if strings.HasPrefix(origin, base+":") {
			return true
		}
	}

	return false
}

// CORSWithDefaults creates a CORS middleware with sensible defaults for APIs
func CORSWithDefaults() gin.HandlerFunc {
	config := CORSConfig{
		Enabled:      true,
		AllowOrigins: []string{}, // Must be explicitly configured
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
			http.MethodHead,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Request-ID",
			"X-Correlation-ID",
			"X-API-Key",
		},
		ExposeHeaders: []string{
			"X-Request-ID",
			"X-Correlation-ID",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
			"Content-Length",
			"Content-Type",
		},
		MaxAge:           12 * time.Hour,
		AllowCredentials: false,
		AllowWildcard:    false,
	}
	return CORS(config)
}

// ValidateOrigin is a helper function to validate origins for use with AllowOriginFunc
func ValidateOrigin(validOrigins []string) func(string) bool {
	originMap := make(map[string]bool)
	for _, origin := range validOrigins {
		originMap[origin] = true
	}

	return func(origin string) bool {
		return originMap[origin]
	}
}

// AllowLocalhost is a helper function that allows all localhost origins
func AllowLocalhost() func(string) bool {
	return func(origin string) bool {
		return strings.HasPrefix(origin, "http://localhost:") ||
			strings.HasPrefix(origin, "https://localhost:") ||
			strings.HasPrefix(origin, "http://127.0.0.1:") ||
			strings.HasPrefix(origin, "https://127.0.0.1:") ||
			origin == "http://localhost" ||
			origin == "https://localhost" ||
			origin == "http://127.0.0.1" ||
			origin == "https://127.0.0.1"
	}
}
