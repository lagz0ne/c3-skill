# Content Separation Reference

Canonical definition for separating component content (domain logic) from ref content (usage patterns).

## The Separation Test

> **"Would this content change if we swapped the underlying technology?"**
> - **Yes** → Integration/usage pattern → belongs in **ref**
> - **No** → Business/domain logic → belongs in **component**

## What Belongs Where

### Components (Domain Logic)

| Signal | Example |
|--------|---------|
| Domain-specific rules | "Users are charged when subscription expires" |
| Entity behavior | "Orders transition to SHIPPED after fulfillment" |
| Business calculations | "Discount is 10% for orders over $100" |
| Feature-specific logic | "Dashboard shows last 30 days of activity" |
| User-facing behavior | "Notifications are sent when..." |

### Refs (Usage Patterns)

| Signal | Example |
|--------|---------|
| Technology conventions | "We use RFC 7807 for error responses" |
| Framework setup | "Redis is configured with 1h TTL" |
| Cross-cutting patterns | "All requests include X-Request-ID" |
| Library usage | "Prisma queries use soft delete pattern" |
| "We use X for..." | Technology usage pattern |
| "Our convention is..." | Cross-cutting convention |

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
