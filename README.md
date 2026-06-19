# C3: Architecture That Agents Can Read

C3 turns your codebase's architecture into something an LLM can navigate, query, and safely change. A sealed `.c3/` tree of markdown is the shared architectural truth — reviewable in Git, validated by machine — while `c3.db` stays a local cache the CLI rebuilds at any time.

Use it through the Claude Code plugin or the `npx` CLI. Both run the same `c3x` engine; only the install shape differs.

## Install / Run

**Claude Code plugin (fat, self-contained):**

```bash
claude plugin install lagz0ne/c3-skill
```

Then: `/c3 onboard this project`

The plugin carries the `c3x` binary and the embedded semantic model — meaning-based search works offline, no downloads, no toolchain.

**`npx` CLI (thin, fetched on demand):**

```bash
npx @c3x/cli check
npx @c3x/cli search "how do users sign in and get permissions"
```

The npm package downloads the matching `c3x` binary and semantic model from the GitHub Release into a versioned local cache on first use.

## Why

Architecture docs rot because nobody enforces them. C3 makes them machine-writable and machine-verifiable:

- **Your entity model, not ours** — every entity type is defined by a *canvas*, user-owned markdown your team edits; validation enforces *your* definition
- **LLMs ask by meaning** — `c3x search "how do users sign in"` surfaces authentication/RBAC/JWT docs even when the wording differs
- **LLMs read before touching code** — `c3x lookup src/auth/login.ts` returns the owning component, governing refs, and applicable rules
- **Writes are validated** — every content update passes canvas enforcement; a missing required section is rejected with a hint
- **Canonical text is reviewable** — Git diffs and merges happen on sealed `.c3/*.md` files, never on an opaque cache
- **The cache is disposable** — `c3x check` rebuilds `c3.db` from canonical text whenever needed
- **Every element is trackable** — headings, paragraphs, table rows, and list items carry stable IDs and hashes; entity-level merkle gives O(1) change detection

## The canvas model

The core idea of C3: **entity definitions are data, owned by your project.**

Every entity type — `container`, `component`, `ref`, `rule`, `adr`, `recipe`, and document types like `prd` or `user-story` — is defined by a **canvas**: markdown declaring its sections, table columns, and reject rules. c3x ships built-in canvases as seeds; on onboard they materialize into `.c3/canvases/<type>.md` and your team owns them from there.

```bash
c3x canvas list             # every entity definition + domain + source
c3x canvas read component   # the canonical definition (yours to edit)
c3x schema component        # render its sections, columns, REJECT IF rules
```

Edit a canvas to shape docs around *your* architecture vocabulary — add a section, tighten a reject rule, define a new entity type — and `c3x write` / `c3x check` enforce *your* definition, not a baked-in one. Definitions travel with the repo and are reviewable in Git like everything else.

The built-in component canvas, for example, requires Goal, Parent Fit, Purpose, Foundational Flow, Business Flow, Governance, Contract, Change Safety, and Derived Materials — but that is a seed default, not a contract baked into the binary.

## What You Get

### Supported operations, one entry point

| Say this | C3 does this |
|----------|-------------|
| `/c3` adopt this project | **onboard** — discovers your architecture through conversation, scaffolds `.c3/` |
| `/c3` where is auth? | **query** — meaning-based discovery via `search`, then `lookup`, `read`, `graph` |
| `/c3` add rate limiting | **change** — ADR-first: impact analysis → decision record → execute → validate |
| `/c3` create a ref for error handling | **ref** — cross-cutting pattern with Choice/Why/How sections and cite wiring |
| `/c3` add a rule for structured logging | **rule** — enforceable standard with golden example and anti-patterns |
| `/c3` edit the canvas for ADRs | **canvas** — inspect or reshape the definitions that govern your docs |
| `/c3` audit the docs | **audit** — structural → semantic → drift → compliance validation |
| `/c3` what breaks if I change payments? | **sweep** — transitive impact across the entity graph, with a verification table |

Query answers follow the skill's Answer Depth Contract: claims bound to reads actually run, causal chains over entity lists, failure boundaries stated, direct vs transitive dependents separated.

### The `c3x` CLI

> **For agents:** the `/c3` skill invokes the CLI automatically via `bash <skill-dir>/bin/c3x.sh`. Never run bare `c3x` — always go through `/c3`. The examples below use `c3x` as shorthand.

**Read:**
```bash
c3x read c3-101                          # entity content
c3x read c3-101 --section Goal           # just one section
c3x read c3-101 --json                   # structured JSON
```

**Write — two shapes, one rule: complex content goes through a file:**
```bash
# Plain-text section edit
echo "Handle JWT authentication" | c3x write c3-101 --section Goal

# Rich content (mermaid, code fences, tables, mixed quotes) — use --file
c3x write c3-101 --section "Foundational Flow" --file flow.md
c3x write c3-101 --file full-body.md

# New entity
c3x add component auth --container c3-1 --file auth-component.md
```

