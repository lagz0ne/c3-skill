## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 search "What owns administrator features for users teams audit and approval configuration?"
c3 read c3-107 --full
c3 read c3-210 --full
c3 read adr-20260121-admin-management-features --full
c3 read ref-admin-page-layout --full
c3 read ref-rbac --full
c3 graph c3-107 --depth 1
c3 graph c3-210 --depth 1
```

## Answer

`c3-107` owns frontend Admin Screens: users, teams, audit logs, notification logs, approval config, and organization view. `c3-210` owns backend Admin Flows: team/role/user management and approval configuration. `adr-20260121-admin-management-features` is the implemented feature decision.

## Grounding

`ref-admin-page-layout` governs because admin screens use admin table/master-detail layouts for audit, notification logs, users, teams, and approval configuration. `ref-rbac` governs because every admin flow extracts the user and requires owner access, returning `NOT_OWNER` when unauthorized.

## Caveats

Fixture has no rule entities.
