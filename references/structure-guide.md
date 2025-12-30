# C3 Structure Guide

Reference for Context and Container level structure.

---

## Frontmatter Format

**CRITICAL: Copy these exactly. Do not modify field names or value formats.**

### Context (c3-0)

```yaml
---
id: c3-0
c3-version: 3
title: {System Name}
summary: {One line description}
---
```

### Container (c3-N)

```yaml
---
id: c3-{N}
c3-version: 3
title: {Container Name}
type: container
parent: c3-0
summary: {One line description}
---
```

**MANDATORY FORMAT RULES:**
- `c3-version: 3` — ALWAYS the integer `3`, NEVER a string like `"1.0"` or `"3"`
- `type: container` — REQUIRED for all containers
- `parent: c3-0` — REQUIRED for all containers
- No extra fields, no missing fields

---

## Context Level (c3-0)

**File:** `.c3/README.md`

**Litmus Test:** "Is this about WHY containers exist and HOW they connect?" Yes: Context. No: Push to Container.

**Required Sections:** Overview, Containers (inventory), Interactions (Mermaid), External Actors

### Container Inventory Format

```markdown
| ID | Name | Responsibility |
|----|------|----------------|
| c3-1 | API Backend | Request handling, business logic |
| c3-2 | Frontend | User interface, client state |
```

**Rule:** List ALL containers. This is the source of truth.

---

## Container Level (c3-N)

**File:** `.c3/c3-{N}-{slug}/README.md`

**Litmus Test:** "Is this about WHAT components do and HOW they connect inside this container?" Yes: Container. Container relationships: Context. How components work: Component.

**Required Sections:** Technology Stack, Components (inventory), Internal Structure (Mermaid)

**Optional Sections:** Inherited From Context, Overview, Key Flows (include when they add clarity)

---

## Inventory-First Model

**CRITICAL:** The inventory table is the source of truth.

### Core Principle

```
Inventories are DISCOVERED, not DESIGNED.
They reflect WHAT EXISTS, compared against WHAT'S DECLARED.
```

### Container Inventory Rules

1. **Inventory is always complete** - List ALL containers, even if new
2. **Overlay comparison** - Compare declared (in .c3/) vs discovered (in code)
3. **User confirms** - Don't scaffold docs until inventory is FIRM

### Component Inventory Rules

1. **Inventory is always complete** - List ALL components, even without detailed docs
2. **Docs appear when conventions mature** - Component doc = conventions exist for consumers
3. **No stubs** - Either a full doc exists or it doesn't
4. **No doc = no consumer conventions** - Just "use it" (e.g., standard logger)

---

## Overlay Comparison

When adopting or auditing, always compare declared vs discovered:

### Overlay Status Semantics

| Status | Symbol | Meaning | Action |
|--------|--------|---------|--------|
| Match | `✓` | Declared item found in code | Keep, optionally update path |
| New | `✚` | Found in code, not declared | Add to inventory or explain why not |
| Missing | `⚠` | Declared but not found in code | Remove, rename, or investigate drift |

### Container Overlay Example

```
CONTAINER INVENTORY OVERLAY
═══════════════════════════════════════════════════════════════

Declared (in .c3/)          Discovered (in code)        Status
───────────────────────────────────────────────────────────────
c3-1 API Backend            services/api                ✓ Match
c3-2 Frontend               frontend                    ✓ Match
-                           services/worker             ✚ NEW
c3-3 Legacy Adapter         -                           ⚠ MISSING

═══════════════════════════════════════════════════════════════
```

### Component Overlay Example

```
COMPONENT INVENTORY: c3-1 API Backend
═══════════════════════════════════════════════════════════════

Declared (in .c3/)          Discovered (in code)        Status
───────────────────────────────────────────────────────────────
c3-101 Request Handler      src/handlers                ✓ Match
c3-102 Auth Service         src/services/auth           ✓ Match
-                           src/services/orders         ✚ NEW
c3-103 Legacy Auth          -                           ⚠ MISSING

═══════════════════════════════════════════════════════════════
```

---

## Components Inventory Format

**CRITICAL: Use exactly these 5 columns in this order. Do not add or rename columns.**

```markdown
## Components

| ID | Name | Type | Responsibility | Status |
|----|------|------|----------------|--------|
| c3-101 | Request Handler | Foundation | HTTP routing | |
| c3-102 | Auth Service | Business | Token validation | Documented |
| c3-103 | Logger | Foundation | Structured logging | Skip: stdlib wrapper |
```

**MANDATORY COLUMNS (in order):**
1. `ID` — Component ID (c3-NNN format)
2. `Name` — Component name
3. `Type` — Either `Foundation` or `Business`
4. `Responsibility` — What it does (brief)
5. `Status` — Empty, `Documented`, or `Skip: {reason}`

**DO NOT USE:** Location, Path, Dependencies, or any other columns

### Status Values

| Status | Meaning | Exit Strategy |
|--------|---------|---------------|
| *(empty)* | Not yet documented, conventions maturing | Document when consumer rules emerge |
| `Documented` | Full component doc exists | - |
| `Skip: {reason}` | Will never need a doc | - |
| `Combined with c3-XXX` | Documented with another component | Split when complexity warrants |

### Component Types

| Type | Purpose |
|------|---------|
| **Foundation** | Cross-cutting (HTTP framework, logger, config) |
| **Business** | Domain logic (auth service, order processor) |

---

## Technology Stack Format

Document tech choices as a table. No patterns - the model knows frameworks.

```markdown
## Technology Stack

| Layer | Technology | Purpose |
|-------|------------|---------|
| Runtime | Node.js 20 | JavaScript runtime |
| Framework | Hono | HTTP server |
| Database | PostgreSQL | Primary data store |
```

---

## Diagrams

Use **Mermaid only**. See `references/diagram-patterns.md` for syntax.

| Layer | Required Diagram |
|-------|-----------------|
| Context | Container interactions (1 diagram) |
| Container | Internal structure (1 diagram) |

---

## ALTER/ADAPT Quick Reference

| Change Type | Layer | Update |
|-------------|-------|--------|
| New container | Context | Add to inventory, create container doc |
| Container relationship | Context | Update interactions diagram |
| New component | Container | Add to inventory (doc when conventions mature) |
| Component relationship | Container | Update internal structure |
| Protocol change | Context | Update all affected containers |
| Tech stack change | Container | Update tech table |

---

## Related

- `v3-structure.md` - Full structure reference with frontmatter
- `container-patterns.md` - Component inventory patterns
- `diagram-patterns.md` - Mermaid diagram syntax
- `adopt-workflow.md` - Full adopt workflow with overlay presentation
- `discovery-engine.md` - Subagent specifications for discovery
