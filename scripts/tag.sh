#!/bin/bash

# Create or remove git tags for modules
# Usage:
#   All modules:    ./scripts/tag.sh v0.1.0
#   Single module:  ./scripts/tag.sh --module oss v0.2.3
#   Remove all:     ./scripts/tag.sh --remove v0.1.0
#   Remove single:  ./scripts/tag.sh --remove --module oss v0.2.3
#   Bump version:   ./scripts/tag.sh --bump [patch|minor|major]

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

REMOVE=0
VERSION=""
MODULE=""
PUSH=""
BUMP=""
BUMP_TYPE="patch"
FROM_VERSION=""

# Show help
show_help() {
    echo "Usage: $0 [OPTIONS] [version]"
    echo ""
    echo "Options:"
    echo "  --bump [TYPE]    Auto-increment version from latest tag"
    echo "                   TYPE: patch (default), minor, major"
    echo "  --from VERSION   Base version for --bump (default: auto-detect latest)"
    echo "  --remove         Remove tags instead of creating them"
    echo "  --module, -m     Target a single module (e.g., oss, data, logging)"
    echo "  --push           Automatically push tags to remote"
    echo "  --help, -h       Show this help message"
    echo ""
    echo "Examples:"
    echo "  # Manual versioning"
    echo "  $0 v0.1.0                          # Tag all modules with v0.1.0"
    echo "  $0 --module oss v0.2.3             # Tag only oss module with v0.2.3"
    echo "  $0 --push v0.1.0                   # Tag all and push to remote"
    echo ""
    echo "  # Auto-increment versioning"
    echo "  $0 --bump                          # Bump all modules (patch: v0.2.2 → v0.2.3)"
    echo "  $0 --bump minor                    # Bump minor (v0.2.2 → v0.3.0)"
    echo "  $0 --bump major                    # Bump major (v0.2.2 → v1.0.0)"
    echo "  $0 --bump --from v0.2.2            # Bump from specific version"
    echo "  $0 --bump --module oss             # Bump only oss module"
    echo "  $0 --bump --push                   # Bump all and push to remote"
    echo ""
    echo "  # Remove tags"
    echo "  $0 --remove v0.1.0                 # Remove v0.1.0 tags from all modules"
    echo "  $0 --remove --module oss v0.2.3    # Remove v0.2.3 tag from oss module"
    exit 0
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --help|-h)
            show_help
            ;;
        --bump)
            BUMP="true"
            # Check if next arg is a bump type
            if [[ "$2" =~ ^(patch|minor|major)$ ]]; then
                BUMP_TYPE="$2"
                shift 2
            else
                shift
            fi
            ;;
        --from)
            FROM_VERSION="$2"
            shift 2
            ;;
        --remove)
            REMOVE=1
            shift
            ;;
        --module|-m)
            MODULE="$2"
            shift 2
            ;;
        --push)
            PUSH="true"
            shift
            ;;
        *)
            VERSION="$1"
            shift
            ;;
    esac
done

# Function to get latest tag for a module
get_latest_tag() {
    local module="$1"
    git tag -l "$module/v*" | sort -V | tail -1 | sed "s|^$module/||"
}

# Function to bump version
bump_version() {
    local version="$1"
    local bump_type="$2"

    # Remove 'v' prefix if present
    version="${version#v}"

    # Split version into parts
    IFS='.' read -r major minor patch <<< "$version"

    case "$bump_type" in
        major)
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        minor)
            minor=$((minor + 1))
            patch=0
            ;;
        patch)
            patch=$((patch + 1))
            ;;
    esac

    echo "v${major}.${minor}.${patch}"
}

