#!/bin/bash

# Run tests for Go modules
# Usage:
#   All modules:    ./scripts/test.sh
#   Single module:  ./scripts/test.sh --module oss
#   With options:   ./scripts/test.sh -v --cover
#   Parallel:       ./scripts/test.sh --parallel

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

VERBOSE=""
COVERAGE=""
MODULE=""
PARALLEL=""

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
    echo "  --parallel       Run tests in parallel (faster)"
    echo "  --help, -h       Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                          # Test all modules"
    echo "  $0 --module oss             # Test only oss module"
    echo "  $0 -v --cover               # Verbose with coverage"
    echo "  $0 --parallel               # Test all modules in parallel"
    echo "  $0 -m data -v               # Test single module verbosely"
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
        --parallel)
            PARALLEL="true"
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

echo -e "${BLUE}Running tests...${NC}"
[ -n "$MODULE" ] && echo -e "${BLUE}Module: ${MODULE}${NC}"
[ -n "$VERBOSE" ] && echo -e "${BLUE}Verbose: enabled${NC}"
[ -n "$COVERAGE" ] && echo -e "${BLUE}Coverage: enabled${NC}"
[ -n "$PARALLEL" ] && echo -e "${BLUE}Parallel: enabled${NC}"
echo "================================"

FAILED_MODULES=()
PASSED_COUNT=0

if [ -n "$PARALLEL" ]; then
    # Parallel execution - compatible with bash 3.2+
    PIDS=()
    MODULE_NAMES=()

    for module in "${MODULES[@]}"; do
        output_file="/tmp/test-${module//\//-}.log"

        (
            echo "Testing: $module" > "$output_file"
            echo "----------------------------" >> "$output_file"

            cd "$module"
            if go test $VERBOSE $COVERAGE ./... >> "$output_file" 2>&1; then
                echo "PASS" > "${output_file}.status"
            else
                echo "FAIL" > "${output_file}.status"
            fi
        ) &

        PIDS+=($!)
        MODULE_NAMES+=("$module")
    done

    # Wait for all processes and collect results
    for i in "${!PIDS[@]}"; do
        wait ${PIDS[$i]} || true
        module="${MODULE_NAMES[$i]}"
        output_file="/tmp/test-${module//\//-}.log"

        if [ -f "${output_file}.status" ] && [ "$(cat ${output_file}.status)" == "PASS" ]; then
            echo -e "${GREEN}[PASS]${NC} $module"
            PASSED_COUNT=$((PASSED_COUNT + 1))
        else
            echo -e "${RED}[FAIL]${NC} $module"
            FAILED_MODULES+=("$module")
            # Show error output in verbose mode
            if [ -n "$VERBOSE" ]; then
                cat "$output_file"
            fi
        fi

        rm -f "$output_file" "${output_file}.status"
    done
else
    # Sequential execution
    for module in "${MODULES[@]}"; do
        echo ""
        echo -e "${BLUE}Testing: ${module}${NC}"
        echo "----------------------------"

        pushd "$module" > /dev/null

        if go test $VERBOSE $COVERAGE ./... 2>&1; then
            echo -e "${GREEN}[PASS]${NC} $module"
            PASSED_COUNT=$((PASSED_COUNT + 1))
        else
            echo -e "${RED}[FAIL]${NC} $module"
            FAILED_MODULES+=("$module")
        fi

        popd > /dev/null
    done
fi

echo ""
echo "================================"
echo -e "${BLUE}Results: ${GREEN}$PASSED_COUNT passed${NC}, ${RED}${#FAILED_MODULES[@]} failed${NC}"

if [ ${#FAILED_MODULES[@]} -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}Failed modules:${NC}"
    for module in "${FAILED_MODULES[@]}"; do
        echo "  - $module"
    done
    exit 1
fi
