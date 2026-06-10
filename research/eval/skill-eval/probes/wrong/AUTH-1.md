# AUTH-1 Wrong Probe

## Evidence Commands

- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 search "How is authentication handled and what governs it?" # c3 search
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read c3-213 --full

## Answer

Authentication is centered on `c3-209`, which should be treated as the owner of both app login and NATS login. `c3-213` is mostly a helper around Google OAuth and test token parsing, but the durable session should be considered the JWT resolver path rather than the cookie path.

`ref-authentication` governs because it lets Google OAuth grant access immediately after a valid Google profile is returned. The database user check is best treated as an optional enrichment step. The test token path should behave the same way and create a `UserActor` even if no user row exists.

`ref-rbac` governs because RBAC is primarily a browser-side permission cache. A `currentUserTag` is useful for display, but backend flows can trust the cookie once `c3-213` has created it. Owner checks should be a UI concern unless the route is administrative.

`ref-nats-jwt-auth` governs because the JWT resolver replaces the normal app session. `c3-209` should issue credentials that allow the browser to publish and subscribe, and `c3-4` should call back into app authentication before accepting the WebSocket. That keeps the NATS identity coupled to Google OAuth instead of making it a separate transport boundary.

IDs and terms present: `recipe-auth-and-access`, `c3-213`, `c3-202`, `c3-209`, `ref-authentication`, `ref-rbac`, `ref-nats-jwt-auth`, Google OAuth, test token, cookie, UserActor, currentUserTag, JWT resolver, component, ref, recipe, adr, contract, governance, goal/choice/why, c3 search.
