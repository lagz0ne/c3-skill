## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 search "NATS JWT auth sync layer coupling WebSocket"
c3 read ref-nats-jwt-auth --full
c3 graph ref-nats-jwt-auth --depth 2 --direction reverse
c3 read c3-209 --full
c3 read c3-4 --full
c3 read c3-101 --full
c3 read ref-sync --full
c3 graph c3-209 --depth 1
c3 graph c3-101 --depth 1
```

## Answer

NATS JWT auth affects sync because browser sync depends on a successful authenticated NATS WebSocket connection. `ref-nats-jwt-auth` defines the JWT resolver design and says there is no auth callout service. `c3-209` generates the per-session JWT and nkey for the browser. `c3-4` validates the JWT and enforces subject permissions. `c3-101` then uses those credentials for `natsSync` and subscribes to `sync.broadcast` and `sync.user.<email>`.

If JWT signing, resolver config, expiry, or permissions change, the HTTP app may still authenticate the user, but browser sync can stop receiving deltas and acks. `ref-sync` is the sync contract: deltas and ack messages travel over NATS WebSocket subjects and client waiting depends on matching those messages.

## Grounding

`ref-nats-jwt-auth` governs the transport identity. `ref-sync` governs the messages that the transport must deliver to clients.

## Caveats

The fixture has known `c3 check` drift. I did not inspect source code beyond C3 docs.
