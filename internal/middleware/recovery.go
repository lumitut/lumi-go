// Package middleware provides HTTP middleware components
package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lumitut/lumi-go/internal/observability/logger"
	"github.com/lumitut/lumi-go/internal/observability/metrics"
	"go.uber.org/zap"
)

// RecoveryConfig provides configuration for recovery middleware
type RecoveryConfig struct {
	// EnableStackTrace enables stack trace in logs
	EnableStackTrace bool
	// StackTraceSize is the size of the stack trace to capture
	StackTraceSize int
	// PrintStack prints stack trace to stderr (useful in development)
	PrintStack bool
	// LogLevel is the level at which to log panics
	LogLevel string
	// IncludeRequest includes request details in panic logs
	IncludeRequest bool
	// CustomHandler allows custom panic handling
	CustomHandler func(c *gin.Context, err interface{})
}

// DefaultRecoveryConfig returns default recovery configuration
func DefaultRecoveryConfig() RecoveryConfig {
	return RecoveryConfig{
		EnableStackTrace: true,
		StackTraceSize:   4096,
		PrintStack:       false,
		LogLevel:         "error",
		IncludeRequest:   true,
	}
}

// Recovery creates a recovery middleware that handles panics
func Recovery() gin.HandlerFunc {
	return RecoveryWithConfig(DefaultRecoveryConfig())
}

// RecoveryWithConfig creates a recovery middleware with custom configuration
func RecoveryWithConfig(config RecoveryConfig) gin.HandlerFunc {
	if config.StackTraceSize == 0 {
		config.StackTraceSize = 4096
	}

	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				handlePanic(c, err, config)
			}
		}()
		c.Next()
	}
}

// handlePanic handles the panic recovery
func handlePanic(c *gin.Context, err interface{}, config RecoveryConfig) {
	// Check for broken connection
	var brokenPipe bool
	if ne, ok := err.(*net.OpError); ok {
		if se, ok := ne.Err.(*os.SyscallError); ok {
			if strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
				strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
				brokenPipe = true
			}
		}
	}

	// Capture stack trace
	var stack []byte
	if config.EnableStackTrace {
		stack = make([]byte, config.StackTraceSize)
		length := runtime.Stack(stack, false)
		stack = stack[:length]
	}

	// Prepare log fields
	fields := []zap.Field{
		zap.Any("error", err),
		zap.String("request_id", ExtractRequestID(c)),
		zap.String("correlation_id", ExtractCorrelationID(c)),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("ip", c.ClientIP()),
	}

	// Add stack trace if enabled
	if config.EnableStackTrace && len(stack) > 0 {
		fields = append(fields, zap.ByteString("stack", stack))
	}

	// Add request details if configured
	if config.IncludeRequest && !brokenPipe {
		if httpRequest, err := httputil.DumpRequest(c.Request, false); err == nil {
			fields = append(fields, zap.ByteString("request", httpRequest))
		}
	}

	// Log the panic
	if brokenPipe {
		// Don't log broken pipe errors at error level
		logger.Warn(c.Request.Context(), "Broken pipe error", fields...)
	} else {
		switch config.LogLevel {
		case "debug":
			logger.Debug(c.Request.Context(), "Panic recovered", fields...)
		case "info":
			logger.Info(c.Request.Context(), "Panic recovered", fields...)
		case "warn":
			logger.Warn(c.Request.Context(), "Panic recovered", fields...)
		case "error":
			logger.Error(c.Request.Context(), "Panic recovered", nil, fields...)
		case "fatal":
			logger.Fatal(c.Request.Context(), "Panic recovered", fields...)
		default:
			logger.Error(c.Request.Context(), "Panic recovered", nil, fields...)
		}
	}

	// Print stack to stderr if configured (useful for development)
	if config.PrintStack && len(stack) > 0 && !brokenPipe {
		fmt.Fprintf(os.Stderr, "[Recovery] panic recovered:\n%s\n%s\n", err, stack)
	}

	// Record panic metric
	if m := metrics.Get(); m != nil {
		m.PanicsTotal.Inc()
	}

	// Handle broken pipe specially
	if brokenPipe {
		// If the connection is dead, we can't write a status
		c.Error(err.(error))
		c.Abort()
		return
	}

	// Use custom handler if provided
	if config.CustomHandler != nil {
		config.CustomHandler(c, err)
		return
	}

	// Default response
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"error":      "internal_server_error",
		"message":    "An internal server error occurred",
		"request_id": ExtractRequestID(c),
		"timestamp":  time.Now().Unix(),
	})
}

