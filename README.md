# C3: Architecture Documentation Toolkit

C3 (Context-Container-Component) is a Claude Code plugin with a Go CLI for managing architecture documentation.

## Claude Code Plugin

```bash
claude plugin install c3-skill
```

Single unified skill with 6 operations:

| Operation | Purpose |
|-----------|---------|
| `onboard` | Create C3 docs from scratch via Socratic discovery |
| `query` | Navigate architecture docs and answer questions |
| `change` | Coordinated change workflow (ADR-first) |
| `ref` | Manage cross-cutting patterns and conventions |
| `audit` | Audit docs for consistency, drift, and completeness |
| `sweep` | Impact assessment before changes |

Just say `/c3` + what you want.

## CLI

The Go CLI (`c3x`) is bundled with the plugin and also usable standalone.

| Command | Description |
|---------|-------------|
| `init` | Scaffold `.c3/` skeleton |
| `list` | Topology view with relationships (`--json`, `--flat`) |
| `check` | Doc integrity: broken links, orphans, duplicates (`--json`) |
| `add <type> <slug>` | Create entity with auto-numbering + wiring |

Types for `add`: `container`, `component`, `ref`, `adr`

## The `.c3/` Directory

```
.c3/
├── README.md                    # System context
├── config.yaml                  # Optional config
├── c3-N-<container>/
│   ├── README.md                # Container overview
│   └── c3-NNN-<component>.md   # One file per component
├── refs/
│   └── ref-<pattern>.md         # Cross-cutting patterns
└── adr/
    └── adr-YYYYMMDD-<slug>.md   # Architecture decisions
```

## Building

```bash
# Cross-compile Go CLI for all platforms
bash scripts/build.sh

# Run Go tests
cd cli && go test ./...
```

## License

MIT
