---
id: c3-211
c3-version: 3
c3-seal: a3b15aafa84cce3203c795ebff690b3fe0c70b61606159e4ac67ad6e6f13b56a
title: Notification System
type: component
category: foundation
parent: c3-2
goal: Publisher/dispatcher pattern for async notifications via NATS JetStream with pluggable channels
uses:
    - ref-pull-dispatcher
    - ref-pumped-fn
    - ref-scope-controlled-config
    - ref-structured-logging
---

# Notification System

## Goal

Publisher/dispatcher pattern for async notifications via NATS JetStream with pluggable channels

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-2 |
| Role | Own Notification System behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep Notification System decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for Notification System so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before Notification System behavior is changed. | ref-pull-dispatcher |
| Inputs | Accept only the files, commands, data, or calls that belong to Notification System ownership. | ref-pull-dispatcher |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | ref-pull-dispatcher |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | ref-pull-dispatcher |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks Notification System to deliver its documented responsibility. | ref-pull-dispatcher |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | ref-pull-dispatcher |
| Alternate paths | When a request falls outside Notification System ownership, hand it to the parent or sibling component. | ref-pull-dispatcher |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | ref-pull-dispatcher |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-pull-dispatcher | ref | Governs Notification System behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| Notification System input | IN | Callers must provide context that matches the component goal and parent fit. | c3-2 boundary | c3x lookup plus targeted tests or review. |
| Notification System output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-2 boundary | c3x check and project test suite. |

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

- NATS JetStream — persistent message queue
- `@pumped-fn/lite` — atoms, controllers, tags
- `notificationLog` / `notificationPreferences` tables — dispatch history and user settings

## Architecture

```
Flow → notificationService → notificationPublisher → [JetStream NOTIFICATIONS]
                                                          ↓
                                               notificationDispatcher
                                              ↙                    ↘
                                    emailChannel              inAppChannel
                                        ↓                         ↓
                                    SMTP Server         NATS (real-time) + JetStream (persistence)
```

## Components

### notificationService

Orchestrates notifications from flows. Two methods:

- `notifyNextApprovers(execCtx, prId)` — looks up next approvers from PR data, publishes one notification per recipient
- `retryNotification(execCtx, logId)` — republishes a failed notification from the log

### notificationPublisher

Publishes to JetStream. Creates the `NOTIFICATIONS` stream on init if missing.

- Subject pattern: `notifications.{type}.{escaped_email}`
- Stream config: Workqueue retention, file storage, 7-day max age, 10K max messages

### notificationDispatcher

Consumes from JetStream with a durable consumer. For each message:

1. Parses the notification
2. Fetches user's preferred channels from `notification_preferences`
3. Creates a `pending` log entry per channel
4. Calls the channel handler
5. Updates log to `sent` or `failed`
6. Acks on success, naks on failure (retry)

Channels self-register via `dispatcher.subscribe({ channel, handler })`. This inverts the dependency — channels depend on the dispatcher, not vice versa.

### notificationController

Lazy-loads required channels on startup using `controller()` wrappers. Validates all required channels are subscribed; throws if any are missing.

## Built-in Channels

| Channel | Name | Delivery |
| --- | --- | --- |
| emailChannel | email | SMTP with HTML template |
| inAppChannel | in_app | NATS publish (real-time) + JetStream (persistence) |
| slackChannel | slack | DM via Slack bot (see c3-215) |

All channels subscribe to the dispatcher in their factory and unsubscribe on cleanup. `slackChannel` requires `slackConfigTag` — gracefully skips if Slack is not configured (no user mapping or bot token).

## User Preferences

`notification_preferences` table stores per-user channel list (JSONB, defaults to `['in_app']`). The dispatcher filters channels against preferences before dispatching.

## Notification Log

`notification_log` table tracks every dispatch attempt with status (`pending`/`sent`/`failed`), channel, payload, and error details. Powers admin UI retry and monitoring.

## Adding a New Channel

1. Create channel atom with `notificationDispatcher` as dependency
2. Tag with `notificationChannelTag('channel_name')`
3. Call `dispatcher.subscribe()` in factory, register cleanup
4. Add `controller(newChannel)` to `notificationController` deps
5. Add to controllers map in factory

## Adding a New Notification Type

1. Add type to `NotificationPayload` union in `@acountee/shared`
2. Create builder method in `notificationService`
3. Add email template in `emailChannel`
