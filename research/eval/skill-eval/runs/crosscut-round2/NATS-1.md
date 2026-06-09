## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 search "What breaks if I change NATS websocket authentication?"
c3 read ref-nats-jwt-auth --full
c3 read c3-209 --full
c3 read c3-4 --full
c3 read ref-sync --full
c3 read adr-20260112-nats-auth-callout --full
c3 read adr-20260113-nats-jwt-resolver --full
c3 graph ref-nats-jwt-auth --depth 2
```

## Answer

Current NATS WebSocket authentication is `ref-nats-jwt-auth`: JWT resolver, not auth callout. `c3-209` creates browser credentials, and `c3-4` validates signatures, expiry, and subject permissions. Changing it can break browser connect, subscribe permissions, credential loading, and sync delivery through `ref-sync`. The historical path is `adr-20260112-nats-auth-callout` -> `adr-20260113-nats-jwt-resolver`.

## Grounding

`ref-nats-jwt-auth` governs because it is the selected auth mechanism and explicitly says no auth callout. `ref-sync` governs because NATS WebSocket is the transport for deltas, acks, and user-specific messages.

## Caveats

ADRs are historical; current refs/components are the active contract.
