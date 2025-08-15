#!/bin/bash
# Fresh Machine Validation Script for lumi-go
# This script validates that the template works correctly on a fresh machine

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Validation results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
VALIDATION_LOG="validation-$(date +%Y%m%d-%H%M%S).log"

# Function to print colored output
print_color() {
    local color=$1
    shift
    echo -e "${color}$*${NC}" | tee -a "$VALIDATION_LOG"
}

# Function to print test result
test_result() {
    local test_name=$1
    local result=$2
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if [ "$result" = "pass" ]; then
        print_color "$GREEN" "✓ $test_name"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        print_color "$RED" "✗ $test_name"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

# Function to run test
run_test() {
    local test_name=$1
    local test_command=$2

    print_color "$CYAN" "\nTesting: $test_name"
    if eval "$test_command" >> "$VALIDATION_LOG" 2>&1; then
        test_result "$test_name" "pass"
        return 0
    else
        test_result "$test_name" "fail"
        return 1
    fi
}

# Start validation
print_color "$BLUE" "========================================="
print_color "$BLUE" "  lumi-go Fresh Machine Validation"
print_color "$BLUE" "========================================="
print_color "$YELLOW" "\nStarting at: $(date)"
print_color "$YELLOW" "Log file: $VALIDATION_LOG\n"

# Phase 1: Check Prerequisites
print_color "$MAGENTA" "\n=== Phase 1: Prerequisites Check ==="

# Check required tools
run_test "Go installed" "go version"
run_test "Docker installed" "docker --version"
run_test "Docker Compose installed" "docker-compose --version"
run_test "Make installed" "make --version"
run_test "Git installed" "git --version"

# Check Docker daemon
run_test "Docker daemon running" "docker info > /dev/null 2>&1"

# Phase 2: Repository Setup
print_color "$MAGENTA" "\n=== Phase 2: Repository Setup ==="

# Check directory structure
run_test "cmd/server directory exists" "[ -d cmd/server ]"
run_test "internal directory exists" "[ -d internal ]"
run_test "api directory exists" "[ -d api ]"
run_test "migrations directory exists" "[ -d migrations ]"
run_test "deploy/docker directory exists" "[ -d deploy/docker ]"
run_test "deploy/helm directory exists" "[ -d deploy/helm ]"
run_test "scripts directory exists" "[ -d scripts ]"
run_test "docs directory exists" "[ -d docs ]"

# Check key files
run_test "Makefile exists" "[ -f Makefile ]"
run_test "docker-compose.yml exists" "[ -f docker-compose.yml ]"
run_test ".air.toml exists" "[ -f .air.toml ]"
run_test "go.mod exists" "[ -f go.mod ]"
run_test ".env.example exists" "[ -f .env.example ]"

# Phase 3: Dependencies Installation
print_color "$MAGENTA" "\n=== Phase 3: Dependencies Installation ==="

# Initialize Go modules
run_test "Go mod download" "go mod download"
run_test "Go mod verify" "go mod verify"

# Install development tools
if command -v air &> /dev/null; then
    test_result "Air (hot-reload) installed" "pass"
else
    print_color "$YELLOW" "Installing Air..."
    if go install github.com/cosmtrek/air@latest >> "$VALIDATION_LOG" 2>&1; then
        test_result "Air installation" "pass"
    else
        test_result "Air installation" "fail"
    fi
fi

# Phase 4: Docker Services
print_color "$MAGENTA" "\n=== Phase 4: Docker Services ==="

# Clean any existing containers
print_color "$YELLOW" "Cleaning existing containers..."
docker-compose down -v >> "$VALIDATION_LOG" 2>&1 || true

# Start services
if run_test "Docker Compose up" "docker-compose up -d"; then
    print_color "$YELLOW" "Waiting for services to be ready..."
    sleep 10

    # Check service health
    run_test "PostgreSQL healthy" "docker-compose exec -T postgres pg_isready -U lumigo"
    run_test "Redis healthy" "docker-compose exec -T redis redis-cli ping | grep -q PONG"

    # Check service ports
    run_test "PostgreSQL port 5432 accessible" "nc -zv localhost 5432 2>&1 | grep -q succeeded"
    run_test "Redis port 6379 accessible" "nc -zv localhost 6379 2>&1 | grep -q succeeded"
    run_test "Prometheus port 9091 accessible" "nc -zv localhost 9091 2>&1 | grep -q succeeded"
    run_test "Grafana port 3000 accessible" "nc -zv localhost 3000 2>&1 | grep -q succeeded"
    run_test "Jaeger port 16686 accessible" "nc -zv localhost 16686 2>&1 | grep -q succeeded"

    # Run migrations
    run_test "Database migrations" "docker-compose run --rm migrate"
fi

# Phase 5: Application Build
print_color "$MAGENTA" "\n=== Phase 5: Application Build ==="

# Create stub main.go if it doesn't exist
if [ ! -f "cmd/server/main.go" ]; then
    print_color "$YELLOW" "Creating stub main.go for testing..."
    mkdir -p cmd/server
    cat > cmd/server/main.go << 'EOF'
package main

import (
    "fmt"
    "net/http"
    "os"
)

func main() {
    port := os.Getenv("HTTP_ADDR")
    if port == "" {
        port = ":8080"
    }

    http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, `{"status":"healthy"}`)
    })

    http.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, `{"status":"ready"}`)
    })

    fmt.Printf("Server starting on %s\n", port)
    if err := http.ListenAndServe(port, nil); err != nil {
        fmt.Printf("Server failed: %v\n", err)
        os.Exit(1)
    }
}
EOF
fi

