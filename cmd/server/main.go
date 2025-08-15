package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lumitut/lumi-go/internal/middleware"
	"github.com/lumitut/lumi-go/internal/observability/logger"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logConfig := logger.Config{
		Level:             getEnv("LOG_LEVEL", "info"),
		Format:            getEnv("LOG_FORMAT", "json"),
		Development:       getEnv("ENVIRONMENT", "development") == "development",
		DisableCaller:     false,
		DisableStacktrace: false,
		SampleInitial:     100,
		SampleThereafter:  100,
	}

	if err := logger.Initialize(logConfig); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Log startup
	logger.Info(context.Background(), "Starting lumi-go service",
		zap.String("version", getEnv("SERVICE_VERSION", "unknown")),
		zap.String("environment", getEnv("ENVIRONMENT", "development")),
	)

	// Setup Gin
	if getEnv("ENVIRONMENT", "development") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := setupRouter()

	// Setup server
	srv := &http.Server{
		Addr:         ":" + getEnv("PORT", "8080"),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info(context.Background(), "HTTP server starting",
			zap.String("address", srv.Addr),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(context.Background(), "Failed to start server",
				zap.Error(err),
			)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info(context.Background(), "Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error(context.Background(), "Server forced to shutdown", err)
		os.Exit(1)
	}

	logger.Info(context.Background(), "Server shutdown complete")
}

func setupRouter() *gin.Engine {
	router := gin.New()

	// Recovery middleware
	router.Use(gin.Recovery())

	// Correlation middleware (must be first)
	router.Use(middleware.Correlation())

	// Logging middleware
	router.Use(middleware.Logging(
		"/health",
		"/ready",
		"/metrics",
	))

	// Health check endpoints
	router.GET("/health", handleHealth)
	router.GET("/ready", handleReady)

	// API routes
	v1 := router.Group("/api/v1")
	{
		v1.GET("/users/:id", handleGetUser)
		v1.POST("/users", handleCreateUser)
		v1.PUT("/users/:id", handleUpdateUser)
		v1.DELETE("/users/:id", handleDeleteUser)
	}

	return router
}

// Health check handlers
func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now().Unix(),
	})
}

func handleReady(c *gin.Context) {
	// TODO: Check database, cache, etc.
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"time":   time.Now().Unix(),
	})
}

// Example API handlers
func handleGetUser(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.Param("id")

	logger.Info(ctx, "Fetching user",
		zap.String("user_id", userID),
	)

	// Simulate some processing
	time.Sleep(50 * time.Millisecond)

	// Example response
	c.JSON(http.StatusOK, gin.H{
		"id":         userID,
		"username":   "john_doe",
		"email":      "[REDACTED]", // PII redacted
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

	// Log audit event
	logger.Audit(ctx, "USER_CREATED", "user:new", "success",
		zap.String("username", user.Username),
	)

	c.JSON(http.StatusCreated, gin.H{
		"id":       "user_123",
		"username": user.Username,
		"email":    "[REDACTED]",
	})
}

func handleUpdateUser(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.Param("id")

	logger.Info(ctx, "Updating user",
		zap.String("user_id", userID),
	)

	// Simulate update
	logger.Audit(ctx, "USER_UPDATED", fmt.Sprintf("user:%s", userID), "success",
		zap.String("updated_fields", "email,profile"),
	)

	c.JSON(http.StatusOK, gin.H{
		"id":         userID,
		"updated_at": time.Now().Unix(),
	})
}

func handleDeleteUser(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.Param("id")

	start := time.Now()

	// Simulate deletion
	time.Sleep(100 * time.Millisecond)

	// Log performance
	logger.Performance(ctx, "user_deletion", time.Since(start),
		zap.String("user_id", userID),
	)

	// Log audit event
	logger.Audit(ctx, "USER_DELETED", fmt.Sprintf("user:%s", userID), "success",
		zap.String("reason", "api_request"),
	)

	c.JSON(http.StatusNoContent, nil)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
