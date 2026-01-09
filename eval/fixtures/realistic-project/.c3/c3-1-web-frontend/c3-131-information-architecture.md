---
id: c3-131
c3-version: 3
title: Information Architecture
type: component
category: documentation
parent: c3-1
summary: Screen inventory with regions at medium abstraction level for UI development
---

# Information Architecture

Documents all screens, panels, dialogs, and regions in the application. Each UI element has a unique ID (SCR-*, PNL-*, DLG-*, GBL-*) used for consistent reference across flows and testing.

## Uses

| Category | Component | For |
|----------|-----------|-----|
| Foundation | c3-101 Router | Route structure definition |
| Foundation | c3-102 AuthLayout | GBL-SIDEBAR wrapper |
| Feature | c3-121 Invoice Screen | SCR-INV implementation |
| Feature | c3-122 Payment Requests Screen | SCR-PR, SCR-APR implementation |
| Feature | c3-124 Admin Screen | SCR-ADMIN implementation |

## Conventions

| Rule | Why |
|------|-----|
| Prefix SCR- for screens | Top-level route views |
| Prefix PNL- for panels | Detail panels within screens |
| Prefix DLG- for dialogs | Drawer/modal overlays |
| Prefix GBL- for global | Shared across screens |
| Suffix -HDR, -FTR, -LIST | Standard region names |

## Screens

| ID | Route | Layout | Permission |
|----|-------|--------|------------|
| SCR-LOGIN | /login | Centered card | Unauthenticated |
| SCR-INV | /invoices | Master-Detail | `invoices` |
| SCR-PR | /prs | Master-Detail | `prs` |
| SCR-APR | /approvals | Master-Detail | `approvals` |
| SCR-PAY | /payments | Master-Detail | `payments` |
| SCR-ADMIN | /admin | Full-width | `isOwner` |

## Global Components

| ID | Location | Purpose |
|----|----------|---------|
| GBL-SIDEBAR | Left side (desktop) | Navigation, theme, user menu |
| GBL-MOBILE-HEADER | Top (mobile) | Back button, hamburger menu |

## Testing

| Scenario | Verifies |
|----------|----------|
| Region ID coverage | All UI elements have documented IDs |
| Route-to-screen mapping | Routes resolve to correct screens |
| Permission matrix | Screens enforce documented permissions |

## References

- `apps/start/src/routes/` - Route definitions
- `apps/start/src/screens/` - Screen implementations
- `apps/start/src/components/` - Shared components
- [c3-132 User Flows](./c3-132-user-flows.md) - Flow documentation using these IDs
- [c3-133 UI Patterns](./c3-133-ui-patterns.md) - Pattern implementations
