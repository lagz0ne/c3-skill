---
id: adr-20260305-slack-bot-integration
c3-seal: 3d38f1e94f163fdad75d1df0e843aee323bdbaf59df4364b6438dc57c4cc83be
title: Slack Bot Integration with chat-sdk
type: adr
goal: 'Add Slack as a notification channel AND interactive bot: receive approval cards, view details, approve/reject directly from Slack, query pending work via slash commands.'
status: implemented
date: "2026-03-05"
affects:
    - c3-2
    - c3-211
    - c3-214
approved-files:
    - apps/start/src/server/resources/slackBot.ts
    - apps/start/src/server/resources/slackChannel.ts
    - apps/start/src/server/resources/slackActions.ts
    - apps/start/src/server/resources/slackCommands.ts
    - apps/start/src/server/resources/slackCards.ts
    - apps/start/src/server/resources/notificationController.ts
    - apps/start/src/routes/api.slack.events.tsx
    - apps/start/src/server/functions/devChat.ts
    - apps/start/src/components/DevChatPanel.tsx
    - apps/start/src/components/DevShell.tsx
    - apps/start/src/server/dbs/queries/slack.ts
    - apps/start/src/server/tags.ts
    - apps/start/src/server/dbs/schema.ts
    - drizzle/XXXX_slack_user_links.sql
    - apps/start/package.json
---

# Slack Bot Integration with chat-sdk

## Goal

Add Slack as a notification channel AND interactive bot: receive approval cards, view details, approve/reject directly from Slack, query pending work via slash commands.

## Problem

Users get approval notifications in-app and via email but can't act without switching to the web app. Slack is the team's primary tool -- actionable notifications there reduce approval latency.

Previous Slack integration (removed in ADR-20260121) was dead infrastructure. This time: proven SDK, proper implementation.

## Decision

### 1. chat-sdk with Slack Adapter

```
pnpm add chat @chat-adapter/slack
```

Single-workspace mode. Env vars: `SLACK_BOT_TOKEN`, `SLACK_SIGNING_SECRET`.

### 2. Architecture

**Outbound** (notifications): dispatcher -> `slackChannel` -> `slackBot.post(card)` to user's DM

**Inbound** (interactions): Slack webhook -> `/api/slack/events` -> `bot.webhooks.slack(request)` -> `onAction` / `onSlashCommand` -> execute flows

The `slackBot` resource atom is the single `Chat` instance shared by both directions.

### 3. Resource Atoms

#### slackBot -- chat-sdk resource

```typescript
export const slackBot = atom({
  deps: { config: tags.optional(slackConfigTag), rootLogger },
  factory: async (ctx, { config, rootLogger }) => {
    if (!config) return null  // graceful skip when no Slack credentials

    const bot = new Chat({
      userName: 'acountee',
      adapters: { slack: createSlackAdapter() },
    })
    // initialize() is auto-called on first webhook -- no eager init needed
    ctx.cleanup(() => bot.shutdown())
    return bot
  },
})
```

#### slackChannel -- outbound notification channel

Same pattern as `emailChannel` / `inAppChannel`:

```typescript
export const slackChannel = atom({
  deps: { bot: slackBot, dispatcher: notificationDispatcher, slackQueries },
  tags: [notificationChannelTag('slack')],
  factory: (ctx, { bot, dispatcher, slackQueries }) => {
    if (!bot) return { name: 'slack', enabled: false }

    const handler = async (notification, log, logger) => {
      const link = await slackQueries.getSlackUserByEmail(notification.recipient_email)
      if (!link) return { success: false, channel: 'slack', error: 'No Slack link' }

      const { payload } = notification
      // openDM takes a Slack user ID, not a channel ID
      const dm = await bot.openDM(link.slack_user_id)
      await dm.post(<ApprovalCard payload={payload} logId={log.id} />)

      return { success: true, channel: 'slack' }
    }

    const unsubscribe = dispatcher.subscribe({ channel: 'slack', handler })
    ctx.cleanup(unsubscribe)
    return { name: 'slack', enabled: true }
  },
})
```

