# C3 Skills Distribution

This branch contains the built distribution of C3 skills.

## Installation

**Claude Code:**
```bash
claude plugin install c3-skill
```

**OpenCode:**
Use the `opencode/` subdirectory.

**Codex:**
Use the `codex/` subdirectory (skills only, no agents).

## Structure

- `skills/` - Self-contained skill packages with bundled references
- `agents/` - Agent definitions
- `.claude-plugin/` - Plugin manifest
- `opencode/` - OpenCode-compatible plugin (skills + agents)
- `codex/` - Codex-compatible skills (no agent support)
