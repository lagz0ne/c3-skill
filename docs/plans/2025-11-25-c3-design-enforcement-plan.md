# c3-design Skill Enforcement Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix c3-design skill so ADR creation is mandatory and handoff executes settings.yaml preferences.

**Architecture:** Add Phase 4 (Handoff) to c3-design, add mandatory gates between phases, add TodoWrite enforcement for phase tracking.

**Tech Stack:** Markdown skill files, YAML settings

---

## Problem Summary

Two issues identified:
1. **ADR not created** - Phase 3 can be skipped; no enforcement
2. **Handoff too quick** - settings.yaml `handoff:` section is defined but never executed

## Tasks

### Task 1: Add Phase Gates and TodoWrite Enforcement

**Files:**
- Modify: `skills/c3-design/SKILL.md:16-23` (Quick Reference table)

**Step 1: Update Quick Reference to show 4 phases with gates**

Find this text in the file:
```markdown
## Quick Reference

| Phase | Key Activities | Output |
|-------|---------------|--------|
| **1. Surface Understanding** | Read TOC, parse request, form hypothesis | Initial hypothesis |
| **2. Iterative Scoping** | HYPOTHESIZE → EXPLORE → DISCOVER loop | Stable scope |
| **3. ADR with Stream** | Document journey, changes, verification | ADR in `.c3/adr/` |
```

Replace with:
```markdown
## Quick Reference

| Phase | Key Activities | Output | Gate |
|-------|---------------|--------|------|
| **1. Surface Understanding** | Read TOC, parse request, form hypothesis | Initial hypothesis | TodoWrite: phases tracked |
| **2. Iterative Scoping** | HYPOTHESIZE → EXPLORE → DISCOVER loop | Stable scope | Scope stable, all IDs named |
| **3. ADR Creation** | Document journey, changes, verification | ADR in `.c3/adr/` | **ADR file exists** |
| **4. Handoff** | Execute settings.yaml handoff steps | Tasks/notifications created | Handoff complete |

**MANDATORY:** You MUST create TodoWrite items for each phase. No phase can be skipped.
```

**Step 2: Verify the change**

Run: `grep -A 8 "## Quick Reference" skills/c3-design/SKILL.md`
Expected: Table shows 4 phases with Gate column

**Step 3: Commit**

```bash
git add skills/c3-design/SKILL.md
git commit -m "docs(c3-design): add phase gates and TodoWrite enforcement to quick reference"
```

---

### Task 2: Add Mandatory TodoWrite Section After Overview

**Files:**
- Modify: `skills/c3-design/SKILL.md:14-15` (after "Announce at start" line)

**Step 1: Add TodoWrite requirement section**

Find this text:
```markdown
**Announce at start:** "I'm using the c3-design skill to guide you through architecture design."

## Quick Reference
```

Replace with:
```markdown
**Announce at start:** "I'm using the c3-design skill to guide you through architecture design."

## Mandatory Phase Tracking

**IMMEDIATELY after announcing, create TodoWrite items:**

```
Phase 1: Surface Understanding - Read TOC, form hypothesis
Phase 2: Iterative Scoping - HYPOTHESIZE → EXPLORE → DISCOVER until stable
Phase 3: ADR Creation - Create ADR file in .c3/adr/ (MANDATORY)
Phase 4: Handoff - Execute settings.yaml handoff steps
```

**Rules:**
- Mark each phase `in_progress` when starting
- Mark `completed` only when gate criteria met
- **Phase 3 gate:** ADR file MUST exist before marking complete
- **Phase 4 gate:** Handoff steps MUST be executed before marking complete

**Skipping phases = skill failure. No exceptions.**

## Quick Reference
```

**Step 2: Verify the change**

Run: `grep -A 15 "## Mandatory Phase Tracking" skills/c3-design/SKILL.md`
Expected: Shows TodoWrite requirements and rules

**Step 3: Commit**

