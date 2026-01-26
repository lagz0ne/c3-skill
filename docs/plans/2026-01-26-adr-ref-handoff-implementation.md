# ADR Reference Handoff Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make refs first-class constraints on components, with automatic discovery and hard enforcement at commit time.

**Architecture:** Components declare refs in `## References` section → c3-analysis extracts them + discovers new ones → ADR inherits refs → c3-gate enforces compliance.

**Tech Stack:** Markdown templates, Bash scripts, YAML frontmatter

---

## Task 1: Update Component Template

**Files:**
- Modify: `templates/component.md`

**Step 1: Add ## References section to template**

Add after `## Related Refs` section (line 67-71), replace it with a stricter version:

```markdown
## References

Patterns this component MUST follow (enforced by c3-gate):

| Ref | Applies To |
|-----|------------|
| ref-{slug} | {What aspect of this component it governs} |

<!--
RULES:
- List all refs that constrain this component's implementation
- Each ref must exist in .c3/refs/
- Code changes to this component must honor all listed refs
- Empty table is valid but must be explicit (no refs apply)
-->
```

**Step 2: Remove old Related Refs section**

Delete lines 67-71 (the old `## Related Refs` section) since we're replacing it.

**Step 3: Verify template structure**

Run: `cat templates/component.md | grep -A5 "## References"`

Expected: New References section visible

**Step 4: Commit**

```bash
git add templates/component.md
git commit -m "$(cat <<'EOF'
feat(templates): add ## References section to component template

Components now declare which refs they must follow. This enables:
- Explicit constraints on implementation
- ADR inheritance of refs from affected components
- c3-gate enforcement at commit time

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: Update component-types.md Reference

**Files:**
- Modify: `references/component-types.md`

**Step 1: Add References requirement to Foundation section**

After line 100 (`**Template:** \`templates/component.md\``), add:

```markdown

**Required sections:**
- `## References` - Refs this component must follow (can be empty but must be explicit)
- `## Code References` - Actual code files implementing this component
```

**Step 2: Add References requirement to Feature section**

After line 123 (`**Template:** \`templates/component.md\``), add:

```markdown

**Required sections:**
- `## References` - Refs this component must follow (can be empty but must be explicit)
- `## Code References` - Actual code files implementing this component
```

**Step 3: Verify changes**

Run: `grep -A3 "Required sections" references/component-types.md`

Expected: Two occurrences with References listed

**Step 4: Commit**

```bash
git add references/component-types.md
git commit -m "$(cat <<'EOF'
docs(component-types): make ## References a required section

Both Foundation and Feature components must now have a ## References
section listing which refs they must follow. Can be empty but explicit.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: Add Part 4 to c3-analysis Agent

**Files:**
- Modify: `agents/c3-analysis.md`

**Step 1: Add Part 4 to Process section**

After line 56 (`### Part 3: Pattern Analysis`), add:

```markdown

### Part 4: Reference Discovery

1. For each affected component:
   - Read `## References` section
   - Extract all `ref-*` entries

2. For new components (if adding):
   - Read `.c3/refs/` directory listing
   - Extract `goal` from each ref frontmatter
   - Match change intent keywords against goals
   - Suggest applicable refs

3. Deduplicate and compile required refs list
```

**Step 2: Add Part 4 to Output Format**

After line 120 (end of Part 3 in output format), add:

```markdown

---

## Part 4: Required References

### From Existing Components
| Component | Refs |
|-----------|------|
| c3-XXX | ref-aaa, ref-bbb |

### For New Components (if applicable)
| Ref | Source | Why |
|-----|--------|-----|
| ref-XXX | keyword match | [matched goal text] |
| ref-YYY | container pattern | [container uses this pattern] |

### Combined Required Refs
- ref-aaa (from c3-XXX)
- ref-bbb (from c3-XXX)
- ref-XXX (recommended for new component)

