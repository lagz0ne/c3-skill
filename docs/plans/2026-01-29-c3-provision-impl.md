# C3 Provision Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add c3-provision skill for architecture-first workflow where components can be documented before implementation.

**Architecture:** New skill (c3-provision) routes provisioning requests through orchestrator, stops after ADR acceptance + doc creation. Provisioned docs live in `.c3/provisioned/` directory. ADR lifecycle extended with `provisioned` and `superseded` statuses.

**Tech Stack:** Markdown skills, YAML frontmatter, existing c3 agents

---

## Task 1: Extend ADR Template with New Status Values

**Files:**
- Modify: `references/adr-template.md`

**Step 1: Read current ADR template**

Read `references/adr-template.md` to understand current structure.

**Step 2: Add provisioned and superseded status values**

Update the Status Values table to include:

```markdown
## Status Values

| Status | Meaning | Gate Behavior | base-commit |
|--------|---------|---------------|-------------|
| `proposed` | Awaiting review | Files NOT editable | Not set |
| `accepted` | Ready for implementation | All code editable | Captured (HEAD at acceptance) |
| `implemented` | Changes applied, verified | Gate relaxed | Used for verification |
| `provisioned` | Design complete, no code yet | Files NOT editable | Not set |
| `superseded` | Replaced by another ADR | Files NOT editable | N/A |
```

**Step 3: Add implements/superseded-by fields to template**

Add to the template frontmatter:

```yaml
---
id: adr-YYYYMMDD-{slug}
title: [Decision Title]
status: proposed
date: YYYY-MM-DD
affects: [c3-0, c3-1]
approved-files: []
base-commit:
implements:           # NEW: Link to provisioned ADR (for implementation ADRs)
superseded-by:        # NEW: Link to implementation ADR (for provisioned ADRs)
---
```

**Step 4: Update Lifecycle diagram**

```markdown
## Lifecycle

```
proposed → accepted → implemented
              ↓
        Generate Plan
              ↓
        Execute Plan
              ↓
        Run Audit

proposed → accepted → provisioned (design-only path)
                          ↓
                    (later) implementation ADR
                          ↓
                    superseded
```
```

**Step 5: Commit**

```bash
git add references/adr-template.md
git commit -m "feat(adr): add provisioned and superseded status values

Extends ADR lifecycle for architecture-first workflow:
- provisioned: design complete, no code yet
- superseded: replaced by implementation ADR
- implements/superseded-by: ADR linking fields"
```

---

## Task 2: Create Component Lifecycle Reference

**Files:**
- Create: `references/component-lifecycle.md`

**Step 1: Create component lifecycle reference**

```markdown
# Component Lifecycle

Components have a lifecycle status indicating their implementation state.

## Status Values

| Status | Meaning | Code References | Location |
|--------|---------|-----------------|----------|
| `provisioned` | Documented, not implemented | None | `.c3/provisioned/` |
| `active` | Implemented, in use | Required | `.c3/c3-X-*/` |
| `deprecated` | Being phased out | May have | `.c3/c3-X-*/` |

## File Organization

### Active Components

Live in standard container directories:

```
.c3/
  c3-1-frontend/
    README.md
    c3-101-component.md    # status: active
  c3-2-api/
    README.md
    c3-201-component.md    # status: active
```

### Provisioned Components

Live in parallel `.c3/provisioned/` directory:

```
.c3/
  provisioned/
    c3-1-frontend/
      c3-105-new-feature.md    # status: provisioned (new)
    c3-2-api/
      c3-201-component.md      # status: provisioned (planned update)
```

**Key insight:** Provisioned directory mirrors container structure. Same ID can exist in both locations - active version vs provisioned version.

## Frontmatter

### Provisioned Component (New)

```yaml
---
id: c3-205-rate-limiter
status: provisioned
adr: adr-20260129-rate-limiter
---
```

### Provisioned Component (Update to Existing)

```yaml
---
id: c3-201-auth
status: provisioned
supersedes: ../c3-2-api/c3-201-auth.md
adr: adr-20260129-oauth-support
---
```

### Active Component

```yaml
---
id: c3-201-auth
status: active
---
```

## Transitions

### Provision → Active (Implementation)

1. Create implementation ADR with `implements: <provisioned-adr>`
2. Execute implementation
3. Move file: `.c3/provisioned/c3-X/c3-XXX.md` → `.c3/c3-X/c3-XXX.md`
4. Update component status: `provisioned` → `active`
5. Add `## Code References` section
6. Update provisioned ADR: `superseded-by: <implementation-adr>`

