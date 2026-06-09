## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 search "Where does approval workflow live and what governs changes to approvals?"
c3 read recipe-approval-workflow --full
c3 read c3-205 --full
c3 read c3-212 --full
c3 read c3-105 --full
c3 read ref-approval-chain --full
c3 read ref-audit-trail --full
c3 read ref-sync --full
c3 graph c3-205 --depth 1
```

## Answer

`recipe-approval-workflow` is the trace. `c3-205` owns core PR mutations: create, request approvals, approve, reject, recall, complete, and bulk approve. `c3-212` extends approved PR handling through workbench paid import/export operations. `c3-105` owns the PR and approvals screen. The lifecycle is `draft -> pending -> approved -> completed`; approval modes include `anyof` and `allof`.

## Grounding

`ref-approval-chain` governs because it defines approval steps, `anyof`/`allof`, and advancement. `ref-audit-trail` governs because approval changes are persistent record mutations and must be reconstructable. `ref-sync` governs because mutations emit client deltas and flow acks.

## Caveats

Fixture has no rule entities.
