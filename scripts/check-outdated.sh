#!/bin/bash

# Check for outdated dependencies in Go modules
# Usage:
#   All modules:    ./scripts/check-outdated.sh
#   Single module:  ./scripts/check-outdated.sh --module oss

set -e

MODULE=""

# Show help
show_help() {
    echo "Usage: $0 [--module <name>]"
    echo ""
    echo "Check for outdated dependencies in Go modules."
    echo ""
    echo "Options:"
    echo "  --module, -m     Target a single module (e.g., oss, data, logging)"
    echo "  --help, -h       Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                      # Check all modules"
    echo "  $0 --module oss         # Check only oss module"
    echo "  $0 -m data/postgres     # Check only data/postgres module"
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

echo "Checking for outdated dependencies..."
[ -n "$MODULE" ] && echo "Module: $MODULE"
echo "====================================="

OUTDATED_COUNT=0

for module in "${MODULES[@]}"; do
    echo ""
    echo "Module: $module"
    echo "----------------------------"

    pushd "$module" > /dev/null

    # List upgradable dependencies (suppress error if none found)
    if output=$(go list -u -m all 2>/dev/null | grep '\[' || true); then
        if [ -n "$output" ]; then
            echo "$output"
            OUTDATED_COUNT=$((OUTDATED_COUNT + 1))
        else
            echo "All dependencies are up to date."
        fi
    else
        echo "Failed to check dependencies."
    fi

    popd > /dev/null
done

echo ""
echo "====================================="
if [ $OUTDATED_COUNT -gt 0 ]; then
    echo "Found outdated dependencies in $OUTDATED_COUNT module(s)."
    echo ""
    echo "To update dependencies, run:"
    echo "  ./scripts/update-deps.sh"
else
    echo "All modules have up-to-date dependencies."
fi
