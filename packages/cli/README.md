# @c3x/cli

Thin CLI wrapper for [c3x](https://github.com/lagz0ne/c3-skill) — the architecture documentation toolkit.

This package does **not** bundle the c3x binary. It discovers an already-installed binary from your Claude Code or Codex skill installations and delegates to it.

## Install

```bash
# Prerequisites: install the c3 skill in Claude Code or Codex
claude plugin install lagz0ne/c3-skill

# Then use c3x directly
npx @c3x/cli list
# or install globally
npm install -g @c3x/cli
c3x list
```

## How it works

`@c3x/cli` searches for the c3x Go binary across agent skill installation paths, picks the highest version, and delegates all commands to it. Your working directory is preserved — `.c3/` discovery works from wherever you run it.

### Resolution order

| Priority | Location |
|----------|----------|
| 1 | `<project>/skills/c3/bin/` (walks up from cwd, stops at `.git`) |
| 2 | `~/.claude/skills/c3/bin/` |
| 3 | `~/.codex/skills/c3/bin/` |
| 4 | `~/.claude/plugins/marketplaces/*/skills/c3/bin/` |

All paths are checked. Highest semver wins. On tie, earlier priority wins.

### Agent filtering

```bash
c3x --agent claude list    # only search Claude paths (+ project)
c3x --agent codex list     # only search Codex paths (+ project)
c3x list                   # search all (default)
```

## Human vs agent output

| Caller | Output |
|--------|--------|
| Human via `@c3x/cli` | Text (default) |
| Agent via `/c3` skill | Agent format through the local skill wrapper |

Explicit `--json` or `--compact` flags override either default.

## License

MIT
