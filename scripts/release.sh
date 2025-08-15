#!/bin/bash
# Release Script for lumi-go
# Creates and publishes a new release

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
VERSION="${1:-}"
DRY_RUN="${DRY_RUN:-false}"
PUSH_REMOTE="${PUSH_REMOTE:-origin}"

# Function to print colored output
print_color() {
    local color=$1
    shift
    echo -e "${color}$*${NC}"
}

# Function to confirm action
confirm() {
    local prompt=$1
    print_color "$YELLOW" "$prompt"
    read -p "Continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_color "$RED" "Aborted by user"
        exit 1
    fi
}

# Show usage
if [ -z "$VERSION" ]; then
    print_color "$BLUE" "========================================="
    print_color "$BLUE" "  lumi-go Release Script"
    print_color "$BLUE" "========================================="
    echo ""
    print_color "$CYAN" "Usage: $0 <version> [DRY_RUN=true]"
    echo ""
    print_color "$YELLOW" "Examples:"
    echo "  $0 v0.0.2                    # Create and push release v0.0.2"
    echo "  DRY_RUN=true $0 v0.0.2      # Dry run without pushing"
    echo ""
    print_color "$YELLOW" "Current tags:"
    git tag -l | tail -5
    echo ""
    exit 1
fi

# Header
print_color "$BLUE" "========================================="
print_color "$BLUE" "  Creating Release $VERSION"
print_color "$BLUE" "========================================="
echo ""

# Validate version format
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    print_color "$RED" "Error: Version must be in format vX.Y.Z (e.g., v0.0.2)"
    exit 1
fi

# Check if we're on main branch
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ] && [ "$CURRENT_BRANCH" != "master" ]; then
    print_color "$YELLOW" "Warning: Not on main branch (current: $CURRENT_BRANCH)"
    confirm "Do you want to continue?"
fi

# Check for uncommitted changes
if [ -n "$(git status --porcelain)" ]; then
    print_color "$RED" "Error: You have uncommitted changes"
    git status --short
    echo ""
    confirm "Do you want to continue anyway?"
fi

# Check if tag already exists
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    print_color "$RED" "Error: Tag $VERSION already exists"
    print_color "$YELLOW" "To delete it: git tag -d $VERSION && git push origin :refs/tags/$VERSION"
    exit 1
fi

# Run validation
print_color "$CYAN" "\n📋 Running validation checks..."
if [ -x "./scripts/validate-fresh.sh" ]; then
    if ./scripts/validate-fresh.sh > /tmp/release-validation.log 2>&1; then
        print_color "$GREEN" "✓ Validation passed"
    else
        print_color "$RED" "✗ Validation failed"
        print_color "$YELLOW" "Check log: /tmp/release-validation.log"
        confirm "Continue despite validation failure?"
    fi
else
    print_color "$YELLOW" "⚠ Validation script not found, skipping..."
fi

# Update version in files (if needed)
print_color "$CYAN" "\n📝 Updating version references..."

# Update Chart.yaml if it exists
if [ -f "deploy/helm/Chart.yaml" ]; then
    sed -i.bak "s/^version:.*/version: ${VERSION#v}/" deploy/helm/Chart.yaml
    sed -i.bak "s/^appVersion:.*/appVersion: \"${VERSION#v}\"/" deploy/helm/Chart.yaml
    rm deploy/helm/Chart.yaml.bak
    print_color "$GREEN" "✓ Updated Helm chart version"
fi

# Update version in go.mod comments if present
if [ -f "go.mod" ]; then
    if grep -q "// Version:" go.mod; then
        sed -i.bak "s|// Version:.*|// Version: $VERSION|" go.mod
        rm go.mod.bak
        print_color "$GREEN" "✓ Updated go.mod version comment"
    fi
fi

# Show what will be released
print_color "$CYAN" "\n📄 Release Information:"
echo "Version:    $VERSION"
echo "Branch:     $CURRENT_BRANCH"
echo "Commit:     $(git rev-parse --short HEAD)"
echo "Author:     $(git config user.name) <$(git config user.email)>"
echo "Date:       $(date -u +"%Y-%m-%d %H:%M:%S UTC")"

# Show recent commits since last tag
LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
if [ -n "$LAST_TAG" ]; then
    print_color "$CYAN" "\n📜 Commits since $LAST_TAG:"
    git log --oneline "$LAST_TAG"..HEAD | head -10
    COMMIT_COUNT=$(git rev-list --count "$LAST_TAG"..HEAD)
    if [ "$COMMIT_COUNT" -gt 10 ]; then
        print_color "$YELLOW" "... and $((COMMIT_COUNT - 10)) more commits"
    fi
