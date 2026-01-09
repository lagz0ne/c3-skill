---
id: c3-101
c3-version: 3
title: Router
type: component
category: foundation
parent: c3-1
summary: TanStack Router configuration with file-based routing and SSR support
---

# Router

Provides the application routing infrastructure using TanStack Router with file-based route generation, scroll restoration, and server-side rendering configuration.

## Contract

| Provides | Expects |
|----------|---------|
| Route matching and navigation | routeTree.gen.ts (auto-generated) |
| History management | Browser history API |
| Scroll restoration | DOM scroll positions |
| Route preloading | staleTime=0 for fresh data |

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Route not found | Renders 404 fallback |
| SSR mismatch | Hydration warnings logged |
| Navigation during loading | Aborts previous navigation |
| Deep linking | Parses URL params correctly |

## References

- `apps/start/src/router.tsx` - Router configuration
- `apps/start/src/routes/` - All route definitions

## Testing

| Scenario | Verifies |
|----------|----------|
| Route matching | URL /prs renders PaymentRequestsScreen |
| Param extraction | /files/:hash passes hash to component |
| Authenticated routes | /_authed/* redirects to /login when unauthenticated |
| Scroll restoration | Navigate back, scroll position restored |
