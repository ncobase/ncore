#!/usr/bin/make

# Metadata
BINARY_NAME=ncore
VERSION=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
BRANCH=$(shell git symbolic-ref -q --short HEAD 2>/dev/null || echo "unknown")
REVISION=$(shell git rev-parse --short HEAD)
BUILT_AT=$(shell date -u '+%Y-%m-%dT%H:%M:%S')
GO_VERSION=$(shell go version | awk '{print $$3}')

# Build flags
BUILD_VARS=github.com/ncobase/ncore/version
LDFLAGS=-X $(BUILD_VARS).Version=$(VERSION) \
        -X $(BUILD_VARS).Branch=$(BRANCH) \
        -X $(BUILD_VARS).Revision=$(REVISION) \
        -X $(BUILD_VARS).BuiltAt=$(BUILT_AT) \
        -X $(BUILD_VARS).GoVersion=$(GO_VERSION)

BUILD_FLAGS= -trimpath -ldflags "$(LDFLAGS)"

# Debug mode
ifeq ($(debug), 1)
BUILD_FLAGS += -gcflags "-N -l"
endif

# Build targets
.PHONY: all build clean test lint vet fmt help version release

# Default target should be help
.DEFAULT_GOAL := help

# Build the application
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@go build $(BUILD_FLAGS) -o $(BINARY_NAME) ./cmd

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -rf dist/

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

# Build all
all: clean build

# Build for multiple platforms
release:
	@echo "Building release binaries..."
	@mkdir -p dist

	@echo "Building for Linux (amd64)..."
	@GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o dist/$(BINARY_NAME)-$(VERSION)-linux-amd64 ./cmd
	@cd dist && sha256sum $(BINARY_NAME)-$(VERSION)-linux-amd64 > $(BINARY_NAME)-$(VERSION)-linux-amd64.sha256

	@echo "Building for Linux (arm64)..."
	@GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) -o dist/$(BINARY_NAME)-$(VERSION)-linux-arm64 ./cmd
	@cd dist && sha256sum $(BINARY_NAME)-$(VERSION)-linux-arm64 > $(BINARY_NAME)-$(VERSION)-linux-arm64.sha256

	@echo "Building for macOS (amd64)..."
	@GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o dist/$(BINARY_NAME)-$(VERSION)-darwin-amd64 ./cmd
	@cd dist && shasum -a 256 $(BINARY_NAME)-$(VERSION)-darwin-amd64 > $(BINARY_NAME)-$(VERSION)-darwin-amd64.sha256

	@echo "Building for macOS (arm64)..."
	@GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) -o dist/$(BINARY_NAME)-$(VERSION)-darwin-arm64 ./cmd
	@cd dist && shasum -a 256 $(BINARY_NAME)-$(VERSION)-darwin-arm64 > $(BINARY_NAME)-$(VERSION)-darwin-arm64.sha256

	@echo "Building for Windows (amd64)..."
	@GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o dist/$(BINARY_NAME)-$(VERSION)-windows-amd64.exe ./cmd
	@cd dist && sha256sum $(BINARY_NAME)-$(VERSION)-windows-amd64.exe > $(BINARY_NAME)-$(VERSION)-windows-amd64.exe.sha256

	@echo "Release binaries built successfully in ./dist directory"

# Show help
help:
	@echo "NCore Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build     Build the application"
	@echo "  clean     Clean build artifacts"
	@echo "  test      Run tests"
	@echo "  lint      Run linter"
	@echo "  vet       Run go vet"
	@echo "  fmt       Format code"
	@echo "  version   Print version information"
	@echo "  all       Clean and build the application"
	@echo "  release   Build release binaries for multiple platforms"
	@echo "  help      Show this help"
	@echo ""
	@echo "Options:"
	@echo "  debug=1   Enable debug symbols in build"
