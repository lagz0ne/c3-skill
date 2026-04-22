# C3: Architecture That Agents Can Read

C3 turns your codebase into something an LLM can navigate. A sealed `.c3/` tree is the shared architectural truth for Git review and merges, while `c3.db` stays a local cache the CLI can rebuild at any time.

One Claude Code plugin. One `/c3` command. The agent figures out the rest.

```bash
claude plugin install lagz0ne/c3-skill
```

Then: `/c3 onboard this project`

## Why

Architecture docs rot because nobody enforces them. C3 fixes this by making the docs machine-writable and machine-verifiable:

- **LLMs read them before touching code** — `c3x lookup src/auth/login.ts` tells the agent which component owns the file, which refs govern it, what rules apply
- **Writes are validated** — every content update passes through schema enforcement. Missing a required section? Rejected with a hint
- **Canonical text is reviewable** — Git diffs and merges happen on sealed `.c3/*.md` files and `code-map.yaml`, not on an opaque cache
- **The cache is disposable** — `c3.db` accelerates queries and writes; `c3x check` rebuilds it from canonical text when needed
- **Every element is trackable** — headings, paragraphs, table rows, list items each have a unique ID and SHA256 hash; entity-level merkle for O(1) change detection

## What You Get

### Supported operations, one entry point

| Say this | C3 does this |
|----------|-------------|
| `/c3` adopt this project | **onboard** — discovers your architecture through conversation, scaffolds `.c3/` |
| `/c3` where is auth? | **query** — topology traversal via `list`, `lookup`, `read`, `graph` |
| `/c3` add rate limiting | **change** — ADR-first: impact analysis → decision record → execute → validate |
| `/c3` create a ref for error handling | **ref** — cross-cutting pattern with Choice/Why/How sections and cite wiring |
| `/c3` add a rule for structured logging | **rule** — enforceable standard with golden example and anti-patterns |
| `/c3` audit the docs | **audit** — structural → semantic → drift → compliance validation |
| `/c3` what breaks if I change payments? | **sweep** — transitive impact across the entity graph |

### The `c3x` CLI

A Go binary bundled inside the plugin. No separate install — the skill carries its own binary for every platform (linux/darwin x amd64/arm64).

> **For agents:** The `/c3` skill handles invocation automatically via `bash <skill-dir>/bin/c3x.sh`. Never run bare `c3x` — always go through `/c3`. The examples below use `c3x` as shorthand for readability.

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
c3x list                                 # topology: goals, file coverage, ref usage
c3x lookup src/auth/login.ts             # file → component + refs + rules
c3x graph c3-1 --format mermaid          # forward subgraph as mermaid
c3x graph ref-jwt --direction reverse    # what breaks if this changes?
c3x schema adr                           # required sections for an entity type
```

**Relationships and removal:**
```bash
c3x wire c3-101 ref-jwt ref-error-handling   # cite one or more refs/rules
c3x wire c3-101 ref-jwt --remove             # unlink
c3x delete ref-obsolete --dry-run
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

**Full command list:** `c3x --help` (11 user-facing commands)

### Schema enforcement

Every entity type has required sections. The CLI enforces them on write:

| Entity | Required sections |
|--------|------------------|
| Component | Goal, Parent Fit, Purpose, Foundational Flow, Business Flow, Governance, Contract, Change Safety, Derived Materials |
| Container | Goal, Components, Responsibilities |
| Ref | Goal, Choice, Why |
| Rule | Goal, Rule, Golden Example |
| ADR, Recipe | Goal |

`c3x write` (full body) validates required sections before accepting. `c3x write --section` on a component validates the full resulting document, so component docs stay all-or-nothing. `c3x check` validates everything post-hoc.

### Canonical `.c3/` tree

```
.c3/
├── README.md       # canonical context doc
├── adr/            # canonical ADR markdown
├── refs/           # canonical refs/rules/recipes/containers/components
├── .gitignore      # ignores local cache and backups inside the C3 tree
└── c3.db           # local cache only (rebuildable)
```

The sealed markdown tree is the shared truth. `c3.db` holds entities, a **content node tree** (every heading, paragraph, list item, table row), relationships, code-map globs, version history, and a mutation changelog as local cache.

User rule:
- review and merge `.c3/` text
- never merge `c3.db`
- after branch switches, selective merges, or conflict resolution, run `c3x check` (it auto-rebuilds cache, verifies seals, then validates)

Every canonical doc carries a `c3-seal` hash. `c3x check` verifies those seals and confirms the current `.c3/` tree matches canonical output.

Entity types: `container`, `component`, `ref`, `rule`, `adr`, `recipe`.

Code-map entries link entities to source files via glob patterns. `c3x lookup` resolves any file to its architecture context. `c3x check` reports coverage as part of validation.

### Marketplace

Share and adopt curated rule collections:

```bash
c3x marketplace add https://github.com/org/go-patterns
c3x marketplace list --tag errors
c3x marketplace show rule-error-wrapping
```

## Daily workflow

- review `.c3/` diffs in Git
- let `c3.db` stay local and disposable
- run `c3x check` before commit (in CI too)
- after branch switches, selective merges, or conflict resolution, `c3x check` rebuilds cache and verifies seals as part of validation
- install once: `c3x git install` (pre-commit guardrails + `.c3/.gitignore` policy)

## Self-contained distribution

The plugin ships with pre-built binaries — no Go toolchain, no npm, no PATH configuration:

```
skills/c3/bin/
├── VERSION                    # current plugin version (read by c3x.sh)
├── c3x.sh                     # detects OS/ARCH, runs the right binary
├── c3x-<version>-linux-amd64
├── c3x-<version>-linux-arm64
├── c3x-<version>-darwin-amd64
└── c3x-<version>-darwin-arm64
```

Each plugin version carries its own binary. Different projects can use different versions without conflict.

## VS Code Extension

**C3 Architecture Navigator** — Ctrl+Click on `c3-101` or `ref-jwt` in your code to jump to the doc. CodeLens, hover previews, glob path navigation.

```bash
curl -fsSL -o c3-nav.vsix https://github.com/Lagz0ne/c3-skill/releases/latest/download/c3-nav.vsix
code --install-extension c3-nav.vsix --force
```

## Development

```bash
cd cli && go test ./...       # 16 packages
bash scripts/build.sh         # cross-compile for 4 targets
```

## License

MIT
