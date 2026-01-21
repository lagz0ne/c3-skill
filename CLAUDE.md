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

# Evaluation Philosophy

## Core Principle

**Test outcomes, not structure.** Meaningful evaluation tests whether artifacts achieve their purpose, not whether they match a template.

## Evaluation Quality Hierarchy

| Level | Approach | What It Tests |
|-------|----------|---------------|
| Best | Outcome-based | Does artifact enable its purpose? |
| Mid | Behavior-based | Does artifact work correctly? |
| Worst | Structure-based | Does artifact match expected shape? |

Always start from outcomes and work backward to find minimal structure requirements.

## Question-Based Evaluation Pattern

The gold standard for evaluating generated artifacts:

```
Codebase Analysis → Question Generation → Fresh-Context Answering → Verification
       ↓                    ↓                       ↓                    ↓
  Extract truth      Simulate real            No cheating         Ground truth
  from source        developer needs          (isolated LLM)      comparison
```

**Why this works:**
- Grounded in THIS system, not generic templates
- Simulates real use with fresh context (prevents information leakage)
- Binary correctness - either the answer matches reality or it doesn't
- Actionable failures - know exactly which question failed and why

**Implementation:** `eval/lib/question-eval.ts`

## Anti-Patterns to Avoid

| Anti-Pattern | Why It Fails | Alternative |
|--------------|--------------|-------------|
| Golden example comparison | Measures similarity, not utility | Question-based testing |
| Structural checklists | Creates template compliance, not quality | Outcome verification |
| Generic dimensions | Same scores for different codebases | Codebase-specific questions |
| Single-context evaluation | LLM can cheat with prior knowledge | Fresh context isolation |
| Bidirectional links | Maintenance burden, stale links | Unidirectional hierarchy |
| Monolithic docs | Hard to maintain and navigate | Separate files per concept |

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

# Running Evaluations

```bash
# Expectations-based (structural)
bun eval/run.ts eval/cases/onboard-simple.yaml

# Question-based (meaningful)
bun eval/test-question-eval.ts <codebase-path> <docs-path>

# Example
bun eval/test-question-eval.ts eval/fixtures/simple-express-app /tmp/output/.c3
```

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