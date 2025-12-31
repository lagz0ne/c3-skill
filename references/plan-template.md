# Plan Template

Implementation plans for ADRs. Lives next to the ADR file.

## File Naming

```
.c3/adr/adr-YYYYMMDD-{slug}.plan.md
```

## Template

```markdown
---
adr: adr-YYYYMMDD-{slug}
status: pending
---

# Plan: [ADR Title]

> Implements [adr-YYYYMMDD-{slug}](./adr-YYYYMMDD-{slug}.md)

## Changes

### c3-0 (Context)

**File:** `.c3/README.md`

| Section | Action | Change |
|---------|--------|--------|
| Overview diagram | Edit | Add `E2[Redis]` node, add `c3-1 --> E2` edge |
| External Systems | Add row | `E2 | Redis | cache | Response caching` |
| Linkages | Add row | `c3-1 | E2 | TCP | Cache hot data` |

### c3-1 (Container)

**File:** `.c3/c3-1-{slug}/README.md`

| Section | Action | Change |
|---------|--------|--------|
| Overview diagram | Edit | Add `c3-103[Cache Client]` node |
| Auxiliary table | Add row | `c3-103 | Cache Client | library-wrapper | | Redis abstraction` |
| Fulfillment | Add row | `c3-1 → E2 | c3-103 | Redis client, connection pooling` |
| Linkages | Add row | `c3-105 | c3-103 | cache-read | Reduce DB load for hot paths` |

### c3-103 (Component) - NEW

**File:** `.c3/c3-1-{slug}/c3-103-cache-client.md`

Action: Create from component template

| Section | Content |
|---------|---------|
| Summary | Redis client wrapper for consistent caching |
| Interface IN | Cache key, TTL, value |
| Interface OUT | Cached value or miss |
| Technology | ioredis, connection pooling |
| Conventions | All cache keys prefixed with service name |

## Order

```
1. Update c3-0 (Context) - adds external system
2. Update c3-1 (Container) - adds component to inventory
3. Create c3-103 (Component) - new component doc
```

## Verification

| From ADR | How to Verify |
|----------|---------------|
| Redis in External Systems | `grep "Redis" .c3/README.md` |
| Cache Client in inventory | `grep "c3-103" .c3/c3-1-*/README.md` |
| Component doc exists | `ls .c3/c3-1-*/c3-103-*.md` |
| Diagrams updated | Visual check, IDs match tables |
```

## Guidelines

- Reference ADR in frontmatter
- One section per affected document
- Table format: Section | Action | Change
- Explicit order of operations
- Verification maps to ADR verification checklist
- No fluff, just actionable steps
