# Acountee round 1 cases

Ground truth was established with the local C3 wrapper:

```bash
C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 <cmd>
```

The fixture `list` output reports `totalCount: 66`. It contains no `rule-*`
entities, so rule expectations below are explicit negatives: do not hallucinate
rules.

The relevant canvas contracts are:

- `schema component`: required `Goal`, `Parent Fit`, `Purpose`, `Foundational Flow`, `Business Flow`, `Governance`, `Contract`, `Change Safety`, `Derived Materials`.
- `schema ref`: required `Goal`, `Choice`, `Why`; optional `How`; reject if `Why` merely restates `Choice`.
- `schema recipe`: required `Goal`.
- `schema adr`: required compliance tables include `Affected Topology`, `Compliance Refs`, `Compliance Rules`, and `Verification`.

## AUTH-1: authentication and access

Question: How is authentication handled and what governs it?

Grounding commands:

```bash
c3 search "How is authentication handled and what governs it?"
c3 read recipe-auth-and-access --full
c3 read c3-213 --full
c3 read c3-202 --full
c3 read c3-209 --full
c3 read ref-authentication --full
c3 read ref-rbac --full
c3 read ref-nats-jwt-auth --full
c3 graph c3-213 --depth 1
c3 graph ref-nats-jwt-auth --depth 2
```

Expected elements:

| Kind | Expected |
| --- | --- |
| Primary recipe | `recipe-auth-and-access` traces login, RBAC, execution context, and NATS auth. |
| Main app auth component | `c3-213` Authentication Flows owns Google OAuth and test-token authentication. |
| Execution context | `c3-202` carries `currentUserTag` / `UserActor` and is part of the auth/access path. |
| NATS auth component | `c3-209` generates per-session JWT + nkey credentials. |
| External service | `c3-4` validates JWTs and enforces NATS WebSocket permissions. |
| Governing refs | `ref-authentication`, `ref-rbac`, `ref-nats-jwt-auth`. |
| Rules | None in fixture. Do not invent `rule-*` ids. |
| Canvas contracts | Component `Governance` and `Contract` rows matter for `c3-213`, `c3-202`, `c3-209`; ref `Goal/Choice/Why` explains why the auth refs govern. |

Why the refs apply:

- `ref-authentication`: governs Google OAuth, test-token bypass, and cookie session behavior.
- `ref-rbac`: governs role permissions, owner checks, inherited permissions, and security events.
- `ref-nats-jwt-auth`: governs the separate NATS transport identity layer.

Evidence snippets:

```text
search returned c3-209 ... uses:ref-nats-jwt-auth
search returned c3-213 ... uses:ref-authentication, ref-rbac
read recipe-auth-and-access: Authentication and authorization span three layers: HTTP session, application RBAC, and NATS transport auth.
read c3-213: uses: ref-authentication,ref-pumped-fn,ref-rbac,ref-structured-logging
read c3-202: currentUserTag | UserActor | Authenticated user with permissions
read c3-209: goal: Generate per-session JWT + nkey for NATS WebSocket authentication
read ref-nats-jwt-auth: no auth callout service needed
graph c3-213: ref-rbac cites c3-205, c3-210, c3-213, recipe-audit-and-compliance, recipe-auth-and-access
```

## NATS-1: NATS websocket auth impact

Question: What breaks if I change NATS websocket authentication?

Grounding commands:

```bash
c3 search "What breaks if I change NATS websocket authentication?"
c3 read ref-nats-jwt-auth --full
c3 read c3-209 --full
c3 read c3-4 --full
c3 read recipe-auth-and-access --full
c3 graph ref-nats-jwt-auth --depth 2
c3 graph c3-4 --depth 1
```

Expected elements:

