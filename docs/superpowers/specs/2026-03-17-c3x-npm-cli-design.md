# @c3x/cli — npm thin client for c3x

## Problem

c3x is currently only usable through the Claude Code skill (agent context). Humans who want to run c3x commands directly have no distribution path — they must navigate to the skill's bin directory manually. We need an npm-distributed thin client so humans can `npx @c3x/cli list` or install globally.

## Design

### Package

- **Name**: `@c3x/cli` (org `c3x` on npm, owned by `lagz0ne`)
- **Location**: `packages/cli/` in the c3-design repo
- **Source**: `src/cli.ts` — TypeScript, zero runtime dependencies (Node.js stdlib only)
- **Built with**: tsdown (Rolldown) → `dist/cli.mjs`
- **Bin**: `{ "c3x": "./dist/cli.mjs" }` — exposes `c3x` command

### Discovery chain

The CLI searches for installed c3x binaries across multiple paths, reads the `VERSION` file from each, and delegates to the highest version found.

Search paths (all checked, highest version wins; on tie, earlier in list wins):

| Priority | Scope | Path pattern |
|----------|-------|-------------|
| 1 | Project | Walk up from cwd toward filesystem root, checking `skills/c3/bin/` at each level. Stops at first `.git` root or `/`. |
| 2 | Claude skills | `~/.claude/skills/c3/bin/` |
| 3 | Codex skills | `~/.codex/skills/c3/bin/` |
| 4 | Marketplace | `~/.claude/plugins/marketplaces/*/skills/c3/bin/` (glob — may match multiple) |

Each location contains a `VERSION` file (single line, semver e.g. `6.11.1`). The CLI:

1. Scans all paths, collects those where `VERSION` file exists and contains valid semver (major.minor.patch). Locations with missing, empty, or malformed VERSION files are skipped silently.
2. Compares versions using numeric semver ordering (major → minor → patch). Pre-release tags are not expected but if present, that location is skipped.
3. Delegates to the winner's `c3x.sh` wrapper (which handles platform/arch resolution)
4. If no valid installation found, exits with error and install instructions pointing to the Claude Code skill marketplace

### Human vs agent mode

Two separate concerns — the npm CLI and the skill each set the env independently:

- **Human (via npm CLI)**: `c3x.mjs` does **not** set `C3X_MODE` → Go binary uses default text output
- **Agent (via skill)**: The skill's `SKILL.md` instructions export `C3X_MODE=agent` before calling `c3x.sh` → Go binary outputs structured format

No changes to `c3x.sh` needed. The caller (npm CLI or skill) controls the env.

The Go binary checks `C3X_MODE` env var:
- Unset or `human` → default text output (existing behavior, unchanged)
- `agent` → equivalent to passing `--json` globally for commands that support it (`list`, `check`, `lookup`, `coverage`, `graph`). Commands without JSON support (`init`, `add`, `set`, `wire`, `unwire`, `delete`, `schema`, `codemap`) are unaffected. Explicit `--json`/`--compact` flags override `C3X_MODE` in either direction.

### Delegation

The npm CLI does **not** resolve platform binaries itself. It delegates to `c3x.sh`, which already handles:
- OS detection (`uname -s` → darwin/linux)
- Arch detection (`uname -m` → amd64/arm64)
- Version-matched binary selection (`c3x-{version}-{os}-{arch}`)
- Stale binary cleanup

**Working directory**: `c3x.mjs` spawns `c3x.sh` by absolute path but does **not** change cwd. `execFileSync` inherits the user's cwd by default. The chain:

1. User runs `c3x list` in `/home/user/myproject`
2. `c3x.mjs` resolves `c3x.sh` at e.g. `~/.claude/skills/c3/bin/c3x.sh`
3. Spawns `bash /absolute/path/to/c3x.sh list` — cwd stays `/home/user/myproject`
4. `c3x.sh` uses its own `SCRIPT_DIR` to locate the platform binary, then `exec`s it — cwd still `/home/user/myproject`
5. Go binary discovers `.c3/` by walking up from cwd — works correctly

### Agent filtering

By default, all paths are scanned. Use `--agent` to restrict to a specific agent type:

```bash
c3x --agent claude list    # only search Claude paths (skills + marketplace)
c3x --agent codex list     # only search Codex paths
c3x list                   # search all (default)
```

`--agent` is consumed by the npm CLI and **not** forwarded to the Go binary.

| `--agent` value | Paths searched |
|-----------------|---------------|
| (omitted) | All 4 scopes |
| `claude` | Project + Claude skills + Marketplace |
| `codex` | Project + Codex skills |

Project scope is always included — it's agent-agnostic.

### Execution flow

```
npx @c3x/cli list
  → bin/c3x.mjs
    → scan applicable path categories (filtered by --agent if set)
    → read VERSION from each found location
    → pick highest semver (ties: earlier priority wins)
    → child_process.execFileSync("bash", [c3x.sh, ...args])
    → inherit stdio
    → exit with same code
```

## Package structure

```
packages/cli/
├── package.json
├── tsdown.config.ts
├── src/
│   └── cli.ts            # Source: discovery + delegation
├── dist/
│   └── cli.mjs           # Built output (gitignored)
└── README.md
```

### Build

Uses **tsdown** (Rolldown-powered bundler) to compile TypeScript source to a single ESM file with Node.js shebang.

```typescript
// tsdown.config.ts
import { defineConfig } from 'tsdown'

export default defineConfig({
  entry: ['src/cli.ts'],
  format: 'esm',
  platform: 'node',
  target: 'node18',
  outDir: 'dist',
  clean: true,
  banner: { js: '#!/usr/bin/env node' },
})
```

### package.json

```json
{
  "name": "@c3x/cli",
  "version": "0.1.0",
  "description": "Thin CLI client for c3x — architecture documentation tool",
  "type": "module",
  "bin": {
    "c3x": "./dist/cli.mjs"
  },
  "files": ["dist/"],
  "scripts": {
    "build": "tsdown"
  },
  "devDependencies": {
    "tsdown": "^0.x"
  },
  "engines": {
    "node": ">=18"
  },
  "license": "MIT"
}
```

## Changes required

### New files
- `packages/cli/package.json` — package manifest
- `packages/cli/tsdown.config.ts` — build config
- `packages/cli/src/cli.ts` — source: discovery + delegation (~60 lines)
- `packages/cli/README.md` — usage instructions

### Modifications
- Go CLI (`cli/cmd/options.go`) — check `C3X_MODE` env var, set `JSON: true` on options when `agent` for commands that support it
- `skills/c3/SKILL.md` — add `export C3X_MODE=agent` before c3x.sh invocations

## Not in scope

- Binary management (download, install, update) — relies on skill installations
- Windows support — c3x.sh is bash, targets macOS/Linux
- Version pinning — always uses highest available version
- Workspace/monorepo config — `packages/cli/` is a standalone publishable package
