---
id: recipe-screen-anatomy
c3-seal: c36451ebedc8b730ec612aefc68bf48caf9999cd01cf35a1fc2daa28365fc173
title: Screen Anatomy
type: recipe
goal: 'Orient fast. When an agent needs to touch UI, this recipe answers: "what screens exist, what does each own, and where do I dig deeper?"'
sources:
    - c3-1
    - c3-104
    - c3-105
    - c3-106
    - c3-107
    - c3-109
    - ref-admin-page-layout
    - ref-detail-content-strategy
    - ref-filter-footer
    - ref-list-view-patterns
    - ref-master-detail-layout
---

# Screen Anatomy

## Goal

Orient fast. When an agent needs to touch UI, this recipe answers: "what screens exist, what does each own, and where do I dig deeper?"

## Screen Inventory

| Screen | Route | What it owns | Archetype | Component doc |
| --- | --- | --- | --- | --- |
| Invoices | /invoices | Import, review, archive invoices | Master-Detail | c3-104 |
| Payment Requests | /prs, /approvals | Create, submit, approve PRs | Master-Detail | c3-105 |
| Payments | /payments | Manage payment methods | Admin Table | c3-106 |
| Workbench | /workbench | Bulk ops, cleanup, export | Tabbed Operational | c3-109 |
| Users | /admin/users | Invite, assign roles, manage teams | Tabbed Tile | c3-107 |
| Approval Config | /admin/approval-config | Approval flows, chains, rules | Admin Table | c3-107 |
| Audit Log | /admin/audit-log | Change history viewer | Admin Table | c3-107 |
| Notifications | /admin/notifications | Delivery log | Admin Table | c3-107 |
| Organization | /admin/organization | Org settings | Admin Table | c3-107 |

## Archetypes at a Glance

**Master-Detail** — list pane + detail pane side-by-side (mobile: stacked slide). Uses `MasterDetailLayout`, `FilterFooter`, virtualized lists. Detail pane has header/content/footer zones.

**Admin Table** — single column with `.admin-page` wrapper, `.admin-table`, footer bar with count.

**Tabbed Operational** — tabs at page level, each tab owns its own content (tables, action bars).

**Tabbed Tile** — tabs switching between entity categories (e.g. users vs teams).

## Where to Look

| Question | Go to |
| --- | --- |
| How does the shell/sidebar/routing work? | recipe-navigation-strategy |
| How does a list filter/sort/paginate? | ref-filter-footer, ref-list-view-patterns |
| How does the detail pane compose its content? | ref-detail-content-strategy |
| How do forms/drawers/modals work? | recipe-modal-dialog, ref-form-patterns |
| How does state flow (atoms, sync, mutations)? | recipe-ui-patterns, recipe-realtime-sync |
| How does responsive layout adapt? | recipe-responsive-design, ref-responsive-layout |
| How do badges, formatting, variants work? | ref-ui-patterns, ref-variant-system |
| How does approval workflow flow through screens? | recipe-approval-workflow |
| What CSS classes does each zone use? | ref-master-detail-layout, ref-admin-page-layout |
