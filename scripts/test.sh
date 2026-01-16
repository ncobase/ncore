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

# Auto-discover all directories containing go.mod
MODULES=()
while IFS= read -r file; do
    dir=$(dirname "$file")
    # Skip the root directory itself if it has a go.mod (though in this repo root has go.work)
    if [ "$dir" == "." ]; then continue; fi
    MODULES+=("$dir")
done < <(find . -name "go.mod" | sort)

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

    pushd "$module" > /dev/null

    if go test $VERBOSE $COVERAGE ./...; then
        echo "✅ $module tests passed"
    else
        echo "❌ $module tests failed"
        FAILED_MODULES+=("$module")
    fi

    popd > /dev/null
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
