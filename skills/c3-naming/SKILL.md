---
name: c3-naming
description: Use when creating or renaming C3 context/container/component/ADR docs to keep names short while encoding parentage (C3 prefix + container digit + component digits) without legacy CON/COM - defines numeric patterns, anchors, examples, red flags, and counters
---

# C3 Naming Conventions

## Overview
Short, container-aware names prevent collisions and make derivation obvious. Drop legacy `CON/COM`; use a stable prefix plus container digit and component digits (`C3-1xx`) to encode parentage. This defines ID/filename/anchor patterns so readers can infer lineage without opening files.

## When to Use
- Creating or renaming C3 docs (CTX/CON/COM/ADR)
- Adding new components to an existing container
- Resolving ID collisions across containers or examples
- Aligning legacy names that lost their parent connection

## Core Principles
- Single context: one `CTX-###-slug` per system.
- Containers own a single digit (extend to two digits if >9 containers): `C3-<C>-slug`.
- Components inherit container digit and add 2-digit sequence: `C3-<C><NN>-slug` (NN zero-padded; extend to three digits if >99 components).
- File names align with IDs; paths mirror hierarchy.
- Downward-only links; names must stand alone outside folder context.

## Quick Reference
| Level | ID Pattern | File Path Pattern | Example |
|-------|------------|-------------------|---------|
| Context | `CTX-###-slug` | `.c3/CTX-###-slug.md` | `CTX-001-system-overview.md` |
| Container (Code/Infra) | `C3-<C>-slug` (`C` = digit) | `.c3/containers/C3-<C>-slug.md` | `C3-1-backend.md` |
| Component | `C3-<C><NN>-slug` (`NN` = 01-99 inside container `C`; use 3 digits if needed) | `.c3/components/<container-slug>/C3-<C><NN>-slug.md` | `C3-101-api-client.md` under `components/backend/` |
| ADR | `ADR-###-slug` | `.c3/adr/ADR-###-slug.md` | `ADR-002-postgresql.md` |

## Search Patterns

Container and component IDs share the `C3-` prefix but differ in digit count. Use these patterns for reliable extraction:

| Element | Simple Search | Regex Pattern | Notes |
|---------|--------------|---------------|-------|
| Context | `CTX-` | `CTX-\d{3}-` | Always 3 digits |
| Container | Find `C3-X-` where X is 1-9 | `C3-\d-[a-z]` | Single digit + dash + letter |
| Component | Find `C3-XXX-` where XXX is 3+ digits | `C3-\d{3,}-` | 3+ digits encode container + sequence |
| ADR | `ADR-` | `ADR-\d{3}-` | Always 3 digits |

**Examples:**
```bash
# Find all containers
grep -E 'C3-[0-9]-[a-z]' .c3/

# Find all components
grep -E 'C3-[0-9]{3,}-' .c3/

# Find components of container 2 specifically
grep -E 'C3-2[0-9]{2}-' .c3/
```

## Flow (small)
```mermaid
flowchart TD
    Need[Creating/renaming doc?] --> Type{Type?}
    Type -->|Context| Ctx[CTX-###-slug<br/>one per system]
    Type -->|Container| Con[C3-<C>-slug<br/>unique digit]
    Type -->|Component| Com[Use parent digit<br/>C3-<C><NN>-slug]
    Type -->|ADR| Adr[ADR-###-slug<br/>optional CTX/CON refs inside]
```

## How to Name
1) Assign/confirm container digit `C` (1–9; extend to two digits if needed). Do not reuse across containers.
2) For components, prefix with `C3-`, container digit, and two/three-digit seq: `C3-<C><NN>-slug` (NN starts at `01` for each container; use three digits if you exceed 99 components).
   - Example: container `C3-2-frontend` → components `C3-201-api-client.md`, `C3-202-auth-guard.md`.
3) Slugs: short, lowercase, hyphenated nouns (no spaces).
4) File paths mirror IDs:
   - Container: `.c3/containers/C3-2-frontend.md`
   - Component: `.c3/components/frontend/C3-201-api-client.md`
5) Update links to use new IDs and anchors (`{#c3-xxx-*}` for containers/components, `{#ctx-xxx-*}` for context, `{#adr-xxx-*}` for ADRs) after renaming.

## Excellent Example (single)
```
Container: C3-1-backend (code)
Components:
- C3-101-db-pool.md      # Resource
- C3-102-auth-middleware.md
- C3-103-task-service.md
Paths:
- .c3/containers/C3-1-backend.md
- .c3/components/backend/C3-101-db-pool.md
```

## Rationalization Table (counters)
| Excuse | Reality & Action |
|--------|------------------|
| "Folder path already shows container; ID can be short" | IDs travel without paths. Encode container digit with prefix in the component ID (e.g., `C3-201`). |
| "We only have one context, collisions unlikely" | Numeric codes collide across examples/tests. Keep the container digit in IDs. |
| "Renaming is overhead; keep legacy COM-001" | Broken links and reader confusion cost more. Rename once; update links immediately. |
| "Numbers are ugly; use names only" | Numbers enable stable references when slugs change. Keep prefix + digit + slug (`C3-203-logger`). |

## Red Flags
- Component ID lacks container digits (e.g., `COM-001-logger`).
- Two containers share the same `CON-###`.
- Slugs omit purpose or exceed 5 hyphen-separated words.
- Links use relative paths without IDs/anchors.

## Common Mistakes
- Mixing container and component digits (e.g., `COM-01-001`).
- Forgetting to move files into matching folders after renaming.
- Using spaces/underscores instead of hyphens.
- Copying examples without updating IDs and anchors.

## Checklist
- IDs follow patterns above (CTX/C3/ADR).
- Components include parent container digit in ID and filename.
- File path matches ID (no mismatches).
- Links updated to new IDs + anchors.
- Anchors follow `{#ctx-*}`, `{#c3-*}`, `{#adr-*}`; include container digit in component anchors (e.g., `{#c3-201-behavior}`).
- One container = one digit; no reuse.
- Slugs stay short, lower-kebab, noun-first.

## Related Skills
- **REQUIRED**: Reference these from c3-adopt, c3-context-design, c3-container-design, c3-component-design when creating docs.
