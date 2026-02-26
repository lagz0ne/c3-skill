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

Native Go binary, bundled with the plugin. Auto-downloads from GitHub releases if missing.

```
c3x <command> [args] [options]

  init                       Scaffold .c3/ skeleton
  list                       Topology view with relationships
  check                      Validate docs, schema, code refs, consistency
  add <type> <slug>          Create entity (auto-numbering + wiring)
  set <id> <field> <value>   Update frontmatter field or section content
  wire <src> cite <tgt>      Link component to ref (3-sided atomic update)
  unwire <src> cite <tgt>    Remove cite link (3-sided)
  schema <type>              Show known sections and column types

  --json                     Machine-readable output
  --c3-dir <path>            Override .c3/ auto-detection
```

The CLI implements a three-layer document engine:

| Layer | What it validates |
|-------|-------------------|
| **1. Structure** | Broken links, orphans, duplicates, missing parents |
| **2. Schema** | Required sections per entity type (Goal, Components, Dependencies, etc.) |
| **3. Types** | Column types in tables: `filepath` exists on disk, `entity_id` in graph, `ref_id` valid, `enum` in allowed set |

## The `.c3/` Directory

```
.c3/
├── README.md                  # System context (c3-0)
├── c3-N-name/                 # Container
│   ├── README.md              # Container overview + component table
│   └── c3-NNN-component.md   # Component with code refs, deps, wiring
├── refs/
│   └── ref-pattern.md         # Cross-cutting convention with cited-by tracking
└── adr/
    └── adr-YYYYMMDD-slug.md   # Architecture decision record
```

Every entity has YAML frontmatter (`id`, `type`, `refs[]`, `status`) and markdown body with schema-defined sections. The CLI keeps wiring consistent — `wire`/`unwire` updates source frontmatter, source Related Refs table, and target Cited By table atomically.

## Development

```bash
cd cli && go test ./...       # Run tests
bash scripts/build.sh         # Cross-compile for 4 targets
```

Binaries are cross-compiled in CI and force-added to `main`. The `dev` branch has source only.

## License

MIT
