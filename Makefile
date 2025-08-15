# Makefile for lumi-go (Go Microservice Template)

# Variables
BINARY_NAME := lumi-go
SERVICE_NAME := lumi-go
DOCKER_REGISTRY := lumitut
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GO_VERSION := 1.22

# Go build flags
LDFLAGS := -ldflags "-s -w \
	-X 'main.Version=$(VERSION)' \
	-X 'main.BuildTime=$(BUILD_TIME)' \
	-X 'main.GitCommit=$(GIT_COMMIT)'"

# Directories
CMD_DIR := ./cmd/server
INTERNAL_DIR := ./internal
MIGRATIONS_DIR := ./migrations
DEPLOY_DIR := ./deploy

# Database
DATABASE_URL ?= postgres://lumigo:lumigo@localhost:5432/lumigo?sslmode=disable

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

# Default target
.DEFAULT_GOAL := help

## help: Display this help message
.PHONY: help
help:
	@echo "$(BLUE)lumi-go Makefile$(NC)"
	@echo "$(YELLOW)Usage:$(NC)"
	@echo "  make [target]"
	@echo ""
	@echo "$(YELLOW)Available targets:$(NC)"
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  $(GREEN)%-20s$(NC) %s\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

## init: Initialize project dependencies and tools
.PHONY: init
init:
	@echo "$(YELLOW)Initializing project...$(NC)"
	go mod download
	go mod tidy
	@echo "$(YELLOW)Installing development tools...$(NC)"
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "$(GREEN)✓ Project initialized$(NC)"

## run: Run the application with hot-reload (development)
.PHONY: run
run:
	@echo "$(YELLOW)Starting application with hot-reload...$(NC)"
	air -c .air.toml

## build: Build the application binary
.PHONY: build
build:
	@echo "$(YELLOW)Building $(BINARY_NAME)...$(NC)"
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) $(CMD_DIR)
	@echo "$(GREEN)✓ Binary built: bin/$(BINARY_NAME)$(NC)"

## build-all: Build for multiple platforms
.PHONY: build-all
build-all:
	@echo "$(YELLOW)Building for multiple platforms...$(NC)"
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)
	@echo "$(GREEN)✓ Multi-platform build complete$(NC)"

## test: Run all tests
.PHONY: test
test:
	@echo "$(YELLOW)Running tests...$(NC)"
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./tests/...
	@echo "$(GREEN)✓ Tests complete$(NC)"

## test-short: Run short tests only
.PHONY: test-short
test-short:
	@echo "$(YELLOW)Running short tests...$(NC)"
	go test -v -short ./tests/...

## test-integration: Run integration tests
.PHONY: test-integration
test-integration:
	@echo "$(YELLOW)Running integration tests...$(NC)"
	go test -v -tags=integration ./tests/...

## coverage: Generate test coverage report
.PHONY: coverage
coverage: test
	@echo "$(YELLOW)Generating coverage report...$(NC)"
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report generated: coverage.html$(NC)"

## benchmark: Run benchmarks
.PHONY: benchmark
benchmark:
	@echo "$(YELLOW)Running benchmarks...$(NC)"
	go test -bench=. -benchmem ./tests/...

## lint: Run linters
.PHONY: lint
lint:
	@echo "$(YELLOW)Running linters...$(NC)"
	golangci-lint run ./...
	@echo "$(GREEN)✓ Linting complete$(NC)"

## fmt: Format code
.PHONY: fmt
fmt:
	@echo "$(YELLOW)Formatting code...$(NC)"
	gofmt -s -w .
	goimports -w .
	go mod tidy
	@echo "$(GREEN)✓ Code formatted$(NC)"

## vet: Run go vet
.PHONY: vet
vet:
	@echo "$(YELLOW)Running go vet...$(NC)"
	go vet ./...
	@echo "$(GREEN)✓ Vet complete$(NC)"

