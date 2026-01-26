# C3 skill design

This is a repository containing a Claude code skill called c3. c3 is a trimmed down concept from c4, focusing on rationalizing architectural and architectural change for large codebase

# Required skills, always load them before hand
- load /superpowers:brainstorming, always do
- load /superpowers-developing-for-claude-code:developing-claude-code-plugins, always do
- use AskUserQuestionTool where possible, that'll give better answer

# Workflow
- Starts with brainstorming to understand clearly the intention
- Once it's all understood, use writing-plan and implement in parallel using subagent
- Delegate to /release command once things is done, confirm with user as needed. Assume to patch by default

---

# Skill Development Philosophy

## The Structure vs Outcome Balance

Shift from prescriptive structure to goal-oriented guidance:

| Structural (Avoid) | Outcome-Oriented (Prefer) |
|--------------------|---------------------------|
| "Create separate file per component" | "Each component should be findable and answerable" |
| "Follow this directory structure" | "Structure should enable navigation" |
| "Add these specific sections" | "Enable developer to understand X" |

## Skill Development Loop

```
Goals → Minimal Skill → Meaningful Eval → Run → Trace Failures → Update Skill → Repeat
```

1. **Start with Goals** - What should the skill enable? What questions should users answer?
2. **Write Minimal Skill** - Focus on outcomes, not structure
3. **Build Meaningful Eval** - Question-based, grounded in actual use
4. **Run Eval** - Identify failing questions
5. **Trace Failures** - Map failures back to skill gaps
6. **Update Skill** - Add targeted guidance, remove unhelpful rules
7. **Repeat** - Improving scores should improve real utility

## Skill Writing Principles

1. **Fewer meaningful sections > many shallow sections** - Don't prescribe structure for structure's sake
2. **Outcome-focused instructions** - "Docs should enable X" not "Docs should contain Y"
3. **Anti-goals are important** - Explicitly state what NOT to do (came from eval failures)
4. **Progressive disclosure** - Core requirements first, details only when needed

## When to Use Prescriptive vs Goal-Oriented

| Use Prescriptive | Use Goal-Oriented |
|------------------|-------------------|
| Hard technical constraints (file names, IDs) | Content quality |
| Integration points (CI expects certain paths) | Organization decisions |
| Non-negotiable conventions | Style and depth |

---

# Plugin Structure Checklist

## Pre-Release Checklist

Run before every release to ensure plugin loads correctly:

| Check | Required | File |
|-------|----------|------|
| `name` field exists | Yes | `.claude-plugin/plugin.json` |
| NO explicit component paths | Yes | `.claude-plugin/plugin.json` (auto-discovery only) |
| Commands have YAML frontmatter with `description` | Yes | `commands/*.md` |
| Skills have `SKILL.md` with frontmatter | Yes | `skills/*/SKILL.md` |
| Agents have YAML frontmatter with `name`, `description` | Yes | `agents/*.md` |
| Hook scripts exist at referenced paths | Yes | `scripts/*` |
| Hooks use `${CLAUDE_PLUGIN_ROOT}` for paths | Yes | `hooks/hooks.json` |

## plugin.json Template

**IMPORTANT:** Claude Code uses auto-discovery. Do NOT add explicit component paths - they break plugin loading.

```json
{
  "name": "c3-skill",
  "version": "x.x.x"
}
```

Components are auto-discovered from standard directories:
- `commands/` - Slash commands
- `skills/` - Agent skills
- `agents/` - Subagent definitions
- `hooks/hooks.json` - Event handlers

## Common Issues

| Symptom | Cause | Fix |
|---------|-------|-----|
| Components not loading | Explicit path declarations in plugin.json | REMOVE `commands`, `skills`, `agents`, `hooks` fields - use auto-discovery |
| Plugin not loading | Version mismatch in `installed_plugins.json` | Update path/version in `~/.claude/plugins/installed_plugins.json` |
| Plugin marked orphaned | `.orphaned_at` file in cache | Remove the file from cached plugin directory |
| Hooks not firing | hooks.json not in `hooks/` directory | Move to `hooks/hooks.json` (auto-discovered) |

## Validation Steps

1. Load `plugin-dev:plugin-structure` skill for structure reference
2. Run `plugin-dev:plugin-validator` agent on plugin root
3. Fix issues reported by validator
4. Restart Claude Code session (components load at startup)

## Testing Plugin Behavior

The plugin structure matches the installed format:

```
c3-design/                    # Repository root
├── .claude-plugin/           # Plugin metadata
│   ├── plugin.json          # Manifest with paths to content
│   └── marketplace.json     # Marketplace publishing config
├── skills/                  # Skill definitions (at root)
├── agents/                  # Agent definitions (at root)
├── commands/                # Slash commands (at root)
├── hooks/                   # Hook configurations (at root)
├── references/              # Shared reference docs
├── templates/               # Doc templates
└── scripts/                 # Helper scripts
```

**Testing locally with --plugin-dir:**

Note: `--plugin-dir` conflicts with installed plugins of the same name. For local testing:

```bash
# Option 1: Temporarily uninstall the marketplace plugin
claude plugin uninstall c3-skill

# Then test with --plugin-dir
claude --plugin-dir /path/to/c3-design -p "list skills"

# Re-install when done
claude plugin install c3-skill

# Option 2: Just release and test the installed version
# Run /release to bump version, then the plugin auto-updates
```

**Structure notes:**
- `.claude-plugin/plugin.json` paths (`"skills": "skills"`) are relative to the repo root
- When installed, the whole repo is copied to `~/.claude/plugins/cache/.../version/`
- The directory name for `--plugin-dir` becomes the skill prefix