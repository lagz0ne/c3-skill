#!/bin/bash
#
# C3 Skill Evaluation Runner
# Usage: ./runner.sh <scenario> <codebase> [options]
#
# Options:
#   --workbench <path>    Path to workbench plugin (default: ./workbench)
#   --output <path>       Output directory (default: ./results/run-TIMESTAMP)
#   --model <model>       Model to use (default: sonnet)
#   --budget <usd>        Max budget in USD (default: 0.50)
#   --skip-rubric         Skip Phase 2 rubric evaluation
#
# Examples:
#   ./runner.sh adopt sample-api
#   ./runner.sh adjust sample-api --model opus --budget 1.00

set -e

# --- Parse arguments ---
SCENARIO="${1:?Usage: runner.sh <scenario> <codebase>}"
CODEBASE="${2:?Usage: runner.sh <scenario> <codebase>}"
shift 2

WORKBENCH="./workbench"
OUTPUT_DIR="./results/run-$(date +%Y%m%d-%H%M%S)"
MODEL="sonnet"
BUDGET="0.50"
SKIP_RUBRIC=false

while [[ $# -gt 0 ]]; do
  case $1 in
    --workbench) WORKBENCH="$2"; shift 2 ;;
    --output) OUTPUT_DIR="$2"; shift 2 ;;
    --model) MODEL="$2"; shift 2 ;;
    --budget) BUDGET="$2"; shift 2 ;;
    --skip-rubric) SKIP_RUBRIC=true; shift ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# --- Validate inputs ---
SCENARIO_FILE="scenarios/${SCENARIO}.yaml"
if [[ ! -f "$SCENARIO_FILE" ]]; then
  echo "Error: Scenario not found: $SCENARIO_FILE"
  exit 1
fi

CODEBASE_DIR="codebases/${CODEBASE}"
if [[ ! -d "$CODEBASE_DIR" ]]; then
  echo "Error: Codebase not found: $CODEBASE_DIR"
  exit 1
fi

if [[ ! -d "$WORKBENCH" ]]; then
  echo "Error: Workbench not found: $WORKBENCH"
  exit 1
fi

# --- Setup ---
echo "=== C3 Skill Evaluation ==="
echo "Scenario:  $SCENARIO"
echo "Codebase:  $CODEBASE"
echo "Model:     $MODEL"
echo "Output:    $OUTPUT_DIR"
echo ""

mkdir -p "$OUTPUT_DIR"

# Export for hooks
export EVAL_OUTPUT_DIR="$(realpath "$OUTPUT_DIR")"
export EVAL_SCENARIO="$SCENARIO"
export EVAL_CODEBASE="$CODEBASE"

# Create temp workspace
WORKSPACE=$(mktemp -d)
trap "rm -rf $WORKSPACE" EXIT

cp -r "$CODEBASE_DIR" "$WORKSPACE/project"

# Copy references to workspace so skill can access them
if [[ -d "$WORKBENCH/references" ]]; then
  cp -r "$WORKBENCH/references" "$WORKSPACE/project/references"
fi

# Helper function to parse YAML
parse_yaml() {
  bun scripts/parse-yaml.ts "$1" "$2"
}

# Run scenario setup commands
echo "=== Setup ==="
while IFS= read -r cmd; do
  if [[ -n "$cmd" && "$cmd" != "null" ]]; then
    echo "Running: $cmd"
    (cd "$WORKSPACE/project" && eval "$cmd")
  fi
done < <(parse_yaml "$SCENARIO_FILE" ".setup[]" 2>/dev/null || true)

# Get prompt
PROMPT=$(parse_yaml "$SCENARIO_FILE" ".prompt")

# --- Execute ---
echo ""
echo "=== Running Skill ==="
WORKBENCH_ABS="$(realpath "$WORKBENCH")"

cd "$WORKSPACE/project"
claude -p \
  --output-format json \
  --no-session-persistence \
  --plugin-dir "$WORKBENCH_ABS" \
  --model "$MODEL" \
  --max-budget-usd "$BUDGET" \
  --dangerously-skip-permissions \
  "$PROMPT" > "$EVAL_OUTPUT_DIR/conversation.json" 2>&1 || true
