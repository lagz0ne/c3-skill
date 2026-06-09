---
id: c3-107
c3-version: 3
c3-seal: 372e8811ff91fa99972e23297b8fce41b7dca899a1b15533b465dc0c905178ab
title: Admin Screens
type: component
category: feature
parent: c3-1
goal: Admin management - users, teams, audit logs, notification logs, approval config, organization view
uses:
    - ref-admin-page-layout
    - ref-form-patterns
    - ref-org-tiles
    - ref-responsive-layout
    - ref-sft-behavioral-spec
    - ref-ui-patterns
    - ref-variant-system
---

# Admin Screens

## Goal

Admin management - users, teams, audit logs, notification logs, approval config, organization view

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | Own Admin Screens behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep Admin Screens decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for Admin Screens so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before Admin Screens behavior is changed. | ref-admin-page-layout |
| Inputs | Accept only the files, commands, data, or calls that belong to Admin Screens ownership. | ref-admin-page-layout |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | ref-admin-page-layout |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | ref-admin-page-layout |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks Admin Screens to deliver its documented responsibility. | ref-admin-page-layout |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | ref-admin-page-layout |
| Alternate paths | When a request falls outside Admin Screens ownership, hand it to the parent or sibling component. | ref-admin-page-layout |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | ref-admin-page-layout |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-admin-page-layout | ref | Governs Admin Screens behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| Admin Screens input | IN | Callers must provide context that matches the component goal and parent fit. | c3-1 boundary | c3x lookup plus targeted tests or review. |
| Admin Screens output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-1 boundary | c3x check and project test suite. |

## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| Contract drift | Goal, boundary, or derived material changes without matching component docs. | Compare Goal, Parent Fit, Contract, and Derived Materials. | Run c3x check and relevant project tests. |
| Governance drift | Cited references, rules, or parent responsibilities change. | Re-read Governance rows and parent container docs. | Run c3x verify plus targeted lookup for changed files. |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| Code, docs, tests, prompts | Goal, Governance, Contract, and Change Safety sections. | Names and framework shape may vary; behavior and boundaries may not. | c3x check, c3x verify, and relevant tests. |

## Architecture Details

## Business Purpose

Organization administration for owners. Manage who has access, what teams exist, how approval workflows are configured, and audit/monitor all system activity. All screens require owner role -- server functions enforce via `rbacQueries.isOwner`.

## Screens

### UserManagementScreen

Manage user accounts and role assignments. Master-detail layout with searchable user list showing email, team, roles, status. Detail pane shows user info and role management.

**Actions**: `adminCreateUser`, `adminUpdateUser`, `adminDeleteUser` (soft delete), `adminAssignRole`, `adminRevokeRole`

### TeamManagementScreen

Manage teams and their capabilities. Master-detail layout with team list showing member count. Detail shows team info and member list. Teams can only be deleted if they have no members.

**Actions**: `adminCreateTeam`, `adminUpdateTeam`, `adminDeleteTeam`

### OrganizationScreen

Combined users + teams in a tabbed tile-grid interface. Card-based layout with avatar tiles for users and expandable tiles for teams. Dropdown menus for edit/delete actions. Alternative visual view of the same data as UserManagement + TeamManagement.

**Tabs**: Users (with count badge), Teams (with count badge). Shared search across tabs.

### ApprovalConfigScreen

Configure multi-step approval workflows. Master-detail: flow list on left, flow detail + steps on right. Each step has a name, mode (`anyof` = any approver suffices, `allof` = all must approve), and a list of approver emails. Steps can be reordered and renumbered.

**Actions**: `adminListApprovalFlows`, `adminToggleApprovalFlowActive`, `adminUpdateApprovalFlow`

### AuditLogScreen

Full-page table of audit trail entries with server-side pagination. Filters: table name, action type, user, date range. Expandable rows show before/after JSON diff of the change.

**Actions**: `adminListAuditEntries`

### NotificationLogScreen

Full-page table of notification dispatch history with server-side pagination. Filters: recipient email, channel (email/in_app), status (sent/failed/pending), date range. Expandable rows show payload JSON, error details, and retry button for failed notifications.

**Actions**: `adminListNotificationLogs`, `adminRetryNotification` (republishes to JetStream)

## UI Pattern Notes (2026-02-26)

- **Aligned with refs**: alert wrapper pattern, search input pattern, tab usage (Organization), admin-page/admin-table structure, and mobile table column prioritization.
- **Implemented mobile enhancements**: Audit/Notification log screens use collapsible filters on mobile and sticky bottom pagination while filters are closed.
- **Known drift**: delete confirmation still uses native `window.confirm()` in UserManagement, TeamManagement, and Organization screens (target pattern is `ConfirmModal`/`ConfirmDrawer`).

## Data Flow

```
All screens use local component state (useState), not shared atoms.
Data fetched via admin server functions on mount and after mutations.
Server functions: @/server/functions/admin (20+ functions)
All operations gated by owner role check.
```

## Key Wiring

- **Server functions**: `@/server/functions/admin` -- user CRUD, team CRUD, role CRUD, approval flow config, audit queries, notification queries + retry
- **Components**: MasterDetailLayout (users, teams, approval config), admin-page/admin-table layout (audit, notifications, organization), Drawer (all CRUD forms), DropdownMenu (org tile actions), native `confirm()` for some delete flows (migration pending)
- **Shared**: `ALL_CAPABILITIES` from `@/lib/capabilities` for team capability assignment

## Dependencies

| Direction | What | From/To |
| --- | --- | --- |
| IN (uses) | See component wiring and code sections in this document. | c3-1 |
| OUT (provides) | Exposes behavior described by this component to dependent callers. | c3-1 |
