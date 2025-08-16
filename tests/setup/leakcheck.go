// Package setup provides shared test setup utilities
package setup

import (
	"testing"

	"go.uber.org/goleak"
)

// VerifyNoLeaks checks for goroutine leaks after tests
func VerifyNoLeaks(m *testing.M) {
	goleak.VerifyTestMain(m,
		// Ignore known goroutines
		goleak.IgnoreTopFunction("github.com/gin-gonic/gin.(*Engine).Run"),
		goleak.IgnoreTopFunction("net/http.(*persistConn).writeLoop"),
		goleak.IgnoreTopFunction("net/http.(*persistConn).readLoop"),
		goleak.IgnoreTopFunction("github.com/prometheus/client_golang/prometheus.(*Registry).Gather"),
		goleak.IgnoreTopFunction("go.opentelemetry.io/otel/sdk/trace.(*batchSpanProcessor).processQueue"),
		goleak.IgnoreTopFunction("go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc.(*Exporter).Export"),
		goleak.IgnoreTopFunction("google.golang.org/grpc.(*ClientConn).WaitForStateChange"),
		goleak.IgnoreTopFunction("google.golang.org/grpc.(*ccBalancerWrapper).watcher"),
		goleak.IgnoreTopFunction("google.golang.org/grpc/internal/transport.(*controlBuffer).get"),
		goleak.IgnoreTopFunction("google.golang.org/grpc/internal/transport.(*http2Client).keepalive"),
		goleak.IgnoreTopFunction("database/sql.(*DB).connectionOpener"),
		goleak.IgnoreTopFunction("github.com/go-redis/redis/v8.(*Pool).reaper"),
	)
}

// CheckLeaks verifies no goroutine leaks in a specific test
func CheckLeaks(t *testing.T) {
	t.Helper()
	defer goleak.VerifyNone(t,
		goleak.IgnoreTopFunction("github.com/gin-gonic/gin.(*Engine).Run"),
		goleak.IgnoreTopFunction("net/http.(*persistConn).writeLoop"),
		goleak.IgnoreTopFunction("net/http.(*persistConn).readLoop"),
		goleak.IgnoreTopFunction("github.com/prometheus/client_golang/prometheus.(*Registry).Gather"),
		goleak.IgnoreTopFunction("go.opentelemetry.io/otel/sdk/trace.(*batchSpanProcessor).processQueue"),
	)
}
