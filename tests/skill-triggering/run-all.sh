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

# NOTE: Current skill/agent names (v4):
# - c3-change skill / c3-lead agent  (replaces c3-alter, c3-provision, c3-orchestrator)
# - c3-query skill / c3-navigator agent
# - c3-ref skill
# - c3-audit skill
# - c3-onboard skill

# Provision routing tests → c3-change skill (provision mode) or c3-provision (installed plugin)
run_test "c3-change,c3-provision" "$PROMPTS_DIR/provision-new.txt"
run_test "c3-change,c3-provision" "$PROMPTS_DIR/provision-design.txt"
run_test "c3-change,c3-provision" "$PROMPTS_DIR/provision-plan.txt"

# Query routing tests → c3-query skill or c3-navigator agent
run_test "c3-query,c3-navigator" "$PROMPTS_DIR/query-where-simple.txt"
run_test "c3-query,c3-navigator" "$PROMPTS_DIR/query-where-cross.txt"
run_test "c3-query,c3-navigator" "$PROMPTS_DIR/query-how.txt"
run_test "c3-query,c3-navigator" "$PROMPTS_DIR/query-explain.txt"
run_test "c3-query,c3-navigator" "$PROMPTS_DIR/query-what.txt"
run_test "c3-query,c3-navigator" "$PROMPTS_DIR/query-show.txt"

# Change routing tests → c3-change skill or c3-lead agent (also accepts c3-alter from installed plugin)
run_test "c3-change,c3-lead,c3-alter" "$PROMPTS_DIR/alter-add.txt"
run_test "c3-change,c3-lead,c3-alter" "$PROMPTS_DIR/alter-modify.txt"
run_test "c3-change,c3-lead,c3-alter" "$PROMPTS_DIR/alter-refactor.txt"
run_test "c3-change,c3-lead,c3-alter" "$PROMPTS_DIR/alter-fix.txt"
run_test "c3-change,c3-lead,c3-alter" "$PROMPTS_DIR/alter-implement.txt"

# Ref routing tests → c3-ref (may route to c3-navigator which then routes to c3-ref)
run_test "c3-ref,c3-navigator" "$PROMPTS_DIR/ref-convention.txt"
run_test "c3-ref,c3-navigator" "$PROMPTS_DIR/ref-standard.txt"
run_test "c3-ref,c3-navigator" "$PROMPTS_DIR/ref-pattern.txt"

# Audit tests → c3-audit skill (also accepts c3 router which dispatches to c3-audit)
run_test "c3-audit,c3" "$PROMPTS_DIR/audit-explicit.txt"
run_test "c3-audit,c3" "$PROMPTS_DIR/audit-check.txt"
run_test "c3-audit,c3" "$PROMPTS_DIR/audit-validate.txt"

# Onboard test → c3-onboard (no .c3/)
run_test "c3-onboard" "$PROMPTS_DIR/onboard-no-c3.txt" "true"

# Summary
echo "========================================"
echo "Skill Triggering Results: $passed passed, $failed failed"
echo "========================================"

[ $failed -eq 0 ]
