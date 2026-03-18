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

Native Go binary, bundled with the plugin. Agents use it through the `/c3` skill automatically. Humans can use it directly via the `@c3x/cli` npm package:

```bash
npx @c3x/cli list              # run without installing
npm install -g @c3x/cli        # or install globally, then:
c3x list                       # use directly
```

`@c3x/cli` is a thin wrapper — it doesn't bundle the Go binary. Instead, it discovers an already-installed binary from your agent skill installations and delegates to it. This means you install the c3 skill once (via Claude or Codex), and both agents and humans share the same binary.

#### How resolution works

The CLI searches these locations for the c3x binary, picks the highest version found:

| Priority | Location | Source |
|----------|----------|--------|
| 1 | `<project>/skills/c3/bin/` | Local project (walks up from cwd, stops at `.git`) |
| 2 | `~/.claude/skills/c3/bin/` | Claude Code skill installation |
| 3 | `~/.codex/skills/c3/bin/` | Codex skill installation |
| 4 | `~/.claude/plugins/marketplaces/*/skills/c3/bin/` | Claude marketplace |

Use `--agent` to restrict the search:

```bash
c3x --agent claude list        # only Claude paths + project
c3x --agent codex list         # only Codex paths + project
```

#### Human vs agent output

When agents invoke c3x through the skill, `C3X_MODE=agent` is set automatically — output is JSON. When humans run `c3x` via the npm CLI, no mode is set — output is human-readable text. Explicit `--json` or `--compact` flags override either default.

```
c3x <command> [args] [options]

  init                       Scaffold .c3/ skeleton
  list                       Topology view with relationships
  check                      Validate docs, schema, code refs, consistency
  add <type> <slug>          Create entity (container, component, ref, adr, recipe)
  set <id> <field> <value>   Update frontmatter field or section content
  wire <src> [cite] <tgt>    Link component to ref (--remove to unlink)
  schema <type>              Show known sections and column types
  codemap                    Scaffold code-map.yaml with stubs for all components + refs
  lookup <file-path>         Map file to component(s) + refs
  coverage                   Code-map coverage + ref governance stats
  graph <entity-id>          Subgraph from entity (--format mermaid for diagrams)
  delete <id>                Remove entity + clean all references (--dry-run)
  capabilities               List all commands as a markdown table

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
| **4. Scope** | Ref scope cross-check: warns when a ref scopes a container but a child component doesn't cite it |

## The `.c3/` Directory

```
.c3/
├── README.md                  # System context (c3-0)
├── code-map.yaml              # Component → source file mappings (validated by check)
├── _index/
│   └── structural.md          # Precomputed entity→files→refs→constraints (auto-rebuilt)
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

Every entity has YAML frontmatter (`id`, `type`, `uses[]`, `status`) and markdown body with schema-defined sections. The CLI keeps wiring consistent — `wire` (and `wire --remove`) updates source frontmatter and source Related Refs table atomically.

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

The CLI automatically maintains a structural index at `.c3/_index/structural.md` after mutating commands (`add`, `set`, `wire`, `delete`). This precomputes entity→files→refs→reverse-deps→constraints mappings from the graph + code-map, giving LLMs instant discovery without multiple CLI calls.

Recipes in `.c3/recipes/` trace cross-cutting concerns end-to-end (e.g. "recipe-auth-flow.md") with YAML frontmatter tracking their entity sources. `c3x check` validates that recipe source citations reference entities that exist in the graph.

## VS Code Extension

The **C3 Architecture Navigator** extension adds in-editor navigation for C3 IDs (`c3-*`, `ref-*`, `adr-*`): CodeLens links, hover previews, Ctrl+Click to definition, and file path navigation in `code-map.yaml`.

**Install from release:**

```bash
# Download the latest .vsix from GitHub Releases
curl -fsSL -o c3-nav.vsix https://github.com/Lagz0ne/c3-skill/releases/latest/download/c3-nav.vsix
code --install-extension c3-nav.vsix --force
```

The extension activates automatically when a workspace contains `.c3/code-map.yaml`. See [`vscode-c3-nav/README.md`](vscode-c3-nav/README.md) for build-from-source instructions.

## Development

```bash
cd cli && go test ./...       # Run tests
bash scripts/build.sh         # Cross-compile for 4 targets
```

Binaries are cross-compiled in CI and force-added to `main`. The `dev` branch has source only.

## License

MIT
