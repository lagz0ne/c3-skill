# CROSSCUT-STEP-ADVANCE-1 — Why do later-step approvers hear about a PR only after the prior step completes?

## Evidence Commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }

c3 search "approval step advance notification PR approvers notified"
c3 read recipe-approval-workflow --full
c3 read ref-approval-chain --full
c3 read c3-205 --full          # PR Flows
c3 read c3-211 --full          # Notification System
c3 read adr-20260202-notification-on-step-advance --full
c3 read adr-20260121-notification-system            # status check
c3 read ref-pull-dispatcher --full
c3 read ref-sync --full
c3 graph c3-205 --depth 1 --format mermaid
c3 graph c3-205 --format mermaid
c3 lookup 'src/**/pr*' ; c3 lookup 'src/**/notification*' ; c3 lookup '**/pr.ts'
```

## Answer

**Layer:** c3-205 (PR Flows) → c3-211 (Notification System), governed by ref-approval-chain and decided in adr-20260202-notification-on-step-advance.

**Short version:** notification of approvers is not an event on "someone approved" — it is an event on "the chain's `current_step` pointer moved". The recipient set is computed *from* `current_step`, and `current_step` only moves when the prior step's mode condition (`anyof`/`allof`) is satisfied. Until then, later-step approvers are by construction not "next approvers", and the flow doesn't even call the notifier. This is a deliberate, ADR-recorded design, not an accident.

### Causal chain (Step 0a++ items 1–6, with the WHY carried by each arrow)

**1. Action owner — `approvePr` / `approveAll` flows (c3-205).**
An approver acts in the Approvals mode of PaymentRequestsScreen (c3-105, from search output) and the mutation enters through the `approvePr` flow (also `approveAll` for bulk). c3-205's Operations table: `approvePr` — "Records current user's approval; if step advances, notifies next approvers | sync, conditional notification."
→ *why the next hop:* flows orchestrate, services mutate — adr-20260202 Rationale explicitly rejected "Notify in prService.approve" because "Services should not trigger side effects; flows orchestrate per C3 pattern". So the flow must be told whether to notify.

**2. State mutation owner — `prService.approve` under ref-approval-chain.**
The service inserts an `approval_records` row, then runs the **app-level mode check** (ref-approval-chain "Mode Validation"): `anyof` = one assigned approver suffices; `allof` = every assigned user must have a record. Only when `isProgressed` is true does `updateApprovalCurrentStep` move `current_step: N → N+1` (Golden Example "Approve + Step Advance"). An `allof` step with user A signed but user B pending keeps `current_step` parked at N (Happy Path trace in ref-approval-chain).
→ *why the next hop:* the carrier is the **return contract** added by adr-20260202 Decision 1 — `stepAdvanced: isProgressed && !willEnd`. Before that ADR, the flow literally could not know the step moved ("this information is not returned to the calling flow, so the flow cannot know to trigger notifications" — ADR Problem section).

**3. The notification gate — conditional dispatch in the flow.**
adr-20260202 Decision 2: `if (result.success && result.stepAdvanced) { await ctx.exec({ fn: notificationService.notifyNextApprovers, ... }) }`. ref-approval-chain's Wiring table confirms this is the live shape: `approvePr.flow → ... → notificationService.notifyNextApprovers (if stepAdvanced)`. `requestApprovals` is the one unconditional caller — it notifies step-0 users when the PR enters `pending` (c3-205 "Approval Integration").
→ *why the next hop:* even when called, **who** gets notified is derived from state — c3-211: "`notifyNextApprovers(execCtx, prId)` — looks up next approvers from PR data". "Next approvers" = the `approval_step_users` of the step `current_step` now points at (ref-approval-chain Data Model: "`current_step` points to the step awaiting approval"). So a step-2 approver only becomes addressable after step 1's completion has moved the pointer to 2. Two gates stack: (a) the flow only calls the notifier on advancement, (b) the notifier's recipient lookup only ever resolves the *current* step's users.

**4. Notification mechanism — c3-211, NATS JetStream, per-recipient subjects.**
`notificationService → notificationPublisher → JetStream NOTIFICATIONS stream` (Workqueue retention, file storage, 7-day max age, 10K max messages), subject pattern `notifications.{type}.{escaped_email}` — one message per recipient. `notificationDispatcher` (durable consumer, ref-pull-dispatcher pattern: channels self-register via `dispatcher.subscribe()`) filters against per-user `notification_preferences` (default `['in_app']`) and dispatches to `emailChannel` (SMTP), `inAppChannel` (NATS real-time `{prefix}.user.{escaped_email}` + JetStream persistence — subject from ref-sync's NATS Subjects table), and `slackChannel` (DM via c3-215 bot; skips gracefully if Slack unconfigured).
→ *why this matters for the question:* notifications are **targeted per-user subjects**, distinct from the sync layer. ref-sync: every mutation also emits a delta on `{prefix}.broadcast` to *all* clients. So a later-step approver with the app open can already *see* the PR change via broadcast deltas — what they don't get early is the personal "act now" notification, because that channel is reserved for actionability.

**5. Emergent property — step-advance-only notification, by decision.**
adr-20260202's Rationale table rejects the alternatives that would notify earlier or later: "Notify on every approval — Wasteful: only notify when step actually advances"; "Notify when PR fully approved — No action required, PR is done"; "Infer from next_approver presence — Fragile: next_approver exists even when step didn't advance (multiple approvers in same step)". The system equates *notification* with *your step is now awaiting you*. Mid-step `allof` approvals, rejects/recalls (chain reset to draft, step 0, records cleared — ref-approval-chain Edge Cases; ADR Step-Advancing Paths Audit marks them "No - approval chain reset"), and the final approval (`willEnd` excludes it via `stepAdvanced = isProgressed && !willEnd`) all produce **no** next-approver notification.

**6. Failure boundary.**
- **Mutation vs notification:** notifications are fire-and-forget — "async with error suppression (logged, not thrown)" (c3-205 Approval Integration; recipe-approval-workflow Cross-Cutting Contracts). A notification failure never rolls back the approval; the DB advance, the audit capture (DB trigger on `pr` table — recipe), and the sync delta/ack all survive. The cost is silent-to-the-approver: a later-step approver may *never* hear about a PR that is now waiting on them; only logs reveal it at the flow layer.
- **Inside the dispatcher:** per-channel `pending` log entry → handler → `sent`/`failed`; ack on success, **nak on failure (retry)** via the durable JetStream consumer (c3-211 notificationDispatcher). Every attempt lands in `notification_log`, which "powers admin UI retry and monitoring"; `retryNotification(execCtx, logId)` republishes from the log. So the observer of failure is the admin (notification logs in c3-107 Admin Screens, per search output), not the approver.
- **Races:** two concurrent step-completing approvals both triggering notifications are deduped — "notificationPublisher.publish is idempotent (JetStream dedupes by message ID)" (adr-20260202 Concurrency/Idempotency). Transaction scope for concurrent approvals comes from the execution context (ref-approval-chain Edge Cases; recipe: "All operations run in transaction scope (c3-202)").
- **Bypass:** if the JetStream leg is down, the broadcast sync delta (separate NATS path, ref-sync) still propagates the PR's new state to connected clients — visibility degrades from "pushed to you" to "visible if you look".

**Graph** (from `c3 graph c3-205` node output — CLI returned TOON node/edge lists in agent mode for both `--format mermaid` invocations; rendered here from those edges):

```mermaid
graph LR
  c3_105[c3-105 PaymentRequestsScreen] -->|approve action| c3_205[c3-205 PR Flows]
  c3_205 -->|governed by| ref_ac[ref-approval-chain]
  c3_205 -->|stepAdvanced=true| c3_211[c3-211 Notification System]
  c3_205 -->|emit delta + ack| ref_sync[ref-sync broadcast]
  c3_211 -->|ref-pull-dispatcher| ch[email / in_app / slack channels]
  adr1[adr-20260202 step-advance notify] -.affects.-> c3_205
  adr1 -.affects.-> c3_211
  adr0[adr-20260121 notification system] -.affects.-> c3_205
