# C3 Skill Testing Framework Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create a comprehensive test framework for c3-skill that validates skill triggering, routing correctness, and workflow behavior using real Claude CLI execution.

**Architecture:** Bash-based test framework modeled after superpowers' tests. Uses archived project fixture (shallow git clone of acountee), extracts per-test to temp directory, runs Claude CLI with plugin loaded from within the project directory (using `cd`), and validates output via grep/assertions.

**Tech Stack:** Bash, Claude CLI, tar/gzip, git

**Pre-Implementation Verification Required (Task 0):**
1. Verify Claude CLI JSON output format for skill/agent invocations
2. Test fixture extraction structure
3. Confirm working directory behavior with `cd` vs `--add-dir`

---

## Task 0: Pre-Implementation Verification

**Files:**
- None (verification only)

**Step 1: Verify Claude CLI JSON output format**

```bash
# Run a test prompt and capture JSON output
claude -p "list available skills" \
    --plugin-dir /home/lagz0ne/c3-design \
    --dangerously-skip-permissions \
    --max-turns 2 \
    --output-format stream-json 2>&1 | head -100 > /tmp/claude-json-sample.json

# Check actual field names for skill invocations
grep -E '"skill"|"Skill"|"name"' /tmp/claude-json-sample.json | head -20

# Document the actual format (may be different from assumptions)
# Expected patterns to verify:
# - Skill tool: Look for "name":"Skill" and adjacent "skill":"..." field
# - Agent dispatch: Look for "name":"Task" and adjacent "subagent_type":"..." field
```

**Step 2: Test working directory behavior**

```bash
# Create a temp directory with .c3/
TEST_DIR=$(mktemp -d)
mkdir -p "$TEST_DIR/.c3"
echo "# Test" > "$TEST_DIR/.c3/README.md"

# Test 1: Using --add-dir (may NOT set CWD)
claude -p "does .c3/README.md exist? just answer yes or no" \
    --add-dir "$TEST_DIR" \
    --dangerously-skip-permissions \
    --max-turns 1

# Test 2: Using cd (SHOULD set CWD correctly)
(cd "$TEST_DIR" && claude -p "does .c3/README.md exist? just answer yes or no" \
    --dangerously-skip-permissions \
    --max-turns 1)

rm -rf "$TEST_DIR"

# If Test 1 fails and Test 2 passes, we MUST use cd approach
```

**Step 3: Document findings**

Create `tests/JSON-FORMAT.md` documenting:
- Actual JSON field names for Skill tool invocations
- Actual JSON field names for Task tool (agent) invocations
- Whether `cd` or `--add-dir` is required for CWD context

---

## Task 1: Create Test Directory Structure

**Files:**
- Create: `tests/test-helpers.sh`
- Create: `tests/run-tests.sh`
- Create: `tests/skill-triggering/run-test.sh`
- Create: `tests/skill-triggering/run-all.sh`

**Step 1: Create test-helpers.sh with core functions**

