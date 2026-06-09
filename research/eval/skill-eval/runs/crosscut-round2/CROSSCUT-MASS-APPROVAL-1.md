## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 search "bulk approve payment requests notify approvers sync"
c3 read c3-105 --full
c3 read c3-205 --full
c3 read ref-approval-chain --full
c3 read ref-sync --full
c3 read c3-101 --full
c3 read c3-211 --full
c3 read adr-20260121-notification-system --full
c3 read adr-20260202-notification-on-step-advance --full
c3 graph c3-205 --depth 1
c3 graph c3-211 --depth 1
```

## Answer

Action starts in `c3-105` Approvals mode, where users can bulk approve. Mutation happens in `c3-205` `approveAll`: it iterates `pr_ids`, approves each PR, collects approved/failed results, and follows `ref-approval-chain`. `adr-20260202-notification-on-step-advance` connects bulk approval to next-step notification when a PR advances.

Sync is `ref-sync`: services emit deltas, flows ack the `executionId`, and `c3-101` receives NATS WebSocket messages on the broadcast subscription. Notification is `c3-211`, backed by `adr-20260121-notification-system`; it dispatches approval notifications to target users. Emergent property: approval and sync do not depend on notification delivery because the flow treats notification as async and errors are logged, not thrown.

## Grounding

`ref-approval-chain` governs because bulk approval must preserve step/mode semantics. `ref-sync` governs because successful PR mutations need NATS WebSocket deltas/acks for connected clients.

## Caveats

No source inspection beyond C3 fixture.
