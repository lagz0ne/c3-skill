# C3 Navigator Agent - Implementation Plan (v2)

**Date:** 2026-01-19
**Goal:** Create a dedicated agent that answers questions using C3 architecture docs with sub-agent summarization

## Summary

Create `c3-navigator` agent that:
- Triggers on ANY question in projects with `.c3/` directory
- Runs in dedicated context window (token efficiency)
- Uses **sub-agent (haiku/sonnet) to analyze and summarize** `.c3/` content
- Produces adaptive output with diagrams via diashort

## Architecture

```
User Question
     │
     ▼
┌─────────────────────────────────────────┐
│          c3-navigator (main)            │
│  - Detects .c3/ exists                  │
│  - Reads TOC/README to identify scope   │
│  - Dispatches to summarizer sub-agent   │
└─────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────┐
│    c3-summarizer (sub-agent, haiku)     │
│  - Reads relevant .c3/ files            │
│  - Extracts key facts for the query     │
│  - Returns condensed summary            │
└─────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────┐
│          c3-navigator (main)            │
│  - Receives summary                     │
│  - Generates diagram from summary       │
│  - Formats response                     │
└─────────────────────────────────────────┘
                    │
                    ▼
            Response with diagram
```

## Implementation Steps

### Step 1: Create Main Agent File

**File:** `agents/c3-navigator.md`

```yaml
---
name: c3-navigator
description: |
  Use this agent when the user asks any question in a project containing .c3/ directory.
  This agent provides dedicated context for architecture queries with visual diagrams.
  Uses a sub-agent to efficiently analyze docs and optimize responses.

  <example>
  Context: User is in a project with .c3/ directory
  user: "How does the authentication system work?"
  assistant: "Let me use the c3-navigator agent to explore the architecture and provide a visual explanation."
  <commentary>
  Project has .c3/ and user asked a question - triggers navigator for dedicated architecture context.
  </commentary>
  </example>

  <example>
  Context: User is in a project with .c3/ directory
  user: "Where is the payment logic?"
  assistant: "I'll use the c3-navigator to trace through the C3 docs and locate the payment components."
  <commentary>
  "Where is X" pattern with .c3/ present - perfect trigger for navigator.
  </commentary>
  </example>

  <example>
  Context: User is in a project with .c3/ directory
  user: "Explain the data flow from frontend to database"
  assistant: "Using c3-navigator to map the data flow across containers with a visual diagram."
  <commentary>
  Cross-container question benefits from dedicated context and diagram generation.
  </commentary>
  </example>
model: inherit
color: cyan
tools: ["Read", "Glob", "Grep", "Bash", "Task"]
---
```

**Note:** Includes `Task` tool to dispatch sub-agent.

### Step 2: Create Summarizer Sub-Agent

**File:** `agents/c3-summarizer.md`

```yaml
---
name: c3-summarizer
description: |
  Internal sub-agent for c3-navigator. Analyzes .c3/ documentation and extracts
  key facts relevant to a specific query. Optimized for token efficiency.

  DO NOT use this agent directly - it is called by c3-navigator.
model: haiku
color: blue
tools: ["Read", "Glob", "Grep"]
---
```

**System prompt focuses on:**
- Reading specific .c3/ files
- Extracting only query-relevant information
- Returning condensed summary (not full docs)
- Identifying code references for follow-up

### Step 3: Write Navigator System Prompt

**Key behaviors:**

1. **Initial Assessment**
   - Read `.c3/README.md` to understand system context
   - Read `.c3/TOC.md` to identify relevant containers/components
   - Determine which docs are relevant to the query

2. **Dispatch Summarizer**
   ```
   Use Task tool with c3-summarizer:
   - Pass: query + list of relevant .c3/ file paths
   - Receive: condensed summary + code references
   ```

3. **Generate Diagram**
   - Based on summary, create Mermaid diagram
   - Submit to diashort
   - Include URL in response

4. **Format Response**
   - Structured: Layer, summary, code refs, diagram
   - Or conversational if query is simple

### Step 4: Write Summarizer System Prompt

**Behaviors:**

```markdown
You are the C3 Summarizer, optimized for extracting query-relevant facts from C3 docs.

**Input:** A user query and list of .c3/ file paths

**Process:**
1. Read each provided file
2. Extract ONLY information relevant to the query
3. Note code references (file paths, line numbers, symbols)
4. Identify related components for "see also"

**Output Format:**
```
## Summary
[2-4 sentences answering the query from docs]

## Key Components
- c3-XXX: [one-line purpose]

## Code References
- `path/file.ts:42` - [what it does]

## Related
- c3-YYY, c3-ZZZ (for follow-up)
```

**Constraints:**
- Output under 500 tokens
- Only facts from docs, no inference
- Explicit when info not found
```

### Step 5: Integrate diashort Diagram Generation

In navigator system prompt:

```markdown
**Diagram Generation:**

After receiving summarizer output, generate a visual:

1. Choose diagram type based on query:
   - "how does X work" → flowchart
   - "what is X" → component diagram
   - "data flow" → sequence diagram

2. Create Mermaid:
   ```mermaid
   graph TD
     [components from summary]
   ```

3. Render via diashort:
   ```bash
   curl -X POST https://diashort.apps.quickable.co/render \
     -H "Content-Type: application/json" \
     -d '{"source": "<mermaid>", "format": "mermaid"}'
   ```

4. Use https:// URL from response
```

### Step 6: Test Integration

1. Manual test in documented-api fixture
2. Verify sub-agent dispatches correctly
3. Verify summarizer returns condensed output
4. Verify diagram generates
5. Verify final response is cohesive

## File Outputs

| File | Purpose |
|------|---------|
| `agents/c3-navigator.md` | Main agent - triggers on questions, orchestrates |
| `agents/c3-summarizer.md` | Sub-agent - haiku, extracts/summarizes docs |

## Token Efficiency Analysis

| Approach | Est. Tokens |
|----------|-------------|
| Read all .c3/ in main context | 5000-20000 |
| Sub-agent summarization | 500-1500 |
| **Savings** | 70-90% |

Sub-agent with haiku:
- Cheap model for extraction
- Output is condensed
- Main agent receives only what's needed

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Sub-agent timeout | Set reasonable timeout, fallback to direct read |
| Summarizer misses key info | Include "verify against docs" instruction |
| Haiku quality insufficient | Option to use sonnet for complex queries |

## Definition of Done

- [ ] `agents/c3-navigator.md` exists with Task tool for sub-agent
- [ ] `agents/c3-summarizer.md` exists with haiku model
- [ ] Navigator triggers on questions in .c3/ projects
- [ ] Navigator dispatches to summarizer
- [ ] Summarizer returns condensed output (<500 tokens)
- [ ] Navigator generates diagram from summary
- [ ] Response is cohesive and accurate
- [ ] Manual testing passes