#### slackActions -- inbound interactive handlers

Calls **flows** (not services) -- same as server functions do. Flows handle the full orchestration: service call + notification-on-step-advance + sync ack.

```typescript
export const slackActions = atom({
  deps: { bot: slackBot, slackQueries, scope },
  factory: (_ctx, { bot, slackQueries, scope }) => {
    if (!bot) return null

    bot.onAction(['approve_pr', 'reject_pr'], async (event) => {
      // 1. Resolve app user from Slack user ID
      const link = await slackQueries.getSlackUserByEmail(event.user.id)
      if (!link) { await event.thread.post('Your Slack account is not linked.'); return }

      // 2. Create execution context with user identity (mirrors transactionMiddleware)
      const execContext = scope.createContext({
        [currentUserTag]: { email: link.user_email },
        [executionIdTag]: crypto.randomUUID(),
      })

      try {
        // 3. Execute the FLOW (not service) -- includes notification + sync side effects
        const flowToRun = event.actionId === 'approve_pr' ? prFlows.approvePr : prFlows.rejectPr
        const result = await execContext.exec({ flow: flowToRun, rawInput: { pr_id: Number(event.value) } })

        // 4. Update the Slack card
        const status = result.success ? 'Approved' : `Failed: ${result.reason}`
        await event.thread.post(`PR #${event.value} -- ${status} by ${link.user_email}`)
      } finally {
        await execContext.close()
      }
    })
  },
})
```

#### slackCommands -- slash command handlers

```typescript
// Registers bot.onSlashCommand("/pending")
// 1. Resolves app user from Slack user ID
// 2. Queries pending PRs for user via prQueries
// 3. Responds with card list (ephemeral)
```

#### slackCards -- JSX card templates

```tsx
/** @jsxImportSource chat */

