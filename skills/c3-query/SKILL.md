---
name: c3-query
description: |
  Navigates C3 architecture docs and explores corresponding code to answer architecture questions.

  Use when the user asks:
  - "where is X", "how does X work", "explain X", "show me the architecture"
  - "find component", "what handles X", "diagram of X", "visualize X"
  - "describe X", "list components", "trace X", "flow of X"
  - References C3 IDs (c3-0, c3-1, adr-*)

  <example>
  Context: Project with .c3/ directory
  user: "explain what c3-101 does and how it connects to other components"
  assistant: "Using c3-query to navigate the architecture docs."
  </example>

  <example>
  Context: Project with .c3/ directory
  user: "show me a diagram of the C3 architecture"
  assistant: "Using c3-query to generate an architecture overview."
  </example>

  DO NOT use for changes (route to c3-change).
  DO NOT use for pattern artifact management — listing, creating, updating refs (route to c3-ref).
  Requires .c3/ to exist.
---

# C3 Query - Architecture Navigation

Navigate C3 docs AND explore corresponding code. Full context = docs + code.

**Relationship to c3-navigator agent:** This skill defines the query workflow. The `c3-navigator` agent implements this workflow. Use the agent when spawning via Task tool; use this skill directly for inline execution.

## Precondition: C3 Adopted

Run `npx -y @lagz0ne/c3x list --json` via Bash. If it fails or returns empty, **STOP**.

If missing:
> This project doesn't have C3 docs yet. Use the c3-onboard skill to create documentation first.

Do NOT proceed until topology is confirmed.

## REQUIRED: Load References

Before proceeding, Read these files (relative to this skill's directory):
1. `references/skill-harness.md` - Red flags and complexity rules
2. `references/layer-navigation.md` - How to traverse C3 docs
3. `references/constraint-chain.md` - Constraint chain query procedure (for "what constraints apply to X")

## Query Flow

```
Query → Load Topology (CLI) → Clarify Intent → Navigate Layers → Extract References → Explore Code
                                     │
                                     └── Use AskUserQuestion if ambiguous
```

## Progress Checklist

```
Query Progress:
- [ ] Step 0a: Topology loaded via `npx -y @lagz0ne/c3x list --json`
- [ ] Step 0b: Intent clarified (or skipped if specific)
- [ ] Step 1: Context navigated (found relevant container/component from JSON)
- [ ] Step 2: References extracted (code paths, symbols — Read docs for body content)
- [ ] Step 3: Code explored (verified against docs)
- [ ] Response delivered with layer, refs, insights
```

---

## Step 0a: Load Topology

**FIRST action:** Run `npx -y @lagz0ne/c3x list --json` via Bash to load full topology.

```bash
npx -y @lagz0ne/c3x list --json
```

The JSON output contains every entity's **id, type, title, path, relationships, and frontmatter**. Use this to:
- Identify which containers, components, refs, and ADRs exist
- Match the user's query to relevant entities by title, type, or relationship
- Resolve C3 IDs (c3-N, c3-NNN, adr-*, ref-*) to file paths

**Do NOT** manually Glob or Read `.c3/` directory listings. The JSON has everything needed for discovery.

Only use the Read tool **after** you've identified specific entities from the JSON — and only when you need body content (prose, Code References sections, edge cases) that isn't in the frontmatter.

---

## Step 0b: Clarify Intent

**Ask when:**
- Query vague ("how does X work?" - which aspect?)
- Multiple interpretations ("authentication" - login? tokens? sessions?)
- Scope unclear ("frontend" - whole container or specific component?)

**Skip when:**
- Query includes C3 ID (c3-102)
- Query specific ("where is login form submitted?")
- User says "show me everything about X"

---

## Step 1: Navigate Layers

Use the JSON topology from Step 0a to navigate: **Context → Container → Component**

1. **Match from JSON** — Find relevant entities by title, type, or relationships in the JSON output
2. **Read for depth** — Only Read entity files when you need body content (prose, `## Code References`, edge cases) not available in the JSON frontmatter

| Doc Section | Extract For Code |
|-------------|------------------|
| Component name | Class/module names |
| `## Code References` | Direct file paths, symbols |
| Technology | Framework patterns |
| Entry points | Main files, handlers |

---

## Step 2: Extract References

From the identified component(s), extract:
- File paths from `## Code References`
- Related patterns from `## Related Refs`

### Reference Lookup

If query relates to patterns/conventions:
1. Find matching `ref-*` entities from the JSON topology (loaded in Step 0a)
2. Read ref file for body content, return ref content + citing components

---

## Step 3: Explore Code

Use extracted references:
- **Glob**: `src/auth/**/*.ts`
- **Grep**: Class names, functions
- **Read**: Specific files from `## Code References`

---

## Query Types

| Type | User Says | Response |
|------|-----------|----------|
| Docs | "where is X", "explain X" | Docs + suggest code exploration |
| Code | "show me code for X" | Full flow through code |
| Deep | "explore X thoroughly" | Docs + Code + Related |
| Constraints | "what rules/constraints apply to X" | Full constraint chain |

---

## Constraint Chain Query

For constraint queries ("what constraints apply to X?"), follow the procedure in `references/constraint-chain.md`. Read upward through the hierarchy, collect cited refs, and synthesize the full constraint chain.

## Edge Cases

| Situation | Action |
|-----------|--------|
| Topic not in C3 docs | Search code directly, suggest documenting if significant |
| Spans multiple containers | List all affected containers, explain relationships |
| Docs seem stale | Note discrepancy, suggest running c3-audit or c3-change skill to fix |

---

## Examples

**Example 1: Component lookup**
```
User: "Where is authentication handled?"

Step 0a: Run `npx -y @lagz0ne/c3x list --json` → JSON shows c3-2-api container, c3-201-auth-middleware component with title "Auth Middleware"
Step 1: Match "authentication" → c3-201-auth-middleware (from JSON relationships + title)
Step 2: Read .c3/c3-2-api/c3-201-auth-middleware.md → Get code refs (need body content)

Response:
**Layer:** c3-201 (Auth Middleware)

Authentication is handled by the AuthMiddleware component in the API container.

**Code References:**
- `src/api/middleware/auth.ts` - JWT validation
- `src/api/middleware/session.ts` - Session management

**Key Insights:**
Uses RS256 for JWT signing. Sessions stored in Redis.

**Related:** See c3-202-user-service for user lookup
```

**Example 2: Pattern lookup**
```
User: "How does error handling work?"

Step 0a: Run `npx -y @lagz0ne/c3x list --json` → JSON shows ref-error-handling entity, cited by c3-201, c3-203, c3-205
Step 1: Match "error handling" → ref-error-handling (from JSON title + relationships)

Response:
**Pattern:** ref-error-handling

Error handling follows a centralized pattern documented in refs.

**Pattern Summary:**
- All errors extend BaseError class
- HTTP errors map to status codes via errorToStatus()
- Logged with correlation ID

**Used By:** c3-201, c3-203, c3-205
```

---

## Response Format

```
**Layer:** <c3-id> (<name>)

<Architecture from docs>

**Code References:**
- `path/file.ts` - <role>

**Key Insights:**
<Observations from code>

**Related:** <navigation hints>
```

For constraint chain queries, use the response format in `references/constraint-chain.md` instead.
