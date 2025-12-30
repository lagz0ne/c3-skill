#!/bin/bash
#
# Main eval command - syncs from main and runs eval
# Persists state for resumption after compaction
#
# Usage:
#   ./eval.sh adopt spree           # Run eval
#   ./eval.sh status                # Show current progress
#   ./eval.sh dump                  # Dump full context for review
#   ./eval.sh reset                 # Reset state, start fresh
#
# Workflow:
#   1. Edit skills/agents in main repo
#   2. Run: ./eval.sh adopt spree
#   3. Review feedback
#   4. Edit main, repeat

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

STATE_FILE="state.json"
BUDGET="${BUDGET:-3.00}"

# Initialize state if not exists
init_state() {
  if [[ ! -f "$STATE_FILE" ]]; then
    echo '{
  "scenario": null,
  "codebase": null,
  "iteration": 0,
  "best_score": 0,
  "best_run": null,
  "runs": [],
  "started_at": null,
  "last_run_at": null
}' > "$STATE_FILE"
  fi
}

# Update state
update_state() {
  local key="$1"
  local value="$2"
  local tmp=$(mktemp)
  jq --arg k "$key" --arg v "$value" '.[$k] = $v' "$STATE_FILE" > "$tmp" && mv "$tmp" "$STATE_FILE"
}

update_state_num() {
  local key="$1"
  local value="$2"
  local tmp=$(mktemp)
  jq --arg k "$key" --argjson v "$value" '.[$k] = $v' "$STATE_FILE" > "$tmp" && mv "$tmp" "$STATE_FILE"
}

add_run() {
  local run_dir="$1"
  local score="$2"
  local tmp=$(mktemp)
  jq --arg dir "$run_dir" --argjson score "$score" '.runs += [{"dir": $dir, "score": $score, "at": now | todate}]' "$STATE_FILE" > "$tmp" && mv "$tmp" "$STATE_FILE"
}

# Commands
cmd_status() {
  init_state

  echo "╔════════════════════════════════════════════════════════════════╗"
  echo "║                    Eval Status                                 ║"
  echo "╚════════════════════════════════════════════════════════════════╝"
  echo ""

  local scenario=$(jq -r '.scenario // "not set"' "$STATE_FILE")
  local codebase=$(jq -r '.codebase // "not set"' "$STATE_FILE")
  local iteration=$(jq -r '.iteration // 0' "$STATE_FILE")
  local best_score=$(jq -r '.best_score // 0' "$STATE_FILE")
  local best_run=$(jq -r '.best_run // "none"' "$STATE_FILE")
  local num_runs=$(jq '.runs | length' "$STATE_FILE")

  echo "Scenario:    $scenario"
  echo "Codebase:    $codebase"
  echo "Iterations:  $iteration"
  echo "Best Score:  $best_score%"
  echo "Total Runs:  $num_runs"
  echo ""

  if [[ "$best_run" != "none" && "$best_run" != "null" ]]; then
    echo "Best Run: $best_run"
    if [[ -f "$best_run/checklist-result.json" ]]; then
      local passed=$(jq '.passed' "$best_run/checklist-result.json")
      local total=$(jq '.total' "$best_run/checklist-result.json")
      echo "         $passed/$total checks passed"
    fi
  fi

  echo ""
  echo "Recent runs:"
  jq -r '.runs | .[-5:] | reverse | .[] | "  \(.at): \(.score)% - \(.dir)"' "$STATE_FILE" 2>/dev/null || echo "  (none)"
}

cmd_dump() {
  init_state

  local best_run=$(jq -r '.best_run // empty' "$STATE_FILE")
  local dump_file="dump-$(date +%Y%m%d-%H%M%S).md"

  echo "# Eval Dump - $(date)" > "$dump_file"
  echo "" >> "$dump_file"

  # State
  echo "## State" >> "$dump_file"
  echo '```json' >> "$dump_file"
  cat "$STATE_FILE" >> "$dump_file"
  echo '```' >> "$dump_file"
  echo "" >> "$dump_file"

  # Current skill
  echo "## Current Skill (workbench)" >> "$dump_file"
  echo '```markdown' >> "$dump_file"
  cat workbench/skills/c3/SKILL.md >> "$dump_file" 2>/dev/null || echo "(not found)"
  echo '```' >> "$dump_file"
  echo "" >> "$dump_file"

  # Best run results
  if [[ -n "$best_run" && -d "$best_run" ]]; then
    echo "## Best Run Analysis" >> "$dump_file"
    echo "" >> "$dump_file"

    if [[ -f "$best_run/checklist-result.json" ]]; then
      echo "### Checklist Results" >> "$dump_file"
      echo '```json' >> "$dump_file"
      cat "$best_run/checklist-result.json" >> "$dump_file"
      echo '```' >> "$dump_file"
      echo "" >> "$dump_file"
    fi

    echo "### Analysis" >> "$dump_file"
    echo '```' >> "$dump_file"
    bun scripts/analyze-results.ts "$best_run" >> "$dump_file" 2>/dev/null || echo "(analysis failed)"
    echo '```' >> "$dump_file"
    echo "" >> "$dump_file"

    # Sample output
    if [[ -f "$best_run/c3-output/README.md" ]]; then
      echo "### Context Output (README.md)" >> "$dump_file"
      echo '```markdown' >> "$dump_file"
      head -100 "$best_run/c3-output/README.md" >> "$dump_file"
      echo '```' >> "$dump_file"
    fi
  fi

  echo ""
  echo "Dumped to: $dump_file"
  echo ""
  echo "This file contains:"
  echo "  - Current state and history"
  echo "  - Current skill content"
  echo "  - Best run analysis"
  echo "  - Sample output"
  echo ""
  echo "Use this to resume after compaction."
}

