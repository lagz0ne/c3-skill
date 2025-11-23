# Derivation Guardrails

Rules for maintaining consistent C3 documentation hierarchy and relationships.

## Reading Order

Always read top-down: **Context → Container → Component**

| Start At | Read Next |
|----------|-----------|
| Unknown entry point | Context (`.c3/README.md`) first |
| Context | Containers listed in Containers section |
| Container | Components listed in Components section |
| Component | Terminal - no further derivation |

## Reference Direction

**References only flow DOWN** - higher layer links to lower layer implementations.

```
Context ──────┬──> Container 1 ──┬──> Component 101
              │                  └──> Component 102
              └──> Container 2 ──┬──> Component 201
                                 └──> Component 202
```

### What This Means

| Level | Links TO | Links FROM |
|-------|----------|------------|
| Context | Container docs, Container#sections | None (root) |
| Container | Component docs, Component#sections | Context |
| Component | None (leaf) | Container |

### Why Downward Only

- **Single source of truth** - No duplicate relationship definitions
- **Natural reading flow** - Reader follows derivation path
- **Maintenance in one place** - Update relationship at parent only

## Infrastructure Containers

Infrastructure containers (databases, caches, queues) are **leaf nodes**:

- No component level beneath them
- Document features provided, not internal structure
- Code containers consume their features

```
Context
  └── Backend Container (code) ───> Component: DB Pool
  └── Postgres Container (infra) ───> [no components]
         ^
         └── Features consumed by DB Pool
```

## Abstraction Check

Before documenting, verify the level is correct:

| Signal | Correct Level |
|--------|---------------|
| Affects system boundary or actors | Context |
| Affects multiple containers | Context |
| Defines protocol between containers | Context |
| Specific to one container's tech stack | Container |
| Defines container's API surface | Container |
| Implementation detail within container | Component |
| Configuration and environment values | Component |

### Litmus Tests

**For Context:**
> "Would changing this require coordinating multiple containers or external parties?"

**For Container:**
> "Is this about WHAT this container does and WITH WHAT, not HOW internally?"

**For Component:**
> "Could a developer implement this from the documentation?"

## Naming Rules

See [v3-structure.md](v3-structure.md) for complete ID/path patterns.

Key points:
- Components embed parent container digit: `c3-101` is in container 1
- Anchors include level prefix: `{#c3-1-middleware}`, `{#c3-101-config}`
- All lowercase, no exceptions

## Common Violations

| Violation | Fix |
|-----------|-----|
| Component linking UP to container | Remove - reader came from container |
| Container linking UP to context | Remove - reader came from context |
| Context with implementation details | Push down to Container or Component |
| Component without parent digit in ID | Rename: `c3-01` → `c3-101` |
