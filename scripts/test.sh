#!/bin/bash

# Test modules
# Usage:
#   ./scripts/test.sh           # run all tests
#   ./scripts/test.sh -v        # verbose output
#   ./scripts/test.sh -cover    # with coverage

set -e

VERBOSE=""
COVERAGE=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE="-v"
            shift
            ;;
        -cover|--coverage)
            COVERAGE="-cover"
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [-v|--verbose] [-cover|--coverage]"
            exit 1
            ;;
    esac
done

# Auto-discover all subdirectories containing Go files
MODULES=()
for dir in */; do
    # Remove trailing slash
    dir=${dir%/}
    # Skip non-directories and hidden ones
    [[ -d "$dir" && "$dir" != .* ]] || continue
    # Check if there are any Go files in this directory or its subdirs
    if find "$dir" -name "*.go" -type f | grep -q .; then
        MODULES+=("$dir")
    fi
done

if [[ ${#MODULES[@]} -eq 0 ]]; then
    echo "No Go modules found."
    exit 0
fi

echo "Running tests for all modules..."
echo "================================"

FAILED_MODULES=()

for module in "${MODULES[@]}"; do
    echo ""
    echo "🧪 Testing module: $module"
    echo "----------------------------"

    cd "$module"

    if go test $VERBOSE $COVERAGE ./...; then
        echo "✅ $module tests passed"
    else
        echo "❌ $module tests failed"
        FAILED_MODULES+=("$module")
    fi

    cd ..
done

echo ""
echo "================================"

if [ ${#FAILED_MODULES[@]} -eq 0 ]; then
    echo "✅ All tests passed!"
    exit 0
else
    echo "❌ Tests failed in the following modules:"
    for module in "${FAILED_MODULES[@]}"; do
        echo "  - $module"
    done
    exit 1
fi
