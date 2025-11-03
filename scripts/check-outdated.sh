#!/bin/bash

# Check for outdated dependencies
# Usage: ./scripts/check-outdated.sh

set -e

# Auto-discover all directories containing a go.mod file
MODULES=($(find . -maxdepth 1 -type d -name '[^.]*' -exec test -f '{}/go.mod' \; -print | sort))

echo "Checking for outdated dependencies..."
echo "====================================="

for module in "${MODULES[@]}"; do
    echo ""
    echo "📦 Module: $module"
    echo "----------------------------"

    cd "$module"

    # List upgradable dependencies
    go list -u -m all | grep '\['

    cd ..
done

echo ""
echo "====================================="
echo "Check complete!"
echo ""
echo "To update all modules, run:"
echo "  ./scripts/update-deps.sh"
