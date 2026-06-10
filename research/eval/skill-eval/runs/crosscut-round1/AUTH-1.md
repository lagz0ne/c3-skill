## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 search "How is authentication handled and what governs it?"
c3 read recipe-auth-and-access --full
c3 read c3-213 --full
c3 read c3-202 --full
c3 read c3-209 --full
c3 read ref-authentication --full
c3 read ref-rbac --full
c3 read ref-nats-jwt-auth --full
c3 graph c3-213 --depth 1
```

## Answer

Authentication is split across `recipe-auth-and-access`, `c3-213`, `c3-202`, `c3-209`, and the external NATS service `c3-4`. `c3-213` owns app login: Google OAuth, test-token auth, and cookie session behavior. `c3-202` carries the authenticated identity as `currentUserTag` / `UserActor`. `c3-209` creates per-session NATS JWT/nkey credentials, and `c3-4` validates JWT signatures and NATS permissions.

## Grounding

`ref-authentication` governs `c3-213` because it defines OAuth, test token, and session behavior. `ref-rbac` governs app access because it defines role permissions and owner checks for flow execution. `ref-nats-jwt-auth` governs the transport identity because NATS WebSocket auth uses JWT resolver credentials, not HTTP cookies.

## Caveats

Fixture has no `rule-*` entities; do not invent auth rules.
