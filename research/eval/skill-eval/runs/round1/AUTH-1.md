# AUTH-1

## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 list
c3 check
c3 search "How is authentication handled and what governs it?"
c3 read recipe-auth-and-access --full
c3 read c3-213 --full
c3 read c3-202 --full
c3 read c3-209 --full
c3 read c3-4 --full
c3 read ref-authentication --full
c3 read ref-rbac --full
c3 read ref-nats-jwt-auth --full
c3 graph c3-213 --depth 1
c3 graph ref-nats-jwt-auth --depth 2
```

## Answer

Authentication is split across HTTP session auth, application authorization,
and NATS transport auth. The orientation doc is `recipe-auth-and-access`, which
states that auth and access span cookies, RBAC, and NATS credentials.

`c3-213` Authentication Flows owns app login. It handles Google OAuth for
production and test token auth for E2E, then returns a `UserActor`; the route
handler stores a Base64 email in an HttpOnly cookie. `ref-authentication`
governs this because its Goal/Choice/Why define Google OAuth, test tokens,
pre-existing users, and cookie session behavior.

`c3-202` Execution Context carries identity through backend execution via
`currentUserTag` and `UserActor`. RBAC is governed by `ref-rbac`: permissions
are JSON roles with inheritance, owner checks go through `rbacQueries.isOwner`,
and security mutations are logged.

NATS WebSocket auth is a separate transport identity layer. `c3-209` NATS
Credential Generator creates per-session JWT plus nkey credentials, while
`c3-4` NATS Server validates JWT signatures, expiration, and subject
permissions. `ref-nats-jwt-auth` governs that layer because it selects the JWT
resolver pattern with memory preload and no external auth callout service.

## Grounding

The recipe names the three layers. Component contracts on `c3-213`, `c3-202`,
and `c3-209` say changes must preserve Goal, Governance, Contract, and Change
Safety. The graph for `c3-213` shows `ref-authentication` and `ref-rbac`; the
graph for `ref-nats-jwt-auth` shows it cited by `c3-209`,
`recipe-auth-and-access`, and `adr-20260113-nats-jwt-resolver`.

## Caveats

The fixture has no `rule-*` entities, so rules should not be invented. `c3
check` rebuilds cache but reports known canonical markdown drift on `c3-0`;
this did not block targeted reads.
