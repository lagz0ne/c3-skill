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
