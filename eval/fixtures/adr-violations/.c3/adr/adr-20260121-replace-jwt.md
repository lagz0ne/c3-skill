---
id: adr-20260121-replace-jwt
title: Replace JWT with session-based auth
status: proposed
date: 2026-01-21
affects: [c3-102]
approved-files:
  - src/middleware/auth.ts
---

# Replace JWT with session-based auth

## Status

**Proposed** - 2026-01-21

## Problem

JWT tokens are stateless and cannot be invalidated.

## Decision

Replace JWT-based authentication with session cookies for better security.

## Rationale

Sessions can be invalidated server-side.

| Considered | Rejected Because |
|------------|------------------|
| JWT blocklist | Complexity |

## Affected Layers

| Layer | Document | Change |
|-------|----------|--------|
| Component | c3-102 | Replace JWT with sessions |

## Approved Files

```yaml
approved-files:
  - src/middleware/auth.ts
```

## Verification

- [ ] Session auth works
- [ ] JWT code removed
