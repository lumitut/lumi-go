#!/bin/bash
# Verification script for lumi-go toolchain
# Checks that all required tools are installed with correct versions

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  lumi-go Toolchain Verification${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Counters for summary
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNING_CHECKS=0

# Function to check tool version
check_tool() {
    local tool=$1
    local version_cmd=$2
    local min_version=$3
    local required=${4:-true}

    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

    if command -v $tool &> /dev/null; then
        version=$($version_cmd 2>&1 | head -n 1)
        echo -e "${GREEN}✓${NC} $tool: $version"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
    else
        if [ "$required" = true ]; then
            echo -e "${RED}✗${NC} $tool: NOT INSTALLED (minimum: $min_version) ${RED}[REQUIRED]${NC}"
            FAILED_CHECKS=$((FAILED_CHECKS + 1))
        else
            echo -e "${YELLOW}⚠${NC} $tool: NOT INSTALLED (minimum: $min_version) ${YELLOW}[OPTIONAL]${NC}"
            WARNING_CHECKS=$((WARNING_CHECKS + 1))
        fi
    fi
}

# Function to check Go tool
check_go_tool() {
    local tool=$1
    local required=${2:-true}

    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

    if command -v $tool &> /dev/null; then
        echo -e "${GREEN}✓${NC} $tool: installed"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
    else
        if [ "$required" = true ]; then
            echo -e "${RED}✗${NC} $tool: NOT INSTALLED ${RED}[REQUIRED]${NC}"
            FAILED_CHECKS=$((FAILED_CHECKS + 1))
        else
            echo -e "${YELLOW}⚠${NC} $tool: NOT INSTALLED ${YELLOW}[OPTIONAL]${NC}"
            WARNING_CHECKS=$((WARNING_CHECKS + 1))
        fi
    fi
}

# Core Requirements
echo -e "${BLUE}Core Requirements:${NC}"
check_tool "go" "go version" "1.21"
check_tool "docker" "docker --version" "20.10"
check_tool "docker-compose" "docker-compose --version" "2.0"
check_tool "make" "make --version" "3.81"
check_tool "git" "git --version" "2.25"
echo ""

# Kubernetes & Deployment
echo -e "${BLUE}Kubernetes & Deployment:${NC}"
check_tool "kubectl" "kubectl version --client --short 2>/dev/null || kubectl version --client" "1.24"
check_tool "helm" "helm version --short" "3.10"
check_tool "kind" "kind version" "0.17" false
check_tool "k9s" "k9s version --short" "0.27" false
echo ""

# Code Generation & Build Tools
echo -e "${BLUE}Code Generation & Build Tools:${NC}"
check_tool "buf" "buf --version" "1.28"
check_tool "protoc" "protoc --version" "3.21" false
check_tool "sqlc" "sqlc version" "1.25"
check_tool "migrate" "migrate -version" "4.16"
echo ""

# Code Quality & Security
echo -e "${BLUE}Code Quality & Security:${NC}"
check_tool "golangci-lint" "golangci-lint version" "1.55"
check_tool "gitleaks" "gitleaks version" "8.18"
check_tool "trivy" "trivy --version" "0.48"
echo ""

# Development Tools
echo -e "${BLUE}Development Tools:${NC}"
check_tool "air" "air -v" "1.49"
check_tool "evans" "evans --version" "0.10" false
check_tool "grpcurl" "grpcurl -version" "1.8" false
check_tool "jq" "jq --version" "1.6" false
check_tool "yq" "yq --version" "4.0" false
check_tool "httpie" "http --version" "3.2" false
echo ""

# Go Tools
echo -e "${BLUE}Go Tools:${NC}"
check_go_tool "wire"
check_go_tool "mockery"
check_go_tool "dlv"
check_go_tool "gofumpt"
check_go_tool "goimports"
check_go_tool "gosec"
check_go_tool "govulncheck"
echo ""

# Environment Variables
echo -e "${BLUE}Environment Variables:${NC}"
if [ -n "$GOPATH" ]; then
    echo -e "${GREEN}✓${NC} GOPATH: $GOPATH"
else
    echo -e "${YELLOW}⚠${NC} GOPATH: NOT SET (Go will use default)"
fi

if [ -n "$GOROOT" ]; then
    echo -e "${GREEN}✓${NC} GOROOT: $GOROOT"
else
    echo -e "${YELLOW}⚠${NC} GOROOT: NOT SET (Go will use default)"
fi

if [ "$DOCKER_BUILDKIT" = "1" ]; then
    echo -e "${GREEN}✓${NC} DOCKER_BUILDKIT: $DOCKER_BUILDKIT"
else
    echo -e "${YELLOW}⚠${NC} DOCKER_BUILDKIT: NOT SET (recommended for faster builds)"
fi

if [ "$COMPOSE_DOCKER_CLI_BUILD" = "1" ]; then
    echo -e "${GREEN}✓${NC} COMPOSE_DOCKER_CLI_BUILD: $COMPOSE_DOCKER_CLI_BUILD"
else
    echo -e "${YELLOW}⚠${NC} COMPOSE_DOCKER_CLI_BUILD: NOT SET (recommended for buildkit with compose)"
fi
echo ""

# Docker Daemon Check
echo -e "${BLUE}Docker Status:${NC}"
if docker info &> /dev/null; then
    echo -e "${GREEN}✓${NC} Docker daemon is running"

    # Check Docker Compose
    if docker-compose version &> /dev/null; then
        echo -e "${GREEN}✓${NC} Docker Compose is working"
    else
        echo -e "${RED}✗${NC} Docker Compose is not working properly"
    fi
else
    echo -e "${RED}✗${NC} Docker daemon is not running"
    echo -e "${YELLOW}  → Start Docker Desktop or Docker daemon${NC}"
fi
echo ""

# Go Module Check
echo -e "${BLUE}Go Modules:${NC}"
if [ -f "go.mod" ]; then
    echo -e "${GREEN}✓${NC} go.mod found"

    # Check if modules are downloaded
    if [ -d "$HOME/go/pkg/mod" ] || [ -d "vendor" ]; then
        echo -e "${GREEN}✓${NC} Go modules cache exists"
    else
        echo -e "${YELLOW}⚠${NC} Go modules not downloaded yet"
        echo -e "${YELLOW}  → Run: go mod download${NC}"
    fi
else
    echo -e "${YELLOW}⚠${NC} go.mod not found in current directory"
fi
echo ""

# Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "Total checks: ${TOTAL_CHECKS}"
echo -e "${GREEN}Passed: ${PASSED_CHECKS}${NC}"
echo -e "${YELLOW}Warnings: ${WARNING_CHECKS}${NC}"
echo -e "${RED}Failed: ${FAILED_CHECKS}${NC}"
echo ""

if [ $FAILED_CHECKS -eq 0 ]; then
    echo -e "${GREEN}✓ All required tools are installed!${NC}"
    echo ""
    echo -e "${BLUE}Next steps:${NC}"
    echo "1. Run: make init"
    echo "2. Run: make up"
    echo "3. Run: make run"
    exit 0
else
    echo -e "${RED}✗ Some required tools are missing!${NC}"
    echo ""
    echo -e "${BLUE}Installation instructions:${NC}"
    echo "• macOS: See docs/engineering.md#macos-installation"
    echo "• Linux: See docs/engineering.md#linux-installation"
    echo "• Windows: See docs/engineering.md#windows-installation"
    echo ""
    echo -e "${BLUE}Quick install (macOS):${NC}"
    echo "  brew install go docker docker-compose kubectl helm buf sqlc golang-migrate golangci-lint gitleaks trivy"
    echo ""
    echo -e "${BLUE}Quick install (Go tools):${NC}"
    echo "  go install github.com/cosmtrek/air@latest"
    echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2"
    echo "  go install github.com/securego/gosec/v2/cmd/gosec@latest"
    echo "  go install golang.org/x/vuln/cmd/govulncheck@latest"
    exit 1
fi