```bash
git add skills/c3-design/SKILL.md
git commit -m "docs(c3-design): add mandatory phase tracking with TodoWrite"
```

---

### Task 3: Add ADR Gate to Phase 2 Exit Criteria

**Files:**
- Modify: `skills/c3-design/SKILL.md:143-150` (Loop Exit Criteria section)

**Step 1: Update exit criteria to mandate ADR creation next**

Find this text:
```markdown
#### Loop Exit Criteria

Scope is stable when:
- You can name all affected documents (by ID)
- You understand why each is affected
- No exploration reveals new upstream/higher impacts
- Socratic confirmation validates understanding
```

Replace with:
```markdown
#### Loop Exit Criteria

Scope is stable when:
- You can name all affected documents (by ID)
- You understand why each is affected
- No exploration reveals new upstream/higher impacts
- Socratic confirmation validates understanding

**MANDATORY NEXT STEP:** Phase 3 (ADR Creation) is required.

You CANNOT:
- Update any C3 documents directly
- Skip to handoff
- Consider the design "done"

Until you have created an ADR file in `.c3/adr/`.
```

**Step 2: Verify the change**

Run: `grep -A 15 "#### Loop Exit Criteria" skills/c3-design/SKILL.md`
Expected: Shows mandatory next step warning

**Step 3: Commit**

```bash
git add skills/c3-design/SKILL.md
git commit -m "docs(c3-design): add mandatory ADR gate after Phase 2"
```

---

### Task 4: Rename Phase 3 and Add Verification Gate

**Files:**
- Modify: `skills/c3-design/SKILL.md:151-154` (Phase 3 header area)

**Step 1: Update Phase 3 header with gate**

Find this text:
```markdown
### Phase 3: ADR with Stream

**Goal:** Document the decision capturing the full scoping journey.
```

Replace with:
```markdown
### Phase 3: ADR Creation (MANDATORY)

**Goal:** Document the decision capturing the full scoping journey.

**This phase is NOT optional.** You must create an ADR before any document updates.
```

**Step 2: Verify the change**

Run: `grep -A 4 "### Phase 3:" skills/c3-design/SKILL.md`
Expected: Shows "(MANDATORY)" and warning text

**Step 3: Commit**

```bash
git add skills/c3-design/SKILL.md
git commit -m "docs(c3-design): mark Phase 3 as mandatory with warning"
```

---

### Task 5: Add ADR Verification After Template

**Files:**
- Modify: `skills/c3-design/SKILL.md:237-239` (After ADR section)

**Step 1: Add verification gate after ADR template**

Find this text:
```markdown
**After ADR:**
1. Update affected documents (CTX/CON/COM) as specified
2. Regenerate TOC: `.c3/scripts/build-toc.sh`
```

Replace with:
```markdown
**Phase 3 Gate - Verify ADR exists:**

```bash
# Verify ADR was created
ls .c3/adr/adr-*.md | tail -1
```

If no ADR file exists, **STOP**. Create the ADR before proceeding.

**After ADR verified:**
1. Update affected documents (CTX/CON/COM) as specified
2. Regenerate TOC: `.c3/scripts/build-toc.sh`
3. **Proceed to Phase 4: Handoff**
```

**Step 2: Verify the change**

Run: `grep -A 12 "Phase 3 Gate" skills/c3-design/SKILL.md`
Expected: Shows verification command and gate

**Step 3: Commit**

```bash
git add skills/c3-design/SKILL.md
git commit -m "docs(c3-design): add ADR verification gate before document updates"
```

---

### Task 6: Add Phase 4 Handoff Section

**Files:**
- Modify: `skills/c3-design/SKILL.md` (insert after Phase 3, before Sub-Skill Invocation)

**Step 1: Add complete Phase 4 section**

Find this text:
```markdown
3. **Proceed to Phase 4: Handoff**

## Sub-Skill Invocation
```