else
    print_color "$CYAN" "\n📜 Recent commits:"
    git log --oneline -10
fi

# Extract release notes for this version
print_color "$CYAN" "\n📋 Release Notes:"
if [ -f "RELEASE_NOTES.md" ]; then
    # Try to extract section for this version
    awk "/## $VERSION/,/^## v[0-9]/" RELEASE_NOTES.md | head -20
    print_color "$GREEN" "✓ Release notes found"
else
    print_color "$YELLOW" "⚠ No RELEASE_NOTES.md file found"
fi

# Confirm release
echo ""
confirm "🚀 Ready to create release $VERSION?"

# Create git tag
print_color "$CYAN" "\n🏷️  Creating git tag..."
if [ "$DRY_RUN" = "true" ]; then
    print_color "$YELLOW" "[DRY RUN] Would create tag: $VERSION"
else
    # Create annotated tag with release notes
    if [ -f "RELEASE_NOTES.md" ]; then
        # Extract release notes for tag message
        RELEASE_MSG=$(awk "/## $VERSION/,/^## v[0-9]/" RELEASE_NOTES.md | head -50)
        git tag -a "$VERSION" -m "Release $VERSION

$RELEASE_MSG"
    else
        git tag -a "$VERSION" -m "Release $VERSION

Phase 1: Local Developer Experience

- Complete Docker Compose stack
- Hot-reload development
- Comprehensive documentation
- Validation scripts"
    fi
    print_color "$GREEN" "✓ Tag $VERSION created"
fi

# Push to remote
print_color "$CYAN" "\n📤 Pushing to remote..."
if [ "$DRY_RUN" = "true" ]; then
    print_color "$YELLOW" "[DRY RUN] Would push:"
    echo "  git push $PUSH_REMOTE $CURRENT_BRANCH"
    echo "  git push $PUSH_REMOTE $VERSION"
else
    # Push commits
    if git push "$PUSH_REMOTE" "$CURRENT_BRANCH"; then
        print_color "$GREEN" "✓ Pushed commits to $PUSH_REMOTE/$CURRENT_BRANCH"
    else
        print_color "$RED" "✗ Failed to push commits"
    fi

    # Push tag
    if git push "$PUSH_REMOTE" "$VERSION"; then
        print_color "$GREEN" "✓ Pushed tag $VERSION to $PUSH_REMOTE"
    else
        print_color "$RED" "✗ Failed to push tag"
        print_color "$YELLOW" "You can push manually: git push $PUSH_REMOTE $VERSION"
    fi
fi

# Build and push Docker image (optional)
if [ -f "deploy/docker/build.sh" ] && [ "$DRY_RUN" != "true" ]; then
    print_color "$CYAN" "\n🐳 Building Docker image..."
    confirm "Do you want to build and push Docker image?"
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        VERSION="$VERSION" ./deploy/docker/build.sh --push
    fi
fi

# Create GitHub release (if gh CLI is available)
if command -v gh &> /dev/null && [ "$DRY_RUN" != "true" ]; then
    print_color "$CYAN" "\n📦 Creating GitHub release..."
    confirm "Do you want to create a GitHub release?"
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        if [ -f "RELEASE_NOTES.md" ]; then
            # Extract release notes for this version
            RELEASE_NOTES=$(awk "/## $VERSION/,/^## v[0-9]/" RELEASE_NOTES.md | tail -n +2)
            echo "$RELEASE_NOTES" | gh release create "$VERSION" \
                --title "Release $VERSION" \
                --notes-file - \
                --target "$CURRENT_BRANCH"
        else
            gh release create "$VERSION" \
                --title "Release $VERSION" \
                --notes "Phase 1: Local Developer Experience - Complete" \
                --target "$CURRENT_BRANCH"
        fi
        print_color "$GREEN" "✓ GitHub release created"
    fi
fi

# Summary
print_color "$BLUE" "\n========================================="
print_color "$BLUE" "  Release Complete!"
print_color "$BLUE" "========================================="
echo ""
print_color "$GREEN" "✅ Successfully created release $VERSION"
echo ""
print_color "$CYAN" "Next steps:"
echo "1. Verify the release on GitHub: https://github.com/lumitut/lumi-go/releases/tag/$VERSION"
echo "2. Update any deployment configurations"
echo "3. Notify team members"
echo "4. Start working on next phase"
echo ""

# Update TODO.md
if [ -f "TODO.md" ] && [ "$DRY_RUN" != "true" ]; then
    print_color "$YELLOW" "Don't forget to update TODO.md:"
    echo "  - Mark current phase as complete"
    echo "  - Update release version reference"
fi

print_color "$GREEN" "🎉 Release $VERSION complete!"
