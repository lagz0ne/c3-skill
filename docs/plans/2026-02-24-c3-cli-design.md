# C3 CLI: File-Based Architecture Toolkit

**Date**: 2026-02-24
**Status**: Draft
**Supersedes**: MCP server approach

---

## The Shift

C3 had an MCP server wrapping SQLite + vector embeddings. We're replacing it with a **pure file-based CLI** that treats `.c3/` as the database and Claude as the semantic engine.

| Before (MCP) | After (CLI) |
|---|---|
| MCP server over stdio | CLI subcommands |
| SQLite + vector index | Direct file access |
| Embedding-based search | Claude's reasoning |
| bun:sqlite dependency | Runtime-agnostic |
| Binary distribution | npm package |
| 7 MCP tools | 15 CLI commands |
| LLM does file I/O | CLI does file I/O, LLM does reasoning |
| c3-apply writes CLAUDE.md | `c3 sync` generates guard skills |

### Why

Three reasons:

1. **The vector index was a middleman.** Claude reasons over text better than TF-IDF or OpenAI embeddings. The CLI gives Claude fast, structured access to `.c3/` docs. Claude provides the understanding.

2. **Skills burn 60-80% of LLM tokens on mechanical work.** File walking, frontmatter parsing, relationship tracing, template expansion, auto-numbering — all things a CLI does in milliseconds for free.

3. **Guard skills replace passive CLAUDE.md.** Instead of a monolithic context file, `c3 sync` generates per-component skills that trigger when Claude touches the corresponding code paths — protection exactly where and when it's needed.

### Token savings by skill

| Skill | Mechanical % | What CLI absorbs | LLM now only does |
|---|---|---|---|
| c3-onboard (82%) | `init` + `add` | 30+ file writes, templates, numbering | Socratic discovery, content decisions |
| c3-audit (75%) | `check` | 10-phase file scanning, validation | Prioritizing findings, recommending fixes |
| c3-ref (67%) | `ref add/usage/check/link` | Template writes, citation scanning, cross-links | Pattern extraction from user intent |
| c3-query (65%) | `list` + `read` | Layer navigation, file reads | Synthesizing answers, reasoning |
| c3-navigator (52%) | `list` + `trace` | Doc-by-doc reading, graph building | Question interpretation, diagrams |
| c3-lead (42%) | `add` + `impact` + `template adr` | ADR template, entity discovery | Decision-making, quality gating |

---

## Commands

```
c3 — Architecture-aware toolkit for C3 projects

Data layer (read):
  list                        Topology view with relationships
  read <path>                 Read a C3 doc
  trace <name>                Follow relationship chains
  check                       Doc integrity + code-doc coverage
  impact <name>               Blast radius (transitive trace)

Action layer (write):
  init                        Scaffold .c3/ skeleton
  add <type> <slug>           Create entity + auto-number + wire relationships
  evolve <path>               Update a doc
  template <type>             Emit template to stdout
  sync                        Generate guard skills from .c3/ docs

Ref subcommands:
  ref add <slug>              Create ref doc
  ref usage <id>              Who cites this ref
  ref check [id]              Compliance check
  ref link <ref-id> <comp-id> Wire ref <-> component

Global options:
  -h, --help                  Show help (per-command or global)
  -v, --version               Print version
  --json                      Machine-readable JSON output
  --c3-dir <path>             Override .c3/ directory detection
```

### Naming: CLI commands vs skills

CLI commands are **mechanical operations**. Skills are **LLM-powered reasoning**. Names must not collide.

| CLI command | Skill | Relationship |
|---|---|---|
| `c3 check` | `c3-audit` | Skill invokes CLI for data, reasons over results |
| `c3 ref check` | `c3-ref` | Skill invokes CLI for citations, reasons about patterns |
| `c3 list`, `c3 read`, `c3 trace` | `c3-query` | Skill invokes CLI for topology, synthesizes answers |
| `c3 impact`, `c3 add` | `c3-change` | Skill invokes CLI for blast radius, orchestrates workers |
| `c3 sync` | — | Pure CLI, no skill equivalent |

