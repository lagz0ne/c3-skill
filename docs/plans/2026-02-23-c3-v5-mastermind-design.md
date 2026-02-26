# C3 v5: The Mastermind

**Date**: 2026-02-23
**Status**: Draft
**Supersedes**: c3 v4 (skills + agents model)

## Architecture Diagram

- [System Overview](https://diashort.apps.quickable.co/d/99318586)
- [Operations Lifecycle](https://diashort.apps.quickable.co/d/1892db66)

---

## The Shift

C3 v4 is a **map** — documentation that describes code. The agent reads docs, reads code, tries to keep them in sync.

C3 v5 is a **brain** — a mastermind agent that owns the knowledge graph and treats code as an expression medium. Changes don't start in code and get documented. They start in the mastermind's understanding and get emitted as code.

| | v4 (Map) | v5 (Brain) |
|---|---|---|
| Source of truth | Code | Knowledge graph |
| Code's role | The thing being documented | Expression medium |
| Change flow | Code → docs catch up | Reference change → code emitted |
| Audit question | "Are docs accurate?" | "Is code on-reference?" |
| Agent role | Navigator/assistant | Sole operator |
| Plugin shape | 5 skills + 2 agents | 1 mastermind agent |

## Core Concepts

### Components = The WHAT

Components are concrete, reusable building blocks. The logger itself. The auth middleware itself. Foundational code that C3 owns and maintains.

A component doc describes:
- What the component does
- Its interface/API
- Which container it belongs to
- Which references govern its usage

### References = The HOW

References are patterns and conventions — not code but rules for how components get used everywhere.

Example: ref-logging says "use debug for flow tracing, info for business events, error for failures requiring action." This isn't code. It's the rule that C3 enforces when writing code that touches logging anywhere in the system.

### The Relationship

Components provide the vocabulary. References provide the grammar. When C3 writes code anywhere:

1. It uses the right components (pulls in the logger)
2. It follows the right references (logs at the correct level, with correct context)

**"On-reference"** means: this code uses the right components in the way the references describe. It's semantic, not line-level.

## The Mastermind

### Why One Agent

Current C3 has 5 skills that only trigger when prompts contain C3-specific keywords ("C3", "component", "container"). Generic prompts like "add rate limiting" bypass C3 entirely.

The mastermind eliminates routing. It IS the interface. Every interaction flows through it, which means every interaction benefits from architectural knowledge.

```
v4: User → Claude → "is this C3?" → maybe invoke c3-change
                                   → maybe just use Explore

v5: User → C3 Mastermind → already knows the architecture
                          → every change goes through knowledge graph
```

### How It Delegates

The mastermind doesn't do all the work. It delegates to worker agents. The difference is the quality of delegation.

When a user says "add rate limiting to the API":

**Without mastermind**: Worker must discover where things are, what patterns exist, how the system works. Explores, guesses, may go off-pattern.

**With mastermind**: Worker receives a perfectly scoped, reference-rich prompt:

> You're modifying component middleware-stack in container api-service.
> Follow ref-rate-limiting: token bucket algorithm, apply at middleware layer.
> Return 429 with Retry-After header.
> Use component logger, level warn, per ref-logging-conventions.
> Here are the interfaces you need: [...]

Workers don't need C3 awareness. The enriched prompt IS the skill. Code is born on-reference because the prompt made it impossible to go off-reference.

## Operations

| Operation | Mode | Description |
|-----------|------|-------------|
| **Understand** | Internal | Vector search → load relevant docs → build scoped context |
| **Explain** | Read-only | Answer architecture questions, trace dependencies, impact analysis |
| **Enrich & Delegate** | Write | Compose architecture-aware prompt → spawn workers |
| **Review** | Internal | Verify worker output against references used during enrichment |
| **Evolve** | Write (docs) | Update knowledge graph, rebuild vector index |
| **Audit** | Read-only | Scan code against knowledge graph, report reference coverage |

### Operation Lifecycle (Write Path)

```
Understand → Enrich → Delegate → Review → Evolve
    ↑                                         |
    └─────────────────────────────────────────┘
              knowledge graph gets richer
```

### Operation Lifecycle (Read Path)

```
Understand → Explain (answer question)
Understand → Audit (report coverage)
```

## Context Management

The mastermind never holds the full knowledge graph in context. Each phase is context-disciplined:

### Phase 1: Intake & Index (lightweight)

```
Context: user request only
Action: vector search against .c3/index.db
Result: list of relevant entities (containers, components, refs) with relevance scores
```

### Phase 2: Deep Load (focused)

```
Context: 2-5 specific docs loaded based on vector search results
Action: read the relevant component docs, ref docs, container docs
Result: scoped architectural context for this specific task
```

### Phase 3: Enrich & Delegate (offloaded)

```
Context: architectural context → composed into enriched prompt
Action: spawn worker agent(s) with the enriched prompt
Result: mastermind releases worker context, workers execute independently
```

### Phase 4: Review & Evolve (separate pass)

```
Context: worker output + the same 2-5 docs
Action: verify on-reference, identify knowledge graph updates needed
Result: docs updated if needed, vector index rebuilt
```

## Storage Architecture

### Source of Truth: .c3/ Directory

```
.c3/
  config.yaml                # C3 configuration (embedding provider, model, etc.)
  containers/
    api-service/
      CONTAINER.md           # Container boundaries, purpose, constraints
      components/
        middleware/
          COMPONENT.md       # Component interface, ownership, dependencies
        routes/
          COMPONENT.md
  refs/
    ref-logging.md           # Pattern: when to log what, at which level
    ref-error-handling.md    # Pattern: RFC 7807, status codes, error shapes
  index.db                   # Vector index (derived, not committed to git)
```

Human-readable docs. Git-tracked. Reviewable in PRs. The directory structure IS the topology — no separate topology file needed. The vector index + directory structure gives the mastermind everything.

### Agent-Optimized Index: .c3/index.db

SQLite database with vector extension (bun + sqlite-vec). Built from .c3/ docs + code signatures.

**What gets indexed:**

| Source | Chunk Strategy | Content |
|--------|---------------|---------|
| Container docs | Whole doc (1 chunk) | Boundaries, purpose, owned components |
| Component docs | Whole doc (1 chunk) | Interface, dependencies, ownership |
| Reference docs | Per ## section (N chunks) | Each pattern/convention independently searchable |
| Code signatures | Per file (1 chunk) | Function/class/interface exports via tree-sitter AST parsing |

**Metadata per entry:**
- `type`: container, component, ref, code_signature
- `name`: entity name
- `path`: file path
- `relationships`: connected entities (parent container, applicable refs)

**Embedding model (configurable):**
- Default: nomic-embed-text-v1.5 (local ONNX, zero config, works offline)
- Optional: OpenAI text-embedding-3-small (API, better quality)
- Configured in `.c3/config.yaml`

**Query flow:**
1. Embed query string
2. Vector similarity search (top 10)
3. Metadata-aware re-ranking (boost architecturally connected entities)
4. Return ranked results with metadata

Rebuilt incrementally when docs change. Full rebuild via `bun scripts/index-build.ts`. Not committed to git (derived artifact).

### Plugin-Embedded

The vector DB runs as part of C3. No external services. No MCP servers.

```
Scripts:
  scripts/index-build.ts     # Build/rebuild vector index from .c3/ docs + code signatures
  scripts/index-query.ts     # Semantic search against index
```

## Self-Evolution (The Flywheel)

Every interaction makes C3 smarter. The mastermind **proposes** evolution — it never changes the knowledge graph without user approval.

### Evolution Types

| Type | Trigger | Example |
|------|---------|---------|
| **Accretive** | New code outside known components | "src/api/cache/ doesn't belong to any component. Create one?" |
| **Corrective** | Code diverged from docs | "middleware's interface changed — update the component doc?" |
| **Emergent** | Pattern repeated 3+ times without a ref | "Rate limiting appeared in 3 recent changes. Create ref-rate-limiting?" |

### Detection Mechanism

After worker completes, mastermind reviews output:
1. What components were touched?
2. Did new files appear outside known components?
3. Did any interface change?
4. Does the output follow the referenced patterns?

Classify as routine (no evolution needed) or structural (propose evolution).

### Proposal Flow

Proposals surface at the **end of a task**, not during:

```
"Done — rate limiter is active on all routes.

 I also noticed:
 1. src/api/cache/ has no component doc. Create one?
 2. The rate limiting pattern could be a ref. Add ref-rate-limiting?"
```

If user ignores proposals, C3 queues them. Periodically:
> "I have 3 pending evolution proposals from previous sessions. Quick review?"

### Episodic Memory: Git

No separate change record files. Git IS the episodic memory:

- **Recent changes**: `git log` — what happened
- **Structural changes**: `git diff` + tree-sitter — what changed in code structure
- **Pattern detection**: Compare git history against knowledge graph via vector search

Since C3 is the sole agent writing code, commit messages include context:
```
feat(api): add rate limiting to middleware

Components: middleware, routes
Refs applied: ref-error-handling
```

C3 analyzes git history on-demand to detect emergent patterns across sessions.

### The Flywheel

```
User request → Enrichment → Worker → Code
                                       ↓
                             Mastermind reviews
                                       ↓
                             Proposes evolution (if structural)
                                       ↓
                             User approves → Knowledge graph grows
                                       ↓
                             Next request gets better enrichment
```

More interactions → richer knowledge graph → better delegations → more consistent code.

## The C3 Invariant: References First, Code Second. Always.

This is the foundational principle of C3 v5.

```
CORRECT: Discuss reference → Establish reference → Implement following reference → Verify against reference
WRONG:   Implement code → Figure out pattern → Document as reference (NEVER)
```

The knowledge graph leads, code follows. Code is an expression of architecture, never the source of it.

### Impact-Driven Verification

The mastermind verifies at the **architecture level**, not the code level. It knows what to verify from impact analysis alone — without reading a single line of code:

```
"Add rate limiting to the API"
         ↓
Impact analysis (pure knowledge graph):
  → Affects: component middleware
  → Touches: ref-error-handling (new 429 error path)
  → Touches: ref-logging (new rate limit events)
  → Touches: ref-testing (new critical path needs coverage)
         ↓
Verification criteria (derived from impact, not code):
  1. ref-error-handling satisfied for the new error path
  2. ref-logging satisfied for the new event type
  3. ref-testing satisfied for the new critical path
  4. Component middleware's contract maintained
```

The mastermind doesn't need to see `res.status(429)` to know error handling matters. The knowledge graph tells it.

### Tests Are References

Testing conventions are documented as references, not ad-hoc decisions:

```
ref-testing:
  What to test: every public function, every API endpoint
  How to test: behavior-driven, not implementation-driven
  What NOT to test: don't mock what you don't own
  Error paths: must have explicit tests
```

Workers follow ref-testing the same way they follow any other reference.

### Workers Stop on Missing References

If a worker encounters a decision not covered by any reference, it STOPS and escalates:

```
Worker: "I need to decide rate limit state storage. No reference covers this."
         ↓
Mastermind discusses with user: "In-memory or shared store?"
         ↓
User decides → Mastermind creates ref-state-management
         ↓
Worker continues following the new reference
```

The invariant holds even mid-task. References are established BEFORE code is written.

### Verification Hierarchy

```
Mastermind verifies (WHAT):     │  Worker is responsible for (HOW):
────────────────────────────────│────────────────────────────────────
Requirement met?                │  Implementation details
References followed?            │  File structure, naming, wiring
Tests follow ref-testing?       │  Test framework, mocks, assertions
Boundaries respected?           │  Import paths, module structure
```

### Reference Coverage

The "100% inspectable, 100% testable, 100% demonstrable" goal.

Every block of code should be traceable to references:

- **On-reference**: Code follows the patterns its references describe
- **Off-reference**: Code contradicts a reference
- **Orphaned**: Code has no reference (shouldn't exist, or needs a new reference)

The Audit operation reports coverage. Over time, as the knowledge graph grows through self-evolution, coverage approaches 100%.

## Enrichment & Delegation

### Enrichment Pipeline

The mastermind doesn't micromanage. It provides a **requirements brief** — not step-by-step instructions.

```
User request (vague)
         ↓
Socratic clarification (architecture-informed, adaptive depth)
  - Clear intent → skip
  - Partial → 1-2 targeted questions
  - Vague → 2-3 questions, always propose-then-refine
  - Rule: never ask what you can infer from the knowledge graph
         ↓
Requirements brief (what the worker receives):
  REQUIREMENT: what needs to happen
  CONTEXT: which container/components are involved
  CONSTRAINTS: applicable references (not how to follow them — the refs themselves)
  VERIFICATION: what must be true when done (derived from impact analysis)
```

The worker gets the WHAT + guardrails. It figures out the HOW.

### Worker Autonomy

Workers are autonomous agent teams, not dumb executors:

1. Receive requirements brief from mastermind
2. Read the referenced refs and component docs themselves
3. Figure out implementation approach
4. Delegate further (spawn sub-workers) if complex
5. Query mastermind if context is missing
6. **STOP if a reference is missing** — escalate to mastermind for discussion
7. Run own verification against the criteria
8. Deliver result

### Delegation Patterns

The mastermind picks a pattern based on impact analysis:

- **Single worker**: one component affected → one worker
- **Parallel workers**: multiple independent components → workers in parallel
- **Sequential workers**: output of A needed by B → chain them

For parallel workers, the mastermind checks integration after all complete.

### Result Flow

```
Worker delivers result
         ↓
Mastermind verifies (at architecture level):
  - References satisfied? (from impact analysis)
  - Boundaries respected?
  - Tests present per ref-testing?
         ↓
All pass → report to user + check for evolution
Some fail → send back with specific reference violations
New ref needed → discuss with user, create ref, worker continues
```

## Plugin Structure (v5)

```
c3-design/
├── .claude-plugin/
│   └── plugin.json
├── agents/
│   └── c3-mastermind.md       # The ONE agent
├── skills/
│   └── c3-onboard/            # Bootstrap: create initial .c3/ docs + index
│       └── SKILL.md
├── scripts/
│   ├── index-build.ts         # Build vector index from .c3/ docs
│   └── index-query.ts         # Semantic search
└── references/
    └── ...                    # Shared reference material for the mastermind
```

### What Stays from v4
- The .c3/ doc format (containers, components, refs)
- The concept of ADRs for tracking decisions
- The onboarding flow for initial setup

### What Changes
- 5 skills → 1 mastermind agent (routing eliminated)
- Workers get enriched prompts instead of C3 skills
- Vector DB for semantic search (no more file-by-file navigation)
- Self-evolution loop after every change

## Implementation: PI Agent Framework

C3 v5 is built on [PI](https://github.com/badlogic/pi-mono), a TypeScript toolkit for constructing AI agents. PI provides:

- **pi-ai**: Multi-provider LLM abstraction (Anthropic, OpenAI, Google, etc.)
- **pi-agent-core**: Agent loop with tool calling
- **pi-coding-agent**: Full coding agent with built-in tools, session persistence, compaction, extensions
- **pi-tui**: Terminal UI with markdown rendering

### Architecture

```
┌─────────────────────────────────────┐
│  C3 Core                            │
│  Knowledge graph, vector search,    │
│  enrichment engine, evolution logic │
├──────────────────┬──────────────────┤
│  PI Agent CLI    │  Claude Code     │
│  (standalone)    │  Plugin          │
│  pi-coding-agent │  (agent .md      │
│  + pi-tui        │  + skill)        │
└──────────────────┴──────────────────┘
```

The core is the shared intelligence. Two deployment modes are thin wrappers.

### Core → PI Mapping

| Mastermind Concept | PI Implementation |
|---|---|
| Mastermind agent | `createAgentSession` with system prompt + custom tools |
| Knowledge graph search | Custom tool: `vector_search` → bun + sqlite-vec |
| Read architecture docs | Custom tool wrapping `read` scoped to `.c3/` |
| Code generation/editing | Built-in `codingTools` (read, write, edit, bash) |
| Prompt enrichment | Extension: `context` event injects relevant C3 docs before every LLM call |
| Self-evolution | Custom tool: `evolve_doc` → update .c3/ docs + rebuild index |
| Session memory | `SessionManager.open()` — JSONL persistence across restarts |
| Context management | Built-in compaction + extension to preserve architectural context |
| Worker delegation | Sub-agent sessions or agent loop with enriched prompts |

### Custom Tools

```typescript
// Knowledge graph search — the mastermind's primary lookup
const vectorSearchTool: AgentTool = {
  name: "c3_search",
  description: "Search the C3 knowledge graph for relevant architecture docs",
  // Calls: bun scripts/index-query.ts <query>
  // Returns: ranked list of relevant components, containers, refs
};

// Read a specific C3 doc
const readC3DocTool: AgentTool = {
  name: "c3_read",
  description: "Read a specific C3 architecture document",
  // Wraps built-in read, scoped to .c3/ directory
};

// Evolve the knowledge graph
const evolveTool: AgentTool = {
  name: "c3_evolve",
  description: "Update a C3 document and rebuild the vector index",
  // Writes updated doc + calls: bun scripts/index-build.ts
};

// Audit code against references
const auditTool: AgentTool = {
  name: "c3_audit",
  description: "Check if code in a file is on-reference or off-reference",
  // Reads code + relevant refs, reports coverage
};
```

### Extension: Auto-Enrichment

The key extension intercepts the `context` event to inject architecture context before every LLM call:

```typescript
export default function c3Enrichment(api: ExtensionAPI): void {
  api.on("context", async (event, ctx) => {
    // Extract the latest user message
    const lastUserMsg = event.messages.findLast(m => m.role === "user");
    if (!lastUserMsg) return;

    // Vector search for relevant C3 docs
    const relevant = await vectorSearch(lastUserMsg.content);

    // Inject as system context (not visible to user, available to LLM)
    const enrichment = formatArchitectureContext(relevant);
    return {
      messages: [
        { role: "system", content: enrichment, timestamp: Date.now() },
        ...event.messages,
      ],
    };
  });

  // Preserve architecture context during compaction
  api.on("session_before_compact", async (event, ctx) => {
    return {
      compaction: {
        summary: await summarizeWithArchitecturePreservation(event.messages),
        firstKeptEntryId: event.firstKeptEntryId,
        tokensBefore: event.tokensBefore,
      },
    };
  });
}
```

### Session Persistence

The mastermind uses JSONL sessions for cross-conversation memory:

```typescript
const sessionFile = path.join(projectRoot, ".c3", "sessions", "mastermind.jsonl");
const sessionManager = SessionManager.open(sessionFile);
```

This means the mastermind remembers previous interactions, decisions, and context across restarts. Combined with the vector index, it has both semantic search (what's architecturally relevant?) and episodic memory (what did we discuss before?).

### Standalone CLI

```bash
# Run C3 as a standalone agent
c3                          # Interactive TUI mode
c3 "add rate limiting"      # One-shot mode
c3 --audit                  # Audit mode
```

### Claude Code Plugin

The same core exposed as a Claude Code agent:

```yaml
# agents/c3-mastermind.md
---
name: c3-mastermind
description: C3 architecture mastermind agent
tools: [Read, Write, Edit, Bash, Grep, Glob]
---
System prompt that loads C3 core and delegates...
```

## Onboarding Flow

Three paths into C3, all ending at the same state: a mastermind with a knowledge graph.

### Path 1: Existing Codebase (most common)

**Phase 0: Silent Scan** (automated, ~30 seconds)

Tree-sitter parses all source files → directory layout, import graph, export signatures. C3 builds an internal model of candidate containers, components, and patterns. User sees a progress indicator, not questions.

**Socratic Phases (non-linear, coherence-driven):**

The phases guide direction but are not a strict pipeline. If a later answer contradicts an earlier conclusion, C3 loops back. Exit condition is coherence, not completion.

```
Phase 1: What Exists?
    ↓ ↑
Phase 2: What Matters?     ← late discovery loops back
    ↓ ↑
Phase 3: What's Invisible? ← "oh wait, that changes the structure"
    ↓
Phase 4: Go (only when coherent)
```

**Phase 1: What Exists?** (goal: shared understanding of system structure)

C3 shows what the scan revealed. Asks leading questions:
> "I see code organized under src/api/ and src/worker/. Are these two independent systems that could run separately, or are they tightly coupled?"

Always proposes, never asks blank questions. Exit when: C3 and user agree on structural boundaries.

**Phase 2: What Matters?** (goal: identify patterns worth enforcing)

C3 presents discovered patterns, asks the user to weigh them:
> "I see every route handler validates auth. Is this a rule you'd want enforced on every new route, or is it flexible?"

Distinguishes rules from preferences. Exit when: at least 1-2 references identified.

**Phase 3: What's Invisible?** (goal: capture intent that code can't reveal)

> "What's the one thing new developers always get wrong in this codebase?"

Surfaces conventions living in people's heads. Exit when: user has shared their key conventions.

**Phase 4: Go** (action, not question)

C3 shows the knowledge graph, creates it, immediately ready.

**Contradiction handling:** After each answer, C3 checks alignment with previous conclusions. If misaligned:
> "Interesting — you mentioned X, but earlier we agreed Y. These pull in different directions. Which is closer to reality?"

C3 challenges assumptions (including its own) and loops back to the relevant phase.

**Target: 4-5 exchanges.** Can extend to 6-7 for complex/ambiguous codebases, but each extra exchange resolves a specific tension.

### Path 2: Greenfield (no code yet)

Socratic discovery: C3 asks about what you're building, boundaries, and patterns. Creates .c3/ docs from answers. Index starts empty, grows as code is written through C3.

### Path 3: Partial (some code, some docs)

Hybrid: scans what exists, asks about gaps. Adapts based on what it finds.

### Design Principles

- **Scan does the heavy lifting** — questions validate, not discover
- **Encouraging tone** — "nice structure", "you're already doing these well"
- **No dead state** — after Q4, C3 is immediately useful
- **Maximum information per question** — each question covers multiple entities

## Open Questions

1. **Embedding model**: Which model generates the vector embeddings? Local (e.g., ONNX) vs API call?
2. **Index rebuild triggers**: On every doc change? Batch? Background process?
3. **Worker trust boundary**: How much does the mastermind verify worker output vs trust it?
4. **Onboarding for existing codebases**: How does the mastermind build the initial knowledge graph from existing code?
5. **Conflict resolution**: What happens if a worker produces code that's on-reference for one ref but off-reference for another?
6. **PI version pinning**: Which PI packages version to target? Latest from npm?
7. **Provider default**: Which LLM provider/model for the mastermind? Anthropic Claude as default, configurable?