```bash
#!/usr/bin/env bash
# Helper functions for C3 skill tests

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGIN_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
FIXTURES_DIR="$SCRIPT_DIR/fixtures"

# Extract fixture to temp directory
# Usage: test_dir=$(extract_fixture)
extract_fixture() {
    local test_dir=$(mktemp -d)
    tar -xzf "$FIXTURES_DIR/acountee-base.tar.gz" -C "$test_dir"
    echo "$test_dir"
}

# Extract fixture without .c3/ (for onboard testing)
# Usage: test_dir=$(extract_fixture_no_c3)
extract_fixture_no_c3() {
    local test_dir=$(mktemp -d)
    tar -xzf "$FIXTURES_DIR/acountee-base.tar.gz" -C "$test_dir"
    rm -rf "$test_dir/.c3"
    echo "$test_dir"
}

# Cleanup test directory (handles various temp paths: /tmp, /var/folders, /run)
# Usage: cleanup_fixture "$test_dir"
cleanup_fixture() {
    local test_dir="$1"
    if [ -d "$test_dir" ] && [[ "$test_dir" == /tmp/* || "$test_dir" == /var/folders/* || "$test_dir" == /run/* ]]; then
        rm -rf "$test_dir"
    fi
}

# Run Claude with a prompt in a project directory
# IMPORTANT: Uses cd to set working directory (--add-dir alone doesn't set CWD)
# Usage: run_claude "prompt" "$project_dir" [timeout] > output.json
run_claude() {
    local prompt="$1"
    local project_dir="$2"
    local timeout="${3:-120}"

    # Must cd into project_dir so Claude sees .c3/ in working directory
    (cd "$project_dir" && timeout "$timeout" claude -p "$prompt" \
        --plugin-dir "$PLUGIN_DIR" \
        --dangerously-skip-permissions \
        --max-turns 5 \
        --output-format stream-json 2>&1)
}

# Check if a skill was triggered
# Usage: assert_skill_triggered "$log_file" "skill-name" "test name"
assert_skill_triggered() {
    local log_file="$1"
    local skill_name="$2"
    local test_name="${3:-test}"

    # Match skill:"skillname" or skill:"namespace:skillname"
    if grep -qE '"skill":"([^"]*:)?'"$skill_name"'"' "$log_file"; then
        echo "  [PASS] $test_name - Skill '$skill_name' triggered"
        return 0
    else
        echo "  [FAIL] $test_name - Skill '$skill_name' NOT triggered"
        echo "  Skills triggered:"
        grep -o '"skill":"[^"]*"' "$log_file" 2>/dev/null | sort -u | sed 's/^/    /' || echo "    (none)"
        return 1
    fi
}

# Check if an agent was triggered (via Task tool)
# Usage: assert_agent_triggered "$log_file" "agent-name" "test name"
assert_agent_triggered() {
    local log_file="$1"
    local agent_name="$2"
    local test_name="${3:-test}"

    # Match subagent_type:"c3-skill:agent-name"
    if grep -qE '"subagent_type":"c3-skill:'"$agent_name"'"' "$log_file"; then
        echo "  [PASS] $test_name - Agent '$agent_name' triggered"
        return 0
    else
        echo "  [FAIL] $test_name - Agent '$agent_name' NOT triggered"
        echo "  Agents triggered:"
        grep -o '"subagent_type":"[^"]*"' "$log_file" 2>/dev/null | sort -u | sed 's/^/    /' || echo "    (none)"
        return 1
    fi
}

# Check if output contains a pattern
# Usage: assert_contains "$output" "pattern" "test name"
assert_contains() {
    local output="$1"
    local pattern="$2"
    local test_name="${3:-test}"

    if echo "$output" | grep -q "$pattern"; then
        echo "  [PASS] $test_name"
        return 0
    else
        echo "  [FAIL] $test_name - Pattern not found: $pattern"
        return 1
    fi
}

# Check that either skill OR agent was triggered (routing flexibility)
# Usage: assert_routed_to "$log_file" "c3-query" "c3-navigator" "test name"
assert_routed_to() {
    local log_file="$1"
    local skill_name="$2"
    local agent_name="$3"
    local test_name="${4:-test}"

    if grep -qE '"skill":"([^"]*:)?'"$skill_name"'"' "$log_file"; then
        echo "  [PASS] $test_name - Routed to skill '$skill_name'"
        return 0
    elif grep -qE '"subagent_type":"c3-skill:'"$agent_name"'"' "$log_file"; then
        echo "  [PASS] $test_name - Routed to agent '$agent_name'"
        return 0
    else
        echo "  [FAIL] $test_name - Not routed to '$skill_name' or '$agent_name'"
        echo "  Skills triggered:"
        grep -o '"skill":"[^"]*"' "$log_file" 2>/dev/null | sort -u | sed 's/^/    /' || echo "    (none)"
        echo "  Agents triggered:"
        grep -o '"subagent_type":"[^"]*"' "$log_file" 2>/dev/null | sort -u | sed 's/^/    /' || echo "    (none)"
        return 1
    fi
}

# Export functions
export -f extract_fixture
export -f extract_fixture_no_c3
export -f cleanup_fixture
export -f run_claude
export -f assert_skill_triggered
export -f assert_agent_triggered
export -f assert_contains
export -f assert_routed_to
export SCRIPT_DIR PLUGIN_DIR FIXTURES_DIR
```

**Step 2: Create run-tests.sh main runner**

