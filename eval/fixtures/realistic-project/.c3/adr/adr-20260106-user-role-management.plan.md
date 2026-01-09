# User/Role Management Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Enable owners to manage users through `/admin` screen with role-based owner protection.

**Architecture:** Backend flows wrap existing `userQueries`/`rbacQueries` with owner guards. Owner role is system-defined with "cannot delete last owner" semantics.

**ADR:** [adr-20260106-user-role-management.md](./adr-20260106-user-role-management.md)

---

## Task 1: Owner Role Seed Migration

Create migration at `apps/start/src/server/dbs/migrations/XXXX_seed_owner_role.ts`:
- Insert role `name='owner'`, `permissions=['users:manage']`, `created_by='system'`
- Use `ON CONFLICT (name) DO NOTHING`

---

## Task 2: Add Owner Check Helpers

Modify `apps/start/src/server/dbs/queries/rbac.ts`, add to `rbacQueries`:

**isOwner(user_email):** Query `user_roles` joined with `roles` where `name='owner'` and not expired. Return `{ isOwner: boolean }`.

**countOwners():** Count users with active owner role. Return `{ count: number }`.

---

## Task 3: Create User Management Flows

Create `apps/start/src/server/flows/user.ts`:

**ownerGuard helper:** Get currentUser from tag, call `rbacQueries.isOwner()`, return `{ allowed: false, reason }` or `{ allowed: true }`.

**Flows (all use ownerGuard):**

| Flow | Logic |
|------|-------|
| listUsers | `userQueries.listUser()` + enrich each with `rbacQueries.getUserRoles()` |
| createUser | `userQueries.newUser({ email, permissions, team })` |
| updateUser | Call `updateUserPermissions/Team/Status` based on input |
| deleteUser | Check `isOwner` → if owner, check `countOwners <= 1` → reject. Else `removeUserByEmail` |
| transferOwnership | Get owner role by name, `assignUserRole` to target |

Export from `flows/index.ts`.

---

## Task 4: Create Admin Server Functions

Create `apps/start/src/server/functions/admin.ts`:

Wrap each flow: `createServerFn` + `middleware` + `validator` (matching flow schema) + handler.
- `adminListUsers` (GET)
- `adminCreateUser`, `adminUpdateUser`, `adminDeleteUser`, `adminTransferOwnership` (POST)

Export from `functions/index.ts`.

---

## Task 5: Create Admin Screen & Route

**Screen** at `apps/start/src/screens/AdminScreen.tsx`:
- `useQuery(['admin','users'])` → `adminListUsers()`
- Mutations for create/delete/transfer with `queryClient.invalidateQueries`
- Table: email, team, status, roles columns + "Make Owner" / "Delete" actions
- Add user form: email input + button
- Access denied state when `!data.success`

**Route** at `apps/start/src/routes/_authed/admin.tsx`:
- Standard TanStack Router file rendering `AdminScreen`

---

## Task 6: Update C3 Documentation

- Create `.c3/c3-2-api-backend/c3-224-user-flows.md` - document flows inventory and owner guard pattern
- Create `.c3/c3-1-web-frontend/c3-124-admin-screen.md` - document admin screen features
- Update container READMEs with new component references
- Update ADR status to `Implemented`

---

## Task 7: Verify & Commit

```bash
cd apps/start && bunx @typescript/native-preview
```

Manual verification:
- Add/delete user works
- Cannot delete last owner (returns error)

```bash
git add -A
git commit -m "feat: implement user/role management with admin screen"
```

---

## Exit Conditions

**The implementation loop exits when ALL conditions pass:**

### 1. Type Check (Required)
```bash
cd apps/start && bunx @typescript/native-preview
```
Exit code 0, no errors.

### 2. E2E Smoke Test (Required)
```bash
cd apps/start && bun run test:e2e -- --grep "admin"
```
Create `apps/start/e2e/admin.spec.ts`:
- Login as owner → navigate to `/admin` → see user list
- Add user → appears in list
- Delete user → removed from list
- Attempt delete last owner → error shown

### 3. Quality Gates (Required)
- No `any` casts in new code
- No `eslint-disable` comments
- Flows follow existing patterns in `pr.ts`
- Screen follows existing patterns in `PaymentRequestsScreen.tsx`

### 4. C3 Docs Updated (Required)
- `c3-124-admin-screen.md` exists
- `c3-224-user-flows.md` exists
- ADR status is `implemented`

---

## Completion Promise

When all exit conditions pass, output:

```
<promise>USER_MANAGEMENT_COMPLETE</promise>
```
