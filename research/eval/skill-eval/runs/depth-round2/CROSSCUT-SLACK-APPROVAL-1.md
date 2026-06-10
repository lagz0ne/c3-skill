# When approval happens from Slack, what keeps web clients and next approvers consistent?

## Evidence Commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 search "Slack approval keeps web clients and next approvers consistent"
c3 read c3-215 --full
c3 read c3-205 --full
c3 read ref-sync --full
c3 read c3-211 --full
c3 read ref-approval-chain --full
c3 read adr-20260305-slack-bot-integration   # status check
c3 read adr-20260202-notification-on-step-advance   # status check
c3 graph c3-215 --format mermaid
```

## Answer

**Short version:** A Slack approve button does NOT take a side door. `c3-215` (Slack Bot Integration) re-enters the system **at the flow layer** — it calls `prFlows.approvePr` on `c3-205` (PR Flows) inside a manually constructed execution context. Because the entry point is at (not below) the flow layer, every side effect attached at the flow layer or lower fires exactly as it does for a web-originated approval: the **service-layer `sync.emit()` delta** keeps web clients consistent over `c3-4` (NATS) `sync.broadcast`, and the **flow-layer `notificationService.notifyNextApprovers` call** (on `stepAdvanced`) keeps next approvers consistent via `c3-211` (Notification System) over NATS JetStream. The audit write is attached even lower — at **storage triggers** — and `c3-215` explicitly feeds it by setting `app.current_user` via SQL `set_config`.

### Causal chain (action → mutation → mechanisms → observers → property → failure boundary)

1. **Action owner — `c3-215` (Slack Bot Integration).** Inbound path: `POST /api/slack/events` → `bot.webhooks.slack` → `slackActions` handles `approve_pr` / `reject_pr`. Because there is no HTTP request/cookie context, `slackActions` builds its own execution context: resolves Slack user → Acountee user via `slackQueries.getSlackUserById`, constructs a `UserActor`, sets `currentUserTag` and `transactionTag`, sets `app.current_user` via SQL `set_config` for audit triggers, then executes `prFlows.approvePr` within a `db.transaction()` (`c3-215` Architecture, slackActions, "Execution Context in Inbound Handlers").

2. **State mutation owner — `c3-205` (PR Flows) → `prService.approve`.** Per `ref-approval-chain` Wiring, `approvePr.flow → prService.approve → approvalQueries` (insert `approval_record`, evaluate `anyof`/`allof` app-level, advance `current_step`, `markPrAsApproved` if final). `prService.approve` returns `stepAdvanced: boolean` so the flow knows whether the chain moved to a new step (`adr-20260202-notification-on-step-advance`, status: **implemented** — current mechanism, confirmed live by `c3-205`'s Operations table: `approvePr | sync, conditional notification`).

3. **Sync mechanism (what keeps web clients consistent) — `ref-sync` (Real-time Sync Pattern).** Two-layer contract: **services emit deltas, flows send acks**, correlated by `executionId` (string). Inside `prService.approve` (after the DB write), `sync.emit({ entity: 'pr', type: 'update', id, data: updatedPr }, executionId)` publishes a full-record `DeltaMessage` via `publisher.publishToAll()` on subject `{prefix}.broadcast` (default `sync.broadcast`) on `c3-4` (NATS Server). Every connected browser's `natsSync` subscription applies the delta to reactive atoms with `applyDelta` (delete → update → add, full-record replacement). So web clients converge on the new PR status because the **delta is attached at the service layer**, below the flow entry point the Slack path uses — it cannot be skipped by entering via Slack.

4. **Notification mechanism (what keeps next approvers consistent) — `c3-211` (Notification System).** Attached at the **flow layer**: `approvePr.flow → notificationService.notifyNextApprovers (if stepAdvanced)` (`ref-approval-chain` Wiring; `c3-205` Uses table). `notifyNextApprovers(execCtx, prId)` looks up next approvers **from PR data** (the just-mutated DB state — so the recipient set is always derived from the canonical approval chain) and publishes one notification per recipient to NATS **JetStream** `NOTIFICATIONS` stream, subject `notifications.{type}.{escaped_email}` (workqueue retention, file storage, 7-day max age). `notificationDispatcher` consumes with a durable consumer, filters channels by `notification_preferences`, writes a `pending` row to `notification_log` per channel, dispatches, marks `sent`/`failed`, acks on success / naks on failure (retry). Channels: `emailChannel` (SMTP), `inAppChannel` (NATS real-time on the per-user subject `sync.user.{escaped_email}` per `ref-sync`'s subject table, + JetStream persistence), `slackChannel` (DM via `c3-215`). Slack-bot enablement was decided in `adr-20260305-slack-bot-integration`, status: **implemented** — current, matches `c3-215`/`c3-211` entity docs.

5. **Emergent property.** Consistency is *layered, not duplicated*: web clients converge via broadcast deltas (service-attached, fire-and-forget, RBAC-filtered client-side per `ref-sync` Convention "Broadcast to all, filter on client"); next approvers converge via targeted, **persistent** JetStream notifications (flow-attached, async, recipient set recomputed from DB after each mutation). Because the Slack path enters at the flow layer, both legs fire identically to a web approval — flow entry preserves side effects. Notification is **step-advance-only**: an `allof` approval that does not complete the step emits a sync delta but no notification (`c3-205` Operations: "conditional notification"; `adr-20260202`).

6. **Failure boundary.**
   - **Notification leg fails:** `c3-205` Approval Integration: "Notifications fire async with error suppression (logged, not thrown)" — the approval mutation and the sync delta still land; only the ping is lost. Below that, JetStream gives retry: dispatcher naks on failure, `notification_log` records `pending`/`sent`/`failed`, and `notificationService.retryNotification(execCtx, logId)` republishes from the log (powers admin UI retry) (`c3-211`).
   - **Slack channel specifically degrades silently:** `slackChannel` skips if no Slack user mapping exists, and all Slack atoms return `null` when `slackConfigTag` is unconfigured (`c3-215`, `c3-211`).
   - **Sync/ack leg:** `result.wait()` is "a UX optimization, not correctness-critical; timeout fallback (2s) prevents permanent hangs" (`ref-sync` Execution ID Contract). For a Slack-originated approval there is no waiting web client, so ack correlation is moot; ack is always guarded `if (executionId)` because "executionId may not exist in non-interactive contexts" (`ref-sync`).
   - **Documented gap:** neither `ref-sync` nor `c3-4` documents replay/recovery of missed `sync.broadcast` deltas for a disconnected web client — deltas are plain NATS pub/sub, only notifications are JetStream-persisted. The docs do not state how an offline web client catches up; reporting that as a gap, not guessing.
   - No `rule-*` entities surfaced for this flow (search output contained refs/ADRs/components only).

### Bar 8 — side-effect attachment layers per entry path

| Side effect | Attached at | Evidence |
| --- | --- | --- |
| `sync.emit` (delta → web clients) | **Service layer** (`prService.approve`, after DB write) | `ref-sync`: "Services call sync.emit() after DB write"; Golden Example "Inside a service method (e.g., prService.approve)" |
| `sync.ack(executionId)` | **Flow layer** (end of flow, guarded `if (executionId)`) | `ref-sync`: "Flows call sync.ack(executionId) at the end"; anti-pattern "Call sync.ack() inside a service" |
| `notifyNextApprovers` | **Flow layer** (`approvePr.flow`, on `stepAdvanced`) | `ref-approval-chain` Wiring; `c3-205` Uses table; `adr-20260202` (implemented) |
| Audit write | **Storage trigger layer** (DB triggers reading `app.current_user`) | `c3-215`: "Sets `app.current_user` via SQL `set_config` for audit triggers" |

| Entry path | Delta emit | Step-advance notification | sync.ack | Audit |
| --- | --- | --- | --- | --- |
| Web UI → server function → `prFlows.approvePr` (`c3-205`) | fires (service layer) | fires (flow layer) | fires — `executionId` set by middleware (`ref-sync`) | fires (trigger layer) |
| Slack button → `slackActions` → `prFlows.approvePr` (`c3-215`) | **fires** — attached below the flow entry point | **fires** — entry is at the flow layer, not below it | **undocumented** — `c3-215` lists `currentUserTag` + `transactionTag` being set, never `executionIdTag`; ack is guarded `if (executionId)` so it is a no-op if unset. Harmless: no web client waits on a Slack-originated executionId, and the delta itself reconciles all clients | **fires** — `slackActions` explicitly sets `app.current_user` for the triggers |
| Slack `/pending` (read-only, `prFlows.listPrs`) | n/a — `c3-205` Operations marks `listPrs` side effects "`-`" | n/a | n/a | n/a |
| Hypothetical direct `prService.approve` call (below flow) | fires (service-attached) | **skipped** — notification is flow-attached | **skipped** — ack is flow-attached | depends on caller setting `app.current_user` |

The attachment point decides survival: nothing entered below its attachment layer fires. The Slack path is safe precisely because `c3-215` deliberately re-enters at the flow layer instead of calling `prService` directly.

### Flow diagram

(`c3 graph c3-215 --format mermaid` returned TOON node rows in agent mode — relationships: `c3-215` ∈ `c3-2`, uses `ref-pumped-fn`, `ref-query-services`, `ref-structured-logging`. Diagram below is assembled from the read docs.)

```mermaid
flowchart TD
    S[Slack approve_pr button] --> W["POST /api/slack/events (c3-215)"]
    W --> A["slackActions: UserActor + currentUserTag + transactionTag + set_config app.current_user"]
    A --> F["prFlows.approvePr (c3-205) — flow layer"]
    F --> SV["prService.approve — anyof/allof, current_step (ref-approval-chain)"]
    SV --> DB[(approval_records / current_step + audit triggers)]
    SV -- "sync.emit delta (service layer)" --> B["c3-4 NATS sync.broadcast"]
    B --> WC[Web clients: applyDelta on atoms]
    F -- "if stepAdvanced (flow layer)" --> N["notificationService.notifyNextApprovers (c3-211)"]
    N --> J["JetStream NOTIFICATIONS: notifications.{type}.{escaped_email}"]
    J --> D[notificationDispatcher + notification_log]
    D --> IA["inAppChannel → sync.user.{escaped_email}"]
    D --> EM[emailChannel → SMTP]
    D --> SC["slackChannel → DM (c3-215)"]
