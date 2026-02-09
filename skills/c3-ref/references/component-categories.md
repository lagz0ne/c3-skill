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

**Must have** `## Code References` section pointing to actual files.

## Feature

Business logic composition. Glues Foundation + Refs together.

Examples: `checkout-flow.ts`, `user-registration.ts`, `order-processor.ts`

**Must have** `## Code References` section pointing to actual files.

## Ref

Guidance only, no code. Describes how/when/why to do something.

Examples: `ref-logging.md` (when to log), `ref-error-handling.md` (error format)

**Must NOT have** `## Code References` section.

## The Canonical Split

| What | Category | Has Code References? |
|------|----------|---------------------|
| `logger.ts` implementation | Foundation | YES |
| "When to log at DEBUG vs INFO" | Ref | NO |
| "Checkout flow using logger + error conventions" | Feature | YES |

## Hard Rules (LLM MUST FOLLOW)

1. **If you cannot name a concrete file, you cannot create a component doc.**
2. **Components MUST have `## Code References`. Refs MUST NOT.**
3. **Never create component files for conventions** — use refs.

## Audit Trigger

If a component doc has no `## Code References`, it is misclassified. Move to `refs/` and remove the component ID.