Replace with:
```markdown
3. **Proceed to Phase 4: Handoff**

### Phase 4: Handoff (MANDATORY)

**Goal:** Execute post-ADR steps configured in settings.yaml.

**Step 1: Load handoff configuration**

```bash
# Check for settings.yaml
cat .c3/settings.yaml 2>/dev/null | grep -A 20 "^handoff:" || echo "NO_HANDOFF_CONFIG"
```

**Step 2: Determine handoff actions**

| If settings.yaml has... | Then do... |
|------------------------|------------|
| `handoff:` section exists | Execute each step listed |
| No `handoff:` section | Use defaults below |
| No settings.yaml | Use defaults below |

**Default handoff steps (when no config):**
1. Summarize ADR to user
2. List affected documents
3. Ask: "Would you like me to create implementation tasks?"

**Step 3: Execute handoff**

For each step in handoff config (or defaults):
- If "create tasks" → Use vibe_kanban MCP or ask user for task system
- If "notify team" → Ask user how to notify (Slack, email, etc.)
- If "update docs" → Already done in Phase 3
- If custom step → Execute as described

**Step 4: Verify handoff complete**

Confirm with user:
- "ADR created: `.c3/adr/adr-YYYYMMDD-slug.md`"
- "Documents updated: [list]"
- "Handoff actions completed: [list what was done]"

**Phase 4 Gate:** User acknowledges handoff is complete.

## Sub-Skill Invocation
```

**Step 2: Verify the change**

Run: `grep -A 40 "### Phase 4: Handoff" skills/c3-design/SKILL.md`
Expected: Shows complete Phase 4 section

**Step 3: Commit**

```bash
git add skills/c3-design/SKILL.md
git commit -m "docs(c3-design): add Phase 4 Handoff with settings.yaml integration"
```

---

### Task 7: Add Red Flags for Phase Skipping

**Files:**
- Modify: `skills/c3-design/SKILL.md:281-292` (Red Flags section)

**Step 1: Add rationalization counters for skipping phases**

Find this text:
```markdown
## Red Flags & Counters

| Rationalization | Counter |
|-----------------|---------|
| "No time to refresh the TOC, I'll just skim files" | Stop and build/read the TOC first; C3 navigation depends on it. |
| "Examples can keep duplicate IDs, they're just sample data" | IDs must be unique or locate/anchor references break—fix collisions before scoping. |
| "I'll ask the user where to change docs instead of hypothesizing" | Hypothesis bounds exploration and prevents confirmation bias; form it before asking questions. |

**Red flags that mean you should pause:**
- `.c3/TOC.md` missing or obviously stale.
- Component IDs reused across containers or layers.
- ADR being drafted without notes from hypothesis, exploration, and discovery.
```

Replace with:
```markdown
## Red Flags & Counters

| Rationalization | Counter |
|-----------------|---------|
| "No time to refresh the TOC, I'll just skim files" | Stop and build/read the TOC first; C3 navigation depends on it. |
| "Examples can keep duplicate IDs, they're just sample data" | IDs must be unique or locate/anchor references break—fix collisions before scoping. |
| "I'll ask the user where to change docs instead of hypothesizing" | Hypothesis bounds exploration and prevents confirmation bias; form it before asking questions. |
| "The scope is clear, I can skip the ADR" | **NO.** ADR is mandatory. It documents the journey and enables review. |
| "I'll just update the docs and mention what I did" | **NO.** ADR first, then doc updates. This is non-negotiable. |
| "Handoff is just cleanup, I can skip it" | **NO.** Handoff ensures tasks are created and team is informed. Execute it. |
| "No settings.yaml means no handoff needed" | **NO.** Use default handoff steps. Always confirm completion with user. |

**Red flags that mean you should pause:**
- `.c3/TOC.md` missing or obviously stale.
- Component IDs reused across containers or layers.
- ADR being drafted without notes from hypothesis, exploration, and discovery.
- **Updating C3 documents without an ADR file existing.**
- **Ending the session without executing handoff.**
- **No TodoWrite items for the 4 phases.**
```