### Active → Deprecated

1. Create deprecation ADR
2. Update component status: `active` → `deprecated`
3. Document replacement/migration path

## Audit Rules

| Rule | Check |
|------|-------|
| Provisioned has no Code References | `.c3/provisioned/**` files must NOT have `## Code References` |
| Active has Code References | `.c3/c3-*/**` files (except README) MUST have `## Code References` |
| Stale provisioned | Warn if active component changed after provisioned version created |
| Orphaned provisioned | Error if ADR is superseded but provisioned file not moved |
```

**Step 2: Commit**

```bash
git add references/component-lifecycle.md
git commit -m "docs(refs): add component lifecycle reference

Documents provisioned/active/deprecated status model and
.c3/provisioned/ directory structure for architecture-first workflow."
```

---

## Task 3: Create c3-provision Skill

**Files:**
- Create: `skills/c3-provision/SKILL.md`

**Step 1: Create skill directory**

```bash
mkdir -p skills/c3-provision
```

**Step 2: Create SKILL.md**

```markdown
---
name: c3-provision
description: |
  Documents architecture WITHOUT implementation. Creates ADR and component docs with status: provisioned.

  MUST use this skill when user wants to DESIGN or PLAN without implementing:
  - "provision X", "provision a component"
  - "design X", "design the architecture for X"
  - "plan X", "plan a new service"
  - "document X architecture"
  - "create architecture docs for X"
  - "envision X", "architect X"

  Key indicator: User wants documentation/design output, NOT code changes.
  Requires .c3/ to exist. Creates ADR with status: provisioned.
  For implementation requests, route to c3-alter.
---

# C3 Provision - Architecture Without Implementation

**Document architecture before code exists.** ADR captures design decision, component docs created with status: provisioned.

**Relationship to c3-orchestrator agent:** This skill uses orchestrator for analysis but stops after ADR acceptance. No c3-dev execution.

## REQUIRED: Load References

Before proceeding, use Glob to find and Read these files:
1. `**/references/skill-harness.md` - Red flags and complexity rules
2. `**/references/layer-navigation.md` - How to traverse C3 docs
3. `**/references/adr-template.md` - ADR structure
4. `**/references/component-lifecycle.md` - Provisioned status model

## Core Workflow

```
User Request ("provision X", "design X", "plan X")
    ↓
Stage 1: Intent Clarification (Socratic)
    ↓
Stage 2: Current State (analyze existing architecture)
    ↓
Stage 3: Scope Impact (what containers/components affected)
    ↓
Stage 4: Create ADR (status: proposed)
    ↓
Stage 4b: User Accepts ADR
    ↓
Stage 5: Create Provisioned Docs
    ├── New component: .c3/provisioned/c3-X/c3-xxx.md
    └── Update to existing: .c3/provisioned/c3-X/c3-xxx.md (with supersedes:)
    ↓
Stage 6: Update ADR status: provisioned
    ↓
DONE (no execution phase)
```

## Progress Checklist

Copy and track as you work:

```
Provision Progress:
- [ ] Stage 1: Intent clarified (what to provision, why)
- [ ] Stage 2: Current state documented
- [ ] Stage 3: Scope assessed (affected containers/components)
- [ ] Stage 4: ADR created (status: proposed)
- [ ] Stage 4b: ADR accepted by user
- [ ] Stage 5: Provisioned docs created
- [ ] Stage 6: ADR status updated to provisioned
```

---

## Stage 1: Intent

| Step | Action |
|------|--------|
| Analyze | What component/feature? New or update? Why provision vs implement? |
| Ask | Use AskUserQuestion: What problem does this solve? Scope? |
| Synthesize | `Intent: Provision [component] Goal: [outcome] Type: [new/update]` |
| Review | User confirms or corrects |

