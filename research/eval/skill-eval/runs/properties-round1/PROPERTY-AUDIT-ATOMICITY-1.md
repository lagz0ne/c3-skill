## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 search "mass approval partial failure audit trail consistency"
c3 read c3-105 --full
c3 read c3-205 --full
c3 read recipe-approval-workflow --full
c3 read ref-audit-trail --full
c3 read c3-208 --full
c3 graph ref-audit-trail --depth 2 --direction reverse
```

## Answer

Mass approval starts in `c3-105` and is executed by `c3-205` through `approveAll`. That flow iterates `pr_ids`, approves each item, and returns approved/failed arrays. Approval semantics still come from `ref-approval-chain`.

Audit is not a separate per-loop explicit write in `approveAll`. `recipe-approval-workflow` says approval mutations on the `pr` table are audit captured by DB trigger, and `ref-audit-trail` says not to also call `createAuditEntry` for trigger-covered tables. `c3-208` is the read/query surface for audit history and export.

So the audit trail should remain consistent for the records that actually changed: successful PR updates are captured by the `pr` trigger, while failed PRs should not show successful change records. The answer is not that the whole bulk operation is guaranteed all-or-nothing; `approveAll` reports per-item failures.

## Grounding

`ref-approval-chain` governs the approval progression. `ref-audit-trail` governs why PR flow code should rely on trigger audit for `pr` mutations instead of duplicating audit writes.

## Caveats

The fixture has known `c3 check` drift. I did not inspect source code beyond C3 docs.
