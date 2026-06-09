---
id: adr-20260121-notification-system
c3-seal: bdc2cd00aa21b2076d712a3ed81f06988820482875f0dbf52ec1ad647df2b399
title: NATS-Based Notification System for Approval Workflows
type: adr
goal: 'Document and implement the architectural decision: NATS-Based Notification System for Approval Workflows.'
status: implemented
date: "2026-01-21"
affects:
    - c3-2
    - c3-204
    - c3-205
    - c3-4
approved-files:
    - apps/start/src/server/dbs/schema.ts
    - apps/start/src/server/dbs/queries/slack.ts
    - apps/start/src/server/dbs/queries/index.ts
    - apps/start/src/server/dbs/queries/types.ts
    - apps/e2e/fixtures/test-database.sql
    - apps/start/src/server/resources/notificationPublisher.ts
    - apps/start/src/server/resources/notificationDispatcher.ts
    - apps/start/src/server/resources/emailTransport.ts
    - apps/start/src/server/services/notification.ts
    - apps/start/src/server/services/pr.ts
    - apps/start/src/server/flows/pr.ts
    - apps/start/src/server/tags.ts
    - apps/start/src/server.tsx
    - apps/start/src/server/dbs/queries/notification.ts
    - infra/nats.conf
    - infra/nats-production.conf
    - drizzle/0011_drop_slack_tables.sql
    - drizzle/0012_create_notification_tables.sql
---

# NATS-Based Notification System for Approval Workflows

## Goal

Document and implement the architectural decision: NATS-Based Notification System for Approval Workflows.

## Status

**Implemented** - 2026-01-21

## Problem

Users who are next in the approval chain have no visibility into pending PRs that need their action unless they actively check the application. This creates approval delays and workflow friction.

The codebase has remnant Slack integration code at the database layer (tables: `slack_users`, `slack_notifications`, `slack_interactions`, `pr_slack_threads`) and query layer (`slackQueries`), but:

- No actual Slack API client exists
- No flows invoke the Slack queries
- The infrastructure is prepared but never implemented

This dead code adds maintenance burden and false expectations. Meanwhile, the project already has a mature NATS infrastructure for real-time sync that could be leveraged for notifications.

## Decision

### 1. Remove Slack Infrastructure Completely

Drop all Slack-related code and database tables:

- Drop tables: `slack_users`, `slack_notifications`, `slack_interactions`, `pr_slack_threads`
- Remove `slackQueries` service
- Remove exports from `queries/index.ts` and `queries/types.ts`
- Update test fixtures

### 2. Build Pluggable Notification System on NATS

Implement a notification system with NATS as the reliable queue backbone:

```
┌─────────────────────────────────────────────────────────────────┐
│                        PR Flow                                   │
│  (requestApprovals, approve, reject, etc.)                      │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                  Notification Publisher                          │
│  Publishes to NATS: notifications.{type}.{recipient}            │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                     NATS JetStream                               │
│  • Durable queue for guaranteed delivery                         │
│  • Work queue pattern for dispatch workers                       │
│  • Retry with backoff on failure                                │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                 Notification Dispatcher                          │
│  Consumes from NATS, routes to channel handlers                 │
└────────┬─────────────────┬─────────────────────┬────────────────┘
         │                 │                     │
         ▼                 ▼                     ▼
   ┌───────────┐    ┌───────────┐        ┌───────────┐
   │  In-App   │    │   Email   │        │  Future:  │
   │  (WebSocket)│  │  (SMTP)   │        │  Slack,   │
   │  (via NATS)│  │  (Future) │        │  Teams... │
   └───────────┘    └───────────┘        └───────────┘
```

### 3. Start with "Awaiting Approval" Notifications Only

Initial scope is narrow:

- Trigger: `requestApprovals` flow transitions PR from `draft` → `requested`
- Recipients: Users in the **next approval step** only (not entire chain)
- Channel: In-app notification via existing NATS WebSocket sync

### 4. Database Schema for Notifications

```sql
-- Notification preferences per user
CREATE TABLE notification_preferences (
    user_email TEXT PRIMARY KEY REFERENCES users(email) ON DELETE CASCADE,
    channels JSONB NOT NULL DEFAULT '["in_app"]',
    -- Future: quiet_hours, per-type preferences
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Notification log for audit and debugging
CREATE TABLE notification_log (
    id SERIAL PRIMARY KEY,
    recipient_email TEXT NOT NULL REFERENCES users(email) ON DELETE CASCADE,
    notification_type TEXT NOT NULL,
    channel TEXT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'failed')),
    error_details TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    sent_at TIMESTAMP
);
CREATE INDEX idx_notification_log_recipient ON notification_log(recipient_email);
CREATE INDEX idx_notification_log_status ON notification_log(status) WHERE status = 'pending';
```

## Rationale

| Considered | Rejected Because |
| --- | --- |
| Keep Slack tables for future use | Dead code, misleading, adds complexity. Can recreate when actually needed. |
| Direct email/webhook dispatch from flow | No durability, no retry, blocks flow execution |
| Separate notification microservice | Overkill for current scale, NATS provides the backbone |
| Notify all approvers at once | Information overload, only next step is actionable |
| Poll-based notification check | Already have WebSocket infrastructure via NATS |

### Why NATS JetStream?

