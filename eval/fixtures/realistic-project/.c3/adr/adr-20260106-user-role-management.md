---
id: adr-20260106-user-role-management
title: User and Role Management System
status: implemented
plan: adr-20260106-user-role-management.plan.md
date: 2026-01-06
affects: [c3-1, c3-2, c3-203]
---

# User and Role Management System

## Status

**Implemented** - 2026-01-06

## Problem

User and role management operations exist in the database layer (RBAC queries, schema tables) but are not exposed through the UI or backend flows. There is no way for administrators to:
- View/create/edit/delete users
- Assign permissions to users
- Manage the "owner" role (first/last manager with special permissions)

This blocks basic administrative operations and ownership transfer.

## Decision

Implement a user management system with:

1. **Single Admin Screen** (`/admin`) with tabs for user management
2. **Owner Role** - Role-based implementation with special semantics:
   - Cannot delete the last owner
   - Transferable to other users
   - Grants access to user management
3. **User Operations Only** - Owners manage users; roles are system-defined
4. **New Backend Flows** - Expose RBAC queries through dedicated flows

## Rationale

| Considered | Rejected Because |
|------------|------------------|
| isOwner column on users | Less flexible than role-based, harder to extend |
| Owner as permission string | Doesn't support "last owner" protection semantics |
| Separate screens per feature | Over-engineering for this use case |
| Full role management UI | Scope creep - roles are system-defined, only user assignment needed |

## Affected Layers

| Layer | Document | Change |
|-------|----------|--------|
| Container | c3-1 | Add admin screen component reference |
| Container | c3-2 | Add user management flows reference |
| Component | c3-124 (NEW) | Admin Screen - user management tabs |
| Component | c3-224 (NEW) | User Management Flows - user CRUD, role assignment |
| Component | c3-203 | Update schema entities to include RBAC tables |

## Implementation Scope

### Backend (c3-224)

| Flow | Purpose |
|------|---------|
| listUsers | List all users with their roles/permissions |
| createUser | Create new user (owner-only) |
| updateUser | Update user permissions/status (owner-only) |
| deleteUser | Remove user (owner-only, cannot delete last owner) |
| assignRole | Assign role to user (owner-only) |
| removeRole | Remove role from user (owner-only) |
| transferOwnership | Transfer owner role to another user |

### Frontend (c3-124)

| Tab | Features |
|-----|----------|
| Users | List users, add/edit/remove users |
| Ownership | View owners, transfer ownership |

### Owner Role Semantics

- System-defined role named "owner"
- Automatically assigned to first user (seed/bootstrap)
- Multiple users can have owner role
- At least one owner must exist at all times
- Only owners can access `/admin` and perform user management

## Plan

See [adr-20260106-user-role-management.plan.md](./adr-20260106-user-role-management.plan.md) for implementation details.

## Verification

- [ ] At least one user has owner role after bootstrap
- [ ] Cannot delete last owner (returns error)
- [ ] Non-owners get 403 on user management flows
- [ ] Admin screen hidden from non-owners
- [ ] Ownership transfer creates security event
- [ ] C3 docs updated and audit passes
