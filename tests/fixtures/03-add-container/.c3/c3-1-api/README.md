---
id: c3-1
c3-version: 3
title: API Backend
type: container
parent: c3-0
summary: Messaging API with WebSocket support
---

# API Backend

## Technology Stack

| Layer | Technology | Purpose |
|-------|------------|---------|
| Runtime | Node.js | JavaScript runtime |
| Framework | Express | HTTP server |
| Realtime | Socket.io | WebSocket connections |
| Email | @sendgrid/mail | Email delivery |

## Components

| ID | Name | Responsibility |
|----|------|----------------|
| c3-101 | Message Handler | Message CRUD and delivery |
| c3-102 | WebSocket Manager | Connection management |
| c3-103 | Email Notifier | Basic email notifications via SendGrid |

## Internal Structure

```mermaid
flowchart TD
    WS[c3-102 WebSocket] --> MH[c3-101 Messages]
    MH --> EN[c3-103 Email]
    EN --> SG((SendGrid))
```

## Notes

- c3-103 Email Notifier is a quick implementation
- Only handles email, no SMS or push yet
- Consider extracting to dedicated notification service
