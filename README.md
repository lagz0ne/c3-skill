# C3: Architecture Documentation Toolkit

C3 (Context-Container-Component) is a Claude Code plugin that brings structured architecture documentation to any codebase. A single `/c3` skill backed by a native Go CLI manages everything in a `.c3/` directory.

![C3 Architecture](https://diashort.apps.quickable.co/e/492da1d8?theme=light)

## Install

```bash
claude plugin install lagz0ne/c3-skill
```

Then say `/c3 onboard` to create architecture docs for your project.

## What You Get

### The `/c3` Skill

Six operations, one entry point:

| Say | What happens |
|-----|-------------|
| `/c3` adopt this project | **onboard** — Socratic discovery → scaffolds `.c3/` with context, containers, components |
| `/c3` where is auth? | **query** — Navigates docs, traces relationships, answers architecture questions |
| `/c3` add rate limiting | **change** — ADR-first workflow: impact analysis → decision record → execute → audit |
| `/c3` create a ref for error handling | **ref** — Manage cross-cutting patterns as first-class docs with cite wiring |
| `/c3` audit the docs | **audit** — 10-phase validation: structural, semantic, drift, consistency |
| `/c3` what breaks if I change payments? | **sweep** — Impact assessment across entity graph |

### The `c3x` CLI

Native Go binary, bundled with the plugin. Version-stamped binaries (`c3x-{version}-{os}-{arch}`) ensure the correct binary is always used after plugin updates.

```
c3x <command> [args] [options]

  init                       Scaffold .c3/ skeleton
  list                       Topology view with relationships
  check                      Validate docs, schema, code refs, consistency
  add <type> <slug>          Create entity (container, component, ref, adr, recipe)
  set <id> <field> <value>   Update frontmatter field or section content
  wire <src> cite <tgt>      Link component to ref (3-sided atomic update)
  unwire <src> cite <tgt>    Remove cite link (3-sided)
  schema <type>              Show known sections and column types
  codemap                    Scaffold code-map.yaml with stubs for all components + refs
  lookup <file-path>         Map file to component(s) + refs
  coverage                   Code-map coverage stats

  --json                     Machine-readable output
  --include-adr              Include ADRs in output (hidden by default)
  --c3-dir <path>            Override .c3/ auto-detection
```

The CLI implements a three-layer document engine:

| Layer | What it validates |
|-------|-------------------|
| **0. Parse** | Broken YAML frontmatter (e.g. unquoted colons in values) |
| **1. Structure** | Broken links, orphans, duplicates, missing parents |
| **2. Schema** | Required sections per entity type (Goal, Components, Dependencies, etc.) |
| **3. Types** | Column types in tables: `filepath` exists on disk, `entity_id` in graph, `ref_id` valid, `enum` in allowed set |

## The `.c3/` Directory

```
.c3/
├── README.md                  # System context (c3-0)
├── code-map.yaml              # Component → source file mappings (validated by check)
├── _index/
│   ├── structural.md          # Precomputed entity→files→refs→constraints (auto-rebuilt)
│   └── notes/                 # LLM-generated cross-cutting topic notes
│       └── *.md               # e.g. authentication-flow.md, data-persistence.md
├── c3-N-name/                 # Container
│   ├── README.md              # Container overview + component table
│   └── c3-NNN-component.md   # Component with deps, wiring
├── refs/
│   └── ref-pattern.md         # Cross-cutting convention (golden examples, no code pointers)
├── recipes/
│   └── recipe-topic.md        # Cross-cutting concern trace (end-to-end narrative with source refs)
└── adr/
    └── adr-YYYYMMDD-slug.md   # Architecture decision record
```

Every entity has YAML frontmatter (`id`, `type`, `refs[]`, `status`) and markdown body with schema-defined sections. The CLI keeps wiring consistent — `wire`/`unwire` updates source frontmatter, source Related Refs table, and target Cited By table atomically.

`code-map.yaml` maps components and refs to their actual source files. Run `c3x codemap` to scaffold stubs for every entity, then fill in the glob patterns:

```yaml
# .c3/code-map.yaml
c3-101:  # Logger
  - src/lib/logger.ts
c3-102:  # Config
  - src/lib/config.ts
  - src/lib/config/**/*.ts
_exclude:
  - "**/*.test.ts"
  - "**/*.spec.ts"
  - dist/**
```

Patterns support `*` and `**` glob syntax, plus literal bracket paths like `[id]` (Next.js/SvelteKit routes). The `_exclude` key marks files that are intentionally unmapped (tests, build output) — they won't count against your coverage percentage.

`c3x check` validates all mappings: component IDs must exist in the graph, paths must be regular files on disk. `c3x coverage` shows how many project files are mapped, excluded, or unmapped. `c3x lookup <file>` resolves any file to its owning component(s) and governing refs.

### Structural Index

The CLI automatically maintains a structural index at `.c3/_index/structural.md` after mutating commands (`add`, `set`, `wire`, `unwire`). This precomputes entity→files→refs→reverse-deps→constraints mappings from the graph + code-map, giving LLMs instant discovery without multiple CLI calls.

Topic notes in `.c3/_index/notes/` are LLM-generated cross-cutting narratives (e.g. "authentication-flow.md", "data-persistence-strategy.md") with YAML frontmatter tracking their sources. `c3x check` validates that note source citations reference entities that exist in the graph.

## Development

```bash
cd cli && go test ./...       # Run tests
bash scripts/build.sh         # Cross-compile for 4 targets
```

Binaries are cross-compiled in CI and force-added to `main`. The `dev` branch has source only.

## License

MIT
