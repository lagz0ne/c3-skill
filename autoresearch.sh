#!/usr/bin/env bash
set -uo pipefail
# Autoresearch benchmark: count signals that ENCOURAGE file tool usage on .c3/
# Excludes prohibitions (NEVER, don't, No manual) and legitimate exceptions (migrate repair)

SKILL_DIR="skills/c3"

# 1. File paths that would trigger Read (exclude directory-existence checks)
FILE_PATHS=$(grep -rn '\.c3/.*\.md' "$SKILL_DIR" --include="*.md" \
  | grep -v 'code-map\|c3\.db\|directory\|layout\|structure\|tree\|File Structure\|\.c3/ directory\|No \.c3\|c3x init\|NEVER\|do not\|migrate' \
  | wc -l | tr -d ' ')

# 2. 'Read' instructions without c3x (exclude prohibitions and migrate exception)
READ_WITHOUT_C3X=$(grep -rn -i 'read the\|read doc\|read file\|Read tool' "$SKILL_DIR" --include="*.md" \
  | grep -v 'c3x read\|c3x lookup\|already read\|NEVER\|Do not\|exception\|migrate' \
  | wc -l | tr -d ' ')

# 3. Edit/Write on .c3/ files (exclude prohibitions and migrate exception)
EDIT_SIGNALS=$(grep -rn -i 'Edit .*\.c3/\|Edit `\.c3' "$SKILL_DIR" --include="*.md" \
  | grep -v 'NEVER\|do not\|HARD RULE\|Edit, Write\|exception\|Do NOT\|migrate' \
  | wc -l | tr -d ' ')

TOTAL=$((FILE_PATHS + READ_WITHOUT_C3X + EDIT_SIGNALS))

echo "METRIC total_signals=$TOTAL"
echo "METRIC file_paths=$FILE_PATHS"
echo "METRIC read_without_c3x=$READ_WITHOUT_C3X"
echo "METRIC edit_signals=$EDIT_SIGNALS"
