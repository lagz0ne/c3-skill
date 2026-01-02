# C3 Plugin Production Cleanup Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Clean up the c3-design plugin to be production-ready by removing dead code, fixing broken commands, and syncing documentation.

**Architecture:** Fix /c3 command to use skill directly (no agents), connect orphan references to skills or delete them, update CLAUDE.md to reflect actual structure, add missing harnesses.

**Tech Stack:** Markdown skills, bash scripts, Claude Code plugin structure

---

## Task 1: Fix /c3 Command

**Files:**
- Modify: `commands/c3.md`

**Step 1: Read current /c3 command**

```bash
cat commands/c3.md
```

**Step 2: Update command to invoke skill directly**

Replace content of `commands/c3.md` with:

```markdown
---
description: Design, explore, configure, or audit architecture using C3 methodology
allowed-tools: Skill
---

# C3 Architecture

Invoke the appropriate C3 skill based on intent:

| Intent | Skill |
|--------|-------|
| Adopt C3 / Initialize | `c3-skill:c3` |
| Navigate / Explore | `c3-skill:c3-query` |
| Change / Modify | `c3-skill:c3-alter` |
| Audit / Validate | `c3-skill:c3` |

For adoption or audit, use: `c3-skill:c3`
For navigation, use: `c3-skill:c3-query`
For changes, use: `c3-skill:c3-alter`
```

**Step 3: Verify command syntax**

```bash
cat commands/c3.md
```

**Step 4: Commit**

```bash
git add commands/c3.md
git commit -m "fix(commands): update /c3 to invoke skills directly instead of missing agent"
```

---

## Task 2: Add Activation Check to c3 Skill

**Files:**
- Modify: `skills/c3/SKILL.md`

**Step 1: Read current c3 skill header**

```bash
head -30 skills/c3/SKILL.md
```

**Step 2: Add REQUIRED harness after Skill Routing section**

Insert after line 18 (after Skill Routing table):

```markdown

## REQUIRED: Check Activation

**Before proceeding, check if `.c3/README.md` exists.**

- If NO `.c3/`: This is **Mode: Adopt**. Suggest `/onboard` or proceed with adoption.
- If YES `.c3/`: Route based on intent (Audit stays here, others route to dedicated skills).
```

**Step 3: Verify the edit**

```bash
head -35 skills/c3/SKILL.md
```

**Step 4: Commit**

```bash
git add skills/c3/SKILL.md
git commit -m "fix(skills): add explicit activation check harness to c3 skill"
```

---

## Task 3: Connect audit-checks.md to c3 Skill

**Files:**
- Modify: `skills/c3/SKILL.md` (Mode: Audit section)

**Step 1: Read current Mode: Audit section**

```bash
grep -A 20 "## Mode: Audit" skills/c3/SKILL.md
```

**Step 2: Update Mode: Audit to reference audit-checks.md**

Replace the Mode: Audit section with:

```markdown
## Mode: Audit

**REQUIRED:** Load `references/audit-checks.md` for comprehensive audit procedures.

Quick reference scopes:
- `audit C3` - full system
- `audit container c3-1` - single container
- `audit adr adr-YYYYMMDD-slug` - single ADR

The reference contains detailed checks for:
- Diagram-inventory consistency
- ID uniqueness
- Linkage reasoning
- Fulfillment coverage
- Orphan detection
```

**Step 3: Verify the edit**

```bash
grep -A 15 "## Mode: Audit" skills/c3/SKILL.md
```

**Step 4: Commit**

```bash
git add skills/c3/SKILL.md
git commit -m "fix(skills): connect audit-checks.md reference to c3 audit mode"
```

---

## Task 4: Delete Orphan Templates

**Files:**
- Delete: `templates/CLAUDE.md.c3-section`
- Delete: `templates/CLAUDE.md.template`

**Step 1: Verify templates are not referenced**

```bash
grep -r "CLAUDE.md.c3-section" . --include="*.md" --include="*.sh" --include="*.json" 2>/dev/null | grep -v node_modules || echo "Not referenced"
grep -r "CLAUDE.md.template" . --include="*.md" --include="*.sh" --include="*.json" 2>/dev/null | grep -v node_modules || echo "Not referenced"
```

**Step 2: Delete orphan templates**

