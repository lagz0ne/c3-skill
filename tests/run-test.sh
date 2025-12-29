#!/bin/bash
# C3 Skill Test Runner
# Runs tests in isolation with only the c3 plugin loaded
# Traces all file access for audit

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PLUGIN_ROOT="$(dirname "$SCRIPT_DIR")"
OUTPUT_DIR="$SCRIPT_DIR/output"
CASES_DIR="$SCRIPT_DIR/cases"
AUDIT_DIR="$SCRIPT_DIR/audit"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Usage
usage() {
    echo "Usage: $0 <test-case> [options]"
    echo ""
    echo "Options:"
    echo "  --workspace <path>   Use specific workspace (default: temp dir)"
    echo "  --no-audit           Disable file access auditing"
    echo ""
    echo "Examples:"
    echo "  $0 01-new-project"
    echo "  $0 03-add-container --workspace /tmp/test-project"
    echo ""
    echo "Available test cases:"
    ls "$CASES_DIR"/*.md 2>/dev/null | xargs -n1 basename | sed 's/.md$//'
    exit 1
}

# Parse arguments
TEST_CASE="$1"
WORKSPACE=""
AUDIT_ENABLED=1

shift || usage

while [[ $# -gt 0 ]]; do
    case $1 in
        --workspace)
            WORKSPACE="$2"
            shift 2
            ;;
        --no-audit)
            AUDIT_ENABLED=0
            shift
            ;;
        *)
            usage
            ;;
    esac
done

if [[ -z "$TEST_CASE" ]]; then
    usage
fi

TEST_FILE="$CASES_DIR/${TEST_CASE}.md"
if [[ ! -f "$TEST_FILE" ]]; then
    echo "Error: Test case not found: $TEST_FILE"
    usage
fi

# Create isolated workspace if not provided
if [[ -z "$WORKSPACE" ]]; then
    WORKSPACE=$(mktemp -d)
    echo "Created isolated workspace: $WORKSPACE"
    CLEANUP_WORKSPACE=1
fi

mkdir -p "$WORKSPACE"

# Check for fixtures and copy if exists
FIXTURE_DIR="$SCRIPT_DIR/fixtures/$TEST_CASE"
if [[ -d "$FIXTURE_DIR" ]]; then
    echo "Loading fixture from: $FIXTURE_DIR"
    cp -r "$FIXTURE_DIR/." "$WORKSPACE/"
    echo "Fixture contents:"
    find "$WORKSPACE" -type f -name "*.md" | head -10
fi

# Timestamp for this run
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

# Set up audit log
AUDIT_LOG="$OUTPUT_DIR/${TEST_CASE}-${TIMESTAMP}.audit.log"
if [[ "$AUDIT_ENABLED" == "1" ]]; then
    echo "Audit log: $AUDIT_LOG"
    echo "# C3 Test Audit Log" > "$AUDIT_LOG"
    echo "# Test: $TEST_CASE" >> "$AUDIT_LOG"
    echo "# Started: $(date)" >> "$AUDIT_LOG"
    echo "# Workspace: $WORKSPACE" >> "$AUDIT_LOG"
    echo "" >> "$AUDIT_LOG"
    export C3_AUDIT_LOG="$AUDIT_LOG"
fi

# Extract query from test file
QUERY=$(sed -n '/^## Query/,/^## Expect/p' "$TEST_FILE" | sed -n '/^```/,/^```/p' | sed '1d;$d')

# Log available skill files (for reference, not embedded)
echo "Available skill files:" >> "$AUDIT_LOG"
echo "  - agents/c3.md ($(wc -l < "$PLUGIN_ROOT/agents/c3.md") lines)" >> "$AUDIT_LOG"
echo "  - references/v3-structure.md ($(wc -l < "$PLUGIN_ROOT/references/v3-structure.md") lines)" >> "$AUDIT_LOG"
echo "  - skills/c3-structure/SKILL.md ($(wc -l < "$PLUGIN_ROOT/skills/c3-structure/SKILL.md") lines)" >> "$AUDIT_LOG"
echo "  - skills/c3-implementation/SKILL.md ($(wc -l < "$PLUGIN_ROOT/skills/c3-implementation/SKILL.md") lines)" >> "$AUDIT_LOG"
echo "" >> "$AUDIT_LOG"
echo "Agent should discover and read these naturally." >> "$AUDIT_LOG"
echo "" >> "$AUDIT_LOG"

# Create minimal test prompt - let agent discover skills naturally
TEST_PROMPT=$(cat <<EOF
You are in workspace: $WORKSPACE

The c3 plugin is available. Use it naturally as you would in a real session.

IMPORTANT: This is a non-interactive test. Make reasonable assumptions and proceed to completion. Do not ask clarifying questions.

$QUERY

## After Completion
When done, output a section:
\`\`\`results
COMPLETED: <what you did>
FILES_CREATED: <list of files>
FILES_MODIFIED: <list of files>
\`\`\`
EOF
)

# Output file
OUTPUT_FILE="$OUTPUT_DIR/${TEST_CASE}-${TIMESTAMP}.md"

echo ""
echo "================================"
echo "C3 Skill Test Runner"
echo "================================"
echo "Test:      $TEST_CASE"
echo "Workspace: $WORKSPACE"
echo "Output:    $OUTPUT_FILE"
echo "Audit:     ${AUDIT_LOG:-disabled}"
echo "================================"
echo ""

# Build claude command
# --plugin-dir: ensure our plugin is loaded
# --dangerously-skip-permissions: allow automated testing in temp workspace
# -p: print mode for non-interactive
CLAUDE_CMD="claude --plugin-dir $PLUGIN_ROOT --dangerously-skip-permissions -p"

# Add audit hooks if enabled
if [[ "$AUDIT_ENABLED" == "1" ]]; then
    # Export for the hook script
    export SKILL_ROOT="$AUDIT_DIR"

    # Create temp settings with audit hooks
    TEMP_SETTINGS=$(mktemp)
    cat > "$TEMP_SETTINGS" <<SETTINGS
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Read|Glob|Grep|Bash|Write|Edit|Task|Skill",
        "hooks": [
          {
            "type": "command",
            "command": "$AUDIT_DIR/log-access.sh"
          }
        ]
      }
    ]
  }
}
SETTINGS
    # Note: Claude Code reads hooks from project .claude/settings.json
    # For now, we'll copy it to workspace
    mkdir -p "$WORKSPACE/.claude"
    cp "$TEMP_SETTINGS" "$WORKSPACE/.claude/settings.json"
    rm "$TEMP_SETTINGS"
fi

# Run claude with isolation
cd "$WORKSPACE"
$CLAUDE_CMD "$TEST_PROMPT" 2>&1 | tee "$OUTPUT_FILE"

# Append expectations to output for comparison
echo "" >> "$OUTPUT_FILE"
echo "---" >> "$OUTPUT_FILE"
echo "# Expected Criteria (from test case)" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"
sed -n '/^## Expect/,$p' "$TEST_FILE" >> "$OUTPUT_FILE"

# Capture workspace state
echo "" >> "$OUTPUT_FILE"
echo "---" >> "$OUTPUT_FILE"
echo "# Workspace State After Test" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"
echo '```' >> "$OUTPUT_FILE"
find "$WORKSPACE" -type f \( -name "*.md" -o -name "*.yaml" \) 2>/dev/null | grep -v ".claude" | head -50 >> "$OUTPUT_FILE"
echo '```' >> "$OUTPUT_FILE"

# Finalize audit log
if [[ "$AUDIT_ENABLED" == "1" ]]; then
    echo "" >> "$AUDIT_LOG"
    echo "# Completed: $(date)" >> "$AUDIT_LOG"

    echo ""
    echo "================================"
    echo "Audit Summary"
    echo "================================"
    echo "Total operations: $(grep -c '^\[' "$AUDIT_LOG" || echo 0)"
    echo ""
    echo "By type:"
    grep -oE '^\[[^]]+\] [A-Z]+:' "$AUDIT_LOG" | sed 's/.*\] //' | sort | uniq -c | sort -rn || true
    echo ""
    echo "Files accessed outside workspace:"
    grep -E '(READ|GLOB|GREP):' "$AUDIT_LOG" | grep -v "$WORKSPACE" | grep -v "$PLUGIN_ROOT" || echo "(none)"
    echo ""
fi

echo ""
echo "================================"
echo "Test Complete"
echo "================================"
echo "Output:    $OUTPUT_FILE"
echo "Audit:     ${AUDIT_LOG:-disabled}"
echo "Workspace: $WORKSPACE"
echo ""
echo "To verify:"
echo "  ./tests/verify-output.sh $TEST_CASE $WORKSPACE"
echo ""
echo "To view audit:"
echo "  cat $AUDIT_LOG"
echo ""

# Cleanup if we created the workspace
if [[ "$CLEANUP_WORKSPACE" == "1" ]]; then
    echo "Temp workspace kept for inspection: $WORKSPACE"
    echo "Delete with: rm -rf $WORKSPACE"
fi
