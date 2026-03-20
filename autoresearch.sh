#!/usr/bin/env bash
set -euo pipefail
# Autoresearch benchmark: count signals in skill docs that lead LLM to use file tools

SKILL_DIR="skills/c3"

# 1. File paths to .c3/ docs (LLM sees path → tries Read tool)
# Exclude structural references (directory layout, file structure descriptions)
FILE_PATHS=$(grep -rn '\.c3/.*\.md\|\.c3/.*\.yaml' "$SKILL_DIR" --include="*.md" \
  | grep -v 'code-map\.yaml\|c3\.db\|directory\|layout\|structure\|tree\|File Structure\|\.c3/ directory' \
  | wc -l | tr -d ' ')

# 2. 'Read' instructions that don't specify c3x
READ_WITHOUT_C3X=$(grep -rn -i 'read the\|read doc\|read file\|Read tool' "$SKILL_DIR" --include="*.md" \
  | grep -v 'c3x read\|c3x lookup\|already read' \
  | wc -l | tr -d ' ')

# 3. References to browsing/walking/scanning .c3/ (implies Glob usage)
BROWSE_SIGNALS=$(grep -rn -i 'walk.*\.c3\|browse.*\.c3\|scan.*\.c3\|Glob.*\.c3\|glob.*c3\|No manual Glob\|manual.*Read' "$SKILL_DIR" --include="*.md" \
  | wc -l | tr -d ' ')

# 4. Edit/Write on .c3/ files (should be c3x write/set)
EDIT_SIGNALS=$(grep -rn -i 'Edit.*\.c3/\|Write.*\.c3/' "$SKILL_DIR" --include="*.md" \
  | grep -v 'NEVER\|do not\|HARD RULE\|Edit, Write' \
  | wc -l | tr -d ' ')

TOTAL=$((FILE_PATHS + READ_WITHOUT_C3X + BROWSE_SIGNALS + EDIT_SIGNALS))

echo "METRIC total_signals=$TOTAL"
echo "METRIC file_paths=$FILE_PATHS"
echo "METRIC read_without_c3x=$READ_WITHOUT_C3X"
echo "METRIC browse_signals=$BROWSE_SIGNALS"
echo "METRIC edit_signals=$EDIT_SIGNALS"
