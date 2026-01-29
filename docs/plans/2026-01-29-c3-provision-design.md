# C3 Provision - Architecture-First Workflow

## Problem

Current c3 skills assume implementation follows design. For consulting/design work, users need to:
- Document architecture before code exists
- Capture requirements with stakeholders
- Plan future roadmap as firm architecture artifacts
- Review and approve designs without committing to immediate implementation

## Decision

Add `c3-provision` skill for design-only architectural changes with explicit lifecycle management.

## Core Model

### Component Lifecycle Status

```yaml
# Component frontmatter
---
id: c3-205-rate-limiter
status: provisioned    # | active | deprecated
---
```

| Status | Meaning | Code References |
|--------|---------|-----------------|
| `provisioned` | Architecture defined, not yet built | None |
| `active` | Implemented and in use | Required |
| `deprecated` | Being phased out | May have |

### File Organization

**New components (no existing active version):**
```
.c3/provisioned/
  c3-2-api/
    c3-205-rate-limiter.md      # status: provisioned
```

**Changes to existing active components:**
```
.c3/c3-2-api/
  c3-201-auth.md                # status: active (current reality)
.c3/provisioned/
  c3-2-api/
    c3-201-auth.md              # planned version (same ID, different location)
```

**Key insight:** Provisioned directory mirrors container structure. No ID collisions because location distinguishes active vs provisioned.

### Provisioned File Structure

```yaml
---
id: c3-201-auth
status: provisioned
supersedes: ../c3-2-api/c3-201-auth.md   # relative path to active version (if exists)
adr: adr-20260129-oauth-support
---
# Auth Middleware (Provisioned)

[Full doc as it will be after implementation]
```

## ADR Lifecycle

### Two Separate ADRs

| ADR | Purpose | Status Progression |
|-----|---------|-------------------|
| Provisioned ADR | Design decision | proposed → accepted → **provisioned** |
| Implementation ADR | Code decision | proposed → accepted → implemented |

### ADR Linking

When implementing a provisioned component:

```yaml
# Implementation ADR frontmatter
---
id: adr-20260215-oauth-impl
implements: adr-20260129-oauth-support   # links to provisioned ADR
status: proposed
---
```

When implementation completes:
- Implementation ADR: `status: implemented`
- Provisioned ADR: `status: superseded`, add `superseded-by: adr-20260215-oauth-impl`

## Workflow: c3-provision

```
User Request ("provision X", "design X", "plan X")
    ↓
Intent Clarification (Socratic)
    ↓
c3-analysis (state + impact + patterns)
    ↓
c3-synthesizer (critical thinking)
    ↓
ADR Created (status: proposed)
    ↓
User Accepts
    ↓
ADR status: provisioned
    ↓
Create Component Doc in .c3/provisioned/
    ├── New component: .c3/provisioned/c3-X/c3-xxx.md
    └── Existing component: .c3/provisioned/c3-X/c3-xxx.md (with supersedes:)
    ↓
DONE (no execution phase)
```

## Workflow: Implementing Provisioned Components

```
User: "implement rate limiter"
    ↓
c3-alter detects: .c3/provisioned/c3-X/c3-xxx.md exists
    ↓
Load provisioned doc as starting point
    ↓
NEW ADR (status: proposed) with implements: link
    ↓
User Accepts
    ↓
c3-dev executes TDD
    ↓
On completion:
    ├── Implementation ADR: status: implemented
    ├── Provisioned ADR: status: superseded
    ├── Move .c3/provisioned/.../c3-xxx.md → .c3/c3-X/c3-xxx.md
    ├── Update status: provisioned → active
    ├── Add Code References
    └── Delete provisioned file (now moved)
```

## Key Behaviors

### c3-provision Skill
- Full orchestrator analysis (same as c3-alter)
- Stops after ADR acceptance + doc creation
- No c3-dev execution
- Terminal ADR status: `provisioned`

### Navigator/Query Behavior
- Default: search `.c3/c3-*` (active components only)
- On "what's planned?", "show provisioned": include `.c3/provisioned/`
- Warn if querying a component that has a provisioned version

### Audit Behavior
- Validate `.c3/provisioned/` components have no Code References
- Validate `.c3/c3-*/` active components have Code References
- Warn on stale provisioned (active component changed after provisioned created)
- Report orphaned provisioned files (ADR superseded but file not moved)

## Changes Required

### New Skill
- `skills/c3-provision/SKILL.md` - provision workflow

### Modified Skills
- `skills/c3-alter/SKILL.md` - detect provisioned components, handle transition
- `skills/c3/SKILL.md` - route to c3-provision for design-only requests

### Modified Agents
- `agents/c3-orchestrator.md` - support ADR status: provisioned
- `agents/c3-navigator.md` - filter/surface provisioned docs appropriately
- `agents/c3-adr-transition.md` - handle provisioned → implemented transition

### New/Modified References
- `references/adr-template.md` - add provisioned status, implements/superseded-by fields
- `references/component-lifecycle.md` - document status field and directory structure

## Trigger Examples

| User Says | Routes To |
|-----------|-----------|
| "provision a rate limiter" | c3-provision |
| "design the auth flow" | c3-provision |
| "plan a new payment service" | c3-provision |
| "add rate limiting" (no .c3/) | c3-onboard first |
| "implement the rate limiter" (provisioned exists) | c3-alter |
| "what's planned?" | c3-navigator (includes provisioned/) |

## Resolved Questions

1. **ID collisions?** → Solved by `.c3/provisioned/` directory. Same ID, different location.

2. **ADR lifecycle?** → Two separate ADRs linked via `implements:` / `superseded-by:`.

3. **Merge/replace rule?** → Move provisioned file to active location, update status, add Code References.

## Open Questions

1. **Multiple provisioned changes to same component?**
   - Only one provisioned version per component at a time
   - New provision overwrites previous (with warning)
   - Or: reject until previous is implemented/cancelled

2. **Staleness detection threshold?**
   - How old is "stale" for provisioned vs active drift?
   - Recommendation: warn on any change, let user decide