```

## Grounding

| Material claim | Source output |
| --- | --- |
| Slack `approve_pr`/`reject_pr` route through `prFlows.approvePr`/`rejectPr`; manual context (`currentUserTag`, `transactionTag`, `set_config app.current_user`), DB transaction; webhook `POST /api/slack/events`; silent skip without user mapping | `c3 read c3-215 --full` (Architecture, slackActions, Execution Context in Inbound Handlers, slackChannel) |
| `approvePr` side effects = "sync, conditional notification"; notifications async with error suppression; `listPrs` has no side effects; flow uses `notificationService` to "Notify next approvers on step advance" | `c3 read c3-205 --full` (Operations table, Approval Integration, Uses table) |
| Services emit deltas / flows ack; `sync.emit` inside `prService.approve` after DB write; subjects `sync.broadcast` and `sync.user.{escaped_email}`; `applyDelta` full-record replacement; ack guarded `if (executionId)`; `wait()` = 2s-timeout UX optimization; broadcast-to-all + client-side filter | `c3 read ref-sync --full` (Architecture, Golden Examples, NATS Subjects, Execution ID Contract, Convention) |
| Flow→service→queries wiring incl. `notificationService.notifyNextApprovers (if stepAdvanced)`; `anyof`/`allof` evaluated app-level in `prService.approve`; data model and state machine | `c3 read ref-approval-chain --full` (Wiring, Mode Validation, Happy Path) |
| JetStream `NOTIFICATIONS` stream config; `notifications.{type}.{escaped_email}`; dispatcher ack/nak retry; `notification_log` statuses; `retryNotification`; channels email/in_app/slack; preferences filtering; `notifyNextApprovers` reads next approvers from PR data | `c3 read c3-211 --full` (notificationService, notificationPublisher, notificationDispatcher, Built-in Channels, Notification Log) |
| `adr-20260305-slack-bot-integration` status: implemented; `adr-20260202-notification-on-step-advance` status: implemented; `stepAdvanced: boolean` return added to `prService.approve` | `c3 read adr-...` outputs (frontmatter `status` + Decision sections) |
| `c3-4` = NATS WebSocket broker for real-time sync; browsers subscribe-only on `sync.broadcast` (`ref-nats-jwt-auth` snippet); candidate ranking | `c3 search "Slack approval keeps web clients and next approvers consistent"` |
| `c3-215` relationships (parent `c3-2`; refs used) | `c3 graph c3-215 --format mermaid` |

## Caveats

- **Ack on the Slack path is undocumented.** `c3-215` enumerates the tags `slackActions` sets (`currentUserTag`, `transactionTag`) and never mentions `executionIdTag`; `ref-sync` says executionId "may not exist in non-interactive contexts" and ack is guarded. Whether `sync.ack` fires for Slack-originated approvals is therefore not stated in the docs. Consistency is unaffected (delta carries the data; no web client waits on that executionId), but verify in code before relying on ack-side behavior.
- **No documented catch-up for missed broadcast deltas.** `ref-sync` persists nothing for `sync.broadcast` (only notifications go through JetStream); no doc read describes replay for a web client that was disconnected during the Slack approval. Explicit gap per the docs read.
- **Notification is step-advance-only** (`c3-205`, `adr-20260202`): a non-advancing `allof` approval updates web clients via delta but notifies nobody — approvers still pending on the *same* step rely on the delta/UI, not a notification.
- **Subject-prefix coupling** (`ref-sync` Subject Prefix Contract): frontend hardcodes `sync.broadcast` / `sync.user.{escaped_email}`; changing `NATS_SUBJECT_PREFIX` breaks web-client consistency unless frontend wiring changes in lockstep.
- ADRs cited (`adr-20260305-slack-bot-integration`, `adr-20260202-notification-on-step-advance`) are both `status: implemented` — labeled current/historical work orders whose content matches the live entity docs read above; the entity docs (`c3-215`, `c3-205`, `c3-211`) remain the architectural truth.
- No `rule-*` entities found in the search output for this flow.
