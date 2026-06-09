# APPROVAL-1

## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 list
c3 check
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

The orientation doc is `recipe-approval-workflow`. It traces payment requests
from `draft -> pending -> approved -> completed`, including workbench bulk
operations and cross-cutting contracts.

Core backend ownership lives in `c3-205` PR Flows. It owns create, update,
request approvals, approve, unapprove, reject, recall, complete, uncomplete,
remove, bulk approve, attachment changes, and PR reads. It is the main owner for
approval state transitions.

Workbench extensions live in `c3-212` Workbench Flows. It owns approved PR
export and paid PR import; importing paid PRs validates that the PR is
approved, writes a payment reference, then completes the PR.

The UI participant is `c3-105` PaymentRequestsScreen. It owns PR and Approvals
screen modes, the approval queue, bulk approve, PR actions, and the approval
chain display. `c3-211` Notification System is related because the recipe says
step advancement notifies next-step approvers.

`ref-approval-chain` governs approval semantics: steps are `anyof` or `allof`,
state advances in app-level service logic, and approval records show who signed
each step. `ref-audit-trail` governs mutation capture; approval mutations on PR
tables are trigger-audited and should not be double-written explicitly.
`ref-sync` governs mutation deltas and acks so clients reconcile approval
changes.

## Grounding

Search returned `recipe-approval-workflow`, `c3-105`, `c3-205`, `c3-212`, and
approval refs. The graph for `ref-approval-chain` shows it cited by `c3-205`
and `recipe-approval-workflow`. Component contract rows separate UI
(`c3-105`), core PR flow (`c3-205`), and workbench extension (`c3-212`)
ownership.

## Caveats

No `rule-*` entities exist in this fixture. Approval changes should start from
current components and refs; ADRs are historical unless the user asks for past
decisions.
