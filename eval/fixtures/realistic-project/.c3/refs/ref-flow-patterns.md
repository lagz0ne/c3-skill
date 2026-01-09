---
id: ref-flow-patterns
title: Flow Patterns
---

# Flow Patterns

## Goal

Establish conventions for defining business logic flows using @pumped-fn/lite with Zod schema parsing, dependency injection, and consistent result types.

## Conventions

| Rule | Why |
|------|-----|
| Define flows with `flow({deps, parse, factory})` | Consistent structure for all business operations |
| Use Zod schema in `parse` | Input validation before execution |
| Namespace pattern for types | Group Input/Success/Failure/Result types together |
| Access current user via `ctx.data.seekTag(currentUserTag)` | Consistent auth context access |
| Return discriminated union results | Type-safe success/failure handling |
| Use `ctx.exec()` for service/query calls | Enables execution tracking and tracing |
| Emit sync events after mutations | Real-time updates to connected clients |
| Check executionId for sync acknowledgment | Prevents duplicate notifications |

## Testing

| Convention | How to Test |
|------------|-------------|
| Zod parsing | Pass invalid input, verify parse error |
| Missing user tag | Call without currentUserTag, verify USER_NOT_FOUND |
| Success result | Valid input and user, verify success: true |
| Failure result | Business rule violation, verify success: false with reason |
| Sync emission | Mutation flow, verify sync.emit called |

## References

- `apps/start/src/server/flows/` - All flow definitions
