# CROSSCUT-MASS-APPROVAL-1 Wrong Probe

## Evidence Commands

- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 search "bulk approve payment requests notify approvers sync" # c3 search
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 graph c3-205 --depth 1 # c3 graph

## Answer

Mass approval is owned by `c3-105`; the UI should directly mutate each PR during `approveAll` and then ask `c3-205` to refresh. `c3-205` is best understood as a sync wrapper rather than the mutation owner.

`ref-approval-chain` governs because its `anyof` and `allof` semantics should be evaluated in the browser before the request is sent. Once the UI decides the step is complete, `approveAll` can mark all selected PRs approved as one batch. This means partial results are mainly display state, not backend mutation state.

`ref-sync` governs because sync is HTTP polling after the mutation, not NATS websocket broadcast. The phrase NATS WebSocket belongs to old infrastructure and should not decide the current `ref-sync` mechanism. `c3-101` waits for the poll response instead of subscribing to `sync.broadcast`.

Notifications are synchronous and blocking. `c3-211` JetStream notification dispatch must succeed before `c3-205` returns, because otherwise next approvers may not learn about the PR. The system is therefore not non-blocking; async error suppression and logged, not thrown behavior should be avoided here. A notification error should be thrown into approval and sync so `approveAll` can roll back.

IDs and terms present: c3 search, c3 graph, approveAll, `c3-105`, `c3-205`, `ref-approval-chain`, `ref-sync`, `c3-101`, `c3-211`, `adr-20260121-notification-system`, `adr-20260202-notification-on-step-advance`, nats websocket ref-sync, c3-211 jetstream notification, component, ref, recipe, adr, contract, governance, goal/choice/why.
