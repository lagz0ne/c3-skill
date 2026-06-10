## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 search "What breaks if I change NATS websocket authentication?"
c3 read ref-nats-jwt-auth --full
c3 read c3-209 --full
c3 read c3-4 --full
c3 read recipe-auth-and-access --full
c3 read ref-sync --full
c3 read adr-20260112-nats-auth-callout --full
c3 read adr-20260113-nats-jwt-resolver --full
c3 graph ref-nats-jwt-auth --depth 2
```

## Answer

Changing NATS WebSocket authentication risks `ref-nats-jwt-auth`, `c3-209`, `c3-4`, and `ref-sync`. Current design is no auth callout: `c3-209` issues ephemeral browser credentials, then `c3-4` validates JWT signatures, expiration, and subject permissions. `ref-sync` rides on that transport, so credential, prefix, or permission drift can break browser subscriptions and server publish behavior.

## Grounding

`ref-nats-jwt-auth` governs because it selects JWT resolver authentication and rejects auth callout. Historical context is `adr-20260112-nats-auth-callout` then `adr-20260113-nats-jwt-resolver`. `ref-sync` governs because sync messages depend on the authenticated NATS WebSocket transport.

## Caveats

ADRs are history; current state must be verified against refs and components.
