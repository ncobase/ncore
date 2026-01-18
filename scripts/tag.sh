#!/bin/bash

# Create or remove git tags for modules
# Usage:
#   All modules:    ./scripts/tag.sh v0.1.0
#   Single module:  ./scripts/tag.sh --module oss v0.2.3
#   Remove all:     ./scripts/tag.sh --remove v0.1.0
#   Remove single:  ./scripts/tag.sh --remove --module oss v0.2.3

set -e

REMOVE=0
VERSION=""
MODULE=""

# Show help
show_help() {
    echo "Usage: $0 [--remove] [--module <name>] <version>"
    echo ""
    echo "Options:"
    echo "  --remove         Remove tags instead of creating them"
    echo "  --module, -m     Target a single module (e.g., oss, data, logging)"
    echo "  --help, -h       Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 v0.1.0                      # Tag all modules with v0.1.0"
    echo "  $0 --module oss v0.2.3         # Tag only oss module with v0.2.3"
    echo "  $0 --remove v0.1.0             # Remove v0.1.0 tags from all modules"
    echo "  $0 --remove --module oss v0.2.3 # Remove v0.2.3 tag from oss module"
    exit 0
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --help|-h)
            show_help
            ;;
        --remove)
            REMOVE=1
            shift
            ;;
        --module|-m)
            MODULE="$2"
            shift 2
            ;;
        *)
            VERSION="$1"
            shift
            ;;
    esac
done

if [ -z "$VERSION" ]; then
    echo "Error: Version is required."
    echo "Run '$0 --help' for usage."
    exit 1
fi

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
        # Remove leading './' and use full relative path
        module="${dir#./}"
        MODULES+=("$module")
    done < <(find . -name 'go.mod' -not -path "./examples/*" -exec dirname {} \; | sort -u)
fi

if [ ${#MODULES[@]} -eq 0 ]; then
    echo "No modules found."
    exit 0
fi

if [ "$REMOVE" -eq 1 ]; then
    echo "Removing tags for version: $VERSION"
    [ -n "$MODULE" ] && echo "Module: $MODULE"
    echo "================================"
    for module in "${MODULES[@]}"; do
        TAG_NAME="$module/$VERSION"
        if git rev-parse -q --verify "refs/tags/$TAG_NAME" >/dev/null; then
            echo "Removing tag: $TAG_NAME"
            git tag -d "$TAG_NAME"
        else
            echo "Tag not found, skipping: $TAG_NAME"
        fi
    done
    echo ""
    echo "Tags removed locally!"
    echo "To delete remote tags, run:"
    for module in "${MODULES[@]}"; do
        echo "  git push origin --delete $module/$VERSION"
    done
else
    echo "Creating tags for version: $VERSION"
    [ -n "$MODULE" ] && echo "Module: $MODULE"
    echo "================================"
    for module in "${MODULES[@]}"; do
        TAG_NAME="$module/$VERSION"
        if git rev-parse -q --verify "refs/tags/$TAG_NAME" >/dev/null; then
            echo "Tag already exists, skipping: $TAG_NAME"
        else
            echo "Creating tag: $TAG_NAME"
            git tag "$TAG_NAME"
        fi
    done
    echo ""
    echo "Tags created successfully!"
    echo "To push tags, run:"
    for module in "${MODULES[@]}"; do
        echo "  git push origin $module/$VERSION"
    done
fi
