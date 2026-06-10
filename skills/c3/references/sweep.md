# Sweep Reference

Impact assessment. Advisory only — no changes.

Flow: `Discover Candidates → Impact Topology → Per-Entity Assessment → Synthesize`

Spawn subagents via Task tool for parallel per-entity assessment.

## Progress

- [ ] Topology loaded
- [ ] Affected entities identified
- [ ] Per-entity assessment complete
- [ ] Constraint chain checked
- [ ] Synthesis delivered

---

## Step 1: Discover Candidates

For natural-language impact questions ("what breaks if...", "what is affected by...",
"is this safe?"), start with conceptual search:

```bash
c3 search "<impact question>"
```

Read the best matching refs, components, recipes, or ADRs, then expand to graph/list
only after the primary affected entity is known. Use `c3 list` first only when the
user asks for topology-wide inventory or search misses.

Known entity/ref/component -> start with the entity graph:

```bash
c3 graph <id> --direction reverse --depth 1
```

For property-style impact questions, do not stop at the first owner:

- **Config/scope/prefix/env changes:** run reverse graph on the config/scope ref
  and on the behavior ref it feeds. Enumerate concrete dependent ids and name the
  result as blast radius/scope of impact.
- **Transport/auth changes that feed sync or delivery:** trace credential
  generation, broker/enforcement, and client/runtime consumption. Name the
  coupling and the shared subject/permission/token that carries the impact.
- **Partial failure or bulk operation safety:** trace the mutation boundary,
  transaction/audit/storage contract, and observation/query surface. Name
  atomicity, consistency, idempotency, or partial-success boundary when those
  properties are what the user is really asking about.

## Step 2: Impact Topology

```bash
c3 list
```

Use topology to confirm containers/components/refs after search identifies candidates.

## Step 3: Affected Entities

From proposed change, identify:
- Containers, components, refs, ADRs
- Match by: title, relationships, code-map entries, ref scopes

## Step 4: Per-Entity Assessment

Subagents for parallelism when multiple containers affected.

**Container:** `c3 read <container-id>` → change affect responsibilities? → identify affected components.

**Component:**
1. `c3 read <component-id>`
2. Per code-map file: `c3 lookup <file>` — loads constraint chain before inspecting code
3. Check code against constraints
4. Change modify behavior, API, dependencies?
5. Check applicable refs. Identify downstream dependents.

**Ref:** `c3 read <ref-id>` → proposed change comply or violate? → note severity + override requirements.

**Rule:** `c3 read <rule-id>` → proposed change violate golden pattern? Note severity + remediation.

## Step 5: Constraint Chain

Per affected component, trace upward:
- Component constraints → container → context → cited refs + rules

Flag any proposed violation.

## Step 6: Synthesize

**Impact Graph:** Include `c3 graph <target-entity> --direction reverse --format mermaid` as mermaid code block atop report (reverse direction = who depends on the changed entity). Graph from most specific affected entity (component > container). For ref/rule impact, graph ref/rule itself to show all citers.

**Direct vs transitive:** a reverse-graph edge is a candidate, not a conclusion. Assign concrete behavior to a dependent only after `c3 read` of that dependent. In the Affected Entities table, `Impact` distinguishes `direct` (cites or consumes the changed entity) from `transitive` (reached through another dependent) — never mark every graph neighbor as affected by default.

```
**C3 Impact Assessment**

**Proposed Change:** [summary]

## Affected Entities
| Entity | Type | Impact | Reason |
|--------|------|--------|--------|
| c3-N | container | direct | [why] |

## Constraint Chain
| Source | Constraint | Status |
|--------|-----------|--------|
| c3-0 | [rule] | compliant/violated |

## File Changes Required
| File | Change | Component |
|------|--------|-----------|
| src/path/file.ts | [mod] | c3-NNN |

## Risks
- [Risk]: [impact + mitigation]

## Verification
| Check | How |
|-------|-----|
| [owner entity/file updated] | [c3 lookup / read to confirm] |
| [config/permission/runtime value] | [command or observable to confirm] |
| [sync/notification observable] | [subject/channel/log to assert] |
| [failure-mode probe] | [what to break + expected degradation] |

## Recommended Approach
1. [Step respecting constraints]
```

An assessment without the Verification table is advice, not an assessment — every sweep ends with checks someone can actually run.