```bash
#!/usr/bin/env bash
# Main test runner for C3 skill tests
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "========================================"
echo " C3 Skill Test Suite"
echo "========================================"
echo ""
echo "Plugin: $(cd "$SCRIPT_DIR/.." && pwd)"
echo "Date: $(date)"
echo "Claude: $(claude --version 2>/dev/null || echo 'not found')"
echo ""

# Check prerequisites
if ! command -v claude &> /dev/null; then
    echo "ERROR: Claude CLI not found"
    exit 1
fi

if [ ! -f "$SCRIPT_DIR/fixtures/acountee-base.tar.gz" ]; then
    echo "ERROR: Fixture not found. Run: ./scripts/create-test-fixture.sh"
    exit 1
fi

# Parse args
VERBOSE=false
SPECIFIC_TEST=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --verbose|-v) VERBOSE=true; shift ;;
        --test|-t) SPECIFIC_TEST="$2"; shift 2 ;;
        --help|-h)
            echo "Usage: $0 [options]"
            echo "  --verbose, -v    Show detailed output"
            echo "  --test, -t NAME  Run specific test only"
            exit 0
            ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

# Run test suites
passed=0
failed=0

run_suite() {
    local suite_dir="$1"
    local suite_name=$(basename "$suite_dir")

    if [ ! -f "$suite_dir/run-all.sh" ]; then
        echo "Skipping $suite_name (no run-all.sh)"
        return
    fi

    echo "----------------------------------------"
    echo "Suite: $suite_name"
    echo "----------------------------------------"

    if bash "$suite_dir/run-all.sh"; then
        passed=$((passed + 1))
    else
        failed=$((failed + 1))
    fi
}

# Run each test suite
for suite in "$SCRIPT_DIR"/*/; do
    if [ -d "$suite" ] && [ "$(basename "$suite")" != "fixtures" ]; then
        run_suite "$suite"
    fi
done

# Summary
echo ""
echo "========================================"
echo " Results"
echo "========================================"
echo "  Passed: $passed"
echo "  Failed: $failed"
echo ""

if [ $failed -gt 0 ]; then
    echo "STATUS: FAILED"
    exit 1
else
    echo "STATUS: PASSED"
    exit 0
fi
```

**Step 3: Create skill-triggering/run-test.sh**

```bash
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

# Try skill first, then agent
if grep -qE '"skill":"([^"]*:)?'"$EXPECTED"'"' "$LOG_FILE" 2>/dev/null; then
    echo "[PASS] Skill '$EXPECTED' triggered"
    exit 0
elif grep -qE '"subagent_type":"c3-skill:'"$EXPECTED"'"' "$LOG_FILE" 2>/dev/null; then
    echo "[PASS] Agent '$EXPECTED' triggered"
    exit 0
else
    echo "[FAIL] Neither skill nor agent '$EXPECTED' triggered"
    echo ""
    echo "Skills triggered:"
    grep -o '"skill":"[^"]*"' "$LOG_FILE" 2>/dev/null | sort -u || echo "  (none)"
    echo "Agents triggered:"
    grep -o '"subagent_type":"[^"]*"' "$LOG_FILE" 2>/dev/null | sort -u || echo "  (none)"
    exit 1
fi
```

**Step 4: Create skill-triggering/run-all.sh**

```bash
#!/usr/bin/env bash
# Run all skill triggering tests
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROMPTS_DIR="$SCRIPT_DIR/prompts"

echo "=== Skill Triggering Tests ==="
echo ""

passed=0
failed=0

run_test() {
    local expected="$1"
    local prompt_file="$2"
    local use_no_c3="${3:-false}"

    if bash "$SCRIPT_DIR/run-test.sh" "$expected" "$prompt_file" "$use_no_c3"; then
        passed=$((passed + 1))
    else
        failed=$((failed + 1))
    fi
    echo ""
}

# NOTE: Routing may go through c3 (router) first, then to sub-skills.
# Tests accept either direct skill match OR the router skill (c3) as a pass.
# The key behavior is that the c3 ecosystem handles the request.

# Query routing tests → c3-query or c3-navigator (or c3 router)
run_test "c3-query" "$PROMPTS_DIR/query-where-simple.txt"
run_test "c3-query" "$PROMPTS_DIR/query-where-cross.txt"
run_test "c3-query" "$PROMPTS_DIR/query-how.txt"
run_test "c3-query" "$PROMPTS_DIR/query-explain.txt"
run_test "c3-query" "$PROMPTS_DIR/query-what.txt"
run_test "c3-query" "$PROMPTS_DIR/query-show.txt"

# Alter routing tests → c3-alter or c3-orchestrator
run_test "c3-alter" "$PROMPTS_DIR/alter-add.txt"
run_test "c3-alter" "$PROMPTS_DIR/alter-modify.txt"
run_test "c3-alter" "$PROMPTS_DIR/alter-refactor.txt"
run_test "c3-alter" "$PROMPTS_DIR/alter-fix.txt"
run_test "c3-alter" "$PROMPTS_DIR/alter-implement.txt"

# Ref routing tests → c3-ref
run_test "c3-ref" "$PROMPTS_DIR/ref-convention.txt"
run_test "c3-ref" "$PROMPTS_DIR/ref-standard.txt"
run_test "c3-ref" "$PROMPTS_DIR/ref-pattern.txt"

# Audit tests → c3 (stays in main skill)
run_test "c3" "$PROMPTS_DIR/audit-explicit.txt"
run_test "c3" "$PROMPTS_DIR/audit-check.txt"
run_test "c3" "$PROMPTS_DIR/audit-validate.txt"

# Onboard test → onboard (no .c3/)
run_test "onboard" "$PROMPTS_DIR/onboard-no-c3.txt" "true"

# Summary
echo "========================================"
echo "Skill Triggering Results: $passed passed, $failed failed"
echo "========================================"

[ $failed -eq 0 ]
```