---

## Command Details

### Data Layer

**`c3 list`** — Topology, not a file listing. Reads frontmatter from all `.c3/` docs, builds parent-child and cross-reference graph, renders as a tree with relationships.

```
$ c3 list

c3-1-api (container)
├── c3-101-auth-provider (foundation)   → ref: ref-auth-patterns
├── c3-102-logger (foundation)          → ref: ref-logging
├── c3-110-auth-middleware (feature)     → affects: c3-2/queue-processor
└── c3-111-token-validator (feature)

c3-2-worker (container)
├── c3-201-queue-processor (foundation)
└── c3-210-data-sync (feature)          → ref: ref-event-sourcing

Cross-cutting:
  ref-event-sourcing         → used by: c3-210
  ref-logging                → used by: c3-102
  ref-auth-patterns          → used by: c3-101

ADRs:
  adr-20260129-event-sourcing  → status: accepted
  adr-20260215-rate-limiting   → status: proposed
```

Options: `--flat` for simple file list.

**`c3 read <path>`** — Read a specific doc by path relative to `.c3/`. Returns full markdown with frontmatter. Accepts ID or path:

```bash
c3 read c3-101                           # by ID
c3 read c3-1-api/c3-101-auth-provider    # by path
c3 read ref-logging                      # ref by ID
```

**`c3 trace <name>`** — Follow relationship chains forward from a component/container/ref.

```bash
c3 trace c3-101                   # what does auth-provider affect?
c3 trace --reverse c3-210         # what points TO data-sync?
c3 trace ref-event-sourcing       # what cites this ref?
```

Options: `--reverse`, `--depth <n>` (default: 1), `--json`.

**`c3 impact <name>`** — Transitive closure of `trace`. Full blast radius: everything affected if this entity changes. Follows chains: A → B → C.

```bash
c3 impact c3-101                  # blast radius of auth-provider
c3 impact c3-1                    # blast radius of entire api container
c3 impact ref-logging             # everything affected if logging pattern changes
```

Options: `--depth <n>` (default: 2), `--json`.

**`c3 check`** — Doc integrity + code-to-doc coverage. Merges structural validation and coverage scanning into one command.

```
$ c3 check

Doc integrity:
  ✓ 12 entities, 0 broken links
  ✗ c3-205: missing parent container
  ✗ ref-logging: 1 uncited component
  ✗ c3-110: ID/filename mismatch

Code coverage:
  ✗ src/api/cache/ has no C3 doc
  ✗ src/worker/retry.ts has no C3 doc
  ✓ 84% code paths documented
```

Doc integrity checks:
- Broken relationships (references to non-existent entities)
- Missing required fields (components without parent container)
- Orphan docs (no incoming relationships)
- Duplicate IDs
- Empty content
- ID/filename mismatches
- Foundation/feature numbering violations (01-09 vs 10+)

Code coverage checks:
- Source files without corresponding C3 docs
- Coverage percentage

Options: `--docs` (integrity only), `--code` (coverage only), `--json`.

**`c3 template <type>`** — Emit template to stdout. Types: `container`, `component`, `ref`, `adr`.

```bash
c3 template component             # print component template
c3 template --list                # show available templates
```

### Action Layer

**`c3 init`** — Scaffold `.c3/` skeleton. Does NOT fill content — that's the LLM's job.

```
$ c3 init
Created .c3/
  ├── config.yaml
  ├── README.md          (context template, id: c3-0)
  ├── containers/        (empty)
  ├── refs/              (empty)
  └── adr/
      └── adr-00000000-c3-adoption.md
```

**`c3 add <type> <slug>`** — Create entity with auto-numbering, correct directory placement, and relationship wiring.

