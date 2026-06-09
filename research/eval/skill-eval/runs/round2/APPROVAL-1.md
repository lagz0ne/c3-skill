# APPROVAL-1

## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 search "Where does approval workflow live and what governs changes to approvals?"
c3 read recipe-approval-workflow --full
c3 read c3-205 --full
c3 read c3-212 --full
c3 read c3-105 --full
c3 read c3-211 --full
c3 read ref-approval-chain --full
c3 read ref-audit-trail --full
c3 read ref-sync --full
c3 graph c3-205 --depth 1
c3 graph c3-212 --depth 1
c3 graph ref-approval-chain --depth 1
```

## Answer

`recipe-approval-workflow` is the best orientation doc. It traces PRs through
`draft -> pending -> approved -> completed`, and says approval semantics are
governed by multi-step chains.

`c3-205` PR Flows owns the core backend approval lifecycle: create, request
approvals, approve, reject, recall, unapprove, complete, uncomplete, remove,
bulk approve, and PR reads/attachment changes. This is the primary owner for
approval state transitions.

`c3-212` Workbench Flows owns workbench extensions to the approval lifecycle,
including approved PR export and paid PR import. Paid import validates approved
PRs, writes payment references, and completes PRs.

`c3-105` PaymentRequestsScreen owns the UI interaction surface: PR mode,
Approvals mode, approval queue, bulk approve, and approval-chain display.
`c3-211` Notification System is related because step advancement notifies the
next approvers.

`ref-approval-chain` governs anyof/allof semantics, approval records, step
advancement, and app-level validation in `prService`. `ref-audit-trail` governs
mutation capture and warns not to double-write audit entries for PR tables that
are trigger-audited. `ref-sync` governs deltas and acks for client
reconciliation.

## Grounding

Search returned the approval recipe, `c3-105`, `c3-205`, `c3-212`, and approval
refs. Reads separated ownership across UI, core flow, and workbench extension.
The `ref-approval-chain` graph shows it cited by `c3-205` and
`recipe-approval-workflow`.

## Caveats

No `rule-*` entities exist in this fixture. If changing approvals, treat ADRs
as historical context and verify against current component/ref contracts.
