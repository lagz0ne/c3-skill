---
id: c3-215
c3-version: 3
c3-seal: 7559219b8e771c925e55483981060e1d76053dd66aa4db042275f67a903c3c70
title: Slack Bot Integration
type: component
category: feature
parent: c3-2
goal: Slack bot for PR approval notifications, interactive approve/reject actions, and /pending command
uses:
    - ref-pumped-fn
    - ref-query-services
    - ref-structured-logging
---

# Slack Bot Integration

## Goal

Slack bot for PR approval notifications, interactive approve/reject actions, and /pending command

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-2 |
| Role | Own Slack Bot Integration behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep Slack Bot Integration decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for Slack Bot Integration so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before Slack Bot Integration behavior is changed. | ref-pumped-fn |
| Inputs | Accept only the files, commands, data, or calls that belong to Slack Bot Integration ownership. | ref-pumped-fn |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | ref-pumped-fn |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | ref-pumped-fn |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks Slack Bot Integration to deliver its documented responsibility. | ref-pumped-fn |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | ref-pumped-fn |
| Alternate paths | When a request falls outside Slack Bot Integration ownership, hand it to the parent or sibling component. | ref-pumped-fn |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | ref-pumped-fn |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-pumped-fn | ref | Governs Slack Bot Integration behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| Slack Bot Integration input | IN | Callers must provide context that matches the component goal and parent fit. | c3-2 boundary | c3x lookup plus targeted tests or review. |
| Slack Bot Integration output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-2 boundary | c3x check and project test suite. |

## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| Contract drift | Goal, boundary, or derived material changes without matching component docs. | Compare Goal, Parent Fit, Contract, and Derived Materials. | Run c3x check and relevant project tests. |
| Governance drift | Cited references, rules, or parent responsibilities change. | Re-read Governance rows and parent container docs. | Run c3x verify plus targeted lookup for changed files. |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| Code, docs, tests, prompts | Goal, Governance, Contract, and Change Safety sections. | Names and framework shape may vary; behavior and boundaries may not. | c3x check, c3x verify, and relevant tests. |

## Architecture Details

## Dependencies

- `chat` + `@chat-adapter/slack` — bot framework and Slack adapter
- `@pumped-fn/lite` — atoms, tags, service
- `slackConfigTag` — optional config; all atoms gracefully return `null` when unconfigured

## Architecture

```
Outbound (notifications):
  notificationDispatcher → slackChannel → bot.openDM → approvalRequestCard

Inbound (commands/actions):
  POST /api/slack/events → bot.webhooks.slack
    → slackCommands (/pending → list PRs pending user approval)
    → slackActions (approve_pr / reject_pr → prFlows.approvePr / rejectPr)
```

## Components

### slackBot

Core `Chat` instance with `SlackAdapter`. Uses in-memory `StateAdapter` for conversation state. Tagged with `slackConfigTag` — returns `null` if Slack is not configured. Registers cleanup on context teardown.

### slackChannel

Notification channel that subscribes to `notificationDispatcher`. For each notification:

1. Looks up recipient's Slack user via `slackQueries.getSlackUserByEmail`
2. Opens a DM thread via `bot.openDM`
3. Sends an `approvalRequestCard` with Approve/Reject buttons
4. Returns `ChannelDispatchResult`

Skips silently if no Slack user mapping exists.

### slackCommands

Registers `/pending` slash command. Lists PRs where the Slack user is the next approver:

1. Resolves Slack user → Acountee email via `slackQueries.getSlackUserById`
2. Fetches all PRs via `prFlows.listPrs`
3. Filters PRs with `status=requested` where `next_approver.users` includes the user
4. Sends `pendingPrCard` per PR (or `noPendingCard` if none)

### slackActions

Handles `approve_pr` and `reject_pr` interactive button actions:

1. Resolves Slack user → Acountee user
2. Builds `UserActor` with permissions and team capabilities
3. Sets `currentUserTag` and `transactionTag` on execution context
4. Executes `prFlows.approvePr` or `prFlows.rejectPr` within a DB transaction
5. Sends `resultCard` with outcome

### slackCards

JSX-like card builders using `chat` primitives (`Card`, `Fields`, `Actions`, `Button`):

| Card | Purpose |
| --- | --- |
| approvalRequestCard | Notification DM with PR details + Approve/Reject buttons |
| pendingPrCard | Per-PR card for /pending command |
| noPendingCard | Empty state for /pending |
| resultCard | Action outcome feedback |

### slackQueries

Query service for `slack_user_links` table:

| Method | Purpose |
| --- | --- |
| getSlackUserByEmail | Look up Slack user by Acountee email |
| getSlackUserById | Look up Acountee user by Slack user ID |
| upsertSlackUser | Create or update Slack ↔ Acountee mapping |
| deleteSlackUser | Remove user mapping |

## Webhook Route

`POST /api/slack/events` — TanStack Router server handler. Resolves `slackBot` from scope, delegates to `bot.webhooks.slack`. Returns 503 if Slack is not configured. Background tasks run via `Promise.allSettled`.

## Execution Context in Inbound Handlers

Slack inbound handlers (commands, actions) create their own execution context since there's no HTTP request/cookie context. The `slackActions` handler manually:

- Creates `scope.createContext({})`
- Sets `currentUserTag` with a constructed `UserActor`
- Sets `transactionTag` within a `db.transaction()` call
- Sets `app.current_user` via SQL `set_config` for audit triggers

## Uses

| Component | For |
| --- | --- |
| c3-211 Notification System | slackChannel subscribes to notificationDispatcher |
| c3-205 PR Flows | prFlows.approvePr, prFlows.rejectPr, prFlows.listPrs |
| c3-202 Execution Context | currentUserTag, transactionTag for inbound actions |
| c3-204 Drizzle ORM | Direct drizzleDb for transaction wrapping |

## Dependencies

| Direction | What | From/To |
| --- | --- | --- |
| IN (uses) | notificationDispatcher, prFlows, drizzleDb, slackQueries, userQueries, teamQueries | c3-2 |
| OUT (provides) | Slack notification channel, interactive PR approval from Slack | c3-2 |
