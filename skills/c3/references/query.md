# Query Reference

Navigate C3 docs AND explore corresponding code. Full context = docs + code.

## Query Flow

```
Query -> Load Topology (CLI) -> Clarify Intent -> Navigate Layers -> Extract References -> Explore Code
```

## Progress Checklist

```
Query Progress:
- [ ] Topology loaded via `c3x list --json`
- [ ] Intent clarified (or skipped if specific)
- [ ] Context navigated (found relevant entity from JSON)
- [ ] References extracted (code paths, symbols)
- [ ] Code explored (verified against docs)
- [ ] Response delivered
```

---

## Step 0a: Load Topology

**FIRST action:**
```bash
bash <skill-dir>/bin/c3x.sh list --json
```

JSON contains every entity's id, type, title, path, relationships, frontmatter. Use to:
- Identify containers, components, refs, ADRs
- Match query to relevant entities by title/type/relationship
- Resolve C3 IDs (c3-N, c3-NNN, adr-*, ref-*) to file paths

**Do NOT** manually Glob or Read `.c3/` directory. JSON has everything for discovery.

Only Read **after** identifying specific entities — when body content (prose, edge cases) is needed. Use `.c3/code-map.yaml` for file paths.

## Step 0b: Clarify Intent

**Ask when** (skip if ASSUMPTION_MODE):
- Query vague ("how does X work?" — which aspect?)
- Multiple interpretations ("authentication" — login? tokens? sessions?)
- Scope unclear ("frontend" — whole container or specific component?)

**Skip when:**
- Query includes C3 ID (c3-102)
- Query specific ("where is login form submitted?")
- User says "show me everything about X"

## Step 1: Navigate Layers

Top-down: Context -> Container -> Component

1. **Match from JSON** — find relevant entities by title, type, relationships
2. **Read for depth** — only Read entity files when body content not in JSON frontmatter

| Source | Extract For Code |
|--------|------------------|
| Component name | Class/module names |
| `.c3/code-map.yaml` | Direct file paths, symbols |
| Technology | Framework patterns |
| Entry points | Main files, handlers |

## Step 2: Extract References

From identified component(s):
- File paths from `.c3/code-map.yaml` (look up component ID)
- Related patterns from `## Related Refs`

**Ref lookup:** Find matching `ref-*` entities from JSON. Read ref file for body content. Return ref content + citing components.

## Step 3: Explore Code

Use extracted references:
- **Glob:** `src/auth/**/*.ts`
- **Grep:** class names, functions
- **Read:** specific files from `.c3/code-map.yaml`

---

## Query Types

| Type | User Says | Response |
|------|-----------|----------|
| Docs | "where is X", "explain X" | Docs + suggest code exploration |
| Code | "show me code for X" | Full flow through code |
| Deep | "explore X thoroughly" | Docs + Code + Related |
| Constraints | "what rules apply to X" | Full constraint chain |

## Constraint Chain Query

For constraint queries ("what constraints apply to X?"):

1. Identify target layer (c3-NNN, c3-N, or c3-0)
2. Read upward through hierarchy:
   - Component -> container -> context
   - Container -> context
3. At each level, extract:
   - Explicit constraints ("MUST...", "MUST NOT...")
   - Implicit boundaries (responsibilities)
   - Layer-specific rules
4. Collect cited refs from `Related Refs` sections
5. Read each cited ref, extract key rules

**Response format:**
```
**Constraint Chain for c3-NNN (Name)**

**Context Constraints (c3-0):**
- [System-wide rule]

**Container Constraints (c3-N):**
- [Container boundary]

**Cited Pattern Constraints:**
- **ref-X:** [Key rules]

**Layer Boundaries:**
This component MAY: [permitted]
This component MUST NOT: [prohibited]
```

## Edge Cases

| Situation | Action |
|-----------|--------|
| Topic not in C3 docs | Search code directly, suggest documenting |
| Spans multiple containers | List all affected, explain relationships |
| Docs seem stale | Note discrepancy, suggest audit |

## Response Format

```
**Layer:** <c3-id> (<name>)

<Architecture from docs>

**Code Map:**
- `path/file.ts` - <role>

**Key Insights:**
<Observations from code>

**Related:** <navigation hints>
```

## ID-to-Path Quick Reference

| Pattern | File |
|---------|------|
| `c3-0` | `.c3/README.md` |
| `c3-N` | `.c3/c3-N-*/README.md` |
| `c3-NNN` | `.c3/c3-N-*/c3-NNN-*.md` |
| `adr-*` | `.c3/adr/adr-*.md` |
| `ref-*` | `.c3/refs/ref-*.md` |

## Reference Resolution

When a pattern/convention is mentioned:
1. Check if component cites a `ref-*`
2. Look up ref in `.c3/refs/ref-{slug}.md`
3. Refs explain patterns; components explain usage

Hierarchy: Component (specific usage) -> Container (cross-component) -> Context (system-wide) -> Refs (pattern definitions)
