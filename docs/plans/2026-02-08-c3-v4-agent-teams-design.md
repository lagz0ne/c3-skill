# C3 v4 Design: Agent Teams Architecture

## Problem

C3 v3's enforcement model blocks file edits at write-time (c3-gate), requires 7 agents with deep handoff chains, and treats implementation drift as a failure to prevent rather than a reality to manage. This creates friction without proportional safety gains.

## Principles

1. **C3 = source of truth for CONCEPTS** (relationships, constraints, chosen options). Abstract down, details up.
2. **Unidirectional by default, regression by exception.** Change request -> ADR -> Tasks -> Execute -> Audit. Loops are explicit and user-confirmed.
3. **Implementation drifts.** Don't guard file editing. Make tasks rich with C3 context. Audit after implementation.
4. **Agent Teams coordinate work.** Lead orchestrates, teammates execute. Shared task list. Peer messaging.
5. **Refs are the consistency engine.** Chosen options with rationale. Cross-cutting. First-class.

---

## C3 Hierarchy

### The 5 Layers

| Layer | Role | Question It Answers |
|-------|------|---------------------|
| **Context** (c3-0) | Abstract constraints | "What must this system do at the highest level?" |
| **Container** (c3-N) | Responsibility allocation + deployment boundary | "What does this unit contribute to satisfying context constraints?" |
| **Foundation** (c3-N0x) | Platform choices | "What did we choose, why, and what does it provide?" |
| **Feature** (c3-Nxx, xx>=10) | Business logic | "What domain problem does this solve, using what's provided?" |
| **Ref** (ref-*) | Cross-cutting choices | "How do we do X consistently across boundaries?" |

### Layer Relationships

```
Context (c3-0)
  defines ABSTRACT constraints
  ("web protocol between frontend and backend", "user data encrypted at rest")
      |
Container (c3-1-backend)
  GLUES components by allocating RESPONSIBILITIES
  to satisfy context constraints
  Also a DEPLOYMENT/RUNTIME boundary
  ("Backend: HTTP API, data persistence, auth enforcement, business logic")
      |
  +-- Foundation (c3-101-express)
  |   CHOICE: Express (why: xyz)
  |   PROVIDES: routes, middleware, request lifecycle
  |   FULFILLS: container's "HTTP API serving" responsibility
  |   CODE AREA: src/middleware/, src/server.ts
  |
  +-- Feature (c3-110-user-service)
      FOCUSES: user registration, profile management
      USES: c3-101-express (routes), c3-102-postgres (storage)
      FOLLOWS: ref-error-handling, ref-validation
      CODE AREA: src/services/user/, src/routes/user/

Ref (ref-error-handling)
  CHOICE: Structured ErrorResponse with error codes
  WHY: Consistent client-side parsing, machine-readable errors
  HOW: [detailed implementation pattern]
  NOT THIS: String error messages, HTTP-only status codes
  SCOPE: c3-1-backend (all components)
  OVERRIDE: Document justification in ADR "Pattern Overrides" section
  APPEARS IN: foundation code (Express middleware) AND feature code (business errors)
```

### Container Role

Container is both:
- **Runtime/deployment boundary**: a separately deployable unit (backend API, frontend app, worker)
- **Responsibility allocator**: takes abstract context constraints and assigns them as concrete responsibilities to its components

This dual role prevents containers from becoming catch-all orchestration layers. If a responsibility can't be mapped to a deployable unit, it's likely a ref (cross-cutting) not a container concern.

### Foundation vs Feature

| | Foundation | Feature |
|-|-----------|---------|
| **Test** | Has concrete code depended on by others | Composes domain logic using foundations |
| **Numbering** | c3-N01 through c3-N09 | c3-N10+ |
| **Relationship** | "I provide capabilities" | "I use what's provided, not responsible for it" |
| **Code area** | Platform setup, infrastructure | Business logic, domain flows |

Mixed components (both providing infrastructure AND implementing business logic) should be split whenever possible. If not splittable, classify by primary responsibility and explicitly document the secondary role.

### Ref Structure

```markdown
# ref-{slug}.md

## Choice
What we chose.

## Why
Why this over alternatives.

## How
Detailed implementation guidance.

## Not This
Alternatives we rejected and why.

## Scope
Where this applies. Container-scoped or system-wide.
Does NOT apply to: [explicit exclusions]

## Override
How to deviate: document justification in ADR "Pattern Overrides" section.
Temporary deviations must include migration timeline.

## Cited By
Components that follow this ref.
```

