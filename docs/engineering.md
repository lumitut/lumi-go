# Engineering Setup Guide

## Overview

This document provides the complete engineering setup for the lumi-go (Go Microservice Template) project, including all required toolchain versions, installation instructions, and configuration steps.

## Required Toolchain Versions

### Core Requirements

| Tool | Minimum Version | Recommended Version | Purpose |
|------|----------------|-------------------|---------|
| **Go** | 1.21 | 1.22+ | Primary programming language |
| **Docker** | 20.10 | 24.0+ | Container runtime |
| **Docker Compose** | 2.0 | 2.23+ | Local orchestration |
| **Make** | 3.81 | 4.3+ | Build automation |
| **Git** | 2.25 | 2.40+ | Version control |

### Kubernetes & Deployment

| Tool | Minimum Version | Recommended Version | Purpose |
|------|----------------|-------------------|---------|
| **kubectl** | 1.24 | 1.28+ | Kubernetes CLI |
| **Helm** | 3.10 | 3.13+ | Kubernetes package manager |
| **kind** | 0.17 | 0.20+ | Local Kubernetes (optional) |
| **k9s** | 0.27 | 0.28+ | Kubernetes TUI (optional) |

### Code Generation & Build Tools

| Tool | Version | Purpose |
|------|---------|---------|
| **buf** | 1.28.1 | Protocol buffer compiler |
| **protoc** | 3.21+ | Protocol buffer compiler (alternative) |
| **sqlc** | 1.25.0 | SQL to Go code generator |
| **wire** | 0.5.0 | Dependency injection |
| **mockery** | 2.38+ | Mock generation |
| **golang-migrate** | 4.16+ | Database migrations |

### Code Quality & Security

| Tool | Version | Purpose |
|------|---------|---------|
| **golangci-lint** | 1.55.2 | Go linters aggregator |
| **gofumpt** | 0.5.0 | Stricter gofmt |
| **gosec** | 2.18+ | Security analyzer |
| **govulncheck** | latest | Vulnerability checker |
| **gitleaks** | 8.18+ | Secret scanner |
| **trivy** | 0.48+ | Container scanner |

### Development Tools

| Tool | Version | Purpose |
|------|---------|---------|
| **air** | 1.49+ | Hot reload for Go |
| **delve** | 1.21+ | Go debugger |
| **evans** | 0.10+ | gRPC client |
| **grpcurl** | 1.8+ | gRPC curl |
| **httpie** | 3.2+ | HTTP client (optional) |

### Observability Stack

| Tool | Version | Purpose |
|------|---------|---------|
| **OTEL Collector** | 0.91+ | Telemetry collection |
| **Prometheus** | 2.48+ | Metrics storage |
| **Grafana** | 10.2+ | Visualization |
| **Jaeger** | 1.52+ | Distributed tracing |

## Installation Instructions

### macOS Installation

```bash
# Install Homebrew (if not already installed)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Core tools
brew install go@1.22
brew install docker
brew install docker-compose
brew install make
brew install git

# Kubernetes tools
brew install kubectl
brew install helm
brew install kind
brew install k9s

# Development tools
brew install buf
brew install sqlc
brew install golang-migrate
brew install protobuf
brew install evans
brew install grpcurl
brew install httpie
brew install jq
brew install yq

# Security tools
brew install gitleaks
brew install trivy

# Go tools (installed via go install)
go install github.com/cosmtrek/air@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2
go install mvdan.cc/gofumpt@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/google/wire/cmd/wire@latest
go install github.com/vektra/mockery/v2@latest
go install github.com/go-delve/delve/cmd/dlv@latest
```

### Linux Installation (Ubuntu/Debian)

```bash
# Update package manager
sudo apt-get update

# Install prerequisites
sudo apt-get install -y curl wget git make build-essential

# Install Go
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
source ~/.bashrc

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.23.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Install kubectl
curl -LO "https://dl.k8s.io/release/v1.28.0/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Install Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Install buf
BIN="/usr/local/bin" && \
VERSION="1.28.1" && \
curl -sSL \
  "https://github.com/bufbuild/buf/releases/download/v${VERSION}/buf-$(uname -s)-$(uname -m)" \
  -o "${BIN}/buf" && \
chmod +x "${BIN}/buf"

# Install sqlc
curl -L https://github.com/sqlc-dev/sqlc/releases/download/v1.25.0/sqlc_1.25.0_linux_amd64.tar.gz | sudo tar -C /usr/local/bin -xz

# Install golang-migrate
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# Install security tools
curl -L https://github.com/aquasecurity/trivy/releases/download/v0.48.0/trivy_0.48.0_Linux-64bit.tar.gz | tar xvz
sudo mv trivy /usr/local/bin/

curl -L https://github.com/gitleaks/gitleaks/releases/download/v8.18.1/gitleaks_8.18.1_linux_x64.tar.gz | tar xvz
sudo mv gitleaks /usr/local/bin/

# Install Go tools
go install github.com/cosmtrek/air@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2
go install mvdan.cc/gofumpt@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/google/wire/cmd/wire@latest
go install github.com/vektra/mockery/v2@latest
go install github.com/go-delve/delve/cmd/dlv@latest
```

### Windows Installation

