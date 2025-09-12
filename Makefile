# FreyjaDB Makefile
# Build, test, and development automation

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod
GOFMT = gofmt
GOLINT = golangci-lint

# Binary and package information
BINARY_NAME = freyja
BINARY_UNIX = $(BINARY_NAME)_unix
MAIN_PACKAGE = ./cmd/freyja
PKG_LIST = $(shell go list ./... | grep -v /vendor/)

# Build information
VERSION ?= $(shell git describe --tags --always --dirty)
BUILD_TIME = $(shell date +%FT%T%z)
LDFLAGS = -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Docker
DOCKER_IMAGE = freyjadb
DOCKER_TAG ?= latest

# Colors for output
RED = \033[0;31m
GREEN = \033[0;32m
YELLOW = \033[0;33m
BLUE = \033[0;34m
NC = \033[0m # No Color

.PHONY: all build clean test test-verbose test-race test-cover help
.PHONY: lint lint-fix format format-check deps deps-update
.PHONY: run bench install docker docker-build docker-run
.PHONY: tools check-tools

# Default target
all: clean deps format lint test build

# Build the binary
build:
	@echo "$(GREEN)Building $(BINARY_NAME)...$(NC)"
	$(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME) $(MAIN_PACKAGE)

# Build for linux
build-linux:
	@echo "$(GREEN)Building $(BINARY_NAME) for Linux...$(NC)"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_UNIX) $(MAIN_PACKAGE)

# Clean build artifacts
clean:
	@echo "$(YELLOW)Cleaning...$(NC)"
	$(GOCLEAN)
	rm -rf bin/
	rm -rf coverage/
	rm -f coverage.out
	rm -f cpu.prof
	rm -f mem.prof
	rm -f trace.out

# Run fast unit tests (excludes fuzz and bench tests)
test:
	@echo "$(BLUE)Running fast unit tests...$(NC)"
	$(GOTEST) -v -short ./...

# Run all tests including slow ones
test-all:
	@echo "$(BLUE)Running all tests (including slow ones)...$(NC)"
	$(GOTEST) -v ./...

# Run tests with verbose output
test-verbose:
	@echo "$(BLUE)Running tests with verbose output...$(NC)"
	$(GOTEST) -v -race -coverprofile=coverage.out -short ./...

# Run tests with race detection
test-race:
	@echo "$(BLUE)Running tests with race detection...$(NC)"
	$(GOTEST) -race -short ./...

# Run tests in parallel
test-parallel:
	@echo "$(BLUE)Running tests in parallel...$(NC)"
	$(GOTEST) -v -short -parallel=4 ./...

# Run tests with coverage
test-cover: test-verbose
	@echo "$(BLUE)Generating coverage report...$(NC)"
	mkdir -p coverage
	$(GOCMD) tool cover -html=coverage.out -o coverage/coverage.html
	$(GOCMD) tool cover -func=coverage.out

# Run benchmarks
bench:
	@echo "$(BLUE)Running benchmarks...$(NC)"
	$(GOTEST) -bench=. -benchmem -tags=bench ./...

# Run fuzz tests
fuzz:
	@echo "$(BLUE)Running fuzz tests...$(NC)"
	$(GOTEST) -fuzz=. -fuzztime=10s -tags=fuzz ./...

# Run benchmarks with CPU profiling
bench-cpu:
	@echo "$(BLUE)Running benchmarks with CPU profiling...$(NC)"
	$(GOTEST) -bench=. -benchmem -cpuprofile=cpu.prof ./...

# Run benchmarks with memory profiling
bench-mem:
	@echo "$(BLUE)Running benchmarks with memory profiling...$(NC)"
	$(GOTEST) -bench=. -benchmem -memprofile=mem.prof ./...

# Lint the code
lint: check-tools
	@echo "$(BLUE)Running linter...$(NC)"
	$(GOLINT) run

# Lint and fix automatically fixable issues
lint-fix: check-tools
	@echo "$(BLUE)Running linter with auto-fix...$(NC)"
	$(GOLINT) run --fix

# Format code
format:
	@echo "$(BLUE)Formatting code...$(NC)"
	$(GOFMT) -s -w .
	$(GOCMD) mod tidy

# Check if code is formatted
format-check:
	@echo "$(BLUE)Checking code format...$(NC)"
	@if [ -n "$$($(GOFMT) -l .)" ]; then \
		echo "$(RED)Code is not formatted. Run 'make format' to fix.$(NC)"; \
		$(GOFMT) -l .; \
		exit 1; \
	fi
	@echo "$(GREEN)Code is properly formatted.$(NC)"

# Update dependencies
deps:
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	$(GOMOD) download
	$(GOMOD) tidy

# Update dependencies to latest versions
deps-update:
	@echo "$(BLUE)Updating dependencies...$(NC)"
	$(GOGET) -u ./...
	$(GOMOD) tidy

# Install the binary
install: build
	@echo "$(GREEN)Installing $(BINARY_NAME)...$(NC)"
	cp bin/$(BINARY_NAME) $(GOPATH)/bin/

# Run the application
run: build
	@echo "$(GREEN)Running $(BINARY_NAME)...$(NC)"
	./bin/$(BINARY_NAME)

# Run with specific command
run-serve: build
	@echo "$(GREEN)Running $(BINARY_NAME) serve...$(NC)"
	./bin/$(BINARY_NAME) serve

# Development tools installation
tools:
	@echo "$(BLUE)Installing development tools...$(NC)"
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Check if required tools are installed
check-tools:
	@which $(GOLINT) > /dev/null || (echo "$(RED)golangci-lint not found. Run 'make tools' to install.$(NC)" && exit 1)

# Docker operations
docker-build:
	@echo "$(BLUE)Building Docker image...$(NC)"
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: docker-build
	@echo "$(GREEN)Running Docker container...$(NC)"
	docker run --rm -it $(DOCKER_IMAGE):$(DOCKER_TAG)

# Security scan
security:
	@echo "$(BLUE)Running security scan...$(NC)"
	$(GOLINT) run --enable gosec

# Generate mocks (if using mockery)
mocks:
	@echo "$(BLUE)Generating mocks...$(NC)"
	@if command -v mockery >/dev/null 2>&1; then \
		mockery --all --output ./mocks; \
	else \
		echo "$(YELLOW)mockery not installed. Skipping mock generation.$(NC)"; \
	fi

# Database operations (for development)
db-setup:
	@echo "$(BLUE)Setting up development database...$(NC)"
	mkdir -p ./testdata/db

db-clean:
	@echo "$(YELLOW)Cleaning development database...$(NC)"
	rm -rf ./testdata/db/*

# Integration tests
test-integration:
	@echo "$(BLUE)Running integration tests...$(NC)"
	$(GOTEST) -tags=integration -v ./tests/integration/...

# Performance tests
test-performance:
	@echo "$(BLUE)Running performance tests...$(NC)"
	$(GOTEST) -tags=performance -v ./tests/performance/...

# Full CI pipeline
ci: deps format-check lint test-race test-cover

# Development workflow
dev: clean deps format lint test build
	@echo "$(GREEN)Development build complete!$(NC)"

# Release preparation
release-check: clean deps format-check lint test-race test-cover build
	@echo "$(GREEN)Release checks passed!$(NC)"

# Generate documentation
docs:
	@echo "$(BLUE)Generating documentation...$(NC)"
	$(GOCMD) doc -all > docs/API.md

# Show project statistics
stats:
	@echo "$(BLUE)Project Statistics:$(NC)"
	@echo "Lines of code:"
	@find . -name "*.go" -not -path "./vendor/*" | xargs wc -l | tail -1
	@echo "Number of packages:"
	@find . -name "*.go" -not -path "./vendor/*" -exec dirname {} \; | sort -u | wc -l
	@echo "Test coverage:"
	@if [ -f coverage.out ]; then \
		$(GOCMD) tool cover -func=coverage.out | tail -1; \
	else \
		echo "No coverage data. Run 'make test-cover' first."; \
	fi

# Help target
help:
	@echo "$(GREEN)FreyjaDB Build System$(NC)"
	@echo ""
	@echo "$(BLUE)Build Targets:$(NC)"
	@echo "  build           Build the binary"
	@echo "  build-linux     Build for Linux"
	@echo "  clean           Clean build artifacts"
	@echo "  install         Install the binary"
	@echo ""
	@echo "$(BLUE)Testing Targets:$(NC)"
	@echo "  test            Run fast unit tests (excludes fuzz/bench)"
	@echo "  test-all        Run all tests including slow ones"
	@echo "  test-verbose    Run tests with verbose output and coverage"
	@echo "  test-race       Run tests with race detection"
	@echo "  test-parallel   Run tests in parallel"
	@echo "  test-cover      Run tests with coverage report"
	@echo "  test-integration Run integration tests"
	@echo "  test-performance Run performance tests"
	@echo "  bench           Run benchmarks"
	@echo "  bench-cpu       Run benchmarks with CPU profiling"
	@echo "  bench-mem       Run benchmarks with memory profiling"
	@echo "  fuzz            Run fuzz tests"
	@echo ""
	@echo "$(BLUE)Code Quality Targets:$(NC)"
	@echo "  lint            Run linter"
	@echo "  lint-fix        Run linter with auto-fix"
	@echo "  format          Format code"
	@echo "  format-check    Check code formatting"
	@echo "  security        Run security scan"
	@echo ""
	@echo "$(BLUE)Development Targets:$(NC)"
	@echo "  deps            Download dependencies"
	@echo "  deps-update     Update dependencies"
	@echo "  tools           Install development tools"
	@echo "  run             Run the application"
	@echo "  run-serve       Run with serve command"
	@echo "  dev             Full development build"
	@echo ""
	@echo "$(BLUE)CI/CD Targets:$(NC)"
	@echo "  ci              Full CI pipeline"
	@echo "  release-check   Release preparation checks"
	@echo ""
	@echo "$(BLUE)Docker Targets:$(NC)"
	@echo "  docker-build    Build Docker image"
	@echo "  docker-run      Build and run Docker container"
	@echo ""
	@echo "$(BLUE)Utility Targets:$(NC)"
	@echo "  docs            Generate documentation"
	@echo "  stats           Show project statistics"
	@echo "  help            Show this help message"
