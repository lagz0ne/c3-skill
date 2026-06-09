# ADMIN-1

## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 search "What owns administrator features for users teams audit and approval configuration?"
c3 read c3-107 --full
c3 read c3-210 --full
c3 read adr-20260121-admin-management-features --full
c3 read ref-admin-page-layout --full
c3 read ref-rbac --full
c3 read ref-sync --full
c3 read recipe-screen-anatomy --full
c3 read recipe-navigation-strategy --full
c3 graph c3-107 --depth 1
c3 graph c3-210 --depth 1
```

## Answer

`c3-107` Admin Screens owns the frontend administrator feature surface:
users, teams, audit logs, notification logs, approval config, and organization
views. It also says all screens require the owner role and server functions
enforce access with `rbacQueries.isOwner`.

`c3-210` Admin Flows owns the backend administrator operations for teams,
roles, users, and approval configuration. Every admin flow extracts the user
from `currentUserTag`, checks `rbacQueries.isOwner`, and returns `NOT_OWNER`
when the caller is not an owner.

`adr-20260121-admin-management-features` is the shipped decision record for the
feature set. It affects `c3-1`, `c3-2`, and `c3-204`.

`ref-admin-page-layout` governs frontend admin table/page structure: header,
filters, scrollable table, pagination, mobile filter behavior, and responsive
columns. `ref-rbac` governs owner-only access, role permissions, owner checks,
and security events. `ref-sync` governs admin mutation broadcasts when data
changes need to reach clients.

Supporting navigation/screen discovery comes from `recipe-screen-anatomy` and
`recipe-navigation-strategy`.

## Grounding

Search returned `c3-107`, `c3-210`, `adr-20260121-admin-management-features`,
and admin layout refs. The `c3-107` graph shows frontend UI/layout refs. The
`c3-210` graph shows backend refs including `ref-rbac`, `ref-sync`,
`ref-server-functions`, and `ref-query-services`.

## Caveats

No `rule-*` entities exist in this fixture. Component Governance, Contract, and
Change Safety rows define owner boundaries; refs explain why the constraints
govern.
