# PROPERTY-CONFIG-BLAST-RADIUS-1 — What is affected if `NATS_SUBJECT_PREFIX` changes away from `sync`?

## Evidence Commands

```bash
c3 search "NATS_SUBJECT_PREFIX subject prefix config sync"
c3 read ref-scope-controlled-config --full        # config mechanism (env -> tag)
c3 read ref-sync --full                           # domain contract the config feeds
c3 graph ref-scope-controlled-config --direction reverse
c3 graph ref-sync --direction reverse
c3 graph ref-scope-controlled-config --direction reverse --format mermaid
c3 graph ref-sync --direction reverse --format mermaid
c3 read c3-209 --full                             # NATS Credential Generator
c3 read c3-101 --full                             # State Management (frontend)
c3 read c3-211 --full                             # Notification System
c3 read c3-4 --full                               # NATS Server (External)
c3 read c3-2 --full                               # API Backend container
c3 read c3-205 --full; c3 read c3-206 --full; c3 read c3-207 --full
c3 read c3-210 --full; c3 read c3-212 --full      # flow components citing ref-sync
c3 read c3-104 --full; c3 read c3-105 --full      # screens named in ref-sync Cited By
c3 read ref-nats-jwt-auth --full
c3 read recipe-realtime-sync --full
c3 read recipe-approval-workflow --full
c3 read recipe-backend-foundations --full         # grep prefix/subject/NATS/sync -> no hits
c3 read adr-20260112-nats-auth-callout            # status check
c3 read adr-20260112-nats-websocket-sync          # status check
c3 read adr-20260126-user-notification-ui         # status check
```

## Answer

### The two mechanisms

`NATS_SUBJECT_PREFIX` is read **once at scope creation** and injected as the tag
`natsConfig.subjectPrefix` (ref-scope-controlled-config, "Scope Creation":
`natsConfig.subjectPrefix(env.NATS_SUBJECT_PREFIX)`; cited by the serverConfig atom
in server.tsx per its "Cited By"). That tag feeds the real-time sync subject contract
(ref-sync, "Subject Prefix Contract"): server-side subjects are prefix-driven —
`{prefix}.broadcast` and `{prefix}.user.{escaped_email}`, default `sync`.

**Causal chain** (each hop bound to a read, below):

