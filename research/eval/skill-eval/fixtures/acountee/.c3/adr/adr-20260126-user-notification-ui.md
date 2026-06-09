---
id: adr-20260126-user-notification-ui
c3-seal: ca732baabf0753354a4d4b656bf2fa1cc62a006b13b92fa8e551effc5cac3552
title: User-Facing Notification UI with Bell Icon Dropdown
type: adr
goal: 'Document and implement the architectural decision: User-Facing Notification UI with Bell Icon Dropdown.'
status: implemented
date: "2026-01-26"
affects:
    - c3-1
    - c3-101
    - c3-211
approved-files:
    - apps/start/src/lib/pumped/types.ts
    - apps/start/src/lib/pumped/atoms/natsSync.ts
    - apps/start/src/lib/pumped/atoms/notifications.ts
    - apps/start/src/lib/pumped/index.ts
    - apps/start/src/components/NotificationBell.tsx
    - apps/start/src/components/AppSidebar.tsx
---

# User-Facing Notification UI with Bell Icon Dropdown

## Goal

Document and implement the architectural decision: User-Facing Notification UI with Bell Icon Dropdown.

## Status

**Implemented** - 2026-01-26

## Problem

The notification system (c3-211) publishes in-app notifications to user-specific NATS subjects (`sync.user.{email}`), but the frontend only subscribes to `sync.broadcast`. Users have no way to see notifications sent to them unless they are admins viewing the notification log.

This creates a gap: notifications are being sent but users can't see them.

## Decision

### 1. Extend SyncMessage Type

Add `notification` variant to `SyncMessage` union in `types.ts`:

```typescript
export type SyncMessage =
  | { type: 'full'; ... }
  | { type: 'delta'; ... }
  | { type: 'ack'; ... }
  | { type: 'notification'; notification: InAppNotification; executionId?: string }
```

### 2. Add User-Subject Subscription to natsSync

Extend `natsSync.ts` to subscribe to both subjects:

- `sync.broadcast` (existing - for data deltas)
- `sync.user.{escaped_email}` (new - for user-specific notifications)

Subject naming uses the same escaping as server: `email.replace(/[@.]/g, '_')`.

### 3. Create Notifications Atom

New `notifications` atom with:

- `items: InAppNotification[]` - received notifications (max 50, FIFO cleanup)
- Derived `unreadCount` - computed from items where `read: false`
- Actions via controller: add, markAsRead, markAllAsRead, dismiss
- localStorage persistence for read IDs across page refresh

### 4. Add NotificationBell Component

Bell icon in AppSidebar with:

- Unread count badge (red dot with number)
- Popover dropdown showing recent notifications
- Click notification to expand details inline
- Mark as read on expand

## Architecture

```mermaid
flowchart TB
    subgraph "Backend (existing)"
        DISP[notificationDispatcher]
        INAPP[inAppChannel]
        PUB[natsPublisher]
        NATS[(NATS)]
    end

    subgraph "Frontend (new)"
        SYNC[natsSync atom]
        NOTIF[notifications atom]
        BELL[NotificationBell]
    end

    DISP --> INAPP
    INAPP -->|publishToUser| PUB
    PUB -->|sync.user.{email}| NATS
    NATS -->|ws subscribe| SYNC
    SYNC -->|type: notification| NOTIF
    NOTIF -->|items, unreadCount| BELL
```

### Dual Subscription in natsSync

```typescript
// natsSync.ts - subscribe to both subjects
const broadcastSub = nc.subscribe('sync.broadcast')
const userSub = nc.subscribe(userSubject(currentUser.email))

// Process messages from either subscription
const processMessage = (msg: SyncMessage) => {
  if (msg.type === 'notification') {
    notifCtrl.update(...)  // forward to notifications atom
  } else if (msg.type === 'delta') {
    // existing delta processing
  }
}
```

## Rationale

| Considered | Rejected Because |
| --- | --- |
| Polling API for notifications | Already have real-time NATS infrastructure |
| Server-side read state | Adds complexity, localStorage sufficient for session |
| Toast-only notifications | Users need to see history, not just ephemeral toasts |
| Full notification center page | Bell dropdown is sufficient for MVP |

