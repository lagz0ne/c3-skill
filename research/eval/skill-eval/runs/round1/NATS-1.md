# NATS-1

## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 list
c3 check
c3 search "What breaks if I change NATS websocket authentication?"
c3 read ref-nats-jwt-auth --full
c3 read c3-209 --full
c3 read c3-4 --full
c3 read ref-sync --full
c3 read recipe-auth-and-access --full
c3 read adr-20260112-nats-auth-callout --full
c3 read adr-20260112-nats-websocket-sync --full
c3 read adr-20260113-nats-jwt-resolver --full
c3 graph ref-nats-jwt-auth --depth 2
c3 graph c3-4 --depth 1
```

## Answer

Changing NATS websocket authentication risks the current JWT resolver path.
`ref-nats-jwt-auth` says the selected design is JWT resolver with memory
preload: the app server signs user JWTs, NATS validates them directly, and
there is no auth callout service. Reintroducing an auth callout or changing the
resolver breaks the selected transport-auth contract.

The implementation owner is `c3-209`. It generates ephemeral per-session JWT
and nkey credentials, limits browser clients to WebSocket, and grants
subscribe-only access to `sync.broadcast` plus the user subject. A change can
break credential generation, JWT expiration, account seed handling, subject
permissions, or loader-to-client credential delivery.

The external service is `c3-4` NATS Server. It validates JWT signatures, checks
expiration, applies embedded permissions, and exposes WSS 8080 to browsers.
Because `c3-4` affects `c3-1` and `c3-2`, a broken auth change can disconnect
frontend subscribers and backend sync/notification behavior.

`ref-sync` also applies because NATS is the real-time sync transport. If auth
or subject permissions change, sync deltas, acks, and user notifications can
stop reaching subscribers.

Historical decisions to inspect are `adr-20260112-nats-auth-callout`,
`adr-20260112-nats-websocket-sync`, and
`adr-20260113-nats-jwt-resolver`.

## Grounding

Search returned the NATS auth ADRs, `c3-4`, `c3-209`, `ref-nats-jwt-auth`, and
`ref-sync`. The `ref-nats-jwt-auth` Goal/Choice/Why explain why JWT resolver is
the current design. Component contract rows on `c3-209` and container wiring on
`c3-4` identify the credential owner and validator.

## Caveats

No `rule-*` entities exist in this fixture. ADRs are historical context; the
current governing docs are `ref-nats-jwt-auth`, `c3-209`, `c3-4`, and
`ref-sync`.
