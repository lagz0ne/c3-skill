---
description: Audit C3 documentation against codebase reality
---

Verify `.c3/` documentation matches actual implementation.

## Flow

1. **Read current state**
   - `.c3/TOC.md` and all docs
   - `.c3/settings.yaml`

2. **Explore codebase** (Task Explore, very thorough)
   - Discover containers (package.json, go.mod, Dockerfile, etc.)
   - Discover components within each container

3. **Audit categories**

   | Category | Check |
   |----------|-------|
   | Config compliance | Docs follow settings.yaml |
   | Vision vs fact | Documented = actual code |
   | ADR status | Proposed/accepted/stale/abandoned |
   | Skill compliance | Frontmatter, anchors, sections |

4. **Generate findings report**

   ```
   ## Findings

   ### Vision vs Fact
   - ❌ c3-3-worker documented but not found
   - ⚠️ /services/notification undocumented
   - ✅ c3-1-backend matches /api

   ### ADR Status
   - ⚠️ ADR-001 proposed 30+ days
   - ✅ ADR-002 implemented

   ### Summary: 1 critical, 2 warnings, 2 passing
   ```

5. **Handoff**

   Check `settings.yaml` `audit:` section for preference.

   If not set, ask:
   ```
   How to handle findings?
   1. Manual - review report yourself
   2. Tasks - create vibe-kanban tasks
   3. Agents - dispatch subagents to fix

   Save preference to settings.yaml?
   ```

   Execute chosen handoff:
   - **Manual:** done
   - **Tasks:** use vibe-kanban MCP to create tasks
   - **Agents:** use Task tool to dispatch fixes in parallel

## References

- [derivation-guardrails.md](../references/derivation-guardrails.md) - Core principles
- [v3-structure.md](../references/v3-structure.md) - Expected structure
