# CROSSCUT-MASS-APPROVAL-1 Gold Probe

## Evidence Commands

- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 search "bulk approve payment requests notify approvers sync" # c3 search
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 search "approval notification sync non blocking NATS websocket"
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read c3-105 --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read c3-205 --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read ref-approval-chain --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read ref-sync --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read c3-101 --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read c3-211 --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read adr-20260121-notification-system --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read adr-20260202-notification-on-step-advance --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read recipe-realtime-sync --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 graph c3-205 --depth 1 # c3 graph
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 graph c3-211 --depth 1
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 graph ref-sync --depth 2 --direction reverse

## Answer

Mass approval starts in the frontend but the durable mutation owner is backend PR Flows. `c3-105` owns PaymentRequestsScreen, including Approvals mode, approve/reject actions, and bulk approve. That is the action surface. `c3-205` owns `approveAll`, and its Operations table says `approveAll` iterates `pr_ids`, approves each PR, collects approved/failed arrays, and has side effects of sync plus conditional notifications per PR.

State mutation: `c3-205` calls PR approval logic governed by `ref-approval-chain`. `ref-approval-chain` governs because it defines approval records, current step, `anyof` and `allof`, and the app-level progression decision inside `prService.approve`. In the causal path, `approveAll` delegates each PR to that same approval mechanism, so a PR can stay pending, advance a step, become approved, or fail independently.

Sync mechanism: `ref-sync` governs because it defines the two-layer NATS WebSocket contract: services emit deltas and flows emit acknowledgements with the same `executionId`. `c3-205` cites `ref-sync`, and the `c3-205` graph shows it as a direct dependency. `c3-101` is the frontend observer because its `natsSync` atom subscribes to `sync.broadcast` for delta and ack messages and updates the PR store through the NATS WebSocket path.

Notification mechanism: `c3-205` calls notification behavior only when the approval result makes notification relevant. `adr-20260202-notification-on-step-advance` is a historical implemented ADR that changed `approvePr` and `approveAll` so they notify next approvers only when `stepAdvanced` is true. `c3-211` is the notification owner: it defines notificationService, notificationPublisher, notificationDispatcher, JetStream `NOTIFICATIONS`, per-channel logs, and in-app delivery. `adr-20260121-notification-system` is the historical implemented ADR that introduced the NATS JetStream notification system for approval workflows.

Direct dependents and observers:

- Action and UI: `c3-105` exposes Approvals mode and bulk approve.
- Mutation and orchestration: `c3-205` owns `approveAll`, uses `ref-approval-chain`, and uses `ref-sync`.
- Sync observer: `c3-101` receives `sync.broadcast` messages and updates client stores.
- Notification dispatcher: `c3-211` owns JetStream notification dispatch.

Indirect dependents:

- `recipe-realtime-sync` explains that broadcast sync and notifications share NATS but are separate mechanisms.
- The reverse graph for `ref-sync` shows other sync users such as `c3-206`, `c3-207`, `c3-210`, and `c3-212`; they are not the mass-approval action owner, but they prove `ref-sync` is a shared contract and should not be changed casually.

Emergent property: mass approval is not held hostage by notification delivery. `c3-205` explicitly says notifications fire async with error suppression, logged not thrown. That creates a non-blocking boundary: approval mutation and sync can complete even if notification dispatch has an error. `c3-211` can persist/log/retry notification attempts through JetStream and notification_log without throwing the notification failure back into the `approveAll` approval path.

Failure boundary:

- Approval failures belong in the `approved`/`failed` result arrays from `approveAll`, not in notification dispatch.
- Notification failures are logged/suppressed by the PR flow and tracked by `c3-211`; they should not roll back approval state or prevent `ref-sync` broadcast/ack.
- Sync and notification are distinct NATS uses: `ref-sync`/`c3-101` handle broadcast state deltas and acks; `c3-211` handles JetStream-backed notification dispatch.

Concrete change checks:

- Exercise `approveAll` with a mix of approvable and failing PRs and assert the approved/failed arrays are correct.
- Include a step-advance bulk approval and assert `adr-20260202-notification-on-step-advance` behavior: next approvers are notified only for stepAdvanced results.
- Force a notification channel failure and assert approval/sync still complete while the error is logged, not thrown.
- Verify `sync.broadcast` updates browser PR atoms through `c3-101`, and verify notification rows/dispatch attempts through `c3-211`.
