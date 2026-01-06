# C3 Skill Harness

Shared behavioral constraints for all C3 skills.

## STOP - Before Any Action

| Check | Action |
|-------|--------|
| No `.c3/README.md`? | Route to `c3` skill (Adopt mode) |
| Wrong intent? | Route to correct skill (see routing table) |
| Uncertain about user intent? | Use `AskUserQuestion` until clear |

## Skill Routing

| User Intent | Skill |
|-------------|-------|
| "Where is X?" / Explore / Explain | `c3-skill:c3-query` |
| Add / Change / Modify / Implement | `c3-skill:c3-alter` |
| Audit / Validate / Check | `c3-skill:c3` |
| Adopt / Onboard / Init | `c3-skill:c3` |

## Red Flags - STOP Immediately

| Situation | Why Stop |
|-----------|----------|
| Making changes without ADR | All changes need ADR. No exceptions. |
| Guessing user intent | Ask, don't guess. Use `AskUserQuestion`. |
| Skipping layer navigation | Always start from Context (c3-0) down. |
| Editing code without docs first | Docs-first. ADR → Plan → Execute. |

## Required Tool: AskUserQuestion

All clarification MUST use `AskUserQuestion` tool:
- Structured answers reduce ambiguity
- Multiple-choice when options are clear
- Continue until no open questions

## Layer Navigation

Always load `layer-navigation.md` before traversing C3 docs.

Traversal order: **Context (c3-0) → Container (c3-N) → Component (c3-NNN)**