```bash
# Add container (auto-assigns c3-3)
$ c3 add container payments
Created: .c3/c3-3-payments/README.md (id: c3-3)

# Add foundation component (auto-assigns c3-301)
$ c3 add component auth-provider --container c3-3
Created: .c3/c3-3-payments/c3-301-auth-provider.md (id: c3-301)
Updated: .c3/c3-3-payments/README.md (component list)

# Add feature component (auto-assigns c3-310)
$ c3 add component checkout-flow --container c3-3 --feature
Created: .c3/c3-3-payments/c3-310-checkout-flow.md (id: c3-310)
Updated: .c3/c3-3-payments/README.md (component list)

# Add ref
$ c3 add ref rate-limiting
Created: .c3/refs/ref-rate-limiting.md (id: ref-rate-limiting)

# Add ADR
$ c3 add adr oauth-support
Created: .c3/adr/adr-20260224-oauth-support.md (id: adr-20260224-oauth-support)
```

Auto-numbering logic:
- Containers: scan existing `c3-{N}-*` dirs, pick N+1
- Foundation components (default): scan existing `c3-{N}0{NN}`, pick next in 01-09
- Feature components (`--feature`): scan existing `c3-{N}{NN}`, pick next in 10+
- ADRs: use today's date `YYYYMMDD`

**`c3 evolve <path>`** — Update a C3 doc. Accepts content from stdin.

```bash
echo "updated content" | c3 evolve c3-101
```

