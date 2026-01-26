#!/usr/bin/env bash
# Create test fixture from acountee project
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SOURCE_REPO="$HOME/dev/acountee"
FIXTURES_DIR="$REPO_ROOT/tests/fixtures"
TEMP_DIR=$(mktemp -d)

echo "Creating test fixture from $SOURCE_REPO..."

# Verify source exists
if [ ! -d "$SOURCE_REPO" ]; then
    echo "ERROR: Source repo not found at $SOURCE_REPO"
    exit 1
fi

if [ ! -d "$SOURCE_REPO/.c3" ]; then
    echo "ERROR: Source repo does not have .c3/ directory"
    exit 1
fi

# Shallow clone
echo "Step 1: Shallow clone..."
git clone --depth 1 "file://$SOURCE_REPO" "$TEMP_DIR/acountee"

# Remove unnecessary files
echo "Step 2: Cleaning up..."
cd "$TEMP_DIR/acountee"
rm -rf .git
rm -rf node_modules
rm -rf dist
rm -rf .turbo
rm -rf coverage
rm -rf .next
rm -rf .output
rm -rf .pnpm-store
rm -rf .bun
find . -name "*.log" -delete 2>/dev/null || true
find . -name ".DS_Store" -delete 2>/dev/null || true
find . -type f -size +1M -delete 2>/dev/null || true

# Create archive (from within the directory so files are at root)
echo "Step 3: Creating archive..."
mkdir -p "$FIXTURES_DIR"
tar -czf "$FIXTURES_DIR/acountee-base.tar.gz" .

# Cleanup
cd "$REPO_ROOT"
rm -rf "$TEMP_DIR"

# Report
echo ""
echo "Fixture created: $FIXTURES_DIR/acountee-base.tar.gz"
ls -lh "$FIXTURES_DIR/acountee-base.tar.gz"

# Size check (cross-platform)
SIZE=$(wc -c < "$FIXTURES_DIR/acountee-base.tar.gz" | tr -d ' ')
if [ "$SIZE" -gt 10485760 ]; then
    echo "WARNING: Fixture is larger than 10MB ($SIZE bytes). Consider removing more files."
fi

echo ""
echo "Contents (verifying .c3/ is at root):"
tar -tzf "$FIXTURES_DIR/acountee-base.tar.gz" | grep "\.c3/" | head -10
if ! tar -tzf "$FIXTURES_DIR/acountee-base.tar.gz" | grep -q "\.c3/README.md"; then
    echo "ERROR: .c3/README.md not found at expected path in archive!"
    exit 1
fi
echo "..."
echo ""
echo "Fixture verification passed."
