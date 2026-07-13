# Sweep Reference

Pre-flight for a change-unit. Given a proposed change to a frozen fact, predict two
things before any patch is authored: the **blast radius** (who depends on it) and
**whether the tool will let the change land** (the destruction gate). Advisory only —
sweep authors nothing.

Discovery and the causal chain belong to **query.md** — start there to find the target
and trace owner → mutation → dependent. Sweep takes a known target and works the
reverse graph.

## The reverse graph is the spine

```bash
C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" graph <id> --direction reverse --depth 1
```

Reverse = who points *at* the changed fact (live children, citers). Each edge is a
**candidate** consequence, not a settled one: confirm a dependent is actually affected
with `C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" read <dependent-id>` before naming it — never mark every neighbor affected by
default. For ref/rule impact, graph the ref/rule itself to surface all citers.

Read the graph `route:` block at the same time. `route.facts` and `route.graph` show
the context pack, `route.anchors` names first files/docs/tests to inspect, `route.lanes`
names the lifecycle or ownership lane, `route.drift` names stale or unreadable anchor
bindings, and `route.hash` is only a change signal. The route helps you inspect the
right path; it does not prove impact, code conformance, or safe deletion.

## Will the destruction gate let it land?

If the change **removes or retires** a fact, the reverse graph *is* the refusal
prediction. `change apply` runs a `retire` gate that **REFUSES** the unit while the
retired fact still has, in the frozen graph:

- **live children** → they would be ORPHANED, *unless this same unit retires them too
  or reparents them away (a `frontmatter` patch to a live parent)*; and
- **live citers** → their citations would DANGLE, *unless this same unit drops that
  citation (or retires the citer)*.

So the sweep deliverable for a removal is the **list of consequences the unit must
also carry**: every orphaned child to reparent/retire, every dangling citer to rewire.
Membership rows are never on that list — a parent's membership table is synthesized
from `parent:` links, so the row drop is automatic.

## Bridge to the saga

The change lands as patches in `.c3/changes/<unit-id>/`, applied all-or-nothing. Once
those patches are staged, preview the post-change graph *before* `apply`:

```bash
C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" graph <id> --unit <adr-id> --direction reverse
```

This renders the graph as it *would* be with the unit's staged patches applied —
confirm the orphans/dangles you predicted are healed before committing. If the preview
includes route facets, compare the anchors/lanes as first-inspection clues, not as a
destruction-gate substitute.

## Deliverable

```
**C3 Impact Assessment**

**Proposed Change:** [summary]

## Affected Entities
| Entity | Type | Impact | Reason |
|--------|------|--------|--------|
| c3-N | container | direct | [why — confirmed via the wrapper's `read` operation] |

## File Changes Required (patches in .c3/changes/<unit-id>/, land all-or-nothing)
| File | Change | Component |
|------|--------|-----------|
| src/path/file.ts | [mod] | c3-NNN |

## Risks
- [Risk]: [impact + mitigation]

## Verification
| Check | How |
|-------|-----|
| destruction gate: does retire/apply succeed or refuse? | reverse graph clean, or every orphan/dangle healed in the unit |
| [owner entity/file updated] | [wrapper's `lookup` / `read` operation used to confirm] |
| [runtime value or observable] | [command or observable to confirm] |
| [failure-mode probe] | [what to break + expected degradation] |
```

`Impact` distinguishes `direct` (cites or consumes the changed fact) from `transitive`
(reached through another dependent). An assessment without the Verification table —
including the destruction-gate row — is advice, not a pre-flight.
