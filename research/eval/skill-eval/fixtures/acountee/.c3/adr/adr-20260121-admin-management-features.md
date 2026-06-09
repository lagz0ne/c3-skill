---
id: adr-20260121-admin-management-features
c3-seal: 5be301f1d7b47343489b7aab8386832c6d6dbe313effdc3ca55dccb839c3053a
title: Admin Management Features (Users, Audit UI, Approval Configuration)
type: adr
goal: 'Document and implement the architectural decision: Admin Management Features (Users, Audit UI, Approval Configuration).'
status: implemented
date: "2026-01-21"
affects:
    - c3-1
    - c3-2
    - c3-204
approved-files:
    - apps/start/src/server/dbs/schema.ts
    - apps/start/src/server/dbs/queries/users.ts
    - apps/start/src/server/dbs/queries/rbac.ts
    - apps/start/src/server/dbs/queries/audit.ts
    - apps/start/src/server/dbs/queries/teams.ts
    - apps/start/src/server/dbs/queries/approvalConfig.ts
    - apps/start/src/server/dbs/queries/index.ts
    - apps/start/src/server/resources/audit.ts
    - apps/start/src/server/flows/user.ts
    - apps/start/src/server/flows/role.ts
    - apps/start/src/server/flows/team.ts
    - apps/start/src/server/flows/audit.ts
    - apps/start/src/server/flows/approvalConfig.ts
    - apps/start/src/server/flows/index.ts
    - apps/start/src/server/flows/types.approvalConfig.ts
    - apps/start/src/server/flows/pr.ts
    - apps/start/src/server/services/pr.ts
    - apps/start/src/server/functions/admin.ts
    - apps/start/src/server/functions/audit.ts
    - apps/start/src/server/functions/approvalConfig.ts
    - apps/start/src/server/functions/index.ts
    - apps/start/src/routes/_authed/admin.tsx
    - apps/start/src/routes/_authed/admin/users.tsx
    - apps/start/src/routes/_authed/admin/roles.tsx
    - apps/start/src/routes/_authed/admin/teams.tsx
    - apps/start/src/routes/_authed/admin/audit.tsx
    - apps/start/src/routes/_authed/admin/approval-config.tsx
    - apps/start/src/screens/paymentRequestHooks.ts
    - apps/start/src/screens/admin/UserManagementScreen.tsx
    - apps/start/src/screens/admin/RoleManagementScreen.tsx
    - apps/start/src/screens/admin/TeamManagementScreen.tsx
    - apps/start/src/screens/admin/AuditLogScreen.tsx
    - apps/start/src/screens/admin/ApprovalConfigScreen.tsx
    - apps/start/src/components/admin/UserForm.tsx
    - apps/start/src/components/admin/RoleForm.tsx
    - apps/start/src/components/admin/TeamForm.tsx
    - apps/start/src/components/admin/AuditFilters.tsx
    - apps/start/src/components/admin/ApprovalFlowEditor.tsx
    - apps/start/src/components/admin/AdminSidebar.tsx
    - apps/start/src/lib/pumped/atoms/admin.ts
    - apps/start/src/lib/pumped/index.ts
    - packages/shared/src/admin.ts
    - packages/shared/src/index.ts
    - apps/e2e/tests/admin.spec.ts
    - apps/e2e/tests/smoke.spec.ts
    - apps/e2e/scripts/run-isolated-tests.sh
    - apps/e2e/playwright.config.ts
    - apps/start/vite.config.ts
    - apps/start/src/styles.css
    - apps/start/src/server/resources/natsConnection.ts
    - .c3/adr/adr-20260121-admin-management-features.plan.md
---

# Admin Management Features

## Goal

Document and implement the architectural decision: Admin Management Features (Users, Audit UI, Approval Configuration).

## Status

**Implemented** - 2026-01-21

## Problem

The system requires administrative capabilities that are currently missing or incomplete:

