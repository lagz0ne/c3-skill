---
id: ref-sft-behavioral-spec
c3-version: 4
c3-seal: 0fcc2f903d0102d5a2b1d923c95ca6148a1dbb99a1bbef68391ec448227e2e5e
title: SFT Behavioral Spec
type: ref
goal: Single source of truth for UI behavioral structure — what screens exist, what regions compose them, what events they emit, what state transitions govern them, and what flows connect them. Complements C3 component docs (which own *what* and *why*) with machine-queryable *behavior*.
scope:
    - c3-104
    - c3-105
    - c3-106
    - c3-107
    - c3-109
---

# SFT Behavioral Spec

## Goal

Single source of truth for UI behavioral structure — what screens exist, what regions compose them, what events they emit, what state transitions govern them, and what flows connect them. Complements C3 component docs (which own *what* and *why*) with machine-queryable *behavior*.

## Choice

**SFT** (`sft` CLI tool) with a SQLite database at `.sft/db`. Screens, regions, events, transitions, tags, flows, components, and attachments are stored relationally and queryable via `sft show`, `sft query`, `sft impact`, and raw SQL.

The spec captures:

- **Screens** — top-level routes with archetype tags (master-detail, admin-table, tabbed-operational, tabbed-tile)
- **Regions** — composable UI zones within screens (lists, detail panes, drawers, dialogs, tabs)
- **Events** — user interactions emitted by regions
- **Transitions** — state changes triggered by events (from → to + action), including `navigate()` for cross-screen nav
- **Flows** — ordered event sequences representing end-to-end user journeys
- **Tags** — classification labels (archetype, access control)
- **Components** — bound React component types with props (maps spec entities to real code)
- **Attachments** — screenshots, wireframes, and other files attached to entities

## Why

C3 docs describe architecture and business purpose. SFT adds the behavioral layer — queryable structure that answers "what can the user do on this screen?" and "what happens when they do it?" without reading code. Key benefits:

- **Impact analysis**: `sft impact` before changes shows what screens/flows are affected
- **Spec drift detection**: `sft validate` catches orphaned events, missing handlers
- **Onboarding**: `sft show` gives a complete behavioral map in seconds
- **Flow tracing**: Named flows connect events across screens (e.g., PRApprovalFlow)
- **Diagrams**: `sft diagram` generates Mermaid state machines, nav graphs, and flow sequences
- **Component mapping**: `sft component` shows which React component implements each entity
- **Rendering**: `sft render` generates a json-render element tree from the spec

## How

### Reading the spec

```bash
sft show                          # full text spec
sft show --json                   # structured output
sft query screens                 # list all screens
sft query regions                 # list all regions
sft query events                  # list all events
sft query flows                   # list all flows
sft query states <entity>         # state machine for entity
sft impact <screen|region> <name> # dependency analysis
sft validate                      # check for issues
```

### Diagrams (Mermaid)

```bash
sft diagram states <name>         # state machine for a screen/region
sft diagram nav                   # screen-to-screen navigation graph
sft diagram flow <name>           # flow sequence diagram
```

### Components

```bash
sft component <entity>                          # show bound component (JSON)
sft component <entity> <Type> [--props <json>]  # bind a UI component
sft component <entity> --rm                     # unbind component
```

All screens and regions are bound to their actual React component names with file paths in props.

### Attachments

```bash
sft attach <entity> <file> [--as <name>]  # attach file to entity
sft list [entity]                          # list attachments
sft cat <entity> <name>                    # read attachment content
```

### Rendering

```bash
sft render                        # generate json-render element tree
sft render | jq '.elements.Home'  # inspect specific screen
```

### Updating the spec

When adding a new screen or modifying UI behavior:

1. `sft show` — understand current state
2. Add entities top-down: screen → regions → events → transitions → flows
3. Bind components: `sft component <entity> <ReactComponent> --props '{"file": "path"}'`
4. Add `navigate()` actions for cross-screen transitions
5. `sft validate` — check for issues
6. Update the corresponding C3 component doc if the change is architectural

### Current inventory

| Screen | Route | Archetype | Access | Component |
| --- | --- | --- | --- | --- |
| Invoices | /invoices | master-detail | all | InvoiceScreen |
| PaymentRequests | /prs, /approvals | master-detail, dual-mode | all | PaymentRequestsScreen |
| Payments | /payments | admin-table | all | PaymentsScreen |
| Workbench | /workbench | tabbed-operational | workbench capability | WorkbenchScreen |
| AdminUsers | /admin/users | master-detail | owner-only | UserManagementScreen |
| AdminTeams | /admin/teams | master-detail | owner-only | TeamManagementScreen |
| AdminOrganization | /admin/organization | tabbed-tile | owner-only | OrganizationScreen |
| AdminApprovalConfig | /admin/approval-config | master-detail | owner-only | ApprovalConfigScreen |
| AdminAuditLog | /admin/audit-log | admin-table | owner-only | AuditLogScreen |
| AdminNotificationLog | /admin/notifications | admin-table | owner-only | NotificationLogScreen |

### Key flows

| Flow | Sequence | Description |
| --- | --- | --- |
| PRApprovalFlow | create → request → approve → complete | Happy path PR lifecycle |
| PRRejectionFlow | create → request → reject → request → approve → complete | Rejection + resubmit |
| InvoiceToPRFlow | import → select → link → select_pr | Invoice import to PR linkage |
| EndOfPeriodFlow | cleanup → export → upload → import | Monthly close operations |
| BulkApproveFlow | switch_mode → bulk_approve | Approvals mode bulk action |

## Scope

**Applies to:**

- All frontend screens in c3-1 (web frontend container)
- New screens MUST be added to SFT when created
- Components MUST be bound when adding new entities

**Does NOT apply to:**

- Backend API flows (c3-2) — those are documented in C3 component docs
- Infrastructure (c3-4)

## Known Issues

- `component` command does not support `--in` for scoped/ambiguous regions ([lagz0ne/sft#1](https://github.com/lagz0ne/sft/issues/1))
- FilterFooter is ambiguous (2 instances) — component binding blocked by above issue

## Cited By

- c3-104 (InvoiceScreen)
- c3-105 (PaymentRequestsScreen)
- c3-106 (PaymentsScreen)
- c3-107 (Admin Screens)
- c3-109 (Workbench Screen)
