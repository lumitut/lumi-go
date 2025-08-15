# Multi-stage Dockerfile for lumi-go (Go Microservice Template)
# Optimized for security, size, and performance

# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    make \
    gcc \
    musl-dev

# Create non-root user for build
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Verify dependencies
RUN go mod verify

# Copy source code
COPY . .

# Build the binary
# CGO_ENABLED=0 for static binary
# -ldflags for smaller binary and version injection
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w \
    -X 'main.Version=${VERSION}' \
    -X 'main.BuildTime=${BUILD_TIME}' \
    -X 'main.GitCommit=${GIT_COMMIT}'" \
    -a -installsuffix cgo \
    -o service \
    ./cmd/server

# Security scan stage (optional)
FROM aquasec/trivy:latest AS security
COPY --from=builder /build/service /service
RUN trivy fs --no-progress --security-checks vuln --exit-code 1 --severity HIGH,CRITICAL /service

# Final stage - distroless for minimal attack surface
FROM gcr.io/distroless/static:nonroot

# Copy timezone data for proper time handling
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy SSL certificates for HTTPS calls
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary from builder
COPY --from=builder /build/service /service

# Copy any static files or migrations if needed
# COPY --from=builder /build/migrations /migrations
# COPY --from=builder /build/static /static

# Use non-root user (distroless nonroot user has UID 65532)
USER nonroot:nonroot

# Expose ports
# HTTP API
EXPOSE 8080
# gRPC/Connect RPC
EXPOSE 8081
# Prometheus metrics
EXPOSE 9090

# Health check (requires Docker 17.05+)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/service", "health"]

# Set entrypoint
ENTRYPOINT ["/service"]

# Default command
CMD ["serve"]
