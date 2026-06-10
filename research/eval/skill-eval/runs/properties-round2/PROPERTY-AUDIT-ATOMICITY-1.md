## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 search "mass approval partial failure audit trail consistency"
c3 read c3-105 --full
c3 read c3-205 --full
c3 read recipe-approval-workflow --full
c3 read ref-audit-trail --full
c3 read recipe-audit-and-compliance --full
c3 read c3-208 --full
c3 read c3-202 --full
c3 graph ref-audit-trail --depth 2 --direction reverse
```

## Answer

Mass approval starts in `c3-105` Approvals mode and executes in `c3-205` `approveAll`. The flow iterates `pr_ids`, approves each PR, and collects approved/failed arrays, so the batch can report partial failure. `ref-approval-chain` governs because the per-PR mutation still has to preserve approval step and mode semantics.

The audit mechanism is trigger-based for PR records. `recipe-approval-workflow` says approval mutations are audit captured by a DB trigger on the `pr` table. `ref-audit-trail` governs because it defines why trigger-covered tables must not be double-written: do not also call `createAuditEntry` for `pr`. `recipe-audit-and-compliance` confirms the DB trigger audit surface for `pr`, and `c3-202` supplies the `transactionTag` context that lets transaction-scoped writes and trigger actor attribution stay tied to the mutation.

Emergent property: **audit atomicity** / audit consistency under partial failure. The consistency boundary is the committed PR mutation, not necessarily the whole bulk batch. A successful PR update gets a trigger audit row; a failed or rolled-back PR attempt should not leave an orphan audit row. `c3-208` then exposes those audit entries through history/list/export/stat flows.

## Grounding

`ref-approval-chain` governs PR approval progression. `ref-audit-trail` governs audit atomicity because it says audit writes must be atomic with the mutation and trigger-covered tables should not receive duplicate explicit audit entries.

## Caveats

The fixture has known `c3 check` drift. No `rule-*` entities found in the fixture.
