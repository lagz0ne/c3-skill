# Skill-effectiveness rubric

This rubric is derived from the `AUTH-1` sample answer inspection in
`cases/acountee-round1.md`. It is intentionally concrete and scoreable. Use the
candidate answer plus its `Evidence commands` section.

## Universal criteria

| ID | Criterion | Score |
| --- | --- | --- |
| U1 | Local C3 only: evidence uses `C3X_MODE=agent bash skills/c3/bin/c3x.sh` or a clearly equivalent local `c3` function bound to `skills/c3/bin/c3x.sh`; no bare/global `c3x`. | 0 or 1 |
| U2 | Search-first for conceptual discovery: first C3 discovery command for the question is `c3 search "<question or close paraphrase>"`, not `list` plus title matching. | 0 or 1 |
| U3 | Targeted confirmation: after search, answer evidence includes at least one targeted `read`, `graph`, `lookup`, or `schema` command relevant to the cited ids. | 0 or 1 |
| U4 | Exact ids: answer names the required entity ids for the case with exact tokens. | 0 to 3: 0 none, 1 some, 2 most, 3 all required core ids. |
| U5 | Governance with why: cited `ref-*` ids are paired with why they govern the answer, not just listed. | 0 to 3: 0 absent, 1 ids only, 2 some why, 3 all core refs have why. |
| U6 | Canvas contract awareness: answer respects relevant component/ref/recipe/ADR contracts, especially component `Governance`/`Contract` and ref `Goal/Choice/Why`; it does not treat refs/rules as interchangeable. | 0 or 1 |
| U7 | No hallucinated governance: answer does not invent `rule-*`, `ref-*`, ADR, or component ids absent from the fixture. | 0 or 1 |
| U8 | Caveat handling: answer notes material fixture limits when relevant, such as no `rule-*` entities or known acountee check drift. | 0 or 1 |

Suggested pass bar for round 1: `U1=1`, `U2=1`, `U3=1`, `U4>=2`, `U5>=2`,
`U7=1`, plus the case-specific must-have ids below.

## Case-specific bars

### AUTH-1

Must-have ids:

- `recipe-auth-and-access`
- `c3-213`
- `c3-202`
- `c3-209`
- `ref-authentication`
- `ref-rbac`
- `ref-nats-jwt-auth`

Strong answer also names `c3-4` for NATS validation.

Must explain:

- Google OAuth/test-token login and cookie session are app auth.
- `UserActor` / `currentUserTag` carries authenticated identity.
- RBAC governs permissions and owner checks.
- NATS JWT auth is separate transport auth and has no external auth callout service.

### NATS-1

Must-have ids:

- `ref-nats-jwt-auth`
- `c3-209`
- `c3-4`
- `ref-sync`
- `adr-20260112-nats-auth-callout`
- `adr-20260113-nats-jwt-resolver`

Must explain:

- Current design is JWT resolver, not auth callout.
- Changing websocket auth can break credential generation, JWT resolver config, expiration, subject permissions, and sync subscribers.
- `c3-1` and `c3-2` are affected through the external NATS service `c3-4`.

### ADMIN-1

Must-have ids:

- `c3-107`
- `c3-210`
- `adr-20260121-admin-management-features`
- `ref-admin-page-layout`
- `ref-rbac`

Strong answer also names `ref-sync` and `recipe-screen-anatomy` or
`recipe-navigation-strategy`.

Must explain:

- `c3-107` owns frontend admin screens.
- `c3-210` owns backend admin flows.
- Owner-only access is enforced with RBAC.
- The admin ADR affects `c3-1`, `c3-2`, and `c3-204`.

### APPROVAL-1

Must-have ids:

- `recipe-approval-workflow`
- `c3-205`
- `c3-212`
- `c3-105`
- `ref-approval-chain`
- `ref-audit-trail`
- `ref-sync`

Must explain:

- PR lifecycle is `draft -> pending -> approved -> completed`.
- Approval semantics use `anyof` and `allof`.
- `c3-205` owns core PR mutations; `c3-212` extends approved PR workbench operations; `c3-105` owns the screen interaction.
- Audit and sync are cross-cutting contracts for mutations.

### UI-1

Must-have ids:

- `recipe-screen-anatomy`
- `c3-104`
- `c3-105`
- `ref-master-detail-layout`
- `ref-detail-content-strategy`
- `ref-list-view-patterns`
- `ref-filter-footer`
- `ref-responsive-layout`

Must explain:

- Invoice and payment-request screens are Master-Detail screens.
- Detail panes follow facet/BIG-grid section conventions.
- Lists are virtualized and filtered via the footer pattern.
- Responsive behavior belongs to the shared layout/ref, not per-screen custom implementations.