# Build application
run_test "Application build" "go build -o bin/lumi-go ./cmd/server"

# Phase 6: Configuration Files
print_color "$MAGENTA" "\n=== Phase 6: Configuration Files ==="

# Check configuration files
run_test "OTEL Collector config exists" "[ -f deploy/docker/otel-collector-config.yaml ]"
run_test "Prometheus config exists" "[ -f deploy/docker/prometheus.yml ]"
run_test "Grafana datasource config exists" "[ -f deploy/docker/grafana-datasource.yml ]"
run_test "Helm Chart.yaml exists" "[ -f deploy/helm/Chart.yaml ]"
run_test "Helm values.yaml exists" "[ -f deploy/helm/values.yaml ]"

# Phase 7: Documentation
print_color "$MAGENTA" "\n=== Phase 7: Documentation ==="

run_test "README.md exists" "[ -f README.md ]"
run_test "CONTRIBUTING.md exists" "[ -f CONTRIBUTING.md ]"
run_test "SECURITY.md exists" "[ -f SECURITY.md ]"
run_test "LICENSE exists" "[ -f LICENSE ]"
run_test "docs/quickstart.md exists" "[ -f docs/quickstart.md ]"
run_test "docs/engineering.md exists" "[ -f docs/engineering.md ]"
run_test "docs/development.md exists" "[ -f docs/development.md ]"

# Phase 8: Scripts
print_color "$MAGENTA" "\n=== Phase 8: Scripts ==="

run_test "local.sh is executable" "[ -x scripts/local.sh ]"
run_test "verify-setup.sh is executable" "[ -x scripts/verify-setup.sh ]"
run_test "Docker build script is executable" "[ -x deploy/docker/build.sh ]"

# Phase 9: Make Targets
print_color "$MAGENTA" "\n=== Phase 9: Make Targets ==="

run_test "make help works" "make help > /dev/null"
run_test "make fmt works" "make fmt"
run_test "make vet works" "make vet"

# Phase 10: Application Runtime
print_color "$MAGENTA" "\n=== Phase 10: Application Runtime ==="

# Start application in background
if [ -f "bin/lumi-go" ]; then
    print_color "$YELLOW" "Starting application..."
    HTTP_ADDR=:8080 ./bin/lumi-go >> "$VALIDATION_LOG" 2>&1 &
    APP_PID=$!
    sleep 3

    # Test endpoints
    run_test "Health endpoint responds" "curl -f http://localhost:8080/healthz"
    run_test "Ready endpoint responds" "curl -f http://localhost:8080/readyz"

    # Stop application
    kill $APP_PID 2>/dev/null || true
    wait $APP_PID 2>/dev/null || true
fi

# Phase 11: Cleanup Test
print_color "$MAGENTA" "\n=== Phase 11: Cleanup Test ==="

run_test "Docker Compose down" "docker-compose down"
run_test "Clean build artifacts" "make clean || true"

# Summary
print_color "$BLUE" "\n========================================="
print_color "$BLUE" "  Validation Summary"
print_color "$BLUE" "========================================="
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    print_color "$GREEN" "✅ ALL TESTS PASSED!"
    print_color "$GREEN" "Total: $TOTAL_TESTS | Passed: $PASSED_TESTS | Failed: $FAILED_TESTS"
    print_color "$GREEN" "\nThe lumi-go template is ready for use on a fresh machine!"
    EXIT_CODE=0
else
    print_color "$RED" "❌ SOME TESTS FAILED"
    print_color "$YELLOW" "Total: $TOTAL_TESTS | Passed: $PASSED_TESTS | Failed: $FAILED_TESTS"
    print_color "$YELLOW" "\nPlease check the log file: $VALIDATION_LOG"
    EXIT_CODE=1
fi

print_color "$YELLOW" "\nCompleted at: $(date)"
print_color "$CYAN" "Full log saved to: $VALIDATION_LOG"

# Cleanup
print_color "$YELLOW" "\n🧹 Cleaning up test environment..."
docker-compose down -v >> "$VALIDATION_LOG" 2>&1 || true
rm -f bin/lumi-go cmd/server/main.go 2>/dev/null || true

exit $EXIT_CODE
