---
id: ref-pull-dispatcher
c3-seal: a49c9bd75e7b4e4cdfe6215497f929983ad9b66bac4379cff22d7c4ac369568c
title: Pull Dispatcher Pattern
type: ref
goal: Consumers subscribe to dispatcher (pull), not dispatcher pushing to consumers.
---

# Pull Dispatcher Pattern

## Goal

Consumers subscribe to dispatcher (pull), not dispatcher pushing to consumers.

## Choice

- Dispatcher exposes a `subscribe()` method; channels self-register by calling it in their factory
- Channels are loaded lazily via `controller()` -- the controller decides which channels to activate
- Active channel list is always queried fresh from the dispatcher, never cached

## When

- Pub/sub within the application
- Optional/configurable consumers
- Lazy-loaded channels

## Why

- Channels are self-registering
- Controller controls which channels load
- No stale references - always query fresh
- Clean dependency inversion

## Conventions

| Rule | Example |
| --- | --- |
| Dispatcher exposes subscribe() | dispatcher.subscribe({ channel, handler }) |
| Channels depend on dispatcher | deps: { dispatcher } |
| Channels subscribe in factory | const unsub = dispatcher.subscribe(...) |
| Channels cleanup on dispose | ctx.cleanup(unsub) |
| Controller uses controller() | controller(channel) for lazy loading |
| Always query fresh | getActiveChannels: () => dispatcher.getSubscribedChannels() |

## Pattern

```typescript
// Dispatcher exposes subscribe()
export const dispatcher = atom({
  factory: (ctx) => {
    const subscribers = new Map()

    return {
      subscribe: (sub) => {
        subscribers.set(sub.channel, sub)
        return () => subscribers.delete(sub.channel)
      },
      getSubscribedChannels: () => Array.from(subscribers.keys()),
    }
  },
})

// Channel depends on dispatcher and subscribes
export const channel = atom({
  deps: { dispatcher },
  factory: (ctx, { dispatcher }) => {
    const unsubscribe = dispatcher.subscribe({
      channel: 'name',
      handler: async (msg) => { ... }
    })
    ctx.cleanup(unsubscribe)
    return { name: 'name' }
  },
})

// Controller loads channels lazily via controller()
export const notificationController = atom({
  deps: {
    dispatcher,
    channelCtrl: controller(channel),
  },
  factory: async (_ctx, { dispatcher, channelCtrl }) => {
    await channelCtrl.resolve()  // lazy load

    return {
      // Always query fresh, never cache
      getActiveChannels: () => dispatcher.getSubscribedChannels(),
    }
  },
})
```

## Anti-Patterns

| Anti-Pattern | Problem | Correct |
| --- | --- | --- |
| Cache subscribed channels | Stale reference | Call getSubscribedChannels() fresh |
| Dispatcher pushes to channels | Tight coupling | Channels subscribe to dispatcher |
| Eager channel loading | Loads unused channels | Use controller() for lazy loading |

## Cited By

- c3-211 (Notification System)