**Key question:** "Are you looking to document the design now and implement later, or do you want to implement immediately?"

---

## Stage 2: Current State

| Step | Action |
|------|--------|
| Analyze | Read affected C3 docs via layer navigation |
| Ask | Any existing components this relates to? |
| Synthesize | List related components, their behavior, dependencies |
| Review | User confirms or corrects |

---

## Stage 3: Scope Impact

| Step | Action |
|------|--------|
| Analyze | Which containers affected? New or update? |
| Ask | Clarify boundaries, dependencies |
| Synthesize | List all affected c3 IDs |
| Review | User confirms scope |

---

## Stage 4: Create ADR

Generate at `.c3/adr/adr-YYYYMMDD-{slug}.md`.

**Provisioning ADR template:**

```yaml
---
id: adr-YYYYMMDD-{slug}
title: [Design Decision Title]
status: proposed
date: YYYY-MM-DD
affects: [c3 IDs]
approved-files: []
base-commit:
implements:
superseded-by:
---
```

**Key sections:**
- Problem (why this design is needed)
- Decision (what we're provisioning)
- Rationale (design tradeoffs)
- Affected Layers (what docs will be created)

**Note:** `approved-files` stays empty for provisioning (no code changes).

---

## Stage 4b: ADR Acceptance

Use AskUserQuestion:

```
question: "Review the ADR. Ready to accept and create provisioned docs?"
options:
  - "Accept - create provisioned component docs"
  - "Revise - update scope or decision"
  - "Cancel - abandon this provisioning"
```

---

## Stage 5: Create Provisioned Docs

### For NEW component:

Create at `.c3/provisioned/c3-X-container/c3-XXX-name.md`:

```yaml
---
id: c3-XXX-name
status: provisioned
adr: adr-YYYYMMDD-{slug}
---

# [Component Name] (Provisioned)

## Purpose

[What this component will do]

## Behavior

[Expected behavior when implemented]

## Dependencies

- Uses: [c3 IDs of components this will use]
- Used by: [c3 IDs of components that will use this]

## Notes

This component is provisioned (designed) but not yet implemented.
Implementation will be tracked via a separate ADR.
```

**IMPORTANT:** Do NOT include `## Code References` section - provisioned components have no code.

### For UPDATE to existing component:

Create at `.c3/provisioned/c3-X-container/c3-XXX-name.md`:

```yaml
---
id: c3-XXX-name
status: provisioned
supersedes: ../c3-X-container/c3-XXX-name.md
adr: adr-YYYYMMDD-{slug}
---

# [Component Name] (Provisioned Update)

[Full component doc as it will be after implementation]
```

### Update container README (optional)

If creating new component, add to container's README under "Provisioned" section:

```markdown
## Provisioned Components

Components designed but not yet implemented:

| ID | Name | ADR |
|----|------|-----|
| c3-XXX | Name | adr-YYYYMMDD-slug |
```

---

## Stage 6: Finalize ADR

Update ADR frontmatter:

```yaml
status: provisioned    # Changed from proposed
```

Add to end of ADR:

```markdown
## Provisioned

Component docs created:
- `.c3/provisioned/c3-X/c3-XXX.md`

To implement this design, create a new ADR with `implements: adr-YYYYMMDD-{slug}`.
```

---

## Examples

**Example 1: Provision new component**

```
User: "provision a rate limiter for the API"

Stage 1 - Intent:
  Intent: Provision rate limiter
  Goal: Document rate limiting design before implementing
  Type: New component

Stage 2 - Current State:
  Related: c3-2-api (API Backend)
  No existing rate limiting

Stage 3 - Scope:
  New: c3-206-rate-limiter in c3-2-api
  Integrates with: c3-201-auth-middleware

Stage 4 - ADR:
  Created: .c3/adr/adr-20260129-rate-limiter.md
  Status: proposed

Stage 4b - Accept:
  User accepts

Stage 5 - Create Docs:
  Created: .c3/provisioned/c3-2-api/c3-206-rate-limiter.md
  Status: provisioned

Stage 6 - Finalize:
  ADR status: provisioned
```

**Example 2: Provision update to existing**

```
User: "design OAuth support for the auth middleware"

Stage 1 - Intent:
  Intent: Provision OAuth addition
  Goal: Document OAuth design before implementing
  Type: Update to existing c3-201

Stage 2 - Current State:
  Existing: c3-201-auth-middleware (basic auth only)

Stage 3 - Scope:
  Update: c3-201-auth-middleware
  No new components

Stage 4 - ADR:
  Created: .c3/adr/adr-20260129-oauth-support.md

Stage 4b - Accept:
  User accepts

Stage 5 - Create Docs:
  Created: .c3/provisioned/c3-2-api/c3-201-auth-middleware.md
  (supersedes: ../c3-2-api/c3-201-auth-middleware.md)

Stage 6 - Finalize:
  ADR status: provisioned
```

---

## Response Format

```
**Stage N: {Name}**
{findings}
**Open Questions:** {list or "None - confident"}
**Next:** {what happens next}
```
```

**Step 3: Commit**

```bash
git add skills/c3-provision/SKILL.md
git commit -m "feat(skill): add c3-provision for architecture-first workflow

New skill for documenting architecture before implementation:
- Full orchestrator analysis
- ADR with status: provisioned
- Component docs in .c3/provisioned/
- No code execution phase"
```

---

## Task 4: Update c3 Router Skill

**Files:**
- Modify: `skills/c3/SKILL.md`

**Step 1: Read current c3 skill**

Read `skills/c3/SKILL.md` to understand routing.

**Step 2: Add provision routing**

Update the routing rules in the description:

```yaml
description: |
  PRIMARY ROUTER for C3 architecture tasks and audit tool for documentation consistency.

  ROUTING RULES (check in order):
  1. **PROVISION-RELATED** (design without implementing):
     → route to c3-provision skill (e.g., "provision X", "design X", "plan X architecture")
  2. **REF-RELATED** (contains "ref", "refs", "pattern", "patterns", "convention", "standard"):
     → MUST route to c3-ref skill (e.g., "what ref", "show refs", "what patterns")
  3. **CHANGES** (add/modify/remove/fix/refactor/implement):
     → route to c3-alter skill
  4. **QUESTIONS** (where/what/how/explain about components):
     → route to c3-query skill
  5. **NO .c3/** directory:
     → route to onboard skill

  AUDIT MODE: Use when user asks to "audit architecture"...
```

Update the Intent Recognition table:

```markdown
## Intent Recognition & Routing

| User Says | Intent | Route To |
|-----------|--------|----------|
| "provision/design/plan/envision/architect X" | Design-only | `/c3-provision` |
| "where/what/how/explain/show me" | Question | `/c3-query` |
| "add/modify/remove/fix/refactor/implement" | Change | `/c3-alter` |
| "pattern/convention/standard/ref/how should we" | Pattern | `/c3-ref` |
| "audit/validate/check/verify/sync" | Audit | this skill |
| (no .c3/ directory) | Initialize | `/onboard` |

**When unclear:** Ask "Do you want to explore (query), design only (provision), change and implement (alter), manage patterns (ref), or audit?"
```

**Step 3: Commit**

```bash
git add skills/c3/SKILL.md
git commit -m "feat(router): add c3-provision routing

Routes provision/design/plan requests to new c3-provision skill
for architecture-first workflow."
```

---

## Task 5: Update c3-alter to Detect Provisioned Components

**Files:**
- Modify: `skills/c3-alter/SKILL.md`

**Step 1: Read current c3-alter skill**

Read `skills/c3-alter/SKILL.md`.

**Step 2: Add provisioned component detection**

Add new section after Stage 2:

```markdown
## Stage 2b: Detect Provisioned Components

Check if component has a provisioned version:

```bash
# Check for provisioned version
ls .c3/provisioned/c3-*/c3-XXX-*.md 2>/dev/null
```

**If provisioned version exists:**

```
AskUserQuestion:
  question: "Found provisioned design for this component at .c3/provisioned/... Implement from this design?"
  options:
    - "Yes - use provisioned design as starting point"
    - "No - start fresh (will supersede provisioned)"
```

**On "Yes":**
1. Load provisioned doc as starting point
2. ADR frontmatter: `implements: <provisioned-adr>`
3. After implementation:
   - Move provisioned file to active location
   - Update provisioned ADR: `superseded-by: <this-adr>`

**On "No":**
1. Proceed with normal alter flow
2. After implementation:
   - Delete orphaned provisioned file
   - Update provisioned ADR: `superseded-by: <this-adr>` with note "Design not used"
```

**Step 3: Update Stage 6 for provisioned promotion**

Add to Stage 6 Execute:

```markdown
### Promoting Provisioned Components

If implementing a provisioned component:

1. **Move file:**
   ```bash
   mv .c3/provisioned/c3-X/c3-XXX.md .c3/c3-X/c3-XXX.md
   ```

2. **Update frontmatter:**
   ```yaml
   status: active    # Changed from provisioned
   # Remove supersedes: field
   # Remove adr: field (or update to implementation ADR)
   ```

3. **Add Code References section**

4. **Update provisioned ADR:**
   ```yaml
   status: superseded
   superseded-by: adr-YYYYMMDD-implementation
   ```
```

**Step 4: Commit**

```bash
git add skills/c3-alter/SKILL.md
git commit -m "feat(alter): detect and promote provisioned components

c3-alter now:
- Detects provisioned versions before changes
- Links implementation ADR via implements: field
- Promotes provisioned docs to active on implementation
- Marks provisioned ADR as superseded"
```

---

## Task 6: Update c3-navigator to Filter Provisioned

**Files:**
- Modify: `agents/c3-navigator.md`

**Step 1: Read current c3-navigator**

Read `agents/c3-navigator.md`.

**Step 2: Add provisioned filtering behavior**

Add to Step 2 (Read Context Layer):

```markdown
### Step 2: Read Context Layer

Always start by reading the C3 context:

```
.c3/README.md   - System overview, actors, containers
.c3/TOC.md      - Full table of contents (if exists)
```

**Provisioned Content:**
- Default: Search only `.c3/c3-*` (active components)
- If user asks "what's planned?", "show provisioned", "what's designed?": Include `.c3/provisioned/`
- If querying a component that has a provisioned version: Warn user

```
Note: Component c3-201-auth has a provisioned update at .c3/provisioned/c3-2-api/c3-201-auth.md
Would you like to see the planned changes?
```
```

**Step 3: Update Edge Cases**

```markdown
## Edge Cases

| Situation | Action |
|-----------|--------|
| No .c3/ directory | Suggest using `/onboard` to create C3 docs |
| Question not in docs | State "not documented", offer to search code |
| Spans multiple containers | List all involved, show cross-container diagram |
| Very complex question | Break into sub-questions, answer each |
| **Component has provisioned version** | Warn user, offer to show planned changes |
| **User asks about planned/provisioned** | Include `.c3/provisioned/` in search |
```

**Step 4: Commit**

```bash
git add agents/c3-navigator.md
git commit -m "feat(navigator): filter provisioned by default, surface on request

Navigator now:
- Searches only active components by default
- Warns when component has provisioned version
- Includes provisioned/ on explicit request"
```

---

## Task 7: Update c3-adr-transition for Superseded Status

**Files:**
- Modify: `agents/c3-adr-transition.md`

**Step 1: Read current transition agent**

Read `agents/c3-adr-transition.md`.

**Step 2: Add superseded transition support**

Add new workflow section:

```markdown
## Workflow: Supersede Provisioned ADR

When an implementation ADR completes, the provisioned ADR it implements should be superseded.

### Step 1: Check for implements link

```bash
# In implementation ADR, check for implements: field
grep "^implements:" .c3/adr/adr-*.md
```

### Step 2: Update provisioned ADR

If `implements:` links to a provisioned ADR:

1. Read the provisioned ADR
2. Update frontmatter:
   ```yaml
   status: superseded
   superseded-by: <implementation-adr-id>
   ```
3. Add note to Status section:
   ```markdown
   ## Status

   **Superseded** - YYYY-MM-DD

   This design was implemented via [adr-YYYYMMDD-implementation](./adr-YYYYMMDD-implementation.md).
   ```
```

Update Error Handling:

```markdown
## Error Handling

**If ADR status is `provisioned`:**
- This is a design-only ADR
- Cannot transition to `implemented` (no code to verify)
- Inform user: "This is a provisioned (design-only) ADR. To implement, create a new ADR with `implements: <this-adr>`"
```

**Step 3: Commit**

```bash
git add agents/c3-adr-transition.md
git commit -m "feat(transition): support superseded status for provisioned ADRs

Transition agent now:
- Recognizes provisioned ADRs (cannot implement directly)
- Supersedes provisioned ADR when implementation completes
- Links ADRs via superseded-by field"
```

---

## Task 8: Add Test Prompts for c3-provision

**Files:**
- Create: `tests/skill-triggering/prompts/provision-new.txt`
- Create: `tests/skill-triggering/prompts/provision-design.txt`
- Create: `tests/skill-triggering/prompts/provision-plan.txt`
- Modify: `tests/skill-triggering/run-all.sh`

**Step 1: Create provision test prompts**

`provision-new.txt`:
```
provision a rate limiter component for the API
```

`provision-design.txt`:
```
design the architecture for a notification service
```

`provision-plan.txt`:
```
plan a new caching layer - don't implement yet, just document the design
```

**Step 2: Update run-all.sh**

Add provision tests:

```bash
# Provision routing tests → c3-provision
run_test "c3-provision" "$PROMPTS_DIR/provision-new.txt"
run_test "c3-provision" "$PROMPTS_DIR/provision-design.txt"
run_test "c3-provision" "$PROMPTS_DIR/provision-plan.txt"
```

**Step 3: Commit**

```bash
git add tests/skill-triggering/prompts/provision-*.txt
git add tests/skill-triggering/run-all.sh
git commit -m "feat(tests): add c3-provision routing tests

Tests provision/design/plan triggers route to c3-provision skill."
```

---

## Task 9: Final Integration Test

**Files:**
- None (verification only)

**Step 1: Run skill triggering tests**

```bash
./tests/skill-triggering/run-all.sh
```

Expected: provision tests pass, routing to c3-provision.

**Step 2: Manual verification**

Test the full provision workflow manually:

```bash
# In a project with .c3/
claude -p "provision a new logging service for the API - just design, don't implement"
```

Expected:
- Routes to c3-provision skill
- Asks clarifying questions (Socratic)
- Creates ADR with status: proposed
- On acceptance, creates `.c3/provisioned/` doc
- ADR status becomes `provisioned`
- No code execution

**Step 3: Verify file structure**

```bash
ls .c3/provisioned/
# Should show container directories with provisioned component docs
```

**Step 4: Final commit**

```bash
git add -A
git commit -m "feat: complete c3-provision implementation

Architecture-first workflow now available:
- c3-provision skill for design without implementation
- ADR status: provisioned for design-only decisions
- .c3/provisioned/ directory for staged docs
- Navigator filters provisioned by default
- Alter detects and promotes provisioned components
- Transition handles superseded status"
```

---

## Summary

| Task | Description | Files |
|------|-------------|-------|
| 1 | Extend ADR template | `references/adr-template.md` |
| 2 | Component lifecycle reference | `references/component-lifecycle.md` |
| 3 | Create c3-provision skill | `skills/c3-provision/SKILL.md` |
| 4 | Update c3 router | `skills/c3/SKILL.md` |
| 5 | Update c3-alter for provisioned | `skills/c3-alter/SKILL.md` |
| 6 | Update navigator filtering | `agents/c3-navigator.md` |
| 7 | Update ADR transition | `agents/c3-adr-transition.md` |
| 8 | Add test prompts | `tests/skill-triggering/prompts/provision-*.txt` |
| 9 | Integration test | Verify full workflow |

**Total: 9 tasks, ~7 file changes**