| Kind | Expected |
| --- | --- |
| Current auth pattern | `ref-nats-jwt-auth` says JWT resolver, memory preload, no auth callout service. |
| Credential owner | `c3-209` generates ephemeral per-session JWT + nkey credentials. |
| External service | `c3-4` validates JWT signatures, expiration, and permissions. |
| Historical ADRs | `adr-20260112-nats-auth-callout`, `adr-20260112-nats-websocket-sync`, `adr-20260113-nats-jwt-resolver`. |
| Related sync ref | `ref-sync` because NATS is the real-time sync transport. |
| Affected topology | `c3-0`, `c3-1`, `c3-2`, `c3-4`; plus `c3-209` implementation owner. |
| Rules | None in fixture. |
| Canvas contracts | ADR compliance rows must say why affected topology and refs apply; component contract rows matter for `c3-209`; ref `Goal/Choice/Why` matters for `ref-nats-jwt-auth`. |

Why the refs apply:

- `ref-nats-jwt-auth`: current selected mechanism. Changing websocket auth risks credential generation, resolver config, permissions, and expiry behavior.
- `ref-sync`: websocket auth protects the transport used for sync broadcasts and user-specific subjects.

Evidence snippets:

```text
search returned adr-20260112-nats-auth-callout, adr-20260112-nats-websocket-sync, adr-20260113-nats-jwt-resolver
read ref-nats-jwt-auth: NATS validates signatures directly without an auth callout service
read c3-209: loader passes credentials to the client via loaderData
read c3-4: NATS | Account JWT resolver | Internal | Validate user JWT signatures + enforce subject permissions
graph ref-nats-jwt-auth: cited by adr-20260113-nats-jwt-resolver, c3-209, recipe-auth-and-access
graph c3-4: affected by adr-20260112-nats-auth-callout and adr-20260121-notification-system; affects c3-1,c3-2
```

## ADMIN-1: administrator feature ownership

Question: What owns administrator features for users, teams, audit, and approval configuration?

Grounding commands:

```bash
c3 search "What owns administrator features for users teams audit and approval configuration?"
c3 read c3-107 --full
c3 read c3-210 --full
c3 read adr-20260121-admin-management-features --full
c3 read ref-admin-page-layout --full
c3 read ref-rbac --full
c3 graph c3-107 --depth 1
c3 graph c3-210 --depth 1
```

Expected elements:

| Kind | Expected |
| --- | --- |
| Frontend owner | `c3-107` Admin Screens owns users, teams, audit logs, notification logs, approval config, organization view. |
| Backend owner | `c3-210` Admin Flows owns teams, roles, users, and approval configuration flows. |
| Decision/history | `adr-20260121-admin-management-features` implemented the admin management feature set. |
| Supporting docs | `recipe-navigation-strategy` and `recipe-screen-anatomy` help route/screen discovery. |
| Governing frontend refs | `ref-admin-page-layout`, `ref-form-patterns`, `ref-org-tiles`, `ref-responsive-layout`, `ref-sft-behavioral-spec`, `ref-ui-patterns`, `ref-variant-system`. |
| Governing backend refs | `ref-pumped-fn`, `ref-query-services`, `ref-rbac`, `ref-server-functions`, `ref-structured-logging`, `ref-sync`. |
| Rules | None in fixture. |
| Canvas contracts | Component `Governance`, `Contract`, and `Change Safety` rows define the owner boundaries; ADR compliance rows should cite `c3-1`, `c3-2`, `c3-204` and explain why. |

Why the refs apply:

- `ref-admin-page-layout`: governs the admin screen structure where full-page admin table patterns apply.
- `ref-rbac`: governs owner-only access for admin flows and role/user permission checks.
- `ref-sync`: governs mutation broadcasts for admin flow changes.

Evidence snippets:

