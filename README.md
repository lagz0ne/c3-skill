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
- **The cache is disposable** — `c3.db` accelerates queries, search, and writes, but `c3x repair` can rebuild it from canonical text at any time
- **Every element is trackable** — headings, paragraphs, table rows, list items each have a unique ID and SHA256 hash. Entity-level merkle for O(1) change detection. Full version history with pruning

## What You Get

### Supported operations, one entry point

| Say this | C3 does this |
|----------|-------------|
| `/c3` adopt this project | **onboard** — discovers your architecture through conversation, scaffolds `.c3/` |
| `/c3` where is auth? | **query** — full-text search, graph traversal, traces relationships |
| `/c3` add rate limiting | **change** — ADR-first: impact analysis → decision record → execute → validate |
| `/c3` create a ref for error handling | **ref** — cross-cutting pattern with Choice/Why/How sections and cite wiring |
| `/c3` add a rule for structured logging | **rule** — enforceable standard with golden example and anti-patterns |
| `/c3` audit the docs | **audit** — 10-phase validation: structural → semantic → drift → compliance |
| `/c3` repair the cache after a branch switch | **migrate** — repair/rebuild cache and handle C3 version upgrades |
| `/c3` what breaks if I change payments? | **sweep** — transitive impact across the entity graph |

### The `c3x` CLI

A Go binary bundled inside the plugin. No separate install — the skill carries its own binary for every platform (linux/darwin x amd64/arm64).

> **For agents:** The `/c3` skill handles invocation automatically via `bash <skill-dir>/bin/c3x.sh`. Never run bare `c3x` — always go through `/c3`. The examples below use `c3x` as shorthand for readability.

**Read/Write cycle:**
```bash
c3x read c3-101                          # entity content rendered from node tree
c3x read c3-101 --section Goal           # just one section
c3x read c3-101 --json                   # structured JSON

echo "New goal." | c3x write c3-101 --section "Goal"   # section update
cat updated.md | c3x write c3-101                       # full replace (validates + versions)
```

**Search and navigate:**
```bash
c3x query "authentication"               # full-text search with ranking
c3x lookup src/auth/login.ts             # file → component + refs + rules
c3x impact ref-jwt                       # what breaks if this changes?
c3x graph c3-1 --format mermaid          # visual subgraph
```

**Manage entities:**
```bash
cat auth-component.md | c3x add component auth --container c3-1
# → {"id":"c3-101","type":"component","sections":["Goal","Parent Fit","Purpose",...]}

c3x wire c3-101 ref-jwt ref-error-handling   # batch wire multiple targets
c3x set c3-101 --section "Goal" "Handle JWT authentication"
c3x set c3-101 codemap "src/auth/**,src/auth.go"   # set code-map patterns
c3x set c3-101 codemap "src/new/**" --append        # add a pattern

# Batch update (fields + sections + codemap in one call):
echo '{"fields":{"goal":"X"},"codemap":["src/**"]}' | c3x set ref-jwt --stdin

c3x delete ref-obsolete --dry-run
```

**Track changes:**
```bash
c3x diff                                 # what changed since last commit
c3x diff --mark abc123                   # stamp changelog with commit hash
c3x check                               # validate everything
c3x coverage                            # code-map completeness stats
```

**Verify and recover:**
```bash
c3x verify                               # verify sealed canonical .c3/ truth
c3x repair                               # rebuild local cache + reseal after branch/merge issues
c3x git install                          # install pre-commit guardrails and .c3/.gitignore policy
```

**Content database:**
```bash
c3x nodes c3-101                         # tree of all nodes with IDs + hashes
c3x nodes c3-101 --json                  # JSON output
c3x hash c3-101                          # root merkle hash
c3x hash c3-101 --recompute             # verify hash integrity
c3x versions c3-101                      # version history
c3x version c3-101 3                     # content at version 3
c3x prune c3-101 --keep 10             # prune old versions
```

**Full command list:** `c3x --help`

### Schema enforcement

Every entity type has required sections. The CLI enforces them on write:

| Entity | Required sections |
|--------|------------------|
| Component | Goal, Parent Fit, Purpose, Foundational Flow, Business Flow, Governance, Contract, Change Safety, Derived Materials |
| Container | Goal, Components, Responsibilities |
| Ref | Goal, Choice, Why |
| Rule | Goal, Rule, Golden Example |
| ADR, Recipe | Goal |

`c3x write` (full body) validates required sections before accepting. Component section updates validate the full resulting document, so component docs stay all-or-nothing. `c3x check` validates everything post-hoc.

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
- after branch switches, selective merges, or conflict resolution, run `c3x repair`

Every canonical doc carries a `c3-seal` hash. `c3x verify` checks those seals and confirms the current `.c3/` tree matches canonical output.

Entity types: `container`, `component`, `ref`, `rule`, `adr`, `recipe`.

Code-map entries link entities to source files via glob patterns. `c3x lookup` resolves any file to its architecture context. `c3x coverage` shows what's mapped and what isn't.

### Marketplace

Share and adopt curated rule collections:

```bash
c3x marketplace add https://github.com/org/go-patterns
c3x marketplace list --tag errors
c3x marketplace show rule-error-wrapping
```

## Migrating

### Upgrading to v9.1.5

v9 is a breaking workflow release. The shared truth moves from “database-first” to “canonical text first”.

Do this once after upgrading:

```bash
# 1. Install the new guardrails
c3x git install

# 2. If the repo still tracks the cache, stop tracking it
git rm --cached .c3/c3.db

# 3. Rebuild and reseal the local tree
c3x repair

# 4. Verify before commit
c3x verify
```

What changes for daily work:
- review `.c3/` diffs in Git
- let `c3.db` stay local and disposable
- use `c3x verify` before commit / in CI
- use `c3x repair` after branch switches, selective merges, or manual conflict resolution
- if broken canonical state blocks read-only commands, repair with the narrow mutating command (`write --section`, `set --section`, or `add adr`) and then run `c3x check --include-adr && c3x verify`

**From v7 (body-based entities):**
```bash
c3x migrate                  # populates node tree for all entities
c3x migrate --dry-run        # preview without changes
c3x migrate --dry-run --json # machine-readable blockers; agent mode returns TOON
```

Expected migration failure flow:
- `BLOCKED: N component(s)` means strict component docs failed preflight and no migration writes occurred. Use `c3x migrate repair-plan`, repair listed sections with `c3x migrate repair <id> --section <name>`, run `c3x cache clear`, run `c3x import --force`, then resume with `c3x migrate --continue`.
- `BLOCKED: migration write failed at <id>` means C3 stopped before canonical export so submitted markdown is not rewritten from a partial cache. Fix the write/cache issue, run `c3x cache clear`, rebuild from canonical text, then resume migration.
- Do not use speculative command chains. Follow the printed fix loop and finish with `c3x check --include-adr && c3x verify`.

**From pre-v7 (file-based `.c3/`):**
```bash
c3x migrate-legacy           # markdown files → SQLite
c3x migrate                  # then populate node tree
```

## Self-contained distribution

The plugin ships with pre-built binaries — no Go toolchain, no npm, no PATH configuration:

```
skills/c3/bin/
├── VERSION                    # "9.1.5"
├── c3x.sh                    # detects OS/ARCH, runs the right binary
├── c3x-9.1.5-linux-amd64
├── c3x-9.1.5-linux-arm64
├── c3x-9.1.5-darwin-amd64
└── c3x-9.1.5-darwin-arm64
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
cd cli && go test ./...       # 14 packages
bash scripts/build.sh         # cross-compile for 4 targets
```

## License

MIT
