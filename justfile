# FreyjaDB Justfile
# Build, test, and development automation

# Variables
binary_name := "freyja"
binary_unix := "{{binary_name}}_unix"
main_package := "./cmd/freyja"
version := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`
build_time := `date +%FT%T%z`
ldflags := "-ldflags \"-X main.Version={{version}} -X main.BuildTime={{build_time}}\""

# Colors for output
red := '''\033[0;31m'''
green := '''\033[0;32m'''
yellow := '''\033[0;33m'''
blue := '''\033[0;34m'''
nc := '''\033[0m''' # No Color

# Default recipe
default:
    just clean deps format lint test build

# Build the binary
build:
    @echo "{{blue}}Building {{binary_name}}...{{nc}}"
    @mkdir -p bin
    go build {{ldflags}} -o bin/{{binary_name}} {{main_package}}

# Build for linux
build-linux:
    @echo "{{blue}}Building {{binary_name}} for Linux...{{nc}}"
    @mkdir -p bin
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build {{ldflags}} -o bin/{{binary_unix}} {{main_package}}

# Clean build artifacts
clean:
    @echo "{{yellow}}Cleaning...{{nc}}"
    go clean
    rm -rf bin/
    rm -f coverage.out
    rm -rf coverage/

# Run fast unit tests (excludes fuzz and bench tests)
test:
    @echo "{{blue}}Running fast unit tests...{{nc}}"
    go test -v -short ./...

# Run all tests including slow ones
test-all:
    @echo "{{blue}}Running all tests (including slow ones)...{{nc}}"
    go test -v ./...

# Run tests with verbose output
test-verbose:
    @echo "{{blue}}Running tests with verbose output...{{nc}}"
    go test -v -race -coverprofile=coverage.out -short ./...

# Run tests with race detection
test-race:
    @echo "{{blue}}Running tests with race detection...{{nc}}"
    go test -race -short ./...

# Run tests in parallel
test-parallel:
    @echo "{{blue}}Running tests in parallel...{{nc}}"
    go test -v -short -parallel=4 ./...

# Run tests with coverage
test-cover: test-verbose
    @echo "{{blue}}Generating coverage report...{{nc}}"
    @mkdir -p coverage
    go tool cover -html=coverage.out -o coverage/coverage.html
    go tool cover -func=coverage.out

# Run benchmarks
bench:
    @echo "{{blue}}Running benchmarks...{{nc}}"
    go test -bench=. -benchmem -tags=bench ./...

# Run fuzz tests
fuzz:
    @echo "{{blue}}Running fuzz tests...{{nc}}"
    go test -fuzz=. -fuzztime=10s -tags=fuzz ./...

# Run benchmarks with CPU profiling
bench-cpu:
    @echo "{{blue}}Running benchmarks with CPU profiling...{{nc}}"
    go test -bench=. -benchmem -cpuprofile=cpu.prof ./...

# Run benchmarks with memory profiling
bench-mem:
    @echo "{{blue}}Running benchmarks with memory profiling...{{nc}}"
    go test -bench=. -benchmem -memprofile=mem.prof ./...

# Check if golangci-lint is installed
check-tools:
    if ! command -v golangci-lint >/dev/null 2>&1; then \
        echo "{{red}}golangci-lint not found. Run 'just tools' to install.{{nc}}"; \
        exit 1; \
    fi

# Lint the code
lint: check-tools
    @echo "{{blue}}Running linter...{{nc}}"
    golangci-lint run

# Lint and fix automatically fixable issues
lint-fix: check-tools
    @echo "{{blue}}Running linter with auto-fix...{{nc}}"
    golangci-lint run --fix

# Format code
format:
    @echo "{{blue}}Formatting code...{{nc}}"
    gofmt -s -w .
    go mod tidy

# Check if code is formatted
format-check:
    @echo "{{blue}}Checking code format...{{nc}}"
    @if [ -n "$(gofmt -l .)" ]; then \
        echo "{{red}}Code is not formatted. Run 'just format' to fix.{{nc}}"; \
        gofmt -l .; \
        exit 1; \
    fi
    @echo "{{green}}Code is properly formatted.{{nc}}"

# Update dependencies
deps:
    @echo "{{blue}}Downloading dependencies...{{nc}}"
    go mod download
    go mod tidy

# Update dependencies to latest versions
deps-update:
    @echo "{{blue}}Updating dependencies...{{nc}}"
    go get -u ./...
    go mod tidy

# Install the binary
install: build
    @echo "{{green}}Installing {{binary_name}}...{{nc}}"
    cp bin/{{binary_name}} $(GOPATH)/bin/

# Run the application
run: build
    @echo "{{green}}Running {{binary_name}}...{{nc}}"
    ./bin/{{binary_name}}

# Run with specific command
run-serve: build
    @echo "{{green}}Running {{binary_name}} serve...{{nc}}"
    ./bin/{{binary_name}} serve

# Development tools installation
tools:
    @echo "{{blue}}Installing development tools...{{nc}}"
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Docker operations
docker-build:
    @echo "{{blue}}Building Docker image...{{nc}}"
    docker build -t freyjadb:latest .

docker-run: docker-build
    @echo "{{green}}Running Docker container...{{nc}}"
    docker run --rm -it freyjadb:latest

# Security scan
security: check-tools
    @echo "{{blue}}Running security scan...{{nc}}"
    golangci-lint run --enable gosec

# Generate mocks (if using mockery)
mocks:
    @echo "{{blue}}Generating mocks...{{nc}}"
    @if command -v mockery >/dev/null 2>&1; then \
        mockery --all --output ./mocks; \
    else \
        echo "{{yellow}}mockery not installed. Skipping mock generation.{{nc}}"; \
    fi

# Database operations (for development)
db-setup:
    @echo "{{blue}}Setting up development database...{{nc}}"
    mkdir -p ./testdata/db

db-clean:
    @echo "{{yellow}}Cleaning development database...{{nc}}"
    rm -rf ./testdata/db/*

# Integration tests
test-integration:
    @echo "{{blue}}Running integration tests...{{nc}}"
    go test -tags=integration -v ./tests/integration/...

# Performance tests
test-performance:
    @echo "{{blue}}Running performance tests...{{nc}}"
    go test -tags=performance -v ./tests/performance/...

# Full CI pipeline
ci: deps format-check lint test-race test-cover

# Development workflow
dev: clean deps format lint test build
    @echo "{{green}}Development build complete!{{nc}}"

# Release preparation
release-check: clean deps format-check lint test-race test-cover build
    @echo "{{green}}Release checks passed!{{nc}}"

# Generate documentation
docs:
    @echo "{{blue}}Generating documentation...{{nc}}"
    go doc -all > docs/API.md

# Generate Swagger documentation
swagger:
    @echo "{{blue}}Generating Swagger documentation...{{nc}}"
    swag init -g pkg/api/server.go -o pkg/api

# Show project statistics
stats:
    @echo "{{blue}}Project Statistics:{{nc}}"
    @echo "Lines of code:"
    @find . -name "*.go" -not -path "./vendor/*" | xargs wc -l | tail -1
    @echo "Number of packages:"
    @find . -name "*.go" -not -path "./vendor/*" -exec dirname {} \; | sort -u | wc -l
    @echo "Test coverage:"
    @if [ -f coverage.out ]; then \
        go tool cover -func=coverage.out | tail -1; \
    else \
        echo "No coverage data. Run 'just test-cover' first."; \
    fi

# Help target
help:
    @echo "{{green}}FreyjaDB Build System{{nc}}"
    @echo ""
    @echo "{{blue}}Build Targets:{{nc}}"
    @echo "  build           Build the binary"
    @echo "  build-linux     Build for Linux"
    @echo "  clean           Clean build artifacts"
    @echo "  install         Install the binary"
    @echo ""
    @echo "{{blue}}Testing Targets:{{nc}}"
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
    @echo "{{blue}}Code Quality Targets:{{nc}}"
    @echo "  lint            Run linter"
    @echo "  lint-fix        Run linter with auto-fix"
    @echo "  format          Format code"
    @echo "  format-check    Check code formatting"
    @echo "  security        Run security scan"
    @echo ""
    @echo "{{blue}}Development Targets:{{nc}}"
    @echo "  deps            Download dependencies"
    @echo "  deps-update     Update dependencies"
    @echo "  tools           Install development tools"
    @echo "  run             Run the application"
    @echo "  run-serve       Run with serve command"
    @echo "  dev             Full development build"
    @echo ""
    @echo "{{blue}}CI/CD Targets:{{nc}}"
    @echo "  ci              Full CI pipeline"
    @echo "  release-check   Release preparation checks"
    @echo ""
    @echo "{{blue}}Docker Targets:{{nc}}"
    @echo "  docker-build    Build Docker image"
    @echo "  docker-run      Build and run Docker container"
    @echo ""
    @echo "{{blue}}Utility Targets:{{nc}}"
    @echo "  docs            Generate documentation"
    @echo "  swagger         Generate Swagger documentation"
    @echo "  stats           Show project statistics"
    @echo "  help            Show this help message"

# Aliases for compatibility
all: default