## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 search "notification system NATS websocket sync async approval"
c3 read c3-205 --full
c3 read c3-211 --full
c3 read ref-sync --full
c3 read c3-101 --full
c3 read adr-20260126-user-notification-ui --full
c3 graph ref-sync --depth 1
c3 graph c3-211 --depth 1
```

## Answer

Do not start at the bell alone. Trace from `c3-205`, because PR approval flows trigger notificationService. Then check `c3-211`, the Notification System, because it publishes and dispatches notification messages. Then check `ref-sync`, because it defines the NATS WebSocket subjects used by clients.

On the frontend, `c3-101` owns `natsSync`: `sync.broadcast` updates shared stores and `sync.user.<email>` carries notifications into the notifications atom. `adr-20260126-user-notification-ui` exists because notifications could be published but invisible when the client listened only to broadcast.

## Grounding

`ref-sync` governs because it defines the subject split and the NATS WebSocket delivery path. `c3-211` owns notification dispatch; `c3-101` owns client receipt.

## Caveats

This traces C3 docs only. Runtime diagnosis would also inspect NATS subscriptions and browser state.
