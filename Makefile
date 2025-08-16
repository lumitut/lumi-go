# Project variables
PROJECT_NAME := lumi-go
MODULE := github.com/lumitut/$(PROJECT_NAME)
MAIN_PATH := ./cmd/server
BINARY_NAME := server

# Go variables
GO := go
GOTEST := $(GO) test
GOBUILD := $(GO) build
GOCLEAN := $(GO) clean
GOGET := $(GO) get
GOMOD := $(GO) mod
GOFMT := gofmt
GOLINT := golangci-lint
GOVET := $(GO) vet
MOCKGEN := mockgen

# Coverage variables
COVERAGE_DIR := coverage
COVERAGE_FILE := $(COVERAGE_DIR)/coverage.out
COVERAGE_HTML := $(COVERAGE_DIR)/coverage.html
COVERAGE_THRESHOLD := 80

# Build variables
BUILD_DIR := build
LDFLAGS := -ldflags "-X main.Version=$$(git describe --tags --always --dirty) -X main.BuildTime=$$(date -u +%Y%m%d-%H%M%S)"

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m # No Color

.PHONY: all build test clean help

## help: Display this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^##' Makefile | sed 's/## /  /'

## all: Build and test the project
all: clean fmt vet test build

## build: Build the binary
build:
	@echo "$(GREEN)Building $(PROJECT_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## run: Run the application locally
run: build
	@echo "$(GREEN)Running $(PROJECT_NAME)...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME)

## run-dev: Run the application with hot reload (requires air)
run-dev:
	@echo "$(GREEN)Running $(PROJECT_NAME) in development mode...$(NC)"
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "$(YELLOW)Air not installed. Install with: go install github.com/air-verse/air@latest$(NC)"; \
		exit 1; \
	fi

## test: Run all tests
test:
	@echo "$(GREEN)Running tests...$(NC)"
	$(GOTEST) -v -race ./...

## test-unit: Run unit tests only
test-unit:
	@echo "$(GREEN)Running unit tests...$(NC)"
	$(GOTEST) -v -race ./tests/unit/...

## test-integration: Run integration tests
test-integration:
	@echo "$(GREEN)Running integration tests...$(NC)"
	$(GOTEST) -v -race -tags=integration ./tests/integration/...

## test-e2e: Run end-to-end tests
test-e2e:
	@echo "$(GREEN)Running e2e tests...$(NC)"
	$(GOTEST) -v -race -tags=e2e ./tests/e2e/...

## test-smoke: Run smoke tests
test-smoke:
	@echo "$(GREEN)Running smoke tests...$(NC)"
	$(GOTEST) -v -race -tags=smoke ./tests/smoke/...

## test-all: Run all test suites
test-all:
	@echo "$(GREEN)Running all tests...$(NC)"
	$(GOTEST) -v -race -tags="integration e2e smoke" ./tests/...

## coverage: Generate test coverage report
coverage:
	@echo "$(GREEN)Generating coverage report...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -race -covermode=atomic -coverprofile=$(COVERAGE_FILE) ./...
	$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "$(GREEN)Coverage report generated: $(COVERAGE_HTML)$(NC)"
	@echo "$(YELLOW)Coverage summary:$(NC)"
	@$(GO) tool cover -func=$(COVERAGE_FILE) | tail -1

## coverage-check: Check if coverage meets threshold
coverage-check: coverage
	@echo "$(GREEN)Checking coverage threshold ($(COVERAGE_THRESHOLD)%)...$(NC)"
	@coverage=$$($(GO) tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ $$(echo "$$coverage < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
		echo "$(RED)Coverage $$coverage% is below threshold $(COVERAGE_THRESHOLD)%$(NC)"; \
		exit 1; \
	else \
		echo "$(GREEN)Coverage $$coverage% meets threshold $(COVERAGE_THRESHOLD)%$(NC)"; \
	fi

## bench: Run benchmarks
bench:
	@echo "$(GREEN)Running benchmarks...$(NC)"
	$(GOTEST) -bench=. -benchmem ./...

## fmt: Format code
fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	$(GOFMT) -w -s .
	$(GOMOD) tidy

## vet: Run go vet
vet:
	@echo "$(GREEN)Running go vet...$(NC)"
	$(GOVET) ./...

## lint: Run linter
lint:
	@echo "$(GREEN)Running linter...$(NC)"
	@if command -v $(GOLINT) >/dev/null 2>&1; then \
		$(GOLINT) run ./...; \
	else \
		echo "$(YELLOW)golangci-lint not installed, skipping...$(NC)"; \
		echo "$(YELLOW)Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)"; \
	fi

## clean: Clean build artifacts
clean:
	@echo "$(GREEN)Cleaning...$(NC)"
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR) $(COVERAGE_DIR) tmp/
	@rm -f $(BINARY_NAME)

## deps: Download and verify dependencies
deps:
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	$(GOMOD) download
	$(GOMOD) verify
	$(GOMOD) tidy

## deps-update: Update dependencies to latest versions
deps-update:
	@echo "$(GREEN)Updating dependencies...$(NC)"
	$(GOGET) -u ./...
	$(GOMOD) tidy

## install-tools: Install development tools
install-tools:
	@echo "$(GREEN)Installing development tools...$(NC)"
	$(GO) install github.com/golang/mock/mockgen@latest
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install golang.org/x/tools/cmd/goimports@latest
	$(GO) install github.com/air-verse/air@latest

## generate: Generate code (mocks, etc.)
generate:
	@echo "$(GREEN)Generating code...$(NC)"
	$(GO) generate ./...

## docker-build: Build Docker image
docker-build:
	@echo "$(GREEN)Building Docker image...$(NC)"
	docker build -t $(PROJECT_NAME):latest -f deploy/docker/Dockerfile .

## docker-build-dev: Build development Docker image
docker-build-dev:
	@echo "$(GREEN)Building development Docker image...$(NC)"
	docker build -t $(PROJECT_NAME):dev -f deploy/docker/Dockerfile.dev .

## docker-run: Run Docker container
docker-run: docker-build
	@echo "$(GREEN)Running Docker container...$(NC)"
	docker run -p 8080:8080 -p 8081:8081 -p 9090:9090 $(PROJECT_NAME):latest

## docker-dev: Run development environment with Docker Compose
docker-dev:
	@echo "$(GREEN)Starting development environment...$(NC)"
	docker-compose -f docker-compose.dev.yml up --build

## docker-stop: Stop Docker containers
docker-stop:
	@echo "$(GREEN)Stopping Docker containers...$(NC)"
	docker-compose -f docker-compose.dev.yml down

## ci: Run CI pipeline locally
ci: clean deps fmt vet lint test coverage-check build
	@echo "$(GREEN)CI pipeline complete!$(NC)"

## release: Create a new release
release: ci
	@echo "$(GREEN)Creating release...$(NC)"
	@read -p "Enter version (e.g., v1.0.0): " version; \
	git tag -a $$version -m "Release $$version"; \
	git push origin $$version; \
	echo "$(GREEN)Release $$version created$(NC)"

## health: Check service health
health:
	@echo "$(GREEN)Checking service health...$(NC)"
	@curl -f http://localhost:8080/health || echo "$(RED)Service is not healthy$(NC)"

## ready: Check service readiness
ready:
	@echo "$(GREEN)Checking service readiness...$(NC)"
	@curl -f http://localhost:8080/ready || echo "$(RED)Service is not ready$(NC)"

## metrics: Display service metrics
metrics:
	@echo "$(GREEN)Fetching service metrics...$(NC)"
	@curl -s http://localhost:9090/metrics | head -20

.DEFAULT_GOAL := help
