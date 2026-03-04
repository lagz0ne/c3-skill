# Query Reference

Navigate C3 docs + corresponding code. Full context = docs + code.

## Flow

`Query → Topology → Clarify → Navigate → Lookup → Explore Code`

## Progress

- [ ] Topology loaded (`c3x list --json`)
- [ ] Intent clarified (or skipped if specific)
- [ ] Entity matched from JSON
- [ ] `c3x lookup` run on every file path surfaced
- [ ] Code explored
- [ ] Response delivered

---

## Step 0a: Topology

**First action:**
```bash
bash <skill-dir>/bin/c3x.sh list --json
```
Returns all entities: id, type, title, path, relationships, frontmatter. Match query to entities by title/type/relationship.

Don't manually Glob/Read `.c3/`. JSON has everything for discovery. Read only after identifying specific entities.

## Step 0a+: Check Recipes

After loading topology, check for recipes that match the query:
1. Filter entities with type `recipe` from `c3x list --json`
2. Match query against recipe title + description
3. If match found → read recipe, serve sources as the narrative trace
4. If no match → proceed with normal query flow

## Step 0b: Clarify Intent

Ask when (skip if ASSUMPTION_MODE):
- Vague ("how does X work?")
- Multiple interpretations ("authentication" — login? tokens?)
- Scope unclear

Skip when: C3 ID given, query is specific, "show me everything about X".

## Step 1: Navigate Layers

Top-down: Context → Container → Component.

Match from JSON. Read entity files only when body content not in frontmatter.

| Source | Use For |
|--------|---------|
| Component name | Class/module names |
| code-map.yaml | Direct file paths, symbols |
| Technology | Framework patterns |

## Step 2: Extract + Lookup

For every file path encountered:
1. **Run `c3x lookup <file>` before reading any source file** — returns component + governing refs. For directory-level context, use `c3x lookup 'src/auth/**'`.
2. Read `## Related Refs` in component doc
3. Find `ref-*` entities from JSON. Read for body content.

Lookup-returned refs = constraints governing that file's code.

## Step 3: Explore Code

```bash
# Glob patterns
src/auth/**/*.ts
# Grep class/function names
# Read specific files from code-map.yaml
```

---

## Query Types

| Type | When | Response |
|------|------|---------|
| Docs | "where is X", "explain X" | Docs + suggest code |
| Code | "show me code for X" | Full flow through code |
| Deep | "explore X thoroughly" | Docs + Code + Related |
| Constraints | "what rules apply to X" | Full constraint chain |

## Constraint Chain Query

1. Identify target (c3-NNN, c3-N, or c3-0)
2. Read upward: component → container → context
3. Extract: explicit constraints (MUST/MUST NOT), boundaries, layer rules
4. Collect cited refs from Related Refs, read key rules

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

## Response Format

```
**Layer:** <c3-id> (<name>)

<Architecture from docs>

**Code Map:** `path/file.ts` - <role>

**Key Insights:** <Observations>

**Related:** <navigation hints>
```

## ID → Path

| Pattern | File |
|---------|------|
| `c3-0` | `.c3/README.md` |
| `c3-N` | `.c3/c3-N-*/README.md` |
| `c3-NNN` | `.c3/c3-N-*/c3-NNN-*.md` |
| `adr-*` | `.c3/adr/adr-*.md` |
| `ref-*` | `.c3/refs/ref-*.md` |
| `recipe-*` | `.c3/recipes/recipe-*.md` |