## Affected Layers

| Layer | Document | Change |
| --- | --- | --- |
| c3-1 | Web Frontend | Add notifications atom, NotificationBell component |
| c3-101 | State Management | Add notifications atom, extend SyncMessage type |
| c3-211 | Notification System | No changes (already sends in-app notifications) |

## Implementation Details

### Phase 1: Type Updates

Update `apps/start/src/lib/pumped/types.ts`:

```typescript
// Add InAppNotification type (mirrors server's InAppNotificationMessage.notification)
export interface InAppNotification {
  id: string
  notification_type: string
  payload: ApprovalRequiredPayload  // or union of payload types
  created_at: string
  read?: boolean  // client-side tracking
}

// Extend SyncMessage union
export type SyncMessage =
  | { type: 'full'; ... }
  | { type: 'delta'; ... }
  | { type: 'ack'; ... }
  | { type: 'notification'; notification: InAppNotification; executionId?: string }
```

### Phase 2: Notifications Atom

Create `apps/start/src/lib/pumped/atoms/notifications.ts`:

```typescript
export interface NotificationState {
  items: InAppNotification[]
}

export const notifications = atom<NotificationState>({
  factory: () => {
    // Load read IDs from localStorage
    const readIds = JSON.parse(localStorage.getItem('notification_read_ids') || '[]')
    return { items: [], readIds }
  }
})

// Computed unreadCount
export const unreadCount = atom({
  deps: { notifs: notifications },
  factory: (_, { notifs }) => notifs.items.filter(n => !n.read).length
})
```

### Phase 3: Update natsSync

Modify `apps/start/src/lib/pumped/atoms/natsSync.ts`:

- Add `notifications` to deps
- Get controller for mutations
- Subscribe to user subject in addition to broadcast
- Route `type: 'notification'` messages to notifications atom

### Phase 4: NotificationBell Component

Create `apps/start/src/components/NotificationBell.tsx`:

- Use Popover from shadcn/ui
- IconBell from @tabler/icons-react
- Badge with unread count
- List notifications with type, time, summary
- Click to expand payload details
- Mark all as read button

### Phase 5: Export and Integration

Update `apps/start/src/lib/pumped/index.ts`:

```typescript
export { notifications, unreadCount } from './atoms/notifications'
export type { InAppNotification, NotificationState } from './types'
```

Add `<NotificationBell />` to AppSidebar header area.

## References & Patterns

| Pattern | Reference | Application |
| --- | --- | --- |
| ref-sync | .c3/refs/ref-sync.md | Message types, subject naming convention |
| ref-ui-patterns | .c3/refs/ref-ui-patterns.md | Popover feedback patterns |
| c3-101 | .c3/c3-1-web-frontend/c3-101-state-management.md | Atom definition, controller pattern |

### Subject Naming

Server (natsPublisher.ts:35):

```typescript
const userSubject = (email: string) => `${subjectPrefix}.user.${email.replace(/[@.]/g, '_')}`
```

Client must use same escaping. The `subjectPrefix` is `sync` (from tags).

### Notification Lifecycle

| Event | Action |
| --- | --- |
| Message received | Add to items (prepend), cap at 50 |
| Notification expanded | Mark as read, persist ID to localStorage |
| Mark all read | Mark all items read, update localStorage |
| Dismiss | Remove from items array |
| Page refresh | Reload items (empty), restore read IDs from localStorage |

## Verification

- [ ] `bunx @typescript/native-preview` passes (types correct)
- [ ] natsSync subscribes to `sync.user.{email}` subject
- [ ] Notifications appear in bell dropdown when approval_required is triggered
- [ ] Unread badge shows correct count (computed)
- [ ] Mark as read updates badge
- [ ] Read IDs persist across page refreshes (localStorage)
- [ ] Clicking notification expands details inline
- [ ] Max 50 notifications (FIFO cleanup)

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
