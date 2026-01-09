---
id: c3-202
c3-version: 3
title: Scope & DI
type: component
category: foundation
parent: c3-2
summary: @pumped-fn/lite scope management and dependency injection patterns
---

# Scope & DI

Provides dependency injection infrastructure using @pumped-fn/lite's createScope, atoms, services, flows, and tags for configuration and context passing.

## Contract

| Provides | Expects |
|----------|---------|
| createScope() for root scope | Extensions and initial tags |
| scope.resolve() for dependency lookup | Atom/service definitions |
| scope.createContext() for execution | Per-request context isolation |
| tags for configuration | Tag definitions with values |
| ctx.data.setTag/seekTag | Context-scoped data |

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Missing dependency | Runtime error with clear message |
| Circular dependency | Detected and throws |
| Scope disposal | All atoms cleaned up |
| Context disposal | Context-specific cleanup runs |

## Testing

| Scenario | Verifies |
|----------|----------|
| Atom resolution | scope.resolve(atom) returns value |
| Service factory | Service receives deps correctly |
| Tag access | seekTag returns set value |
| Cleanup registration | ctx.cleanup() called on dispose |

## References

- `apps/start/src/server/tags.ts` - DI tag definitions
- `apps/start/src/server/resources/` - Resource atoms
