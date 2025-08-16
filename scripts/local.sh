#!/bin/bash
# Development helper script for Lumi-Go
#
# This script provides all development tools in one place:
# - One-time environment setup: ./scripts/local.sh setup
# - Daily service management: start/stop/restart/status/logs
# - Testing and cleanup: test/clean
#
# Run './scripts/local.sh help' to see all available commands.

set -e

# Configuration
SERVICE_NAME="lumi-go"
HTTP_PORT="${LUMI_SERVER_HTTPPORT:-8080}"
RPC_PORT="${LUMI_SERVER_RPCPORT:-8081}"
METRICS_PORT="${LUMI_OBSERVABILITY_METRICSPORT:-9090}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Functions
print_header() {
    echo -e "${BLUE}================================================${NC}"
    echo -e "${BLUE}    $1${NC}"
    echo -e "${BLUE}================================================${NC}"
}

print_success() { echo -e "${GREEN}✓ $1${NC}"; }
print_error() { echo -e "${RED}✗ $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠ $1${NC}"; }
print_info() { echo -e "ℹ $1"; }

# Check if service is running
check_service() {
    if curl -f -s http://localhost:$HTTP_PORT/health > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Start the service
start_service() {
    print_header "Starting $SERVICE_NAME"

    if check_service; then
        print_warning "Service is already running on port $HTTP_PORT"
        return 0
    fi

    print_info "Starting service..."

    # Load environment variables
    if [ -f .env ]; then
        export $(grep -v '^#' .env | xargs)
    fi

    # Build the service
    print_info "Building service..."
    go build -o build/server ./cmd/server

    # Start in background
    nohup ./build/server > logs/service.log 2>&1 &
    echo $! > .pid

    # Wait for service to be ready
    print_info "Waiting for service to be ready..."
    for i in {1..30}; do
        if check_service; then
            print_success "Service started successfully!"
            print_info "HTTP API: http://localhost:$HTTP_PORT"
            print_info "gRPC API: localhost:$RPC_PORT"
            print_info "Metrics: http://localhost:$METRICS_PORT/metrics"
            print_info "Logs: tail -f logs/service.log"
            return 0
        fi
        sleep 1
    done

    print_error "Service failed to start"
    return 1
}

# Stop the service
stop_service() {
    print_header "Stopping $SERVICE_NAME"

    if [ -f .pid ]; then
        PID=$(cat .pid)
        if ps -p $PID > /dev/null; then
            print_info "Stopping service (PID: $PID)..."
            kill $PID
            rm .pid
            print_success "Service stopped"
        else
            print_warning "Service not running (stale PID file)"
            rm .pid
        fi
    else
        print_warning "No PID file found"
    fi
}

# Restart the service
restart_service() {
    stop_service
    sleep 2
    start_service
}

# Show service status
status_service() {
    print_header "Service Status"

    if check_service; then
        print_success "Service is running"

        # Get health status
        HEALTH=$(curl -s http://localhost:$HTTP_PORT/health)
        print_info "Health: $HEALTH"

        # Get readiness status
        READY=$(curl -s http://localhost:$HTTP_PORT/ready)
        print_info "Readiness: $READY"

        # Show process info if PID file exists
        if [ -f .pid ]; then
            PID=$(cat .pid)
            print_info "Process ID: $PID"
            ps -p $PID -o pid,vsz,rss,comm
        fi
    else
        print_error "Service is not running"
    fi
}

# Tail logs
tail_logs() {
    print_header "Service Logs"

    if [ ! -f logs/service.log ]; then
        print_warning "No log file found"
        return 1
    fi

    tail -f logs/service.log
}

# Run tests
run_tests() {
    print_header "Running Tests"

    if command -v make &> /dev/null; then
        make test
    else
        print_info "Running unit tests..."
        go test -v -race ./tests/unit/...

        print_info "Running integration tests..."
        go test -v -race -tags=integration ./tests/integration/...
    fi

    print_success "All tests passed!"
}

# Clean up
cleanup() {
    print_header "Cleaning Up"

    # Stop service if running
    stop_service

    # Clean build artifacts
    print_info "Cleaning build artifacts..."
    rm -rf build/
    rm -rf tmp/
    rm -f .pid

    # Clean test cache
    print_info "Cleaning test cache..."
    go clean -testcache

    print_success "Cleanup complete"
}

# Setup development environment (one-time)
setup_environment() {
    print_header "Setting up Development Environment"

    # Check prerequisites
    print_info "Checking prerequisites..."

    # Check Go
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}')
        print_success "Go installed: $GO_VERSION"

        REQUIRED_VERSION="go1.22"
        if [[ "$GO_VERSION" < "$REQUIRED_VERSION" ]]; then
            print_warning "Go version $REQUIRED_VERSION or higher recommended"
        fi
    else
        print_error "Go is not installed"
        echo "Please install Go from https://golang.org/dl/"
        return 1
    fi

    # Check Docker
    if command -v docker &> /dev/null; then
        DOCKER_VERSION=$(docker --version | awk '{print $3}' | sed 's/,//')
        print_success "Docker installed: $DOCKER_VERSION"
    else
        print_warning "Docker is not installed (optional for containerized development)"
    fi

    # Check Make
    if command -v make &> /dev/null; then
        print_success "Make installed"
    else
        print_warning "Make is not installed (recommended for build commands)"
    fi

    # Install Go development tools
    print_info "Installing Go development tools..."

    # Air for hot reload
    if ! command -v air &> /dev/null; then
        print_info "Installing air (hot reload)..."
        go install github.com/air-verse/air@latest
        print_success "Air installed"
    else
        print_success "Air already installed"
    fi

    # golangci-lint
    if ! command -v golangci-lint &> /dev/null; then
        print_info "Installing golangci-lint..."
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
        print_success "golangci-lint installed"
    else
        print_success "golangci-lint already installed"
    fi

    # goimports
    if ! command -v goimports &> /dev/null; then
        print_info "Installing goimports..."
        go install golang.org/x/tools/cmd/goimports@latest
        print_success "goimports installed"
    else
        print_success "goimports already installed"
    fi

    # mockgen
    if ! command -v mockgen &> /dev/null; then
        print_info "Installing mockgen..."
        go install github.com/golang/mock/mockgen@latest
        print_success "mockgen installed"
    else
        print_success "mockgen already installed"
    fi

    # Setup environment file
    if [ ! -f .env ]; then
        print_info "Creating .env file from env.example..."
        cp env.example .env
        print_success ".env file created"
        print_warning "Please review and update .env with your configuration"
    else
        print_success ".env file already exists"
    fi

    # Download dependencies
    print_info "Downloading Go dependencies..."
    go mod download
    go mod verify
    print_success "Dependencies downloaded and verified"

    # Initial build
    print_info "Running initial build..."
    if command -v make &> /dev/null; then
        make build
    else
        mkdir -p build
        go build -o build/server ./cmd/server
    fi
    print_success "Initial build successful"

    # Run tests
    print_info "Running tests..."
    if command -v make &> /dev/null; then
        make test
    else
        go test ./...
    fi
    print_success "Tests passed"

    print_success "Development environment setup complete! 🎉"
    echo ""
    echo "Next steps:"
    echo "  1. Review and update .env configuration"
    echo "  2. Run './scripts/local.sh start' to start the service"
    echo "  3. Visit http://localhost:$HTTP_PORT/health"
    echo ""
    echo "Useful commands:"
    echo "  ./scripts/local.sh start   - Start the service"
    echo "  ./scripts/local.sh status  - Check service status"
    echo "  ./scripts/local.sh logs    - View service logs"
    echo "  ./scripts/local.sh help    - Show all commands"
}

# Show usage
usage() {
    echo "Usage: $0 {setup|start|stop|restart|status|logs|test|clean|help}"
    echo ""
    echo "Commands:"
    echo "  setup    - Setup development environment (run once)"
    echo "  start    - Start the service"
    echo "  stop     - Stop the service"
    echo "  restart  - Restart the service"
    echo "  status   - Show service status"
    echo "  logs     - Tail service logs"
    echo "  test     - Run tests"
    echo "  clean    - Clean up artifacts and stop service"
    echo "  help     - Show this help message"
}

# Create necessary directories
mkdir -p logs build

# Main script
case "$1" in
    setup)
        setup_environment
        ;;
    start)
        start_service
        ;;
    stop)
        stop_service
        ;;
    restart)
        restart_service
        ;;
    status)
        status_service
        ;;
    logs)
        tail_logs
        ;;
    test)
        run_tests
        ;;
    clean)
        cleanup
        ;;
    help|--help|-h)
        usage
        ;;
    *)
        usage
        exit 1
        ;;
esac
