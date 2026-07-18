# Sweep Reference

Pre-flight for a change-unit. Given a proposed change to a frozen fact, predict two
things before any patch is authored: the **blast radius** (who depends on it) and
**whether the tool will let the change land** (the destruction gate). Advisory only —
sweep authors nothing.

Discovery and the causal chain belong to **query.md** — start there to find the target
and trace owner → mutation → dependent. Sweep takes a known target and works the
reverse graph.

Start discovery with `search "<change>" --pack --limit 3`. Treat returned
`record_claims` as a compact checklist of source-backed current mechanisms and failure
boundaries to verify. They are stored claims, not automatic runtime proof; inspect the
cited source before classifying impact.

If the selected pack has no matching `record_claims`, treat that as low C3 coverage.
Do not expand semantically adjacent records into a confident answer. Search the
repository directly for every requested surface, attach one exact current-mechanism
anchor per surface, and leave any unclosed surface unknown with the next check.

## Use three bounded routes

Run these as independent routes. Redirect duplicate findings instead of repeating
the same read.

1. **Reverse dependency route** — run
   `C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" graph <id> --direction reverse --depth 1`
   to find parent-owned children and inbound wiring/citers.
2. **Code propagation route** — start from `lookup` and `route.anchors`; inspect narrow
   callers, consumers, persistence/event boundaries, and focused tests.
3. **Contract and failure route** — read governing refs/rules and inspect the failure,
   compatibility, migration, and isolation lane named by `route.lanes`.

Reverse = who points *at* the changed fact (live children, citers). Each edge is a
**candidate** consequence, not a settled one: confirm a dependent is actually affected
with `C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" read <dependent-id>` before naming
it — never mark every neighbor affected by default. For ref/rule impact, graph the
ref/rule itself to surface all citers.

Containment is traversal adjacency for impact and destruction checks, not a wiring
relationship edge. A child appears because retiring its parent would orphan it; that
does not prove runtime propagation between parent and child.

Read the graph `route:` block at the same time. `route.facts` and `route.graph` show
the context pack, `route.anchors` names first files/docs/tests to inspect, `route.lanes`
names the lifecycle or ownership lane, `route.drift` names stale or unreadable anchor
bindings, and `route.hash` is only a change signal. The route helps you inspect the
right path; it does not prove impact, code conformance, or safe deletion.

Expand beyond depth 1 only when a confirmed mechanism carries the change into a new
hop. Stop when every surfaced lane is classified **affected / unaffected / unknown**,
each affected hop has evidence, and every unknown has the smallest next check. A flat
list of files is not closure.

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

## Impact Classification
| Surface | Direct/transitive | Required change | Evidence | Status |
|---------|-------------------|-----------------|----------|--------|
| c3-N / runtime surface | direct | [change or preserved contract] | [exact c3 id + section, or file:symbol/test] | affected / unaffected / unknown |

## Route Coverage
Write one row for every surfaced lane; do not compress several lanes into one summary.

| Route | Lane | Evidence checked | Status | Next check |
|-------|------|------------------|--------|------------|
| Reverse dependency route | [child or citer lane] | [exact read/graph evidence] | affected / unaffected / unknown | [smallest check, or none] |
| Code propagation route | [caller/consumer/state/test lane] | [source/test evidence] | affected / unaffected / unknown | [smallest check, or none] |
| Contract and failure route | [rule/failure/compatibility/isolation lane] | [ref/rule/runtime evidence] | affected / unaffected / unknown | [smallest check, or none] |

Unknown rows require evidence checked and a next check. An unseen lane is not
`unaffected`; record why it was not inspectable.

## Code Changes Proposed
| File or symbol | Change | Owning component | Evidence |
|----------------|--------|------------------|----------|
| src/path/file.ts | [mod] | c3-NNN | [lookup plus narrow source/test read] |

## C3 Fact Patches Required
| Fact | Section | Reason |
|------|---------|--------|
| c3-NNN | [section, or none] | [shared truth changed, or no C3 patch required] |

## Isolation Boundaries
- [Surface that must not change]: [evidence and verification]

## Unknowns
- [Unresolved category]: [smallest next command/check; never guess]

## Risks
- [Risk]: [impact + mitigation]

## Verification
| Check | How |
|-------|-----|
| destruction gate: does retire/apply succeed or refuse? | reverse graph clean, every orphan/dangle healed in the unit, or N/A for a change that does not remove or retire a fact |
| [owner entity/file updated] | [wrapper's `lookup` / `read` operation used to confirm] |
| [runtime value or observable] | [command or observable to confirm] |
| [failure-mode probe] | [what to break + expected degradation] |
```

`Direct/transitive` distinguishes a surface that cites or consumes the changed fact
from one reached through another confirmed mechanism. `Evidence` binds every material
claim to the exact C3 section, source symbol, test, or runtime check used. `Unknown`
is a valid status; converting it to `affected` without evidence is a false claim.

For pre-change work, write current truth before the proposed boundary:

| Current claim | Stored record class/state | Source anchor | Runtime proof | Verdict |
|---|---|---|---|---|
| [what exists now] | fact / change-doc + state | [path:symbol or citation] | [narrow source/test read] | proved / contradicted / unknown |

`fact` means a frozen C3 record, not automatically implemented runtime behavior.
Change-doc state describes the decision record. Never turn intended or planned
text into a current implementation claim without source proof.

An assessment without route closure, isolation boundaries, unknowns, and the
Verification table is advice, not a pre-flight.
