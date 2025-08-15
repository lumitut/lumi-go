// Package middleware provides HTTP middleware components
package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lumitut/lumi-go/internal/observability/logger"
	"go.opentelemetry.io/otel/trace"
)

// Common header names for correlation
const (
	HeaderRequestID     = "X-Request-ID"
	HeaderCorrelationID = "X-Correlation-ID" 
	HeaderTraceID       = "X-Trace-ID"
	HeaderSpanID        = "X-Span-ID"
	HeaderUserID        = "X-User-ID"
	HeaderTenantID      = "X-Tenant-ID"
	HeaderForwardedFor  = "X-Forwarded-For"
	HeaderRealIP        = "X-Real-IP"
)

// Correlation adds correlation IDs to the request context
func Correlation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get or generate request ID
		requestID := c.GetHeader(HeaderRequestID)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Writer.Header().Set(HeaderRequestID, requestID)

		// Get or generate correlation ID
		correlationID := c.GetHeader(HeaderCorrelationID)
		if correlationID == "" {
			correlationID = requestID // Use request ID as correlation ID if not provided
		}
		c.Set("correlation_id", correlationID)
		c.Writer.Header().Set(HeaderCorrelationID, correlationID)

		// Extract trace context if available
		if span := trace.SpanFromContext(c.Request.Context()); span.SpanContext().IsValid() {
			traceID := span.SpanContext().TraceID().String()
			spanID := span.SpanContext().SpanID().String()
			
			c.Set("trace_id", traceID)
			c.Set("span_id", spanID)
			c.Writer.Header().Set(HeaderTraceID, traceID)
			c.Writer.Header().Set(HeaderSpanID, spanID)
		}

		// Extract user context if available
		if userID := c.GetHeader(HeaderUserID); userID != "" {
			c.Set("user_id", userID)
		}
		if tenantID := c.GetHeader(HeaderTenantID); tenantID != "" {
			c.Set("tenant_id", tenantID)
		}

		// Create context with correlation values
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, logger.RequestIDKey, requestID)
		ctx = context.WithValue(ctx, logger.CorrelationIDKey, correlationID)
		
		if traceID, exists := c.Get("trace_id"); exists {
			ctx = context.WithValue(ctx, logger.TraceIDKey, traceID)
		}
		if spanID, exists := c.Get("span_id"); exists {
			ctx = context.WithValue(ctx, logger.SpanIDKey, spanID)
		}
		if userID, exists := c.Get("user_id"); exists {
			ctx = context.WithValue(ctx, logger.UserIDKey, userID)
		}
		if tenantID, exists := c.Get("tenant_id"); exists {
			ctx = context.WithValue(ctx, logger.TenantIDKey, tenantID)
		}

		// Update request with new context
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// ExtractCorrelationID extracts correlation ID from gin context
func ExtractCorrelationID(c *gin.Context) string {
	if id, exists := c.Get("correlation_id"); exists {
		if strID, ok := id.(string); ok {
			return strID
		}
	}
	return ""
}

// ExtractRequestID extracts request ID from gin context
func ExtractRequestID(c *gin.Context) string {
	if id, exists := c.Get("request_id"); exists {
		if strID, ok := id.(string); ok {
			return strID
		}
	}
	return ""
}

// ExtractTraceID extracts trace ID from gin context
func ExtractTraceID(c *gin.Context) string {
	if id, exists := c.Get("trace_id"); exists {
		if strID, ok := id.(string); ok {
			return strID
		}
	}
	return ""
}

// ExtractUserID extracts user ID from gin context
func ExtractUserID(c *gin.Context) string {
	if id, exists := c.Get("user_id"); exists {
		if strID, ok := id.(string); ok {
			return strID
		}
	}
	return ""
}

// ContextFromGin creates a context with all correlation values from gin context
func ContextFromGin(c *gin.Context) context.Context {
	ctx := c.Request.Context()
	
	if requestID := ExtractRequestID(c); requestID != "" {
		ctx = context.WithValue(ctx, logger.RequestIDKey, requestID)
	}
	if correlationID := ExtractCorrelationID(c); correlationID != "" {
		ctx = context.WithValue(ctx, logger.CorrelationIDKey, correlationID)
	}
	if traceID := ExtractTraceID(c); traceID != "" {
		ctx = context.WithValue(ctx, logger.TraceIDKey, traceID)
	}
	if userID := ExtractUserID(c); userID != "" {
		ctx = context.WithValue(ctx, logger.UserIDKey, userID)
	}
	
	return ctx
}
