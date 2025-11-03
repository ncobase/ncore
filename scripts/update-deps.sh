#!/bin/bash

# Update dependencies
# Usage:
#   ./scripts/update-deps.sh        # Update all modules
#   ./scripts/update-deps.sh data   # Update only the data module

set -e

# Auto-detect all subdirectories that contain a go.mod file
MODULES=()
for dir in */; do
    if [ -f "${dir}go.mod" ]; then
        MODULES+=("${dir%/}")
    fi
done

# If a module name is provided, only update that module
if [ -n "$1" ]; then
    MODULES=("$1")
fi

echo "Updating dependencies for modules..."
echo "===================================="

for module in "${MODULES[@]}"; do
    if [ ! -d "$module" ]; then
        echo "⚠️  Module $module not found, skipping..."
        continue
    fi

    echo ""
    echo "📦 Updating module: $module"
    echo "----------------------------"

    cd "$module"

    # Upgrade all dependencies to the latest minor or patch version
    echo "Running: go get -u ./..."
    go get -u ./...

    # Clean up unused dependencies
    echo "Running: go mod tidy"
    go mod tidy

    cd ..

    echo "✅ $module updated"
done

echo ""
echo "===================================="
echo "All modules updated successfully!"
echo ""
echo "Next steps:"
echo "  1. Run: go work sync"
echo "  2. Test: bash scripts/test.sh"
echo "  3. Commit changes if everything works"
