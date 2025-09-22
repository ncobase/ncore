#!/usr/bin/make

# Metadata
VERSION=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
BRANCH=$(shell git symbolic-ref -q --short HEAD 2>/dev/null || echo "unknown")
REVISION=$(shell git rev-parse --short HEAD)
BUILT_AT=$(shell date -u '+%Y-%m-%dT%H:%M:%S')
GO_VERSION=$(shell go version | awk '{print $$3}')

# Build targets
.PHONY: test lint vet fmt help version clean deps mod-tidy

# Default target should be help
.DEFAULT_GOAL := help

# Clean artifacts
clean:
	@echo "Cleaning artifacts..."
	@go clean
	@go clean -modcache

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download

# Tidy dependencies
mod-tidy:
	@echo "Tidying dependencies..."
	@go mod tidy

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed"; \
		exit 1; \
	fi

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Print version information
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(REVISION)"
	@echo "Build Time: $(BUILT_AT)"
	@echo "Go Version: $(GO_VERSION)"

# Show help
help:
	@echo "NCore Library Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  clean     Clean artifacts and module cache"
	@echo "  deps      Download dependencies"
	@echo "  mod-tidy  Tidy dependencies"
	@echo "  test      Run tests"
	@echo "  lint      Run linter"
	@echo "  vet       Run go vet"
	@echo "  fmt       Format code"
	@echo "  version   Print version information"
	@echo "  help      Show this help"
