# Acountee cross-cutting cases

Ground truth was established with the local C3 wrapper:

```bash
C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 <cmd>
```

These cases are intentionally indirect. A complete answer must trace the path
across several mechanisms and name the emergent property, not just list ids.

## CROSSCUT-MASS-APPROVAL-1: mass approval and user notification

Question: Explain how mass-approval gets done and informs other users.

Grounding commands:

```bash
c3 search "bulk approve payment requests notify approvers sync"
c3 search "approval notification sync non blocking NATS websocket"
c3 read c3-105 --full
c3 read c3-205 --full
c3 read c3-211 --full
c3 read ref-sync --full
c3 read recipe-realtime-sync --full
c3 read adr-20260202-notification-on-step-advance --full
c3 graph c3-205 --depth 1
c3 graph c3-211 --depth 1
```

Expected trace:

| Segment | Expected ids and facts |
| --- | --- |
| Action / command location | `c3-105` exposes Approvals mode and bulk `approveAll`; `c3-205` owns the PR Flows command, where `approveAll` iterates `pr_ids`, approves each PR, and returns approved/failed arrays. |
| State change | `c3-205` calls approval-chain logic governed by `ref-approval-chain`; PR status/step may advance. |
| Sync mechanism | `ref-sync` defines NATS WebSocket sync: services emit deltas, flows emit ack with `executionId`; `c3-101` `natsSync` subscribes to `sync.broadcast` for deltas/acks. |
| Notification target/model | `c3-211` is the Notification System; notifications are persisted/logged and dispatched to recipients. |
| Notification mechanism | `adr-20260121-notification-system` and `c3-211` use NATS JetStream `NOTIFICATIONS`, notificationPublisher/notificationDispatcher, and in-app NATS user subjects. |
| Emergent property | Notifications are async/non-blocking for approval/sync: `c3-205` says notifications fire async with error suppression, logged not thrown. |

Evidence snippets:

```text
search: recipe-approval-workflow ... c3-205 ... ref-approval-chain ... ref-sync
search: c3-205 ... approvePr ... approveAll ... notificationService ... sync
read c3-105: Approvals mode ... Approve, reject, bulk approve
read c3-205: approveAll | Bulk approval: iterates pr_ids ... sync, conditional notifications per PR
read c3-205: Notifications fire async with error suppression (logged, not thrown)
read ref-sync: services emit deltas, flows emit acknowledgements using executionId
read c3-101: subscribes to sync.broadcast (delta) and sync.user.<email> (notifications)
read c3-211: Publisher/dispatcher pattern for async notifications via NATS JetStream
graph c3-205: uses ref-approval-chain and ref-sync; affected by adr-20260202-notification-on-step-advance
graph c3-211: affected by adr-20260126-user-notification-ui and adr-20260202-notification-on-step-advance
```

## CROSSCUT-STEP-ADVANCE-1: later approver visibility

Question: Why do later-step approvers hear about a PR only after the prior step completes?

Grounding commands:

```bash
c3 search "bulk approve payment requests notify approvers sync"
c3 search "notification system NATS websocket sync async approval"
c3 read c3-205 --full
c3 read ref-approval-chain --full
c3 read adr-20260202-notification-on-step-advance --full
c3 read c3-211 --full
c3 read ref-pull-dispatcher --full
c3 read ref-sync --full
c3 read c3-101 --full
c3 graph c3-205 --depth 1
c3 graph c3-211 --depth 1
```

Expected trace:

| Segment | Expected ids and facts |
| --- | --- |
| Action / command location | `c3-205` owns `requestApprovals`, `approvePr`, and `approveAll`; `adr-20260202-notification-on-step-advance` changed `approvePr` and `approveAll`. |
| Approval semantics | `ref-approval-chain` defines multi-step chains and `anyof`/`allof`; notification depends on actual step advancement, not merely any approval. |
| Sync mechanism | `ref-sync` defines the NATS WebSocket delta/ack path; `c3-101` receives broadcast deltas and resolves `executionTracker`. |
| Notification target/model | `c3-211` notificationService/notificationPublisher/notificationDispatcher sends next approver notifications. |
| Notification mechanism | `ref-pull-dispatcher` governs channels self-registering with the dispatcher; `c3-211` channels include in-app, email, and Slack. |
| Emergent property | The explicit `stepAdvanced` signal avoids fragile inference; notifications are sent only when the step actually advances and not when PR is fully approved. |

Evidence snippets:

```text
read adr-20260202-notification-on-step-advance: approvePr flow calls notificationService.notifyNextApprovers when result.stepAdvanced
read adr-20260202-notification-on-step-advance: approveAll has same issue/fix for each PR with stepAdvanced: true
read adr-20260202-notification-on-step-advance: Notify on every approval rejected as wasteful; only notify when step actually advances
read c3-205: requestApprovals notifies first-step approvers; approvePr notifies next approvers if step advances
read ref-approval-chain: mode anyof = first approver advances; allof = every approver must sign
read c3-211: notificationService.notifyNextApprovers publishes one notification per recipient
read ref-pull-dispatcher: channels self-register by subscribe()
read ref-sync: flows send sync.ack(executionId); NATS broadcast updates clients
```

