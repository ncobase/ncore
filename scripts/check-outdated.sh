#!/bin/bash

# Check for outdated dependencies in Go modules
# Usage:
#   All modules:    ./scripts/check-outdated.sh
#   Single module:  ./scripts/check-outdated.sh --module oss

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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
    echo "  $0 -m data              # Check only data module"
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
            echo -e "${RED}Unknown option: $1${NC}"
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
        echo -e "${RED}Error: Module '$MODULE' not found (no go.mod in $MODULE/)${NC}"
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
    echo -e "${YELLOW}No modules found.${NC}"
    exit 0
fi

echo -e "${BLUE}Checking for outdated dependencies...${NC}"
[ -n "$MODULE" ] && echo -e "${BLUE}Module: ${MODULE}${NC}"
echo "====================================="

OUTDATED_COUNT=0

for module in "${MODULES[@]}"; do
    echo ""
    echo -e "${BLUE}Module: ${module}${NC}"
    echo "----------------------------"

    pushd "$module" > /dev/null

    # List upgradable dependencies (suppress error if none found)
    if output=$(go list -u -m all 2>/dev/null | grep '\[' || true); then
        if [ -n "$output" ]; then
            echo -e "${YELLOW}$output${NC}"
            OUTDATED_COUNT=$((OUTDATED_COUNT + 1))
        else
            echo -e "${GREEN}✅ All dependencies are up to date.${NC}"
        fi
    else
        echo -e "${RED}Failed to check dependencies.${NC}"
    fi

    popd > /dev/null
done

echo ""
echo "====================================="
if [ $OUTDATED_COUNT -gt 0 ]; then
    echo -e "${YELLOW}⚠️  Found outdated dependencies in $OUTDATED_COUNT module(s).${NC}"
    echo ""
    echo "To update dependencies, run:"
    echo "  ./scripts/update-deps.sh"
    echo ""
    echo "To update and test:"
    echo "  ./scripts/update-deps.sh --clean --test"
else
    echo -e "${GREEN}✅ All modules have up-to-date dependencies!${NC}"
fi