1. **Already deployed** - NATS is running for real-time sync
2. **Durable delivery** - JetStream provides guaranteed delivery with ack
3. **Work queue pattern** - Multiple dispatchers can consume, natural scaling
4. **Retry with backoff** - Built-in for failed deliveries
5. **Observability** - NATS monitoring shows queue depth, consumer lag

### Why In-App First?

1. **Zero external dependencies** - Uses existing NATS WebSocket connection
2. **Immediate feedback** - No email delays or deliverability issues
3. **Foundation for more** - Same infrastructure extends to email, Slack, etc.

## Affected Layers

| Layer | Document | Change |
| --- | --- | --- |
| c3-2 | API Backend | Add notification publisher, dispatcher, service |
| c3-204 | Drizzle ORM | Drop Slack tables, add notification tables |
| c3-205 | PR Flows | Hook requestApprovals to emit notification event |
| c3-4 | NATS Server | Add JetStream stream for notifications |

## Implementation Approach

### Phase 1: Slack Removal

1. Create migration to drop Slack tables
2. Remove `slackQueries` service
3. Update exports and types
4. Update test fixtures

### Phase 2: Notification Infrastructure

1. Create `notificationPublisher` resource (publishes to NATS)
2. Create JetStream stream config for `notifications.>`
3. Create `notificationDispatcher` resource (consumes from NATS)
4. Create channel handlers (start with in-app via `sync.broadcast`)

### Phase 3: Integration

1. Create `notificationService` with `notifyNextApprovers(prId)` method
2. Hook into `requestApprovals` **flow** (not service) - follows C3 pattern where flows orchestrate side effects
3. Create notification preferences table and default all users to `in_app`

Integration point (in `apps/start/src/server/flows/pr.ts`):

```typescript
// requestApprovals flow
await ctx.exec({ fn: prService.requestForApprovals, params: [prId] })
await ctx.exec({ fn: notificationService.notifyNextApprovers, params: [prId] })  // NEW
```

### Phase 4: Frontend (Optional, Separate ADR)

1. Add notification indicator/badge
2. Add notification dropdown/panel
3. Mark as read functionality

## NATS JetStream Configuration

JetStream must be enabled in NATS config files:

```conf
# Add to infra/nats.conf and infra/nats-production.conf
jetstream {
    store_dir: /data/jetstream
    max_mem: 256MB
    max_file: 1GB
}
```

**Note:** Email addresses in subjects are escaped (replace `@` and `.` with `_`) for NATS compatibility, consistent with existing `natsPublisher.ts` pattern.

## NATS Subject Design

```
notifications.approval_required.{recipient_email}
notifications.approval_complete.{maker_email}
notifications.pr_rejected.{maker_email}
```

JetStream stream: `NOTIFICATIONS`

- Subjects: `notifications.>`
- Retention: WorkQueue (each message delivered to exactly one consumer)
- Max age: 7 days
- Replicas: 1 (single node for now)

## Notification Payload

```typescript
interface ApprovalRequiredNotification {
  type: 'approval_required'
  pr_id: number
  pr_name: string
  maker: string
  amount: string | null
  step_name: string
  requested_at: string
}
```

## Approved Files

The following files are approved for modification under this ADR:

```yaml
approved-files:
  # Slack removal
  - apps/start/src/server/dbs/schema.ts
  - apps/start/src/server/dbs/queries/slack.ts
  - apps/start/src/server/dbs/queries/index.ts
  - apps/start/src/server/dbs/queries/types.ts
  - apps/e2e/fixtures/test-database.sql
  # New notification infrastructure
  - apps/start/src/server/resources/notificationPublisher.ts
  - apps/start/src/server/resources/notificationDispatcher.ts
  - apps/start/src/server/services/notification.ts
  - apps/start/src/server/services/pr.ts
  - apps/start/src/server/flows/pr.ts
  # Database migrations
  - migrations/xxx-drop-slack-tables.sql
  - migrations/xxx-create-notification-tables.sql
```

**Gate behavior:** Only these files can be edited when status is `accepted`.

## Verification

- [ ] Slack tables dropped (check `\dt` in psql shows no slack_* tables)
- [ ] `slackQueries` removed (grep returns no results)
- [ ] No broken imports (TypeScript compiles clean)
- [ ] JetStream stream created for `notifications.>`
- [ ] `requestApprovals` flow emits notification event
- [ ] Users in next approval step receive in-app notification
- [ ] Notification appears in browser via NATS WebSocket
- [ ] E2E tests pass with updated fixtures
- [ ] Notification log records delivery attempts

## Context

N.A - historical ADR; original context is captured in the git log around the ADR date and in the current code that implements the decision.

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| N.A - historical | Shipped via git commits; the c3 topology and code-map reflect the resulting structure. | c3x list --include-adr and git log around the ADR date |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| N.A - historical | Current .c3 entities, refs, and code-map are the post-change state. | c3x verify and c3x check |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| N.A - historical | Enforcement is implicit in the currently linked components and refs. | c3x graph and cited ref ids on the relevant components |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| N.A - historical | Alternatives were considered at decision time; rationale is preserved in the original commit message or branch discussion. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| N.A - historical | Risks were assessed pre-merge; the decision has since shipped without outstanding incidents tied to this ADR. | git log and project test suite |
