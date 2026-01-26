# Content Separation Reference

Canonical definition for separating Foundation (code), Feature (composition), and Ref (guidance).

## The Three-Way Test

| Question | Answer | Category |
|----------|--------|----------|
| Does it have **actual code** others import/depend on? | Yes | **Foundation** (component) |
| Is it **how/when/why to do something** with no code? | Yes | **Ref** (pattern) |
| Does it **compose foundation + follow refs** for business? | Yes | **Feature** (component) |

## STRICT RULE (LLM MUST FOLLOW)

```
┌────────────────────────────────────────────────────────────┐
│  Foundation = HAS CODE (logger.ts, db-client.ts)           │
│  Ref = NO CODE (when to log, how to handle errors)         │
│  Feature = COMPOSES CODE (uses foundation, cites refs)     │
└────────────────────────────────────────────────────────────┘
```

- FOUNDATION and FEATURE docs **MUST** include a `## Code References` section that points to real files.
- REF docs **MUST NOT** include a `## Code References` section.

**Hard test:** If the doc can't list at least one real code file in `## Code References`, it cannot be a component.

If you are describing "when/how/why" with no concrete file → it is a REF.
If you can point to a concrete file → it is a FOUNDATION or FEATURE component.

## What Belongs Where

### Foundation Components (Actual Implementations)

| Signal | Example | Why Foundation |
|--------|---------|----------------|
| Concrete implementation | `logger.ts`, `db-client.ts` | Others import this code |
| Framework integration | `hono-routes.ts`, `auth-provider.ts` | System depends on it |
| Shared utilities | `event-bus.ts`, `config-loader.ts` | Reused across features |

### Feature Components (Business Composition)

| Signal | Example | Why Feature |
|--------|---------|-------------|
| Domain-specific rules | "Users charged when subscription expires" | Business logic |
| Entity behavior | "Orders transition to SHIPPED" | Domain state |
| Business calculations | "10% discount for orders > $100" | Non-reusable |
| User-facing flows | "Checkout flow", "Registration" | Composes foundation |

### Refs (Patterns/Conventions — NO CODE)

| Signal | Example | Why Ref |
|--------|---------|---------|
| "When to X" | "When to log at DEBUG vs INFO" | Guidance, no code |
| "How to X" | "How to structure error responses" | Convention |
| "Our convention is..." | "All APIs return RFC 7807 errors" | Pattern |
| Style/approach | "UI composition principles" | No implementation |
| Cross-cutting patterns | "All requests include X-Request-ID" | Applied everywhere |

**Logger example (the canonical split):**
- "When to log at DEBUG vs INFO" → **REF** (`ref-logging.md`)
- "How log messages must be structured" → **REF** (`ref-logging.md`)
- "`logger.ts` implementation" → **FOUNDATION** (not a ref, has Code References)

## What Refs Capture

Refs answer **"how WE use X HERE"** - specific decisions and conventions for THIS codebase:

| Question | What It Captures |
|----------|------------------|
| **When** do we use this? | Context, triggers, conditions |
| **Why** this over alternatives? | Decision rationale |
| **Where** is the entry point? | How to invoke, where to start |
| **What** conventions apply? | Constraints, patterns we follow |

## What Refs Are NOT

- API reference (that's the library's docs)
- Step-by-step tutorials
- Configuration deep-dives
- Generic technology documentation (link to official docs instead)

## Scoping Principle

If a ref topic is too broad to answer "what patterns?" concisely → split it.

- `ref-react` too big → `ref-react-components`, `ref-react-data`, `ref-react-testing`
- Or scope by concern → `ref-async-data` (covers react-query, SWR, whatever)

## Related

- `audit-checks.md` - Phase 9 validates content separation
- `c3-content-classifier.md` - Agent for LLM-based content classification
- `templates/ref.md` - Ref template with When/Why/Conventions structure
