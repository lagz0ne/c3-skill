---
name: c3-query
description: |
  Navigates C3 architecture docs and explores corresponding code to answer architecture questions.
  Use when the user asks "where is X", "how does X work", "explain X", "show me the architecture",
  "find component", "what handles X", or references C3 IDs (c3-0, c3-1, adr-*).
  Requires .c3/ to exist. For changes, route to c3-alter instead.
---

# C3 Query - Architecture Navigation

Navigate C3 docs AND explore corresponding code. Full context = docs + code.

## REQUIRED: Load References

Before proceeding, load these files:
1. `../../references/skill-harness.md` - Red flags and complexity rules
2. `../../references/layer-navigation.md` - How to traverse C3 docs

## Query Flow

```
Query → Clarify Intent → Navigate Layers → Extract References → Explore Code
              │
              └── Use AskUserQuestion if ambiguous
```

## Progress Checklist

```
Query Progress:
- [ ] Step 0: Intent clarified (or skipped if specific)
- [ ] Step 1: Context navigated (found relevant container)
- [ ] Step 2: References extracted (code paths, symbols)
- [ ] Step 3: Code explored (verified against docs)
- [ ] Response delivered with layer, refs, insights
```

---

## Step 0: Clarify Intent

**Ask when:**
- Query vague ("how does X work?" - which aspect?)
- Multiple interpretations ("authentication" - login? tokens? sessions?)
- Scope unclear ("frontend" - whole container or specific component?)

**Skip when:**
- Query includes C3 ID (c3-102)
- Query specific ("where is login form submitted?")
- User says "show me everything about X"

---

## Step 1-2: Navigate and Extract

Follow layer navigation: **Context → Container → Component**

| Doc Section | Extract For Code |
|-------------|------------------|
| Component name | Class/module names |
| `## References` | Direct file paths, symbols |
| Technology | Framework patterns |
| Entry points | Main files, handlers |

### Reference Lookup

If query relates to patterns/conventions:
1. Check `.c3/refs/` for `ref-*` matching topic
2. Return ref content + citing components

---

## Step 3: Explore Code

Use extracted references:
- **Glob**: `src/auth/**/*.ts`
- **Grep**: Class names, functions
- **Read**: Specific files from `## References`

---

## Query Types

| Type | User Says | Response |
|------|-----------|----------|
| Docs | "where is X", "explain X" | Docs + suggest code exploration |
| Code | "show me code for X" | Full flow through code |
| Deep | "explore X thoroughly" | Docs + Code + Related |

## Edge Cases

| Situation | Action |
|-----------|--------|
| Topic not in C3 docs | Search code directly, suggest documenting if significant |
| Spans multiple containers | List all affected containers, explain relationships |
| Docs seem stale | Note discrepancy, suggest `/c3 audit` or `/alter` to fix |

---

## Examples

**Example 1: Component lookup**
```
User: "Where is authentication handled?"

Step 1: Read .c3/README.md → Find c3-2-api container handles auth
Step 2: Read .c3/c3-2-api/README.md → Find c3-201-auth-middleware
Step 3: Read .c3/c3-2-api/c3-201-auth-middleware.md → Get code refs

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

Step 1: Read .c3/README.md → Multiple containers
Step 2: Check .c3/refs/ → Find ref-error-handling.md

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
