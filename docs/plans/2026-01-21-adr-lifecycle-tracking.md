# ADR Lifecycle Tracking Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Simplify the gate to check for any approved ADR, capture base-commit on acceptance, and verify/document actual changes on transition to implemented.

**Architecture:** Gate becomes a simple "approved ADR exists?" check. Base-commit captured in ADR frontmatter on acceptance. New verifier script diffs base-commit..HEAD on transition, compares against approved-files, and documents verification in ADR.

**Tech Stack:** Bash scripts, YAML frontmatter parsing, git

---

## Task 1: Simplify c3-gate Script

**Files:**
- Modify: `scripts/c3-gate`

**Step 1: Read current c3-gate implementation**

Understand the existing file-specific approval logic that will be removed.

**Step 2: Write simplified gate logic**

Replace the file-specific checking with simple "any approved ADR exists?" check:

```bash
#!/usr/bin/env bash
#
# C3 Architecture Gate (Simplified)
#
# Gates Edit/Write operations in C3-adopted projects.
# Requires an accepted ADR to exist before any code changes.
#
# Behavior:
# - If no .c3/README.md: Allow (not a C3 project)
# - If file is in .c3/: Allow (documentation)
# - If any ADR with status: accepted exists: Allow
# - Otherwise: Block with guidance
#
# Exit codes:
# - 0: Allow the operation
# - 2: Block the operation (stderr shown to user)

set -uo pipefail

# Read JSON from stdin
INPUT=$(cat)

# Parse file_path
FILE_PATH=$(echo "$INPUT" | grep -o '"file_path"[[:space:]]*:[[:space:]]*"[^"]*"' | sed 's/.*:.*"\([^"]*\)".*/\1/' 2>/dev/null || echo "")

# No file path means we can't check - allow
if [[ -z "$FILE_PATH" ]]; then
    exit 0
fi

# Get project directory
PROJECT_DIR="${CLAUDE_PROJECT_DIR:-$(pwd)}"

# Make path relative for matching
REL_PATH="${FILE_PATH#$PROJECT_DIR/}"

# Skip gate for .c3/ documentation files
if [[ "$REL_PATH" == .c3/* ]]; then
    exit 0
fi

# Only active if project has adopted C3
if [[ ! -f "$PROJECT_DIR/.c3/README.md" ]]; then
    exit 0
fi

# Check if ANY accepted ADR exists
ADR_DIR="$PROJECT_DIR/.c3/adr"

if [[ -d "$ADR_DIR" ]]; then
    for adr_file in "$ADR_DIR"/adr-*.md; do
        [[ -f "$adr_file" ]] || continue
        [[ "$adr_file" == *.plan.md ]] && continue

        # Check if ADR is accepted
        if grep -q "^status:[[:space:]]*accepted" "$adr_file" 2>/dev/null; then
            # Found an accepted ADR - allow the change
            exit 0
        fi
    done
fi

# No accepted ADR found - block with guidance
cat >&2 << EOF
[c3-gate] Change blocked: ${REL_PATH}

No accepted ADR found in this C3 project.

To modify code in a C3 project:
1. Use c3-orchestrator agent to analyze the change
2. Create ADR with approved-files list
3. Set ADR status to 'accepted'
4. Then Edit/Write will be allowed

Run: "I want to change ${REL_PATH}" to start analysis.
EOF

exit 2
```

**Step 3: Verify script is executable**

Run: `chmod +x scripts/c3-gate`

**Step 4: Commit**

```bash
git add scripts/c3-gate
git commit -m "refactor(gate): simplify to check for any accepted ADR

Instead of checking if specific files are in approved-files list,
gate now just checks if any accepted ADR exists. This allows
flexibility during implementation while maintaining commitment.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Task 2: Update ADR Template with base-commit Field

**Files:**
- Modify: `references/adr-template.md`

**Step 1: Add base-commit to frontmatter template**

Add `base-commit` field to the template frontmatter:

```yaml
---
id: adr-YYYYMMDD-{slug}
title: [Decision Title]
status: proposed
date: YYYY-MM-DD
affects: [c3-0, c3-1]
approved-files: []
base-commit:        # Captured when status becomes accepted
---
```

**Step 2: Update Status Values table**

Update the table to document base-commit capture:

```markdown
## Status Values

