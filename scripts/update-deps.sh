#!/bin/bash

# Update dependencies for Go modules
# Usage:
#   All modules:    ./scripts/update-deps.sh
#   Single module:  ./scripts/update-deps.sh --module oss
#   With options:   ./scripts/update-deps.sh --clean --test

set -e

MODULE=""
CLEAN=""
RUN_TESTS=""

# Show help
show_help() {
    echo "Usage: $0 [options] [--module <name>]"
    echo ""
    echo "Update dependencies for Go modules."
    echo ""
    echo "Options:"
    echo "  --module, -m     Target a single module (e.g., oss, data, logging)"
    echo "  --clean          Remove unused dependencies (go mod tidy)"
    echo "  --test, -t       Run tests after updating"
    echo "  --help, -h       Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                          # Update all modules"
    echo "  $0 --module oss             # Update only oss module"
    echo "  $0 --clean --test           # Update, clean, and test all modules"
    echo "  $0 -m data/postgres -t      # Update and test single module"
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
        --clean)
            CLEAN="true"
            shift
            ;;
        --test|-t)
            RUN_TESTS="true"
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

echo "Updating dependencies..."
[ -n "$MODULE" ] && echo "Module: $MODULE"
[ -n "$CLEAN" ] && echo "Clean mode: enabled"
[ -n "$RUN_TESTS" ] && echo "Test mode: enabled"
echo "================================"

FAILED_MODULES=()
UPDATED_COUNT=0

for module in "${MODULES[@]}"; do
    echo ""
    echo "Module: $module"
    echo "----------------------------"

    pushd "$module" > /dev/null

    # Upgrade all dependencies
    echo "Running: go get -u ./..."
    if go get -u ./... 2>&1; then
        echo "[UPDATE] Dependencies updated"
    else
        echo "[ERROR] Failed to update dependencies"
        FAILED_MODULES+=("$module")
        popd > /dev/null
        continue
    fi

    # Clean up if requested
    if [ -n "$CLEAN" ]; then
        echo "Running: go mod tidy"
        if go mod tidy 2>&1; then
            echo "[CLEAN] Dependencies tidied"
        else
            echo "[ERROR] Failed to tidy dependencies"
            FAILED_MODULES+=("$module")
            popd > /dev/null
            continue
        fi
    fi

    # Run tests if requested
    if [ -n "$RUN_TESTS" ]; then
        echo "Running: go test ./..."
        if go test ./... 2>&1; then
            echo "[TEST] Tests passed"
        else
            echo "[ERROR] Tests failed"
            FAILED_MODULES+=("$module")
            popd > /dev/null
            continue
        fi
    fi

    popd > /dev/null
    UPDATED_COUNT=$((UPDATED_COUNT + 1))
    echo "[SUCCESS] $module updated successfully"
done

echo ""
echo "================================"
echo "Results: $UPDATED_COUNT updated, ${#FAILED_MODULES[@]} failed"

if [ ${#FAILED_MODULES[@]} -ne 0 ]; then
    echo ""
    echo "Failed modules:"
    for module in "${FAILED_MODULES[@]}"; do
        echo "  - $module"
    done
    exit 1
else
    echo ""
    echo "All modules updated successfully!"
    echo ""
    echo "Next steps:"
    echo "  1. Run: go work sync"
    echo "  2. Commit changes if everything works"
fi