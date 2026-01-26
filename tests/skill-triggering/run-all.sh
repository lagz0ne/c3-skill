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