## CROSSCUT-NOTIFICATION-BELL-1: visible notification path

Question: A notification was published but a user cannot see it. What path must you trace before blaming the UI?

Grounding commands:

```bash
c3 search "notification system NATS websocket sync async approval"
c3 read c3-205 --full
c3 read c3-211 --full
c3 read ref-sync --full
c3 read c3-101 --full
c3 read adr-20260126-user-notification-ui --full
c3 read recipe-realtime-sync --full
c3 graph ref-sync --depth 1
c3 graph c3-211 --depth 1
```

Expected trace:

| Segment | Expected ids and facts |
| --- | --- |
| Action / command location | `c3-205` PR Flows triggers notificationService from approval flows such as `requestApprovals` or step-advancing `approvePr`/`approveAll`. |
| Sync mechanism | `ref-sync` defines two NATS subjects: `sync.broadcast` for deltas/acks and `sync.user.{escaped_email}` for user-specific notifications. |
| Frontend receiver | `c3-101` `natsSync` subscribes to both `sync.broadcast` and `sync.user.<email>`; notification messages update the `notifications` atom. |
| Notification target/model | `c3-211` Notification System publishes/dispatches in-app notifications and records notification logs. |
| UI decision/history | `adr-20260126-user-notification-ui` added the notification message type, user subject subscription, notifications atom, and bell dropdown because broadcast-only sync hid delivered notifications. |
| Emergent property | Sync and notifications share NATS but are separate: broadcast sync is ephemeral/global; notifications are durable/targeted through JetStream and user subjects. |

Evidence snippets:

```text
read adr-20260126-user-notification-ui: c3-211 publishes in-app notifications to sync.user.{email}, but frontend only subscribed to sync.broadcast
read adr-20260126-user-notification-ui: extend natsSync to subscribe to sync.broadcast and sync.user.{escaped_email}
read c3-101: natsSync subscribes to sync.broadcast (delta) and sync.user.<email> (notifications)
read ref-sync: {prefix}.broadcast = deltas + acks; {prefix}.user.{escaped_email} = notifications to specific user
read recipe-realtime-sync: notification system uses separate NATS JetStream stream, independently from broadcast sync
read c3-211: inAppChannel delivers by NATS publish real-time plus JetStream persistence
```

## CROSSCUT-SLACK-APPROVAL-1: Slack action consistency

Question: When approval happens from Slack, what keeps web clients and next approvers consistent?

Grounding commands:

```bash
c3 search "Slack approve reject payment request notification sync"
c3 read c3-215 --full
c3 read adr-20260305-slack-bot-integration --full
c3 read c3-205 --full
c3 read c3-202 --full
c3 read c3-211 --full
c3 read ref-pull-dispatcher --full
c3 read ref-sync --full
c3 read c3-101 --full
c3 graph c3-215 --depth 1
c3 graph c3-205 --depth 1
c3 graph c3-211 --depth 1
```

Expected trace:

| Segment | Expected ids and facts |
| --- | --- |
| Action / command location | `c3-215` Slack Bot Integration handles `approve_pr` / `reject_pr` actions, resolves Slack user identity, builds execution context, then calls `prFlows.approvePr` or `prFlows.rejectPr`. |
| Execution context | `c3-202` supplies `currentUserTag`, `UserActor`, `transactionTag`, and `executionIdTag` concepts used when a non-HTTP Slack action enters backend flows. |
| State change | `c3-205` owns PR mutation and approval-chain side effects; Slack must call flows, not lower-level services. |
| Sync mechanism | `ref-sync` keeps web clients consistent via NATS WebSocket deltas/acks; `c3-101` receives those on `sync.broadcast`. |
| Notification target/model | `c3-211` sends next-approver notifications; Slack also registers as a notification channel through the dispatcher. |
| Notification mechanism | `ref-pull-dispatcher` governs channel self-registration; `adr-20260305-slack-bot-integration` adds `slackChannel` and says inbound actions call flows because flows include notification + sync side effects. |
| Emergent property | Flow entry preserves cross-cutting side effects. Calling the service directly from Slack would bypass sync/ack and notification-on-step-advance behavior. |

Evidence snippets:

```text
read c3-215: slackActions handles approve_pr and reject_pr, then executes prFlows.approvePr or prFlows.rejectPr
read adr-20260305-slack-bot-integration: Calls flows (not services) -- same as server functions do
read adr-20260305-slack-bot-integration: flows handle service call + notification-on-step-advance + sync ack
read c3-215: outbound notifications: notificationDispatcher -> slackChannel -> bot.openDM -> approvalRequestCard
read c3-202: currentUserTag carries UserActor; executionIdTag is optional execution id tag
read ref-sync: flows call sync.ack(executionId) at the end; clients resolve executionTracker
read c3-101: sync.broadcast updates prs/invoices/payments stores; sync.user.<email> updates notifications
read ref-pull-dispatcher: channels self-register with dispatcher.subscribe()
```
