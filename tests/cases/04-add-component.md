# Test: Add New Component

## Setup

**Fixture:** `fixtures/04-add-component/`

Pre-existing state (STUB + RELATED COMPONENTS):
- `.c3/README.md` - Context with one container
- `.c3/c3-1-api/README.md` - Container with:
  - c3-101 Cache Service (status: **STUB**)
  - c3-102 Routes (status: NOT DOCUMENTED)
  - c3-103 Rate Limiter (status: NOT DOCUMENTED) - shares Redis with c3-101
  - Component Dependencies section noting shared Redis pool
  - Notes about "invalidation strategy not finalized"
- `.c3/c3-1-api/c3-101-cache-service.md` - **STUB file exists** with:
  - Basic frontmatter (status: stub)
  - TODO list of what needs documenting

## Query

```
Document the cache service component (c3-101).
It uses Redis with TTL-based eviction, handles cache invalidation on updates,
and provides a typed wrapper for the rest of the application.
```

## Expect

### Pre-Checks
- [ ] Found existing stub file for c3-101
- [ ] Read parent container for context
- [ ] Noticed c3-103 Rate Limiter shares Redis pool
- [ ] Noticed "invalidation strategy not finalized" note

---

## Component Level Criteria

### PASS: Should Include
| Element | Check |
|---------|-------|
| Contract reference | "As defined in Container, this component..." |
| Interface diagram | IN → Processing → OUT boundary |
| Hand-offs table | What exchanges with c3-102, relationship with c3-103 |
| Conventions table | TTL policies, key patterns |
| Edge cases table | Cache miss, connection failure, eviction |
| Error scenarios | Redis down, timeout, serialization errors |
| Processing flow | Get/Set/Invalidate logic steps |
| Dependencies | Redis client (ioredis from container tech stack) |

### PASS: Addressed Related Context
| Element | Check |
|---------|-------|
| Shared Redis pool | Documented connection sharing with c3-103 |
| Invalidation strategy | Addressed the "not finalized" note |
| Container dependencies | Reflected what container says about relationships |

### FAIL: Should NOT Include
| Element | Failure Reason |
|---------|----------------|
| WHAT cache service does | Already in Container inventory |
| Component relationships | Already in Container (just reference) |
| Rate limiter details | That's c3-103's documentation |
| Redis setup/config | That's operational, not architectural |

---

## Stub Handling Criteria

### PASS: Completed Stub Properly
- [ ] Updated status from "stub" to something meaningful
- [ ] Addressed all TODOs from stub
- [ ] Maintained valid frontmatter structure

### FAIL: Ignored Stub
| Element | Failure Reason |
|---------|----------------|
| Created new file ignoring stub | Stub exists, should complete it |
| Left status as "stub" | Should update status |
| Didn't address TODO items | Stub gave specific guidance |

---

## NO CODE Rule Criteria

### PASS: No Code Present
- [ ] No TypeScript/JavaScript code blocks
- [ ] No JSON/YAML examples
- [ ] No `interface`, `type`, `function` keywords
- [ ] Only Mermaid diagrams in code blocks
- [ ] Tables used for data structures

### FAIL: Code Present
| Element | Failure Reason |
|---------|----------------|
| `interface CacheOptions {...}` | Use table: Field \| Type \| Purpose |
| `{ "ttl": 3600 }` config example | Use table: Setting \| Default \| Purpose |
| `await redis.get(key)` | Describe pattern, not syntax |
| Redis commands | Describe pattern, not syntax |

---

## Diagram Criteria

### PASS: Correct Diagrams
| Type | Check |
|------|-------|
| Interface diagram | Shows IN/OUT boundary (REQUIRED) |
| Flowchart | Processing steps, decision logic |
| Sequence | Calls to Redis if helpful |

### FAIL: Wrong Diagrams
| Type | Failure Reason |
|------|----------------|
| Container overview | Push to Container level |
| Component relationships | Already in Container |
| Class diagram with methods | Too implementation-focused |
