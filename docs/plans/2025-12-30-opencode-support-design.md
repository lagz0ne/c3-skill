# Design: Dual Claude Code + OpenCode Support

**Date:** 2025-12-30
**Status:** Draft
**Author:** Lagz0ne + Claude

## Summary

Add OpenCode compatibility to the C3 plugin, enabling the same skills and agent to work in both Claude Code and OpenCode environments. Uses a monorepo approach with Claude Code as the source of truth and a Bun build script to transform/compile for OpenCode distribution via npm.

## Goals

- Support both Claude Code and OpenCode from a single codebase
- Maintain Claude Code as the primary/source format
- Publish OpenCode plugin to npm for easy installation
- Include starter hooks for future extensibility

## Non-Goals

- Replacing Claude Code support
- Feature parity with programmatic OpenCode capabilities (hooks are starters only)
- Supporting other AI CLI tools

## Architecture

### Directory Structure

```
c3-design/
├── .claude-plugin/              # Claude Code manifests (unchanged)
│   ├── plugin.json
│   └── marketplace.json
│
├── skills/                      # Source of truth (Claude format)
│   ├── c3-structure/SKILL.md
│   └── c3-implementation/SKILL.md
│
├── agents/                      # Source of truth (Claude format)
│   └── c3.md
│
├── commands/                    # Claude Code only
│   └── c3.md
│
├── references/                  # Shared (read at runtime by both)
│   └── *.md
│
├── src/
│   └── opencode/
│       ├── plugin.ts            # Main plugin export with hooks
│       └── hooks.ts             # Hook implementations (optional split)
│
├── scripts/
│   └── build-opencode.ts        # Bun build script
│
├── dist/                        # Generated (git-ignored)
│   └── opencode-c3/
│       ├── package.json
│       ├── plugin.js
│       ├── skill/
│       │   ├── c3-structure/SKILL.md
│       │   └── c3-implementation/SKILL.md
│       └── agent/
│           └── c3.md
│
├── package.json                 # Root with build scripts
└── .gitignore                   # dist/, .opencode/
```

### Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│                    Source (Claude Format)                    │
├─────────────────────────────────────────────────────────────┤
│  skills/*.md    agents/*.md    src/opencode/*.ts            │
└──────────┬─────────────┬────────────────┬───────────────────┘
           │             │                │
           ▼             ▼                ▼
┌──────────────────────────────────────────────────────────────┐
│              scripts/build-opencode.ts                       │
│  ┌────────────────┐ ┌────────────────┐ ┌──────────────────┐  │
│  │ Transform      │ │ Transform      │ │ Compile          │  │
│  │ Skills         │ │ Agent          │ │ TypeScript       │  │
│  │ (dir rename)   │ │ (frontmatter)  │ │ (Bun bundle)     │  │
│  └────────────────┘ └────────────────┘ └──────────────────┘  │
│                            │                                  │
│                   Generate package.json                       │
│                            │                                  │
│                      Verify outputs                           │
└──────────────────────────────┬───────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                   dist/opencode-c3/                          │
│         (Ready for npm publish or local testing)             │
└─────────────────────────────────────────────────────────────┘
```

## Build Script Responsibilities

### 1. Skill Transformation

```
skills/<name>/SKILL.md → dist/opencode-c3/skill/<name>/SKILL.md
```

- Validate name matches OpenCode pattern: `^[a-z0-9]+(-[a-z0-9]+)*$`
- Copy frontmatter (name, description) - already compatible
- Copy content as-is

### 2. Agent Transformation

```
agents/c3.md → dist/opencode-c3/agent/c3.md
```

| Claude Format | OpenCode Format |
|---------------|-----------------|
| `tools: Glob, Grep, Read...` | `tools: { glob: true, grep: true, read: true... }` |
| `model: opus` | `model: anthropic/claude-opus-4-5` |
| `color: cyan` | (removed - not supported) |
| (implicit) | `mode: subagent` |

### 3. Plugin Compilation

```
src/opencode/*.ts → dist/opencode-c3/plugin.js
```

- Compile TypeScript with Bun
- Bundle all hooks into single file

### 4. Package Generation

```
.claude-plugin/plugin.json → dist/opencode-c3/package.json
```

| Source Field | Target Field |
|--------------|--------------|
| `name: c3-skill` | `name: opencode-c3` |
| `version` | `version` |
| `author` | `author` |
| `description` | `description` |
| (new) | `main: "./plugin.js"` |
| (new) | `peerDependencies: { "@opencode-ai/plugin": "*" }` |

### 5. Build Verification

Script validates all required outputs exist before exiting:

```typescript
const required = [
  'dist/opencode-c3/package.json',
  'dist/opencode-c3/plugin.js',
  'dist/opencode-c3/skill/c3-structure/SKILL.md',
  'dist/opencode-c3/skill/c3-implementation/SKILL.md',
  'dist/opencode-c3/agent/c3.md',
]
// Exit non-zero if any missing
```

## Plugin Hooks

Starter implementations for all major hook types:

### tool.execute.before
- Warn on Context doc (`c3-0`) edits
- Block deletion of `.c3/` directory

### tool.execute.after
- Log C3 doc modifications

### file.edited
- Track ADR status changes
- Track container doc modifications

### session.created
- Auto-detect C3 project presence

### permission.ask
- Auto-allow reads on C3 docs
- Default to user decision for others

## CI/CD

### Workflow: `.github/workflows/publish-opencode.yml`

Triggers:
- On release published
- Manual dispatch

Steps:
1. Checkout
2. Setup Bun
3. Install dependencies
4. Run build script (includes verification)
5. Publish to npm

## Usage

### For OpenCode Users

```jsonc
// opencode.json
{
  "plugin": ["opencode-c3"]
}
```

### For Local Development

```jsonc
// opencode.json (in test project)
{
  "plugin": ["file:///path/to/c3-design/dist/opencode-c3"]
}
```

### Build Commands

```bash
bun run build:opencode      # Build to dist/
bun run dev:opencode        # Watch mode (future)
```

## Files to Create

| File | Purpose |
|------|---------|
| `src/opencode/plugin.ts` | Main plugin with hooks |
| `scripts/build-opencode.ts` | Bun build script |
| `package.json` | Add bun deps + build scripts |
| `.github/workflows/publish-opencode.yml` | CI for npm publish |

## Files to Modify

| File | Change |
|------|--------|
| `.gitignore` | Add `dist/`, `.opencode/` |

## Files Unchanged

- `.claude-plugin/*` - Claude Code manifests
- `skills/*` - Source of truth
- `agents/*` - Source of truth
- `commands/*` - Claude Code only
- `references/*` - Shared runtime resources

## Open Questions

1. **npm package name**: `opencode-c3` or `@c3-skill/opencode`?
2. **Watch mode**: Implement in v1 or defer?
3. **Local `.opencode/` generation**: Generate for local testing or only `dist/`?

## References

- [OpenCode Skills Documentation](https://opencode.ai/docs/skills/)
- [OpenCode Plugins Documentation](https://opencode.ai/docs/plugins/)
- [OpenCode Agents Documentation](https://opencode.ai/docs/agents/)
- [Claude Code to OpenCode Migration Guide](https://gist.github.com/RichardHightower/827c4b655f894a1dd2d14b15be6a33c0)
