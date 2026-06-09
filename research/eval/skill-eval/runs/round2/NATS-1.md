# NATS-1

## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
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

Changing NATS websocket authentication can break the selected JWT resolver
design, credential generation, server validation, and real-time sync.

`ref-nats-jwt-auth` is the governing ref. Its Choice says JWT resolver with
memory preload: the app signs user JWTs, NATS validates signatures directly,
and there is no auth callout service. Replacing that with callout behavior, or
changing resolver keys, breaks the current contract.

`c3-209` owns credentials. It generates ephemeral per-session JWT plus nkey
credentials, returns them to the loader/client, restricts browser connections
to WebSocket, and grants subscribe-only access to the broadcast and user
subjects. Risks include broken account seed use, invalid JWT signatures,
expiry mistakes, subject permission drift, and missing client credentials.

`c3-4` owns the external NATS validation surface. It validates JWT signatures,
checks expiration, applies subject permissions, and affects `c3-1` and `c3-2`.
`ref-sync` applies because frontend subscribers and backend mutation broadcasts
depend on that NATS transport.

Historical context to inspect: `adr-20260112-nats-auth-callout`,
`adr-20260112-nats-websocket-sync`, and
`adr-20260113-nats-jwt-resolver`.

## Grounding

Search returned the three NATS auth/sync ADRs plus `c3-4`, `c3-209`,
`ref-nats-jwt-auth`, and `ref-sync`. The `ref-nats-jwt-auth` read provides the
Goal/Choice/Why for the current JWT resolver design. The graphs show
`ref-nats-jwt-auth` cited by `c3-209` and the auth recipe, and `c3-4`
affecting frontend/backend containers.

## Caveats

No `rule-*` entities exist in this fixture. ADRs here are historical; current
breakage analysis should anchor on `ref-nats-jwt-auth`, `c3-209`, `c3-4`, and
`ref-sync`.
