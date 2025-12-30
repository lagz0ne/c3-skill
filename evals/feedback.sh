#!/bin/bash
#
# Quick feedback on latest or specified eval run
#
# Usage:
#   ./feedback.sh                    # Analyze latest run
#   ./feedback.sh <results-dir>      # Analyze specific run

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

if [[ -n "$1" ]]; then
  RESULTS_DIR="$1"
else
  # Find latest results
  RESULTS_DIR=$(ls -td results/*/  2>/dev/null | head -1)
  if [[ -z "$RESULTS_DIR" ]]; then
    echo "No results found. Run an eval first:"
    echo "  ./runner.sh adopt spree --skip-rubric"
    exit 1
  fi
fi

if [[ ! -f "$RESULTS_DIR/checklist-result.json" ]]; then
  echo "No checklist results in $RESULTS_DIR"
  exit 1
fi

echo "Analyzing: $RESULTS_DIR"
echo ""

# Run analysis
ANALYSIS=$(bun scripts/analyze-results.ts "$RESULTS_DIR")

SCORE=$(echo "$ANALYSIS" | jq -r '.score')
PASSED=$(echo "$ANALYSIS" | jq -r '.passed')
TOTAL=$(echo "$ANALYSIS" | jq -r '.total')
SUMMARY=$(echo "$ANALYSIS" | jq -r '.summary')

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
fi

# Show what was created
if [[ -f "$RESULTS_DIR/file-list.txt" ]]; then
  FILE_COUNT=$(wc -l < "$RESULTS_DIR/file-list.txt")
  echo "Files created: $FILE_COUNT"
  echo "$ANALYSIS" | jq -r empty  # just to use jq
  head -10 "$RESULTS_DIR/file-list.txt" | sed 's|^|  |'
  if [[ $FILE_COUNT -gt 10 ]]; then
    echo "  ... and $((FILE_COUNT - 10)) more"
  fi
fi
