# C3: Architecture That Agents Can Read

C3 turns your codebase into something an LLM can navigate. A single `.c3/` directory holds architecture docs in a SQLite database — every component, every cross-cutting pattern, every decision, queryable and enforceable.

One Claude Code plugin. One `/c3` command. The agent figures out the rest.

```bash
claude plugin install lagz0ne/c3-skill
```

Then: `/c3 onboard this project`

## Why

Architecture docs rot because nobody enforces them. C3 fixes this by making the docs machine-writable and machine-verifiable:

- **LLMs read them before touching code** — `c3x lookup src/auth/login.ts` tells the agent which component owns the file, which refs govern it, what rules apply
- **Writes are validated** — every content update passes through schema enforcement. Missing a required section? Rejected with a hint
- **One database, not scattered files** — entities, content node trees, relationships, code-map, version history, and changelog in a single `c3.db`. Full-text search. Graph traversal. Impact analysis
- **Every element is trackable** — headings, paragraphs, table rows, list items each have a unique ID and SHA256 hash. Entity-level merkle for O(1) change detection. Full version history with pruning

## What You Get

### Seven operations, one entry point

| Say this | C3 does this |
|----------|-------------|
| `/c3` adopt this project | **onboard** — discovers your architecture through conversation, scaffolds `.c3/` |
| `/c3` where is auth? | **query** — full-text search, graph traversal, traces relationships |
| `/c3` add rate limiting | **change** — ADR-first: impact analysis → decision record → execute → validate |
| `/c3` create a ref for error handling | **ref** — cross-cutting pattern with Choice/Why/How sections and cite wiring |
| `/c3` add a rule for structured logging | **rule** — enforceable standard with golden example and anti-patterns |
| `/c3` audit the docs | **audit** — 10-phase validation: structural → semantic → drift → compliance |
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
c3x add component auth --container c3-1 --goal "JWT auth" --json
# → {"id":"c3-101","type":"component","sections":["Goal","Dependencies",...]}

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
| Component | Goal, Dependencies |
| Container | Goal, Components, Responsibilities |
| Ref | Goal, Choice, Why |
| Rule | Goal, Rule, Golden Example |
| ADR, Recipe | Goal |

`c3x write` (full body) validates required sections before accepting. Section-level updates (`write --section`, `set --section`) skip validation to allow incremental filling. `c3x check` validates everything post-hoc.

### The database

```
.c3/
├── c3.db           # everything lives here
└── config.yaml
```

`c3.db` holds entities, a **content node tree** (every heading, paragraph, list item, table row), relationships, code-map globs, version history, and a mutation changelog. No scattered markdown files — but `c3x export` dumps to files any time you need them.

Every content element has an **ID** and a **SHA256 hash** for change tracking. Entity-level merkle hashes detect any content change with a single comparison. Full version history with configurable pruning.

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

**From v7 (body-based entities):**
```bash
c3x migrate                  # populates node tree for all entities
c3x migrate --dry-run        # preview without changes
```

**From pre-v7 (file-based `.c3/`):**
```bash
c3x migrate-legacy           # markdown files → SQLite
c3x migrate                  # then populate node tree
```

## Self-contained distribution

The plugin ships with pre-built binaries — no Go toolchain, no npm, no PATH configuration:

```
skills/c3/bin/
├── VERSION                    # "8.0.0"
├── c3x.sh                    # detects OS/ARCH, runs the right binary
├── c3x-8.0.0-linux-amd64
├── c3x-8.0.0-linux-arm64
├── c3x-8.0.0-darwin-amd64
└── c3x-8.0.0-darwin-arm64
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
