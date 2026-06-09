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

Frontend ownership is `c3-107`; backend ownership is `c3-210`. `c3-107` owns users, teams, audit logs, notification logs, approval config, and organization admin screens. `c3-210` owns team, role, user, and approval configuration flows. `adr-20260121-admin-management-features` records the shipped admin feature decision.

## Grounding

`ref-admin-page-layout` governs `c3-107` because admin audit/notification tables and master-detail admin screens use that page layout. `ref-rbac` governs admin flow access because every admin flow checks owner status and rejects `NOT_OWNER`.

## Caveats

No `rule-admin` entity exists in this fixture.
