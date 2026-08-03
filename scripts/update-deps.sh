#!/bin/bash

# Update direct dependencies for Go modules
# Usage:
#   All modules:    ./scripts/update-deps.sh
#   Single module:  ./scripts/update-deps.sh --module oss
#   With options:   ./scripts/update-deps.sh --no-tidy --test

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

MODULE=""
TIDY="true"
RUN_TESTS=""

# Show help
show_help() {
    echo "Usage: $0 [options] [--module <name>]"
    echo ""
    echo "Update direct dependencies for Go modules using bounded patch upgrades."
    echo ""
    echo "Options:"
    echo "  --module, -m     Target a single module (e.g., oss, data, logging)"
    echo "  --clean          Run go mod tidy (compatibility alias; enabled by default)"
    echo "  --no-tidy        Skip go mod tidy"
    echo "  --test, -t       Run tests after updating"
    echo "  --help, -h       Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                          # Update all modules"
    echo "  $0 --module oss             # Update only oss module"
    echo "  $0 --clean --test           # Update, clean, and test all modules"
    echo "  $0 -m data -t               # Update and test single module"
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
        --clean)
            TIDY="true"
            shift
            ;;
        --no-tidy)
            TIDY=""
            shift
            ;;
        --test|-t)
            RUN_TESTS="true"
            shift
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

echo -e "${BLUE}Updating dependencies...${NC}"
[ -n "$MODULE" ] && echo -e "${BLUE}Module: ${MODULE}${NC}"
[ -n "$TIDY" ] && echo -e "${BLUE}Tidy mode: enabled${NC}"
[ -n "$RUN_TESTS" ] && echo -e "${BLUE}Test mode: enabled${NC}"
echo "================================"

FAILED_MODULES=()
UPDATED_COUNT=0

for module in "${MODULES[@]}"; do
    echo ""
    echo -e "${BLUE}Module: ${module}${NC}"
    echo "----------------------------"

    pushd "$module" > /dev/null

    # Upgrade only direct dependencies with patch-level bounds. Updating every
    # transitive module with `go get -u ./...` can select incompatible module
    # paths such as github.com/armon/go-metrics and github.com/go-ini/ini.
    DIRECT_DEPS=()
    while IFS= read -r dependency; do
        if [ -n "$dependency" ] && [[ "$dependency" != github.com/ncobase/ncore/* ]]; then
            DIRECT_DEPS+=("$dependency")
        fi
    done < <(go list -m -f '{{if and (not .Main) (not .Indirect)}}{{.Path}}{{end}}' all)

    if [ ${#DIRECT_DEPS[@]} -gt 0 ]; then
        echo "Running: go get -u=patch <direct dependencies>"
        if output=$(go get -u=patch "${DIRECT_DEPS[@]}" 2>&1); then
            if [ -n "$output" ]; then
                echo "$output"
            fi
            echo -e "${GREEN}[UPDATE] Direct dependencies updated${NC}"
        else
            echo "$output"
            echo -e "${RED}[ERROR] Failed to update direct dependencies in ${module}${NC}"
            echo -e "${YELLOW}Review renamed module paths manually; do not replace a module with a different declared path.${NC}"
            FAILED_MODULES+=("$module")
            popd > /dev/null
            continue
        fi
    else
        echo -e "${YELLOW}[UPDATE] No external direct dependencies require an upgrade${NC}"
    fi

    # Clean up if requested
    if [ -n "$TIDY" ]; then
        echo "Running: go mod tidy"
        if output=$(go mod tidy 2>&1); then
            if [ -n "$output" ]; then
                echo "$output"
            fi
            echo -e "${GREEN}[CLEAN] Dependencies tidied${NC}"
        else
            echo "$output"
            echo -e "${RED}[ERROR] Failed to tidy dependencies in ${module}${NC}"
            FAILED_MODULES+=("$module")
            popd > /dev/null
            continue
        fi
    fi

    # Run tests if requested
    if [ -n "$RUN_TESTS" ]; then
        echo "Running: go test ./..."
        if output=$(go test ./... 2>&1); then
            echo "$output" | head -20
            echo -e "${GREEN}[TEST] Tests passed${NC}"
        else
            echo "$output" | head -20
            echo -e "${RED}[ERROR] Tests failed${NC}"
            FAILED_MODULES+=("$module")
            popd > /dev/null
            continue
        fi
    fi

    popd > /dev/null
    UPDATED_COUNT=$((UPDATED_COUNT + 1))
    echo -e "${GREEN}[SUCCESS] ${module} updated successfully${NC}"
done

echo ""
echo "================================"
echo -e "${BLUE}Results: ${GREEN}$UPDATED_COUNT updated${NC}, ${RED}${#FAILED_MODULES[@]} failed${NC}"

if [ ${#FAILED_MODULES[@]} -ne 0 ]; then
    echo ""
    echo -e "${RED}Failed modules:${NC}"
    for module in "${FAILED_MODULES[@]}"; do
        echo "  - $module"
    done
    exit 1
else
    echo ""
    echo -e "${GREEN}✅ All modules updated successfully!${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Run: go work sync"
    echo "  2. Review changes: git diff go.work.sum"
    echo "  3. Commit changes if everything works"
fi
