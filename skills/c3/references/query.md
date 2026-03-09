# Query Reference

Navigate C3 docs and corresponding code. Full context = docs + code together.

## Flow

`Topology → Match Entity → Read Docs → Explore Code → Respond`

## Progress

- [ ] Topology loaded (`c3x list --json`)
- [ ] Entity matched
- [ ] Docs read (component, container, refs)
- [ ] Code explored
- [ ] Response delivered

---

## Step 1: Load Topology

```bash
bash <skill-dir>/bin/c3x.sh list --json
```

Returns all entities with IDs, types, titles, paths, relationships, frontmatter. Match the user's query to entities by title, type, or relationship.

Don't manually Glob/Read `.c3/` for discovery — the JSON has everything. Only read individual docs after identifying them.

Also check for **recipes** in the topology. If a recipe matches the query, read it and trace its sources as the narrative.

## Step 2: Clarify Intent (if needed)

Ask when (skip if ASSUMPTION_MODE):
- Query is vague ("how does X work?" with multiple interpretations)
- Multiple entities could match
- Scope is unclear

Skip when: specific C3 ID given, query is precise, "show me everything about X".

## Step 3: Navigate and Read

Navigate top-down: Context → Container → Component.

Read the matched entity doc(s) for architecture context. Check the Related Refs table in component docs — read those refs for constraints and patterns.

## Step 4: Explore Code

How deeply you explore code depends on the query type:

| Query Type | Approach |
|------------|----------|
| **"Where is X?"** / **"Explain X"** | Read docs, reference code-map entries, suggest code files to explore |
| **"Trace X"** / **"Show me the code"** | Read docs for context, then follow the code path through source files |
| **"What constraints apply to X?"** | Read upward through entity chain + all cited refs |
| **"Explore X thoroughly"** | Full treatment: docs + code + related entities |

**Code navigation: LSP first.** When tracing code paths, always start with LSP tools (go-to-definition, find-references, hover, workspace symbols) to follow function calls, resolve types, and find usages. LSP gives precise, type-aware results that are more reliable than text search. Only fall back to Grep/Glob when LSP is unavailable or returns no results.

Use `c3x lookup <file>` when you want to understand which component/refs govern a file — but don't treat it as a gate before every file read. It's most valuable when you encounter unfamiliar files or need to check constraints.

## Step 5: Respond

```
**Layer:** <c3-id> (<name>)

<Architecture from docs>

**Code Map:** `path/file.ts` - <role>

**Key Insights:** <Observations from code>

**Related:** <navigation hints to related entities>
```

---

## Constraint Chain Query

When the user asks "what rules apply to X":

1. Identify target entity
2. Read upward: component → container → context
3. Extract: explicit constraints (MUST/MUST NOT), boundaries, layer rules
4. Collect cited refs, read key rules

```
**Constraint Chain for c3-NNN (Name)**

**Context (c3-0):** [system-wide rule]
**Container (c3-N):** [container boundary]
**Patterns:** ref-X: [key rules]
**Layer Boundaries:** MAY: [...] MUST NOT: [...]
```

## Edge Cases

| Situation | Action |
|-----------|--------|
| Topic not in C3 | Search code directly, suggest documenting |
| Spans containers | List all affected, explain relationships |
| Docs seem stale | Note, suggest audit |

## ID → Path

| Pattern | File |
|---------|------|
| `c3-0` | `.c3/README.md` |
| `c3-N` | `.c3/c3-N-*/README.md` |
| `c3-NNN` | `.c3/c3-N-*/c3-NNN-*.md` |
| `adr-*` | `.c3/adr/adr-*.md` |
| `ref-*` | `.c3/refs/ref-*.md` |
| `recipe-*` | `.c3/recipes/recipe-*.md` |
