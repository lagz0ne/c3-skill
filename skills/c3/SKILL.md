---
name: c3
description: |
  Use when project needs C3 adoption (no .c3/) or auditing existing C3 docs.
  Triggers: "adopt C3", "onboard", "scaffold", "audit", "validate", "check C3", "init".
  NOT for navigation (use c3-query) or changes (use c3-alter).
---

# C3 Architecture Assistant

## Skill Routing

| Intent | Skill |
|--------|-------|
| "Where is X?" / Explore | `c3-query` |
| "Add/change X" | `c3-alter` |
| "Audit C3" | This skill |
| No `.c3/` exists | This skill (Adopt) |

## REQUIRED: Check Activation

**Before proceeding, check if `.c3/README.md` exists.**

- If NO `.c3/`: This is **Mode: Adopt**. Suggest `/onboard` or proceed with adoption.
- If YES `.c3/`: Route based on intent (Audit stays here, others route to dedicated skills).

## ADR Lifecycle

`proposed` → `accepted` → `implemented`

See `../../references/adr-template.md` for template.

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
2. Categorize by: Foundation (primitives) / Auxiliary (conventions) / Feature (domain-specific)
3. Draw internal mermaid diagram with component IDs
4. Fill fulfillment section (which components handle Context links)
5. Fill linkages with REASONING

**Each Component (c3-NNN):**
1. Create file: `.c3/c3-N-slug/c3-NNN-component-slug.md`
2. Use sequential IDs: c3-101, c3-102, etc.
3. Use template by category:
   - Foundation → `../../templates/component-foundation.md`
   - Auxiliary → `../../templates/component-auxiliary.md`
   - Feature → `../../templates/component-feature.md`
4. After drafting, add `## References` pointing to implementation code (symbols first, then patterns, then paths; 3-7 items typical)

**ADR-000:**
1. Document why C3 was adopted
2. List all containers created
3. Mark verification checklist

**Reference:** See container patterns reference below for component categorization.

### Subagent Prompt Template

When dispatching a subagent, **you** (the main agent) must first read `../../references/container-patterns.md` and include its content in the prompt below:

```
## Context

You are filling C3 templates for {{PROJECT}}.

### Container Patterns (for reference)
{{PASTE_CONTAINER_PATTERNS_CONTENT_HERE}}

Your job:
1. Analyze codebase
2. Fill inventory tables (Foundation/Auxiliary/Feature)
3. Create mermaid diagrams with IDs
4. Add linkages with reasoning
5. Create component files using category-specific templates

Rules:
- Diagram goes FIRST, uses IDs from tables
- Every linkage needs REASONING
- Sequential component IDs: c3-101, c3-102, etc.
```

## Mode: Design

**Defer to `c3-alter` skill** for all architecture and code changes.

`c3-alter` enforces ADR-first workflow with impact discovery.

## Mode: Audit

**REQUIRED:** Load `../../references/audit-checks.md` for comprehensive audit procedures.

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

## Guidelines

- Diagram first, tables second, linkages third
- Every linkage needs reasoning
- Container fulfills Context links (documents constraints)
- Component documents implementation (technology, conventions, edge cases)
- Inventory ready for growth (empty sections OK)
- References are a lookup index from architecture to code; keep them concise (3-7 items)