## security-scan: Run security scans
.PHONY: security-scan
security-scan:
	@echo "$(YELLOW)Running security scans...$(NC)"
	govulncheck ./...
	gosec -fmt sarif -out gosec-results.sarif ./...
	@echo "$(GREEN)✓ Security scan complete$(NC)"

## clean: Clean build artifacts
.PHONY: clean
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	rm -rf bin/ tmp/ coverage.* *.test *.out
	@echo "$(GREEN)✓ Clean complete$(NC)"

# Docker targets
## docker-build: Build Docker image
.PHONY: docker-build
docker-build:
	@echo "$(YELLOW)Building Docker image...$(NC)"
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-f deploy/docker/Dockerfile \
		-t $(DOCKER_REGISTRY)/$(SERVICE_NAME):$(VERSION) \
		-t $(DOCKER_REGISTRY)/$(SERVICE_NAME):latest \
		.
	@echo "$(GREEN)✓ Docker image built$(NC)"

## docker-push: Push Docker image to registry
.PHONY: docker-push
docker-push:
	@echo "$(YELLOW)Pushing Docker image...$(NC)"
	docker push $(DOCKER_REGISTRY)/$(SERVICE_NAME):$(VERSION)
	docker push $(DOCKER_REGISTRY)/$(SERVICE_NAME):latest
	@echo "$(GREEN)✓ Docker image pushed$(NC)"

## docker-run: Run Docker container
.PHONY: docker-run
docker-run:
	@echo "$(YELLOW)Running Docker container...$(NC)"
	docker run -p 8080:8080 -p 8081:8081 -p 9090:9090 \
		$(DOCKER_REGISTRY)/$(SERVICE_NAME):$(VERSION)

# Docker Compose targets
## up: Start all services with docker-compose
.PHONY: up
up:
	@echo "$(YELLOW)Starting services with docker-compose...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)✓ Services started$(NC)"
	@echo "$(BLUE)Services:$(NC)"
	@echo "  • Application: http://localhost:8080"
	@echo "  • Metrics: http://localhost:9090/metrics"
	@echo "  • Prometheus: http://localhost:9091"
	@echo "  • Grafana: http://localhost:3000 (admin/admin)"
	@echo "  • Jaeger: http://localhost:16686"

## down: Stop all services
.PHONY: down
down:
	@echo "$(YELLOW)Stopping services...$(NC)"
	docker-compose down
	@echo "$(GREEN)✓ Services stopped$(NC)"

## restart: Restart all services
.PHONY: restart
restart: down up

## logs: View docker-compose logs
.PHONY: logs
logs:
	docker-compose logs -f

## ps: List running services
.PHONY: ps
ps:
	docker-compose ps

# Database migration targets
## migrate-up: Run database migrations up
.PHONY: migrate-up
migrate-up:
	@echo "$(YELLOW)Running migrations up...$(NC)"
	migrate -database "$(DATABASE_URL)" -path $(MIGRATIONS_DIR) up
	@echo "$(GREEN)✓ Migrations complete$(NC)"

## migrate-down: Rollback last migration
.PHONY: migrate-down
migrate-down:
	@echo "$(YELLOW)Rolling back last migration...$(NC)"
	migrate -database "$(DATABASE_URL)" -path $(MIGRATIONS_DIR) down 1
	@echo "$(GREEN)✓ Rollback complete$(NC)"

## migrate-reset: Reset database
.PHONY: migrate-reset
migrate-reset:
	@echo "$(YELLOW)Resetting database...$(NC)"
	migrate -database "$(DATABASE_URL)" -path $(MIGRATIONS_DIR) down -all
	migrate -database "$(DATABASE_URL)" -path $(MIGRATIONS_DIR) up
	@echo "$(GREEN)✓ Database reset complete$(NC)"

## migrate-create: Create new migration
.PHONY: migrate-create
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $$name
	@echo "$(GREEN)✓ Migration created$(NC)"

## migrate-version: Show current migration version
.PHONY: migrate-version
migrate-version:
	@echo "$(YELLOW)Current migration version:$(NC)"
	@migrate -database "$(DATABASE_URL)" -path $(MIGRATIONS_DIR) version

