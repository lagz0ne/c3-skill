# ADMIN-1

## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 list
c3 check
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

Frontend ownership is `c3-107` Admin Screens. It owns users, teams, audit logs,
notification logs, approval config, and organization views. Its business
purpose says every admin screen requires the owner role and that server
functions enforce access through `rbacQueries.isOwner`.

Backend ownership is `c3-210` Admin Flows. It owns teams, roles, users, and
approval configuration operations. Its authorization section says every admin
flow extracts the user from `currentUserTag`, checks `rbacQueries.isOwner`, and
rejects with `NOT_OWNER` when the caller is not an owner.

The shipped decision record is `adr-20260121-admin-management-features`, which
implemented the admin management feature set and affected `c3-1`, `c3-2`, and
`c3-204`.

The main frontend governing ref is `ref-admin-page-layout`; it governs admin
full-page table structure, pagination, filters, responsive table columns, and
empty states. Backend access is governed by `ref-rbac`, because owner checks,
roles, permissions, and security events are defined there. `ref-sync` applies
to admin mutations that broadcast client updates.

`recipe-screen-anatomy` and `recipe-navigation-strategy` are supporting docs
for route and screen discovery.

## Grounding

Search returned `c3-107`, `c3-210`, the admin ADR, and admin layout refs.
Graphs show `c3-107` uses `ref-admin-page-layout`,
`ref-responsive-layout`, UI refs, and variant refs; `c3-210` uses
`ref-rbac`, `ref-sync`, `ref-server-functions`, `ref-query-services`,
`ref-pumped-fn`, and `ref-structured-logging`.

## Caveats

No `rule-*` entities exist in this fixture, so admin governance should be
explained through component contracts, refs, and the historical ADR.
