#!/usr/bin/env bash
# Run a single skill triggering test
# Usage: ./run-test.sh <expected-skill-or-agent> <prompt-file>
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../test-helpers.sh"

EXPECTED="$1"
PROMPT_FILE="$2"
USE_NO_C3="${3:-false}"

if [ -z "$EXPECTED" ] || [ -z "$PROMPT_FILE" ]; then
    echo "Usage: $0 <expected-skill-or-agent> <prompt-file> [use-no-c3]"
    exit 1
fi

PROMPT=$(cat "$PROMPT_FILE")
TEST_NAME=$(basename "$PROMPT_FILE" .txt)

echo "=== Test: $TEST_NAME ==="
echo "Expected: $EXPECTED"
echo "Prompt: $PROMPT"
echo ""

# Setup
if [ "$USE_NO_C3" = "true" ]; then
    TEST_DIR=$(extract_fixture_no_c3)
else
    TEST_DIR=$(extract_fixture)
fi
LOG_FILE=$(mktemp)
trap "cleanup_fixture \"$TEST_DIR\"; rm -f \"$LOG_FILE\"" EXIT

# Run
echo "Running Claude..."
run_claude "$PROMPT" "$TEST_DIR" 180 > "$LOG_FILE" || true

# Verify
echo ""
echo "=== Results ==="

# Support comma-separated alternatives (e.g., "c3-query,c3-navigator")
IFS=',' read -ra EXPECTED_LIST <<< "$EXPECTED"
for exp in "${EXPECTED_LIST[@]}"; do
    if grep -qE '"skill":"([^"]*:)?'"$exp"'"' "$LOG_FILE" 2>/dev/null; then
        echo "[PASS] Skill '$exp' triggered"
        exit 0
    elif grep -qE '"subagent_type":"c3-skill:'"$exp"'"' "$LOG_FILE" 2>/dev/null; then
        echo "[PASS] Agent '$exp' triggered"
        exit 0
    fi
done

echo "[FAIL] None of '$EXPECTED' triggered"
echo ""
echo "Skills triggered:"
grep -o '"skill":"[^"]*"' "$LOG_FILE" 2>/dev/null | sort -u || echo "  (none)"
echo "Agents triggered:"
grep -o '"subagent_type":"[^"]*"' "$LOG_FILE" 2>/dev/null | sort -u || echo "  (none)"
exit 1