# Handle --bump mode
if [ -n "$BUMP" ]; then
    if [ -n "$FROM_VERSION" ]; then
        # Validate FROM_VERSION format
        if [[ ! "$FROM_VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
            echo -e "${RED}Error: Invalid --from version format. Expected vX.Y.Z (e.g., v0.2.2)${NC}"
            exit 1
        fi
        VERSION=$(bump_version "$FROM_VERSION" "$BUMP_TYPE")
    else
        # We'll determine VERSION per-module later based on each module's latest tag
        VERSION=""
    fi
elif [ -z "$VERSION" ]; then
    echo -e "${RED}Error: Version is required (or use --bump to auto-increment).${NC}"
    echo "Run '$0 --help' for usage."
    exit 1
fi

# Validate version format (only if VERSION is set)
if [ -n "$VERSION" ] && [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}Error: Invalid version format. Expected vX.Y.Z (e.g., v0.3.0)${NC}"
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
        echo -e "${RED}Error: Module '$MODULE' not found (no go.mod in $MODULE/)${NC}"
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
    echo -e "${YELLOW}No modules found.${NC}"
    exit 0
fi

if [ "$REMOVE" -eq 1 ]; then
    echo -e "${BLUE}Removing tags for version: ${VERSION}${NC}"
    [ -n "$MODULE" ] && echo -e "${BLUE}Module: ${MODULE}${NC}"
    echo "================================"
    for module in "${MODULES[@]}"; do
        TAG_NAME="$module/$VERSION"
        if git rev-parse -q --verify "refs/tags/$TAG_NAME" >/dev/null; then
            echo -e "${YELLOW}Removing tag: ${TAG_NAME}${NC}"
            git tag -d "$TAG_NAME"
        else
            echo -e "${BLUE}Tag not found, skipping: ${TAG_NAME}${NC}"
        fi
    done
    echo ""
    echo -e "${GREEN}Tags removed locally!${NC}"
    if [ -n "$PUSH" ]; then
        echo ""
        echo -e "${BLUE}Pushing deletions to remote...${NC}"
        for module in "${MODULES[@]}"; do
            git push origin --delete "$module/$VERSION" 2>/dev/null || true
        done
        echo -e "${GREEN}✅ Remote tags deleted${NC}"
    else
        echo ""
        echo "To delete remote tags, run:"
        for module in "${MODULES[@]}"; do
            echo "  git push origin --delete $module/$VERSION"
        done
    fi
else
    # Tag creation mode
    if [ -n "$BUMP" ] && [ -z "$VERSION" ]; then
        # Bump mode with auto-detection
        echo -e "${BLUE}Auto-bumping version (${BUMP_TYPE})${NC}"
        [ -n "$MODULE" ] && echo -e "${BLUE}Module: ${MODULE}${NC}"
        echo "================================"

        CREATED_TAGS=()
        for module in "${MODULES[@]}"; do
            # Get latest tag for this module
            LATEST=$(get_latest_tag "$module")

            if [ -z "$LATEST" ]; then
                echo -e "${YELLOW}No existing tags for $module, skipping${NC}"
                echo -e "${YELLOW}Hint: Create initial tag with: $0 --module $module v0.1.0${NC}"
                continue
            fi

            # Bump the version
            NEW_VERSION=$(bump_version "$LATEST" "$BUMP_TYPE")
            TAG_NAME="$module/$NEW_VERSION"

            if git rev-parse -q --verify "refs/tags/$TAG_NAME" >/dev/null; then
                echo -e "${YELLOW}[$module] Tag already exists: $LATEST → $NEW_VERSION${NC}"
            else
                echo -e "${GREEN}[$module] $LATEST → $NEW_VERSION${NC}"
                git tag "$TAG_NAME"
                CREATED_TAGS+=("$TAG_NAME")
            fi
        done

        if [ ${#CREATED_TAGS[@]} -eq 0 ]; then
            echo ""
            echo -e "${YELLOW}No new tags created${NC}"
        else
            echo ""
            echo -e "${GREEN}✅ Created ${#CREATED_TAGS[@]} tag(s) successfully!${NC}"

            if [ -n "$PUSH" ]; then
                echo ""
                echo -e "${BLUE}Pushing tags to remote...${NC}"
                for tag in "${CREATED_TAGS[@]}"; do
                    git push origin "$tag"
                done
                echo -e "${GREEN}✅ Tags pushed to remote${NC}"
            else
                echo ""
                echo "To push tags, run:"
                for tag in "${CREATED_TAGS[@]}"; do
                    echo "  git push origin $tag"
                done
            fi
        fi
    else
        # Standard tag creation with explicit version
        echo -e "${BLUE}Creating tags for version: ${VERSION}${NC}"
        [ -n "$MODULE" ] && echo -e "${BLUE}Module: ${MODULE}${NC}"
        echo "================================"

        CREATED_TAGS=()
        for module in "${MODULES[@]}"; do
            TAG_NAME="$module/$VERSION"
            if git rev-parse -q --verify "refs/tags/$TAG_NAME" >/dev/null; then
                echo -e "${YELLOW}Tag already exists, skipping: ${TAG_NAME}${NC}"
            else
                echo -e "${GREEN}Creating tag: ${TAG_NAME}${NC}"
                git tag "$TAG_NAME"
                CREATED_TAGS+=("$TAG_NAME")
            fi
        done

        echo ""
        echo -e "${GREEN}✅ Tags created successfully!${NC}"
        if [ -n "$PUSH" ]; then
            echo ""
            echo -e "${BLUE}Pushing tags to remote...${NC}"
            for tag in "${CREATED_TAGS[@]}"; do
                git push origin "$tag"
            done
            echo -e "${GREEN}✅ Tags pushed to remote${NC}"
        else
            echo ""
            echo "To push tags, run:"
            for tag in "${CREATED_TAGS[@]}"; do
                echo "  git push origin $tag"
            done
            echo ""
            echo "Or use: $0 --push $VERSION"
        fi
    fi
fi