1. **User Management**: Backend exists (`userFlows.ts`) but there is no UI. Admins cannot visually manage users, assign roles, or manage team membership.
**User Management**: Backend exists (`userFlows.ts`) but there is no UI. Admins cannot visually manage users, assign roles, or manage team membership.
2. **Teams**: Currently hardcoded as an enum (`admin`, `finance`, `bod`) in the users table with a CHECK constraint. This prevents flexible team management - admins cannot create custom teams like "Procurement" or "HR".
**Teams**: Currently hardcoded as an enum (`admin`, `finance`, `bod`) in the users table with a CHECK constraint. This prevents flexible team management - admins cannot create custom teams like "Procurement" or "HR".
3. **Audit Log**: Full backend exists (c3-208) with filtering, export, and statistics capabilities, but there is no UI for discovery and exploration of audit data.
**Audit Log**: Full backend exists (c3-208) with filtering, export, and statistics capabilities, but there is no UI for discovery and exploration of audit data.
4. **Approval Flow Configuration**: Approval flows are hardcoded in `types.approvalConfig.ts` with specific user emails. Any change requires code deployment.
**Approval Flow Configuration**: Approval flows are hardcoded in `types.approvalConfig.ts` with specific user emails. Any change requires code deployment.

## Decision

Implement three admin feature areas with a new `/admin/*` route structure:

### Feature 1: User Management (Full CRUD)

Create UI for complete user lifecycle management:

- List users with their roles and team assignments
- Create new users with initial permissions
- Edit user permissions, team, and status
- Deactivate/suspend users (soft delete)
- Assign/revoke roles to users
- Create and edit custom roles

**Database Change**: Add `teams` table to replace enum constraint.

### Feature 2: Audit Log UI

Create discovery-focused UI for audit data:

- Server-side pagination (50-100 records per page)
- Multi-dimensional filtering (table, user, action, date range)
- Grouping by table/user/date
- Export to CSV/JSON
- Quick navigation to related records

### Feature 3: Approval Flow Configuration

Create UI to edit approval flows stored in database:

- Replace hardcoded `approvalFlow` constant with database table
- UI to edit existing flows (finance, procurement)
- Add/remove steps in a flow
- Configure step mode (`anyof`/`allof`) and assignees
- Activation/deactivation toggle (`is_active`) in UI
- No templates or per-PR customization in this phase

### UI Structure

All admin features under `/_authed/admin/*`:

- `/admin/users` - User management
- `/admin/roles` - Role management
- `/admin/teams` - Team management
- `/admin/audit` - Audit log viewer
- `/admin/approval-config` - Approval flow editor

## In-Flight PR Handling

**Decision: PRs use approval flow snapshot at creation time.**

When an approval flow configuration is modified, pending PRs (those not yet fully approved or rejected) continue using the approval chain that was active when they were created. This ensures:

1. **Predictability**: Users know what approvals are required when they submit a PR
2. **Auditability**: The approval chain used is the one that was agreed upon at submission
3. **No retroactive changes**: Changing config doesn't suddenly add/remove approvers from existing PRs

**Implementation**: When a PR is created, the current approval flow configuration is snapshotted and stored with the PR record. The PR flow (`pr.ts`) and PR service (`pr.ts`) consume this snapshot rather than reading live config.

## Rationale

| Considered | Rejected Because |
| --- | --- |
| Single /admin page with tabs | Too cramped for complex features; URL sharing difficult |
| Modal-based editing | Breaks back button expectations; poor mobile UX |
| Client-side pagination for audit | Memory issues with large audit tables; poor performance |
| Per-PR approval customization | Adds complexity; start simple with global flows |
| Keep teams as enum | Inflexible; requires code changes for new teams |
| Apply config changes to in-flight PRs | Unpredictable UX; potential security concerns |

**Why MasterDetailLayout for most screens:**

