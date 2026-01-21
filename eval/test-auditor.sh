#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PLUGIN_DIR="$(dirname "$SCRIPT_DIR")"
FIXTURE_DIR="$SCRIPT_DIR/fixtures/adr-violations"

# Create temp dir and copy fixture
TEMP_DIR=$(mktemp -d)
cp -r "$FIXTURE_DIR/." "$TEMP_DIR/"

echo "=== Test: Audit ADR for architectural violations ==="
echo "Fixture: $TEMP_DIR"
echo "Plugin: $PLUGIN_DIR"
echo ""

cd "$TEMP_DIR"

# Test 1: OAuth violation (should FAIL)
echo "--- Test 1: OAuth in task routes (expect FAIL) ---"
cat << 'EOF' | claude -p --dangerously-skip-permissions --plugin-dir "$PLUGIN_DIR"
Audit the ADR at .c3/adr/adr-20260121-add-oauth-to-tasks.md

Check if this ADR violates C3 architectural principles:
- Does the change respect layer hierarchy?
- Does c3-103 have authority to add auth logic?

Read the ADR, then read the affected component (c3-103) and the container doc (c3-1-api).
Output: PASS or FAIL with explanation.
EOF
echo ""
echo "--- Test 1 complete ---"
echo ""

# Test 2: Valid ADR (should PASS)
echo "--- Test 2: Add archive (expect PASS) ---"
cat << 'EOF' | claude -p --dangerously-skip-permissions --plugin-dir "$PLUGIN_DIR"
Audit the ADR at .c3/adr/adr-20260121-add-archive.md

Check if this ADR violates C3 architectural principles:
- Does the change respect layer hierarchy?
- Is archiving within c3-103 task routes responsibility?

Read the ADR, then read the affected component (c3-103) and the container doc (c3-1-api).
Output: PASS or FAIL with explanation.
EOF
echo ""
echo "--- Test 2 complete ---"
echo ""

# Test 3: Orchestration violation (should FAIL)
echo "--- Test 3: Task orchestration (expect FAIL) ---"
cat << 'EOF' | claude -p --dangerously-skip-permissions --plugin-dir "$PLUGIN_DIR"
Audit the ADR at .c3/adr/adr-20260121-task-orchestration.md

Check if this ADR violates C3 architectural principles:
- Do components hand-off or orchestrate? (hand-off is correct)
- Does the decision use orchestration language like "coordinate" or "manage flow"?

Read the ADR, then read the container doc (c3-1-api) which states the composition rule.
Output: PASS or FAIL with explanation.
EOF
echo ""
echo "--- Test 3 complete ---"
echo ""

# Test 4: Context override without justification (should FAIL)
echo "--- Test 4: Replace JWT (expect FAIL) ---"
cat << 'EOF' | claude -p --dangerously-skip-permissions --plugin-dir "$PLUGIN_DIR"
Audit the ADR at .c3/adr/adr-20260121-replace-jwt.md

Check if this ADR violates C3 architectural principles:
- Does it contradict context-level decisions?
- If contradicting, is there a Pattern Overrides section with justification?

Read the ADR, then read the context doc (.c3/README.md) which has Key Decisions.
Output: PASS or FAIL with explanation.
EOF
echo ""
echo "--- Test 4 complete ---"

# Cleanup
cd /
rm -rf "$TEMP_DIR"
echo ""
echo "=== All tests complete ==="