export function ApprovalCard({ payload, logId }) {
  return (
    <Card title={`Approval Required: ${payload.pr_name}`}>
      <Fields>
        <Field label="Created by" value={payload.maker} />
        <Field label="Amount" value={payload.amount ?? 'N/A'} />
        <Field label="Step" value={payload.step_name} />
      </Fields>
      <Actions>
        <Button id="approve_pr" value={String(payload.pr_id)} style="primary">Approve</Button>
        <Button id="reject_pr" value={String(payload.pr_id)} style="danger">Reject</Button>
      </Actions>
    </Card>
  )
}
```

### 4. Webhook Route

```typescript
export const Route = createFileRoute('/api/slack/events')({
  server: {
    handlers: {
      POST: async ({ request, context: { scope } }) => {
        const bot = await scope.resolve(slackBot)
        if (!bot) return new Response('Slack not configured', { status: 503 })

        // waitUntil: chat-sdk needs this for async work after 200 response.
        // TanStack Start has no built-in waitUntil -- use fire-and-forget promise collector.
        const pending: Promise<unknown>[] = []
        const response = await bot.webhooks.slack(request, {
          waitUntil: (task) => { pending.push(task) },
        })

        // Don't await pending -- Slack requires <3s response.
        // Tasks (flow execution, card updates) run after response is sent.
        if (pending.length) Promise.allSettled(pending).catch(() => {})

        return response
      },
    },
  },
})
```

### 5. User Linking

```sql
CREATE TABLE slack_user_links (
  slack_user_id TEXT PRIMARY KEY,
  slack_team_id TEXT NOT NULL,
  user_email TEXT NOT NULL REFERENCES users(email) ON DELETE CASCADE,
  slack_dm_channel_id TEXT,
  linked_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_slack_user_links_email ON slack_user_links(user_email);
```

Auto-link on first interaction: fetch Slack user's email via `users:read.email` scope, match to app user.

### 6. Notification Controller Update

Current controller has a static map of `{ in_app: inAppCtrl, email: emailCtrl }` + `tags.required(requiredChannels)`. To add Slack as optional:

```typescript
// Before:
const controllers = { in_app: inAppCtrl, email: emailCtrl }

// After:
deps: {
  // ... existing
  slackCtrl: controller(slackChannel),
},
factory: async (_ctx, { ..., slackCtrl }) => {
  const controllers = { in_app: inAppCtrl, email: emailCtrl, slack: slackCtrl }
  // requiredChannels defaults to [] -- Slack is never "required"
  // But if slackChannel is enabled, it self-registers with dispatcher via subscribe()
  // Controller validates required channels; Slack registers independently
}
```

The key: `slackChannel.subscribe()` registers with the dispatcher regardless of `requiredChannels`. The dispatcher routes to all subscribed channels that match user preferences. No controller change needed for dispatch -- only for lifecycle/health tracking.

### 7. Dev Chat Panel

Dev-only slide-up panel in DevShell (same pattern as the Logs panel, `c` to toggle). Tests bot interactions without Slack.

**Server function** -- calls flows directly, same as `apps/start/src/server/functions/pr.ts`:

```typescript
// apps/start/src/server/functions/devChat.ts (DEV ONLY)
export const devChatAction = createServerFn({ method: 'POST' })
  .middleware([transactionMiddleware])
  .inputValidator((data: { type: 'command' | 'action'; command?: string; actionId?: string; value?: string }) => data)
  .handler(async ({ data, context: { execContext } }) => {
    if (data.type === 'command' && data.command === '/pending') {
      const prQueries = await execContext.scope.resolve(prQueriesAtom)
      const prs = await execContext.exec({ fn: prQueries.listPrs, params: [{ status: 'requested' }] })
      return { type: 'card_list' as const, prs }
    }
    if (data.type === 'action' && data.actionId === 'approve_pr') {
      const result = await execContext.exec({ flow: prFlows.approvePr, rawInput: { pr_id: Number(data.value) } })
      return { type: 'result' as const, ...result }
    }
    if (data.type === 'action' && data.actionId === 'reject_pr') {
      const result = await execContext.exec({ flow: prFlows.rejectPr, rawInput: { pr_id: Number(data.value) } })
      return { type: 'result' as const, ...result }
    }
    return { type: 'error' as const, message: 'Unknown command' }
  })
```

**DevChatPanel** -- React component rendering message list + cards with action buttons. Subscribes to NATS for real-time incoming notification cards.

## Work Breakdown

### Phase 1: Infrastructure

Install deps, `slackBot` atom, `slackConfigTag`, webhook route, `slack_user_links` migration

### Phase 2: Outbound

`slackCards.tsx`, `slackChannel`, controller update, user preference migration

### Phase 3: Inbound

`slackActions` (approve/reject via flows), `slackCommands` (`/pending`), auto-linking

### Phase 4: Dev Chat

`devChatAction` server fn, `DevChatPanel` component, DevShell integration

## Rationale

| Considered | Rejected Because |
| --- | --- |
| Bolt.js (official Slack SDK) | Lower-level, no JSX cards, no cross-platform potential |
| Direct Slack API calls | Manual webhook parsing, signature verification |
| Separate microservice | Same process shares DI context, simpler testing |
| Webhook-only (no bot) | Can't do interactive actions from messages |

## Open Questions

- **Execution context in Slack inbound**: Need to replicate what `transactionMiddleware` does (user resolution, DB transaction wrapping) without HTTP cookies. May need a `slackExecContextFactory` utility.
- **Notification preferences migration**: Existing users default to `['in_app']`. Need migration or admin UI to opt users into `'slack'`.

## Verification

- [ ] App starts clean without `SLACK_BOT_TOKEN` (no crash)
- [ ] With token: webhook route returns 200 for Slack challenge
- [ ] Notification to Slack-linked user delivers card with Approve/Reject buttons
- [ ] Approve button executes `approvePr` flow, card updates
- [ ] Reject button executes `rejectPr` flow, card updates
- [ ] `/pending` lists PRs awaiting user's approval
- [ ] Auto-link resolves Slack user by email
- [ ] Dev chat panel: `/pending`, Approve, Reject all work
- [ ] Dev chat panel not bundled in production
- [ ] TypeScript compiles clean

## Context

N.A - historical ADR; original context is captured in the git log around the ADR date and in the current code that implements the decision.

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
