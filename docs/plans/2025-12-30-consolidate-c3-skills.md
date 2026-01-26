# C3 Skills Consolidation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Consolidate three C3 skills into one workflow-focused skill with reference delegation.

**Architecture:** Single `skills/c3/SKILL.md` (~200 words) focused on decision flow, delegating detailed templates/guidance to `references/`. Agent wrapper remains thin.

**Tech Stack:** Markdown, YAML frontmatter

---

## Current State Analysis

| File | Lines | Purpose | Issue |
|------|-------|---------|-------|
| `skills/c3/SKILL.md` | 314 | Modes, ADR lifecycle | Too verbose, duplicates mode details |
| `skills/c3-structure/SKILL.md` | 235 | Context/Container templates | Used as reference, not invoked |
| `skills/c3-implementation/SKILL.md` | 197 | Component templates | Used as reference, not invoked |
| `agents/c3.md` | 170 | Enhanced capabilities | References old skill paths |

**Total:** 916 lines across 4 files

**Target:** ~400 lines total (skill ~200, references ~150 each, agent ~50)

---

## Task 1: Create Structure Reference

**Files:**
- Create: `references/structure-guide.md`
- Read: `skills/c3-structure/SKILL.md`

**Step 1: Read source skill**

Read `skills/c3-structure/SKILL.md` to extract:
- Context level sections
- Container level sections
- Inventory-first model rules
- Technology stack format
- Diagram guidance

**Step 2: Create reference file**

Write `references/structure-guide.md` with:

```markdown
# C3 Structure Guide

Reference for Context (c3-0) and Container (c3-N) documentation.

## Context Level (c3-0)

**File:** `.c3/README.md`

### Litmus Test

> "Is this about WHY containers exist and HOW they connect?"

### Required Sections

1. **Overview** - System purpose, boundary
2. **Containers** - Inventory table (always complete)
3. **Interactions** - Mermaid diagram
4. **External Actors** - Who/what interacts from outside

### Container Inventory Format

| ID | Name | Responsibility |
|----|------|----------------|
| c3-1 | API Backend | Request handling, business logic |

---

## Container Level (c3-N)

**File:** `.c3/c3-{N}-{slug}/README.md`

### Litmus Test

> "Is this about WHAT components do and HOW they connect inside?"

### Required Sections

1. **Inherited From Context** - What this container is responsible for
2. **Overview** - Container purpose
3. **Technology Stack** - Table only
4. **Components** - 5-column inventory table
5. **Internal Structure** - Mermaid diagram
6. **Key Flows** - 1-2 critical paths

---

## Inventory-First Model

**CRITICAL:** Components table is source of truth.

### Rules

1. Inventory is always complete - list ALL components
2. Docs appear when conventions mature
3. No stubs - full doc or nothing
4. No doc = no consumer conventions

### Components Inventory Format

| ID | Name | Type | Responsibility | Status |
|----|------|------|----------------|--------|
| c3-101 | Auth Service | Business | Authentication | Documented |
| c3-102 | Logger | Foundation | Structured logging | Skip: stdlib |

### Status Values

| Status | Meaning |
|--------|---------|
| *(empty)* | Conventions maturing |
| `Documented` | Full doc exists |
| `Skip: {reason}` | Never needs doc |
| `Combined with c3-XXX` | Documented together |

---

## Technology Stack Format

| Layer | Technology | Purpose |
|-------|------------|---------|
| Runtime | Node.js 20 | JavaScript runtime |
| Framework | Hono | HTTP server |

---

## Diagrams

Use Mermaid only. See `references/diagram-patterns.md`.

| Layer | Required Diagram |
|-------|-----------------|
| Context | Container interactions |
| Container | Internal structure |
```

**Step 3: Verify file created**

Run: `wc -l references/structure-guide.md`
Expected: ~100 lines

**Step 4: Commit**

```bash
git add references/structure-guide.md
git commit -m "refactor(references): extract structure guide from skill"
```

---