**Frontmatter fields (no body touched):**
```bash
c3x set c3-101 goal "Handle JWT auth"
c3x set c3-101 codemap "src/auth/**,src/auth.go"
c3x set c3-101 codemap "src/new/**" --append
```

**Navigate the architecture:**
```bash
c3x search "what handles tenant permissions?" # meaning → candidate entities
c3x list                                 # topology: goals, file coverage, ref usage
c3x lookup src/auth/login.ts             # file → component + refs + rules
c3x graph c3-1 --format mermaid          # forward subgraph as mermaid
c3x graph ref-jwt --direction reverse    # what breaks if this changes?
c3x schema adr                           # required sections + pre-draft workorder
```

### Find by meaning

Use `c3x search "<question in plain English>"` when you know the concept but not the entity name or file path.

```bash
c3x search "how do users sign in and get permissions"
```

Hybrid search fuses three signals:

- **semantic** — local ONNX all-MiniLM embeddings rank docs by meaning, so "sign in" can match "authentication" and "permissions" can match "RBAC"
- **keyword/BM25** — exact terms still matter when the wording lines up
- **graph** — related components, refs, rules, and code-map paths add architectural context

`match_sources` in the output shows which signals ranked each candidate. Compare with keyword-only mode when you need to see what semantic search added:

```bash
c3x search "how do users sign in and get permissions" --no-semantic
```

**Citations and removal:** a component cites refs/rules in its body — the Governance
table column *is* the edge, re-derived on `write`. Frozen facts are re-edged through a
change-unit (a `frontmatter` patch). A `retire` is refused if it would orphan a child
or leave a citation dangling, so the graph never strands.
```bash
c3x delete ref-obsolete --dry-run            # preview removing a fact
```

**Validate:**
```bash
c3x check                                # canonical seal + schema + refs + coverage
c3x check --only-touched                 # scope to branch-touched entities
c3x check --only c3-101                  # scope to one entity
c3x check --include-adr                  # include ADR validation (skip by default)
c3x check --rule rule-xyz                # scope to citers of a rule
c3x check --fix                          # auto-fix title-matched references
```

**Full command list:** `c3x --help`

### Canvas enforcement

`c3x write` (full body) validates required sections before accepting. `c3x write --section` on a component validates the full resulting document, so component docs stay all-or-nothing. `c3x check` validates everything post-hoc — always against the project's own canvas definitions.

ADR schema output carries a pre-draft workorder: create a volatile Discovery Brief from the task goal and targeted `c3x` reads before writing the ADR body, so agents read the governing material without flooding add-time errors.

### Canonical `.c3/` tree

```
.c3/
├── README.md       # canonical context doc
├── canvases/       # entity definitions (user-owned shape of each type)
├── adr/            # canonical ADR markdown
├── refs/           # canonical refs
├── rules/          # canonical rules
├── recipes/        # cross-cutting trace shortcuts
├── .gitignore      # ignores local cache and backups inside the C3 tree
└── c3.db           # local cache only (rebuildable)
```

The sealed markdown tree is the shared truth. `c3.db` holds entities, a content node tree, relationships, code-map globs, version history, and a mutation changelog — as local cache.

User rule:
- review and merge `.c3/` text
- never merge `c3.db`
- after branch switches, selective merges, or conflict resolution, run `c3x check` (it auto-rebuilds cache, verifies seals, then validates)

Every canonical doc carries a `c3-seal` hash. `c3x check` verifies those seals and confirms the current `.c3/` tree matches canonical output.

Code-map entries link entities to source files via glob patterns. `c3x lookup` resolves any file to its architecture context; `c3x check` reports coverage as part of validation.

## Daily workflow

- review `.c3/` diffs in Git
- let `c3.db` stay local and disposable
- run `c3x check` before commit (in CI too)
- install once: `c3x git install` (pre-commit guardrails + `.c3/.gitignore` policy)

## Distribution

Two shapes, one engine:

**Plugin (fat).** The Claude Code plugin bundles pre-built binaries plus the embedded semantic model. No Go toolchain, npm, PATH setup, or model download:

```
skills/c3/bin/
├── VERSION                    # current plugin version (read by c3x.sh)
├── c3x.sh                     # detects OS/ARCH, runs the right binary
├── c3x-<version>-linux-amd64
├── c3x-<version>-linux-arm64
└── c3x-<version>-darwin-arm64
```

Each plugin version carries its own binary, so different projects can run different versions without conflict.

**npm (thin).** `@c3x/cli` is a small manager that resolves OS/ARCH, downloads the matching GitHub Release binary and semantic model into a versioned local cache, then executes the cached binary. The npm package version matches the c3x release it pins.

## Development

```bash
cd cli && go test ./...       # run the Go test suite
bash scripts/build.sh         # cross-compile release targets
```

The skill itself is developed eval-first: see `research/eval/skill-eval/` for the graded eval (deterministic scorer + LLM judge) that drives guidance changes.

## License

MIT
