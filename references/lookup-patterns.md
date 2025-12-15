# C3 Lookup Patterns

Quick reference for ID-based document retrieval in `.c3/` documentation.

## Document ID Patterns (v3)

| Pattern | Level | Example | File Path |
|---------|-------|---------|-----------|
| `c3-0` | Context | c3-0 | `.c3/README.md` |
| `c3-{N}` | Container | c3-1 | `.c3/c3-{N}-*/README.md` |
| `c3-{N}{NN}` | Component | c3-101 | `.c3/c3-{N}-*/c3-{N}{NN}-*.md` |
| `adr-{YYYYMMDD}-{slug}` | Decision | adr-20251124-caching | `.c3/adr/adr-{YYYYMMDD}-{slug}.md` |
| `#c3-xxx-section` | Heading | #c3-1-middleware | (within any doc) |

## Bash Lookup Commands

### By Document ID

```bash
# Context (c3-0)
cat .c3/README.md

# Container (c3-1, c3-2, etc.)
cat .c3/c3-1-*/README.md

# Component (c3-101, c3-215, etc.)
cat .c3/c3-1-*/c3-101-*.md
cat .c3/c3-2-*/c3-215-*.md

# ADR
cat .c3/adr/adr-20251124-caching.md
```

### Finding Documents

```bash
# Find all containers
ls -d .c3/c3-[1-9]-*/

# Find all components in container 2
ls .c3/c3-2-*/c3-2*.md

# Find all ADRs
ls .c3/adr/adr-*.md

# Find ADRs by year
ls .c3/adr/adr-2025*.md
```

### Extracting Frontmatter

```bash
# Extract frontmatter from any document
awk '/^---$/,/^---$/ {print}' .c3/c3-1-backend/README.md
```

### Finding Heading by ID

```bash
# Find heading and extract content until next heading
awk -v hid="c3-1-middleware" '
  $0 ~ "{#" hid "}" {
    found = 1
    print
    next
  }
  found && /^## / { exit }
  found { print }
' .c3/c3-1-backend/README.md
```

## ID Resolution Logic

Parse the ID to determine document type and path:

1. **Identify ID pattern:**
   - `c3-0` → Context (root README)
   - `c3-{N}` (single digit) → Container
   - `c3-{N}{NN}` (3+ digits) → Component
   - `adr-{YYYYMMDD}-{slug}` → ADR decision

2. **Determine file path:**
   - `c3-0` → `.c3/README.md`
   - `c3-{N}` → `.c3/c3-{N}-*/README.md`
   - `c3-{N}{NN}` → `.c3/c3-{N}-*/c3-{N}{NN}-*.md`
   - `adr-*` → `.c3/adr/adr-*.md`

## V2 Fallback (Backward Compatibility)

| Pattern | Level | Example | File Path |
|---------|-------|---------|-----------|
| `context` | Primary Context | context | `.c3/README.md` |
| `CTX-slug` | Auxiliary Context | CTX-actors | `.c3/CTX-actors.md` |
| `C3-{N}-slug` | Container | C3-1-backend | `.c3/containers/C3-1-backend.md` |
| `C3-{N}{NN}-slug` | Component | C3-102-auth | `.c3/components/C3-102-auth.md` |
| `ADR-###-slug` | Decision | ADR-003-cache | `.c3/adr/ADR-###-slug.md` |

```bash
# V2 fallback patterns
find .c3/containers -name "C3-[0-9]-*.md" 2>/dev/null
find .c3/components -name "C3-[0-9][0-9][0-9]-*.md" 2>/dev/null
find .c3/adr -name "ADR-*.md" 2>/dev/null
```

## Usage During Exploration

When investigating a hypothesis:

```
Hypothesis: "This affects c3-1 middleware"

Exploration:
1. cat .c3/c3-1-*/README.md
   → Get overview, see what components exist

2. Find middleware section (grep for heading ID)
   → Read middleware section details

3. cat .c3/c3-1-*/c3-101-*.md
   → Deeper exploration of specific component

4. Discover: "This actually touches c3-101-auth-middleware"
```

## Related

- [v3-structure.md](v3-structure.md) - Full structure reference
- [naming-conventions.md](naming-conventions.md) - Naming patterns
