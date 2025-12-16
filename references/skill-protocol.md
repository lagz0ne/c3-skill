# Skill Protocol

Standard patterns for C3 skill structure and invocation.

## Announcement Format

Every skill announces itself at invocation start:

```
"I'm using the {skill-name} skill to {action}."
```

Examples:
- "I'm using the c3-design skill to guide you through architecture design."
- "I'm using the c3-adopt skill to initialize architecture documentation for this project."

## Quick Reference Table

Every skill includes a quick reference table near the top:

```markdown
## Quick Reference

| Phase | Key Activities | Output |
|-------|---------------|--------|
| **1. [Phase Name]** | [Activities] | [Deliverable] |
| **2. [Phase Name]** | [Activities] | [Deliverable] |
```

This gives Claude and users immediate orientation before detailed instructions.

## Section Ordering

Standard skill SKILL.md structure:

1. **Frontmatter** - name, description
2. **Overview** - 1-2 sentences on purpose
3. **When to Use** - Triggering conditions (for model invocation)
4. **Quick Reference** - Table of phases/outputs
5. **[Core Workflow]** - The main process (skill-specific)
6. **[Templates/Checklists]** - If applicable
7. **Related Skills** - Cross-references

## Description Field

The `description` in frontmatter must include WHAT and WHEN:

```yaml
---
name: c3-component-design
description: Explore Component level impact during scoping - implementation details, configuration, dependencies, and technical specifics
---
```

Pattern: `[Action verb] [what] - [key elements covered]`

## Sub-Skill Delegation

When delegating to another skill, use this pattern:

```markdown
> "I'll use the {skill-name} skill to {specific task}."
```

Then provide:
1. Context/understanding built so far
2. Specific document or section to create
3. Key elements to include

## Checklist Format

Use checkboxes for verification checklists:

```markdown
### Checklist (must be true to call [LEVEL] done)

- [ ] Item 1 completed
- [ ] Item 2 verified
- [ ] Item 3 present
```

## Tool Usage Patterns

| Task | Tool |
|------|------|
| Read existing C3 docs | Read tool |
| Find C3 files | Glob tool with `.c3/**/*.md` |
| Search C3 content | Grep tool |
| Create/edit docs | Write/Edit tool |
| Generate TOC | build-toc.sh script in plugin |
| User confirmation | AskUserQuestion tool |

## Error Handling

When prerequisites aren't met:

```markdown
If `.c3/` doesn't exist:
- Stop and inform user
- Suggest: "Use the `c3-adopt` skill to initialize C3 documentation"
```

## Red Flags Section

Skills with decision points include rationalization counters:

```markdown
## Red Flags

| Rationalization | Counter |
|-----------------|---------|
| "[excuse]" | [why it's wrong and what to do instead] |
```
