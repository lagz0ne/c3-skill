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
```

Create as `proposed` → human accepts → update layer docs → implement → audit passes → mark `implemented`.

## Mode Selection

| Intent | Mode |
|--------|------|
| "Where is X?" | Navigate |
| "Add/change X" | Design |
| "Audit C3" | Audit |
| No `.c3/` exists | Adopt |

## Document Structure

Every C3 document follows consistent structure:

```
┌─────────────────────────────────────┐
│ 1. DIAGRAM (mermaid)                │
│    Visual grab of connectivity      │
│    Uses IDs for quick reference     │
├─────────────────────────────────────┤
│ 2. INVENTORY TABLE                  │
│    ID | Name | Type | Status        │
├─────────────────────────────────────┤
│ 3. LINKAGES                         │
│    From → To + WHY (reasoning)      │
└─────────────────────────────────────┘
```

## Layer Hierarchy

```
CONTEXT (Strategic)
│   WHAT containers exist
│   WHY they connect (inter-container protocols)
│
└──→ CONTAINER (Bridge)
     │   Fulfills protocol expectations from Context
     │   WHAT components exist
     │   WHY components connect internally
     │
     └──→ COMPONENT (Tactical)
          HOW the protocol is implemented
```

Each layer explains its own flows. No need to reference up.

## Mode: Navigate

ID-based lookup. Parse ID and read corresponding file:

| Pattern | File |
|---------|------|
| `c3-0` | `.c3/README.md` |
| `c3-N` | `.c3/c3-N-*/README.md` |
| `c3-NNN` | `.c3/c3-N-*/c3-NNN-*.md` |
| `adr-YYYYMMDD-slug` | `.c3/adr/adr-YYYYMMDD-slug.md` |

## Mode: Adopt

**Template-first approach.** Two rounds:

### Round 1: Structure (User runs bash)

```bash
PROJECT="MyApp" C1="backend" C2="frontend" ./scripts/c3-init.sh
```

Creates:
```
.c3/
├── README.md                       (Context)
├── c3-1-backend/README.md          (Container)
├── c3-2-frontend/README.md         (Container)
└── adr/adr-00000000-c3-adoption.md (ADR)
```

### Round 2: Fill (AI subagent)

Dispatch subagent to analyze codebase and fill templates:

**Context (c3-0):**
1. Analyze codebase for actors (users, schedulers, webhooks)
2. Confirm containers match code structure
3. Identify external systems (databases, APIs, caches)
4. Draw mermaid diagram with IDs (A1, c3-1, E1, etc.)
5. Fill linkages with REASONING (why they connect)

**Each Container (c3-N):**
1. Analyze container scope for components
2. Categorize by concern:
   - Foundation: entry, identity, integration
   - Auxiliary: library wrappers, framework usage, utilities
   - Business: domain services
   - Presentation: styling, composition, state (or N/A)
3. Draw internal mermaid diagram with component IDs
4. Fill fulfillment section (which components handle Context links)
5. Fill linkages with REASONING

**ADR-000:**
1. Document why C3 was adopted
2. List all containers created
3. Mark verification checklist

### Subagent Prompt Template

```
You are filling C3 templates for {{PROJECT}}.

Templates are already in place at .c3/. Your job:
1. Analyze codebase
2. Fill inventory tables
3. Create mermaid diagrams with IDs
4. Add linkages with reasoning

Rules:
- Diagram goes FIRST, uses IDs from tables
- Every linkage needs REASONING (why, not just that)
- Foundation/Auxiliary/Business/Presentation categories
- Fulfillment section maps Context links to components
- Keep structure, fill content
```

## Mode: Design

Conversational discovery → ADR → Plan → Execute.

### Step 1: Understand

Ask clarifying questions:
- What's the change? (add, modify, remove)
- Why? (problem being solved)
- Scope? (which containers/components)

### Step 2: Scope

Read current `.c3/` structure:
- Identify affected layers (Context, Container, Component)
- Check existing relationships
- Summarize impact for user confirmation

### Step 3: ADR

Generate tactical ADR (see `references/adr-template.md`):
- Problem (2-3 sentences)
- Decision (clear and direct)
- Rationale (tradeoffs considered)
- Affected Layers (document-level)
- Verification checklist

Create: `.c3/adr/adr-YYYYMMDD-{slug}.md` with status `proposed`

### Step 4: Accept

User reviews ADR. On acceptance:
- Update status to `accepted`
- Proceed to Plan

### Step 5: Plan

Generate detailed plan (see `references/plan-template.md`):
- Exact changes per file
- Section-by-section edits
- Order of operations
- Verification steps

Create: `.c3/adr/adr-YYYYMMDD-{slug}.plan.md`

If `superpowers:writing-plans` available, use it.

### Step 6: Execute

Apply changes from plan:
- Update `.c3/` docs in order specified
- Create new component docs if needed
- Update diagrams with correct IDs

### Step 7: Verify

Run audit on affected scope:
- Check diagrams match tables
- Verify linkages have reasoning
- Confirm fulfillment coverage

On pass: Update ADR status to `implemented`

## Mode: Audit

Scopes:
- `audit C3` - full system
- `audit container c3-1` - single container
- `audit adr adr-YYYYMMDD-slug` - single ADR

Checks:
- Diagrams match inventory tables
- All IDs consistent
- Linkages have reasoning
- Fulfillment covers Context links
- No orphan components

## Guidelines

- Diagram first, tables second, linkages third
- Every linkage needs reasoning
- Container fulfills Context links (documents constraints)
- Component documents implementation (technology, conventions, edge cases)
- Inventory ready for growth (empty sections OK)
