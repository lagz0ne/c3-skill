#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PLUGIN_DIR="$(dirname "$SCRIPT_DIR")"
FIXTURE_DIR="$SCRIPT_DIR/fixtures/adr-violations"

# Create temp dir and copy fixture
TEMP_DIR=$(mktemp -d)
cp -r "$FIXTURE_DIR/." "$TEMP_DIR/"

echo "=== Test: Orchestrator with ADR Audit Gate ==="
echo "Fixture: $TEMP_DIR"
echo "Plugin: $PLUGIN_DIR"
echo ""

cd "$TEMP_DIR"

# Test that orchestrator mentions the audit step
echo "--- Test: Check orchestrator knows about audit ---"
cat << 'EOF' | claude -p --dangerously-skip-permissions --plugin-dir "$PLUGIN_DIR"
I want to understand the c3-orchestrator workflow.

1. Read the orchestrator agent definition
2. Tell me: What happens between ADR generation and user acceptance? Is there an audit step?

Just summarize the workflow phases briefly.
EOF

echo ""
echo "--- Test complete ---"

# Cleanup
cd /
rm -rf "$TEMP_DIR"
echo ""
echo "=== Done ==="