```powershell
# Install Chocolatey (if not installed)
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

# Install tools via Chocolatey
choco install golang --version=1.22.0
choco install docker-desktop
choco install make
choco install git
choco install kubernetes-cli
choco install kubernetes-helm
choco install kind

# Install Go tools
go install github.com/cosmtrek/air@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2
go install mvdan.cc/gofumpt@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/google/wire/cmd/wire@latest
go install github.com/vektra/mockery/v2@latest
go install github.com/go-delve/delve/cmd/dlv@latest

# Install buf (download from GitHub releases)
# Visit: https://github.com/bufbuild/buf/releases

# Install sqlc (download from GitHub releases)
# Visit: https://github.com/sqlc-dev/sqlc/releases
```

## Environment Configuration

### Go Environment

```bash
# Add to your shell profile (~/.bashrc, ~/.zshrc, etc.)
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org,direct
export GOPRIVATE=github.com/lumitut/*
```

### Docker Configuration

```bash
# Increase Docker resources (Docker Desktop)
# - CPUs: 4+
# - Memory: 8GB+
# - Disk: 50GB+

# Configure buildkit for better builds
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1
```

### Git Configuration

```bash
# Set up Git
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# Configure Git for Go
git config --global url."ssh://git@github.com/".insteadOf "https://github.com/"

# Set up GPG signing (optional but recommended)
git config --global commit.gpgsign true
git config --global user.signingkey YOUR_GPG_KEY_ID
```

## IDE Setup

### Visual Studio Code

Install recommended extensions:
```bash
code --install-extension golang.go
code --install-extension ms-vscode.makefile-tools
code --install-extension ms-azuretools.vscode-docker
code --install-extension ms-kubernetes-tools.vscode-kubernetes-tools
code --install-extension redhat.vscode-yaml
code --install-extension 42Crunch.vscode-openapi
code --install-extension zxh404.vscode-proto3
```

### GoLand/IntelliJ IDEA

1. Install Go plugin
2. Configure GOPATH and GOROOT
3. Enable Go modules support
4. Install Docker and Kubernetes plugins
5. Configure code style to use gofumpt

## Verification Script

Create and run this verification script to check your setup:

```bash
#!/bin/bash
# Save as verify-setup.sh

echo "Checking tool versions..."

check_tool() {
    local tool=$1
    local version_cmd=$2
    local min_version=$3
    
    if command -v $tool &> /dev/null; then
        version=$($version_cmd 2>&1 | head -n 1)
        echo "✓ $tool: $version"
    else
        echo "✗ $tool: NOT INSTALLED (minimum: $min_version)"
    fi
}

check_tool "go" "go version" "1.21"
check_tool "docker" "docker --version" "20.10"
check_tool "docker-compose" "docker-compose --version" "2.0"
check_tool "make" "make --version" "3.81"
check_tool "kubectl" "kubectl version --client --short" "1.24"
check_tool "helm" "helm version --short" "3.10"
check_tool "buf" "buf --version" "1.28"
check_tool "sqlc" "sqlc version" "1.25"
check_tool "migrate" "migrate -version" "4.16"
check_tool "golangci-lint" "golangci-lint version" "1.55"
check_tool "air" "air -v" "1.49"
check_tool "gitleaks" "gitleaks version" "8.18"
check_tool "trivy" "trivy --version" "0.48"

echo ""
echo "Checking Go tools..."
for tool in gosec govulncheck wire mockery dlv gofumpt; do
    if command -v $tool &> /dev/null; then
        echo "✓ $tool: installed"
    else
        echo "✗ $tool: NOT INSTALLED"
    fi
done

echo ""
echo "Checking environment variables..."
[ -n "$GOPATH" ] && echo "✓ GOPATH: $GOPATH" || echo "✗ GOPATH: NOT SET"
[ -n "$DOCKER_BUILDKIT" ] && echo "✓ DOCKER_BUILDKIT: $DOCKER_BUILDKIT" || echo "⚠ DOCKER_BUILDKIT: NOT SET (optional)"
```

## Quick Setup

For a quick setup, run:
```bash
# Clone the repository
git clone https://github.com/lumitut/lumi-go.git
cd lumi-go

# Run the setup script
make init

# Verify installation
./scripts/verify-setup.sh

# Start local environment
make up
```

## Troubleshooting

### Common Issues

1. **Go modules not downloading**
   ```bash
   go clean -modcache
   go mod download
   ```

2. **Docker permission denied**
   ```bash
   sudo usermod -aG docker $USER
   newgrp docker
   ```

3. **Port already in use**
   ```bash
   # Find and kill process using port
   lsof -ti:8080 | xargs kill -9
   ```

4. **Kubernetes context issues**
   ```bash
   kubectl config get-contexts
   kubectl config use-context docker-desktop
   ```

## Additional Resources

- [Go Documentation](https://go.dev/doc/)
- [Docker Documentation](https://docs.docker.com/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Buf Documentation](https://buf.build/docs/)
- [SQLC Documentation](https://docs.sqlc.dev/)
- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)

## Support

For issues or questions:
- Create an issue in the GitHub repository
- Contact the platform team: platform@lumitut.com
- Check the [FAQ](./FAQ.md) document
