# c3x: Architecture-Aware CLI for C3 Projects

c3x is a CLI tool for managing C3 architecture documentation. It provides commands for listing entities, checking integrity, scaffolding new docs, and more.

## Installation

### Per-project (recommended)

```bash
npm install --save-dev c3-kit
```

Then use via `npx`:

```bash
npx c3-kit list --json
npx c3-kit check
```

### Global

```bash
npm install -g c3-kit
```

Then use directly:

```bash
c3x list --json
c3x check
```

### Zero-install

```bash
npx -y c3-kit list --json
```

Downloads from npm on first use (~2s), then cached. The `-y` flag suppresses the install prompt.

## Commands

| Command | Description |
|---------|-------------|
| `list` | Topology view with relationships (`--json` for machine-readable) |
| `check` | Doc integrity: broken links, orphans, duplicates |
| `init` | Scaffold `.c3/` skeleton |
| `add <type> <slug>` | Create entity with auto-numbering + wiring |
| `sync` | Generate guard skills from component docs |

Run `npx c3-kit <command> --help` for details and examples.

## Claude Code Plugin

c3x also ships as a Claude Code plugin with architecture-aware skills:

```bash
claude plugin install c3-skill
```

| Skill | Purpose |
|-------|---------|
| `c3-onboard` | Create C3 docs from scratch via Socratic discovery |
| `c3-query` | Navigate architecture docs and answer questions |
| `c3-change` | Coordinated change workflow via Agent Teams (ADR-first) |
| `c3-ref` | Manage cross-cutting patterns and conventions |
| `c3-audit` | Audit docs for consistency, drift, and completeness |
| `c3-sweep` | Impact assessment before changes |

## The `.c3/` Directory

```
.c3/
├── README.md                    # System context
├── config.yaml                  # Optional config
├── c3-1-<container>/
│   ├── README.md                # Container overview
│   └── c3-101-<component>.md    # One file per component
├── refs/
│   └── ref-<pattern>.md         # Cross-cutting patterns
└── adr/
    └── adr-YYYYMMDD-<slug>.md   # Architecture decisions
```

## Building

```bash
bun install
bun run build:cli    # CLI → dist/cli.cjs
bun run build        # Plugin → dist/claude-code/
bun run check-refs   # Verify bundled references match source
bun run fix-refs     # Sync shared references into skills
```

## License

MIT
