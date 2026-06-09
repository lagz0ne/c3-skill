## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 search "notification system NATS websocket sync async approval"
c3 read c3-205 --full
c3 read c3-211 --full
c3 read ref-sync --full
c3 read c3-101 --full
c3 read adr-20260126-user-notification-ui --full
c3 read recipe-realtime-sync --full
c3 graph ref-sync --depth 1
c3 graph c3-211 --depth 1
```

## Answer

Start before UI. `c3-205` triggers notifications from approval flows. `c3-211` publishes and dispatches notification messages. `ref-sync` defines NATS WebSocket subject behavior: broadcast messages carry deltas/acks, and user-specific subjects carry notification messages. `c3-101` must subscribe to both and route notification messages into the notifications atom.

`adr-20260126-user-notification-ui` explains the visibility bug: notifications were published but users could not see them while the frontend listened only to broadcast. `recipe-realtime-sync` adds the cross-cut property: sync and notifications share NATS but are separate; notification is durable and targeted, broadcast sync is ephemeral/global.

## Grounding

`ref-sync` governs because its subject split is the contract that separates data sync from notification delivery.

## Caveats

Runtime debugging would still check actual NATS subscription state.
