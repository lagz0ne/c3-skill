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

## Cross-cutting criteria

The cross-cut cases add scoreable requirements beyond flat ids:

| ID | Criterion | Score |
| --- | --- | --- |
| X1 | Trace coverage: answer connects the action/command location, sync path, and notification path using the required real ids for that case. | 0 to 3: one point per complete segment. |
| X2 | Sync mechanism named: answer explicitly names NATS WebSocket / `ref-sync`, not generic "realtime". | 0 or 1 |
| X3 | Notification mechanism named: answer explicitly identifies the notification mechanism for the case, such as `c3-211`, NATS JetStream, `sync.user`, or `slackChannel`. | 0 or 1 |
| X4 | Emergent property: answer surfaces the cross-cutting property, such as async/non-blocking notifications, explicit `stepAdvanced`, user-specific notification subjects, or flow entry preserving side effects. | 0 or 1 |

Cross-cut answers should not pass by naming one central entity. They must show
how a user action mutates state, how sync reaches clients, how notifications
reach recipients, and what system property emerges from the combination.

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

### CROSSCUT-MASS-APPROVAL-1

Must trace:

- `c3-105` bulk approval UI in Approvals mode.
- `c3-205` `approveAll` PR Flow and `ref-approval-chain`.
- `ref-sync` plus `c3-101` NATS WebSocket sync.
- `c3-211` plus `adr-20260121-notification-system` notification dispatch.
- Emergent property: notifications are async/non-blocking and errors are logged/suppressed, not thrown into approval/sync.

### CROSSCUT-STEP-ADVANCE-1

Must trace:

- `c3-205` `approvePr` / `approveAll`.
- `ref-approval-chain` `anyof` / `allof`.
- `adr-20260202-notification-on-step-advance` explicit `stepAdvanced`.
- `c3-211`, `ref-pull-dispatcher`, `ref-sync`, and `c3-101`.
- Emergent property: next approvers are notified only when a step actually advances, not on every approval or final completion.

### CROSSCUT-NOTIFICATION-BELL-1

Must trace:

- `c3-205` approval notification trigger.
- `c3-211` notification dispatch.
- `ref-sync` subjects `sync.broadcast` vs `sync.user.{escaped_email}`.
- `c3-101` dual subscription and `adr-20260126-user-notification-ui`.
- Emergent property: sync and notifications share NATS but are separate; notification is durable/targeted while broadcast sync is ephemeral/global.

### CROSSCUT-SLACK-APPROVAL-1

Must trace:

- `c3-215` Slack inbound action and `adr-20260305-slack-bot-integration`.
- `c3-202` execution context concepts for non-HTTP entry.
- `c3-205` PR Flows, not direct service calls.
- `ref-sync` / `c3-101` NATS WebSocket client consistency.
- `c3-211` / `ref-pull-dispatcher` notification channel dispatch.
- Emergent property: entering through flows preserves sync and notification side effects.
