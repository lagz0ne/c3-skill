---
name: c3-query
description: |
  Use when project has .c3/ and user asks about architecture or code location.
  Triggers: "explain", "where is", "look for", "how does", "what is", "find", "show me", C3 IDs (c3-0, c3-1, adr-*).
---

# C3 Query - Architecture Navigation + Code Exploration

Navigate architecture docs AND explore corresponding code. Full context = docs + code.

## NOT This Skill

| Intent | Use Instead |
|--------|-------------|
| Change code/architecture | `c3-alter` |
| Set up C3 from scratch | `c3` (Adopt mode) or `/onboard` |
| Validate C3 docs | `c3` (Audit mode) |

## REQUIRED: Load Navigation Reference

**Before navigating layers, you MUST read `../../references/layer-navigation.md`.**

This reference contains:
- Activation check
- Traversal order (Context → Container → Component)
- ID-to-path mapping

## Critical: Always Use AskUserQuestionTool

When clarifying intent, MUST use the `AskUserQuestion` tool - never plain text questions.
- Structured questions get structured answers
- Multiple-choice reduces ambiguity
- User can select rather than type

## Core Principle

To understand fully, you need BOTH:
- **C3 Docs**: What it is, why it exists, how it connects
- **Code**: How it's actually implemented

## Query Resolution Flow

```
Query
  │
  ├─→ 0. Clarify Intent (Socratic)
  │      If ambiguous → ask until clear
  │
  ├─→ 1. Navigate Layers (per layer-navigation.md)
  │      Context → Container → Component
  │
  ├─→ 2. Extract Code References
  │      From component docs, extract:
  │      - File paths, class/function names
  │      - Directory patterns, technology keywords
  │
  └─→ 3. Explore Code
         Use extracted references to read/search actual code
```

## Step 0: Clarify Intent (Socratic)

Before searching, ensure you understand what the user is looking for.

**Ask clarifying questions when:**
- Query is vague ("how does X work?" - which aspect?)
- Multiple interpretations exist ("authentication" - login flow? token validation? session management?)
- Scope is unclear ("frontend" - the whole container? a specific component?)

**Example Socratic questions:**
- "Are you looking for how it's documented or how it's implemented in code?"
- "Which aspect of authentication: login flow, token refresh, or permission checks?"
- "Do you need the high-level architecture or specific implementation details?"

**Continue until intent is clear.** Don't guess - a wrong interpretation wastes time.

**Skip clarification when:**
- Query includes a C3 ID (c3-102) - direct lookup
- Query is specific ("where is the login form submitted?")
- User says "just show me everything about X"

## Step 1-2: Navigate and Extract

Follow layer-navigation.md to traverse layers, then extract code references:

From C3 docs, extract keywords for code exploration:

| Doc Section | Code References |
|-------------|-----------------|
| Component name | Class names, module names |
| File paths | Direct paths to read |
| Technology | Framework patterns to search |
| Entry points | Main files, handlers, routes |
| Dependencies | Import patterns |

Look for:
- Explicit paths: `src/auth/service.ts`
- Patterns: `**/auth/**`, `*.controller.ts`
- Keywords: class names, function names, route patterns
- Config files: related configs, env vars

## Step 3: Explore Code

Use Task tool with Explore agent OR direct tools:
- **Glob**: Find files matching patterns from docs
- **Grep**: Search for keywords, class names, functions
- **Read**: Load specific files mentioned in docs

Build understanding from both docs and code.

## Query Types

### Documentation Query (default)
"where is X", "explain X" → Docs only, suggest code exploration

### Code Query
"show me the code for X", "implementation of X" → Full flow through code

### Deep Query
"explore X thoroughly", "understand X completely" → Docs + Code + Related

## Response Format

```
**Layer:** <layer-id> (<name>)

<Architecture understanding from docs>

**Code References:**
- `path/to/file.ts` - <what it does>
- `ClassName` in `path/` - <role>

**Key Code Insights:**
<Observations from actual code exploration>

**Related:** <navigation hints>
```

## Examples

**Query:** "where is authentication handled?"
1. Load `.c3/README.md` → find auth in c3-1 (backend)
2. Load `.c3/c3-1-backend/README.md` → find c3-101 (auth-service)
3. Extract: `src/auth/`, `AuthService`, `jwt`
4. Answer with layer + code pointers

**Query:** "show me the code for payment processing"
1. Locate: Context → c3-1 (backend) → c3-103 (payments)
2. Extract from docs: `src/payments/`, `PaymentService`, `Stripe`
3. Glob: `src/payments/**/*.ts`
4. Read key files, summarize implementation
5. Answer with docs context + code walkthrough

**Query:** "how does frontend call the API?"
1. Load Context → check c3-1 ↔ c3-2 linkage
2. Load both containers for integration details
3. Extract: API client patterns, endpoint definitions
4. Search: `fetch`, `axios`, route handlers
5. Answer with protocol from docs + actual implementation

## Navigation Hints

| User Intent | Suggest |
|-------------|---------|
| Wants code | "Let me explore the code at <paths>" |
| Wants detail | "See c3-NNN for implementation docs" |
| Wants context | "See c3-0 for system overview" |
| Wants history | "See adr-* for decision rationale" |
