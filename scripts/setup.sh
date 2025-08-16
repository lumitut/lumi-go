#!/bin/bash
# Setup script for Lumi-Go development environment
#
# This script is used ONCE when setting up a new development environment.
# It installs required tools, sets up configuration, and verifies everything works.
#
# For daily development workflow (start/stop/restart service), use local.sh instead.

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print colored output
print_success() { echo -e "${GREEN}✓ $1${NC}"; }
print_error() { echo -e "${RED}✗ $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠ $1${NC}"; }
print_info() { echo -e "ℹ $1"; }

echo "================================================"
echo "    Lumi-Go Development Environment Setup"
echo "================================================"
echo ""

# Check Go installation
check_go() {
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}')
        print_success "Go installed: $GO_VERSION"

        # Check minimum version (1.22)
        REQUIRED_VERSION="go1.22"
        if [[ "$GO_VERSION" < "$REQUIRED_VERSION" ]]; then
            print_warning "Go version $REQUIRED_VERSION or higher recommended"
        fi
    else
        print_error "Go is not installed"
        echo "Please install Go from https://golang.org/dl/"
        exit 1
    fi
}

# Check Docker installation
check_docker() {
    if command -v docker &> /dev/null; then
        DOCKER_VERSION=$(docker --version | awk '{print $3}' | sed 's/,//')
        print_success "Docker installed: $DOCKER_VERSION"
    else
        print_warning "Docker is not installed (optional for containerized development)"
    fi
}

# Check Make installation
check_make() {
    if command -v make &> /dev/null; then
        print_success "Make installed"
    else
        print_warning "Make is not installed (recommended for build commands)"
    fi
}

# Install Go development tools
install_go_tools() {
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

    # mockgen for generating mocks
    if ! command -v mockgen &> /dev/null; then
        print_info "Installing mockgen..."
        go install github.com/golang/mock/mockgen@latest
        print_success "mockgen installed"
    else
        print_success "mockgen already installed"
    fi
}

# Setup environment file
setup_env() {
    if [ ! -f .env ]; then
        print_info "Creating .env file from env.example..."
        cp env.example .env
        print_success ".env file created"
        print_warning "Please review and update .env with your configuration"
    else
        print_success ".env file already exists"
    fi
}

# Download Go dependencies
download_deps() {
    print_info "Downloading Go dependencies..."
    go mod download
    go mod verify
    print_success "Dependencies downloaded and verified"
}

# Run initial build
initial_build() {
    print_info "Running initial build..."
    if command -v make &> /dev/null; then
        make build
    else
        go build -o build/server ./cmd/server
    fi
    print_success "Initial build successful"
}

# Run tests
run_tests() {
    print_info "Running tests..."
    if command -v make &> /dev/null; then
        make test
    else
        go test ./...
    fi
    print_success "Tests passed"
}

# Main setup flow
main() {
    echo "Checking prerequisites..."
    echo ""

    check_go
    check_docker
    check_make

    echo ""
    echo "Setting up development environment..."
    echo ""

    install_go_tools
    setup_env
    download_deps
    initial_build
    run_tests

    echo ""
    echo "================================================"
    echo "    Setup Complete! 🎉"
    echo "================================================"
    echo ""
    echo "Next steps:"
    echo "  1. Review and update .env configuration"
    echo "  2. Run 'make run' to start the service"
    echo "  3. Run 'make run-dev' for hot reload mode"
    echo "  4. Visit http://localhost:8080/health"
    echo ""
    echo "Useful commands:"
    echo "  make help         - Show all available commands"
    echo "  make test         - Run tests"
    echo "  make coverage     - Generate coverage report"
    echo "  make docker-build - Build Docker image"
    echo ""
    print_success "Happy coding! 🚀"
}

# Run main function
main "$@"