# Code generation targets
## gen: Run all code generators
.PHONY: gen
gen:
	@echo "$(YELLOW)Running code generators...$(NC)"
	@# Add your generators here
	@# go generate ./...
	@# wire ./...
	@# sqlc generate
	@# buf generate
	@echo "$(GREEN)✓ Code generation complete$(NC)"

## gen-check: Check if generated code is up to date
.PHONY: gen-check
gen-check: gen
	@echo "$(YELLOW)Checking generated code...$(NC)"
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "$(RED)✗ Generated code is not up to date$(NC)"; \
		git status --short; \
		exit 1; \
	else \
		echo "$(GREEN)✓ Generated code is up to date$(NC)"; \
	fi

# Helm targets
## helm-lint: Lint Helm chart
.PHONY: helm-lint
helm-lint:
	@echo "$(YELLOW)Linting Helm chart...$(NC)"
	helm lint $(DEPLOY_DIR)/helm
	@echo "$(GREEN)✓ Helm lint complete$(NC)"

## helm-template: Render Helm templates
.PHONY: helm-template
helm-template:
	@echo "$(YELLOW)Rendering Helm templates...$(NC)"
	helm template $(SERVICE_NAME) $(DEPLOY_DIR)/helm

## helm-install: Install Helm chart
.PHONY: helm-install
helm-install:
	@echo "$(YELLOW)Installing Helm chart...$(NC)"
	helm install $(SERVICE_NAME) $(DEPLOY_DIR)/helm
	@echo "$(GREEN)✓ Helm chart installed$(NC)"

## helm-upgrade: Upgrade Helm release
.PHONY: helm-upgrade
helm-upgrade:
	@echo "$(YELLOW)Upgrading Helm release...$(NC)"
	helm upgrade $(SERVICE_NAME) $(DEPLOY_DIR)/helm
	@echo "$(GREEN)✓ Helm release upgraded$(NC)"

## helm-uninstall: Uninstall Helm release
.PHONY: helm-uninstall
helm-uninstall:
	@echo "$(YELLOW)Uninstalling Helm release...$(NC)"
	helm uninstall $(SERVICE_NAME)
	@echo "$(GREEN)✓ Helm release uninstalled$(NC)"

# CI/CD targets
## ci: Run CI pipeline locally
.PHONY: ci
ci: fmt lint vet test security-scan
	@echo "$(GREEN)✓ CI pipeline complete$(NC)"

## release: Create a new release
.PHONY: release
release:
	@echo "$(YELLOW)Creating release...$(NC)"
	@read -p "Enter version (e.g., v1.0.0): " version; \
	git tag -a $$version -m "Release $$version"; \
	git push origin $$version
	@echo "$(GREEN)✓ Release created$(NC)"

# Utility targets
## info: Display project information
.PHONY: info
info:
	@echo "$(BLUE)Project Information:$(NC)"
	@echo "  • Service: $(SERVICE_NAME)"
	@echo "  • Version: $(VERSION)"
	@echo "  • Git Commit: $(GIT_COMMIT)"
	@echo "  • Build Time: $(BUILD_TIME)"
	@echo "  • Go Version: $(shell go version)"
	@echo "  • Docker Registry: $(DOCKER_REGISTRY)"

## setup-labels: Setup GitHub labels
.PHONY: setup-labels
setup-labels:
	@echo "$(YELLOW)Setting up GitHub labels...$(NC)"
	./.github/scripts/setup-labels.sh
	@echo "$(GREEN)✓ Labels configured$(NC)"

## pre-commit: Install pre-commit hooks
.PHONY: pre-commit
pre-commit:
	@echo "$(YELLOW)Installing pre-commit hooks...$(NC)"
	pip install pre-commit
	pre-commit install
	@echo "$(GREEN)✓ Pre-commit hooks installed$(NC)"

.PHONY: all
all: clean fmt lint vet test build