**Step 5: Make scripts executable and commit**

```bash
chmod +x tests/test-helpers.sh
chmod +x tests/run-tests.sh
chmod +x tests/skill-triggering/run-test.sh
chmod +x tests/skill-triggering/run-all.sh
git add tests/
git commit -m "feat(tests): add test framework structure and helpers"
```

---

## Task 2: Create Test Fixture

**Files:**
- Create: `scripts/create-test-fixture.sh`
- Create: `tests/fixtures/.gitkeep`
- Create: `tests/fixtures/.gitignore`

**Step 1: Create fixture creation script**

```bash
#!/usr/bin/env bash
# Create test fixture from acountee project
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SOURCE_REPO="$HOME/dev/acountee"
FIXTURES_DIR="$REPO_ROOT/tests/fixtures"
TEMP_DIR=$(mktemp -d)

echo "Creating test fixture from $SOURCE_REPO..."

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
find . -name "*.log" -delete
find . -name ".DS_Store" -delete
find . -type f -size +1M -delete  # Remove files > 1MB

# Create archive
echo "Step 3: Creating archive..."
mkdir -p "$FIXTURES_DIR"
tar -czf "$FIXTURES_DIR/acountee-base.tar.gz" .

# Cleanup
rm -rf "$TEMP_DIR"

# Report
echo ""
echo "Fixture created: $FIXTURES_DIR/acountee-base.tar.gz"
ls -lh "$FIXTURES_DIR/acountee-base.tar.gz"

# Size check
SIZE=$(stat -f%z "$FIXTURES_DIR/acountee-base.tar.gz" 2>/dev/null || stat -c%s "$FIXTURES_DIR/acountee-base.tar.gz")
if [ "$SIZE" -gt 10485760 ]; then
    echo "WARNING: Fixture is larger than 10MB ($SIZE bytes). Consider removing more files."
fi

echo ""
echo "Contents (verifying .c3/ is at root):"
tar -tzf "$FIXTURES_DIR/acountee-base.tar.gz" | grep "^\.c3/" | head -10
if ! tar -tzf "$FIXTURES_DIR/acountee-base.tar.gz" | grep -q "^\.c3/README.md"; then
    echo "ERROR: .c3/README.md not found at expected path in archive!"
    exit 1
fi
echo "..."
```

**Step 2: Create fixtures directory with gitignore**

Create `tests/fixtures/.gitignore`:
```
# Generated fixtures - not committed
acountee-base.tar.gz
```

Create `tests/fixtures/.gitkeep`:
```
# Placeholder for fixtures directory
```

**Step 3: Run fixture creation**

```bash
chmod +x scripts/create-test-fixture.sh
./scripts/create-test-fixture.sh
```

**Step 4: Verify fixture**

```bash
# Should show fixture file
ls -lh tests/fixtures/acountee-base.tar.gz

# Should extract successfully
TEST_DIR=$(mktemp -d)
tar -xzf tests/fixtures/acountee-base.tar.gz -C "$TEST_DIR"
ls "$TEST_DIR/.c3/"  # Should show README.md, c3-1-*, c3-2-*, etc.
rm -rf "$TEST_DIR"
```

**Step 5: Commit fixture script**

```bash
git add scripts/create-test-fixture.sh tests/fixtures/.gitkeep tests/fixtures/.gitignore
git commit -m "feat(tests): add fixture creation script"
```

---

## Task 3: Create Test Prompts - Query Category

**Files:**
- Create: `tests/skill-triggering/prompts/query-where-simple.txt`
- Create: `tests/skill-triggering/prompts/query-where-cross.txt`
- Create: `tests/skill-triggering/prompts/query-how.txt`
- Create: `tests/skill-triggering/prompts/query-explain.txt`
- Create: `tests/skill-triggering/prompts/query-what.txt`
- Create: `tests/skill-triggering/prompts/query-show.txt`

**Step 1: Create all query prompt files**

`query-where-simple.txt`:
```
where is authentication handled in this project?
```

`query-where-cross.txt`:
```
where does the frontend call the API backend? show me the connection points
```