- Established pattern in c3-1 (InvoiceScreen, PaymentRequestsScreen)
- Consistent UX across the application
- Good for list + detail workflows

**Why table-based layout for Audit:**

- Discovery-focused, not detail-focused
- Users scan many rows quickly
- Filtering/sorting are primary interactions

**Approval Mode Naming**: Using `anyof`/`allof` to match existing codebase conventions. Note: `ref-approval-chain` pattern should be updated to standardize on this naming.

## Affected Layers

| Layer | Document | Change |
| --- | --- | --- |
| c3-204 | Drizzle ORM | Add teams table, approval_flows table, approval_flow_steps table |
| c3-2 | API Backend | New flows for teams, roles, approval config; extend audit flows |
| c3-1 | Web Frontend | New admin routes, screens, components, state atoms |

## Database Schema Changes

### New Table: teams

```sql
CREATE TABLE teams (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  description TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  created_by TEXT,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_teams_name ON teams(name);
```

### New Table: approval_flows

```sql
CREATE TABLE approval_flows (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,  -- 'finance', 'procurement', etc.
  description TEXT,
  version INTEGER NOT NULL DEFAULT 1,  -- Incremented on each edit for auditability
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  created_by TEXT,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_approval_flows_name ON approval_flows(name);
CREATE INDEX idx_approval_flows_active ON approval_flows(is_active);
```

### New Table: approval_flow_steps

```sql
CREATE TABLE approval_flow_steps (
  id SERIAL PRIMARY KEY,
  flow_id INTEGER NOT NULL REFERENCES approval_flows(id) ON DELETE CASCADE,
  step_number INTEGER NOT NULL,
  name TEXT NOT NULL,
  mode TEXT NOT NULL CHECK (mode IN ('anyof', 'allof')),
  user_emails TEXT[] NOT NULL DEFAULT '{}',
  UNIQUE(flow_id, step_number)
);

CREATE INDEX idx_approval_flow_steps_flow_id ON approval_flow_steps(flow_id);
```

### Modify: users table

```sql
-- Remove CHECK constraint for team enum
ALTER TABLE users DROP CONSTRAINT users_team_check;

-- Add foreign key to teams (optional, allows NULL for backwards compat)
ALTER TABLE users ADD COLUMN team_id INTEGER REFERENCES teams(id);

-- Migration: Create default teams and link existing users
INSERT INTO teams (name) VALUES ('admin'), ('finance'), ('bod');
UPDATE users SET team_id = (SELECT id FROM teams WHERE name = users.team);
```

## Component Architecture

```
routes/_authed/admin.tsx          <- Layout with AdminSidebar
  routes/_authed/admin/users.tsx  <- UserManagementScreen
  routes/_authed/admin/roles.tsx  <- RoleManagementScreen
  routes/_authed/admin/teams.tsx  <- TeamManagementScreen
  routes/_authed/admin/audit.tsx  <- AuditLogScreen
  routes/_authed/admin/approval-config.tsx <- ApprovalConfigScreen

screens/admin/
  UserManagementScreen.tsx        <- MasterDetailLayout
  RoleManagementScreen.tsx        <- MasterDetailLayout
  TeamManagementScreen.tsx        <- MasterDetailLayout
  AuditLogScreen.tsx              <- Table layout with filters
  ApprovalConfigScreen.tsx        <- MasterDetailLayout + is_active toggle

components/admin/
  AdminSidebar.tsx                <- Navigation within /admin
  UserForm.tsx                    <- Create/Edit user
  RoleForm.tsx                    <- Create/Edit role
  TeamForm.tsx                    <- Create/Edit team
  AuditFilters.tsx                <- Filter panel for audit
  ApprovalFlowEditor.tsx          <- Visual flow step editor
```

## Data Flow

### User Management Flow

