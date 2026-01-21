#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PLUGIN_DIR="$(dirname "$SCRIPT_DIR")"
FIXTURE_DIR="$SCRIPT_DIR/fixtures/adr-violations"

# Create temp dir and copy fixture
TEMP_DIR=$(mktemp -d)
cp -r "$FIXTURE_DIR/." "$TEMP_DIR/"

echo "=== Test: c3-adr-auditor agent ==="
echo "Fixture: $TEMP_DIR"
echo "Plugin: $PLUGIN_DIR"
echo ""

cd "$TEMP_DIR"

# Test using the agent
echo "--- Test: Dispatch auditor agent for OAuth ADR (expect FAIL) ---"
cat << 'EOF' | claude -p --dangerously-skip-permissions --plugin-dir "$PLUGIN_DIR"
Use the c3-adr-auditor agent to audit this ADR:

ADR Path: .c3/adr/adr-20260121-add-oauth-to-tasks.md

Dispatch the agent and return its verdict.
EOF

echo ""
echo "--- Test complete ---"

# Cleanup
cd /
rm -rf "$TEMP_DIR"
echo ""
echo "=== Done ==="