cd "$SCRIPT_DIR"

# --- Snapshot ---
echo ""
echo "=== Capturing Output ==="
if [[ -d "$WORKSPACE/project/.c3" ]]; then
  cp -r "$WORKSPACE/project/.c3" "$OUTPUT_DIR/c3-output"
  find "$OUTPUT_DIR/c3-output" -type f -name "*.md" | sort > "$OUTPUT_DIR/file-list.txt"
  echo "Captured $(wc -l < "$OUTPUT_DIR/file-list.txt") files"
else
  mkdir -p "$OUTPUT_DIR/c3-output"
  touch "$OUTPUT_DIR/file-list.txt"
  echo "Warning: No .c3/ directory created"
fi

# --- Phase 1: Checklist ---
echo ""
echo "=== Phase 1: Checklist ==="
CHECKLIST=$(parse_yaml "$SCENARIO_FILE" ".checklist")
bun scripts/check.ts "checklists/$CHECKLIST" "$OUTPUT_DIR/c3-output" > "$OUTPUT_DIR/checklist-result.json"

PASSED=$(jq '.passed' "$OUTPUT_DIR/checklist-result.json")
FAILED=$(jq '.failed' "$OUTPUT_DIR/checklist-result.json")
TOTAL=$(jq '.total' "$OUTPUT_DIR/checklist-result.json")

echo "Checklist: $PASSED/$TOTAL passed"

if [[ "$FAILED" -gt 0 ]]; then
  echo ""
  echo "Failed checks:"
  jq -r '.checks[] | select(.status == "fail") | "  - \(.id): \(.reason)"' "$OUTPUT_DIR/checklist-result.json"
fi

# --- Phase 2: Rubric ---
if [[ "$SKIP_RUBRIC" == "true" ]]; then
  echo ""
  echo "=== Phase 2: Skipped ==="
  echo '{"skipped": true, "reason": "user_request"}' > "$OUTPUT_DIR/rubric-result.json"
elif [[ "$FAILED" -gt 0 ]]; then
  echo ""
  echo "=== Phase 2: Skipped (checklist failed) ==="
  echo '{"skipped": true, "reason": "checklist_failed"}' > "$OUTPUT_DIR/rubric-result.json"
else
  echo ""
  echo "=== Phase 2: Rubric Evaluation ==="
  bun scripts/evaluate.ts "$OUTPUT_DIR" > "$OUTPUT_DIR/rubric-result.json"

  SCORE=$(jq '.percentage' "$OUTPUT_DIR/rubric-result.json")
  echo "Quality Score: ${SCORE}%"
fi

# --- Summary ---
echo ""
echo "=== Summary ==="
jq -n \
  --arg scenario "$SCENARIO" \
  --arg codebase "$CODEBASE" \
  --arg model "$MODEL" \
  --arg timestamp "$(date -Iseconds)" \
  --slurpfile checklist "$OUTPUT_DIR/checklist-result.json" \
  --slurpfile rubric "$OUTPUT_DIR/rubric-result.json" \
  '{
    scenario: $scenario,
    codebase: $codebase,
    model: $model,
    timestamp: $timestamp,
    checklist: $checklist[0],
    rubric: $rubric[0]
  }' > "$OUTPUT_DIR/summary.json"

echo "Results saved to: $OUTPUT_DIR"
echo ""

# Print quick summary
if [[ "$FAILED" -eq 0 ]]; then
  if [[ "$SKIP_RUBRIC" == "false" ]]; then
    SCORE=$(jq '.percentage' "$OUTPUT_DIR/rubric-result.json")
    echo "Result: PASS (Checklist: $PASSED/$TOTAL, Quality: ${SCORE}%)"
  else
    echo "Result: PASS (Checklist: $PASSED/$TOTAL)"
  fi
else
  echo "Result: FAIL (Checklist: $PASSED/$TOTAL, $FAILED failed)"
fi
