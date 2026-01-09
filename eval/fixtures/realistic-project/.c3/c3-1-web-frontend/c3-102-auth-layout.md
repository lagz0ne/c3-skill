---
id: c3-102
c3-version: 3
title: AuthLayout
type: component
category: foundation
parent: c3-1
summary: Authenticated layout wrapper with navigation, user menu, and state initialization
---

# AuthLayout

Provides the authenticated shell for all protected routes including sidebar navigation, user avatar menu, theme toggle, and initialization of state atoms with SSR-prefetched data via ScopeProvider.

## Contract

| Provides | Expects |
|----------|---------|
| Navigation sidebar | Authenticated user from loader |
| User menu with logout | User email and avatar |
| Theme toggle (light/dark) | localStorage for persistence |
| ScopeProvider with presets | SSR data for invoices, prs, payments |
| Outlet for child routes | Route children to render |

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Unauthenticated access | Redirect to /login |
| SSR data missing | Shows loading screen |
| Scope creation failure | ErrorBoundary catches |
| Mobile viewport | Collapsible sidebar |

## Testing

| Scenario | Verifies |
|----------|----------|
| Auth redirect | Visit /_authed without cookie, redirected to /login |
| Scope initialization | Loader data populates atoms |
| Navigation items | Sidebar shows Invoices, PRs, Payments, Approvals |
| User menu | Avatar click shows logout option |
| Theme persistence | Toggle theme, refresh, theme persists |

## References

- `apps/start/src/routes/_authed.tsx` - Layout wrapper
- `apps/start/src/routes/_authed/` - Protected routes
