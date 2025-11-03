#!/bin/bash

# Create or remove git tags
# Usage: ./scripts/tag.sh v0.1.0
#        ./scripts/tag.sh --remove v0.1.0

set -e

REMOVE=0
VERSION=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --remove)
            REMOVE=1
            shift
            ;;
        *)
            VERSION="$1"
            shift
            ;;
    esac
done

if [ -z "$VERSION" ]; then
    echo "Usage: $0 [--remove] <version>"
    echo "Example: $0 v0.1.0"
    echo "         $0 --remove v0.1.0"
    exit 1
fi

# Auto-discover modules: directories containing go.mod
MODULES=()
while IFS= read -r dir; do
    MODULES+=("$(basename "$dir")")
done < <(find . -name 'go.mod' -exec dirname {} \; | sort -u)

if [ ${#MODULES[@]} -eq 0 ]; then
    echo "No modules found (no go.mod files detected)."
    exit 0
fi

if [ "$REMOVE" -eq 1 ]; then
    echo "Removing tags for version: $VERSION"
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
    echo "  git push origin --delete <module>/$VERSION"
else
    echo "Creating tags for version: $VERSION"
    echo "================================"
    for module in "${MODULES[@]}"; do
        TAG_NAME="$module/$VERSION"
        echo "Creating tag: $TAG_NAME"
        git tag "$TAG_NAME"
    done
    echo ""
    echo "Tags created successfully!"
    echo "To push all tags, run:"
    echo "  git push origin --tags"
    echo ""
    echo "To push specific module tags, run:"
    echo "  git push origin <module>/$VERSION"
fi
