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

`recipe-approval-workflow` is the narrative trace. Core approval work lives in `c3-205` PR Flows: create, request approvals, approve, reject, recall, complete, and bulk approve. `c3-212` extends the workflow through workbench paid-PR import, and `c3-105` owns the PR/approvals UI.

The lifecycle is `draft -> pending -> approved -> completed`. Approval steps use `anyof` and `allof`.

## Grounding

`ref-approval-chain` governs approval semantics because it defines multi-step approval, `anyof`/`allof`, and step advancement. `ref-audit-trail` governs change capture because approval mutations must remain reconstructable without duplicate audit writes. `ref-sync` governs client consistency because approval mutations emit deltas and flow acks.

## Caveats

No `rule-approval` entity exists.
