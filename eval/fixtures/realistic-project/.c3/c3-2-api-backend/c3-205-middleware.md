---
id: c3-205
c3-version: 3
title: Middleware
type: component
category: foundation
parent: c3-2
summary: TanStack middleware for execution context and user authentication from cookies
---

# Middleware

Provides TanStack Start middleware for creating execution contexts per request and extracting authenticated user from cookies.

## Contract

| Provides | Expects |
|----------|---------|
| executionContextMiddleware | scope in context |
| getCurrentUserMiddleware | Cookies with 'user' email |
| execContext in context | For flow execution |
| currentUserTag set | When user cookie valid |

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| No cookie header | Proceeds without user |
| Invalid user cookie | Proceeds without user |
| User not in database | Proceeds without user |
| execContext cleanup | Runs in finally block |

## Testing

| Scenario | Verifies |
|----------|----------|
| Context creation | execContext available to handler |
| User extraction | Valid cookie sets currentUserTag |
| Missing user | No crash, user undefined |
| Context cleanup | close() called after handler |

## References

- `apps/start/src/server/middlewares/` - Request middleware
- `apps/start/src/server/functions/middleware.ts` - Server function middleware
