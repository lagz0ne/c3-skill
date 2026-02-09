# Component Categories

The three-way split for C3 documentation. **Strict separation required.**

## The Test

| Question | If Yes | Category |
|----------|--------|----------|
| Can you name a concrete code file (e.g., `logger.ts`)? | Yes | **FOUNDATION** |
| Is it only rules/conventions about how/when to use code? | Yes | **REF** |
| Is it domain logic that composes Foundation + follows Refs? | Yes | **FEATURE** |

## Foundation

Actual code that other components import. **You can point to a real file.**

Examples: `logger.ts`, `db-client.ts`, `auth-provider.ts`, `event-bus.ts`

**`## Code References`** marks this component as implemented — it has a code counterpart.

## Feature

Business logic composition. Glues Foundation + Refs together.

Examples: `checkout-flow.ts`, `user-registration.ts`, `order-processor.ts`

**`## Code References`** marks this component as implemented — it has a code counterpart.

## Ref

Guidance and conventions. Describes how/when/why to do something.

May include **code examples as golden references** — canonical snippets that implementation code should be reviewed against. These are prescriptive patterns, not an implementation counterpart.

Examples: `ref-logging.md` (when to log, with canonical log format snippet), `ref-error-handling.md` (error format with example response shape)

**No `## Code References`** — refs have no code counterpart. They ARE the deliverable.

## The Canonical Split

| What | Category | Has Code References? |
|------|----------|---------------------|
| `logger.ts` implementation | Foundation | YES |
| "When to log at DEBUG vs INFO" (may include golden code examples) | Ref | NO |
| "Checkout flow using logger + error conventions" | Feature | YES |

## Hard Rules (LLM MUST FOLLOW)

1. **If you cannot name a concrete file, you cannot create a component doc.**
2. **`## Code References` = implemented.** Components have it when code exists. Provisioned components don't (yet). Refs never do.
3. **Refs MAY contain code examples** as golden references — canonical patterns to review against.
4. **Never create component files for conventions** — use refs.

## Audit Trigger

If an implemented component doc has no `## Code References`, it is either misclassified (move to `refs/`) or still provisioned (add `status: provisioned` to frontmatter).