**Step 2: Verify the change**

Run: `grep -A 20 "## Red Flags & Counters" skills/c3-design/SKILL.md`
Expected: Shows new rationalizations and red flags

**Step 3: Commit**

```bash
git add skills/c3-design/SKILL.md
git commit -m "docs(c3-design): add red flags for phase skipping and missing handoff"
```

---

### Task 8: Update Common Mistakes Section

**Files:**
- Modify: `skills/c3-design/SKILL.md:274-279` (Common Mistakes section)

**Step 1: Add mistakes about skipping phases**

Find this text:
```markdown
## Common Mistakes

- Skipping the TOC/ID sweep and diving into files, which hides upstream impacts and duplicate IDs.
- Asking the user to point to files instead of forming a hypothesis from the TOC and existing docs.
- Drafting an ADR before the hypothesis → explore → discover loop stabilizes.
- Treating examples as throwaway and allowing duplicate IDs or missing TOC to persist.
```

Replace with:
```markdown
## Common Mistakes

- Skipping the TOC/ID sweep and diving into files, which hides upstream impacts and duplicate IDs.
- Asking the user to point to files instead of forming a hypothesis from the TOC and existing docs.
- Drafting an ADR before the hypothesis → explore → discover loop stabilizes.
- Treating examples as throwaway and allowing duplicate IDs or missing TOC to persist.
- **Skipping Phase 3 (ADR creation)** and updating documents directly.
- **Skipping Phase 4 (Handoff)** and ending the session without executing settings.yaml steps.
- **Not creating TodoWrite items** for phase tracking.
- **Ignoring settings.yaml** handoff configuration.
```

**Step 2: Verify the change**

Run: `grep -A 12 "## Common Mistakes" skills/c3-design/SKILL.md`
Expected: Shows new mistakes about phase skipping

**Step 3: Commit**

```bash
git add skills/c3-design/SKILL.md
git commit -m "docs(c3-design): add common mistakes for phase skipping"
```

---

### Task 9: Final Verification

**Files:**
- Read: `skills/c3-design/SKILL.md` (full file)

**Step 1: Verify all changes are present**

Run these commands:
```bash
# Check 4 phases in quick reference
grep -c "Phase 4" skills/c3-design/SKILL.md

# Check mandatory markers
grep -c "MANDATORY" skills/c3-design/SKILL.md

# Check TodoWrite requirement
grep -c "TodoWrite" skills/c3-design/SKILL.md

# Check handoff section exists
grep -c "### Phase 4: Handoff" skills/c3-design/SKILL.md
```

Expected:
- Phase 4 count: at least 3
- MANDATORY count: at least 4
- TodoWrite count: at least 3
- Phase 4 Handoff: 1

**Step 2: Run a smoke test by reading the skill**

Run: `cat skills/c3-design/SKILL.md | head -100`
Verify: Shows updated Quick Reference with 4 phases and gates

**Step 3: Final commit if any uncommitted changes**

```bash
git status
# If changes exist:
git add skills/c3-design/SKILL.md
git commit -m "docs(c3-design): complete enforcement overhaul"
```

---

## Summary of Changes

| Change | Purpose |
|--------|---------|
| 4-phase Quick Reference with gates | Make phases visible and mandatory |
| Mandatory Phase Tracking section | Force TodoWrite usage |
| ADR gate after Phase 2 | Prevent skipping ADR |
| Phase 3 marked MANDATORY | Explicit requirement |
| ADR verification before doc updates | Technical gate |
| Phase 4 Handoff section | Execute settings.yaml preferences |
| Updated Red Flags | Counter rationalizations |
| Updated Common Mistakes | Document failure modes |

## Test Plan

After implementation, test by:
1. Start a new session
2. Run `/c3-skill:c3` command
3. Verify agent creates TodoWrite items for 4 phases
4. Verify agent cannot skip Phase 3 (ADR)
5. Verify agent executes Phase 4 (Handoff)
