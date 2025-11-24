# Container Archetype Hints

Archetypes are **starting points for exploration**, not templates to fill.

Use these to:
- Recognize container patterns during discovery
- Know what roles to look for
- Ask informed questions

## Backend Service

**Typical Roles:** Request Handler, Business Logic, Database Access

**Discovery approach:**
1. Look for: package.json (express, fastify, nestjs), go.mod (chi, gin), requirements.txt (fastapi, django)
2. Find entry points: main.ts, app.py, main.go
3. Trace routes → handlers → services → data layer

**Questions to ask:**
- "What endpoints does this expose?"
- "Where does business logic live?"
- "What database/storage does this use?"

## Frontend SPA

**Typical Roles:** View Layer, State Container, Router, Integration

**Discovery approach:**
1. Look for: package.json (react, vue, svelte, angular)
2. Find component structure: src/components/, src/pages/
3. Find state management: store/, redux/, zustand

**Questions to ask:**
- "What UI framework?"
- "How is state managed?"
- "How does it communicate with backend?"

## Meta-Framework (Next.js, Nuxt, SvelteKit)

**Typical Roles:** Varies by execution context

**Discovery approach:**
1. Identify framework: next.config.js, nuxt.config.ts, svelte.config.js
2. Map execution contexts:
   - Server build-time: getStaticProps, generateStaticParams
   - Server runtime: API routes, server actions, RSC
   - Client runtime: use client components, client state

**Execution Context Table:**

| Context | Available Roles |
|---------|-----------------|
| Server Build-time | Data Access, View Layer, Configuration |
| Server Runtime | Request Handler, Business Logic, Data Access |
| Client Runtime | View Layer, State Container, Router, Form Handler |

**Questions to ask:**
- "What renders on server vs client?"
- "Are there API routes?"
- "How is data fetched (server vs client)?"

## Worker / Background Processor

**Typical Roles:** Event Consumer, Business Logic, Database Access

**Discovery approach:**
1. Look for: queue connection config, job definitions
2. Find handlers: workers/, jobs/, consumers/
3. Trace: queue → handler → processing → output

**Questions to ask:**
- "What queue/broker does this consume from?"
- "What triggers job execution?"
- "Where do results go?"

## API Gateway / BFF

**Typical Roles:** Request Handler, Aggregator, Internal Client

**Discovery approach:**
1. Look for: proxy config, route aggregation
2. Find upstream service definitions
3. Check for auth/rate limiting middleware

**Questions to ask:**
- "What services does this route to?"
- "Any request transformation or aggregation?"
- "Where is auth handled?"

## Infrastructure Container

**No code components - has Configuration Surface instead**

**Discovery approach:**
1. Identify technology: PostgreSQL, Redis, RabbitMQ, etc.
2. Find configuration: terraform, helm, docker-compose
3. Document interfaces and operational characteristics

**Documentation pattern:**

```markdown
## Configuration Surface
| Parameter | Value | Rationale |

## Interfaces
- Inbound: [port, protocol, auth method]
- Outbound: [replication, backups]

## Operational Characteristics
- Backup: [strategy, retention]
- Failover: [automatic/manual, RTO]
- Scaling: [horizontal/vertical, constraints]

## Dependencies
- Requires: [VPC, IAM, encryption keys]
- Consumed by: [list of containers]
```

## CLI Tool

**Typical Roles:** Request Handler (args), Business Logic, Integration

**Discovery approach:**
1. Look for: bin/ directory, CLI framework (commander, cobra, click)
2. Find command definitions
3. Trace: command → handler → action

**Questions to ask:**
- "What commands are available?"
- "What does each command do?"
- "Any external service calls?"
