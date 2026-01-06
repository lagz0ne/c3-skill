# Component Types

Three categories organize components by role and impact.

## Quick Selection

```mermaid
graph TD
    Q1{Reusable within container?}
    Q1 -->|Yes| Q2{Establishes conventions?}
    Q1 -->|No| Feature
    Q2 -->|Yes| Auxiliary
    Q2 -->|No| Foundation
```

| Ask | If Yes → | If No → |
|-----|----------|---------|
| Is this reusable by other components? | Foundation or Auxiliary | Feature |
| Does this establish HOW to use something? | Auxiliary | Foundation |
| Does this deliver user-facing value directly? | Feature | Foundation or Auxiliary |

## Foundation

**What:** Core primitives that others build on. High impact when changed.

**Examples:**
- Layout, Button, Input (UI primitives)
- Router, AuthProvider (infrastructure)
- DataStore, EventBus (shared state)
- Logger, ConfigLoader (utilities)

**Characteristics:**
- Changes ripple to many dependents
- Stable API surface
- Minimal external dependencies
- Well-tested edge cases

**Template:** `templates/component-foundation.md`

## Auxiliary

**What:** Conventions for using external tools. "How we use X here."

**Examples:**
- TailwindPatterns (CSS conventions)
- TypeConventions (prefer type over interface)
- APIClientPatterns (error handling, retries)
- TestingPatterns (mocking approach)

**Characteristics:**
- Documents decisions, not code
- Guides consistent usage
- References external docs
- Applied across Feature components

**Template:** `templates/component-auxiliary.md`

## Feature

**What:** Domain-specific. Delivers user value. Uses Foundation + Auxiliary.

**Examples:**
- ProductCard, CheckoutFlow (e-commerce)
- UserProfile, SettingsPage (user management)
- OrderHistory, InvoiceGenerator (transactions)
- Dashboard, ReportViewer (analytics)

**Characteristics:**
- Not reusable outside context
- Combines Foundation + Auxiliary
- User-facing behavior
- Domain-specific logic

**Template:** `templates/component-feature.md`

## Dependency Direction

```
Feature
   │
   ├──→ Foundation
   │
   └──→ Auxiliary (conventions)
```

**Rules:**
- Feature depends on Foundation (allowed)
- Feature follows Auxiliary conventions (allowed)
- Foundation never depends on Feature (violation)
- Auxiliary never depends on Feature (violation)

## Common Mistakes

| Mistake | Why Wrong | Fix |
|---------|-----------|-----|
| Feature marked Foundation | Not reusable, domain-specific | Reclassify as Feature |
| Helper functions as Auxiliary | Auxiliary is conventions, not code | Move to Foundation |
| UI component as Auxiliary | Auxiliary has no visual output | Move to Foundation |
| Config as Feature | Config is infrastructure | Move to Foundation |

## Migration Guide

When reclassifying components:

1. Check dependents (who uses this?)
2. Check dependencies (what does this use?)
3. Apply selection flowchart
4. Update component's `category` in frontmatter
5. Move file if needed (path reflects category)
6. Update container inventory