| Status | Meaning | Gate Behavior | base-commit |
|--------|---------|---------------|-------------|
| `proposed` | Awaiting review | Files NOT editable | Not set |
| `accepted` | Ready for implementation | All code editable | Captured (HEAD at acceptance) |
| `implemented` | Changes applied, verified | Gate relaxed | Used for verification |
| `superseded` | Replaced by another ADR | Files NOT editable | N/A |
```

**Step 3: Add Verification section to template**

Add fields for verification output:

```markdown
## Verification Results (populated on implementation)

```yaml
actual-files: []      # Files actually changed (from git diff)
verification:
  matched: []         # In approved-files AND touched
  unplanned: []       # Touched but not in approved-files
  untouched: []       # In approved-files but not touched
  verified-at:        # ISO timestamp
  verified-by: claude
```
```

**Step 4: Commit**

```bash
git add references/adr-template.md
git commit -m "docs(adr): add base-commit and verification fields to template

- base-commit: captured when ADR transitions to accepted
- actual-files: populated by verifier on implemented transition
- verification section: documents matched/unplanned/untouched files

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Task 3: Update c3-orchestrator to Capture base-commit

**Files:**
- Modify: `agents/c3-orchestrator.md`

**Step 1: Add base-commit capture instruction to Phase 5b**

Find the "Phase 5b: ADR Acceptance" section and add:

```markdown
**On Accept:**
1. Update ADR `status: proposed` → `status: accepted`
2. **Capture base-commit:** Add `base-commit: <current HEAD>` to frontmatter
   ```bash
   git rev-parse HEAD
   ```
3. c3-gate will now allow Edit/Write operations
```

**Step 2: Add example of the update**

```markdown
**Example frontmatter update on acceptance:**

```yaml
# Before acceptance
status: proposed
base-commit:

# After acceptance
status: accepted
base-commit: abc123f
```
```

**Step 3: Commit**

```bash
git add agents/c3-orchestrator.md
git commit -m "feat(orchestrator): capture base-commit on ADR acceptance

When transitioning ADR from proposed to accepted, orchestrator
now captures current HEAD as base-commit. This enables verification
to diff base-commit..HEAD when transitioning to implemented.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Task 4: Create c3-verifier Script

**Files:**
- Create: `scripts/c3-verifier`

**Step 1: Write the verifier script**

```bash
#!/usr/bin/env bash
#
# C3 ADR Verifier
#
# Verifies ADR implementation by comparing approved-files against actual changes.
# Called when transitioning ADR from accepted to implemented.
#
# Usage: c3-verifier <adr-file>
#
# Outputs verification results to stdout and updates the ADR file.

set -uo pipefail

ADR_FILE="${1:-}"

if [[ -z "$ADR_FILE" || ! -f "$ADR_FILE" ]]; then
    echo "Usage: c3-verifier <adr-file>" >&2
    exit 1
fi

# Get project directory
PROJECT_DIR="${CLAUDE_PROJECT_DIR:-$(pwd)}"

# Extract base-commit from frontmatter
BASE_COMMIT=$(grep -E "^base-commit:" "$ADR_FILE" | sed 's/base-commit:[[:space:]]*//' | tr -d '[:space:]')

if [[ -z "$BASE_COMMIT" ]]; then
    echo "Error: No base-commit found in ADR frontmatter" >&2
    exit 1
fi

# Verify base-commit exists
if ! git cat-file -e "$BASE_COMMIT" 2>/dev/null; then
    echo "Error: base-commit $BASE_COMMIT not found in git history" >&2
    exit 1
fi

# Get actual files changed since base-commit
ACTUAL_FILES=$(git diff --name-only "$BASE_COMMIT"..HEAD | grep -v "^\.c3/" | sort)

