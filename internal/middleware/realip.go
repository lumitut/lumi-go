// Package middleware provides HTTP middleware components
package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RealIP extracts the real client IP address from request headers
// It checks headers in the following order:
// 1. CF-Connecting-IP (Cloudflare)
// 2. X-Real-IP
// 3. X-Forwarded-For (first IP if multiple)
// 4. Falls back to RemoteAddr
func RealIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := extractRealIP(c.Request)
		c.Set("client_ip", clientIP)

		// Override Gin's ClientIP method result
		c.Request.Header.Set("X-Real-IP", clientIP)

		c.Next()
	}
}

// extractRealIP extracts the real IP from the request
func extractRealIP(r *http.Request) string {
	// Check CF-Connecting-IP (Cloudflare)
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}

	// Check X-Real-IP
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Check X-Forwarded-For
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if i := strings.Index(xff, ","); i != -1 {
			xff = xff[:i]
		}
		xff = strings.TrimSpace(xff)
		if xff != "" {
			return xff
		}
	}

	// Check X-Forwarded
	if xf := r.Header.Get("X-Forwarded"); xf != "" {
		return xf
	}

	// Check Forwarded header (RFC 7239)
	if f := r.Header.Get("Forwarded"); f != "" {
		// Parse the Forwarded header
		for _, pair := range strings.Split(f, ";") {
			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(pair)), "for=") {
				ip := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(pair)), "for=")
				// Remove quotes if present
				ip = strings.Trim(ip, "\"")
				// Remove port if present
				if host, _, err := net.SplitHostPort(ip); err == nil {
					return host
				}
				return ip
			}
		}
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// RealIPConfig provides configuration for the RealIP middleware
type RealIPConfig struct {
	// TrustedProxies is a list of trusted proxy IP addresses or CIDR ranges
	TrustedProxies []string
	// TrustAll trusts all proxies (use with caution)
	TrustAll bool
}

// RealIPWithConfig creates a RealIP middleware with custom configuration
func RealIPWithConfig(config RealIPConfig) gin.HandlerFunc {
	// Parse trusted proxies
	var trustedNets []*net.IPNet
	for _, proxy := range config.TrustedProxies {
		if strings.Contains(proxy, "/") {
			// CIDR notation
			_, ipNet, err := net.ParseCIDR(proxy)
			if err == nil {
				trustedNets = append(trustedNets, ipNet)
			}
		} else {
			// Single IP
			ip := net.ParseIP(proxy)
			if ip != nil {
				// Convert to CIDR
				mask := net.CIDRMask(32, 32)
				if ip.To4() == nil {
					mask = net.CIDRMask(128, 128)
				}
				trustedNets = append(trustedNets, &net.IPNet{
					IP:   ip,
					Mask: mask,
				})
			}
		}
	}

	return func(c *gin.Context) {
		// Get the immediate client IP
		immediateIP, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		if err != nil {
			immediateIP = c.Request.RemoteAddr
		}

		// Check if we should trust this proxy
		trustProxy := config.TrustAll
		if !trustProxy && len(trustedNets) > 0 {
			ip := net.ParseIP(immediateIP)
			if ip != nil {
				for _, trustedNet := range trustedNets {
					if trustedNet.Contains(ip) {
						trustProxy = true
						break
					}
				}
			}
		}

		var clientIP string
		if trustProxy {
			// Use headers to determine real IP
			clientIP = extractRealIP(c.Request)
		} else {
			// Don't trust proxy headers, use immediate IP
			clientIP = immediateIP
		}

		c.Set("client_ip", clientIP)
		c.Request.Header.Set("X-Real-IP", clientIP)

		c.Next()
	}
}
