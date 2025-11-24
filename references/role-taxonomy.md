# Role Taxonomy

Role patterns are **vocabulary for discussion**, not requirements to fill in.

Use these patterns to:
- Recognize what you discover in code
- Ask informed questions during discovery
- Document using shared language

## Communication Roles

How the component/container connects to the outside world.

| Role | Description | Discovery Cues |
|------|-------------|----------------|
| Request Handler | Receives sync requests (HTTP, gRPC, GraphQL) | Route files, controllers, handlers |
| Event Consumer | Receives async messages/events | Queue listeners, subscribers, on() handlers |
| Event Producer | Emits messages/events | publish(), emit(), send() calls |
| Scheduled Handler | Time-triggered execution | Cron config, @Scheduled, setInterval |
| Gateway/Edge | External traffic ingress | Proxy config, load balancer rules |

## Processing Roles

What the component does with data/requests.

| Role | Description | Discovery Cues |
|------|-------------|----------------|
| Business Logic | Domain rules, workflows, validation | Service classes, use cases, domain/ |
| Transformer | Data mapping, format conversion | Mappers, serializers, adapters |
| Aggregator | Combines data from multiple sources | BFF patterns, Promise.all, concurrent calls |
| Saga Coordinator | Multi-step distributed workflows | State machines, orchestrators, compensating transactions |
| Form Handler | Input capture, validation, submission | Form components, useForm, validators |

## State Roles

How the component manages data persistence.

| Role | Description | Discovery Cues |
|------|-------------|----------------|
| Database Access | ACID-compliant storage | ORM calls, repositories, SQL queries |
| Cache Access | Ephemeral/best-effort storage | Redis clients, cache decorators, memoization |
| Object Storage | Large blob storage | S3 clients, file upload handlers |
| State Container | Application state (frontend) | Redux, Zustand, Pinia, stores/ |
| Client Cache | API response caching (frontend) | React Query, SWR, Apollo cache |

## Presentation Roles

Frontend-specific UI concerns.

| Role | Description | Discovery Cues |
|------|-------------|----------------|
| View Layer | Renders UI from state | Components, templates, JSX/TSX |
| Router | URL-to-view mapping | Route config, pages/, app/ directory |
| Hydration | SSR-to-client state transfer | use client, hydration boundaries |

## Integration Roles

How the component communicates with other systems.

| Role | Description | Discovery Cues |
|------|-------------|----------------|
| Internal Client | Calls within trust boundary | Service clients, internal SDK, gRPC stubs |
| External Client | Calls third-party systems | API adapters, webhook handlers, OAuth flows |

## Operational Roles

Deployment and runtime concerns.

| Role | Description | Discovery Cues |
|------|-------------|----------------|
| Sidecar | Co-located helper process | Envoy config, log shippers, secret injectors |
| Init/Bootstrap | Pre-start setup | Migration scripts, secret fetch, cert rotation |
| Health/Readiness | Orchestrator signals | /health, /ready endpoints, probes |

## Using Role Vocabulary

**During discovery:**
- "I see route handlers here - this looks like a Request Handler role. Accurate?"
- "There's a Redis client - is this Cache Access or something else?"

**When documenting:**
- Use role names in component descriptions
- Don't force every role to be present
- Document what exists, not what "should" exist

**Anti-patterns:**
- Treating roles as checkboxes to fill
- Forcing code into role categories that don't fit
- Creating roles for hypothetical future needs
