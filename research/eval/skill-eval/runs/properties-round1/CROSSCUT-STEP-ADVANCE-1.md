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
c3 read c3-101 --full
c3 graph c3-205 --depth 1
c3 graph c3-211 --depth 1
```

## Answer

`c3-205` owns `requestApprovals`, `approvePr`, and `approveAll`. `ref-approval-chain` defines whether a step completes via `anyof` or `allof`. `adr-20260202-notification-on-step-advance` added `stepAdvanced` so the flow can notify only when the step actually advances; it avoids notifying on every approval and sends no notification when fully approved.

Sync remains `ref-sync`: NATS WebSocket deltas/acks update clients, and `c3-101` resolves the execution tracker. Notification is `c3-211`, with channel self-registration governed by `ref-pull-dispatcher`.

## Grounding

`ref-approval-chain` governs because step advancement depends on mode semantics. `ref-pull-dispatcher` governs because notification channels subscribe to the dispatcher. `ref-sync` governs because approval mutations still need browser state updates.

## Caveats

The ADR is historical but matches current `c3-205` behavior.
