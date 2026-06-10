## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 search "bulk approve payment requests notify approvers sync"
c3 read c3-105 --full
c3 read c3-205 --full
c3 read c3-211 --full
c3 read ref-sync --full
c3 graph c3-205 --depth 1
c3 graph c3-211 --depth 1
```

## Answer

Mass approval starts in `c3-105`: Approvals mode lets approvers use bulk approve. The backend command is `c3-205` `approveAll`, which iterates `pr_ids`, approves each PR, and returns approved/failed arrays. The approval state follows `ref-approval-chain`.

After mutation, sync is handled by `ref-sync`: services emit deltas and flows ack the `executionId`. `c3-101` receives those messages through NATS WebSocket subscriptions so PR state updates on clients. User-facing notifications go through `c3-211`, which owns notification service, publisher, dispatcher, logs, and channels.

## Grounding

`ref-approval-chain` governs because bulk approval still uses normal PR approval semantics. `ref-sync` governs because the PR mutation must be propagated to connected clients after the flow changes state.

## Caveats

This answer follows the component/ref graph. It does not inspect source code paths beyond the C3 fixture.