```

### Concrete checks (if you need to verify or change this behavior)

1. **Behavioral probe** (mirrors adr-20260202 Verification checklist): 3-step chain (A step 0, B step 1, C step 2). `requestApprovals` → only A notified; A approves → B notified; B approves → C notified; C approves → no notification. Assert rows in `notification_log` after each.
2. **`allof` mid-step negative:** step with users {A, B} in `allof`; A approves → assert `current_step` unchanged and **no** new `notification_log` entries.
3. **Owner surfaces to touch for any change:** the `stepAdvanced` return contract in `prService.approve`, the conditional in `approvePr`/`approveAll` flows (c3-205), and `notifyNextApprovers`'s current-step lookup (c3-211). Per recipe-approval-workflow Risk: "Approval chain logic spans c3-205 (flows) + prService + approvalQueries... blast radius across all three layers plus the notification dispatch path."
4. **Runtime observables:** JetStream `NOTIFICATIONS` stream / `notifications.{type}.{escaped_email}` subjects for publish; `{prefix}.user.{escaped_email}` for in-app delivery; `notification_log` status transitions `pending → sent|failed`; flow log line "notified next approvers after step advance".
5. **Failure-mode probe:** kill a channel (e.g. bad SMTP) and approve through a step — assert the approval still commits, sync delta still broadcasts, `notification_log` shows `failed`, and admin retry republishes it.

**ADR status labels:** adr-20260202-notification-on-step-advance — `status: implemented`, **current** (its decision matches the live wiring in c3-205's Operations table and ref-approval-chain's Wiring table; no newer ADR supersedes it in search/graph output). adr-20260121-notification-system — `status: implemented`, **historical foundation** (built c3-211 itself; partially superseded in one respect: it *removed* Slack, which adr-20260305-slack-bot-integration later re-added as a channel — both visible in search output, c3-211 lists slackChannel as current).

## Grounding

| Material claim | Source (command output) |
| --- | --- |
| `approvePr` notifies next approvers only when step advances; notifications async, error-suppressed | `c3 read c3-205 --full` — Operations table + "Approval Integration" section |
| `requestApprovals` notifies first-step (step 0) approvers unconditionally | `c3 read c3-205 --full` Approval Integration; `c3 read adr-20260202 --full` Step-Advancing Paths Audit |
| `anyof`/`allof` mode check is app-level in `prService.approve`; `current_step` advances only when satisfied; `allof` mid-step keeps step parked | `c3 read ref-approval-chain --full` — Mode Validation, Happy Path, Golden Example "Approve + Step Advance" |
| `current_step` points at the step awaiting approval; recipients per step live in `approval_step_users` | `c3 read ref-approval-chain --full` — Data Model |
| `stepAdvanced = isProgressed && !willEnd` return contract; flow conditional `if (result.success && result.stepAdvanced)`; rejected alternatives (notify-every-approval = wasteful, notify-on-final = no action needed, infer-from-next_approver = fragile, notify-in-service = services don't do side effects) | `c3 read adr-20260202-notification-on-step-advance --full` — Decision 1, Decision 2, Rationale table |
| Before the ADR, step 1+ users were never notified because the flow couldn't know the step advanced | same ADR — Problem section |
| `notifyNextApprovers` looks up next approvers from PR data, one notification per recipient; JetStream `NOTIFICATIONS` stream, subject `notifications.{type}.{escaped_email}`; dispatcher durable consumer, preference filtering, pending/sent/failed log, ack/nak retry; channels email/in_app/slack; `notification_log` powers admin retry; `retryNotification` | `c3 read c3-211 --full` — Components, notificationPublisher, notificationDispatcher, Built-in Channels, Notification Log |
| Channels self-register via `dispatcher.subscribe()` (dependency inversion) | `c3 read ref-pull-dispatcher --full` — Choice/Pattern |
| Sync layer is separate: deltas broadcast to all on `{prefix}.broadcast`, per-user subject `{prefix}.user.{escaped_email}`; services emit deltas, flows ack | `c3 read ref-sync --full` — Architecture, NATS Subjects table |
| Notifications fire-and-forget; audit via DB trigger; transaction scope c3-202; blast radius spans c3-205 + prService + approvalQueries + dispatch path | `c3 read recipe-approval-workflow --full` — Cross-Cutting Contracts, Risk |
| JetStream dedupes racing notification publishes by message ID | `c3 read adr-20260202 --full` — Concurrency/Idempotency |
| Approver enters via PaymentRequestsScreen approvals mode (c3-105); admin notification logs UI (c3-107) | `c3 search` output — c3-105 and c3-107 rows |
| Wiring `approvePr.flow → prService.approve → ... → notificationService.notifyNextApprovers (if stepAdvanced)` is current | `c3 read ref-approval-chain --full` — Wiring table |
| ADR statuses (both `implemented`); Slack removed then re-added | `c3 read adr-20260202 --full`, `c3 read adr-20260121-notification-system` (frontmatter + Problem/Decision), `c3 search` output (adr-20260305-slack-bot-integration row) |
| Graph relationships c3-205 ↔ refs/ADRs/container | `c3 graph c3-205 --depth 1 --format mermaid` node/edge output |

## Caveats

- **No `rule-*` entities found** governing this path: `c3 graph c3-205` and `c3 read c3-205/c3-211` show only `ref-*` citations (ref-approval-chain, ref-sync, ref-pull-dispatcher, etc.); the search output returned no rule rows.
- **Code could not be inspected:** `c3 lookup 'src/**/pr*'`, `'src/**/notification*'`, and `'**/pr.ts'` all returned empty `files`/`components` (codemap coverage gap reported by the CLI), and the fixture directory contains only `README.md` beside `.c3/`. All flow/service code claims above therefore rest on the C3 docs (ref-approval-chain Wiring/Golden Examples, ADR code excerpts), not on read source files.
- **Graph mermaid rendering:** `c3 graph c3-205 --format mermaid` returned TOON node lists (twice) rather than mermaid text in agent mode; the mermaid block above is hand-rendered from those returned nodes/edges.
- **Approver-side failure visibility gap:** the docs state errors are "logged, not thrown" (c3-205) and that `notification_log`/admin UI observe failures (c3-211); no doc describes any mechanism that re-alerts the *approver* if their step-advance notification permanently fails — the documented recovery is admin retry. The docs don't state whether nak-retry is bounded (no max-deliveries figure appears in `c3 read c3-211 --full`); reporting that as a gap, not a guess.
- **First `c3 search` attempt errored** (`SQL logic error: no such column: step`) on a hyphenated query ("later-step"); rerun without hyphens succeeded — results above come from the successful query only.