Refs cover ANY cross-cutting chosen option:
- Coding style and conventions
- State management approach (Zustand, Redux, Context — we chose X)
- Naming conventions (DB fields, API routes, component names)
- Testing patterns (integration vs unit, mocking approach)
- Error handling flow
- API conventions (REST vs GraphQL, envelope format)
- Security requirements (ref-security)
- Performance budgets (ref-performance)
- Non-functional constraints

---

## Agent Teams Flow

### Overview

```
User Request
    |
    v
Phase 1: Understand (Lead + analyst + reviewer teammates)
    |
    v
Phase 2: ADR (Lead writes, user accepts)
    |
    v
Phase 3: Execute (Lead decomposes -> implementer teammates)
    |
    v
Phase 4: Audit (Auditor teammate -> C3 doc updates)
    |
    v
New version of C3 docs
```

### Phase 1: Understand

Lead enters delegate mode (coordination only, no code).

1. Lead reads C3 docs relevant to the change request
2. Lead clarifies intent with user via Socratic questions
3. Lead spawns two teammates:
   - **Analyst**: investigates impact — reads affected components, traces dependencies, checks refs
   - **Reviewer**: challenges the analyst's findings — plays devil's advocate, looks for missed impacts
4. Analyst and reviewer message each other (peer debate)
5. Lead synthesizes findings, presents to user

### Phase 2: ADR

1. Lead writes ADR with Work Breakdown section (no separate plan files)
2. ADR sections: Problem, Decision, Rationale, Affected Layers, Work Breakdown, Verification
3. User reviews and accepts ADR
4. On acceptance: `status: proposed` -> `status: accepted`, `base-commit: <HEAD>`

### Phase 3: Execute

1. Lead decomposes Work Breakdown into shared tasks (TaskCreate)
2. Lead sets task dependencies (TaskUpdate with addBlockedBy)
3. Lead ensures no two tasks target the same files (conflict avoidance)
4. Lead spawns N implementer teammates
5. Each implementer:
   - Claims a task from shared list
   - Reads the ADR for decision context
   - Reads affected component docs for code references + relationships
   - Reads ALL cited refs for implementation constraints
   - Implements, conforming to every cited ref
   - Runs verification command
   - Marks task complete
6. Lead reviews each completed task:
   - Acceptance criteria met?
   - Changes conform to cited refs?
   - Files changed match expected scope?
   - If not: rejects with feedback, teammate revises
7. Lead monitors progress, redirects stalled teammates

### Phase 4: Audit

1. Lead spawns auditor teammate
2. Auditor compares C3 docs vs actual code changes:
   - Component docs still accurate? (code references, relationships)
   - Refs still hold? (pattern compliance across boundaries)
   - Any new cross-cutting patterns emerged?
