// Package httpapi provides HTTP server setup and routing
package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lumitut/lumi-go/internal/config"
	"github.com/lumitut/lumi-go/internal/middleware"
	"github.com/lumitut/lumi-go/internal/observability/logger"
	"github.com/lumitut/lumi-go/internal/observability/metrics"
	"go.uber.org/zap"
)

// Server represents the HTTP server
type Server struct {
	config     *config.Config
	router     *gin.Engine
	httpServer *http.Server
	isReady    bool
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.Config) *Server {
	// Set Gin mode based on environment
	if cfg.Service.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := setupRouter(cfg)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         ":" + cfg.Server.HTTPPort,
		Handler:      router,
		ReadTimeout:  cfg.Server.HTTPReadTimeout,
		WriteTimeout: cfg.Server.HTTPWriteTimeout,
		IdleTimeout:  cfg.Server.HTTPIdleTimeout,
	}

	return &Server{
		config:     cfg,
		router:     router,
		httpServer: httpServer,
		isReady:    false,
	}
}

// setupRouter configures the Gin router with all middleware and routes
func setupRouter(cfg *config.Config) *gin.Engine {
	// Create router without default middleware
	router := gin.New()

	// Global middleware - Order matters!

	// 1. Recovery (must be first to catch panics)
	router.Use(middleware.RecoveryWithConfig(middleware.RecoveryConfig{
		EnableStackTrace: cfg.Middleware.RecoveryStackTrace,
		StackTraceSize:   cfg.Middleware.RecoveryStackSize,
		PrintStack:       cfg.Middleware.RecoveryPrintStack,
		LogLevel:         "error",
		IncludeRequest:   true,
	}))

	// 2. Real IP extraction (before anything that needs client IP)
	if cfg.Middleware.TrustAllProxies {
		router.Use(middleware.RealIP())
	} else if len(cfg.Middleware.TrustedProxies) > 0 {
		router.Use(middleware.RealIPWithConfig(middleware.RealIPConfig{
			TrustedProxies: cfg.Middleware.TrustedProxies,
			TrustAll:       false,
		}))
	} else {
		router.Use(middleware.RealIP())
	}

	// 3. Correlation IDs (before logging/tracing)
	router.Use(middleware.Correlation())

	// 4. OpenTelemetry tracing
	if cfg.Observability.TracingEnabled {
		router.Use(middleware.TracingWithConfig(middleware.TracingConfig{
			ServiceName:   cfg.Service.Name,
			SkipPaths:     cfg.Middleware.LogSkipPaths,
			RecordError:   true,
			RecordHeaders: false,
		}))
	}

	// 5. Access logging
	router.Use(middleware.LoggingWithConfig(middleware.LoggingConfig{
		SkipPaths:       cfg.Middleware.LogSkipPaths,
		LogRequestBody:  cfg.Middleware.LogRequestBody,
		LogResponseBody: cfg.Middleware.LogResponseBody,
		SlowThreshold:   cfg.Middleware.LogSlowThreshold,
	}))

	// 6. Metrics
	if cfg.Observability.MetricsEnabled {
		router.Use(middleware.MetricsWithConfig(middleware.MetricsConfig{
			SkipPaths: cfg.Middleware.LogSkipPaths,
		}))
	}

	// 7. Rate limiting
	if cfg.Middleware.RateLimitEnabled {
		var rateLimitMiddleware gin.HandlerFunc
		switch cfg.Middleware.RateLimitType {
		case "user":
			rateLimitMiddleware = middleware.UserRateLimit(cfg.Middleware.RateLimitRate)
		case "api_key":
			rateLimitMiddleware = middleware.APIKeyRateLimit(cfg.Middleware.RateLimitRate)
		default: // "ip"
			rateLimitMiddleware = middleware.IPRateLimit(cfg.Middleware.RateLimitRate)
		}
		router.Use(rateLimitMiddleware)
	}

	// 8. CORS (if enabled)
	if cfg.Middleware.CORSEnabled {
		corsConfig := middleware.CORSConfig{
			Enabled:          true,
			AllowOrigins:     cfg.Middleware.CORSAllowOrigins,
			AllowMethods:     cfg.Middleware.CORSAllowMethods,
			AllowHeaders:     cfg.Middleware.CORSAllowHeaders,
			ExposeHeaders:    cfg.Middleware.CORSExposeHeaders,
			AllowCredentials: cfg.Middleware.CORSAllowCredentials,
			MaxAge:           cfg.Middleware.CORSMaxAge,
			AllowWildcard:    cfg.Service.Environment == "development",
		}

		// Use development config if in development and no origins specified
		if cfg.Service.Environment == "development" && len(cfg.Middleware.CORSAllowOrigins) == 0 {
			corsConfig = middleware.DevelopmentCORSConfig()
		}

		router.Use(middleware.CORS(corsConfig))
	}

	// Register routes
	registerOpsRoutes(router, cfg)
	registerAPIRoutes(router, cfg)

	return router
}