```UserManagementScreen
  -> useAtomValue(adminUsersAtom)     <- Fetches via adminListUsers
  -> actions.createUser(data)         <- Calls adminCreateUser server fn
  -> actions.updateUser(data)         <- Calls adminUpdateUser server fn
  -> actions.deleteUser(email)        <- Calls adminDeleteUser server fn
  -> actions.assignRole(email, role)  <- NEW: Calls assignUserRole server fn
```

### Audit Log Flow

```AuditLogScreen
  -> useState for filters (local)
  -> useQuery for paginated data       <- Calls listAuditEntries with offset/limit
  -> Server returns { entries, total, page, pageSize }
  -> UI renders DataTable with pagination controls
```

### Approval Config Flow

```ApprovalConfigScreen
  -> useAtomValue(approvalFlowsAtom)   <- NEW atom
  -> actions.updateApprovalFlow(...)   <- Calls updateApprovalFlow server fn
  -> actions.toggleFlowActive(...)     <- Calls toggleApprovalFlowActive server fn
  -> On save: invalidate PR creation UI (uses updated flow)
  -> Version column auto-increments on each edit
```

## Implementation Phases

### Phase 1: Database & Backend Foundation (2-3 days)

1. Create migration for `teams` table
2. Create migration for `approval_flows` and `approval_flow_steps` tables (with `version` column)
3. Update Drizzle schema (`schema.ts`)
4. Create `teamsQueries.ts` and `approvalConfigQueries.ts`
5. Create `team.ts` and `approvalConfig.ts` flows
6. Extend `rbacQueries` for role CRUD operations
7. Seed default teams and migrate approval flows from hardcoded config

### Phase 2: Admin Routes & Layout (1 day)

1. Create `/_authed/admin.tsx` layout with AdminSidebar
2. Create route files for each admin section
3. Implement permission check (owner-only access)

### Phase 3: User Management UI (2-3 days)

1. Create `UserManagementScreen.tsx` with MasterDetailLayout
2. Implement user list with role/team badges
3. Implement UserForm for create/edit
4. Implement role assignment UI
5. Add state atoms for admin users

### Phase 4: Team & Role Management UI (1-2 days)

1. Create `TeamManagementScreen.tsx`
2. Create `RoleManagementScreen.tsx`
3. Implement CRUD forms for both

### Phase 5: Audit Log UI (2 days) [Can run parallel to Phases 3-4]

1. Extend `auditFlows` with pagination support
2. Create `AuditLogScreen.tsx` with table layout
3. Implement `AuditFilters.tsx` component
4. Add pagination controls
5. Add export functionality

**Note:** Phase 5 has no dependencies on Phases 3-4 and can be executed in parallel to optimize timeline.

### Phase 6: Approval Config UI (2-3 days)

1. Create `ApprovalConfigScreen.tsx` with `is_active` toggle
2. Implement `ApprovalFlowEditor.tsx` for visual editing
3. Create server functions for config CRUD
4. Update PR creation to read from database instead of hardcoded config
5. Update `pr.ts` flow and `pr.ts` service to snapshot approval config at PR creation
6. Add data migration to seed existing flows

### Phase 7: Testing & Polish (1-2 days)

1. Add E2E tests for admin flows (`apps/e2e/tests/admin.spec.ts`)
2. Add permission boundary tests
3. Verify audit trail captures admin actions
4. Test in-flight PR behavior when config changes

## Approved Files

The following files are approved for modification under this ADR:

