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
# NOTE: --verbose is required with --output-format stream-json
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
        --output-format stream-json \
        --verbose 2>&1)
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
