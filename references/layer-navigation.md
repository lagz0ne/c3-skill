# Layer Navigation Pattern

## Activation Check

Only proceed if `.c3/README.md` exists. Otherwise, suggest `/onboard`.

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

### 1. Load Context

Read `.c3/README.md`:
- Scan inventory for matching container(s)
- Check linkages for relationship context
- Identify which container(s) own the concept

### 2. Load Container(s)

Read `.c3/c3-N-*/README.md`:
- Scan component inventory
- Match by name, category, or description
- Note component IDs and roles
- Check fulfillment section

### 3. Load Component(s) (if exists)

Read `.c3/c3-N-*/c3-NNN-*.md`:
- Implementation details
- Edge cases documented
- Code references (paths, classes, patterns)

## ID-to-Path Quick Reference

| Pattern | File |
|---------|------|
| `c3-0` | `.c3/README.md` |
| `c3-N` | `.c3/c3-N-*/README.md` |
| `c3-NNN` | `.c3/c3-N-*/c3-NNN-*.md` |
| `adr-*` | `.c3/adr/adr-*.md` |

See `references/v3-structure.md` for complete patterns.