`query-how.txt`:
```
how does the flow pattern work in this codebase?
```

`query-explain.txt`:
```
explain the middleware stack and how requests are processed
```

`query-what.txt`:
```
what components are in the API container? list them all
```

`query-show.txt`:
```
show me the system architecture overview
```

**Step 2: Commit prompts**

```bash
mkdir -p tests/skill-triggering/prompts
git add tests/skill-triggering/prompts/query-*.txt
git commit -m "feat(tests): add query routing test prompts"
```

---

## Task 4: Create Test Prompts - Alter Category

**Files:**
- Create: `tests/skill-triggering/prompts/alter-add.txt`
- Create: `tests/skill-triggering/prompts/alter-modify.txt`
- Create: `tests/skill-triggering/prompts/alter-refactor.txt`
- Create: `tests/skill-triggering/prompts/alter-fix.txt`
- Create: `tests/skill-triggering/prompts/alter-implement.txt`

**Step 1: Create all alter prompt files**

`alter-add.txt`:
```
add rate limiting to the API backend
```

`alter-modify.txt`:
```
modify the auth middleware to support API keys in addition to OAuth
```

`alter-refactor.txt`:
```
refactor the payment flows to extract common validation logic
```

`alter-fix.txt`:
```
fix the notification system to retry on failure with exponential backoff
```

`alter-implement.txt`:
```
implement caching for the query services to improve performance
```

**Step 2: Commit prompts**

```bash
git add tests/skill-triggering/prompts/alter-*.txt
git commit -m "feat(tests): add alter routing test prompts"
```

---

## Task 5: Create Test Prompts - Ref, Audit, and Onboard Categories

**Files:**
- Create: `tests/skill-triggering/prompts/ref-convention.txt`
- Create: `tests/skill-triggering/prompts/ref-standard.txt`
- Create: `tests/skill-triggering/prompts/ref-pattern.txt`
- Create: `tests/skill-triggering/prompts/audit-explicit.txt`
- Create: `tests/skill-triggering/prompts/audit-check.txt`
- Create: `tests/skill-triggering/prompts/audit-validate.txt`
- Create: `tests/skill-triggering/prompts/onboard-no-c3.txt`

**Step 1: Create all remaining prompt files**

`ref-convention.txt`:
```
document our error handling convention - what patterns do we use?
```

`ref-standard.txt`:
```
what ref documents our query services pattern?
```

`ref-pattern.txt`:
```
show me the refs - what patterns are documented?
```

`audit-explicit.txt`:
```
run C3 audit on this project
```

`audit-check.txt`:
```
check if the C3 documentation is up to date with the code
```

`audit-validate.txt`:
```
validate the architecture docs
```

`onboard-no-c3.txt`:
```
help me document this project's architecture using C3
```

**Step 2: Commit prompts**

```bash
git add tests/skill-triggering/prompts/ref-*.txt
git add tests/skill-triggering/prompts/audit-*.txt
git add tests/skill-triggering/prompts/onboard-*.txt
git commit -m "feat(tests): add ref, audit, and onboard test prompts"
```

---

## Task 6: Integration Test - Run Full Suite

**Files:**
- None (verification only)

**Step 1: Verify fixture exists**

```bash
ls -lh tests/fixtures/acountee-base.tar.gz
# Expected: file exists, ~2-10MB
```

**Step 2: Run single test to verify framework**

```bash
cd /home/lagz0ne/c3-design
./tests/skill-triggering/run-test.sh c3-query tests/skill-triggering/prompts/query-where-simple.txt
# Expected: [PASS] or clear error message
```

**Step 3: Run full test suite**

```bash
./tests/run-tests.sh --verbose
# Expected: All tests run, see pass/fail for each
```

**Step 4: Document any failures**

If tests fail, create issue list:
- Which prompts failed?
- What was triggered instead?
- Is it a skill trigger issue or test expectation issue?

**Step 5: Final commit if all pass**

```bash
git add -A
git commit -m "feat(tests): complete c3 skill test framework v1"
```

---

## Summary

| Task | Description | Files |
|------|-------------|-------|
| 1 | Test directory structure | test-helpers.sh, run-tests.sh, skill-triggering/*.sh |
| 2 | Create test fixture | scripts/create-test-fixture.sh |
| 3 | Query prompts (6) | prompts/query-*.txt |
| 4 | Alter prompts (5) | prompts/alter-*.txt |
| 5 | Ref/Audit/Onboard prompts (7) | prompts/ref-*.txt, audit-*.txt, onboard-*.txt |
| 6 | Integration test | Verify all works |

**Total: 18 test prompts across 5 routing categories**
