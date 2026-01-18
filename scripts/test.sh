#!/bin/bash

# Run tests for Go modules
# Usage:
#   All modules:    ./scripts/test.sh
#   Single module:  ./scripts/test.sh --module oss
#   With options:   ./scripts/test.sh -v --cover

set -e

VERBOSE=""
COVERAGE=""
MODULE=""

# Show help
show_help() {
    echo "Usage: $0 [options] [--module <name>]"
    echo ""
    echo "Run tests for Go modules."
    echo ""
    echo "Options:"
    echo "  --module, -m     Target a single module (e.g., oss, data, logging)"
    echo "  -v, --verbose    Verbose test output"
    echo "  --cover          Enable coverage reporting"
    echo "  --help, -h       Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                          # Test all modules"
    echo "  $0 --module oss             # Test only oss module"
    echo "  $0 -v --cover               # Verbose with coverage"
    echo "  $0 -m data/postgres -v      # Test single module verbosely"
    exit 0
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --help|-h)
            show_help
            ;;
        --module|-m)
            MODULE="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE="-v"
            shift
            ;;
        --cover|--coverage)
            COVERAGE="-cover"
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Run '$0 --help' for usage."
            exit 1
            ;;
    esac
done

# Build module list
MODULES=()

if [ -n "$MODULE" ]; then
    # Single module mode
    if [ -f "$MODULE/go.mod" ]; then
        MODULES+=("$MODULE")
    elif [ -f "./$MODULE/go.mod" ]; then
        MODULES+=("$MODULE")
    else
        echo "Error: Module '$MODULE' not found (no go.mod in $MODULE/)"
        exit 1
    fi
else
    # Auto-discover modules: directories containing go.mod
    while IFS= read -r dir; do
        module="${dir#./}"
        MODULES+=("$module")
    done < <(find . -name 'go.mod' -not -path "./examples/*" -exec dirname {} \; | sort -u)
fi

if [ ${#MODULES[@]} -eq 0 ]; then
    echo "No modules found."
    exit 0
fi

echo "Running tests..."
[ -n "$MODULE" ] && echo "Module: $MODULE"
[ -n "$VERBOSE" ] && echo "Verbose: enabled"
[ -n "$COVERAGE" ] && echo "Coverage: enabled"
echo "================================"

FAILED_MODULES=()
PASSED_COUNT=0

for module in "${MODULES[@]}"; do
    echo ""
    echo "Testing: $module"
    echo "----------------------------"

    pushd "$module" > /dev/null

    if go test $VERBOSE $COVERAGE ./... 2>&1; then
        echo "[PASS] $module"
        PASSED_COUNT=$((PASSED_COUNT + 1))
    else
        echo "[FAIL] $module"
        FAILED_MODULES+=("$module")
    fi

    popd > /dev/null
done

echo ""
echo "================================"
echo "Results: $PASSED_COUNT passed, ${#FAILED_MODULES[@]} failed"

if [ ${#FAILED_MODULES[@]} -eq 0 ]; then
    echo "All tests passed."
    exit 0
else
    echo ""
    echo "Failed modules:"
    for module in "${FAILED_MODULES[@]}"; do
        echo "  - $module"
    done
    exit 1
fi
