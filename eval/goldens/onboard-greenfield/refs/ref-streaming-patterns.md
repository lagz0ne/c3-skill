---
id: ref-streaming-patterns
title: Streaming Response Patterns
---

# Streaming Response Patterns

## Goal

Establish patterns for Server-Sent Events (SSE) streaming including event formatting, client-side handling, progressive rendering, and error recovery for AI responses.

## Overview

Conventions for Server-Sent Events (SSE) and streaming responses.

## SSE Protocol

### Event Format

| Field | Description |
|-------|-------------|
| event | Event type (optional) |
| data | JSON payload |
| id | Event ID for resumption |
| retry | Reconnect delay hint |

### Event Types

| Event | Payload | Purpose |
|-------|---------|---------|
| chunk | {text: string} | Incremental content |
| citation | {conceptId, title} | Referenced concept |
| suggestion | {type, action} | Action recommendation |
| done | {usage} | Stream complete |
| error | {code, message} | Error occurred |

## Client Handling

### State Machine

```
Idle → Streaming → Done
         ↓
       Error → Retry/Idle
```

### Buffer Management

| Scenario | Handling |
|----------|----------|
| Fast chunks | Batch renders at 60fps |
| Slow chunks | Render immediately |
| Large payload | Chunk reassembly |

## Rendering

### Progressive Display

| Approach | When |
|----------|------|
| Character-by-character | Chat responses |
| Block-by-block | Structured content |
| All-at-once | Short responses |

### Markdown Streaming

| Challenge | Solution |
|-----------|----------|
| Incomplete code blocks | Defer render until closed |
| Partial links | Buffer until pattern complete |
| Tables | Buffer until row complete |

## Error Recovery

| Error | Recovery |
|-------|----------|
| Connection drop | Retry with last-event-id |
| Timeout | Show partial, offer retry |
| Invalid JSON | Skip chunk, log warning |
| Server error | Show error, offer retry |

## Performance

| Metric | Target |
|--------|--------|
| Time to first byte | <200ms |
| Chunk interval | ~50ms |
| Render latency | <16ms |
| Memory growth | Bounded buffer |

## Cited By

- c3-113 (AI Chat Panel)
- c3-214 (Chat Handler)
