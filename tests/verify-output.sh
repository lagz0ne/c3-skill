#!/bin/bash
# Verify test output against expected criteria
# Checks workspace for PASS/FAIL conditions

# Don't exit on error - we want to count failures
set +e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CASES_DIR="$SCRIPT_DIR/cases"

usage() {
    echo "Usage: $0 <test-case> <workspace-path>"
    echo ""
    echo "Verifies workspace state against test case expectations"
    exit 1
}

TEST_CASE="$1"
WORKSPACE="$2"

[[ -z "$TEST_CASE" || -z "$WORKSPACE" ]] && usage

TEST_FILE="$CASES_DIR/${TEST_CASE}.md"
[[ ! -f "$TEST_FILE" ]] && echo "Test case not found: $TEST_FILE" && exit 1
[[ ! -d "$WORKSPACE" ]] && echo "Workspace not found: $WORKSPACE" && exit 1

echo "Verifying: $TEST_CASE"
echo "Workspace: $WORKSPACE"
echo ""

PASS=0
FAIL=0

# Helper: check if file exists
check_file() {
    local pattern="$1"
    local desc="$2"
    if compgen -G "$WORKSPACE/$pattern" > /dev/null; then
        echo "✓ PASS: $desc"
        ((PASS++))
    else
        echo "✗ FAIL: $desc (missing: $pattern)"
        ((FAIL++))
    fi
}

# Helper: check if file contains pattern
check_content() {
    local file="$1"
    local pattern="$2"
    local desc="$3"
    if [[ -f "$WORKSPACE/$file" ]] && grep -q "$pattern" "$WORKSPACE/$file"; then
        echo "✓ PASS: $desc"
        ((PASS++))
    else
        echo "✗ FAIL: $desc (pattern not found: $pattern)"
        ((FAIL++))
    fi
}

# Helper: check file does NOT contain pattern (failure condition)
check_no_content() {
    local file="$1"
    local pattern="$2"
    local desc="$3"
    if [[ -f "$WORKSPACE/$file" ]] && grep -q "$pattern" "$WORKSPACE/$file"; then
        echo "✗ FAIL: $desc (prohibited pattern found: $pattern)"
        ((FAIL++))
    else
        echo "✓ PASS: $desc"
        ((PASS++))
    fi
}

echo "=== Structure Checks ==="

# Common structure checks
check_file ".c3/README.md" "Context file exists"

# Check for valid frontmatter in context
# Required per c3-structure: id, c3-version, title, summary
if [[ -f "$WORKSPACE/.c3/README.md" ]]; then
    check_content ".c3/README.md" "^id: c3-0" "Context has correct ID"
    check_content ".c3/README.md" "^c3-version:" "Context has c3-version"
    check_content ".c3/README.md" "^title:" "Context has title"
fi

# Check for container(s)
if compgen -G "$WORKSPACE/.c3/c3-[0-9]-*" > /dev/null; then
    echo "✓ PASS: At least one container folder exists"
    ((PASS++))

    # Check first container
    CONTAINER_DIR=$(ls -d "$WORKSPACE"/.c3/c3-[0-9]-* 2>/dev/null | head -1)
    if [[ -f "$CONTAINER_DIR/README.md" ]]; then
        echo "✓ PASS: Container README exists"
        ((PASS++))

        # Check container frontmatter (type can be "container", "Service", etc.)
        if grep -qE "^type:" "$CONTAINER_DIR/README.md"; then
            TYPE=$(grep -E "^type:" "$CONTAINER_DIR/README.md" | head -1 | cut -d: -f2 | tr -d ' ')
            echo "✓ PASS: Container has type: $TYPE"
            ((PASS++))
        else
            echo "✗ FAIL: Container missing type field"
            ((FAIL++))
        fi
    else
        echo "✗ FAIL: Container README missing"
        ((FAIL++))
    fi
else
    echo "✗ FAIL: No container folders found"
    ((FAIL++))
fi

echo ""
echo "=== Layer Content Checks ==="

# Context level - should NOT have component details
if [[ -f "$WORKSPACE/.c3/README.md" ]]; then
    # Should have
    check_content ".c3/README.md" "## Container" "Context has container section"

    # Should NOT have (failure conditions)
    check_no_content ".c3/README.md" '\`\`\`typescript' "Context has no TypeScript code"
    check_no_content ".c3/README.md" '\`\`\`javascript' "Context has no JavaScript code"
    check_no_content ".c3/README.md" '\`\`\`json' "Context has no JSON code"
fi

# Container level checks
if [[ -n "$CONTAINER_DIR" && -f "$CONTAINER_DIR/README.md" ]]; then
    # Should have (flexible matching)
    if grep -qE "## Component" "$CONTAINER_DIR/README.md"; then
        echo "✓ PASS: Container has component section"
        ((PASS++))
    elif grep -qE "## Components" "$CONTAINER_DIR/README.md"; then
        echo "✓ PASS: Container has component section (as '## Components')"
        ((PASS++))
    else
        echo "✗ FAIL: Container missing component section"
        ((FAIL++))
    fi

    if grep -qE "## Tech" "$CONTAINER_DIR/README.md"; then
        echo "✓ PASS: Container has tech stack section"
        ((PASS++))
    elif grep -qE "## Technology" "$CONTAINER_DIR/README.md"; then
        echo "✓ PASS: Container has tech stack section (as '## Technology')"
        ((PASS++))
    else
        echo "✗ FAIL: Container missing tech stack section"
        ((FAIL++))
    fi

    # Should NOT have
    check_no_content "$CONTAINER_DIR/README.md" '\`\`\`typescript' "Container has no TypeScript code"
    check_no_content "$CONTAINER_DIR/README.md" '\`\`\`javascript' "Container has no JavaScript code"
fi

echo ""
echo "=== Summary ==="
echo "PASS: $PASS"
echo "FAIL: $FAIL"
echo ""

if [[ $FAIL -eq 0 ]]; then
    echo "✓ All checks passed!"
    exit 0
else
    echo "✗ Some checks failed"
    exit 1
fi
