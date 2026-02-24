# Layer Navigation Pattern

## Activation Check

Run `npx -y c3-kit list --json` to load topology. If it fails or returns empty, suggest the c3-onboard skill.

## Discovery via CLI

**Primary discovery:** Use `npx -y c3-kit list --json` output to identify all entities, their types, relationships, and frontmatter. This replaces manual Glob/Read for discovering what exists.

**Read for depth:** Only use the Read tool when you need body content (prose, Code References, edge cases) from a specific entity already identified in the JSON.

## Traversal Order

Always navigate top-down: Context → Container → Component.

```
Context (.c3/README.md)
│   WHAT containers exist
│   External dependencies
│   Inter-container linkages
│
└──→ Container (.c3/c3-N-*/README.md)
     │   WHAT components exist
     │   Internal relationships
     │   Fulfillment mappings (how Context links are handled)
     │
     └──→ Component (.c3/c3-N-*/c3-NNN-*.md)
          HOW implemented
          Edge cases
          Code references
```

## Step-by-Step

### 1. Load Topology (CLI)

Run `npx -y c3-kit list --json` to get all entities with id, type, title, path, relationships, and frontmatter. Use this to:
- Identify all containers, components, refs, and ADRs
- Match the query to relevant entities by title, type, or relationship
- Resolve C3 IDs to file paths

### 2. Match Entities (from JSON)

Using the JSON output:
- Match by entity title, type, or relationships
- Identify which container(s) own the concept
- Note component IDs, roles, and inter-entity links

### 3. Read for Depth (selective)

Read specific entity files only when body content is needed:
- Implementation details and prose
- Edge cases documented
- Code references (paths, classes, patterns)
- Fulfillment sections

## ID-to-Path Quick Reference

| Pattern | File |
|---------|------|
| `c3-0` | `.c3/README.md` |
| `c3-N` | `.c3/c3-N-*/README.md` |
| `c3-NNN` | `.c3/c3-N-*/c3-NNN-*.md` |
| `adr-*` | `.c3/adr/adr-*.md` |

Use the ID-to-Path table above to resolve C3 IDs to file paths.

## Reference Resolution

When navigating and a pattern/convention is mentioned:

1. Check if component cites a `ref-*`
2. Look up ref in `.c3/refs/ref-{slug}.md`
3. Refs explain patterns; components explain usage

**Resolution hierarchy:**
- Component: specific usage of patterns
- Container: cross-component patterns
- Context: system-wide patterns
- Refs: pattern definitions and rationale
