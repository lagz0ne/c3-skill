# AUTH-1 Gold Probe

## Evidence Commands

- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 search "How is authentication handled and what governs it?" # c3 search
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read recipe-auth-and-access --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read c3-213 --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read c3-202 --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read c3-209 --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read c3-4 --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read ref-authentication --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read ref-rbac --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read ref-nats-jwt-auth --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 graph c3-213 --depth 1
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 graph ref-nats-jwt-auth --depth 2 --direction reverse

## Answer

Authentication is layered, not a single gate. The top-level recipe, `recipe-auth-and-access`, states that access spans HTTP session cookies, application RBAC, and separate NATS transport auth. The action owner for interactive app login is `c3-213`: it owns Google OAuth and test token authentication, normalizes the email, calls `initUserActor`, and returns a `UserActor` or an explicit failure such as `INVALID_TOKEN` or `USER_NOT_FOUND`.

The first state boundary is the HTTP session. `ref-authentication` governs this because its Goal/Choice/Why define Google OAuth for production, a test token path for E2E, an HttpOnly Base64 email cookie, and the negative rule that OAuth alone does not grant access unless the user already exists in the database. `c3-213` confirms the same route-level cookie behavior in its Cookie Management section.

The second state boundary is request context and permission evaluation. `c3-202` defines `currentUserTag`, `UserActor`, `transactionTag`, and `executionIdTag`; the recipe says flows consume `currentUserTag` through `seekTag()`. `ref-rbac` governs this because it defines JSON role permissions, parent inheritance, the owner check through `rbacQueries.isOwner`, and security event logging for RBAC mutations. The dependent behavior is that backend flows should authorize through the `UserActor` and RBAC queries instead of re-interpreting the cookie.

The third boundary is NATS identity. `c3-209` owns generation of per-session JWT plus nkey credentials and passes them through loaderData to the client. `ref-nats-jwt-auth` governs this because it chooses the JWT resolver pattern, subscribe-only WebSocket credentials, one-hour expiry, and no auth callout service. The direct observer is `c3-4`, which validates JWT signatures and expiration, enforces subject permissions, and accepts browser WSS connections only within the embedded permissions.

Direct dependents: `c3-213` is the HTTP auth owner, `c3-202` is the execution-context carrier, `c3-209` is the NATS credential owner, and `c3-4` is the external broker enforcement boundary. Indirect dependents are the backend flows that read `currentUserTag` and use RBAC, plus browser sync code that depends on `c3-209` credentials being accepted by `c3-4`. The graph for `c3-213` ties it to `ref-authentication` and `ref-rbac`; the reverse graph for `ref-nats-jwt-auth` ties the transport-auth ref to `c3-209`, `recipe-auth-and-access`, and historical `adr-20260113-nats-jwt-resolver`.

Emergent property: app authentication, app authorization, and NATS transport identity fail independently. A valid Google account can still fail app login if no user row exists; a valid cookie can still fail a mutation if `UserActor.asserts()` or an owner check rejects it; a valid app session can still lose real-time transport if the JWT resolver, expiry, or subject permissions fail.

Historical ADR context: `adr-20260113-nats-jwt-resolver` is historical evidence from the `ref-nats-jwt-auth` reverse graph for why the current transport mechanism is JWT resolver based. The search also surfaced older NATS auth-callout context, but the current ref explicitly says no auth callout service is needed.

Caveats and change checks:

- There are no fixture rule entities in this auth answer; do not invent auth, RBAC, or NATS rule IDs.
- For login changes, check both Google OAuth and test token paths, then confirm the cookie remains HttpOnly and user-existence checks still reject unknown users.
- For authorization changes, test an allowed role, a denied role, and an owner-only path through `UserActor`/`ref-rbac`.
- For NATS auth changes, test that `c3-209` still issues subscribe-only credentials and `c3-4` still rejects expired or over-permissioned JWTs.