**`c3 sync`** — Generate guard skills from `.c3/` component docs into `.claude/skills/`. See [Guard Skills](#guard-skills) section below.

```
$ c3 sync
Generated 9 guard skills in .claude/skills/
  c3-guard-api-auth.md           (src/api/auth/**)
  c3-guard-api-middleware.md     (src/api/middleware/**)
  c3-guard-worker-queue.md       (src/worker/queue/**)
  ...
```

### Ref Subcommands

**`c3 ref add <slug>`** — Create ref doc (alias for `c3 add ref <slug>`).

**`c3 ref usage <id>`** — Find all components that cite this ref.

```
$ c3 ref usage ref-logging
c3-102-logger (container: c3-1-api)
c3-205-event-bus (container: c3-2-worker)
```

**`c3 ref check [id]`** — Compliance check. Verifies that components claiming to follow a ref actually reference it in frontmatter.

```
$ c3 ref check ref-logging
✓ c3-102-logger: compliant
✗ c3-205-event-bus: missing ref in frontmatter
```

Without `id`, checks all refs.

**`c3 ref link <ref-id> <comp-id>`** — Wire a ref to a component. Updates both the component's frontmatter and the ref's usage list.

```
$ c3 ref link ref-logging c3-301
Linked: ref-logging → c3-301-auth-provider
Updated: .c3/c3-3-payments/c3-301-auth-provider.md (refs field)
```

---

## Guard Skills

The CLI's most distinctive feature. `c3 sync` reads all component docs, extracts their `## Code References` (file paths), and generates **guard skills** — Claude Code skills that trigger when Claude touches those code paths.

### Why guard skills

| c3-apply (old) | c3 sync (new) |
|---|---|
| Writes one monolithic CLAUDE.md | Generates per-component skills |
| Passive context (Claude reads once) | Active triggering (fires when relevant) |
| All architecture context always loaded | Only the relevant component's rules load |
| No code-path awareness | Triggers by matching file paths |

### What gets generated

One guard skill per component that has `## Code References`. Each skill:

1. **Triggers on the component's code paths** — skill description includes file globs
2. **Contains the rules** — applicable refs inlined as constraints
3. **Shows blast radius** — what depends on this component (static info)
4. **Points to post-change verification** — invoke the c3-audit skill

### Example guard skill

```markdown
# .claude/skills/c3-guard-api-auth.md
---
name: c3-guard-api-auth
description: >
  Use when modifying files in src/api/auth/** or src/lib/auth.ts.
  Guards auth-provider component (c3-101) in the API container.
---

## Component: c3-101-auth-provider (foundation)
Container: c3-1-api

## Rules
- **ref-logging**: debug for flow tracing, info for auth events, error for failures
- **ref-error-handling**: RFC 7807, 401/403 status codes

## Blast radius
- c3-110-auth-middleware depends on AuthProvider interface
- c3-111-token-validator depends on token validation exports

## After changes
Invoke: c3-audit skill
```

### How it works in practice

1. User says: "fix the auth token refresh logic"
2. Claude is about to edit `src/api/auth/refresh.ts`
3. Claude's routing matches `c3-guard-api-auth` skill (description mentions `src/api/auth/**`)
4. Skill loads → Claude sees: follow ref-logging, follow ref-error-handling, auth-middleware depends on you
5. Claude writes code that follows the refs
6. After changes, Claude invokes c3-audit skill for structural verification

The guard skill ensures code is born **on-reference** — the rules are injected at the moment of coding, not discovered after the fact.

### Generated file structure

```
.claude/skills/          (generated by c3 sync, do not edit manually)
├── c3-guard-api-auth.md
├── c3-guard-api-middleware.md
├── c3-guard-api-routes.md
├── c3-guard-worker-queue.md
├── c3-guard-worker-sync.md
└── ...
```

Components without `## Code References` get no guard skill — they have no code paths to protect.

### When to run c3 sync

- After `c3 add` (new component with code references)
- After `c3 evolve` (updated code references or refs)
- After `c3 ref link` (new ref wired to a component)
- After onboarding (`c3-onboard` skill completes)

Manual invocation. No auto-sync, no watchers.

---

## Entity Naming Conventions

The CLI enforces these conventions for all `add` operations:

| Entity | ID Format | Directory Pattern | Example |
|---|---|---|---|
| Context | `c3-0` | `.c3/README.md` | Always `c3-0` |
| Container | `c3-{N}` | `.c3/c3-{N}-{slug}/README.md` | `c3-1`, `c3-2` |
| Foundation | `c3-{N}0{NN}` | `.c3/c3-{N}-{slug}/c3-{N}0{NN}-{slug}.md` | `c3-101`, `c3-302` |
| Feature | `c3-{N}{NN}` | `.c3/c3-{N}-{slug}/c3-{N}{NN}-{slug}.md` | `c3-110`, `c3-211` |
| Ref | `ref-{slug}` | `.c3/refs/ref-{slug}.md` | `ref-logging` |
| ADR | `adr-{YYYYMMDD}-{slug}` | `.c3/adr/adr-{date}-{slug}.md` | `adr-20260224-cli` |

Component numbering: 01-09 = foundation (infrastructure), 10+ = feature (business logic).

---

## Help System

Each subcommand has rich help with usage, description, options, and concrete examples. Modeled after agent-browser.

```
$ c3 --help

c3 — Architecture-aware toolkit for C3 projects

Usage: c3 <command> [options]

Data:
  list                   Topology view with relationships
  read <path>            Read a C3 doc (by ID or path)
  trace <name>           Follow relationship chains
  check                  Doc integrity + code coverage
  impact <name>          Blast radius of a change

Actions:
  init                   Scaffold .c3/ skeleton
  add <type> <slug>      Create entity with auto-numbering
  evolve <path>          Update a C3 doc
  template <type>        Emit doc template
  sync                   Generate guard skills

Refs:
  ref add <slug>         Create a reference pattern
  ref usage <id>         Find components citing a ref
  ref check [id]         Verify ref compliance
  ref link <ref> <comp>  Wire ref to component

Options:
  -h, --help             Show help (with command for details)
  -v, --version          Print version
  --json                 Machine-readable output
  --c3-dir <path>        Override .c3/ detection

Run 'c3 <command> --help' for details and examples.
```

---

## Architecture

### Source structure

```
src/
├── core/
│   ├── config.ts            # findC3Dir(), loadConfig()
│   ├── frontmatter.ts       # Parse YAML frontmatter, extract relationships
│   ├── walker.ts            # Walk .c3/ docs, classify types, build relationship graph
│   ├── numbering.ts         # Auto-numbering: next container N, next component NN
│   └── wiring.ts            # Relationship wiring: update cross-links between entities
├── cli/
│   ├── index.ts             # Entry point, arg parsing, scope setup, error handling
│   ├── commands/
│   │   ├── list.ts
│   │   ├── read.ts
│   │   ├── trace.ts
│   │   ├── check.ts         # doc integrity + code coverage (merged)
│   │   ├── impact.ts
│   │   ├── template.ts
│   │   ├── evolve.ts
│   │   ├── init.ts
│   │   ├── add.ts
│   │   ├── sync.ts          # guard skill generation
│   │   └── ref.ts           # ref add/usage/check/link
│   ├── help.ts              # Rich help text generation
│   └── output.ts            # Plain text / JSON rendering
└── index.ts                 # Library exports
```

### Removed from MCP era

- `embedding.ts` — no embeddings
- `vector-index.ts` — no SQLite index
- `mcp-server.ts` — no MCP transport
- `index.db` — no database file

### Context wiring with pumped-fn

Each command is a `flow`. Long-lived resources are `atom`s. Ambient context via `tag`s. Each command is the glue for its own scope and context — no hidden middleware.

```typescript
// Tags — ambient context per invocation
const c3DirTag   = tag<string>({ label: 'c3Dir' })
const optionsTag = tag<CliOptions>({ label: 'options' })

// Atoms — cached per scope, built once
const configAtom = atom({
  deps: { c3Dir: tags.required(c3DirTag) },
  factory: (ctx, { c3Dir }) => loadConfig(c3Dir)
})

const graphAtom = atom({
  deps: { c3Dir: tags.required(c3DirTag) },
  factory: async (ctx, { c3Dir }) => {
    const docs = await walkC3Docs(c3Dir)
    return buildRelationshipGraph(docs)
  }
})

// Read commands use graphAtom for pre-built topology
const listFlow = flow({
  deps: { graph: graphAtom, options: tags.required(optionsTag) },
  factory: async (ctx, { graph, options }) => {
    return options.flat ? flatList(graph) : topologyView(graph)
  }
})

// Write commands use graphAtom for numbering + wiring
const addFlow = flow({
  parse: (raw) => parseAddArgs(raw),
  deps: { graph: graphAtom, c3Dir: tags.required(c3DirTag) },
  factory: async (ctx, { graph, c3Dir }) => {
    const id = nextId(graph, ctx.input.type, ctx.input.container)
    const path = createDoc(c3Dir, id, ctx.input.slug, ctx.input.type)
    wireRelationships(c3Dir, graph, id, ctx.input.refs)
    return { id, path }
  }
})
```

### Error handling & scope cleanup

```typescript
async function main() {
  const options = parseGlobalOptions(process.argv.slice(2))

  if (options.help)    return showHelp(options.command)
  if (options.version) return showVersion()

  // init doesn't need an existing .c3/
  if (options.command === 'init') {
    return initFlow(options)
  }

  const c3Dir = options.c3Dir ?? findC3Dir(process.cwd())
  if (!c3Dir) bail("No .c3/ directory found. Run 'c3 init' first.")

  const scope = createScope({
    tags: [c3DirTag(c3Dir), optionsTag(options)]
  })

  try {
    const ctx = scope.createContext()
    const output = await ctx.exec({ flow: commandMap[options.command], input: options.args })
    render(output, options)
    await ctx.close({ ok: true })
  } catch (err) {
    if (err instanceof C3Error) {
      console.error(`error: ${err.message}`)
      if (err.hint) console.error(`hint: ${err.hint}`)
      process.exitCode = 1
    } else {
      console.error(`unexpected error: ${err}`)
      console.error(`Report: https://github.com/lagz0ne/c3-skill/issues`)
      process.exitCode = 2
    }
  } finally {
    await scope.dispose()
  }
}
```

Error hierarchy:
- `C3Error` — user-facing with optional `hint` field
- Everything else — unexpected crash with issue link
- `scope.dispose()` in `finally` — always releases resources

---

## Dependencies

| Package | Purpose |
|---|---|
| `@pumped-fn/lite` | Context wiring, scope lifecycle, flows |
| `zod` | Arg validation, frontmatter schemas |
| `yaml` | YAML frontmatter parsing (replaces Bun.YAML) |

**No runtime-specific APIs.** Source runs on Node, Bun, or Deno.

---

## Distribution

**npm package only.** No binary compilation needed.

```bash
# Install globally
npm i -g c3-cli
bun add -g c3-cli

