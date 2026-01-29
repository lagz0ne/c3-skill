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

# NOTE: Routing behavior:
# - Provision questions → c3-provision skill
# - Query questions → c3-navigator agent
# - Alter questions → c3-orchestrator agent
# - Ref questions → c3-ref skill
# - Audit → c3 skill
# - No .c3/ → onboard skill

# Provision routing tests → c3-provision skill
run_test "c3-provision" "$PROMPTS_DIR/provision-new.txt"
run_test "c3-provision" "$PROMPTS_DIR/provision-design.txt"
run_test "c3-provision" "$PROMPTS_DIR/provision-plan.txt"

# Query routing tests → c3-navigator agent
run_test "c3-navigator" "$PROMPTS_DIR/query-where-simple.txt"
run_test "c3-navigator" "$PROMPTS_DIR/query-where-cross.txt"
run_test "c3-navigator" "$PROMPTS_DIR/query-how.txt"
run_test "c3-navigator" "$PROMPTS_DIR/query-explain.txt"
run_test "c3-navigator" "$PROMPTS_DIR/query-what.txt"
run_test "c3-navigator" "$PROMPTS_DIR/query-show.txt"

# Alter routing tests → c3-orchestrator agent
run_test "c3-orchestrator" "$PROMPTS_DIR/alter-add.txt"
run_test "c3-orchestrator" "$PROMPTS_DIR/alter-modify.txt"
run_test "c3-orchestrator" "$PROMPTS_DIR/alter-refactor.txt"
run_test "c3-orchestrator" "$PROMPTS_DIR/alter-fix.txt"
run_test "c3-orchestrator" "$PROMPTS_DIR/alter-implement.txt"

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