```bash
rm templates/CLAUDE.md.c3-section
rm templates/CLAUDE.md.template
```

**Step 3: Verify deletion**

```bash
ls templates/
```

Expected: `adr-000.md  component.md  container.md  context.md`

**Step 4: Commit**

```bash
git add -A
git commit -m "chore(templates): remove unused CLAUDE.md template files"
```

---

## Task 5: Decide on Orphan References

**Files:**
- Review: `references/container-patterns.md`
- Review: `references/implementation-guide.md`
- Review: `references/structure-guide.md`
- Review: `references/v3-structure.md`

**Step 1: Check what each reference contains**

```bash
head -20 references/container-patterns.md
head -20 references/implementation-guide.md
head -20 references/structure-guide.md
head -20 references/v3-structure.md
```

**Step 2: Determine overlap and usage**

- `v3-structure.md` - Referenced by `layer-navigation.md`, keep
- `structure-guide.md` - Overlaps with v3-structure, merge or delete
- `container-patterns.md` - Could be useful for Adopt mode
- `implementation-guide.md` - Component-level guidance, could be useful

**Step 3: Add v3-structure.md reference to layer-navigation.md (already done)**

Verify:
```bash
grep "v3-structure" references/layer-navigation.md
```

**Step 4: Delete structure-guide.md (duplicates v3-structure)**

```bash
rm references/structure-guide.md
```

**Step 5: Connect container-patterns.md to c3 skill Adopt mode**

Add to c3/SKILL.md after "### Round 2: Fill (AI subagent)":

```markdown
**Reference:** See `references/container-patterns.md` for component categorization patterns.
```

**Step 6: Connect implementation-guide.md to c3-alter**

Add to c3-alter/SKILL.md in Step 7: Execute section:

```markdown
**Reference:** See `references/implementation-guide.md` for component documentation conventions.
```

**Step 7: Commit**

```bash
git add -A
git commit -m "refactor(references): connect orphan references to skills, remove duplicate structure-guide.md"
```

---

## Task 6: Update c3-session-start Hook Message

**Files:**
- Modify: `scripts/c3-session-start`

**Step 1: Read current hook**

```bash
cat scripts/c3-session-start
```

**Step 2: Update message to mention all commands**

Replace the echo line with:

```bash
echo "C3 architecture documentation exists. Commands: /c3, /query, /alter. Skills: c3-skill:c3, c3-skill:c3-query, c3-skill:c3-alter"
```

**Step 3: Verify the edit**

```bash
cat scripts/c3-session-start
```

**Step 4: Test the hook**

```bash
CLAUDE_PROJECT_DIR=/tmp/test-c3 mkdir -p /tmp/test-c3/.c3 && echo "test" > /tmp/test-c3/.c3/README.md && CLAUDE_PROJECT_DIR=/tmp/test-c3 ./scripts/c3-session-start && rm -rf /tmp/test-c3
```

**Step 5: Commit**

```bash
git add scripts/c3-session-start
git commit -m "fix(hooks): update session start message to mention all C3 commands"
```

---

## Task 7: Update CLAUDE.md - Skills Organization Section

**Files:**
- Modify: `CLAUDE.md`

**Step 1: Find Skills Organization section**

```bash
grep -n "## Skills Organization" CLAUDE.md
```

**Step 2: Replace outdated skills info with current structure**

Find and replace the entire "## Skills Organization" section with:

```markdown
## Skills Organization

### Skills (3 total)

| Skill | Purpose | When Used |
|-------|---------|-----------|
| `c3` | Adoption and Audit | No .c3/ exists (adopt), or validating existing docs (audit) |
| `c3-query` | Navigation + Code Exploration | Understanding architecture, finding where things are |
| `c3-alter` | Changes through ADR | Any modification to code or architecture |

### Commands (4 total)

| Command | Purpose | Delegates To |
|---------|---------|--------------|
| `/c3` | Main entry point | Routes to appropriate skill |
| `/onboard` | Initialize C3 for new project | Scripts + c3 skill |
| `/query` | Navigate architecture | c3-query skill |
| `/alter` | Make changes via ADR | c3-alter skill |

### Layer Mapping

| Layer | Skill |
|-------|-------|
| Context (c3-0) | `c3` (adopt), `c3-query` (navigate) |
| Container (c3-N) | `c3-query` (navigate), `c3-alter` (change) |
| Component (c3-NNN) | `c3-query` (navigate), `c3-alter` (change) |
```

