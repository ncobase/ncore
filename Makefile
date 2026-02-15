.PHONY: help sync test test-v test-cover test-parallel update check-outdated tag clean fmt lint

help:
	@echo "NCore - Multi-module Repository"
	@echo ""
	@echo "Available targets:"
	@echo ""
	@echo "Testing:"
	@echo "  make test          - Run all tests"
	@echo "  make test-v        - Run all tests (verbose)"
	@echo "  make test-cover    - Run all tests with coverage"
	@echo "  make test-parallel - Run all tests in parallel (faster)"
	@echo ""
	@echo "Dependencies:"
	@echo "  make sync          - Sync workspace dependencies"
	@echo "  make update        - Update all dependencies"
	@echo "  make check-outdated- Check for outdated dependencies"
	@echo ""
	@echo "Versioning:"
	@echo "  make tag VERSION=v0.3.0          - Tag all modules"
	@echo "  make tag-push VERSION=v0.3.0     - Tag all modules and push"
	@echo ""
	@echo "Code Quality:"
	@echo "  make fmt           - Format all code"
	@echo "  make lint          - Run linter (requires golangci-lint)"
	@echo "  make clean         - Clean build artifacts"
	@echo ""

sync:
	@echo "Syncing workspace dependencies..."
	@go work sync
	@echo "✅ Sync complete"

# Testing targets
test:
	@./scripts/test.sh

test-v:
	@./scripts/test.sh -v

test-cover:
	@./scripts/test.sh --cover

test-parallel:
	@./scripts/test.sh --parallel

# Dependency management
update:
	@./scripts/update-deps.sh
	@echo ""
	@echo "Don't forget to run: make sync"

check-outdated:
	@./scripts/check-outdated.sh

# Version management via git tags
tag:
ifndef VERSION
	@echo "Error: VERSION is required"
	@echo "Usage: make tag VERSION=v0.3.0"
	@exit 1
endif
	@./scripts/tag.sh $(VERSION)

tag-push:
ifndef VERSION
	@echo "Error: VERSION is required"
	@echo "Usage: make tag-push VERSION=v0.3.0"
	@exit 1
endif
	@./scripts/tag.sh --push $(VERSION)

# Code quality
fmt:
	@echo "Formatting code..."
	@for dir in */; do \
		if [ -f "$$dir/go.mod" ]; then \
			echo "Formatting $$dir"; \
			cd "$$dir" && go fmt ./... && cd ..; \
		fi \
	done
	@echo "✅ Format complete"

lint:
	@echo "Running linter..."
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "❌ golangci-lint not found. Install it from: https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi
	@for dir in */; do \
		if [ -f "$$dir/go.mod" ]; then \
			echo "Linting $$dir"; \
			cd "$$dir" && golangci-lint run && cd ..; \
		fi \
	done
	@echo "✅ Lint complete"

clean:
	@echo "Cleaning build artifacts..."
	@find . -name "*.test" -delete
	@find . -name "*.out" -delete
	@find . -name "coverage.txt" -delete
	@rm -f go.work.sum
	@echo "✅ Clean complete"
