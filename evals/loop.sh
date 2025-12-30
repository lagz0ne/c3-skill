#!/bin/bash
#
# C3 Skill Eval Loop
# Interactive improvement loop with feedback
#
# Usage: ./loop.sh <scenario> <codebase> [options]
#
# Options:
#   --target <score>    Target score to achieve (default: 100)
#   --max-iter <n>      Maximum iterations (default: 5)
#   --budget <usd>      Budget per iteration (default: 3.00)
#   --auto              Auto-improve without prompting (uses Claude to fix skill)
#
# Examples:
#   ./loop.sh adopt spree
#   ./loop.sh adopt spree --target 80 --max-iter 3

set -e

SCENARIO="${1:?Usage: loop.sh <scenario> <codebase>}"
CODEBASE="${2:?Usage: loop.sh <scenario> <codebase>}"
shift 2

TARGET_SCORE=100
MAX_ITER=5
BUDGET="3.00"
AUTO_MODE=false

while [[ $# -gt 0 ]]; do
  case $1 in
    --target) TARGET_SCORE="$2"; shift 2 ;;
    --max-iter) MAX_ITER="$2"; shift 2 ;;
    --budget) BUDGET="$2"; shift 2 ;;
    --auto) AUTO_MODE=true; shift ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

ITERATION=0
BEST_SCORE=0
BEST_RUN=""

echo "╔════════════════════════════════════════════════════════════════╗"
echo "║                    C3 Skill Eval Loop                          ║"
echo "╠════════════════════════════════════════════════════════════════╣"
echo "║ Scenario:     $SCENARIO"
echo "║ Codebase:     $CODEBASE"
echo "║ Target Score: $TARGET_SCORE%"
echo "║ Max Iters:    $MAX_ITER"
echo "║ Auto Mode:    $AUTO_MODE"
echo "╚════════════════════════════════════════════════════════════════╝"
echo ""

while [[ $ITERATION -lt $MAX_ITER ]]; do
  ITERATION=$((ITERATION + 1))

  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "  Iteration $ITERATION of $MAX_ITER"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo ""

  # Run eval
  OUTPUT_DIR="./results/loop-$(date +%Y%m%d-%H%M%S)-iter$ITERATION"
  echo "→ Running eval..."
  ./runner.sh "$SCENARIO" "$CODEBASE" --skip-rubric --budget "$BUDGET" --output "$OUTPUT_DIR" > /dev/null 2>&1 || true

  # Analyze results
  echo "→ Analyzing results..."
  ANALYSIS=$(bun scripts/analyze-results.ts "$OUTPUT_DIR")

  SCORE=$(echo "$ANALYSIS" | jq -r '.score')
  PASSED=$(echo "$ANALYSIS" | jq -r '.passed')
  TOTAL=$(echo "$ANALYSIS" | jq -r '.total')
  SUMMARY=$(echo "$ANALYSIS" | jq -r '.summary')

  echo ""
  echo "┌────────────────────────────────────────────────────────────────┐"
  echo "│ Score: $SCORE% ($PASSED/$TOTAL checks passed)"
  echo "│ $SUMMARY"
  echo "└────────────────────────────────────────────────────────────────┘"
  echo ""

  # Track best
  if [[ $SCORE -gt $BEST_SCORE ]]; then
    BEST_SCORE=$SCORE
    BEST_RUN=$OUTPUT_DIR
  fi

  # Check if target reached
  if [[ $SCORE -ge $TARGET_SCORE ]]; then
    echo "✓ Target score reached!"
    echo ""
    echo "Results: $OUTPUT_DIR"
    exit 0
  fi

  # Show failures
  echo "Failures:"
  echo "$ANALYSIS" | jq -r '.failures[] | "  ✗ \(.id): \(.reason)"'
  echo ""

  # Show suggested actions
  echo "Suggested fixes:"
  echo "$ANALYSIS" | jq -r '.next_actions[] | "  → \(.)"'
  echo ""

  # Check if more iterations
  if [[ $ITERATION -ge $MAX_ITER ]]; then
    echo "Maximum iterations reached."
    break
  fi

  if [[ "$AUTO_MODE" == "true" ]]; then
    # Auto-improve mode: use Claude to fix the skill
    echo "→ Auto-improving skill..."

    FAILURES_JSON=$(echo "$ANALYSIS" | jq -c '.failures')
    NEXT_ACTIONS=$(echo "$ANALYSIS" | jq -r '.next_actions | join("; ")')

    # Read current skill
    SKILL_CONTENT=$(cat workbench/skills/c3/SKILL.md)
    STRUCTURE_GUIDE=$(cat workbench/references/structure-guide.md 2>/dev/null || echo "Not found")

    # Use Claude to improve
    IMPROVE_PROMPT="You are improving a C3 skill based on eval feedback.

Current skill:
\`\`\`markdown
$SKILL_CONTENT
\`\`\`

Structure guide for reference:
\`\`\`markdown
$STRUCTURE_GUIDE
\`\`\`

Eval failures:
$FAILURES_JSON

Suggested fixes:
$NEXT_ACTIONS

Update the skill to fix these issues. Output ONLY the complete updated SKILL.md content, nothing else."

    IMPROVED=$(claude -p --model sonnet --max-budget-usd 0.50 "$IMPROVE_PROMPT" 2>/dev/null || echo "")

    if [[ -n "$IMPROVED" && "$IMPROVED" != *"Error"* ]]; then
      echo "$IMPROVED" > workbench/skills/c3/SKILL.md
      echo "  Skill updated."
    else
      echo "  Auto-improve failed. Manual intervention needed."
      break
    fi

  else
    # Interactive mode: ask user
    echo "Options:"
    echo "  [r] Re-run eval (after manual skill edits)"
    echo "  [v] View skill file"
    echo "  [e] Edit skill file (opens \$EDITOR)"
    echo "  [a] Auto-fix with Claude"
    echo "  [q] Quit"
    echo ""
    read -p "Choice: " CHOICE

    case $CHOICE in
      r|R)
        echo "Re-running..."
        ;;
      v|V)
        echo ""
        cat workbench/skills/c3/SKILL.md
        echo ""
        read -p "Press Enter to continue..."
        ITERATION=$((ITERATION - 1))  # Don't count as iteration
        ;;
      e|E)
        ${EDITOR:-vim} workbench/skills/c3/SKILL.md
        ;;
      a|A)
        AUTO_MODE=true
        ITERATION=$((ITERATION - 1))  # Re-run this iteration in auto mode
        ;;
      q|Q)
        echo "Quitting."
        break
        ;;
      *)
        echo "Invalid choice. Re-running..."
        ;;
    esac
  fi

  echo ""
done

echo ""
echo "════════════════════════════════════════════════════════════════════"
echo "  Loop Complete"
echo "════════════════════════════════════════════════════════════════════"
echo ""
echo "  Best Score: $BEST_SCORE%"
echo "  Best Run:   $BEST_RUN"
echo ""
echo "  To view results:"
echo "    cat $BEST_RUN/summary.json"
echo ""
echo "  To continue improving:"
echo "    1. Edit workbench/skills/c3/SKILL.md"
echo "    2. Run: ./loop.sh $SCENARIO $CODEBASE"
echo ""