3. Auditor reports findings to lead
4. Lead updates C3 docs (or creates tasks for doc updates)
5. Lead updates CLAUDE.md files (replacing c3-apply's role)
6. Lead transitions ADR: `status: accepted` -> `status: implemented`

### Task Anatomy

Tasks point to C3 docs (teammates read the source of truth directly):

```
TaskCreate:
  subject: "Add rate limiting to API gateway"
  description: |
    ## ADR
    .c3/adr/adr-20260206-rate-limiting.md

    ## Components (WHAT + WHERE)
    - .c3/c3-2-backend/c3-201-api-gateway.md (primary)
    - .c3/c3-2-backend/c3-101-auth.md (dependency)

    ## Refs (HOW - mandatory reading)
    - .c3/refs/ref-error-handling.md
    - .c3/refs/ref-api-conventions.md
    - .c3/refs/ref-testing.md

    ## Task
    Implement rate limiting middleware. Read components for code
    locations and relationships. Read ALL refs for constraints -
    code must conform to every cited ref.

    ## Acceptance Criteria
    - [ ] Rate limiter middleware exists
    - [ ] Conforms to ref-api-conventions (middleware chain order)
    - [ ] Conforms to ref-error-handling (429 response shape)
    - [ ] Tests follow ref-testing patterns
    - [ ] Verification command passes

    ## Verification
    Run: npm test -- --grep "rate-limit"
  activeForm: "Implementing rate limiting"
```

---

## Regression Model

### Universal Escape

Any phase can discover issues. Teammates surface discoveries to the lead. The lead checks against the ADR structure:

```
Teammate surfaces discovery
         |
Does this change the PROBLEM?       -> Phase 1 (user confirms)
Does this change the DECISION?      -> Phase 2 (user confirms)
Does this change AFFECTED LAYERS?   -> Amend ADR (user confirms)
Implementation detail only?         -> Adjust tasks (lead handles)
```

### Decision Examples

| Discovery | ADR Section | Lead Action |
|-----------|-------------|-------------|
| "Auth service was replaced by SSO gateway" | Problem | Phase 1 (re-understand) |
| "Can't do this with REST, needs WebSocket" | Decision | Phase 2 (new ADR) |
| "Component c3-205 also needs updating" | Affected Layers | Amend ADR + add tasks |
| "Function is called processAuth not handleAuth" | None | Adjust task |

### Confirmation Rules

- Phase 1/2 regression: always confirm with user
- Amend ADR: confirm with user
- Adjust tasks: lead handles autonomously

### ADR Versioning

When an ADR is amended during execution, increment a revision counter in frontmatter. Teammates working on stale revisions get a broadcast from the lead with updated context.

---

## Plugin Inventory

### Skills (5)

| Skill | Purpose | Uses Teams? |
|-------|---------|-------------|
| **onboard** | Create C3 from scratch (3-stage discovery) | Optional |
| **c3-change** | Team-based architectural change flow | Yes (core) |
| **c3-query** | Navigate C3 docs, answer questions | No |
| **c3-ref** | Manage cross-cutting patterns (refs) | No |
| **c3-audit** | Standalone audit (also Phase 4 teammate role) | No |

### Agents (2)

| Agent | Purpose |
|-------|---------|
| **c3-navigator** | Powers c3-query (navigation + code exploration) |
| **c3-lead** | System prompt for team lead in c3-change |

### Teammate Roles (4 spawn prompts, not agent files)

| Role | Spawned In | Purpose |
|------|-----------|---------|
| **Analyst** | Phase 1 | Impact analysis, dependency tracing |
| **Reviewer** | Phase 1 | Challenge analyst findings, devil's advocate |
| **Implementer** | Phase 3 | Execute tasks, write code conforming to refs |
| **Auditor** | Phase 4 | Compare C3 docs vs code, detect drift |

### Hooks

**None.** Zero hooks. All context is agent-mediated:
- CLAUDE.md files (auto-loaded by Claude Code) provide ambient context
- Task descriptions include C3 doc pointers
- Lead reviews task completion (no TaskCompleted hook needed)

### What Was Removed (from v3)

| Removed | Replacement |
|---------|-------------|
| c3-gate hook | No pre-edit blocking. Audit after implementation. |
| c3-edit-context hook | Task descriptions include component context. |
| c3-context-loader hook | CLAUDE.md provides ambient context. |
| c3-dev agent | Implementer teammates |
| c3-adr-transition agent | Lead updates ADR status directly |
| c3-synthesizer agent | Lead synthesizes from analyst + reviewer |
| c3-summarizer agent | Navigator handles inline |
| c3-apply skill | Auditor teammate updates CLAUDE.md in Phase 4 |
| c3-provision skill | Merged into c3-change (provision = change where Phase 3 produces docs) |
| c3-alter skill | Replaced by c3-change |
| Plan files (.plan.md) | ADR has Work Breakdown section |
| All scripts (c3-gate, c3-edit-context, etc.) | Agent-mediated gates |

### File Structure

```
c3-design/
+-- skills/
|   +-- onboard/SKILL.md
|   +-- c3-change/SKILL.md
|   +-- c3-query/SKILL.md
|   +-- c3-ref/SKILL.md
|   +-- c3-audit/SKILL.md
+-- agents/
|   +-- c3-navigator.md
|   +-- c3-lead.md
+-- references/              # Shared reference docs
+-- templates/               # Doc templates (ADR, component, ref, etc.)
+-- .claude-plugin/plugin.json
```

No `hooks/`, no `scripts/`, no `commands/`.

---

## Approved-Files: Soft Contract

Without c3-gate, "approved-files" shifts from hard enforcement to guidance:

- **In ADR**: Expected file touch set (guidance for task creation)
- **In tasks**: Lead includes expected files from ADR
- **In audit**: Auditor checks actual changes vs expected scope
- **On out-of-scope edits**: Teammate surfaces to lead -> lead decides (adjust task vs amend ADR)
- **No blocking**: Files are never blocked from editing

---

## C3 Doc Updates After Implementation

| Who | Updates What |
|-----|-------------|
| Implementers | Component docs they touched (code references, relationships) |
| Lead | ADR status (accepted -> implemented), affected layers |
| Auditor | Verifies drift, flags stale docs, updates CLAUDE.md |

Doc updates happen in the same change cycle, not deferred.

---

## Open Items for Implementation

1. c3-lead agent system prompt (detailed orchestration instructions)
2. Teammate spawn prompts (analyst, reviewer, implementer, auditor)
3. c3-change skill SKILL.md (entry point, team setup, 4-phase flow)
4. Updated templates (ADR with Work Breakdown, ref with Scope + Override)
5. Updated onboard skill (new ref structure, container definition)
6. Migration guide from v3 -> v4
7. Enable `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` in plugin settings