UI mutation → flow (c3-205/206/207/210/212, side effect "sync" per their Operations
tables) → service `sync.emit()` / flow `sync.ack()` (ref-sync, Architecture) →
publisher publishes on `{prefix}.broadcast` / `{prefix}.user.{escaped_email}`
(ref-sync, NATS Subjects table; c3-2 Wiring row "NATS Publisher | c3-4 (NATS) | TCP |
Broadcast sync events") → c3-4 routes `{prefix}.>` and enforces JWT subject
permissions (c3-4, Responsibilities + Required Permissions) → browser connects with a
JWT from c3-209 whose subscribe permissions are `{prefix}.broadcast,
{prefix}.user.{escaped_email}` (c3-209, Permission Model) → frontend atom in c3-101
subscribes to the **literal** `sync.broadcast` / `sync.user.<email>` (c3-101, "NATS
Sync Wiring") → deltas update stores, acks resolve `executionTracker` /
`result.wait()` (ref-sync, Golden Examples).

The prefix is therefore a **shared coupling value across four parties**: server
publisher config tag, JWT permission generator (c3-209), broker permission/routing
config (c3-4), and frontend subscription strings (c3-101). The first three are
prefix-driven; the frontend is hardcoded to `sync.*` — that asymmetry is the blast
radius's breaking point. Both c3-101 and ref-sync state it explicitly: "if prefix
changes, frontend subscription wiring must be updated to match" / "must change in
lockstep."

### Impact graph (constructed from the two reverse-graph outputs)

```mermaid
graph RL
  subgraph config mechanism — c3 graph ref-scope-controlled-config --direction reverse
    c3-202[c3-202 Execution Context] --> RSCC[ref-scope-controlled-config]
    c3-203[c3-203 Middleware Stack] --> RSCC
    c3-204[c3-204 Drizzle ORM] --> RSCC
    c3-209[c3-209 NATS Credential Generator] --> RSCC
    c3-211[c3-211 Notification System] --> RSCC
    RBF[recipe-backend-foundations] --> RSCC
  end
  subgraph domain contract — c3 graph ref-sync --direction reverse
    c3-101[c3-101 State Management] --> RS[ref-sync]
    c3-205[c3-205 PR Flows] --> RS
    c3-206[c3-206 Invoice Flows] --> RS
    c3-207[c3-207 Payment Flows] --> RS
    c3-210[c3-210 Admin Flows] --> RS
    c3-212[c3-212 Workbench Flows] --> RS
    RAW[recipe-approval-workflow] --> RS
    RRS[recipe-realtime-sync] --> RS
    ADR1[adr-20260112-nats-auth-callout · implemented/historical] --> RS
    ADR2[adr-20260112-nats-websocket-sync · implemented/historical] --> RS
  end
```

### Affected entities — direct consumers vs transitive dependents

Each row's behavior is bound to that entity's own read output (see Grounding).

| Entity | Type | Impact | Why (from its own doc) |
|---|---|---|---|
| c3-209 (NATS Credential Generator) | component | **direct** | Consumes tag `natsConfig.subjectPrefix` (Configuration table, example `sync`); embeds subscribe permissions `{prefix}.broadcast`, `{prefix}.user.{escaped_email}` into every user JWT. New prefix ⇒ new JWTs grant different subjects. |
| c3-101 (State Management) | component | **direct — breaking** | `natsSync` atom subscribes to hardcoded `sync.broadcast` + `sync.user.<email>`; doc: "if prefix changes, frontend subscription wiring must be updated to match." Without lockstep change, no deltas/acks/notifications arrive; `executionTracker` waits fall to 2s timeout. |
| c3-4 (NATS Server (External)) | container | **direct** | Routes `{prefix}.>`; APP account required pub/sub permissions `{prefix}.>   # default prefix: sync`; browser permission model keyed on `{prefix}.broadcast` / `{prefix}.user.{escaped_email}`. Broker permission config must carry the new prefix. |
| c3-2 (API Backend) | container | **direct** | Owns the "NATS Publisher" wiring to c3-4 ("Broadcast sync events"); ref-scope-controlled-config "Cited By" places the serverConfig atom (which aggregates `natsConfig.subjectPrefix`) in server.tsx here. The env var itself is read at this container's scope creation. |
| ref-sync (Real-time Sync Pattern) | ref | **direct (doc)** | Its "Subject Prefix Contract" and subject table state the default `sync`; client-side golden example hardcodes `const prefix = 'sync'`. Doc must be updated with the new default. |
| ref-scope-controlled-config (Scope-Controlled Configuration) | ref | **direct (doc)** | Names `natsConfig.subjectPrefix` / `env.NATS_SUBJECT_PREFIX` as its worked example; mechanism unchanged, example value stale. |
| c3-205, c3-206, c3-207, c3-210, c3-212 (PR/Invoice/Payment/Admin/Workbench Flows) | components | **transitive** | Each lists `sync — Client notification via NATS after mutations` in Uses and `sync` side effects per operation; none references the prefix or subjects directly. Reached through ref-sync's publisher. Their mutations still commit; only the client-notification side effect stops being observed if subscriptions mismatch. |
| c3-211 (Notification System) | component | **transitive (in-app real-time leg only)** | Its JetStream pipeline uses subjects `notifications.{type}.{escaped_email}` — not `{prefix}`-driven, unaffected. Its `inAppChannel` delivers via "NATS publish (real-time) + JetStream (persistence)"; the real-time publish is the `{prefix}.user.{escaped_email}` path (ref-sync subject table: `publisher.publishToUser()`), which breaks for un-updated clients. Durable JetStream persistence keeps working. |
| c3-104 (InvoiceScreen), c3-105 (PaymentRequestsScreen) | components | **transitive** | Both docs state "Real-time updates via NATS sync" in Data Flow; they reach NATS only through the c3-101 atoms. They lose live updates, not core CRUD. |
| recipe-realtime-sync, recipe-approval-workflow | recipes | **transitive (doc)** | recipe-realtime-sync restates the `{prefix}.broadcast` / `{prefix}.user.{escaped_email}` subjects; recipe-approval-workflow restates "every mutation emits a sync delta (ref-sync)". Narrative docs to re-verify, no runtime behavior. |
| c3-202, c3-203, c3-204 | components | **not affected** | In the ref-scope-controlled-config reverse graph (they cite the config *mechanism*), but their read bodies contain no `natsConfig.subjectPrefix` / subject usage — they consume other tags. Graph neighbors, not consumers of this value. |
| recipe-backend-foundations | recipe | **not affected** | In the config-mechanism reverse graph; full read grepped for prefix/subject/NATS/sync returned no hits. |
| adr-20260112-nats-auth-callout, adr-20260112-nats-websocket-sync, adr-20260126-user-notification-ui | adr | **historical only** | All `status: implemented` ⇒ historical work orders, not live mechanisms. adr-20260126-user-notification-ui records "The `subjectPrefix` is `sync` (from tags). Client must use same escaping" — consistent with current docs, no action. |

### Failure boundary

If only the env var changes (server side) and nothing else:

1. **Permission boundary moves first.** New JWTs from c3-209 grant subscribe on
   `<new>.broadcast` / `<new>.user.*`; the c3-101 client still subscribes to
   `sync.broadcast` / `sync.user.*`, which the JWT no longer permits — c3-4 enforces
   "pub/sub permissions from JWT" (c3-4 Responsibilities), so the subscription is
   denied/yields nothing.
2. **Broker config may reject the publisher.** c3-4's APP account permission lists
   `{prefix}.>` with "default prefix: sync" baked into its config example; if the
   broker config literal stays `sync.>`, backend publishes to `<new>.>` are denied at
   the broker even before any client concern.
3. **Degradation is silent, not crashing.** HTTP mutations, DB writes, flows, audit,
   and JetStream notifications all still succeed (flows don't touch subjects; c3-211's
   `notifications.>` stream is prefix-independent). What is lost: live deltas, acks,
   and in-app real-time notifications. Per ref-sync's Execution ID Contract,
   `result.wait()` is "a UX optimization, not correctness-critical; timeout fallback
   (2s)" — so every originating client's wait degrades to a 2s timeout, other clients
   see stale data until reload. Per ref-sync anti-patterns, a missing ack means
   "executionTracker falls back to timeout, causing sluggish UI" — the observable
   symptom of the whole mismatch.

### Verification checks (concrete)

| Check | How |
|---|---|
| Server tag actually carries the new value | Confirm scope creation injects `natsConfig.subjectPrefix(env.NATS_SUBJECT_PREFIX)` in the server.tsx serverConfig atom (ref-scope-controlled-config "Scope Creation" + "Cited By"); log/inspect resolved `config.natsSubjectPrefix` at startup |
| c3-209 JWTs follow the prefix | Decode a generated user JWT; assert subscribe permissions are `<new>.broadcast` and `<new>.user.<escaped_email>` (c3-209 Permission Model) |
| c3-4 broker permissions updated | Update/verify APP account `pub`/`sub` lists contain `<new>.>` (c3-4 Required Permissions); health: `curl http://nats-server:8222/healthz` |
| c3-101 frontend wiring updated in lockstep | Change `natsSync` atom subscription strings from `sync.broadcast` / `sync.user.<email>` to the new prefix (c3-101 "NATS Sync Wiring"; ref-sync "Subject Prefix Contract") |
| Sync observable end-to-end | With two browser sessions, run a mutation from c3-205 (e.g. `approvePr`); assert a delta arrives on `<new>.broadcast` in the second session and `result.wait()` resolves before the 2s timeout in the first (ref-sync Golden Examples) |
| Targeted notification observable | Trigger `requestApprovals` (c3-205, "notifies next approvers"); assert the in-app notification arrives on `<new>.user.<escaped_email>` (ref-sync subject table; c3-211 inAppChannel) |
| Notification durability untouched | Confirm JetStream `NOTIFICATIONS` stream subjects `notifications.{type}.{escaped_email}` still consume/dispatch (c3-211 notificationPublisher/Dispatcher) — should be prefix-independent |
| Failure-mode probe | Intentionally leave one leg stale (e.g. frontend still on `sync.*`): assert mutation still succeeds over HTTP, `result.wait()` times out at 2s, no delta applied — confirms silent-degradation boundary, then fix |
| Docs re-synced | Update default-prefix mentions in ref-sync, c3-101, c3-209, c3-4, recipe-realtime-sync; run `c3 check` |

## Grounding

| Material claim | Evidence source |
|---|---|
| Env var read once at scope creation, injected as `natsConfig.subjectPrefix` tag; serverConfig atom in server.tsx (c3-2) | `c3 read ref-scope-controlled-config --full` — "Scope Creation", Conventions table, "Cited By" |
| Subjects are `{prefix}.broadcast` / `{prefix}.user.{escaped_email}`, default `sync`; lockstep requirement; publisher methods `publishToAll()` / `publishToUser()` | `c3 read ref-sync --full` — "NATS Subjects", "Subject Prefix Contract" |
| Config-mechanism dependents = c3-202, c3-203, c3-204, c3-209, c3-211, recipe-backend-foundations | `c3 graph ref-scope-controlled-config --direction reverse` (node list) |
| Sync-contract dependents = c3-101, c3-205, c3-206, c3-207, c3-210, c3-212, recipe-approval-workflow, recipe-realtime-sync, adr-20260112-nats-auth-callout, adr-20260112-nats-websocket-sync | `c3 graph ref-sync --direction reverse` (node list) |
| c3-209 consumes `natsConfig.subjectPrefix` (example `sync`); JWT subscribe permissions `{prefix}.broadcast`, `{prefix}.user.{escaped_email}`; publish empty; WebSocket-only | `c3 read c3-209 --full` — Dependencies, Permission Model, Configuration tables |
| c3-101 subscribes to literal `sync.broadcast` / `sync.user.<email>`; "if prefix changes, frontend subscription wiring must be updated to match"; 2s `wait()` timeout fallback | `c3 read c3-101 --full` — "NATS Sync Wiring", Atoms table |
| c3-4 routes `{prefix}.>`, enforces JWT pub/sub permissions, APP account needs `{prefix}.>` (default sync), monitoring on 8222 | `c3 read c3-4 --full` — Wiring, Responsibilities, Required Permissions, Health Check |
| c3-2 owns NATS Publisher wiring ("Broadcast sync events") and the flow components | `c3 read c3-2 --full` — Wiring, Components, Entry Points |
| Flows c3-205/206/207/212 list `sync — Client notification via NATS after mutations` in Uses with `sync` side effects per operation; no prefix/subject reference in their bodies | `c3 read c3-205/206/207/212 --full` — Uses + Operations tables (c3-210 cites ref-sync per its frontmatter `uses` from the reverse graph and `c3 read c3-210`; its truncated body showed no subject usage either) |
| c3-211 JetStream subjects `notifications.{type}.{escaped_email}`; inAppChannel = "NATS publish (real-time) + JetStream (persistence)"; NOTIFICATIONS stream workqueue/file/7-day | `c3 read c3-211 --full` — notificationPublisher, Built-in Channels |
| c3-104 / c3-105 receive "Real-time updates via NATS sync" via atoms | `c3 read c3-104 --full`, `c3 read c3-105 --full` — Data Flow sections |
| `result.wait()` is UX optimization, timeout fallback 2s; missing ack ⇒ "sluggish UI" | `c3 read ref-sync --full` — Execution ID Contract, Anti-patterns |
| Sync ephemeral vs notifications durable; subjects restated in narrative | `c3 read recipe-realtime-sync --full` — Narrative, Risk |
| "Every mutation emits a sync delta (ref-sync), then the flow acks"; notifications fire-and-forget | `c3 read recipe-approval-workflow --full` — Cross-Cutting Contracts |
| c3-202/c3-203/c3-204 bodies show no subjectPrefix/subject usage | `c3 read c3-202/c3-203/c3-204` outputs (no prefix/NATS-subject content) |
| recipe-backend-foundations has no prefix/subject/NATS/sync content | `c3 read recipe-backend-foundations --full` piped through grep — zero hits |
| All three cited ADRs are `status: implemented` ⇒ historical; adr-20260126-user-notification-ui records `subjectPrefix` is `sync` (from tags) | `c3 read adr-...` status lines; `c3 search` snippet for adr-20260126-user-notification-ui |

## Caveats

- **Cited By vs graph drift on ref-sync:** ref-sync's body "Cited By" lists c3-104,
  c3-105, c3-209, and c3-211, but `c3 graph ref-sync --direction reverse` returns none
  of them (and their frontmatter `uses` lines omit ref-sync). The prose citation list
  and the wired graph disagree; I treated the four as dependents anyway after reading
  each, but the topology wiring is stale on one side.
- **Permission-table inconsistency:** ref-nats-jwt-auth's Permissions Model says
  browser clients subscribe only to `sync.broadcast` (literal, no user subject), while
  c3-209's Permission Model grants `{prefix}.broadcast` **and**
  `{prefix}.user.{escaped_email}`. The docs disagree on both prefix-drivenness and the
  user subject; c3-209 (the credential owner) is the more specific source.
- **Broker config literalness is unconfirmed:** c3-4 shows `{prefix}.>` with comment
  "default prefix: sync" in its Required Permissions block, but whether the deployed
  nats.conf derives the prefix from `NATS_SUBJECT_PREFIX` or hardcodes `sync.>` is not
  stated in any read doc — c3-4's Environment Variables table lists only JWT/account
  vars, not `NATS_SUBJECT_PREFIX`. Verify against the actual broker config (failure
  boundary item 2 depends on it).
- `--format mermaid` in this fixture returned the same TOON node list as the plain
  graph command; the mermaid block above is constructed verbatim from those node/edge
  lists, not from a rendered mermaid output.
- Truncated first reads of c3-205/206/207/210/212 were followed by `--full` reads for
  all except c3-210, whose `--full` output surfaced only frontmatter in my filter; its
  ref-sync dependence is grounded in its `uses` frontmatter and the reverse graph, and
  I assign it only the same transitive flow-layer behavior its siblings document.
