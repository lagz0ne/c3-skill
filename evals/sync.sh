#!/bin/bash
#
# Sync workbench from main skills/agents/references
# Run this before each eval to test latest changes
#
# Usage: ./sync.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "Syncing workbench from main..."

# Skills
rm -rf workbench/skills/c3
mkdir -p workbench/skills/c3
cp -r ../skills/c3/* workbench/skills/c3/
echo "  ✓ skills/c3"

# Agents
rm -rf workbench/agents
mkdir -p workbench/agents
cp ../agents/c3.md workbench/agents/
echo "  ✓ agents/c3.md"

# References
rm -rf workbench/references
cp -r ../references workbench/
echo "  ✓ references/"

# Commands
rm -rf workbench/commands
mkdir -p workbench/commands
cp -r ../commands/* workbench/commands/
echo "  ✓ commands/"

echo ""
echo "Workbench synced. Ready for eval."