**Step 3: Verify the edit**

```bash
grep -A 30 "## Skills Organization" CLAUDE.md
```

**Step 4: Commit**

```bash
git add CLAUDE.md
git commit -m "docs(CLAUDE.md): update Skills Organization to reflect current structure"
```

---

## Task 8: Update CLAUDE.md - Remove Dead References

**Files:**
- Modify: `CLAUDE.md`

**Step 1: Find Key Files to Know section**

```bash
grep -n "## Key Files" CLAUDE.md
```

**Step 2: Update Reference Files table**

Replace the Reference Files table with:

```markdown
### Reference Files

| File | Purpose |
|------|---------|
| `references/adr-template.md` | ADR structure template |
| `references/plan-template.md` | Execution plan template |
| `references/layer-navigation.md` | Shared layer traversal pattern |
| `references/v3-structure.md` | File paths, ID patterns, frontmatter |
| `references/audit-checks.md` | Comprehensive audit procedures |
| `references/container-patterns.md` | Component categorization patterns |
| `references/implementation-guide.md` | Component documentation conventions |
```

**Step 3: Remove references to non-existent files**

Search and remove any references to:
- `references/naming-conventions.md`
- `references/lookup-patterns.md`
- `references/core-principle.md`
- `references/diagram-patterns.md`
- `agents/c3.md`
- `skills/c3-structure/`
- `skills/c3-implementation/`

**Step 4: Commit**

```bash
git add CLAUDE.md
git commit -m "docs(CLAUDE.md): remove references to non-existent files, update reference table"
```

---

## Task 9: Add component.md Template to c3-init.sh

**Files:**
- Modify: `scripts/c3-init.sh`
- Keep: `templates/component.md`

**Step 1: Check if component.md is a useful template**

```bash
cat templates/component.md
```

**Step 2: Decide - Document as manual creation or integrate**

The component template is for creating component docs when needed (not during init). Document this in the c3 skill.

Add to c3/SKILL.md in Mode: Design section (or leave as-is since c3-alter handles this):

No change needed - component docs are created via c3-alter workflow, not during init.

**Step 3: Verify component.md is referenced somewhere**

Add to c3-alter/SKILL.md Step 7: Execute section:

```markdown
For new components, use `templates/component.md` as starting template.
```

**Step 4: Commit**

```bash
git add skills/c3-alter/SKILL.md
git commit -m "docs(skills): document component.md template usage in c3-alter"
```

---

## Task 10: Final Verification

**Step 1: Check for any remaining dead references**

```bash
grep -r "c3-structure\|c3-implementation\|agents/c3" . --include="*.md" 2>/dev/null | grep -v node_modules | grep -v ".git"
```

Expected: No output (or only in migration history files)

**Step 2: Verify all skills reference their dependencies**

```bash
echo "=== c3 skill references ===" && grep -E "references/|REQUIRED" skills/c3/SKILL.md
echo "=== c3-query skill references ===" && grep -E "references/|REQUIRED" skills/c3-query/SKILL.md
echo "=== c3-alter skill references ===" && grep -E "references/|REQUIRED" skills/c3-alter/SKILL.md
```

**Step 3: Verify all commands work**

```bash
cat commands/c3.md
cat commands/query.md
cat commands/alter.md
cat commands/onboard.md
```

**Step 4: Final commit**

```bash
git add -A
git commit -m "chore: production cleanup complete - all dead code removed, references connected"
```

---

## Summary of Changes

| Category | Action | Files |
|----------|--------|-------|
| Commands | Fix /c3 to use skills | `commands/c3.md` |
| Skills | Add activation harness to c3 | `skills/c3/SKILL.md` |
| Skills | Connect audit-checks.md | `skills/c3/SKILL.md` |
| Templates | Delete unused | `templates/CLAUDE.md.*` |
| References | Delete duplicate | `references/structure-guide.md` |
| References | Connect orphans | `skills/c3/SKILL.md`, `skills/c3-alter/SKILL.md` |
| Hooks | Update message | `scripts/c3-session-start` |
| Docs | Sync CLAUDE.md | `CLAUDE.md` |

**Post-cleanup health score: 90+/100**
