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
c3 read recipe-realtime-sync --full
c3 read recipe-auth-and-access --full
c3 read adr-20260113-nats-jwt-resolver --full
c3 read adr-20260112-nats-auth-callout --full
c3 graph c3-209 --depth 1
c3 graph c3-101 --depth 1
```

## Answer

Emergent property: **transport-auth/sync coupling**. Browser sync is gated by NATS transport identity, so changing JWT resolver behavior can break sync even when HTTP application auth still works.

Trace:

| Segment | Entities | Mechanism |
| --- | --- | --- |
| Credential generation | `ref-nats-jwt-auth`, `c3-209`, `adr-20260113-nats-jwt-resolver` | JWT resolver, no auth callout, per-session JWT + nkey generated at page load and passed through loaderData. |
| Broker enforcement | `c3-4`, `adr-20260112-nats-auth-callout` | NATS validates signature/expiration and applies subscribe-only permissions for browser WebSocket clients. |
| Sync runtime | `c3-101`, `ref-sync`, `recipe-realtime-sync`, `recipe-auth-and-access` | `natsSync` connects over WebSocket, subscribes to `sync.broadcast` and user subjects, and uses executionId to resolve deltas/acks. |

If signing keys, resolver preload, token expiration, or subject permissions change, the client may still have an app session but fail to establish the NATS WebSocket or receive `sync.broadcast` deltas. `ref-nats-jwt-auth` governs the credential and permission layer; `ref-sync` governs the sync delta/ack contract that depends on that transport.

## Grounding

`ref-nats-jwt-auth` governs because it defines JWT resolver auth, loaderData credentials, and subscribe-only browser access. `ref-sync` governs because its WebSocket delivery contract requires those credentials and exact subject permissions.

## Caveats

The fixture has known `c3 check` drift. No `rule-*` entities found in the fixture.
