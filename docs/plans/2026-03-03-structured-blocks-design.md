# Structured Blocks

> Structured blocks give architecture facts stable, typed addresses — so retrieving design knowledge is a direct lookup, not a document comprehension task.

## Problem

.c3/ docs have two layers:

1. **Frontmatter (structured)** — YAML with typed fields (`id`, `type`, `parent`, `uses`). Machine-readable, queryable, validated.
2. **Body (unstructured)** — Markdown sections with implicit structure. Tables have columns, code blocks have languages, but none of this is declared. Retrieving a fact from the body requires reading the file, finding the section, parsing the format, and interpreting the content.

The body contains the most useful information (dependencies, code references, pattern decisions, constraints), but it has no addresses. You can't ask "give me c3-101's dependencies" — you must read the file and find them yourself.

## Solution

Extend the structured frontier from frontmatter into the body. Specific `## Sections` become **structured blocks** — declared in the schema registry with type, format, and purpose. The CLI can parse, validate, extract, and serve them.

A structured block is identified by its heading name + the entity type it belongs to. No extra syntax in the markdown. The schema registry is the source of truth.

## Block Types

**text** — Free-form prose. Validated for non-empty when required.

**table** — Pipe-delimited markdown table with declared columns. Column types enable validation (enum values, entity ID references, file paths).

**code** — Fenced code blocks. Language tag preserved. Extractable as raw code.

## Schema Registry

Enhanced `schema.go` with purpose strings and richer column definitions:

```go
type SectionDef struct {
    Name        string      // Heading name (e.g., "Dependencies")
    ContentType string      // "text", "table", "code"
    Required    bool
    Purpose     string      // What question this block answers
    Columns     []ColumnDef // For tables
}

type ColumnDef struct {
    Name   string   // Column header
    Type   string   // "text", "enum", "entity_id", "ref_id", "filepath"
    Values []string // For enums: allowed values
}
```

Each entity type (component, container, ref, context, adr) declares its expected blocks with purposes:

| Entity Type | Block | Type | Purpose |
|-------------|-------|------|---------|
| component | Goal | text | What this component exists to do |
| component | Dependencies | table | What it consumes and produces |
| component | Code References | table | Which files implement this |
| component | Related Refs | table | Which patterns govern this component |
| component | Container Connection | text | Why this component matters to its container |
| container | Goal | text | What this container exists to do |
| container | Components | table | What parts make up this container |
| container | Responsibilities | text | What this container owns |
| container | Complexity Assessment | text | Where the hard parts are and why |
| ref | Goal | text | Why this pattern exists |
| ref | Choice | text | What was decided |
| ref | Why | text | Reasoning behind the decision |
| ref | How | code | Working code to follow the pattern |
| context | Goal | text | What this system exists to do |
| context | Containers | table | What major parts make up the system |
| context | Abstract Constraints | table | Non-negotiable rules across containers |

## Query: `c3x query`

Simple addressing. Give an ID, get the content.

### No args — catalog

```bash
c3x query
```

Lists all entities with their blocks, fill status, and purpose. Teaches the consumer what's available and why.

### Entity ID — snapshot

```bash
c3x query c3-101
```

Returns all structured blocks for that entity, pre-extracted. Tables as JSON arrays, text as strings, code as raw code.

### Entity + section — single block

```bash
c3x query c3-101 dependencies
```

Returns just the dependencies table for c3-101.

### File path — resolve and extract

```bash
c3x query cli/cmd/set.go
```

Resolves file to owning component via codemap, returns component snapshot.

### Chain — constraint traversal

```bash
c3x query c3-101 --chain
```

Traverses component -> container -> context -> cited refs. Returns the full governance context.

### Cross-entity — future extension

```bash
c3x query '*.dependencies'
c3x query 'refs.*.choice'
```

Same block across many entities. Add when needed, not now.

## What This Enables

1. **LLM working on a file** — `c3x query <file>` gives full context in one call instead of 4-6 file reads
2. **Impact assessment** — `c3x query <id> --chain` shows everything that governs a component
3. **Architecture overview** — `c3x query` catalogs all queryable knowledge
4. **Doc authoring** — empty blocks show what needs filling; schema tells you the format
5. **Validation** — `c3x check` uses the same schema to validate block content quality

## Implementation Phases

### Phase 1: Schema enhancement
- Add `Purpose` field to `SectionDef` in `schema.go`
- Define purpose strings for all existing blocks
- No new commands yet — just enriching the registry

### Phase 2: `c3x query <id>` — entity snapshot
- New `query.go` command
- Walks entity, parses all sections via existing `ParseSections` + `ParseTable`
- Returns structured output (JSON for tables, text for prose, raw for code)
- Smart format detection based on schema ContentType

### Phase 3: `c3x query` — catalog
- No args = list all entities with blocks, fill status, purpose
- Uses schema registry for purpose strings
- Shows what's queryable and why

### Phase 4: `c3x query <id> <section>` — single block extraction
- Extract one specific block from an entity
- Subset of Phase 2 output

### Phase 5: `c3x query <file>` — file resolution
- Detect file path input (contains `/` or `.`)
- Resolve via codemap to component
- Return component snapshot

### Phase 6: `c3x query <id> --chain` — constraint traversal
- Walk up: component -> container -> context
- Collect cited refs
- Return assembled chain

### Future: Cross-entity, filtering, validation queries
- `*.dependencies`, `refs.*.choice` etc.
- Add when real use cases demand it
