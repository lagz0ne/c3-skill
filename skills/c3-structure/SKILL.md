---
name: c3-structure
description: Use when understanding or changing system structure - containers, relationships, component inventory. Answers "WHERE in the system?" and "WHAT is connected?"
---

# C3 Structure Design

## Critical Gate

> **STOP** - Before ANY structure work:
> ```bash
> cat .c3/README.md 2>/dev/null || echo "NO_CONTEXT"
> cat .c3/c3-{N}-*/README.md 2>/dev/null || echo "NO_CONTAINER"
> ```

- If "NO_CONTEXT" and working at Context level → Create Context first
- If "NO_CONTEXT" and working at Container level → Context must exist first
- If container doc exists → Read before proposing changes

**DO NOT read ADRs** unless asked - they add context load without helping current structure decisions.

---

## The Principle

> **Upper layer defines WHAT. Lower layer implements HOW.**

See `references/core-principle.md` for full model.

| Layer | Defines | Implemented By |
|-------|---------|----------------|
| Context | WHAT containers exist | Container docs |
| Container | WHAT components exist | Component docs |

---

## ALTER/ADAPT Decision

| Change Type | Layer | Update |
|-------------|-------|--------|
| New container | Context | Add to inventory, create container doc |
| Container relationship | Context | Update interactions diagram |
| New component | Container | Add to inventory (doc when conventions mature) |
| Component relationship | Container | Update internal structure |
| Protocol change | Context | Update all affected containers |
| Tech stack change | Container | Update tech table |

---

## Context Level (c3-0)

**File:** `.c3/README.md`

### Purpose

Document WHAT containers exist and HOW they relate. Bird's-eye view.

### Litmus Test

> "Is this about WHY containers exist and HOW they connect?"

Yes → Context. No → Push to Container.

### Required Sections

1. **Overview** - System purpose, boundary
2. **Containers** - Inventory table (always complete)
3. **Interactions** - Mermaid diagram showing relationships
4. **External Actors** - Who/what interacts from outside

### Container Inventory

```markdown
## Containers

| ID | Name | Responsibility |
|----|------|----------------|
| c3-1 | API Backend | Request handling, business logic |
| c3-2 | Frontend | User interface, client state |
| c3-3 | Worker | Async job processing |
```

**Rule:** List ALL containers. This is the source of truth for what exists.

---

## Container Level (c3-N)

**File:** `.c3/c3-{N}-{slug}/README.md`

### Purpose

Document WHAT components exist and HOW they relate within this container.

### Litmus Test

> "Is this about WHAT components do and HOW they connect inside this container?"

Yes → Container. About container relationships → Push to Context. About how components work internally → Push to Component.

### Required Sections

1. **Inherited From Context** - What this container is responsible for (from parent)
2. **Overview** - Container purpose
3. **Technology Stack** - Table only, no patterns
4. **Components** - Inventory table (always complete)
5. **Internal Structure** - Mermaid diagram
6. **Key Flows** - 1-2 critical paths (brief)

---

## Inventory-First Model

**CRITICAL:** The components table is the source of truth.

### Rules

1. **Inventory is always complete** - List ALL components, even without detailed docs
2. **Docs appear when conventions mature** - Component doc = conventions exist for consumers
3. **No stubs** - Either a full doc exists or it doesn't
4. **No doc = no consumer conventions** - Just "use it" (e.g., standard logger)

```
WRONG:
.c3/c3-1-backend/
├── README.md
├── c3-101-handler.md     ← NO! Don't create at structure time
├── c3-102-auth.md        ← NO! Conventions not mature yet
└── c3-103-db.md          ← NO! Just "we use postgres"

RIGHT:
.c3/c3-1-backend/
└── README.md             ← Contains component INVENTORY table
    ## Components
    | ID | Name | Type | Responsibility | Status |
    | c3-101 | Handler | Foundation | HTTP routing | |
    | c3-102 | Auth | Business | Authentication | Documented |
    | c3-103 | DB Client | Foundation | Database access | Skip: pg library |
```

### Components Inventory

```markdown
## Components

| ID | Name | Type | Responsibility | Status |
|----|------|------|----------------|--------|
| c3-101 | Request Handler | Foundation | HTTP routing, middleware | |
| c3-102 | Auth Service | Business | Token validation, sessions | Documented |
| c3-103 | Logger | Foundation | Structured logging | Skip: stdlib wrapper |
| c3-104 | User Service | Business | User CRUD | Combined with c3-102 |
```

### Status Column

Status explains WHY a component doc doesn't exist (if applicable):

| Status | Meaning | Exit Strategy |
|--------|---------|---------------|
| *(empty)* | Not yet documented, conventions maturing | Document when consumer rules emerge |
| `Documented` | Full component doc exists | - |
| `Skip: {reason}` | Will never need a doc | - |
| `Combined with c3-XXX` | Documented together with another component | Split when complexity warrants separate doc |

**Examples:**
- `Skip: stdlib wrapper` - Standard library usage, no custom conventions
- `Skip: framework default` - Using framework as-is
- `Combined with c3-102` - Auth and sessions documented together
- *(empty)* - Conventions not yet mature, may document later

### When to Create Component Doc

| Has conventions for consumers? | Action |
|-------------------------------|--------|
| Yes - rules consumers must follow | Create component doc, mark `Documented` |
| No - just "we use X library" | Mark `Skip: {reason}` |
| Part of another component | Mark `Combined with c3-XXX` |
| Not yet decided | Leave Status empty |

### Foundation vs Business

| Type | Purpose |
|------|---------|
| **Foundation** | Cross-cutting, provides to business (HTTP framework, logger, config) |
| **Business** | Domain logic (auth service, order processor) |

---

## Diagrams

Use **Mermaid only**. See `references/diagram-patterns.md`.

| Layer | Required Diagram |
|-------|-----------------|
| Context | Container interactions (1 diagram) |
| Container | Internal structure (1 diagram) |

---

## Technology Stack

Document tech choices as a table. No patterns - the model knows frameworks.

```markdown
## Technology Stack

| Layer | Technology | Purpose |
|-------|------------|---------|
| Runtime | Node.js 20 | JavaScript runtime |
| Framework | Hono | HTTP server |
| Database | PostgreSQL | Primary data store |
| Cache | Redis | Session and response caching |
```

---

## Verification

Before completing:

- [ ] Layer integrity: parent exists before child
- [ ] Inventory complete: all containers/components listed
- [ ] Diagrams present: Mermaid, not ASCII
- [ ] No implementation details: HOW belongs in Component layer

---

## Escalation

| Situation | Action |
|-----------|--------|
| Change affects multiple containers | Work at Context level |
| Change is internal to one container | Work at Container level |
| Change is how a component works | Use c3-implementation skill |