# Run without install
npx c3-cli list
bunx c3-cli list
```

### Build

TypeScript → JavaScript bundle via tsdown or bun build (no --compile).

```json
{
  "name": "c3-cli",
  "bin": { "c3": "./dist/cli.js" },
  "type": "module",
  "files": ["dist/"]
}
```

### GitHub Actions

Single workflow: build + publish to npm on version bump.

```yaml
jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: oven-sh/setup-bun@v2
      - run: bun install
      - run: bun run build
      - run: npm publish
```

No platform matrix. No binary artifacts. No install.sh.

---

## How Skills Use the CLI

Before and after for each skill:

### c3-onboard (82% mechanical → ~10%)

```bash
# Before: 30+ Write calls, template expansion, manual numbering
# After:
c3 init                                          # scaffold skeleton
c3 add container api                             # auto-numbered c3-1
c3 add component auth --container c3-1           # auto-numbered c3-101
c3 add ref logging                               # creates ref doc
# LLM fills in content via c3 evolve
c3 sync                                          # generate guard skills
```

### c3-audit (75% → ~5%)

```bash
# Before: 10-phase file-by-file scanning, manual validation
# After:
c3 check --json                                  # doc integrity + code coverage
# LLM (c3-audit skill) interprets findings, recommends fixes
```

### c3-ref (67% → ~10%)

```bash
# Before: Glob for refs, read each citation, update cross-links manually
# After:
c3 ref add rate-limiting                         # create doc
c3 ref link ref-rate-limiting c3-110             # wire relationship
c3 ref usage ref-logging                         # find all citations
c3 ref check ref-logging                         # compliance report
# LLM extracts pattern from user intent, decides what to link
```

### c3-query (65% → ~5%)

```bash
# Before: Read README → Read container → Read component (3+ Read calls per question)
# After:
c3 list --json                                   # full topology in one call
c3 read c3-101                                   # specific doc by ID
c3 trace c3-101                                  # relationships
# LLM synthesizes the answer
```

### c3-lead (42% → ~10%)

```bash
# Before: Manual entity discovery, ADR template expansion, relationship scanning
# After:
c3 impact c3-1 --json                            # blast radius for planning
c3 template adr                                  # get ADR template
c3 add adr rate-limiting                         # create ADR doc
c3 add component rate-limiter --container c3-1   # provision new component
# LLM does decision-making, quality gating, worker orchestration
```

---

## Migration from MCP

### Removed
- `mcp.json` — no MCP config
- `src/mcp-server.ts` — no MCP server
- `src/core/embedding.ts` — no embeddings
- `src/core/vector-index.ts` — no vector index
- `install.sh` — no binary installer
- Binary build targets in release.yml
- `@modelcontextprotocol/sdk` dependency
- `bun:sqlite` usage

### Added
- `src/cli/` — CLI entry point and all commands
- `src/cli/commands/sync.ts` — guard skill generation
- `src/cli/commands/check.ts` — merged audit + verify
- `src/core/numbering.ts` — auto-numbering logic
- `src/core/wiring.ts` — relationship wiring
- `@pumped-fn/lite` dependency
- `yaml` dependency
- npm publish in CI

### Unchanged
- `src/core/config.ts` — still finds `.c3/` and loads config
- `.c3/` doc format — containers, components, refs, ADRs
- Entity naming conventions (c3-{N}, c3-{N}{NN}, ref-{slug}, adr-{date}-{slug})
- Skills and agents (updated to shell out to CLI)
- Plugin structure (skills/ and agents/ directories)

---

## Open Questions

1. **Package name**: `c3-cli`? `@c3/cli`? `c3-arch`?
2. **check --code scope**: How deep should code coverage scanning go? Tree-sitter AST or simple file presence?
3. **Guard skill placement**: `.claude/skills/` (project-level) or `.claude/skills/c3-guards/` (namespaced)?
