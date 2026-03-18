# c3x capabilities command

## Problem

c3x has 12 commands but onboard only advertises 2 (`lookup`, `coverage`). The CLI help text and onboard capabilities table are maintained separately, causing drift.

## Solution

Registry-driven command metadata. A single `[]CommandMeta` slice becomes the source of truth for both `c3x --help` and a new `c3x capabilities` subcommand.

## Data structure

```go
type CommandMeta struct {
    Name     string // "lookup"
    Args     string // "<file-or-glob>"
    OneLiner string // "Map file to component(s) + refs"
    Help     string // detailed help text
    Hidden   bool   // true = skip in capabilities output (e.g. init)
}

var commands = []CommandMeta{...}
```

## Consumers

1. **`c3x --help`** — generates Commands section from `Name + Args + OneLiner`
2. **`c3x capabilities`** — emits markdown table (skips `Hidden` entries)
3. **`c3x <cmd> --help`** — uses `Help` field (unchanged detailed text)

## capabilities output format

```markdown
| Command | What it does |
|---------|-------------|
| `c3x list` | Topology view with relationships |
| `c3x check` | Validate docs, schema, code refs, consistency |
| `c3x add <type> <slug>` | Create entity (auto-numbering + wiring) |
| `c3x set <id> <field> <value>` | Update frontmatter field |
| `c3x wire <src> [cite] <tgt>` | Link component to ref (--remove to unlink) |
| `c3x schema <type>` | Show known sections for entity type |
| `c3x codemap` | Scaffold code-map.yaml for all components + refs |
| `c3x lookup <file-path>` | Map file to component(s) + refs |
| `c3x coverage` | Code-map coverage stats |
| `c3x graph <entity-id>` | Subgraph from entity (LLM-friendly output) |
| `c3x delete <id>` | Remove entity + clean all references |
```

Hidden: `init` (handled by onboard, not user-facing).

## Onboard integration

Post-onboard reveal replaces the current inventory dump with two things:

### 1. Typical flow introduction

A short narrative showing how c3x is used in practice — not a command reference, but the designed workflow:

```
## Your C3 toolkit is ready

**Typical flow:**

1. Understand what exists: `c3x list` → topology, then `c3x lookup <file>` → which component owns it
2. Make changes: `c3x add` / `c3x set` / `c3x wire` to create and connect entities
3. Validate: `c3x check` catches broken links, schema gaps, orphans
4. Explore impact: `c3x graph <id>` shows what connects to what

For architecture questions, changes, audits → just say `/c3` + what you want.
```

### 2. Discovery pointer

Instead of listing every command, point to self-discovery:

```
Run `c3x capabilities` to see all available commands.
Run `c3x <command> --help` for detailed usage.
```

This way the reveal teaches the mental model (the flow) and gives the user a way to explore further, rather than dumping a table they'll forget.

## Changes

| File | Change |
|------|--------|
| `cli/cmd/help.go` | Replace `globalHelp` string + `commandHelp` map with `[]CommandMeta` registry. Add `ShowCapabilities(w)`. Generate `globalHelp` from registry. |
| `cli/cmd/args.go` | Add `capabilities` to command parsing |
| `cli/main.go` | Add `capabilities` case to switch |
| `skills/c3/references/onboard.md` | Replace hardcoded c3x rows with instruction to run `c3x capabilities` |

## Non-goals

- CLI does not emit `/c3` skill operation rows (skill owns those)
- No changes to detailed per-command help text (just moves into `Help` field)
- No `--json` for capabilities (markdown is the only consumer)