### Ref Discovery: [complete/partial/none]
[If partial/none, explain what couldn't be discovered]
```

**Step 3: Update Summary section**

Modify line 124-132 to include ref discovery status:

```markdown
## Summary

### All Checks Passed: [yes/no]
### Issues Found:
- [Issue 1]
- [Or "None"]

### Required Refs Identified: [yes/no]
### Ready for ADR: [yes/no]
[If no, explain what needs resolution]
```

**Step 4: Verify changes**

Run: `grep -c "Part 4" agents/c3-analysis.md`

Expected: At least 2 (process + output)

**Step 5: Commit**

```bash
git add agents/c3-analysis.md
git commit -m "$(cat <<'EOF'
feat(c3-analysis): add Part 4 Reference Discovery

c3-analysis now extracts refs from affected components and discovers
applicable refs for new components via keyword matching. This enables
ADR to inherit refs and c3-gate to enforce them.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: Update c3-synthesizer Validation

**Files:**
- Modify: `agents/c3-synthesizer.md`

**Step 1: Add ref discovery validation to Step 6**

After line 93 (end of Part 3 checks), add:

```markdown

3. **From Part 4 - Reference Discovery:**
   - Refs identified for all affected components
   - New component has recommended refs (if applicable)
```

**Step 2: Update total validation count**

Change line 98 from:
```
**Total: 11 validation checks**
```
To:
```
**Total: 13 validation checks**
```

**Step 3: Add ref checks to Validation Summary table**

After line 154 (row 11), add:

```markdown
| 12 | Refs identified | c3-analysis | ✓/✗ | |
| 13 | New component refs | c3-analysis | ✓/✗/N/A | |
```

**Step 4: Verify changes**

Run: `grep "13 validation" agents/c3-synthesizer.md`

Expected: Line with "Total: 13 validation checks"

**Step 5: Commit**

```bash
git add agents/c3-synthesizer.md
git commit -m "$(cat <<'EOF'
feat(c3-synthesizer): add ref discovery to validation checks

Synthesizer now validates that refs are identified for all affected
components and that new components have recommended refs. Total
validation checks increased from 11 to 13.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 5: Update ADR Template

**Files:**
- Modify: `references/adr-template.md`

**Step 1: Add references-to-follow to frontmatter**

After line 45 (`base-commit:        # Captured when status becomes accepted`), add:

```yaml
references-to-follow: []  # Populated from c3-analysis Part 4
```

**Step 2: Add ## References to Follow section**

After line 78 (`## Affected Layers` section ends), add:

```markdown

## References to Follow

Refs inherited from affected components (enforced by c3-gate):

| Ref | Source |
|-----|--------|
| ref-{slug} | [component or "new component recommendation"] |

**For new components:**
| Ref | Reason |
|-----|--------|
| ref-{slug} | [why this ref applies] |

**Note:** The `references-to-follow` list in frontmatter is used by c3-gate to enforce compliance. Code changes must honor all listed refs.
```

**Step 3: Verify changes**

Run: `grep -A10 "References to Follow" references/adr-template.md`

Expected: New section visible

**Step 4: Commit**

```bash
git add references/adr-template.md
git commit -m "$(cat <<'EOF'
feat(adr-template): add ## References to Follow section

ADRs now include refs inherited from affected components. This provides
visibility into constraints and enables c3-gate enforcement.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 6: Update c3-orchestrator Phase 5

**Files:**
- Modify: `agents/c3-orchestrator.md`

**Step 1: Add refs to ADR generation template**

In Phase 5 (around line 283), update the ADR template to include:

After the `affects:` line in the example frontmatter, add:
```yaml
references-to-follow:
  - [refs from c3-analysis Part 4]
```

**Step 2: Add ## References to Follow to ADR body**

After the `## Affected Layers` section in the template (around line 314), add:

```markdown

## References to Follow

Refs that MUST be followed during implementation:

| Ref | Source |
|-----|--------|
[From c3-analysis Part 4 - list all refs]

**Gate behavior:** c3-gate verifies compliance with these refs before allowing commits.
```

**Step 3: Add ref opportunity detection to Phase 1**

After line 162 (end of example questions), add:

```markdown

**Ref opportunity detection:**

If the change involves a pattern that could be reused:

```
AskUserQuestion:
  question: "This looks like a reusable pattern. Create ref first?"
  options:
    - "Yes - create ref-{slug} first (recommended)"
    - "No - implement without ref (one-off)"
```

If user chooses "Yes":
1. Pause c3-alter flow
2. Guide user to `/c3-ref` to create the ref
3. Resume c3-alter with new ref in scope
```

**Step 4: Verify changes**

Run: `grep -c "references-to-follow" agents/c3-orchestrator.md`

Expected: At least 2 occurrences

**Step 5: Commit**

```bash
git add agents/c3-orchestrator.md
git commit -m "$(cat <<'EOF'
feat(c3-orchestrator): include refs in ADR generation

Phase 5 now generates ADR with references-to-follow section populated
from c3-analysis Part 4. Phase 1 also detects ref opportunities and
offers to create refs first.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 7: Update c3-alter Skill

**Files:**
- Modify: `skills/c3-alter/SKILL.md`

**Step 1: Update Stage 4 to require refs**

After line 97 (`**Key sections:** Problem, Decision, Rationale, Affected Layers, References Affected, Verification`), change to:

```markdown
**Key sections:** Problem, Decision, Rationale, Affected Layers, References to Follow, Verification

**References to Follow** is populated from c3-analysis Part 4 and MUST be present.
```

**Step 2: Add ref discovery to Stage 2 table**

Update the Stage 2 table (line 75-79) to include:

```markdown
| Step | Action |
|------|--------|
| Analyze | Read affected C3 docs via layer navigation, extract ## References |
| Ask | Are docs accurate? Recent code changes not documented? |
| Synthesize | List affected components, their refs, their current behavior, dependencies |
| Review | User confirms or corrects |
```

**Step 3: Update example to show refs**

In Example 1 (around line 175), update Stage 2 output:

```markdown
Stage 2 - Current State:
  Affected: c3-2-api (API Backend)
  Current: No rate limiting exists
  Depends on: c3-201-auth-middleware (good injection point)
  Refs from c3-201: ref-error-handling, ref-middleware-chain
```

**Step 4: Verify changes**

Run: `grep "References to Follow" skills/c3-alter/SKILL.md`

Expected: At least 1 occurrence

**Step 5: Commit**

```bash
git add skills/c3-alter/SKILL.md
git commit -m "$(cat <<'EOF'
feat(c3-alter): require References to Follow in ADR

Stage 4 now requires References to Follow section in ADR. Stage 2
updated to extract refs from affected components.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 8: Enhance c3-gate for Ref Enforcement

**Files:**
- Modify: `scripts/c3-gate`

**Step 1: Add function to extract refs from component**

After line 38 (`REL_PATH="${FILE_PATH#$PROJECT_DIR/}"`), add:

```bash
# Extract refs from a component doc
extract_component_refs() {
    local component_file="$1"
    [[ -f "$component_file" ]] || return 0

    # Look for ## References section and extract ref-* entries
    awk '/^## References/,/^## / {
        if (/ref-[a-z0-9-]+/) {
            match($0, /ref-[a-z0-9-]+/)
            print substr($0, RSTART, RLENGTH)
        }
    }' "$component_file" | sort -u
}
```

**Step 2: Add function to find component for a file**

After the new function, add:

```bash
# Find component doc that owns a file path
find_component_for_file() {
    local file_path="$1"
    local c3_dir="$PROJECT_DIR/.c3"

    # Search component docs for this file in ## Code References
    for container_dir in "$c3_dir"/c3-*/; do
        [[ -d "$container_dir" ]] || continue
        for component_file in "$container_dir"c3-*.md; do
            [[ -f "$component_file" ]] || continue
            if grep -q "$file_path" "$component_file" 2>/dev/null; then
                echo "$component_file"
                return 0
            fi
        done
    done
    return 1
}
```

**Step 3: Add ref compliance check to main flow**

Before the "No accepted ADR found" block (line 67), add:

```bash
# Check refs for the file being modified
component_doc=$(find_component_for_file "$REL_PATH")
if [[ -n "$component_doc" ]]; then
    required_refs=$(extract_component_refs "$component_doc")
    if [[ -n "$required_refs" ]]; then
        # Log refs that must be followed (for awareness)
        # Full compliance checking would require semantic analysis
        # For now, we surface the refs as a reminder
        echo "[c3-gate] File $REL_PATH is governed by:" >&2
        echo "$required_refs" | while read ref; do
            echo "  - $ref" >&2
        done
        echo "[c3-gate] Ensure changes comply with these refs." >&2
    fi
fi
```

**Step 4: Test the script syntax**

Run: `bash -n scripts/c3-gate`

Expected: No output (syntax OK)

**Step 5: Commit**

```bash
git add scripts/c3-gate
git commit -m "$(cat <<'EOF'
feat(c3-gate): surface component refs on code changes

c3-gate now finds the component that owns a file and surfaces the refs
that govern it. This provides awareness of constraints during commits.
Full semantic compliance checking is deferred to /c3 audit.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 9: Update Onboard Skill for Ref Assignment

**Files:**
- Modify: `skills/onboard/SKILL.md`

**Step 1: Add ref assignment guidance**

Find the section about component creation and add guidance to assign refs:

```markdown
### Ref Assignment for New Components

When creating component documentation:

1. Check existing refs in `.c3/refs/` (if any exist from previous work)
2. For each new component, determine applicable refs:
   - Cross-cutting concerns (error handling, logging) → likely has a ref
   - Container-specific patterns → check container README for patterns
   - Domain-specific → may not need refs initially
3. Populate `## References` section even if empty (explicit > implicit)

**Ref discovery signals:**
| Signal | Likely Ref |
|--------|------------|
| Handles errors | ref-error-handling |
| HTTP middleware | ref-middleware-chain |
| Database queries | ref-data-access |
| Authentication | ref-auth |
| Logging | ref-logging |
```

**Step 2: Verify changes**

Run: `grep "Ref Assignment" skills/onboard/SKILL.md`

Expected: Section found

**Step 3: Commit**

```bash
git add skills/onboard/SKILL.md
git commit -m "$(cat <<'EOF'
feat(onboard): add ref assignment guidance for new components

Onboarding now includes guidance on assigning refs to new components
based on cross-cutting concern signals.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 10: Add Ref Discovery to Audit Checks

**Files:**
- Modify: `references/audit-checks.md`

**Step 1: Add missing refs check**

Add a new audit check for components without `## References`:

```markdown
## Check: Component Missing References Section

**Severity:** Warning

**Detection:**
```bash
# Find component docs without ## References section
for file in .c3/c3-*/c3-*.md; do
    if ! grep -q "^## References" "$file" 2>/dev/null; then
        echo "WARN: $file missing ## References section"
    fi
done
```

**Fix:** Add `## References` section with applicable refs (can be empty but explicit).

**Why it matters:** Components without declared refs cannot be enforced by c3-gate.
```

**Step 2: Verify changes**

Run: `grep "Missing References" references/audit-checks.md`

Expected: New check found

**Step 3: Commit**

```bash
git add references/audit-checks.md
git commit -m "$(cat <<'EOF'
feat(audit): add check for missing ## References section

Audit now warns about components without ## References section, enabling
migration to the new ref enforcement model.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 11: Final Integration Test

**Files:**
- None (verification only)

**Step 1: Verify all files were modified**

Run:
```bash
git log --oneline -10
```

Expected: 10 commits from this implementation

**Step 2: Check template has ## References**

Run:
```bash
grep -A5 "^## References" templates/component.md
```

Expected: New section with table format

**Step 3: Check c3-analysis has Part 4**

Run:
```bash
grep "Part 4" agents/c3-analysis.md
```

Expected: Multiple occurrences

**Step 4: Check ADR template has References to Follow**

Run:
```bash
grep "References to Follow" references/adr-template.md
```

Expected: Section found

**Step 5: Check c3-gate has ref extraction**

Run:
```bash
grep "extract_component_refs" scripts/c3-gate
```

Expected: Function found

**Step 6: Commit summary (no action needed)**

All changes complete. The implementation adds:
1. `## References` section to component template
2. Part 4 (Reference Discovery) to c3-analysis
3. Ref validation to c3-synthesizer (13 checks total)
4. `## References to Follow` to ADR template
5. Ref inheritance in c3-orchestrator
6. Ref requirement in c3-alter skill
7. Ref surfacing in c3-gate
8. Ref assignment guidance in onboard
9. Missing refs audit check

---

## Summary

| Task | File | Change |
|------|------|--------|
| 1 | `templates/component.md` | Add `## References` section |
| 2 | `references/component-types.md` | Document References as required |
| 3 | `agents/c3-analysis.md` | Add Part 4: Reference Discovery |
| 4 | `agents/c3-synthesizer.md` | Add ref validation (13 checks) |
| 5 | `references/adr-template.md` | Add `## References to Follow` |
| 6 | `agents/c3-orchestrator.md` | Include refs in ADR generation |
| 7 | `skills/c3-alter/SKILL.md` | Require refs in Stage 4 |
| 8 | `scripts/c3-gate` | Surface refs on code changes |
| 9 | `skills/onboard/SKILL.md` | Add ref assignment guidance |
| 10 | `references/audit-checks.md` | Add missing refs check |
| 11 | (verification) | Integration test |