# Extract approved-files from frontmatter (handle YAML list format)
APPROVED_FILES=$(awk '
    /^approved-files:/ { in_list=1; next }
    in_list && /^[[:space:]]*-[[:space:]]/ {
        gsub(/^[[:space:]]*-[[:space:]]*/, "")
        gsub(/["\047]/, "")  # Remove quotes
        gsub(/[[:space:]]*$/, "")
        print
    }
    in_list && /^[a-z]/ { exit }
' "$ADR_FILE" | sort)

# Calculate matched, unplanned, untouched
MATCHED=$(comm -12 <(echo "$APPROVED_FILES") <(echo "$ACTUAL_FILES"))
UNPLANNED=$(comm -13 <(echo "$APPROVED_FILES") <(echo "$ACTUAL_FILES"))
UNTOUCHED=$(comm -23 <(echo "$APPROVED_FILES") <(echo "$ACTUAL_FILES"))

# Format as YAML lists
format_yaml_list() {
    local items="$1"
    if [[ -z "$items" ]]; then
        echo "[]"
    else
        echo ""
        echo "$items" | while read -r item; do
            [[ -n "$item" ]] && echo "  - $item"
        done
    fi
}

# Output verification results
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

cat << EOF

## Verification Results

\`\`\`yaml
actual-files:$(format_yaml_list "$ACTUAL_FILES")

verification:
  matched:$(format_yaml_list "$MATCHED")
  unplanned:$(format_yaml_list "$UNPLANNED")
  untouched:$(format_yaml_list "$UNTOUCHED")
  verified-at: $TIMESTAMP
  verified-by: claude
\`\`\`
EOF

# Summary
MATCHED_COUNT=$(echo "$MATCHED" | grep -c . || echo 0)
UNPLANNED_COUNT=$(echo "$UNPLANNED" | grep -c . || echo 0)
UNTOUCHED_COUNT=$(echo "$UNTOUCHED" | grep -c . || echo 0)

echo ""
echo "Summary: $MATCHED_COUNT matched, $UNPLANNED_COUNT unplanned, $UNTOUCHED_COUNT untouched"

# Exit code based on results
if [[ $UNPLANNED_COUNT -gt 0 || $UNTOUCHED_COUNT -gt 0 ]]; then
    echo ""
    echo "Note: Discrepancies found between approved and actual files."
    echo "This is informational - the verification documents what happened."
fi

exit 0
```

**Step 2: Make executable**

Run: `chmod +x scripts/c3-verifier`

**Step 3: Test with a sample ADR (manual verification)**

Create a test scenario mentally:
- ADR with base-commit pointing to a commit
- approved-files listing some files
- Some files changed, some not

**Step 4: Commit**

```bash
git add scripts/c3-verifier
git commit -m "feat(verifier): add script to verify ADR implementation

Compares approved-files against git diff from base-commit to HEAD.
Documents matched, unplanned, and untouched files in verification
output. Called when transitioning ADR to implemented status.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Task 5: Create c3-adr-transition Agent

**Files:**
- Create: `agents/c3-adr-transition.md`

**Step 1: Write the transition agent**

```markdown
---
name: c3-adr-transition
description: |
  Transitions ADR status with proper verification.
  Use when moving ADR from accepted to implemented.
  Runs verification, documents results, updates status.
model: haiku
color: cyan
tools: ["Read", "Bash", "Write", "Grep", "Glob"]
---

You are the C3 ADR Transition agent. You handle status transitions with verification.

## Your Mission

Transition an ADR from `accepted` to `implemented` status:
1. Run verification script
2. Append verification results to ADR
3. Update status in frontmatter
4. Report summary

## Workflow

### Step 1: Locate the ADR

Find the ADR to transition:

```bash
# List accepted ADRs
grep -l "^status:[[:space:]]*accepted" .c3/adr/adr-*.md 2>/dev/null | grep -v ".plan.md"
```

If multiple accepted ADRs exist, ask which one to transition.

### Step 2: Run Verifier

```bash
${CLAUDE_PLUGIN_ROOT}/scripts/c3-verifier <adr-path>
```

Capture the output.

### Step 3: Append Verification to ADR

Add the verification results section to the end of the ADR file.

### Step 4: Update Status

Change frontmatter:
- `status: accepted` → `status: implemented`

### Step 5: Report Summary

Output:
- Files matched vs approved
- Unplanned changes (informational)
- Untouched approved files (informational)
- ADR path

## Example Output

```
ADR Transition: adr-20260121-auth-refactor.md

Verification:
- 3 files matched approved list
- 1 unplanned change: e2e/auth.spec.ts (incidental fix)
- 0 approved files untouched

Status: accepted → implemented
```
```

**Step 2: Register agent in plugin.json (verify already auto-discovered)**

Agents are auto-discovered from the `agents/` directory per plugin.json `"agents": "agents"`.

**Step 3: Commit**

```bash
git add agents/c3-adr-transition.md
git commit -m "feat(agents): add c3-adr-transition for status transitions

New agent handles accepted → implemented transition:
- Runs c3-verifier to compare approved vs actual files
- Appends verification results to ADR
- Updates status in frontmatter
- Reports summary of discrepancies

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Task 6: Update c3-orchestrator with Transition Option

**Files:**
- Modify: `agents/c3-orchestrator.md`

**Step 1: Add Phase 7 for implementation completion**

Add after Phase 6:

```markdown
## Phase 7: Implementation Completion

When the user indicates implementation is complete:

```
AskUserQuestion:
  question: "Implementation complete. Ready to transition ADR to implemented?"
  options:
    - "Yes - run verification and transition"
    - "Not yet - still working"
    - "Skip verification - mark implemented directly"
```

**On "Yes":**

Dispatch c3-adr-transition:

```
Task with subagent_type: c3-skill:c3-adr-transition
Prompt:
  Transition ADR at .c3/adr/<current-adr>.md from accepted to implemented.
  Run verification and document results.
```

**On "Skip verification":**

Update ADR status directly (not recommended, loses traceability).
```

**Step 2: Commit**

```bash
git add agents/c3-orchestrator.md
git commit -m "feat(orchestrator): add Phase 7 for implementation completion

New phase handles transitioning ADR to implemented status:
- Asks user if ready to transition
- Dispatches c3-adr-transition agent for verification
- Documents verification results in ADR

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Task 7: Update Test Script

**Files:**
- Modify: `scripts/test-c3-gate`

**Step 1: Read current test script**

Understand existing test cases.

**Step 2: Update tests for simplified gate behavior**

Update test cases to reflect new behavior:
- Test: No ADR → block
- Test: Proposed ADR → block
- Test: Accepted ADR → allow (any file)
- Test: .c3/ files → always allow

**Step 3: Add test for verifier**

Add basic verifier test case.

**Step 4: Commit**

```bash
git add scripts/test-c3-gate
git commit -m "test(gate): update tests for simplified gate behavior

- Remove file-specific approval tests
- Add tests for any-accepted-ADR logic
- Add basic verifier integration test

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Task 8: Final Integration Test

**Step 1: Manual end-to-end verification**

1. Create a test project with .c3/
2. Create an ADR with status: proposed
3. Verify gate blocks
4. Change status to accepted, add base-commit
5. Verify gate allows
6. Make some changes
7. Run c3-verifier
8. Verify output shows matched/unplanned/untouched
9. Verify ADR updated correctly

**Step 2: Run full test suite**

```bash
./scripts/test-c3-gate
```

**Step 3: Commit any fixes**

If issues found, fix and commit.

---

## Summary

| Task | Component | Change |
|------|-----------|--------|
| 1 | c3-gate | Simplify to "any accepted ADR" check |
| 2 | adr-template | Add base-commit and verification fields |
| 3 | c3-orchestrator | Capture base-commit on acceptance |
| 4 | c3-verifier | New script for verification |
| 5 | c3-adr-transition | New agent for status transition |
| 6 | c3-orchestrator | Add Phase 7 for completion |
| 7 | test-c3-gate | Update tests |
| 8 | Integration | End-to-end verification |