## Task 2: Create Implementation Reference

**Files:**
- Create: `references/implementation-guide.md`
- Read: `skills/c3-implementation/SKILL.md`

**Step 1: Read source skill**

Read `skills/c3-implementation/SKILL.md` to extract:
- Component doc structure
- NO CODE enforcement rules
- Required/optional sections
- Foundation vs Business guidance

**Step 2: Create reference file**

Write `references/implementation-guide.md` with:

```markdown
# C3 Implementation Guide

Reference for Component (c3-NNN) documentation.

## When to Create Component Doc

| Has conventions for consumers? | Action |
|-------------------------------|--------|
| Yes - rules consumers must follow | Create doc |
| No - just "use X library" | Inventory entry only |

**No stubs.** Full doc or nothing.

---

## NO CODE Enforcement

Component docs describe HOW, not actual implementation.

| Prohibited | Use Instead |
|------------|-------------|
| `function foo() {}` | Flow diagram |
| `interface X {}` | Table: Field, Type, Purpose |
| JSON/YAML examples | Table with dot notation |
| Config snippets | Settings table |

**Why:** Code changes → doc drift. Tables are scannable.

**Mermaid is allowed** - visual architecture, not data syntax.

---

## Required Sections

### 1. Contract

What Container says about this component.

```markdown
## Contract

From c3-1 (API Backend): "Handles authentication and sessions"
```

### 2. Interface Diagram (REQUIRED)

Shows boundary and hand-offs. Use Mermaid IN/PROCESS/OUT pattern.

### 3. Hand-offs Table

| Direction | What | To/From |
|-----------|------|---------|
| IN | Credentials | Request Handler |
| OUT | Auth Result | Calling Service |

### 4. Conventions Table

| Rule | Why |
|------|-----|
| Always validate before processing | Security |

### 5. Edge Cases & Errors

| Scenario | Behavior |
|----------|----------|
| Invalid credentials | Return 401, log attempt |

---

## Optional Sections

Include only if relevant:
- **Configuration** - Significant config surface
- **Dependencies** - External dependencies matter
- **State/Lifecycle** - Component has states
- **Performance** - Throughput/latency matters

---

## Foundation vs Business

| Type | Doc Focus |
|------|-----------|
| **Foundation** | What it PROVIDES, interface conventions |
| **Business** | Processing flow, domain rules, edge cases |
```

**Step 3: Verify file created**

Run: `wc -l references/implementation-guide.md`
Expected: ~90 lines

**Step 4: Commit**

```bash
git add references/implementation-guide.md
git commit -m "refactor(references): extract implementation guide from skill"
```

---

## Task 3: Rewrite Consolidated Skill

**Files:**
- Modify: `skills/c3/SKILL.md`

**Step 1: Write new skill focused on workflow/decisions**

```markdown
---
name: c3
description: |
  Use when working with C3 architecture docs - navigating, understanding, designing, or auditing.
  Triggers: "C3", "architecture", "where is X documented", "impact of changing X".
---

# C3 Architecture Assistant

## ADR Lifecycle

```
proposed → accepted → implemented
    ↓         ↓           ↓
 Create    Human      After audit
           accepts    passes