```yaml
approved-files:
  # Database Schema & Queries
  - apps/start/src/server/dbs/schema.ts
  - apps/start/src/server/dbs/queries/users.ts
  - apps/start/src/server/dbs/queries/rbac.ts
  - apps/start/src/server/dbs/queries/audit.ts
  - apps/start/src/server/dbs/queries/teams.ts
  - apps/start/src/server/dbs/queries/approvalConfig.ts
  # Flows
  - apps/start/src/server/flows/user.ts
  - apps/start/src/server/flows/role.ts
  - apps/start/src/server/flows/team.ts
  - apps/start/src/server/flows/audit.ts
  - apps/start/src/server/flows/approvalConfig.ts
  - apps/start/src/server/flows/index.ts
  - apps/start/src/server/flows/types.approvalConfig.ts
  - apps/start/src/server/flows/pr.ts
  - apps/start/src/server/services/pr.ts
  # Server Functions
  - apps/start/src/server/functions/admin.ts
  - apps/start/src/server/functions/audit.ts
  - apps/start/src/server/functions/approvalConfig.ts
  - apps/start/src/server/functions/index.ts
  # Routes
  - apps/start/src/routes/_authed/admin.tsx
  - apps/start/src/routes/_authed/admin/users.tsx
  - apps/start/src/routes/_authed/admin/roles.tsx
  - apps/start/src/routes/_authed/admin/teams.tsx
  - apps/start/src/routes/_authed/admin/audit.tsx
  - apps/start/src/routes/_authed/admin/approval-config.tsx
  # Screens
  - apps/start/src/screens/paymentRequestHooks.ts
  - apps/start/src/screens/admin/UserManagementScreen.tsx
  - apps/start/src/screens/admin/RoleManagementScreen.tsx
  - apps/start/src/screens/admin/TeamManagementScreen.tsx
  - apps/start/src/screens/admin/AuditLogScreen.tsx
  - apps/start/src/screens/admin/ApprovalConfigScreen.tsx
  # Components
  - apps/start/src/components/admin/UserForm.tsx
  - apps/start/src/components/admin/RoleForm.tsx
  - apps/start/src/components/admin/TeamForm.tsx
  - apps/start/src/components/admin/AuditFilters.tsx
  - apps/start/src/components/admin/ApprovalFlowEditor.tsx
  - apps/start/src/components/admin/AdminSidebar.tsx
  # State
  - apps/start/src/lib/pumped/atoms/admin.ts
  - apps/start/src/lib/pumped/index.ts
  # Shared Types
  - packages/shared/src/admin.ts
  - packages/shared/src/index.ts
  # E2E Tests
  - apps/e2e/tests/admin.spec.ts
  - apps/e2e/scripts/run-isolated-tests.sh
  - apps/e2e/playwright.config.ts
```

**Gate behavior:** Only these files can be edited when status is `accepted`.

## Verification

- [ ] Teams can be created, edited, and users assigned to them
- [ ] Users can be created, edited, deactivated via UI
- [ ] Roles can be created and assigned to users
- [ ] Audit log displays with server-side pagination (50+ records load correctly)
- [ ] Audit log filtering works across all dimensions
- [ ] Audit log export produces valid CSV/JSON
- [ ] Approval flows can be edited and changes reflect in new PRs
- [ ] Approval flow `is_active` toggle works correctly in UI
- [ ] Approval flow `version` increments on each edit
- [ ] Existing PRs continue to use their original approval flow (snapshot behavior)
- [ ] All admin actions are recorded in audit log
- [ ] Non-owners cannot access /admin/* routes
- [ ] E2E tests pass for admin workflows (`apps/e2e/tests/admin.spec.ts`)

## Context

N.A - historical ADR; original context is captured in the git log around the ADR date and in the current code that implements the decision.

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| N.A - historical | Shipped via git commits; the c3 topology and code-map reflect the resulting structure. | c3x list --include-adr and git log around the ADR date |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| N.A - historical | Current .c3 entities, refs, and code-map are the post-change state. | c3x verify and c3x check |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| N.A - historical | Enforcement is implicit in the currently linked components and refs. | c3x graph and cited ref ids on the relevant components |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| N.A - historical | Alternatives were considered at decision time; rationale is preserved in the original commit message or branch discussion. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| N.A - historical | Risks were assessed pre-merge; the decision has since shipped without outstanding incidents tied to this ADR. | git log and project test suite |
