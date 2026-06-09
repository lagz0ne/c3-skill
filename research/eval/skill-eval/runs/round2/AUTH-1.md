# AUTH-1

## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 search "How is authentication handled and what governs it?"
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

Acountee authentication is layered: HTTP session authentication, application
RBAC authorization, and separate NATS transport authentication.

`recipe-auth-and-access` is the trace doc. It says authentication and
authorization span cookies, RBAC, and NATS credentials. `c3-213`
Authentication Flows owns the login layer: Google OAuth for production, test
token auth for E2E, user lookup through `initUserActor`, and session restore
through a cookie. `ref-authentication` governs this because its
Goal/Choice/Why choose Google OAuth, test tokens, pre-existing users, and
cookie session behavior.

`c3-202` Execution Context carries the authenticated identity as
`currentUserTag` / `UserActor`. `ref-rbac` governs authorization: role
permissions, inherited permissions, owner checks with `rbacQueries.isOwner`,
and security-event logging.

NATS auth is not the HTTP session. `c3-209` NATS Credential Generator creates
per-session JWT plus nkey credentials, and `c3-4` NATS Server validates JWT
signatures, expiration, and permissions. `ref-nats-jwt-auth` governs that
transport layer because it chooses the JWT resolver pattern with memory preload
and no external auth callout service.

## Grounding

Search surfaced `c3-209`, `c3-213`, `c3-4`, `ref-authentication`, and
`ref-nats-jwt-auth`. Targeted reads provided the component contracts and each
ref's Goal/Choice/Why. The `c3-213` graph shows `ref-authentication` and
`ref-rbac`; the `ref-nats-jwt-auth` graph shows `c3-209` and
`recipe-auth-and-access`.

## Caveats

The first cold search hit a missing fixture cache; `c3 check` rebuilt it but
reported known `c3-0` drift. No `rule-*` entities exist in this fixture.