```

You can only:
- Create ADRs as `proposed`
- Move to `accepted` after human confirms → then update layer docs
- Move to `implemented` only after audit passes

---

## Mode Selection

| Intent | Mode | Action |
|--------|------|--------|
| "Where is X?" | Navigate | Path + summary |
| "How does X work?" | Understand | Explain with citations |
| "Show architecture" | Overview | System summary |
| "Add/change X" | Design | Impact → ADR (proposed) |
| "Accept ADR" | Lifecycle | Update status + layer docs |
| "Mark implemented" | Lifecycle | Audit first |
| "Audit C3" | Audit | Health check |

---

## Layer Decision

| Working on... | Layer | Reference |
|---------------|-------|-----------|
| Container inventory, relationships | Context (c3-0) | `references/structure-guide.md` |
| Component inventory, tech stack | Container (c3-N) | `references/structure-guide.md` |
| Component conventions, hand-offs | Component (c3-NNN) | `references/implementation-guide.md` |

---

## Mode: Adopt

**Trigger:** No `.c3/` exists

1. Create `.c3/adr/` directory
2. Create Context doc (`.c3/README.md`) - see `references/structure-guide.md`
3. Create Container docs - see `references/structure-guide.md`

**CRITICAL:** Inventory-first. NO component docs at adopt time.

---

## Mode: Design (Analyze + ADR)

1. **Discover** - What, Why, Where?
2. **Assess Impact** - Which layers affected?
3. **ADR Decision** - Crosses boundaries? Changes contracts? → ADR needed
4. **Create ADR** - Use `references/adr-template.md`
5. **Handoff** - Provide context for implementation

---

## Mode: Lifecycle

### proposed → accepted

1. Update ADR status
2. Parse "Changes Across Layers"
3. Update affected layer docs (read reference for structure)
4. Update Audit Record

### accepted → implemented

1. Run audit (verification checklist)
2. If PASS → update status
3. If FAIL → report issues, no status change

---

## Mode: Audit

Check docs against code. See `references/audit-checks.md`.

| Check | Fail Criteria |
|-------|---------------|
| Frontmatter | Missing required fields |
| ID Pattern | Wrong format |
| Inventory vs Code | Missing module |
| Structure Integrity | Missing parent |

---

## Guidelines

- Fast for Navigate/Understand
- Thorough for Design
- Never skip ADR lifecycle steps
- Cite specific files
```

**Step 2: Verify word count**

Run: `wc -w skills/c3/SKILL.md`
Expected: <250 words (target <200)

**Step 3: Commit**

```bash
git add skills/c3/SKILL.md
git commit -m "refactor(skills): consolidate c3 skill with reference delegation"
```

---

## Task 4: Update Agent

**Files:**
- Modify: `agents/c3.md`

**Step 1: Update agent to reference new structure**

The agent should:
- Load skill for core behavior
- Add discovery-based adopt (Task + AskUserQuestion)
- Add drift detection audit (Task)
- Reference new paths

**Step 2: Slim agent further**

Remove any duplicated content, keep only enhancements.

**Step 3: Commit**

```bash
git add agents/c3.md
git commit -m "refactor(agents): update c3 agent for new reference paths"
```

---

## Task 5: Delete Old Skills

**Files:**
- Delete: `skills/c3-structure/SKILL.md`
- Delete: `skills/c3-structure/structure-template.md` (if exists)
- Delete: `skills/c3-implementation/SKILL.md`
- Delete: `skills/c3-implementation/implementation-template.md` (if exists)
- Delete: directories if empty

**Step 1: Remove files**

```bash
rm -rf skills/c3-structure
rm -rf skills/c3-implementation
```

**Step 2: Verify removal**

```bash
ls skills/
```

Expected: Only `c3/` directory remains

**Step 3: Commit**

```bash
git add -A
git commit -m "refactor(skills): remove deprecated layer skills"
```

---

## Task 6: Update Plugin Manifest

**Files:**
- Check: `.claude-plugin/plugin.json`

**Step 1: Verify skill entries**

Check if plugin.json references the old skills. If so, remove them.

**Step 2: Commit if changed**

```bash
git add .claude-plugin/plugin.json
git commit -m "chore(plugin): update skill entries after consolidation"
```

---

## Task 7: Verify Build

**Step 1: Run OpenCode build**

```bash
bun run build:opencode
```

**Step 2: Check for errors**

Expected: Build succeeds, `dist/opencode-c3/` updated

**Step 3: Commit dist if needed**

(Only if dist is tracked)

---

## Summary

| Before | After |
|--------|-------|
| 4 files, 916 lines | 4 files, ~400 lines |
| 3 skills | 1 skill |
| Duplicated content | Reference delegation |
| Skill used as reference | Proper references/ files |

**Line reduction:** ~56%
