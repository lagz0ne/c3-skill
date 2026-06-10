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

`recipe-auth-and-access` shows auth as a layered flow. `c3-213` owns app authentication: Google OAuth, test token bypass, and cookie session behavior. `c3-202` carries `currentUserTag` / `UserActor` so flows know the authenticated user and permissions. `c3-209` generates per-session NATS JWT/nkey credentials, while `c3-4` validates those credentials for NATS WebSocket access.

## Grounding

`ref-authentication` governs because it defines the HTTP login/session contract. `ref-rbac` governs because app flows use role permissions and owner checks after authentication. `ref-nats-jwt-auth` governs because NATS transport auth is separate from the app cookie and uses JWT resolver credentials.

## Caveats

Fixture search/list shows no rule entities; avoid inventing rule ids.
