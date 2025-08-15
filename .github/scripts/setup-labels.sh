#!/bin/bash
# Script to set up GitHub labels using GitHub CLI
# Prerequisites: GitHub CLI (gh) must be installed and authenticated

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo -e "${RED}Error: GitHub CLI (gh) is not installed${NC}"
    echo "Install it from: https://cli.github.com/"
    exit 1
fi

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo -e "${RED}Error: Not in a git repository${NC}"
    exit 1
fi

# Get repository name
REPO=$(gh repo view --json nameWithOwner -q .nameWithOwner)

if [ -z "$REPO" ]; then
    echo -e "${RED}Error: Could not determine repository name${NC}"
    echo "Make sure you're authenticated with GitHub CLI: gh auth login"
    exit 1
fi

echo -e "${GREEN}Setting up labels for repository: $REPO${NC}"
echo ""

# Read labels from JSON file
LABELS_FILE=".github/labels.json"

if [ ! -f "$LABELS_FILE" ]; then
    echo -e "${RED}Error: $LABELS_FILE not found${NC}"
    exit 1
fi

# Function to create or update a label
create_or_update_label() {
    local name=$1
    local color=$2
    local description=$3
    
    # Check if label exists
    if gh label list --repo "$REPO" --json name | grep -q "\"$name\""; then
        echo -e "${YELLOW}Updating label: $name${NC}"
        gh label edit "$name" \
            --repo "$REPO" \
            --color "$color" \
            --description "$description" 2>/dev/null || true
    else
        echo -e "${GREEN}Creating label: $name${NC}"
        gh label create "$name" \
            --repo "$REPO" \
            --color "$color" \
            --description "$description" 2>/dev/null || true
    fi
}

# Parse JSON and create/update labels
echo "Processing labels..."
echo ""

# Use jq if available, otherwise fall back to python
if command -v jq &> /dev/null; then
    # Using jq
    while IFS= read -r label; do
        name=$(echo "$label" | jq -r '.name')
        color=$(echo "$label" | jq -r '.color')
        description=$(echo "$label" | jq -r '.description')
        create_or_update_label "$name" "$color" "$description"
    done < <(jq -c '.[]' "$LABELS_FILE")
elif command -v python3 &> /dev/null; then
    # Using python as fallback
    python3 -c "
import json
import subprocess

with open('$LABELS_FILE', 'r') as f:
    labels = json.load(f)
    
for label in labels:
    print(f\"Processing: {label['name']}\")
    subprocess.run([
        'bash', '-c',
        f\"$(declare -f create_or_update_label); create_or_update_label '{label['name']}' '{label['color']}' '{label['description']}'\"
    ])
"
else
    echo -e "${RED}Error: Neither jq nor python3 is installed${NC}"
    echo "Install one of them to parse the JSON file"
    exit 1
fi

echo ""
echo -e "${GREEN}✅ Labels setup completed!${NC}"
echo ""
echo "To view all labels, run:"
echo "  gh label list --repo $REPO"
