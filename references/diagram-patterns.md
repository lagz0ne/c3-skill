# Diagram Patterns

Diagrams should **clarify**, not mandate. Generate when helpful, skip when trivial.

---

## ⛔ MERMAID-ONLY ENFORCEMENT (MANDATORY)

**This is non-negotiable. ALL diagrams in C3 documentation MUST use Mermaid syntax.**

### What is Prohibited (NO EXCEPTIONS)

| Prohibited | Example | Why Prohibited |
|------------|---------|----------------|
| ASCII art diagrams | `+---+ → +---+` boxes | Not renderable, poor accessibility |
| Text-based flowcharts | `A --> B --> C` plain text | Not a diagram, just text |
| Unicode box drawing | `┌──┐ │ │ └──┘` | Platform-dependent rendering |
| Indented hierarchies | Tree structures with `├──` | Use Mermaid `graph TD` instead |
| Any non-Mermaid visual | Anything not in ` ```mermaid ` block | Inconsistent, not interactive |

### Why Mermaid is Required

1. **Renderable** - GitHub, GitLab, and most markdown viewers render Mermaid natively
2. **Accessible** - Screen readers can parse structure
3. **Maintainable** - Standard syntax, easy to diff/review in PRs
4. **Consistent** - Single format across all documentation

### Validation Checklist (RUN BEFORE FINALIZING)

Before completing any C3 document with diagrams, verify:

- [ ] **All diagrams use ` ```mermaid ` code blocks**
- [ ] **No ASCII art or Unicode box drawing anywhere**
- [ ] **No plain-text "diagrams" that should be visual**

### Red Flags - STOP and Rewrite If You See

🚩 `+---+` or `|   |` box drawing characters
🚩 `───>` or `-->` arrows outside Mermaid blocks
🚩 `├──` or `└──` tree characters
🚩 Indentation-based hierarchies meant to show structure
🚩 "Here's the flow: A then B then C" without a mermaid diagram

---

## When to Use Diagrams

**Use a diagram when:**
- Relationships are non-obvious
- Data flow has multiple steps
- Test orchestration is complex
- Architecture overview aids understanding

**Skip diagrams when:**
- Relationships are trivial (A calls B)
- Text description is clearer
- Diagram would just repeat prose

## Component Relationships (Container Level)

Show how components within a container interact.

```mermaid
flowchart TD
    subgraph Container["Backend Service"]
        RH[Request Handler]
        BL[Business Logic]
        DA[Database Access]
        IC[Integration Client]
    end

    RH --> BL
    BL --> DA
    BL --> IC
    DA --> DB[(Database)]
    IC --> ExtSvc[External Service]
```

**When to use:** Container has 3+ components with non-trivial relationships.

## Data Flow (Sequence)

Show how a request flows through the system.

```mermaid
sequenceDiagram
    participant Client
    participant Handler
    participant Service
    participant DB

    Client->>Handler: POST /orders
    Handler->>Service: createOrder(data)
    Service->>DB: INSERT
    Service-->>Handler: order
    Handler-->>Client: 201 Created
```

**When to use:** Request involves multiple components/services, error handling branches, or async steps.

## Container Overview (Context Level)

Show all containers and their relationships.

```mermaid
flowchart TB
    subgraph System
        FE[Frontend]
        API[API Service]
        Worker[Worker]
        DB[(PostgreSQL)]
        Queue[RabbitMQ]
    end

    User --> FE
    FE --> API
    API --> DB
    API --> Queue
    Worker --> Queue
    Worker --> DB
```

**When to use:** System has 3+ containers.

## Platform Topology

Show infrastructure layout.

```mermaid
flowchart TB
    subgraph Internet
        User
    end

    subgraph Platform
        LB[Load Balancer]
        subgraph Private["Private Subnet"]
            API
            Worker
        end
        subgraph Data["Data Subnet"]
            DB[(Database)]
            Cache[(Redis)]
        end
    end

    User --> LB
    LB --> API
    API --> DB
    API --> Cache
    Worker --> DB
```

**When to use:** Documenting platform/networking.

## Execution Context (Meta-Frameworks)

Show server vs client code paths.

```mermaid
flowchart LR
    subgraph Build["Build Time"]
        SSG[Static Gen]
    end

    subgraph Server["Server Runtime"]
        RSC[Server Components]
        API[API Routes]
    end

    subgraph Client["Client Runtime"]
        Hydrate[Hydration]
        State[Client State]
    end

    Build --> Server
    Server --> Client
```

**When to use:** Documenting Next.js, Nuxt, SvelteKit applications.

## Test Orchestration

Show test setup and teardown.

```mermaid
sequenceDiagram
    participant CI
    participant TestDB
    participant App
    participant Tests

    CI->>TestDB: docker-compose up
    CI->>App: npm start
    CI->>Tests: npm test
    Tests->>App: requests
    Tests-->>CI: results
    CI->>App: stop
    CI->>TestDB: docker-compose down
```

**When to use:** Integration tests with external dependencies.

## State Machine (Complex Business Logic)

Show state transitions for workflows.

```mermaid
stateDiagram-v2
    [*] --> Pending
    Pending --> Processing: start
    Processing --> Completed: success
    Processing --> Failed: error
    Failed --> Processing: retry
    Completed --> [*]
```

**When to use:** Component manages state transitions (orders, workflows, sagas).

## Anti-Patterns

**Don't:**
- Generate diagrams for every component (noise)
- Create diagrams that just mirror the prose
- Use complex diagram types when simple flowchart works
- Force diagrams where text is clearer

**Do:**
- Ask "does this diagram add clarity?"
- Keep diagrams focused (5-10 nodes max)
- Use consistent naming with documentation
- Update diagrams when code changes