// RecoveryJSON creates a recovery middleware that always returns JSON
func RecoveryJSON() gin.HandlerFunc {
	config := DefaultRecoveryConfig()
	config.CustomHandler = func(c *gin.Context, err interface{}) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "internal_server_error",
			"message":    "An internal server error occurred",
			"request_id": ExtractRequestID(c),
			"timestamp":  time.Now().Unix(),
		})
		c.Abort()
	}
	return RecoveryWithConfig(config)
}

// RecoveryHTML creates a recovery middleware that returns HTML error page
func RecoveryHTML() gin.HandlerFunc {
	config := DefaultRecoveryConfig()
	config.CustomHandler = func(c *gin.Context, err interface{}) {
		c.HTML(http.StatusInternalServerError, "error/500.html", gin.H{
			"error":      "Internal Server Error",
			"message":    "Something went wrong on our end. Please try again later.",
			"request_id": ExtractRequestID(c),
		})
		c.Abort()
	}
	return RecoveryWithConfig(config)
}

// DevelopmentRecovery creates a recovery middleware suitable for development
// It includes more details about the error and stack trace
func DevelopmentRecovery() gin.HandlerFunc {
	config := RecoveryConfig{
		EnableStackTrace: true,
		StackTraceSize:   8192,
		PrintStack:       true,
		LogLevel:         "error",
		IncludeRequest:   true,
		CustomHandler: func(c *gin.Context, err interface{}) {
			stack := make([]byte, 8192)
			length := runtime.Stack(stack, false)
			stack = stack[:length]

			c.JSON(http.StatusInternalServerError, gin.H{
				"error":       "internal_server_error",
				"message":     fmt.Sprintf("%v", err),
				"stack_trace": string(stack),
				"request_id":  ExtractRequestID(c),
				"method":      c.Request.Method,
				"path":        c.Request.URL.Path,
				"timestamp":   time.Now().Unix(),
			})
			c.Abort()
		},
	}
	return RecoveryWithConfig(config)
}

// RecoveryWithWriter creates a recovery middleware that writes to a custom writer
func RecoveryWithWriter(out io.Writer) gin.HandlerFunc {
	config := DefaultRecoveryConfig()
	config.CustomHandler = func(c *gin.Context, err interface{}) {
		stack := make([]byte, 4096)
		length := runtime.Stack(stack, false)
		stack = stack[:length]

		timestamp := time.Now().Format(time.RFC3339)
		fmt.Fprintf(out, "[Recovery] %s panic recovered:\n%s\n%s\n", timestamp, err, stack)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "internal_server_error",
			"message":    "An internal server error occurred",
			"request_id": ExtractRequestID(c),
			"timestamp":  time.Now().Unix(),
		})
		c.Abort()
	}
	return RecoveryWithConfig(config)
}

// CustomRecoveryWithWriter creates a recovery middleware with custom writer and formatter
func CustomRecoveryWithWriter(out io.Writer, handler func(c *gin.Context, err interface{})) gin.HandlerFunc {
	config := DefaultRecoveryConfig()
	config.CustomHandler = handler
	return RecoveryWithConfig(config)
}

// stack returns a nicely formatted stack frame
func stack(skip int) []byte {
	buf := new(bytes.Buffer)
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Skip runtime functions
		if strings.Contains(file, "runtime/") {
			continue
		}
		// Print this much at least
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// source returns a space-trimmed slice of the n'th line
func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return []byte("???")
	}
	return bytes.TrimSpace(lines[n])
}

// function returns, if possible, the name of the function containing the PC
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return []byte("???")
	}
	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included. Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	// Also the package path might contain dot (e.g. code.google.com/...),
	// so first eliminate the path prefix
	if lastSlash := bytes.LastIndex(name, []byte("/")); lastSlash >= 0 {
		name = name[lastSlash+1:]
	}
	if period := bytes.Index(name, []byte(".")); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, []byte("·"), []byte("."), -1)
	return name
}