cmd_reset() {
  echo "Resetting eval state..."
  rm -f "$STATE_FILE"
  init_state
  echo "State reset. Ready for fresh start."
}

cmd_run() {
  local scenario="$1"
  local codebase="$2"

  if [[ -z "$scenario" || -z "$codebase" ]]; then
    echo "Usage: ./eval.sh <scenario> <codebase>"
    echo ""
    echo "Examples:"
    echo "  ./eval.sh adopt spree"
    echo "  ./eval.sh status"
    echo "  ./eval.sh dump"
    echo "  ./eval.sh reset"
    exit 1
  fi

  init_state

  # Update state
  update_state "scenario" "$scenario"
  update_state "codebase" "$codebase"
  update_state "last_run_at" "$(date -Iseconds)"

  local iteration=$(jq '.iteration' "$STATE_FILE")
  iteration=$((iteration + 1))
  update_state_num "iteration" "$iteration"

  if [[ $(jq '.started_at' "$STATE_FILE") == "null" ]]; then
    update_state "started_at" "$(date -Iseconds)"
  fi

  echo "╔════════════════════════════════════════════════════════════════╗"
  echo "║  Eval Run #$iteration"
  echo "╠════════════════════════════════════════════════════════════════╣"
  echo "║  Scenario: $scenario"
  echo "║  Codebase: $codebase"
  echo "╚════════════════════════════════════════════════════════════════╝"
  echo ""

  # Sync from main
  echo "→ Syncing from main..."
  ./sync.sh > /dev/null
  echo "  Done."
  echo ""

  # Run eval
  echo "→ Running eval..."
  OUTPUT_DIR="./results/run-$(date +%Y%m%d-%H%M%S)"
  ./runner.sh "$scenario" "$codebase" --skip-rubric --budget "$BUDGET" --output "$OUTPUT_DIR" > /dev/null 2>&1 || true
  echo "  Done."
  echo ""

  # Analyze
  echo "→ Analyzing..."
  ANALYSIS=$(bun scripts/analyze-results.ts "$OUTPUT_DIR" 2>/dev/null || echo '{"score":0,"passed":0,"total":0,"failures":[],"summary":"Analysis failed","next_actions":[]}')

  SCORE=$(echo "$ANALYSIS" | jq -r '.score')
  PASSED=$(echo "$ANALYSIS" | jq -r '.passed')
  TOTAL=$(echo "$ANALYSIS" | jq -r '.total')
  SUMMARY=$(echo "$ANALYSIS" | jq -r '.summary')

  # Record run
  add_run "$OUTPUT_DIR" "$SCORE"

  # Update best
  local best_score=$(jq '.best_score' "$STATE_FILE")
  if [[ $SCORE -gt $best_score ]]; then
    update_state_num "best_score" "$SCORE"
    update_state "best_run" "$OUTPUT_DIR"
  fi

  # Display results
  echo ""
  echo "┌────────────────────────────────────────────────────────────────┐"
  echo "│ Score: $SCORE% ($PASSED/$TOTAL checks passed)"
  echo "│"
  echo "│ $SUMMARY"
  echo "└────────────────────────────────────────────────────────────────┘"
  echo ""

  if [[ $SCORE -lt 100 ]]; then
    echo "Failures:"
    echo "$ANALYSIS" | jq -r '.failures[] | "  ✗ [\(.category)] \(.id)\n    └─ \(.reason)"'
    echo ""

    echo "Next Actions:"
    echo "$ANALYSIS" | jq -r '.next_actions[] | "  → \(.)"'
    echo ""
  else
    echo "✓ All checks passed!"
    echo ""
  fi

  # Show files
  if [[ -f "$OUTPUT_DIR/file-list.txt" ]]; then
    local file_count=$(wc -l < "$OUTPUT_DIR/file-list.txt")
    echo "Files created: $file_count"
    head -5 "$OUTPUT_DIR/file-list.txt" | sed 's|^|  |'
    if [[ $file_count -gt 5 ]]; then
      echo "  ... and $((file_count - 5)) more"
    fi
    echo ""
  fi

  echo "Results: $OUTPUT_DIR"
  echo ""
  echo "Next steps:"
  echo "  1. Edit main: skills/c3/SKILL.md or references/"
  echo "  2. Run again: ./eval.sh $scenario $codebase"
  echo "  3. Check status: ./eval.sh status"
  echo "  4. Dump context: ./eval.sh dump"
}

# Main
case "${1:-}" in
  status)
    cmd_status
    ;;
  dump)
    cmd_dump
    ;;
  reset)
    cmd_reset
    ;;
  "")
    echo "Usage: ./eval.sh <command> [args]"
    echo ""
    echo "Commands:"
    echo "  <scenario> <codebase>  Run eval (e.g., ./eval.sh adopt spree)"
    echo "  status                 Show current progress"
    echo "  dump                   Dump full context for review"
    echo "  reset                  Reset state, start fresh"
    ;;
  *)
    cmd_run "$1" "$2"
    ;;
esac
