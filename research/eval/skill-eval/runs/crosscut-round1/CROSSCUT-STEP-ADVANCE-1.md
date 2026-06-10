## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 search "bulk approve payment requests notify approvers sync"
c3 read c3-205 --full
c3 read ref-approval-chain --full
c3 read adr-20260202-notification-on-step-advance --full
c3 read c3-211 --full
c3 read ref-pull-dispatcher --full
c3 read ref-sync --full
c3 graph c3-205 --depth 1
c3 graph c3-211 --depth 1
```

## Answer

Later-step approvers hear only after the prior step completes because the PR flow uses explicit step advancement. `c3-205` owns `requestApprovals`, `approvePr`, and `approveAll`; `adr-20260202-notification-on-step-advance` added `stepAdvanced` to signal when `approvePr` or `approveAll` should notify the next step.

`ref-approval-chain` governs `anyof` and `allof`: the step may advance after one approver or after all approvers in the current step. `c3-211` then sends the notification, with channel registration shaped by `ref-pull-dispatcher`. Sync still uses `ref-sync`.

## Grounding

`ref-approval-chain` governs because approval semantics decide whether the current step has completed. `ref-pull-dispatcher` governs the notification mechanism because channels subscribe to the dispatcher. `ref-sync` governs state consistency because PR updates still move over NATS WebSocket.

## Caveats

Historical ADR evidence is used only to explain why current component docs mention step-advance notifications.