// registerOpsRoutes registers operational endpoints
func registerOpsRoutes(router *gin.Engine, cfg *config.Config) {
	// Health check - always returns 200 if service is running
	router.GET("/healthz", handleHealth)
	router.GET("/health", handleHealth)

	// Readiness check - returns 200 when service is ready to handle requests
	router.GET("/readyz", func(c *gin.Context) {
		// Access the server instance through context if needed
		// For now, just check basic readiness
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
			"time":   time.Now().Unix(),
		})
	})
	router.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
			"time":   time.Now().Unix(),
		})
	})

	// Metrics endpoint
	if cfg.Observability.MetricsEnabled {
		router.GET("/metrics", gin.WrapH(metrics.Handler()))
	}

	// pprof endpoints (only in non-production or if explicitly enabled)
	if cfg.Service.Environment != "production" || cfg.Server.EnablePProf {
		pprofGroup := router.Group("/debug/pprof")
		{
			pprofGroup.GET("/", gin.WrapF(pprof.Index))
			pprofGroup.GET("/cmdline", gin.WrapF(pprof.Cmdline))
			pprofGroup.GET("/profile", gin.WrapF(pprof.Profile))
			pprofGroup.GET("/symbol", gin.WrapF(pprof.Symbol))
			pprofGroup.GET("/trace", gin.WrapF(pprof.Trace))
			pprofGroup.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
			pprofGroup.GET("/block", gin.WrapH(pprof.Handler("block")))
			pprofGroup.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
			pprofGroup.GET("/heap", gin.WrapH(pprof.Handler("heap")))
			pprofGroup.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
			pprofGroup.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
		}
	}

	// Version endpoint
	router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":     cfg.Service.Name,
			"version":     cfg.Service.Version,
			"environment": cfg.Service.Environment,
		})
	})
}

// registerAPIRoutes registers application API routes
func registerAPIRoutes(router *gin.Engine, cfg *config.Config) {
	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Example user routes
		users := v1.Group("/users")
		{
			users.GET("/:id", handleGetUser)
			users.POST("", handleCreateUser)
			users.PUT("/:id", handleUpdateUser)
			users.DELETE("/:id", handleDeleteUser)
			users.GET("", handleListUsers)
		}

		// Add more API groups here
	}
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	logger.Info(ctx, "Starting HTTP server",
		zap.String("address", s.httpServer.Addr),
		zap.String("environment", s.config.Service.Environment),
	)

	// Mark server as ready after a brief initialization
	go func() {
		time.Sleep(100 * time.Millisecond)
		s.setReady(true)
		logger.Info(ctx, "HTTP server ready to accept requests")
	}()

	// Start server
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	logger.Info(ctx, "Shutting down HTTP server")

	// Mark as not ready
	s.setReady(false)

	// Wait a bit for load balancers to detect
	time.Sleep(5 * time.Second)

	// Shutdown server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	logger.Info(ctx, "HTTP server shutdown complete")
	return nil
}

// Router returns the Gin router
func (s *Server) Router() *gin.Engine {
	return s.router
}

// setReady sets the readiness state
func (s *Server) setReady(ready bool) {
	s.isReady = ready
}

// IsReady returns the readiness state
func (s *Server) IsReady() bool {
	return s.isReady
}

// Health check handler
func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now().Unix(),
	})
}

// Example API handlers (to be moved to separate handlers package)
func handleGetUser(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.Param("id")

	logger.Info(ctx, "Fetching user",
		zap.String("user_id", userID),
	)

	// Example response
	c.JSON(http.StatusOK, gin.H{
		"id":         userID,
		"username":   "john_doe",
		"email":      "john@example.com",
		"created_at": time.Now().Unix(),
	})
}

func handleCreateUser(c *gin.Context) {
	ctx := c.Request.Context()

	var user struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&user); err != nil {
		logger.Warn(ctx, "Invalid user data",
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	logger.Info(ctx, "Creating user",
		zap.String("username", user.Username),
	)

	c.JSON(http.StatusCreated, gin.H{
		"id":       "user_123",
		"username": user.Username,
		"email":    user.Email,
	})
}

func handleUpdateUser(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.Param("id")

	logger.Info(ctx, "Updating user",
		zap.String("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{
		"id":         userID,
		"updated_at": time.Now().Unix(),
	})
}

func handleDeleteUser(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.Param("id")

	logger.Info(ctx, "Deleting user",
		zap.String("user_id", userID),
	)

	c.Status(http.StatusNoContent)
}

func handleListUsers(c *gin.Context) {
	ctx := c.Request.Context()

	logger.Info(ctx, "Listing users")

	c.JSON(http.StatusOK, gin.H{
		"users": []gin.H{
			{
				"id":       "user_1",
				"username": "john_doe",
				"email":    "john@example.com",
			},
			{
				"id":       "user_2",
				"username": "jane_doe",
				"email":    "jane@example.com",
			},
		},
		"total": 2,
	})
}