```text
search returned c3-107 Admin Screens and c3-210 Admin Flows
read c3-107: All screens require owner role -- server functions enforce via rbacQueries.isOwner
read c3-210: Every admin flow: extract user from currentUserTag -> check rbacQueries.isOwner -> reject with NOT_OWNER
read adr-20260121-admin-management-features: affects: c3-1,c3-2,c3-204
graph c3-107: uses ref-admin-page-layout, ref-form-patterns, ref-org-tiles, ref-responsive-layout, ref-sft-behavioral-spec, ref-ui-patterns, ref-variant-system
graph c3-210: uses ref-pumped-fn, ref-query-services, ref-rbac, ref-server-functions, ref-structured-logging, ref-sync
```

## APPROVAL-1: approval workflow ownership

Question: Where does approval workflow live and what governs changes to approvals?

Grounding commands:

```bash
c3 search "Where does approval workflow live and what governs changes to approvals?"
c3 read recipe-approval-workflow --full
c3 read c3-205 --full
c3 read c3-212 --full
c3 read ref-approval-chain --full
c3 graph c3-205 --depth 1
c3 graph c3-212 --depth 1
c3 graph ref-approval-chain --depth 1
```

Expected elements:

| Kind | Expected |
| --- | --- |
| Primary recipe | `recipe-approval-workflow` traces PR creation through approval and workbench bulk operations. |
| Core backend owner | `c3-205` PR Flows owns create, request approvals, approve, reject, recall, complete, and bulk approve. |
| Workbench extension | `c3-212` Workbench Flows owns invoice cleanup, approved PR export, and paid PR import. |
| Frontend participant | `c3-105` PaymentRequestsScreen owns PR and approvals UI modes. |
| Governing refs | `ref-approval-chain`, `ref-audit-trail`, `ref-sync`, plus backend flow refs such as `ref-rbac`, `ref-server-functions`, `ref-query-services`. |
| Related notification owner | `c3-211` is mentioned by the recipe for next-step approver notifications. |
| Rules | None in fixture. |
| Canvas contracts | Component contracts separate UI (`c3-105`), core flows (`c3-205`), and workbench flows (`c3-212`); ref `Goal/Choice/Why` explains approval semantics. |

Why the refs apply:

- `ref-approval-chain`: governs multi-step approval chains, `anyof`/`allof`, step advancement, records, and state machine.
- `ref-audit-trail`: approval mutations are audit captured; the recipe says not to double-write audit entries for PR mutations.
- `ref-sync`: every mutation emits a sync delta and ack.

Evidence snippets:

```text
list: recipe-approval-workflow sources c3-0,c3-205,c3-212,ref-approval-chain,ref-audit-trail,ref-sync
read recipe-approval-workflow: draft -> pending -> approved -> completed
read c3-205: approvePr records current user's approval; if step advances, notifies next approvers
read c3-212: importPaidPrs validates approved PR, sets payment_reference, completes PR
read ref-approval-chain: modes anyof/allof; mode validation is app-level logic in prService.approve
graph ref-approval-chain: cited by c3-205 and recipe-approval-workflow
```

## UI-1: detail and list layout consistency

Question: How should invoice and payment request screens stay consistent across detail and list layouts?

Grounding commands:

```bash
c3 search "How should invoice and payment request screens stay consistent across detail and list layouts?"
c3 read recipe-screen-anatomy --full
c3 read c3-104 --full
c3 read c3-105 --full
c3 read ref-master-detail-layout --full
c3 read ref-detail-content-strategy --full
c3 read ref-list-view-patterns --full
c3 read ref-filter-footer --full
c3 read ref-responsive-layout --full
c3 graph ref-master-detail-layout --depth 1
c3 graph ref-detail-content-strategy --depth 1
```

Expected elements:

| Kind | Expected |
| --- | --- |
| Orientation recipe | `recipe-screen-anatomy` identifies Invoices and Payment Requests as Master-Detail screens. |
| Screen owners | `c3-104` InvoiceScreen and `c3-105` PaymentRequestsScreen. |
| Governing refs | `ref-master-detail-layout`, `ref-detail-content-strategy`, `ref-list-view-patterns`, `ref-filter-footer`, `ref-responsive-layout`, plus shared UI refs from the components. |
| Rules | None in fixture. |
| Canvas contracts | Component `Governance`/`Contract` rows preserve ownership; ref `Goal/Choice/Why` explains why each UI pattern applies. |

Why the refs apply:

- `ref-master-detail-layout`: both screens use list plus detail and should let the layout handle desktop/tablet/mobile behavior.
- `ref-detail-content-strategy`: detail panes use facet grids, BIG grids, section ordering, and empty-state conventions.
- `ref-list-view-patterns`: both feature screens use virtualized lists and sticky group headers where appropriate.
- `ref-filter-footer`: both screens use filter UI in the list footer.
- `ref-responsive-layout`: defines the mobile/tablet/desktop breakpoints used by MasterDetailLayout.

Evidence snippets:

```text
read recipe-screen-anatomy: Invoices -> c3-104; Payment Requests -> c3-105; both Master-Detail
read c3-104: Layout: Master-detail with FilterFooter. Virtualized list on left, tabbed detail pane on right.
read c3-105: Dual mode PRs/Approvals; MasterDetailLayout, FilterFooter, PRFilterContent, AuditLogPanel
graph ref-master-detail-layout: cited by c3-104, c3-105, recipe-navigation-strategy, recipe-responsive-design, recipe-screen-anatomy
graph ref-detail-content-strategy: cited by c3-104, c3-105, recipe-screen-anatomy
```

## Sample skill-guided answer

Case: `AUTH-1`.

Evidence commands used, in order:

```bash
c3 search "How is authentication handled and what governs it?"
c3 read recipe-auth-and-access --full
c3 read c3-213 --full
c3 read c3-202 --full
c3 read c3-209 --full
c3 read ref-authentication --full
c3 read ref-rbac --full
c3 read ref-nats-jwt-auth --full
c3 graph c3-213 --depth 1
c3 graph ref-nats-jwt-auth --depth 2
```

Answer:

Acountee splits auth into three related layers.

`c3-213` Authentication Flows owns application login. It implements Google
OAuth for production and test-token auth for E2E, then returns a `UserActor`;
the route handler stores the session as an HttpOnly `user` cookie. This is
governed by `ref-authentication`, because that ref chooses Google OAuth/test
tokens, cookie sessions, and the pre-existing-user requirement.

`c3-202` Execution Context carries the authenticated actor through the backend
via `currentUserTag`. Authorization is governed by `ref-rbac`: roles carry JSON
permissions, owner checks use `rbacQueries.isOwner`, permissions can inherit
from parent roles, and security mutations are logged.

NATS websocket auth is separate from HTTP auth. `c3-209` NATS Credential
Generator creates a per-session JWT plus nkey, and `c3-4` NATS Server validates
the JWT signature, expiration, and permissions. That path is governed by
`ref-nats-jwt-auth`, which chooses a JWT resolver with memory preload and says
there is no external auth callout service. The recipe `recipe-auth-and-access`
is the cross-cutting trace tying these layers together.

No `rule-*` entities exist in this fixture, so I would cite refs and component
contracts, not rules.

Inspection against ground truth:

| Result | Notes |
| --- | --- |
| Got right | Started with `c3 search`; named `recipe-auth-and-access`, `c3-213`, `c3-202`, `c3-209`, `c3-4`, `ref-authentication`, `ref-rbac`, `ref-nats-jwt-auth`; explained why each governing ref applies. |
| Missed / weak | Did not mention supporting refs `ref-pumped-fn` or `ref-structured-logging` from component `uses`; acceptable unless the question asks for all component refs. Did not mention the fixture `check` drift; acceptable for a query answer but useful as caveat. |
| Rubric implication | Require core ids and why-applies governance; do not require every shared backend ref unless the question asks for exhaustive constraints. Require no hallucinated rules. |
